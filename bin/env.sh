#!/bin/bash

# 3rd party repos and version to use.
export NOMS_REPO=https://github.com/attic-labs/noms.git
export NOMS_SHA=a1f990c94dcc03f9f1845d25a55e84108f1be673
export TENDERMINT_REPO=https://github.com/tendermint/tendermint.git
export TENDERMINT_VER=v0.25.0

# Port numbers.
export NODE_CHAOS_PORT=26658
export NODE_NDAU_PORT=26663
export NOMS_CHAOS_PORT=8000
export NOMS_NDAU_PORT=8001
export REDIS_CHAOS_PORT=6379
export REDIS_NDAU_PORT=6380
export TM_CHAOS_P2P_PORT=26656
export TM_CHAOS_RPC_PORT=26657
export TM_NDAU_P2P_PORT=26661
export TM_NDAU_RPC_PORT=26662

# Go source path.
GO_DIR=$(go env GOPATH)
export GO_DIR

# Repository locations.
export ATTICLABS_DIR=$GO_DIR/src/github.com/attic-labs
export NDEV_DIR=$GO_DIR/src/github.com/oneiro-ndev
export TM_DIR=$GO_DIR/src/github.com/tendermint

# Build locations.
export CHAOS_DIR=$NDEV_DIR/chaos
export COMMANDS_DIR=$NDEV_DIR/commands
export NDAU_DIR=$NDEV_DIR/ndau
export NOMS_DIR=$ATTICLABS_DIR/noms
export TENDERMINT_DIR=$TM_DIR/tendermint

# Data directories.
export NODE_DATA_DIR=~/.ndau
export NOMS_CHAOS_DATA_DIR=~/.noms-chaos
export NOMS_NDAU_DATA_DIR=~/.noms-ndau
export REDIS_CHAOS_DATA_DIR=~/.redis-chaos
export REDIS_NDAU_DATA_DIR=~/.redis-ndau
export TENDERMINT_CHAOS_DATA_DIR=~/.tendermint-chaos
export TENDERMINT_NDAU_DATA_DIR=~/.tendermint-ndau

# Command source subdirectories.  We build all tools in their respective repo roots, though.
export CHAOS_CMD=cmd/chaos
export CHAOSNODE_CMD=cmd/chaosnode
export GENERATE_CMD=cmd/generate
export GENESIS_CMD=cmd/genesis
export NDAU_CMD=cmd/ndau
export NDAUAPI_CMD=cmd/ndauapi
export NDAUNODE_CMD=cmd/ndaunode
export NOMS_CMD=cmd/noms
export TENDERMINT_CMD=cmd/tendermint

# File used by conf.sh to tell run.sh to import genesis data on first run after a reset.
export NEEDS_UPDATE_FLAG_FILE=$NODE_DATA_DIR/needsupdate
