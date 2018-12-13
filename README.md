# Oneiro ndev Developer Setup

## Overview

This document contains steps for getting set up to build and test ndev applications.  By the end you will be able to run `chaos` and `ndau` blockchains, talking to `redis`, `noms` and `tendermint`, from the command line.  This is the way to do it if you would eventually like to debug the applications, as they run simultaneously and interact with each other from their own shells.

The `/bin` directory also contains other scripts useful for developing within a local development environment.

## Setup Tools

### Prerequisites

* Ensure that you have SSH clone access to the [oneiro-ndev](https://github.com/oneiro-ndev) repos
* The following instructions have been tested on a fresh user account using macOS High Sierra version 10.13.6

### Install

These steps only need to be performed once:

1. Install [Xcode](https://itunes.apple.com/us/app/xcode/id497799835)
1. Install [Go](https://golang.org/doc/install)
1. Install [Python3](https://www.python.org/downloads/)
1. Restart your terminal if necessary to update `$PATH`.
1. Install [`remarshal`](https://github.com/dbohdan/remarshal):
    ```sh
    python3 -m pip install remarshal --user
    ```
1. Install [Brew](https://brew.sh/)
1. Install `dep`: `brew install dep`
1. Install Redis: `brew install redis`
1. Install `jq`: `brew install jq`
1. Clone this repo:
    ```sh
    git clone git@github.com:oneiro-ndev/commands.git $GOPATH/src/github.com/oneiro-ndev/commands
    ```
1. Run `bin/setup.sh` from the repo root

### Demo mode

[`demo.sh`](demo.sh) sets everything up, runs the node group, creates a `demo` ndau account, gives it some money, creates a `demo` chaos id associated with that ndau account, sends some transactions, and shows that the chaos transactions validated themselves on the ndau chain, before finally shutting everything down.

### Running

Use `./run.sh` from the `bin/` directory.

This will run all the tasks in the proper sequence and create a set of appropriately-named .pid and .log files, one for each task.  All tasks will run in the background.

### Shutting it down

Use `./kill.sh`

This will shut down any running tasks in the reverse order from which they were run. If a task doesn't shut itself down nicely, it will be killed.

### Reset

To run with fresh databases, run `./reset.sh` before your next `./run.sh`.

### Individual commands

Both `run.sh` and `kill.sh` take a single argument, which is the name of the task you wish to run or kill. Valid task names are:

* chaos_redis
* chaos_noms
* chaos_node
* chaos_tm
* ndau_redis
* ndau_noms
* ndau_node
* ndau_tm

### Rebuild

Use `./build.sh` from the `bin/` directory if you make changes to any of the tools and want to rebuild them before running again.

## Other Tools

### linkdep

This tool is useful when you want to make changes to one of our dependency projects and test it locally without first having to push it up to github.

Normally we have cloned `chaos` and `ndau` into `~/go/src/github.com/oneiro-ndev` and we make changes there to those projects like any other git repos.  But if you want to make changes on one of our dependency probjects, say, `metanode`, then you can use the `linkdep.sh` tool to set that up for you.

Steps:

1. Clone `metanode` next to `commands`
1. Run `./linkdep.sh metanode` from anywhere

What this does is it creates symbolic links from the `commands` vendor directory for metanode back to your cloned copy of metanode.  You then can make changes from within your cloned directory and interact with git as usual.  When you want to test any changes you've made to metanode, you can run `./build.sh` and `./test.sh` as usual.

Any time you run a `dep ensure` from `commands`, you must run `./linkdep.sh metanode` again if you'd like to test more local changes to metanode that haven't yet been pushed and landed to the appropriate branch (usually master) on github.

#### Rationale

We tried various other approaches that didn't work out as well as this:

* `go get github.com/oneiro-ndev/metanode`
    - Doesn't get metanode's dependencies, so `go build ./...` fails
* `glide mirror set git@github.com:oneiro-ndev/metanode.git file:///Users/<username>/go/src/github.com/oneiro-ndev/metanode --vcs git`
    - One extra global developer step to config
    - Have to commit changes to your metanode branch and run `glide install` to test
* `glide init` with `glide install` within `metanode`
    - Hacky way of using glide
    - Still have to edit `chaos` and `ndau` glide.yaml to pull from your branch to test locally
