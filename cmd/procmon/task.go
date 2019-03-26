package main

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// Task is a restartable process; it can be monitored and
// restarted.
// By default, a task is monitored simply as a running process. If the
// process terminates for any reason, the task's status channel
// is notified.
// The task can also be given a set of monitors that can also
// watch the task for incorrect performance; they also communicate
// on the status channel.
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
	MaxShutdown time.Duration
	Status      chan Eventer
	Ready       func() Eventer
	Stdout      io.WriteCloser
	Stderr      io.WriteCloser

	mutex      sync.Mutex
	Stopping   bool
	Dependents []*Task

	cmd    *exec.Cmd
	cancel context.CancelFunc
}

// NewTask creates a Task (but does not start it)
// The default Ready() function simply returns true
func NewTask(name string, path string, args ...string) *Task {
	return &Task{
		Name:        name,
		Path:        path,
		Args:        args,
		MaxShutdown: 5 * time.Second,
		Status:      make(chan Eventer),
		Ready:       func() Eventer { return OK },
	}
}

// Listen implements Listener, and should be called as a goroutine.
func (t *Task) Listen(done chan struct{}) {
	for {
		select {
		case e := <-t.Status:
			if IsFailed(e) {
				if t.Stopping {
					fmt.Printf("Ignoring deliberate shutdown of %s\n", t.Name)
					continue
				}
				fmt.Printf("Task %s failed, restarting.\n", t.Name)
				// make sure the task and its children are gone
				t.Kill()
				// and then start it again
				t.Start(done)
				// start created a new Listener so this one can go away
				return
			}
			fmt.Println("Looping")
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
	fmt.Printf("Starting %s\n", t.Name)
	ctx, cancel := context.WithCancel(context.Background())

	t.cmd = exec.CommandContext(ctx, t.Path, t.Args...)

	if t.Stdout != nil {
		pipe, err := t.cmd.StdoutPipe()
		if err != nil {
			panic(err) // replace with log
		}
		go streamCopy(t.Stdout, pipe)
	}

	if t.Stderr != nil {
		pipe, err := t.cmd.StderrPipe()
		if err != nil {
			panic(err) // replace with log
		}
		go streamCopy(t.Stderr, pipe)
	}

	t.cancel = cancel

	// clear the stopping flag
	fmt.Printf("Clearing demo flag %s\n", t.Name)
	t.mutex.Lock()
	t.Stopping = false
	t.mutex.Unlock()
	fmt.Printf("Finished clearing demo flag %s\n", t.Name)

	// start the task and wait for it to be ready
	fmt.Printf("Running process for %s\n", t.Name)
	t.cmd.Start()
	for !IsOK(t.Ready()) {
		select {
		case <-time.After(50 * time.Millisecond):
			// go check again
		case <-time.After(30 * time.Second):
			fmt.Printf("%s took too long to start up.\n", t.Name)
			return
		}
	}

	// now we can start any dependent children
	// we want the unlock to happen before the Wait() so this is in a func
	func() {
		fmt.Printf("Starting children %s\n", t.Name)
		t.mutex.Lock()
		defer t.mutex.Unlock()

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
	}()
	fmt.Printf("Done with children %s\n", t.Name)

	// and then spin off a goroutine that will tell us if it dies
	go func() {
		t.cmd.Wait()
		t.Status <- Failed
	}()

	// now we need to monitor for that death
	go t.Listen(done)
}

// Kill ends a running task and all of its children
func (t *Task) Kill() {
	// record that we're stopping so we don't try to start things up again
	fmt.Printf("Starting to kill %s\n", t.Name)
	func() {
		fmt.Printf("Setting stopping flag for %s\n", t.Name)
		t.mutex.Lock()
		defer t.mutex.Unlock()
		t.Stopping = true
	}()
	fmt.Printf("Finished setting stopping flag for %s\n", t.Name)

	t.killDependents()
	if !t.Exited() {
		fmt.Printf("Shutting down %s\n", t.Name)
		t.cmd.Process.Signal(syscall.SIGTERM)
		for !t.Exited() {
			select {
			case <-time.After(t.MaxShutdown / 100):
				// go check again
			case <-time.After(t.MaxShutdown):
				fmt.Printf("%s did not shut down -- killing it.\n", t.Name)
				t.cancel()
			}
		}
	}
	fmt.Printf("Done Killing %s\n", t.Name)
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
	fmt.Printf("Killing dependents of %s", t.Name)
	t.mutex.Lock()
	defer t.mutex.Unlock()
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
	fmt.Printf("Done killing dependents of %s", t.Name)
	return
}

// Exited tells if a task has terminated for any reason
func (t *Task) Exited() bool {
	// fmt.Printf("Checking exit for %s: %#v\n", t.Name, t.cmd)
	if t.cmd.ProcessState == nil {
		if t.cmd.Process == nil {
			fmt.Printf("Process %s was nil\n", t.Name)
			return true
		}
		if err := t.cmd.Process.Signal(syscall.Signal(0)); err != nil {
			fmt.Printf("Process %s did not respond\n", t.Name)
			return true
		}
		return false
	}
	b := t.cmd.ProcessState.Exited()
	if b {
		fmt.Printf("Process %s exited\n", t.Name)
	}
	return b
}

// AddDependent adds a new dependent task. If this task terminates, all dependent tasks
// will also be terminated. Similarly, when this task is started, after the task
// is alive, its children will also be started.
func (t *Task) AddDependent(ch *Task) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.Dependents = append(t.Dependents, ch)
}
