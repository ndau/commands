package main

/*
Procmon is a command that manages a tree of os-level tasks. If any task terminates or
otherwise fails a validity test, that task and all of its descendants are stopped
and re-run in order.

This allows us to manage the delicate dance of databases, the tendermint stack, and
the application, making sure that all the dependencies are satisfied.

The key abstractions are Task and Monitor.

A Task has a couple of key operations:

* Start
* Stop
* AddDependent
* AddMonitor

Monitor is constructed with a Task and a chan Eventer, and it has a Listen method.

It's possible to write Monitor wrappers.

An Eventer is any object that implements Event(), which returns one of the following status codes:

* Start
* OK
* Failing
* Failed
* Shutdown

A task is constructed, then has any dependents added, and then any monitors. Then it is Started.
Part of the start process is that it launches its associated Eventer named "ready" and
it will not continue until that monitor has sent OK. The default Ready eventer simply
returns OK immediately.

Once a task is Started, its monitors are also started.

Finally, a task has a special listener that watches for the task to terminate at the
operating system level. This behaves as if one of the task monitors had failed.

The process monitor itself, as well as any additional monitors, will send events on
the task's status channel.

If a task receives a Failed message on its status channel, it:
* Stops all its children
* Stops its associated process (if it hadn't already crashed)
* Starts its associated process
* Waits for it to be ready
* Starts each of its children
* Waits for them to be ready

Other messages are advisory; individual task or monitor implementations may choose to use them.

# Environments

The config has a section for environment variables. Obviously the runtime environment exists
as well. If the same value is set in both, the config file should be considered to have
default values in it but you can override them by specifying them in your global environment
before running procmon.

Each task also has a section for task-specific config values.

When each task is run, it is started with an environment composed of:
* The global env values specified in the config
* The global environment from the procmon task

In the case of duplicate keys, the global environment overrides values specified in the
config file.

Values (not keys)  in the config file may contain interpolation strings of the forms:
* $key
* ${key}
Key will be interpreted as all of the consecutive characters [a-zA-Z0-9_], which is why
the second form is safer. $$ means a single $.

Because the environment is considered to be an unordered map, values in the
[env] section are interpolated using only the global environment. Values in the
rest of the file are interpolated using the combination of environments in the
order above.

Interpolation happens only once, so attempting to do something like ${FOO} where FOO is defined
as $OTHER will not result in the value of $OTHER being injected.

*/
