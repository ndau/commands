package main

import (
	"errors"
	"io"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
)

// Config is the overall structure for a procmon TOML config file
// It looks like sample.toml.
// These are the major sections
type Config struct {
	Env    map[string]string
	Logger map[string]string
	Task   []ConfigTask
}

// The ConfigTask section is a map ("table") of tasks
type ConfigTask struct {
	Name        string
	Path        string
	Args        []string
	Stdout      string
	Stderr      string
	Parent      string
	MaxStartup  string
	MaxShutdown string
	Monitors    []map[string]string
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
func Load(filename string) (Config, error) {
	cfg := Config{}
	_, err := toml.DecodeFile(filename, &cfg)

	// now interpolate the config's environment variables
	globalenv := envmap(os.Environ(), make(map[string]string))
	for k, v := range cfg.Env {
		cfg.Env[k] = interpolate(v, globalenv)
	}
	// Create the real env for the rest of what we do by
	// updating the loaded env from config with the global
	// environment, and then save that in the cfg.
	fullenv := envmap(os.Environ(), cfg.Env)
	cfg.Env = fullenv

	// Now we can use that to interpolate the rest
	// of the loaded configuration
	cfg.Logger = interpolateAll(cfg.Logger, cfg.Env).(map[string]string)

	for i := range cfg.Task {
		cfg.Task[i].interpolate(cfg.Env)
	}

	return cfg, err
}

// BuildMonitor constructs a monitor from an element in the
// Monitors map
func BuildMonitor(mon map[string]string) (func() Eventer, error) {
	switch mon["type"] {
	case "openport":
		if mon["port"] == "" {
			return nil, errors.New("openport requires a port parm")
		}
		m := OpenPort(mon["port"])
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
		m := HTTPPinger(mon["url"], timeout)
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
		// set up any monitors we need
		for _, mon := range ct.Monitors {
			m, err := BuildMonitor(mon)
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
				nm := NewMonitor(t.Status, period, m)
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
