package main

import (
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

// Task is a restartable process; it can be monitored and
// restarted.
//
// Each task publishes and listens to a Status channel. Any sender can
// send a Stop event on that channel.
// Other events are also permitted, and could be used to communicate
// between tasks or between monitors and tasks. However, the Stop
// event takes precedence and forces shutdown of the task.
//
// Each task also publishes a Stopped channel.
// All of the task's monitors and dependents should listen on the
// Stopped channel, and when that channel is closed,
// they should also terminate themselves (and propagate that
// information to their listeners in turn).
//
// It is safe for multiple monitors to send Stop events.
//
// Each task also listens to the Stopped channel of its dependents.
// If any of a task's dependents stop without having been
// stopped by the parent, then the dependent will be restarted.
//
// At the base level, tasks without a parent in the config are
// actually children of a master task. If the master task is
// shut down it will not be automatically restarted.
//
// A bit of pseudocode:
//
// Start (parentstop):
//     run prefix tasks
//     run task
//     wait for task to start
//     create stop channel
//     run master monitor(parentstop, stop)
//     run task exit monitor (stop)
//     run all other monitors (stop)
//     run children (stop)
//     run
//
// master monitor:
//     if status gets Stop message
//         close taskstop
//         terminate
//     if parentstop closed, close taskstop, terminate
//
// stopMonitor:
//     if t.Stopped is closed:
//         kill task
//         terminate
//
// task exit monitor(status, stop):
//     if task exits send stop on status
//
// behavior monitors:
//     if task fails send stop on status
//     if stop closed terminate
//
// child monitor:
//     if child's stop is closed:
//         record it
//         wait for fallback time
//         call child.Start() only if parent task is not Stopped
type Task struct {
	Name         string
	Path         string
	Args         []string
	Env          []string
	Onetime      bool
	MaxShutdown  time.Duration
	MaxStartup   time.Duration
	Status       chan Eventer
	Stopped      chan struct{}
	Ready        func() Eventer
	Stdout       io.WriteCloser
	Stderr       io.WriteCloser
	Logger       logrus.FieldLogger
	FailCount    int
	RestartDelay time.Duration
	Monitors     []*FailMonitor
	Prerun       []*Task
	Dependents   []*Task

	cmd   *exec.Cmd
	dying bool
}

// NewTask creates a Task (but does not start it)
// The default Ready() function simply returns true
func NewTask(name string, path string, args ...string) *Task {
	return &Task{
		Name:         name,
		Path:         path,
		Args:         args,
		MaxStartup:   10 * time.Second,
		MaxShutdown:  5 * time.Second,
		Ready:        func() Eventer { return OK },
		Monitors:     make([]*FailMonitor, 0),
		RestartDelay: time.Second,
	}
}

// The masterMonitor is listening to the task's
// Status channel and if it receives a stop message,
// is the only place the t.Stopped channel is closed.
// If the parent task's Stopped channel is closed,
// it closes this task's Stopped channel also.
func (t *Task) masterMonitor(parentstop chan struct{}) {
	for {
		select {
		case <-parentstop:
			t.Logger.Info("parent task stopped; shutting down")
			close(t.Stopped)
			time.Sleep(50 * time.Millisecond)
			return
		case e := <-t.Status:
			t.Logger.WithField("status", e).Debug("event")
			if e == Stop {
				t.Logger.Warn("received Stop message; shutting down")
				close(t.Stopped)
				time.Sleep(50 * time.Millisecond)
				return
			}
		}
	}
}

// The exitMonitor waits for the task itself to exit
// and then sends the Stop message on its Status channel.
// It should be launched as a goroutine and will not
// terminate until the task does.
func (t *Task) exitMonitor() {
	// make a copy of the Status chan so
	// if it gets recreated we don't send a message on
	// the new one by mistake
	status := t.Status
	err := t.cmd.Wait()
	if err != nil {
		t.Logger.WithError(err).Error("task terminated")
	} else {
		t.Logger.Warn("terminated")
	}
	status <- Stop
}

// The childMonitor is given a child task;
// If the child task's Stopped channel is closed
// before the current task, childMonitor:
// * waits the child's startDelay amount of time
// * increases the child's startDelay (to try to slow down
//   a task that's flapping)
// * call child.Start()
// If the current task's Stopped channel is closed, this
// monitor terminates
// It also reduces restart time if a task is well-behaved after having failed.
func (t *Task) childMonitor(child *Task) {
	// allow the default restart delay to be overridden in the environment
	var defaultRestartDelay = 10 * time.Second
	if s := os.Getenv("DEFAULT_RESTART_DELAY"); s != "" {
		if drd, err := time.ParseDuration(s); err == nil {
			defaultRestartDelay = drd
		}
	}
	for {
		select {
		case <-time.After(t.RestartDelay):
			// we double the RestartDelay every time the task restarts,
			// and we reduce it asymptotically to the default value
			// as the task lives longer without restarting
			t.RestartDelay = defaultRestartDelay +
				((t.RestartDelay-defaultRestartDelay)*9)/10
		case <-t.Stopped:
			return
		case <-child.Stopped:
			// we want to delay for the sleep time but
			// we don't want to miss it if our task is stopped
			// because we don't want to restart the child
			// if its parent is restarting
			select {
			case <-time.After(child.RestartDelay):
				child.FailCount++
				child.RestartDelay *= 2
				// this will replace the child's Stopped channel
				t.Logger.WithField("child", child.Name).
					WithField("delay", child.RestartDelay).
					WithField("failcount", child.FailCount).
					Debugf("childmonitor restarting child")
				child.Start(t.Stopped)
			case <-t.Stopped:
				t.Logger.WithField("child", child.Name).
					Debugf("childmonitor detected child stop but parent also stopped")
				return
			}
		}
	}
}

// stopMonitor is the one that listens to the Stopped channel
// and shuts the task down when it's stopped
func (t *Task) stopMonitor() {
	select {
	case <-t.Stopped:
		t.Logger.Info("Stopped channel closed; killing task")
		t.Kill()
		return
	}
}

func (t *Task) setOutputStreams() {
	// streamCopy is meant to be run as a goroutine and it simply runs, copying
	// src to dst until src is closed. It's used for logging.
	streamCopy := func(dst io.WriteCloser, src io.ReadCloser) {
		// we want a small buffer so it keeps the output current with the input
		buf := make([]byte, 100)
		// copy until we got nothing left
		io.CopyBuffer(dst, src, buf)
	}

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
}

// Start begins a new version of the task.
func (t *Task) Start(parentstop chan struct{}) {
	// run the prerun tasks first
	for _, prerun := range t.Prerun {
		prerun.Start(parentstop)
	}

	// if the task's path is empty, it's just a placeholder for prerun and we're done
	if t.Path == "" {
		return
	}

	t.Logger.Info("Starting")
	t.cmd = exec.Command(t.Path, t.Args...)
	t.setOutputStreams()

	// start the task and wait for it to be ready
	t.Logger.Info("Running process")
	t.Logger.WithField("path", t.Path).
		WithField("args", t.Args).
		WithField("failcount", t.FailCount).
		Debug("task info")
	t.cmd.Env = t.Env
	t.dying = false

	// if it's a onetime task, just run it and be done
	if t.Onetime {
		t.Logger.Debug("running onetime task")
		err := t.cmd.Run()
		if err != nil {
			t.Logger.WithError(err).Error("onetime task failed")
			panic(t.Name + " failed but mustsucceed was set")
		} else {
			t.Logger.Debug("onetime task succeeded")
		}
		return
	}

	err := t.cmd.Start()
	if err != nil {
		t.Logger.WithError(err).Error("errored on startup")
		return
	}

	if t.cmd != nil && t.cmd.Process != nil {
		t.Logger.WithField("pid", t.cmd.Process.Pid).Info("waiting for ready")
	}
	looptime := 50 * time.Millisecond
	loopticker := time.NewTicker(looptime)
	toolong := time.NewTimer(t.MaxStartup)
	for t.Ready() != OK {
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

	// now we need the Status channel
	t.Status = make(chan Eventer, 1)
	// make a Stopped channel
	t.Stopped = make(chan struct{})

	// run the masterMonitor
	go t.masterMonitor(parentstop)
	// spin off a goroutine that will tell us if it dies
	go t.exitMonitor()
	// finally, we need to start a monitor to listen to the status channel
	go t.stopMonitor()
	// the task is running now, start all the behavior monitors
	t.startBehaviorMonitors()
	// now we can start any dependent children
	t.StartChildren()
}

// StartChildren starts all of the task's children
func (t *Task) StartChildren() {
	t.Logger.Info("starting children")

	// we're going to start all the children in parallel and wait until
	// the slowest of them gets going
	wg := sync.WaitGroup{}
	for _, ch := range t.Dependents {
		// copy ch
		ch := ch
		wg.Add(1)
		go func() {
			ch.Start(t.Stopped)
			// start a child monitor to keep it running
			go t.childMonitor(ch)
			wg.Done()
		}()
	}
	wg.Wait()
	t.Logger.Debug("done with children")
}

// startBehaviorMonitors starts all the monitors
// that watch the task for bad behavior
func (t *Task) startBehaviorMonitors() {
	for _, m := range t.Monitors {
		go m.Listen(t.Stopped)
	}
	t.Logger.WithField("monitorcount", len(t.Monitors)).Debug("behavior monitors started")
}

// waitForShutdown assumes the task has already begun shutdown and
// that we just have to wait for it.
func (t *Task) waitForShutdown() {
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
			t.Destroy()
		}
	}
	looptimer.Stop()
	toolong.Stop()
}

// Kill ends a running task and all of its children
// In a properly functioning system, child tasks should
// have already been notified, but this will ensure
// that they shut down and that we wait for it
func (t *Task) Kill() {
	// if we're already shutting down, just wait for it
	if t.dying {
		t.waitForShutdown()
		return
	}

	// record that we're stopping
	t.dying = true
	t.Logger.Warn("starting to kill process")
	t.killDependents()
	if !t.Exited() {
		t.Logger.Info("shutting down")
		t.cmd.Process.Signal(syscall.SIGTERM)
		t.waitForShutdown()
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

// Destroy does an os-level kill on a task and all its dependents
// Use only as a last resort.
func (t *Task) Destroy() {
	for _, ch := range t.Dependents {
		ch.Destroy()
	}

	if t.cmd == nil || t.cmd.Process == nil {
		t.Logger.Error("destroy called with no process to kill")
		return
	}
	t.Logger.WithField("pid", t.cmd.Process.Pid).Print("cancelling task")
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

// CollectPIDs appends this task's PID plus those of all its dependents
// to the slice passed in.
func (t *Task) CollectPIDs(pids []int) []int {
	for _, ch := range t.Dependents {
		pids = ch.CollectPIDs(pids)
	}
	if t.cmd != nil && t.cmd.Process != nil {
		pid := t.cmd.Process.Pid
		pids = append(pids, pid)
	}
	return pids
}
