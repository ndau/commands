# Oneiro ndev Developer Setup

## Overview

This document contains steps for getting set up to build and test ndev applications.  By the end you will be able to run the `ndau` blockchain, talking to `redis`, `noms` and `tendermint`, from the command line.  This is the way to do it if you would eventually like to debug the applications, as they run simultaneously and interact with each other from their own shells.

The `/bin` directory also contains other scripts useful for developing within a local development environment.  More information can be found in its [README](bin/README.md).

## Setup Tools

### Prerequisites

* Ensure that you have SSH clone access to the [oneiro-ndev](https://github.com/oneiro-ndev) repos
* The following instructions have been tested on a fresh user account using macOS High Sierra version 10.13.6

### Install

These steps only need to be performed once:

1. Install [Xcode](https://itunes.apple.com/us/app/xcode/id497799835)
1. Install [Go](https://golang.org/doc/install)
1. Install [Python3](https://www.python.org/downloads/)
1. Restart your terminal if necessary to update `$PATH`
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
    git clone git@github.com:oneiro-ndev/commands.git ~/go/src/github.com/oneiro-ndev/commands
    ```
1. Set up genesis files: (Optional)
    - Using custom genesis files is useful if running a localnet for mainnet-specific tasks; they are generated automatically if not supplied
    - Get a copy of `genesis_files.tar` from Oneiro's 1password account
    - Create the directory `~/.localnet`
    - Extract `genesis_files.tar` within `~/.localnet`
    - You should now see the following items in your `~/.localnet/genesis_files/` directory:
        - `system_accounts.toml`
        - `system_vars.toml`
        - (any other files or subdirectories in here are not needed and can be removed if desired)
1. Run `./bin/setup.sh N` where `N` is the number of nodes you'd like to run

### Running

Use `./bin/run.sh`.

This will run all the tasks in the proper sequence and create a set of appropriately-named .pid and .log files, one for each task.  All tasks will run in the background.

### Shutting it down

Use `./bin/kill.sh`.

This will shut down any running tasks in the reverse order from which they were run. If a task doesn't shut itself down nicely, it will be killed.

### Reset

To run with fresh databases, run `./bin/reset.sh` before your next `./bin/run.sh`.

### Individual commands

Both `./bin/run.sh` and `./bin/kill.sh` take an argument, which is the name of the task you wish to run or kill. Valid task names are:

* ndau_redis
* ndau_noms
* ndau_node
* ndau_tm

You can also specify the node number for each.  For example, if you ran `./bin/setup.sh` with a node count greater than 1, then you can `./bin/run.sh ndau_redis 1` to run ndau redis for the zero-based node number 1.  If you leave off the node number in these commands, the default 0'th node will be used.

### Rebuild

Use `./bin/build.sh` if you make changes to any of the tools and want to rebuild them before running again.

### Test

Use `./bin/test.sh` to run unit tests on the latest built tools.
Use `./bin/test.sh -i` to run integration tests found in `/ndauapi/routes`.

## Chaincode tools

### Building the tools

From the root of the commands repository, you can use `make`. It basically expects that you are working from within goroot and that the chaincode repo is at `../chaincode` and also expects `../validation_scripts` relative to this `commands` repo.

Given that, you should be able to do `make build` to create all the tools.

The tools it creates are:

* opcodes (the code generator that ensures that all the chaincode sources use the same set of opcodes)
* chasm (the chaincode assembler)
* chfmt (the chasm formatter)
* crank (the chaincode debugger, repl, and test tool)

Once you have built the tools, you can do:

* `make validations` to build all the validation scripts.
* `make vtests` will test all the validation scripts based on finding files with the .crank extension in the `../validation_scripts` directory.
* `make vformat` will run the formatter over all the scripts in that directory. Note that the formatter currently has the potential to damage a file if it cannot be parsed, so you would be wise to commit an unformatted version before you run it; the safest bet is to make sure it compiles first.

### Notes on crank

Crank was originally intended to be a chaincode repl. It can definitely be used that way, but usually you'll be better off running it from a script.

If you encounter a puzzling bug, you can use the -verbose switch; if this is set, when crank encounters a script error it will drop into the repl so you can try to look around and maybe reset and walk through it.

The `help` and `help verbose` commands will dump some helpful text about how to use crank. Also see the readme in cmd/crank.

## Other tools

See the [README](bin/README.md) in the `./bin` directory for more information on the tools found there
