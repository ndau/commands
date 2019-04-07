package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Config is the overall structure for a procmon TOML config file
// It looks like sample.toml.
// These are the major sections
type Config struct {
	Env      map[string]string
	Logger   map[string]string
	Prologue []map[string]string
	Task     []ConfigTask
}

// The ConfigTask section is a map ("table") of tasks
type ConfigTask struct {
	Name        string
	Path        string
	Args        []string
	Onetime     bool
	Stdout      string
	Stderr      string
	Parent      string
	MaxStartup  string
	MaxShutdown string
	Monitors    []map[string]string
	Prerun      []string
}

func (ct *ConfigTask) interpolate(env map[string]string) {
	ct.Name = interpolate(ct.Name, env)
	ct.Path = interpolate(ct.Path, env)
	ct.Stdout = interpolate(ct.Stdout, env)
	ct.Stderr = interpolate(ct.Stderr, env)
	ct.Parent = interpolate(ct.Parent, env)
	ct.MaxShutdown = interpolate(ct.MaxShutdown, env)
	ct.Args = interpolateAll(ct.Args, env).([]string)
	for i := range ct.Monitors {
		ct.Monitors[i] = interpolateAll(ct.Monitors[i], env).(map[string]string)
	}
}

// Load does the toml load into a config object
func Load(filename string, nocheck bool) (Config, error) {
	cfg := Config{}
	metadata, err := toml.DecodeFile(filename, &cfg)
	if err != nil {
		return cfg, errors.Wrap(err, fmt.Sprintf("metadata = %#v", metadata))
	}

	// start by interpolating the config's environment variables with the global ones
	globalenv := envmap(os.Environ(), make(map[string]string))
	for k, v := range cfg.Env {
		cfg.Env[k] = interpolate(v, globalenv)
	}
	// and do it again only this time with the envvars we just interpolated
	// we do it 5 times to deal with nested cases (the environment is a map,
	// so there's no iteration order, and we need to be careful not to
	// recurse forever in case someone tries something like A=$A).
	cfg.Env = envmap(os.Environ(), cfg.Env)
	for i := 0; i < 5; i++ {
		for k, v := range cfg.Env {
			cfg.Env[k] = interpolate(v, cfg.Env)
		}
	}
	// unless they tell us not to, check to see if all the environment variables were
	// processed by looking for leftover things that look like envvar expansions.
	if !nocheck {
		for k, v := range cfg.Env {
			if found, _ := regexp.MatchString(`\$\(?[A-Za-z0-9]+\)?`, v); found {
				return cfg, errors.New("Unprocessed envvars still found in Env: " + k + ":" + v)
			}
		}
	}

	// Now we can use that to interpolate the rest
	// of the loaded configuration
	cfg.Logger = interpolateAll(cfg.Logger, cfg.Env).(map[string]string)

	for i := range cfg.Prologue {
		cfg.Prologue[i] = interpolateAll(cfg.Prologue[i], cfg.Env).(map[string]string)
	}

	for i := range cfg.Task {
		cfg.Task[i].interpolate(cfg.Env)
	}

	return cfg, err
}

// RunPrologue creates and runs pingers in order to establish that everything is
// ready to go.
func (c *Config) RunPrologue(logger *logrus.Logger) error {
	for _, p := range c.Prologue {
		pinger, err := BuildMonitor(p, logger)
		if err != nil {
			return err
		}
		if status := pinger(); status != OK {
			l := logger.WithField("status", status)
			for k, v := range p {
				l = l.WithField(k, v)
			}
			l.Error("did not return ok")
			return errors.New("pinger failed: " + p["name"])
		}
	}
	return nil
}

// BuildMonitor constructs a monitor from an element in the
// Monitors map
func BuildMonitor(mon map[string]string, logger *logrus.Logger) (func() Eventer, error) {
	switch mon["type"] {
	case "portavailable":
		if mon["port"] == "" {
			return nil, errors.New("portavailable requires a port parm")
		}
		m := PortAvailable(mon["port"])
		return m, nil
	case "portinuse":
		if mon["port"] == "" {
			return nil, errors.New("portinuse requires a port parm")
		}
		if mon["timeout"] == "" {
			mon["timeout"] = "100ms"
		}
		timeout, err := time.ParseDuration(mon["timeout"])
		if err != nil {
			return nil, err
		}
		m := PortInUse(mon["port"], timeout, logger)
		return m, nil
	case "ensuredir":
		if mon["path"] == "" {
			return nil, errors.New("ensuredir requires a path parm")
		}
		if mon["perm"] == "" {
			mon["perm"] = "0755"
		}
		// we always parse permissions in base 8
		perm, err := strconv.ParseInt(mon["perm"], 8, 32)
		if err != nil {
			return nil, err
		}
		m := EnsureDir(mon["path"], os.FileMode(perm))
		return m, nil
	case "redis":
		if mon["addr"] == "" {
			mon["addr"] = "localhost:6379"
		}
		m := RedisPinger(mon["addr"])
		return m, nil
	case "http":
		if mon["timeout"] == "" {
			mon["timeout"] = "1s"
		}
		timeout, err := time.ParseDuration(mon["timeout"])
		if err != nil {
			return nil, err
		}
		m := HTTPPinger(mon["url"], timeout, logger)
		return m, nil
	default:
		return nil, errors.New("unknown monitor type " + mon["type"])
	}
}

func fileparse(name string) (io.WriteCloser, error) {
	// "SUPPRESS" (also "") meaning "discard this stream"
	// "HONEYCOMB" sends the message to honeycomb
	// Anything else is a named file
	switch name {
	case "SUPPRESS", "":
		return nil, nil
	case "HONEYCOMB":
		return nil, errors.New("honeycomb is not currently supported as a log destination")
	default:
		f, err := os.Create(name)
		if err != nil {
			return nil, err
		}
		return f, nil
	}
}

// BuildTasks constructs all the tasks from a loaded config
// It returns an array of the tasks that need to be individually
// started. All child tasks will be descendants of these.
func (c *Config) BuildTasks(logger *logrus.Logger) ([]*Task, error) {
	taskm := make(map[string]*Task)
	tasks := make([]*Task, 0)
	for _, ct := range c.Task {
		t := NewTask(ct.Name, ct.Path, ct.Args...)
		t.Env = c.Getenv()
		t.Onetime = ct.Onetime
		// set up any monitors we need
		for _, mon := range ct.Monitors {
			m, err := BuildMonitor(mon, logger)
			if err != nil {
				return nil, err
			}
			switch mon["name"] {
			case "ready":
				t.Ready = m
			default:
				if mon["period"] == "" {
					mon["period"] = "15s"
				}
				period, err := time.ParseDuration(mon["period"])
				if err != nil {
					return nil, err
				}
				nm := NewFailMonitor(NewMonitor(t.Status, period, m))
				t.Monitors = append(t.Monitors, nm)
			}
		}
		// check for stdout/err assignments
		stdout, err := fileparse(ct.Stdout)
		if err != nil {
			return nil, err
		}
		t.Stdout = stdout
		stderr, err := fileparse(ct.Stderr)
		if err != nil {
			return nil, err
		}
		t.Stderr = stderr

		// MaxStartup
		if ct.MaxStartup != "" {
			maxstartup, err := time.ParseDuration(ct.MaxStartup)
			if err != nil {
				return nil, err
			}
			t.MaxStartup = maxstartup
		}

		// MaxShutdown
		if ct.MaxShutdown != "" {
			maxshutdown, err := time.ParseDuration(ct.MaxShutdown)
			if err != nil {
				return nil, err
			}
			t.MaxShutdown = maxshutdown
		}

		// set up the logger
		t.Logger = logger.WithField("task", t.Name).WithField("bin", "procmon")

		for _, prerun := range ct.Prerun {
			if _, ok := taskm[prerun]; !ok {
				return nil, errors.New("did not find prerun task " + prerun)
			}
			t.Prerun = append(t.Prerun, taskm[prerun])
		}

		// if the task has a parent, assign it as a dependent
		if ct.Parent != "" {
			if _, ok := taskm[ct.Parent]; !ok {
				return nil, errors.New("did not find parent task " + ct.Parent)
			}
			taskm[ct.Parent].AddDependent(t)
		} else {
			// if no parent, then it's in the root set of tasks that have to
			// be started directly
			tasks = append(tasks, t)
		}
		taskm[ct.Name] = t
	}

	// return the root set
	return tasks, nil
}

// BuildLogger constructs a logger given the configuration info.
func (c *Config) BuildLogger() *logrus.Logger {
	var formatter logrus.Formatter
	var out io.Writer
	var level logrus.Level
	switch c.Logger["output"] {
	case "stdout":
		out = os.Stdout
	case "stderr", "":
		out = os.Stderr
	case "SUPPRESS":
		out = nil
	default:
		out = nil
	}

	switch c.Logger["format"] {
	case "json", "":
		formatter = new(logrus.JSONFormatter)
	case "text":
		formatter = new(logrus.TextFormatter)
	default:
		formatter = new(logrus.JSONFormatter)
	}

	switch c.Logger["level"] {
	case "info", "":
		level = logrus.InfoLevel
	case "debug":
		level = logrus.DebugLevel
	case "warn", "warning":
		level = logrus.WarnLevel
	case "err", "error":
		level = logrus.ErrorLevel
	default:
		level = logrus.InfoLevel
	}

	logger := logrus.New()
	logger.Out = out
	logger.Formatter = formatter
	logger.Level = level
	return logger
}

// Getenv returns a composite environment with the
// same interface as os.Getenv, but any env vars
// in the config are also included
func (c *Config) Getenv() []string {
	env := make([]string, 0, len(c.Env))
	for k, v := range c.Env {
		env = append(env, k+"="+v)
	}
	return env
}
