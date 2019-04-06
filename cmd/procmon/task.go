package main

import (
	"context"
	"io"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

// Task is a restartable process; it can be monitored and
// restarted.
// By default, a task is monitored simply as a running process. If the
// process terminates for any reason, the task's status channel
// is notified.
// The task can also be given a set of monitors that can also
// watch the task for incorrect performance; they also communicate
// on the status channel.
// The task's monitors listen on a done channel for the task; when the
// task is detected to have failed, that done channel is closed and the
// monitors shut down (the status channel is drained to prevent leaks).
// If the status channel receives a termination notice, the task is
// shut down if it was not already stopped.
// When a task stops, its child tasks are also shut down.
// During shutdown, the tasks that are being deliberately terminated
// will not be automatically restarted.
// Once a task and all of its children have been shut down, the task is
// then restarted (which will cause all of its children also to restart).
type Task struct {
	Name        string
	Path        string
	Args        []string
	Env         []string
	Onetime     bool
	MaxShutdown time.Duration
	MaxStartup  time.Duration
	Status      chan Eventer
	Ready       func() Eventer
	Stdout      io.WriteCloser
	Stderr      io.WriteCloser
	Logger      logrus.FieldLogger
	FailCount   int
	Monitors    []*Monitor
	Prerun      []*Task
	Dependents  []*Task

	cmd    *exec.Cmd
	cancel context.CancelFunc
	done   chan struct{}
}

// NewTask creates a Task (but does not start it)
// The default Ready() function simply returns true
func NewTask(name string, path string, args ...string) *Task {
	return &Task{
		Name:        name,
		Path:        path,
		Args:        args,
		MaxStartup:  10 * time.Second,
		MaxShutdown: 5 * time.Second,
		Status:      make(chan Eventer, 1),
		Ready:       func() Eventer { return OK },
		Monitors:    make([]*Monitor, 0),
	}
}

// Listen implements Listener, and should be called as a goroutine.
func (t *Task) Listen(done chan struct{}) {
	for {
		select {
		case <-done:
			// we're being asked to quit
			t.Logger.Info("done channel closed; listener shutting down")
			return
		case e := <-t.Status:
			t.Logger.WithField("status", e).Info("listener")
			if IsFailed(e) {
				t.Logger.Info("shutdown and restart")
				// we need to shut down, so turn off all our monitors
				t.StopMonitors()
				// record the fact that we failed
				t.FailCount++
				logger := t.Logger.WithField("failcount", t.FailCount)
				if err, ok := e.(ErrorEvent); ok {
					logger.WithError(err).Warn("task failed with error")
				} else {
					logger.WithField("code", e.Code()).Warn("task failed")
				}
				// make sure the task and its children are gone
				t.Kill()
				// For every failure, wait a bit longer to restart
				// TODO: increment something on fail and decrement it
				// on success so we don't wait a long time to restart tasks
				// that don't fail often.
				pause := time.Duration(t.FailCount+1) * time.Second
				time.Sleep(pause)
				// and then start it again
				logger.Warn("restarting after failure")
				t.Start(done)
				// start created a new Listener so this one can go away
				return
			}
		}
	}
}

// streamCopy is meant to be run as a goroutine and it simply runs, copying
// src to dst until src is closed.
func streamCopy(dst io.WriteCloser, src io.ReadCloser) {
	// we want a small buffer so it keeps the output current with the input
	buf := make([]byte, 100)
	// copy until we got nothing left
	io.CopyBuffer(dst, src, buf)
}

// Start begins a new version of the task.
func (t *Task) Start(done chan struct{}) {
	// run the prerun tasks first
	for _, prerun := range t.Prerun {
		prerun.Start(done)
	}

	// if the task's path is empty, it's just a placeholder for prerun and we're done
	if t.Path == "" {
		return
	}

	t.Logger.Info("Starting")
	t.cmd = exec.Command(t.Path, t.Args...)

	if t.Stdout != nil {
		pipe, err := t.cmd.StdoutPipe()
		if err != nil {
			t.Logger.WithError(err).Error("could not construct stdout pipe")
		} else {
			go streamCopy(t.Stdout, pipe)
		}
	}

	if t.Stderr != nil {
		pipe, err := t.cmd.StderrPipe()
		if err != nil {
			t.Logger.WithError(err).Error("could not construct stderr pipe")
		} else {
			go streamCopy(t.Stderr, pipe)
		}
	}

	t.cancel = func() {
		t.Logger.WithField("pid", t.cmd.Process.Pid).Print("cancelling task")
		if t.cmd.Process == nil {
			t.Logger.Info("process was already killed")
			return
		}
		err := t.cmd.Process.Kill()
		if err != nil {
			t.Logger.WithError(err).Error("unable to kill process")
			return
		}
		state, err := t.cmd.Process.Wait()
		if err != nil {
			t.Logger.WithError(err).Error("cancel error")
			return
		}
		t.Logger.WithField("processstate", state).Info("terminated")
	}

	// start the task and wait for it to be ready
	t.Logger.Info("Running process")
	t.Logger.WithField("path", t.Path).
		WithField("args", t.Args).
		WithField("failcount", t.FailCount).
		Debug("task info")
	t.cmd.Env = t.Env
	err := t.cmd.Start()
	if err != nil {
		t.Logger.WithError(err).Error("errored on startup")
		return
	}

	// if it's a onetime task, just run it and be done
	if t.Onetime {
		t.Logger.Debug("running onetime task")
		err := t.cmd.Wait()
		if err != nil {
			t.Logger.WithError(err).Error("onetime task failed")
			panic(t.Name + " failed but mustsucceed was set")
		} else {
			t.Logger.Debug("onetime task succeeded")
		}
		return
	}

	if t.cmd != nil && t.cmd.Process != nil {
		t.Logger.WithField("pid", t.cmd.Process.Pid).Info("waiting for ready")
	}
	looptime := 50 * time.Millisecond
	loopticker := time.NewTicker(looptime)
	toolong := time.NewTimer(t.MaxStartup)
	for !IsOK(t.Ready()) {
		select {
		case <-loopticker.C:
			if t.Exited() {
				t.Logger.Error("task exited while starting up")
				return
			}
			// go check again
		case <-toolong.C:
			t.Logger.Error("took too long to start up")
			return
		}
	}
	t.Logger.WithField("pid", t.cmd.Process.Pid).Debug("task started and is ready")

	// the task is running now, start monitoring it
	// this creates t.done
	t.StartMonitors()

	// now we can start any dependent children
	t.StartChildren(done)

	// and then spin off a goroutine that will tell us if it dies
	go func() {
		err := t.cmd.Wait()
		if err != nil {
			t.Logger.WithError(err).Error("task terminated")
		}
		t.Logger.Warn("terminated")
		t.Status <- Failed
	}()

	// finally, we need to start a monitor to listen to the status channel
	go t.Listen(t.done)
}

// StartChildren starts all of the task's children
func (t *Task) StartChildren(done chan struct{}) {
	t.Logger.Info("starting children")

	// we're going to start all the children in parallel and wait until
	// the slowest of them gets going
	wg := sync.WaitGroup{}
	for _, ch := range t.Dependents {
		// copy ch
		ch := ch
		wg.Add(1)
		go func() {
			ch.Start(done)
			wg.Done()
		}()
	}
	wg.Wait()
	t.Logger.Debug("done with children")
}

// StartMonitors creates the done channel for the task
// and starts all the associated monitors with it
func (t *Task) StartMonitors() {
	t.done = make(chan struct{})
	for _, m := range t.Monitors {
		go m.Listen(t.done)
	}
	t.Logger.WithField("monitorcount", len(t.Monitors)).Debug("monitors started")
}

// StopMonitors stops all the monitors associated with this task and
// drains the Status channel
func (t *Task) StopMonitors() {
	// close the done channel
	close(t.done)
	// sleep briefly to give other tasks a chance to run
	time.Sleep(100 * time.Millisecond)
	// post a semaphore to the Status channel
	t.Status <- Drain
	// now read until we get a Drain back; since
	// we're the only one who will ever post a Drain, we know we have
	// drained the channel when we get there.
	for {
		stat := <-t.Status
		if stat == Drain {
			break
		}
	}
	t.Logger.Debug("monitors stopped")
}

// Kill ends a running task and all of its children
func (t *Task) Kill() {
	// record that we're stopping so we don't try to start things up again
	t.Logger.Warn("starting to kill process descendants")
	t.killDependents()
	if !t.Exited() {
		t.Logger.Info("shutting down")
		t.cmd.Process.Signal(syscall.SIGTERM)
		looptime := t.MaxShutdown / 64
		looptimer := time.NewTimer(looptime)
		toolong := time.NewTimer(t.MaxShutdown)
		for !t.Exited() {
			select {
			case <-looptimer.C:
				// go check again
				looptime *= 2
				looptimer.Reset(looptime)
			case <-toolong.C:
				t.Logger.Error("did not shut down nicely, killing it")
				t.cancel()
			}
		}
	}
	t.Logger.Debug("done killing")
	return
}

// killDependents ends the dependents of a running task
// but doesn't touch the task itself.
// The tasks are killed in parallel and this function only returns when
// they have all died.
func (t *Task) killDependents() {
	if len(t.Dependents) == 0 {
		return
	}
	wg := sync.WaitGroup{}
	for _, ch := range t.Dependents {
		ch := ch
		wg.Add(1)
		go func() {
			ch.Kill()
			wg.Done()
		}()
	}
	wg.Wait()
	t.Logger.Info("Done killing dependents")
	return
}

// Exited tells if a task has terminated for any reason
func (t *Task) Exited() bool {
	if t.cmd == nil {
		return true
	}
	if t.cmd.ProcessState == nil {
		if t.cmd.Process == nil {
			t.Logger.Error("process was nil")
			return true
		}
		if err := t.cmd.Process.Signal(syscall.Signal(0)); err != nil {
			t.Logger.Error("process did not respond")
			return true
		}
		return false
	}
	b := t.cmd.ProcessState.Exited()
	if b {
		t.Logger.Warn("process exited")
	}
	return b
}

// AddDependent adds a new dependent task. If this task terminates, all dependent tasks
// will also be terminated. Similarly, when this task is started, after the task
// is alive, its children will also be started.
func (t *Task) AddDependent(ch *Task) {
	t.Dependents = append(t.Dependents, ch)
}

// Destroy does an os-level kill on a task
// Use only as a last resort.
func (t *Task) Destroy() {
	for _, ch := range t.Dependents {
		ch.Destroy()
	}
	pid := t.cmd.Process.Pid
	t.Logger.WithField("pid", pid).Error("trying to force shut down")
	t.cmd.Process.Kill()
	state, err := t.cmd.Process.Wait()
	t.Logger.Printf("shutdown state %v, err %v", state, err)
}

// CollectPIDs appends this task's PID plus those of all its dependents
// to the slice passed in.
func (t *Task) CollectPIDs(pids []int) []int {
	for _, ch := range t.Dependents {
		pids = ch.CollectPIDs(pids)
	}
	if t.cmd.Process != nil {
		pid := t.cmd.Process.Pid
		pids = append(pids, pid)
	}
	return pids
}
