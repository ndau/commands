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
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	arg "github.com/alexflint/go-arg"
)

// WatchSignals can set up functions to call on various operating system signals.
func WatchSignals(sigs map[os.Signal]func()) {
	signals := make([]os.Signal, 0)
	for s := range sigs {
		signals = append(signals, s)
	}
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, signals...)
		for {
			sig := <-sigchan
			switch sig {
			// we always want to terminate on SIGTERM
			case syscall.SIGTERM:
				sigs[sig]()
				os.Exit(0)
			default:
				sigs[sig]()
			}
		}
	}()
}

// loads the arguments and the configuration and returns the loaded
// config. If the config fails to load, it does not return.
func loadConfig() Config {
	var args struct {
		Configfile string `arg:"positional" help:"the name of the .toml config file to load"`
		NoCheck    bool   `help:"set this to disable checking that envvar substitutions are fully resolved"`
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
		return cfg
	}
	fmt.Printf("a config file name is required!")
	os.Exit(1)
	return cfg
}

// killall returns a function that kills everything by telling
// the root tasks to kill themselves.
func killall(mainTasks []*Task) func() {
	return func() {
		for _, t := range mainTasks {
			t.Logger.Println("shutting down tasks by killing them")
			t.Kill()
		}
		os.Exit(0)
	}
}

// shutdown returns a function that can shut things down more politely, kinda
func shutdown(root *Task, mainTasks []*Task) func() {
	return func() {
		root.Logger.Print("shutting down by closing Stopped channel on the root")
		close(root.Stopped)
		exitcode := waitForTasksToDie(root, mainTasks)
		os.Exit(exitcode)
	}
}

// runfunc creates a function that allows us to run a task with a Signal or
// from a timer.
// If task.Shutdown is defined, we shut everything else down first.
// Then we run the task, and when it is finished, we check task.Terminate.
// If task.Terminate is defined, we call killall. Otherwise, if necessary
// (task.Shutdown was true) we run the root task again.
func runfunc(task, root *Task, mainTasks, periodicTasks []*Task) func() {
	return func() {
		if task.Shutdown {
			root.Logger.Warn("running shutdown task, temporarily stopping all tasks")
			close(root.Stopped)
			exitcode := waitForTasksToDie(root, mainTasks)
			root.Logger.WithField("exitcode", exitcode).Debug("all tasks terminated")
		}
		tempstop := make(chan struct{})
		task.Start(tempstop)
		close(tempstop)
		root.Logger.Debug("finished running shutdown task")
		if task.Terminate {
			killall(mainTasks)()
		}
		if task.Shutdown {
			root.Logger.Debug("restarting main tasks")
			startChildren(root, mainTasks, periodicTasks)
			root.Logger.Warn("shutdown processing complete")
		}
	}
}

// Helper function to set up root.Stopped and start its child tasks.
func startChildren(root *Task, mainTasks, periodicTasks []*Task) {
	root.Stopped = make(chan struct{})
	root.StartChildren()
	setupPeriodic(root, mainTasks, periodicTasks)
}

func waitForTasksToDie(root *Task, mainTasks []*Task) int {
	looptime, err := parseDuration(os.Getenv("SHUTDOWN_LOOPTIME"), 250*time.Millisecond)
	if err != nil {
		root.Logger.WithError(err).Info("SHUTDOWN_LOOPTIME could not be parsed")
	}
	looptimer := time.NewTimer(looptime)
	longtime, err := parseDuration(os.Getenv("SHUTDOWN_MAX"), 75*time.Second)
	if err != nil {
		root.Logger.WithError(err).Info("SHUTDOWN_MAX could not be parsed")
	}
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
				return 0
			}
			root.Logger.Printf("waiting for all main tasks to die...")
			looptime *= 2
			looptimer.Reset(looptime)
		case <-toolong.C:
			root.Logger.Error("requested shutdown did not complete within " + longtime.String())
			for _, t := range mainTasks {
				t.Destroy()
			}
			time.Sleep(1 * time.Second)
			return 1
		}
	}
}

func setupSighandlers(root *Task, tasks Tasks) {
	// define some default sighandlers; they can be overridden in the
	// config file and additional ones can be defined
	sighandlers := map[os.Signal]func(){
		syscall.SIGTERM: killall(tasks.Main),
		syscall.SIGINT:  shutdown(root, tasks.Main),
	}
	for sig, task := range tasks.Signals {
		sighandlers[sig] = runfunc(task, root, tasks.Main, tasks.Periodic)
	}
	WatchSignals(sighandlers)
}

func setupPeriodic(root *Task, mainTasks, periodicTasks []*Task) {
	// set up the execution of any periodic tasks
	for _, t := range periodicTasks {
		f := runfunc(t, root, mainTasks, periodicTasks)
		dur := t.Periodic
		logger := t.Logger
		logger.WithField("period", dur).Info("setting up periodic task")
		go func() {
			ticker := time.NewTicker(dur)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					logger.Info("periodic task running")
					f()
				case <-root.Stopped:
					return
				}
			}
		}()
	}
}

func main() {
	cfg := loadConfig()

	// Init honeycomb filters if applicable; no-op otherwise.
	// Do this before building the root logger and before building tasks,
	// so that useHoneycomb is initialized.
	filterError := initFilters()

	// Create a logger even if we failed to init filters.
	logger := cfg.BuildLogger(rootTaskName)

	// Now that we have a logger, we can log any filter error that may have occurred.
	if filterError != nil {
		logger.WithError(filterError).Fatal("problems initializing filters")
	}

	err := cfg.RunPrologue(logger)
	if err != nil {
		logger.WithError(err).Fatal("problems running prologue")
	}

	tasks, err := cfg.BuildTasks(logger)
	// if we can't read the tasks we shouldn't even continue
	if err != nil {
		logger.WithError(err).Fatal("aborting because task file was invalid")
	}

	// now build a special task to act as the parent of the root tasks
	root := NewTask(rootTaskName, "")
	root.Logger = logger
	for i := range tasks.Main {
		root.AddDependent(tasks.Main[i])
	}
	startChildren(root, tasks.Main, tasks.Periodic)
	setupSighandlers(root, tasks)

	// and run almost forever
	logstatus := time.NewTicker(15 * time.Second)
	for {
		select {
		case <-logstatus.C:
			pids := []int{}
			pids = root.CollectPIDs(pids)
			sort.Sort(sort.IntSlice(pids))
			logger.WithField("pids", pids).WithField("npids", len(pids)).Print("pidinfo")
		}
	}
}
