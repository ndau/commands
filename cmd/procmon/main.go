package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/oneiro-ndev/o11y/pkg/honeycomb"

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
		Configfile string
	}
	arg.MustParse(&args)

	var cfg Config
	var err error
	if args.Configfile != "" {
		cfg, err = Load(args.Configfile)
		if err != nil {
			panic(err)
		}
	}

	logger := cfg.BuildLogger()
	if os.Getenv("HONEYCOMB_KEY") != "" {
		logger = honeycomb.Setup(logger)
	}

	err = cfg.RunPrologue(logger)
	if err != nil {
		logger.WithError(err).Error("problems running prologue")
		panic(err)
	}

	rootTasks, err := cfg.BuildTasks(logger)
	// if we can't read the tasks we shouldn't even continue
	if err != nil {
		logger.WithError(err).Error("aborting because task file was invalid")
		panic(err)
	}

	// this is the channel we can use to shut everything down
	donech := make(chan struct{})
	// when we get SIGINT or SIGTERM, kill everything by telling
	// the root tasks to kill themselves.
	// SIGHUP does a shutdown by closing donech which
	// should basically do the same thing. However, it's
	// not entirely reliable; better to use SIGTERM if you can.
	inKillall := false
	killall := func() {
		// if they hit ctl-C a second time they must mean it,
		// just shut down now.
		if !inKillall {
			inKillall = true
			for _, t := range rootTasks {
				t.Logger.Println("shutting down on command")
				t.Kill()
			}
		}
		os.Exit(0)
	}
	shutdown := func() {
		logger.Print("shutting down by closing done channel")
		close(donech)
	}
	WatchSignals(shutdown, killall, killall)

	// start the main tasks, which will start everything else
	for _, t := range rootTasks {
		t.Logger.Print("starting root task " + t.Name)
		t.Start(donech)
	}

	// and run almost forever
	select {
	case <-donech:
		logger.Print("done channel was closed")
		// if we see this, then we were told to shutdown,
		// but we don't want to do it too early, so loop until
		// there are no root tasks left that haven't exited.
		looptime := 250 * time.Millisecond
		looptimer := time.NewTimer(looptime)
		longtime := 75 * time.Second
		toolong := time.NewTimer(longtime)
		for {
			select {
			case <-looptimer.C:
				canExit := true
				for _, t := range rootTasks {
					if !t.Exited() {
						canExit = false
					}
				}
				if canExit {
					os.Exit(0)
				}
				logger.Printf("waiting for all root tasks to die...")
				looptime *= 2
				looptimer.Reset(looptime)
			case <-toolong.C:
				logger.Error("requested shutdown did not complete within " + longtime.String())
				for _, t := range rootTasks {
					t.Destroy()
				}
				time.Sleep(1 * time.Second)
				os.Exit(1)
			}
		}
	}
}
