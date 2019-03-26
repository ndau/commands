package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	arg "github.com/alexflint/go-arg"
)

// WatchSignals can set up functions to call on various operating system signals.
func WatchSignals(fhup, fint, fterm func()) {
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
		for {
			sig := <-sigchan
			switch sig {
			case syscall.SIGHUP:
				if fhup != nil {
					fhup()
				}
			case syscall.SIGINT:
				if fint != nil {
					fint()
				}
			case syscall.SIGTERM:
				if fterm != nil {
					fterm()
				}
				os.Exit(0)
			}
		}
	}()
}

func main() {
	var args struct {
		Test string
	}
	arg.MustParse(&args)

	t1 := NewTask("demo parent",
		"/Users/kentquirk/go/src/github.com/oneiro-ndev/rest/cmd/demo/demo",
		"--port=9999",
	)
	t2 := NewTask("demo child",
		"/Users/kentquirk/go/src/github.com/oneiro-ndev/rest/cmd/demo/demo",
		"--port=9998",
	)
	t1.AddDependent(t2)
	m1 := NewMonitor(t1.Status, 3*time.Second, HTTPPinger("http://localhost:9999/health", 2*time.Second))
	m2 := NewMonitor(t2.Status, 1*time.Second, HTTPPinger("http://localhost:9998/health", 1*time.Second))
	rm2 := NewRetryMonitor(m2, 3)

	donech := make(chan struct{})
	defer close(donech)

	go m1.Listen(donech)
	go rm2.Listen(donech)

	killall := func() {
		t1.Kill()
		os.Exit(0)
	}
	WatchSignals(nil, killall, killall)

	t1.Start(donech)
	select {
	case <-donech:
	}
}
