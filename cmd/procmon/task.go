package main

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// Task is a restartable program; it can be monitored and
// restarted if a monitor fails
type Task struct {
	Name        string
	Path        string
	Args        []string
	MaxShutdown time.Duration

	mutex      sync.Mutex
	Dependents []*Task

	cmd    *exec.Cmd
	cancel context.CancelFunc
}

// NewTask creates a Task (but does not start it)
func NewTask(name string, path string, args ...string) *Task {
	return &Task{
		Name:        name,
		Path:        path,
		Args:        args,
		MaxShutdown: 5 * time.Second,
	}
}

// Kill ends a running task
func (t *Task) Kill() []*Task {
	killed := t.KillDependents()
	fmt.Printf("Killing %s\n", t.Name)
	t.cancel()
	return append(killed, t)
}

// KillDependents ends the dependents of a running task
// but doesn't touch the task itself.
func (t *Task) KillDependents() []*Task {
	t.mutex.Lock()
	for _, ch := range t.Dependents {
		ch.Kill()
	}
	t.mutex.Unlock()
	return t.Dependents
}

// Start begins a new version of the task.
// It completes only when the task exits; before doing so it
// sends the task back on the terminated channel.
func (t *Task) Start(terminated chan *Task) {
	fmt.Printf("Starting %s\n", t.Name)
	ctx, cancel := context.WithCancel(context.Background())

	t.cmd = exec.CommandContext(ctx, t.Path, t.Args...)
	t.cancel = cancel
	t.cmd.Start()

	// now start any dependent children
	t.mutex.Lock()
	for _, ch := range t.Dependents {
		go ch.Start(terminated)
	}
	t.mutex.Unlock()

	// and then hang here
	t.cmd.Wait()
	terminated <- t
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
	t.Dependents = append(t.Dependents, ch)
	t.mutex.Unlock()
}
