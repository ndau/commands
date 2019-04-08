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

// HupTask is the name of the task that will be run on SIGHUP
const HupTask = "HUPTASK"

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

// this searches the main task list for a task called
// "HUPTASK" and if it finds it, removes it from
// the list and returns it separately.
func findHupTask(maintasks []*Task) ([]*Task, *Task) {
	ret := make([]*Task, 0, len(maintasks))
	var huptask *Task
	for _, t := range maintasks {
		if t.Name == HupTask {
			huptask = t
		} else {
			ret = append(ret, t)
		}
	}
	return ret, huptask
}

// when we get SIGINT or SIGTERM, kill everything by telling
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

// Shut things down more politely, kinda
func shutdown(root *Task, mainTasks []*Task) func() {
	return func() {
		root.Logger.Print("shutting down by closing Stopped channel on the root")
		close(root.Stopped)
		exitcode := waitForTasksToDie(root, mainTasks)
		os.Exit(exitcode)
	}
}

// When we receive sighup, if huptask is defined, we shut everything else down,
// run huptask, and when it terminates we run the root task again
func runhup(huptask, root *Task, mainTasks []*Task) func() {
	if huptask == nil {
		root.Logger.Debugf("no task called %s was defined so SIGHUP will do nothing", HupTask)
	}
	return func() {
		if huptask == nil {
			root.Logger.Warnf("SIGHUP received but no task called %s was found", HupTask)
			return
		}
		root.Logger.Warn("SIGHUP received, temporarily stopping all tasks")
		close(root.Stopped)
		exitcode := waitForTasksToDie(root, mainTasks)
		root.Logger.WithField("exitcode", exitcode).Warnf("all tasks terminated -- running %s", HupTask)
		tempstop := make(chan struct{})
		huptask.Start(tempstop)
		close(tempstop)
		root.Logger.Warnf("%s finished, restarting main tasks", HupTask)
		root.Stopped = make(chan struct{})
		root.StartChildren()
		root.Logger.Warn("SIGHUP processing complete")
	}
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

func main() {
	cfg := loadConfig()
	logger := cfg.BuildLogger()
	if os.Getenv("HONEYCOMB_KEY") != "" {
		logger = honeycomb.Setup(logger)
	}

	err := cfg.RunPrologue(logger)
	if err != nil {
		logger.WithError(err).Fatal("problems running prologue")
	}

	mainTasks, err := cfg.BuildTasks(logger)
	// if we can't read the tasks we shouldn't even continue
	if err != nil {
		logger.WithError(err).Fatal("aborting because task file was invalid")
	}

	mainTasks, huptask := findHupTask(mainTasks)

	// now build a special task to act as the parent of the root tasks
	root := NewTask("root", "")
	root.Logger = logger
	root.Stopped = make(chan struct{})
	for i := range mainTasks {
		root.AddDependent(mainTasks[i])
	}
	root.StartChildren()

	WatchSignals(runhup(huptask, root, mainTasks), shutdown(root, mainTasks), killall(mainTasks))

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
