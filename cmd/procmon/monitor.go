package main

import (
	"fmt"
	"net/http"
	"time"
)

// Monitor defines a listener that is watching a task by calling a Test function
// periodically. If the Test function fails, the Monitor will
// send the task on the failed channel.
// It is expected that the Listen function is called as a goroutine.
// If the done channel is closed, the Monitor terminates.
// Monitors are designed to be easily aggregated and wrapped like middleware.
type Monitor struct {
	T      *Task
	D      time.Duration
	Failed chan *Task
	Test   func() bool
}

// Listener is the interface for the Listen method. It expects to be called as a goroutine.
type Listener interface {
	Listen(done chan struct{})
}

// asserts that Monitor is in fact a Listener
var _ Listener = (*Monitor)(nil)

// NewMonitor returns a new monitor object
func NewMonitor(t *Task, d time.Duration, failed chan *Task, test func() bool) *Monitor {
	return &Monitor{
		T:      t,
		D:      d,
		Failed: failed,
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
			if !m.Test() {
				x := m.T
				m.Failed <- x
			}
		}
	}
}

// HTTPPinger returns a function compatible with the Monitor's Test
// parameter that pings an HTTP address with a timeout.
func HTTPPinger(u string, timeout time.Duration) func() bool {
	client := http.Client{Timeout: time.Duration(timeout)}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		panic(err)
	}
	return func() bool {
		resp, err := client.Do(req)
		if err != nil {
			return false
		}
		resp.Body.Close()
		if resp.StatusCode > 299 || resp.StatusCode < 200 {
			return false
		}
		return true
	}
}

// RetryMonitor wraps a monitor and will only fail after receiving a number of
// successive failures that exceeds the Retries value.
// Note that this won't reliably detect an intermittent problem. For that,
// consider a LeakyBucketMonitor.
type RetryMonitor struct {
	M           *Monitor
	Retries     int
	test        func() bool
	failed      chan *Task
	childFailed chan *Task
	failCount   int
}

// asserts that RetryMonitor is in fact a Listener
var _ Listener = (*RetryMonitor)(nil)

// NewRetryMonitor wraps an existing monitor with retry logic.
func NewRetryMonitor(m *Monitor, retries int) *RetryMonitor {
	rm := &RetryMonitor{
		M:           m,
		Retries:     retries,
		failed:      m.Failed,
		childFailed: make(chan *Task),
		test:        m.Test,
	}
	// wrap the test function with a closure that resets failCount if
	// the system worked
	m.Test = func() bool {
		t := rm.test()
		if !t {
			rm.failCount = 0
		}
		return t
	}
	m.Failed = rm.childFailed
	return rm
}

// Listen implements Listener for RetryMonitor.
// It should be called as a goroutine.
func (m *RetryMonitor) Listen(done chan struct{}) {
	for {
		select {
		case <-done:
			return
		case t := <-m.childFailed:
			m.failCount++
			fmt.Printf("retry monitor failcount = %d\n", m.failCount)
			if m.failCount > m.Retries {
				m.failed <- t
			}
		}
	}
}

// type MultiMonitor struct {
// 	Fail     chan<- *Task
// 	Done     <-chan struct{}
// 	Monitors []Monitor
// }

// func NewMultiMonitor(ms ...Monitor) Monitor {
// 	return &MultiMonitor{
// 		Fail:     fail,
// 		Done:     done,
// 		Monitors: ms,
// 	}
// }

// // Listen implements Monitor.
// // It should be called as a goroutine.
// func (m *MultiMonitor) Listen(failed chan *Task, done chan struct{}) {
// 	failch := make(chan<- *Task)
// 	donech := make(<-chan struct{})
// 	for _, mon := range m.Monitors {
// 		go mon.Listen(t, failch, donech)
// 	}
// 	for {
// 		select {
// 		case <-done:
// 			close(failch)
// 			return
// 		case t <- failch:
// 			failed <- t
// 		}
// 	}
// }
