package main

import (
	"fmt"
	"os"
	"os/signal"
	"sort"
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
		Configfile string `help:"the name of the .toml file to load"`
		NoCheck    bool   `help:"set this to disable envvar checking"`
	}
	arg.MustParse(&args)

	var cfg Config
	var err error
	if args.Configfile != "" {
		cfg, err = Load(args.Configfile, args.NoCheck)
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}
	}

	logger := cfg.BuildLogger()
	if os.Getenv("HONEYCOMB_KEY") != "" {
		logger = honeycomb.Setup(logger)
	}

	err = cfg.RunPrologue(logger)
	if err != nil {
		logger.WithError(err).Fatal("problems running prologue")
	}

	mainTasks, err := cfg.BuildTasks(logger)
	// if we can't read the tasks we shouldn't even continue
	if err != nil {
		logger.WithError(err).Fatal("aborting because task file was invalid")
	}

	// now build a special task to act as the parent of the root tasks
	root := NewTask("root", "")
	root.Logger = logger
	root.Stopped = make(chan struct{})
	for i := range mainTasks {
		root.AddDependent(mainTasks[i])
	}
	root.StartChildren()

	// when we get SIGINT or SIGTERM, kill everything by telling
	// the root tasks to kill themselves.
	killall := func() {
		for _, t := range mainTasks {
			t.Logger.Println("shutting down tasks by killing them")
			t.Kill()
		}
		os.Exit(0)
	}
	shutdown := func() {
		logger.Print("shutting down by closing done channel")
		close(root.Stopped)
	}
	WatchSignals(killall, shutdown, shutdown)

	// and run almost forever
	logstatus := time.NewTicker(15 * time.Second)
	for {
		select {
		case <-logstatus.C:
			pids := []int{}
			pids = root.CollectPIDs(pids)
			sort.Sort(sort.IntSlice(pids))
			logger.WithField("pids", pids).WithField("npids", len(pids)).Print("pidinfo")
		case <-root.Stopped:
			logger.Print("done channel was closed")
			// if we see this, then we were told to shutdown,
			// but we don't want to do it too early, so loop until
			// there are no main tasks left that haven't exited.
			looptime := 250 * time.Millisecond
			looptimer := time.NewTimer(looptime)
			longtime := 75 * time.Second
			toolong := time.NewTimer(longtime)
			for {
				select {
				case <-looptimer.C:
					canExit := true
					for _, t := range mainTasks {
						if !t.Exited() {
							canExit = false
						}
					}
					if canExit {
						os.Exit(0)
					}
					logger.Printf("waiting for all main tasks to die...")
					looptime *= 2
					looptimer.Reset(looptime)
				case <-toolong.C:
					logger.Error("requested shutdown did not complete within " + longtime.String())
					for _, t := range mainTasks {
						t.Destroy()
					}
					time.Sleep(1 * time.Second)
					os.Exit(1)
				}
			}
		}
	}
}
