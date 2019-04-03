# Oneiro ndev Developer Setup

### Overview

This document contains steps for getting set up to build and test ndev applications.  By the end you will be able to run the `ndau` blockchain, talking to `redis`, `noms` and `tendermint`, from the command line.  This is the way to do it if you would eventually like to debug the applications, as they run simultaneously and interact with each other from their own shells.

The `/bin` directory also contains other scripts useful for developing within a local development environment.  More information can be found in its [README](bin/README.md).

The following instructions have been tested on clean installs of macOS Mojave version 10.4.4 and Ubuntu 18.10.

### Prerequisites

Ensure that you have SSH clone access to the [oneiro-ndev](https://github.com/oneiro-ndev) repos required: `chaincode  genesis  json2msgp  metanode  msgp-well-known-types  mt19937_64  ndau  ndaumath  noms-util  o11y  system_vars  writers`.

#### macOS:

The Homebrew package manager is by far the easiest way to install these tools, but each can be installed separately from the distribution's standard download package.
1. Install the Xcode command-line tools: `xcode-select —install`
2. Install [Brew](https://brew.sh/)
3. Install `go`: `brew install go`
4. Install `dep`: `brew install dep`
5. Install `redis`: `brew install redis`
6. Install `jq`: `brew install jq`

#### Ubuntu:

Install tooling: `sudo apt install golang go-dep redis jq git -y`
2. Install `dep`: `sudo apt install go-dep -y`
3. Install `redis`: `sudo apt install redis -y`
4. Install `jq`: `sudo apt install jq -y`
5. Install `git`: `sudo apt install git -y`

### ndau Tools

1. Clone the ndau commands repo:
    ```sh
    git clone git@github.com:oneiro-ndev/commands.git "$GOPATH"/src/github.com/oneiro-ndev/commands
    ```
1. Build all tools, set up for a single-node localnet for testing:
   ```sh
   $GOPATH/src/github.com/oneiro-ndev/commands
   ./bin/setup.sh 1
   ```
   Replace `1` with the desired number of nodes for a larger localnet configuration.

### Custom genesis configuration

To create a custom configuration (usually to replicate a testnet or mainnet configuration), do the following **before** running `./bin/run.sh` for the first time. If you're already running with the default pre-installed configuration, remove the `~/.localnet` directory first.

1. Set up custom genesis files: (Optional)
    
    - Create the directory `~/.localnet/genesis_files`
    - Create the configuration files in your `~/.localnet/genesis_files/` directory:
    
    `go run $GOPATH/src/github.com/oneiro-ndev/commands/cmd/generate \
       -g ~/.localnet/genesis_files/system_vars.toml \
       -a ~/.localnet/genesis_files/system_accounts.toml`

### Running

```sh
./bin/run.sh
```

To start a new localnet with a default configuration pre-installed, answer `y` to the prompt
```sh
Cannot find genesis file: ~/.localnet/genesis_files/system_vars.toml
Generate new? [y|n]: y
```

This will run all the tasks in the proper sequence and create a set of appropriately-named .pid and .log files, one for each task.  All tasks will run in the background.

### Shutting it down

Use `./bin/kill.sh`.

This will shut down any running tasks in the reverse order from which they were run. If a task doesn't shut itself down nicely, it will be killed.

### Reset

To run with fresh databases, run `./bin/reset.sh` before your next `./bin/run.sh`.

### Individual commands

Both `./bin/run.sh` and `./bin/kill.sh` take an argument, which is the name of the task you wish to run or kill. Valid task names are:

* `ndau_redis`
* `ndau_noms`
* `ndau_node`
* `ndau_tm`

You can also specify the node number for each.  For example, if you ran `./bin/setup.sh` with a node count greater than 1, then you can `./bin/run.sh ndau_redis 1` to run ndau redis for the zero-based node number 1.  If you leave off the node number in these commands, the default 0'th node will be used.

### Rebuild

Use `./bin/build.sh` if you make changes to any of the tools and want to rebuild them before running again.

### Test

Use `./bin/test.sh` to run unit tests on the latest built tools.
Use `./bin/test.sh -i` to run integration tests found in `/ndauapi/routes`.

## Running the ndau API

The ndau API is a REST server for interacting with the ndau blockchain, and is the standard method for doing so. The default local server runs at `https://localhost:3030` and can be started as:
```sh
./bin/ndauapi.sh
```

## Chaincode tools

Chaincode is the scripting language ndau uses for validation rules, fee and rate calculations, and other configurable behaviors. It is usually only needed when creating custom validation rules.

### Building the tools

From the root of the commands repository, you can use `make`. It basically expects that you are working from within goroot and that the chaincode repo is at `../chaincode` and also expects `../chaincode_scripts` relative to this `commands` repo. The `../chaincode_scripts` repo is not included in the required set described above.
```sh
    git clone git@github.com:oneiro-ndev/chaincode_scripts.git ~/go/src/github.com/oneiro-ndev/chaincode_scripts
```

Given that, you should be able to do `make build` to create all the tools.

The tools it creates are:

* opcodes (the code generator that ensures that all the chaincode sources use the same set of opcodes)
* chasm (the chaincode assembler)
* chfmt (the chasm formatter)
* crank (the chaincode debugger, repl, and test tool)

Once you have built the tools, you can do:

* `make scripts` to build all the validation scripts.
* `make scripttests` will test all the validation scripts based on finding files with the .crank extension in the `../chaincode_scripts` directory.
* `make scriptformat` will run the formatter over all the scripts in that directory. Note that the formatter currently has the potential to damage a file if it cannot be parsed, so you would be wise to commit an unformatted version before you run it; the safest bet is to make sure it compiles first.

### Notes on crank

Crank was originally intended to be a chaincode repl. It can definitely be used that way, but usually you'll be better off running it from a script.

If you encounter a puzzling bug, you can use the `--verbose` switch; if this is set, when crank encounters a script error it will drop into the repl so you can try to look around and maybe reset and walk through it.

The `help` and `help verbose` commands will dump some helpful text about how to use crank. Also see the readme in cmd/crank.

## Other tools

See the [README](bin/README.md) in the `./bin` directory for more information on the tools found there.
