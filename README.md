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
1. Download `github_chaos_deploy` from the Oneiro 1password account
    - Have someone on the team securely send it to you if needed
    - Copy it into the `bin/` directory
1. Run `./setup.sh` from the `bin/` directory

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
