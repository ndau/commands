package main

import (
	"fmt"
	"os/exec"
)

// Event is the type that indicates the status of a Task
type Event int

// assert that Event satisfies Eventer
var _ Eventer = (*Event)(nil)

// These constants are used to indicate possible event types
const (
	OK      Event = iota
	Stop    Event = iota
	Failing Event = iota
	Failed  Event = iota
)

// Eventer is an interface for an object that is carrying an event
type Eventer interface {
	Code() Event
}

// Code implements Eventer for an Event object
func (e Event) Code() Event {
	return e
}

// ErrorEvent is an Event that can carry an error as well
type ErrorEvent struct {
	Evt Event
	Err error
}

// assert that ErrorEvent satisfies Eventer
var _ Eventer = (*ErrorEvent)(nil)

// and also the error interface
var _ error = (*ErrorEvent)(nil)

// NewErrorEvent constructs an ErrorEvent
func NewErrorEvent(evt Event, err error) Eventer {
	return ErrorEvent{
		Evt: evt,
		Err: err,
	}
}

// Code implements Eventer for an ErrorEvent object
func (e ErrorEvent) Code() Event {
	return e.Evt
}

// Error() implements the error interface for ErrorEvent
func (e ErrorEvent) Error() string {
	return fmt.Sprintf("ErrorEvent %d - %s", e.Evt, e.Err.Error())
}

// TerminateEvent is an event constructed when a Task exits
//
// Construct it by wrapping exec.Wait
type TerminateEvent struct {
	Err error
}

// assert that ErrorEvent satisfies Eventer
var _ Eventer = (*TerminateEvent)(nil)

// and also the error interface
var _ error = (*TerminateEvent)(nil)

// Code implements Eventer for TerminateEvent. It always returns Stop.
func (e TerminateEvent) Code() Event {
	return Stop
}

// Error implementes error for TerminateEvent
func (e TerminateEvent) Error() string {
	if e.Err == nil {
		return "<nil>"
	}
	return e.Err.Error()
}

// ExitCode returns the exit code of the Task
//
// If the task was terminated by a signal, returns -1
// If the task's exit code couldn't be determined, returns -2
// Otherwise, returns the task's exit code
func (e TerminateEvent) ExitCode() int {
	if e.Err == nil {
		return 0
	}
	eerr, ok := e.Err.(*exec.ExitError)
	if !ok {
		// we don't know what happened, but it wasn't a normal exit error
		return -2
	}
	return eerr.ExitCode()
}
