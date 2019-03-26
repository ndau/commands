package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oneiro-ndev/o11y/pkg/honeycomb"
	"github.com/sirupsen/logrus"

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

	logger := honeycomb.Setup(&logrus.Logger{
		Out:       os.Stderr,
		Formatter: new(logrus.JSONFormatter),
		Level:     logrus.InfoLevel,
	})

	// set up the main task
	t1 := NewTask("demo parent",
		"/Users/kentquirk/go/src/github.com/oneiro-ndev/rest/cmd/demo/demo",
		"--port=9999",
	)
	// give it a way to tell if it's ready
	t1.Ready = HTTPPinger("http://localhost:9999/health", 100*time.Millisecond)
	// and send its logs to a file
	f, _ := os.Create("t1.log")
	t1.Stdout = f
	t1.Stderr = f
	t1.Logger = logger.WithField("task", t1.Name)
	t1.Logger.Println("Starting task")

	// do the same for the child task
	t2 := NewTask("demo child",
		"/Users/kentquirk/go/src/github.com/oneiro-ndev/rest/cmd/demo/demo",
		"--port=9998",
	)
	t2.Logger = logger.WithField("task", t2.Name)
	t2.Logger.Println("Starting task")

	// and here's how it gets ready
	t1.Ready = HTTPPinger("http://localhost:9999/health", 100*time.Millisecond)
	// and set the parent/child relationship
	t1.AddDependent(t2)

	// Now we have a health check monitor for the main task
	m1 := NewMonitor(t1.Status, 3*time.Second, HTTPPinger("http://localhost:9999/health", 2*time.Second))
	// and health check for the child with a shorter timeout
	m2 := NewMonitor(t2.Status, 1*time.Second, HTTPPinger("http://localhost:9998/health", 1*time.Second))
	// but some tolerance for occasional failure
	rm2 := NewRetryMonitor(m2, 3)

	// this is the channel we can use to shut everything down
	donech := make(chan struct{})
	// defer close(donech)

	// start the listeners for the monitors
	go m1.Listen(donech)
	go rm2.Listen(donech)

	// when we get SIGINT or SIGTERM, kill everything
	// SIGHUP is currently ignored
	killall := func() {
		t1.Kill()
		os.Exit(0)
	}
	shutdown := func() {
		close(donech)
	}
	WatchSignals(shutdown, killall, killall)

	// start the main task, which will start everything else
	t1.Start(donech)
	// and run almost forever
	select {
	case <-donech:
		t1.Kill()
	}
}
