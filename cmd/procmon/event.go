package main

import "fmt"

// Event is the type that indicates the status of a Task
type Event int

// assert that Event satisfies Eventer
var _ Eventer = (*Event)(nil)

// These constants are used to indicate possible event types
const (
	Started  Event = iota
	OK       Event = iota
	Failing  Event = iota
	Failed   Event = iota
	Stopping Event = iota
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
	return fmt.Sprintf("ErrorEvent %d - %s", e.Evt, e.Error())
}

// IsFailed returns true if the event is a Failed event
func IsFailed(e Eventer) bool {
	return e.Code() == Failed
}

// IsOK returns true if the event is an OK event
func IsOK(e Eventer) bool {
	return e.Code() == OK
}
