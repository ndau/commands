package main

import (
	"fmt"
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

	failch := make(chan *Task)

	t1 := NewTask("demo parent",
		"/Users/kentquirk/go/src/github.com/oneiro-ndev/rest/cmd/demo/demo",
		"--port=9999",
	)
	t2 := NewTask("demo child",
		"/Users/kentquirk/go/src/github.com/oneiro-ndev/rest/cmd/demo/demo",
		"--port=9998",
	)
	t1.AddDependent(t2)
	go t1.Start(failch)
	killall := func() {
		t1.Kill()
		os.Exit(0)
	}
	WatchSignals(nil, killall, killall)

	donech := make(chan struct{})
	defer close(donech)
	m1 := NewSimpleHTTPMonitor(t1, "http://localhost:9999/health", 3*time.Second)
	m2 := NewSimpleHTTPMonitor(t2, "http://localhost:9998/health", 3*time.Second)
	go m1.Listen(failch, donech)
	go m2.Listen(failch, donech)

	ignores := TaskSet{}
	for {
		select {
		case t := <-failch:
			if ignores.Has(t) {
				fmt.Printf("Ignoring deliberate shutdown of %s\n", t.Name)
				ignores.Delete(t)
				continue
			}
			fmt.Printf("Task %s failed, restarting.\n", t.Name)
			// kill the task and all its dependents in reverse order
			ignores.Add(t.Kill()...)
			t.Start(failch)
		}
		fmt.Println("Looping")
	}
}
