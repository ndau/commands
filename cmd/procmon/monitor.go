package main

import (
	"net/http"
	"time"
)

// Monitor defines a listener that is watching a task, and if the
// task fails to respond (however the monitor defines that)
// will send the task on the failed channel.
// It is expected that the Listen function is called as a goroutine.
// If the done channel is closed, the Monitor terminates.
// Monitors are designed to be easily aggregated and wrapped like middleware.
type Monitor interface {
	Listen(failed chan *Task, done chan struct{})
}

// A SimpleHTTPMonitor pings a URL every duration d, and expects a 200 response
// It sends the task on the failed channel if it doesn't get one.
type SimpleHTTPMonitor struct {
	T *Task
	U string
	D time.Duration
}

// NewSimpleHTTPMonitor constructs what it says
func NewSimpleHTTPMonitor(t *Task, u string, d time.Duration) Monitor {
	return &SimpleHTTPMonitor{
		T: t,
		U: u,
		D: d,
	}
}

// Listen implements Monitor.
// It should be called as a goroutine.
func (m *SimpleHTTPMonitor) Listen(failed chan *Task, done chan struct{}) {
	client := http.Client{Timeout: time.Duration(m.D)}
	req, err := http.NewRequest("GET", m.U, nil)
	if err != nil {
		panic(err)
	}
	for {
		select {
		case <-done:
			return
		case <-time.After(m.D):
			resp, err := client.Do(req)
			if err != nil {
				failed <- m.T
				continue
			}
			resp.Body.Close()
			if resp.StatusCode > 299 || resp.StatusCode < 200 {
				failed <- m.T
			}
		}
	}
}

// AlivenessMonitor checks the task every duration d, and expects it to be found
// still running.
// It sends the task on the failed channel if it has exited.
type AlivenessMonitor struct {
	T *Task
	D time.Duration
}

// NewAlivenessMonitor constructs what it says
func NewAlivenessMonitor(t *Task, d time.Duration) Monitor {
	return &AlivenessMonitor{
		T: t,
		D: d,
	}
}

// Listen implements Monitor.
// It should be called as a goroutine.
func (m *AlivenessMonitor) Listen(failed chan *Task, done chan struct{}) {
	for {
		select {
		case <-done:
			return
		case <-time.After(m.D):
			if m.T.Exited() {
				failed <- m.T
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
