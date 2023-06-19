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
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// LoggerOutput string constants.
const (
	LoggerOutputSuppress = "SUPPRESS"
	LoggerOutputStdout   = "STDOUT"
	LoggerOutputStderr   = "STDERR"
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
	Specials    map[string]interface{}
	Stdout      string
	Stderr      string
	Parent      string
	MaxStartup  string
	MaxShutdown string
	Monitors    []map[string]string
	Prerun      []string
}

// Tasks is the container for all the task types that get manipulated.
type Tasks struct {
	Main     []*Task
	Signals  map[os.Signal]*Task
	Periodic []*Task
	All      map[string]*Task
}

// NewTasks creates the Tasks object
func NewTasks() Tasks {
	return Tasks{
		Main:     make([]*Task, 0),
		Signals:  make(map[os.Signal]*Task),
		Periodic: make([]*Task, 0),
		All:      make(map[string]*Task),
	}
}

// parseDuration parses a duration (a string from the environment) and
// accepts a default value. Even when it returns an error, it also returns
// the default, so that the error can be logged but let the system continue.
func parseDuration(idur interface{}, def time.Duration) (time.Duration, error) {
	switch dur := idur.(type) {
	case string:
		// if it's empty, there's no error, just return the default
		if dur == "" {
			return def, nil
		}
		// if it's not empty, and can't be parsed, return the error for logging
		// as well as the default.
		v, err := time.ParseDuration(dur)
		if err != nil {
			return def, err
		}
		return v, nil
	default:
		return def, nil
	}
}

// parseBool turns an interface value into a bool
func parseBool(v interface{}, def bool) bool {
	switch b := v.(type) {
	case string:
		switch strings.ToLower(b) {
		case "t", "y", "true", "yes":
			return true
		case "f", "n", "false", "no":
			return false
		default:
			return def
		}
	case bool:
		return b
	default:
		return def
	}
}

// parseSignal interprets a signal name in an interface and returns the associated signal
func parseSignal(v interface{}) os.Signal {
	switch s := v.(type) {
	case string:
		switch s {
		case "SIGHUP", "HUP":
			return syscall.SIGHUP
		case "SIGINT", "INT":
			return syscall.SIGINT
		case "SIGTERM", "TERM":
			return syscall.SIGTERM
		case "SIGUSR1", "USR1":
			return syscall.SIGUSR1
		case "SIGUSR2", "USR2":
			return syscall.SIGUSR2
		default:
			return nil
		}
	default:
		return nil
	}
}

func (ct *ConfigTask) interpolate(env map[string]string) {
	ct.Name = interpolate(ct.Name, env)
	ct.Path = interpolate(ct.Path, env)
	ct.Stdout = interpolate(ct.Stdout, env)
	ct.Stderr = interpolate(ct.Stderr, env)
	ct.Parent = interpolate(ct.Parent, env)
	ct.MaxShutdown = interpolate(ct.MaxShutdown, env)
	ct.Args = interpolateAll(ct.Args, env).([]string)
	ct.Specials = interpolateAll(ct.Specials, env).(map[string]interface{})
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
func (c *Config) RunPrologue(logger logrus.FieldLogger) error {
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
func BuildMonitor(mon map[string]string, logger logrus.FieldLogger) (func() Eventer, error) {
	switch mon["type"] {
	case "portavailable":
		if mon["port"] == "" {
			return nil, errors.New("portavailable " + mon["name"] + " requires a port parm")
		}
		m := PortAvailable(mon["port"])
		return m, nil
	case "portinuse":
		if mon["port"] == "" {
			return nil, errors.New("portinuse requires a port parm")
		}
		timeout, err := parseDuration(mon["timeout"], 100*time.Millisecond)
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
		m := RedisPinger(mon["addr"], logger)
		return m, nil
	case "http":
		timeout, err := parseDuration(mon["timeout"], time.Second)
		if err != nil {
			return nil, err
		}
		m := HTTPPinger(mon["url"], timeout, logger)
		return m, nil
	default:
		return nil, errors.New("unknown monitor type " + mon["type"])
	}
}

// Pass in one of the LoggerOutput* contants.
// If blank, the given default is used.
// Otherwise, the logger output is assumed to be a file name.
// If the HONEYCOMB_* env vars are set, then all logging goes to honeycomb.
func fileparse(taskName, loggerOutput string, def io.Writer) (io.Writer, error) {
	if useHoneycomb {
		// Route all output from a given task to its own honeycomb filter.
		return newFilter(taskName), nil
	}

	switch loggerOutput {
	case "":
		return def, nil
	case LoggerOutputStdout:
		return os.Stdout, nil
	case LoggerOutputStderr:
		return os.Stderr, nil
	case LoggerOutputSuppress:
		return ioutil.Discard, nil
	default:
		f, err := os.OpenFile(loggerOutput, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		return f, nil
	}
}

// BuildTasks constructs all the tasks from a loaded config
// It returns an array of the tasks that need to be individually
// started. All child tasks will be descendants of these.
func (c *Config) BuildTasks(logger logrus.FieldLogger) (Tasks, error) {
	tasks := NewTasks()
	// taskm := make(map[string]*Task)
	// tasks := make([]*Task, 0)
	for _, ct := range c.Task {
		t := NewTask(ct.Name, ct.Path, ct.Args...)
		t.Env = c.Getenv()
		// set up any monitors we need
		for _, mon := range ct.Monitors {
			m, err := BuildMonitor(mon, logger)
			if err != nil {
				return tasks, err
			}
			switch mon["name"] {
			case "ready":
				t.Ready = m
			default:
				period, err := parseDuration(mon["period"], 15*time.Second)
				if err != nil {
					return tasks, err
				}
				nm := NewFailMonitor(NewMonitor(t.Status, period, m))
				t.Monitors = append(t.Monitors, nm)
			}
		}
		// check for stdout/err assignments
		stdout, err := fileparse(t.Name, ct.Stdout, os.Stdout)
		if err != nil {
			return tasks, err
		}
		t.Stdout = stdout
		stderr, err := fileparse(t.Name, ct.Stderr, os.Stderr)
		if err != nil {
			return tasks, err
		}
		t.Stderr = stderr

		// MaxStartup
		maxstartup, err := parseDuration(ct.MaxStartup, t.MaxStartup)
		if err != nil {
			return tasks, err
		}
		t.MaxStartup = maxstartup

		// MaxShutdown
		maxshutdown, err := parseDuration(ct.MaxShutdown, t.MaxShutdown)
		if err != nil {
			return tasks, err
		}
		t.MaxShutdown = maxshutdown

		// set up the logger
		t.Logger = logger

		for _, prerun := range ct.Prerun {
			if _, ok := tasks.All[prerun]; !ok {
				return tasks, errors.New("did not find prerun task " + prerun)
			}
			t.Prerun = append(t.Prerun, tasks.All[prerun])
		}

		t.Onetime = parseBool(ct.Specials["onetime"], false)
		t.Periodic, err = parseDuration(ct.Specials["periodic"], 0)
		t.Terminate = parseBool(ct.Specials["terminate"], false)
		t.Shutdown = parseBool(ct.Specials["shutdown"], false)
		if err != nil {
			return tasks, err
		}
		sig := parseSignal(ct.Specials["signal"])
		switch {
		case sig != nil:
			tasks.Signals[sig] = t
		case ct.Parent != "":
			if _, ok := tasks.All[ct.Parent]; !ok {
				return tasks, errors.New("did not find parent task " + ct.Parent)
			}
			tasks.All[ct.Parent].AddDependent(t)
		case t.Periodic != 0:
			tasks.Periodic = append(tasks.Periodic, t)
		case ct.Parent == "" && t.Onetime == false:
			// if no parent and not a onetime task, then it's in the root set of tasks that have to
			// be started directly
			t.Logger.Infof(
				"Adding to main: name=%v onetime=%v periodic=%v terminate=%v shutdown=%v",
				t.Name, t.Onetime, t.Periodic, t.Terminate, t.Shutdown)
			tasks.Main = append(tasks.Main, t)
		}
		tasks.All[t.Name] = t
	}

	return tasks, nil
}

// BuildLogger constructs a logger given the configuration info.
// The returned logger is used by all tasks including the root task.  Tasks can apply additional
// logger.WithField()s, but if honeycomb logging is enabled, then all tasks will log to honeycomb
// regardless of their logger output configuration.
func (c *Config) BuildLogger(rootTaskName string) logrus.FieldLogger {
	var formatter logrus.Formatter
	var out io.Writer
	var level logrus.Level

	if useHoneycomb {
		// Output json format for logging from the procmon app itself.
		formatter = &logrus.JSONFormatter{}
		// Filter that output through the honeycomb logger.
		out = newFilter(rootTaskName)
	} else {
		switch c.Logger["output"] {
		case LoggerOutputStdout:
			out = os.Stdout
		case LoggerOutputStderr, "":
			out = os.Stderr
		case LoggerOutputSuppress:
			out = ioutil.Discard
		default:
			out = ioutil.Discard
		}

		switch c.Logger["format"] {
		case "json", "":
			formatter = &logrus.JSONFormatter{}
		case "text", "plain":
			formatter = &logrus.TextFormatter{}
		default:
			formatter = &logrus.JSONFormatter{}
		}
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
