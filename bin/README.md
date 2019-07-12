# Oneiro ndev Tools

## Overview

This document contains information on the tools found in the bin directory that weren't covered in the documentation found at the root of this repo.  See the [README](../README.md) in the root directory for more information on the main tools needed to run a localnet before using any tools found here.

## Tools

### status

Running `./status.sh` will display the current running state of each component in the localnet node group.

### ndauapi

Once a localnet is up and running (using `./run.sh`), you can use `./ndauapi.sh` to fire up an `ndauapi` web server listening on port 3030.  Then you can run commands against it such as:

```sh
curl --get "http://localhost:3030/version" | jq .
curl --get "http://localhost:3030/block/current" | jq .
```

### populate

Once a localnet is up and running (using `./run.sh`), you can use the `./populate.sh` tool to do some basic transactoins against the ndau blockchain.  For example:

```sh
./populate.sh create
./populate.sh set-validation
./populate.sh issue
./populate.sh xfer
./populate.sh refund
./populate.sh status
```

### linkdep

This tool is useful when you want to make changes to one of our dependency projects and test it locally without first having to push it up to github.

Normally we have cloned `ndau` into `~/go/src/github.com/oneiro-ndev` and we make changes there to those projects like any other git repos.  But if you want to make changes on one of our dependency probjects, say, `metanode`, then you can use the `./linkdep.sh` tool to set that up for you.

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
