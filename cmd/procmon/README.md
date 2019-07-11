# Procmon

Procmon is a process runner/monitor tool that can start and stop a set of tasks and rerun them if
they fail.

What's unique about it compared to things like [supervisor](http://supervisord.org/) is that it
lets you define a dependency tree of processes, so that when a parent process dies, its children
can be terminated and re-run after the parent has started again.

It's designed to work when run inside a container.

## Features

* Watches long-running tasks to see if they terminate
* Kills off children if the parent terminates, then restarts in order
* Flexible idea of additional "behavioral" monitors that can monitor a variety of things (http health, RPC, etc).
* A monitor reports failure as if the task has died
* Signal management; shuts everything down on exit
* Kill off tasks that fail monitoring but are still running
* Timeouts to force restart if things don't shut down when requested
* Composable monitors allow making them more sophisticated when desired
* Redirection of stdout/stderr
* Logging its own behavior to log files or to honeycomb
* Use SIGHUP to trigger a special task after shutting down everything (for example, for backup)
* A task definition language (config) so we don't need to compile the tool when tasks change

## Task definition language

The tasks are defined in a TOML file; see sample.toml for an example.
