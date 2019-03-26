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

Once a task is Started, it listens to its monitors.

A task is Start-ed, and then it's told to Listen.

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

*/
