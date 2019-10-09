package main

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

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

// FailMonitor wraps a monitor and only sends Failure events; it sends one
// when it receives one from its wrapped monitor.
type FailMonitor struct {
	Child  *Monitor
	Status chan Eventer
}

// asserts that FailMonitor is in fact a Listener
var _ Listener = (*FailMonitor)(nil)

// NewFailMonitor creates a FailMonitor from a Monitor
func NewFailMonitor(child *Monitor) *FailMonitor {
	fm := FailMonitor{
		Child:  child,
		Status: child.Status,
	}
	child.Status = make(chan Eventer)
	return &fm
}

// Listen implements listener, and should be called as a goroutine.
// It returns Stop when its child monitor returns Failed or Stop,
// otherwise it swallows the event.
func (m *FailMonitor) Listen(done chan struct{}) {
	go m.Child.Listen(done)
	for {
		select {
		case <-done:
			return
		case stat := <-m.Child.Status:
			if stat == Failed || stat == Stop {
				m.Status <- Stop
			}
		}
	}
}

// RetryMonitor wraps a monitor and will only fail after receiving a number of
// successive failures that exceeds the Retries value.
// Note that this won't reliably detect an intermittent problem. For that,
// consider a LeakyBucketMonitor.
type RetryMonitor struct {
	Child       *Monitor
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
		Child:       m,
		Retries:     retries,
		status:      m.Status,
		childStatus: make(chan Eventer),
		test:        m.Test,
	}
	// wrap the test function with a closure that resets failCount if
	// the system worked
	m.Test = func() Eventer {
		e := rm.test()
		if e == OK {
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
	go m.Child.Listen(done)
	for {
		select {
		case <-done:
			return
		case e := <-m.childStatus:
			if e == Failed || e == Stop {
				m.failCount++
				if m.failCount > m.Retries {
					m.status <- Stop
				} else {
					m.status <- Failing
				}
			} else {
				m.status <- e
			}
		}
	}
}
