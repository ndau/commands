package main

import (
	"time"
)

// Monitor defines a listener that is calling a Test function
// periodically. The test function returns an Event and that event is sent
// on the Status channel.
// It is expected that the Listen function is called as a goroutine.
// If the done channel is closed, the Monitor terminates.
// Monitors are designed to be easily aggregated and wrapped like middleware.
type Monitor struct {
	D      time.Duration
	Status chan Eventer
	Test   func() Eventer
}

// Listener is the interface for the Listen method. It expects to be called as a goroutine.
type Listener interface {
	Listen(done chan struct{})
}

// asserts that Monitor is in fact a Listener
var _ Listener = (*Monitor)(nil)

// NewMonitor returns a new monitor object
func NewMonitor(status chan Eventer, d time.Duration, test func() Eventer) *Monitor {
	return &Monitor{
		D:      d,
		Status: status,
		Test:   test,
	}
}

// Listen implements listener, and should be called as a goroutine.
func (m *Monitor) Listen(done chan struct{}) {
	for {
		select {
		case <-done:
			return
		case <-time.After(m.D):
			m.Status <- m.Test()
		}
	}
}

// RetryMonitor wraps a monitor and will only fail after receiving a number of
// successive failures that exceeds the Retries value.
// Note that this won't reliably detect an intermittent problem. For that,
// consider a LeakyBucketMonitor.
type RetryMonitor struct {
	M           *Monitor
	Retries     int
	test        func() Eventer
	status      chan Eventer
	childStatus chan Eventer
	failCount   int
}

// asserts that RetryMonitor is in fact a Listener
var _ Listener = (*RetryMonitor)(nil)

// NewRetryMonitor wraps an existing monitor with retry logic.
func NewRetryMonitor(m *Monitor, retries int) *RetryMonitor {
	rm := &RetryMonitor{
		M:           m,
		Retries:     retries,
		status:      m.Status,
		childStatus: make(chan Eventer),
		test:        m.Test,
	}
	// wrap the test function with a closure that resets failCount if
	// the system worked
	m.Test = func() Eventer {
		e := rm.test()
		if IsOK(e) {
			rm.failCount = 0
		}
		return e
	}
	m.Status = rm.childStatus
	return rm
}

// Listen implements Listener for RetryMonitor.
// It should be called as a goroutine.
func (m *RetryMonitor) Listen(done chan struct{}) {
	for {
		select {
		case <-done:
			return
		case e := <-m.childStatus:
			if IsFailed(e) {
				m.failCount++
				if m.failCount > m.Retries {
					m.status <- e
				} else {
					m.status <- Failing
				}
			} else {
				m.status <- e
			}
		}
	}
}
