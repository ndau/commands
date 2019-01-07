#!/bin/bash

# 3rd party repos and version to use.
export NOMS_REPO=https://github.com/oneiro-ndev/noms.git
export TENDERMINT_REPO=https://github.com/tendermint/tendermint.git
export TENDERMINT_VER=v0.25.0

# For multi-node support.
export MAX_NODE_COUNT=5

# Port numbers.  They come in chaos-ndau pairs, one pair per node.
# For example, here are the noms ports for each node:
#   node 0:
#     chaos: 8000
#     ndau : 8001
#   node 1:
#     chaos: 8002
#     ndau : 8003
#   node N:
#     chaos: NOMS_PORT + 2 * N
#     ndau : NOMS_PORT + 2 * N + 1
# Therefore, we must leave room for 2 x MAX_NODE_COUNT values in each port number space.
export NODE_PORT=26650
export NOMS_PORT=8000
export REDIS_PORT=6379
export TM_P2P_PORT=26660
export TM_RPC_PORT=26670

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

# Localnet directories common to all nodes.
export LOCALNET_DIR=~/.localnet
export ROOT_DATA_DIR="$LOCALNET_DIR"/data

# Data directories.  These get "-$node_num" appended to them when they are used.
export NODE_DATA_DIR="$ROOT_DATA_DIR"/ndau
export NOMS_CHAOS_DATA_DIR="$ROOT_DATA_DIR"/noms-chaos
export NOMS_NDAU_DATA_DIR="$ROOT_DATA_DIR"/noms-ndau
export REDIS_CHAOS_DATA_DIR="$ROOT_DATA_DIR"/redis-chaos
export REDIS_NDAU_DATA_DIR="$ROOT_DATA_DIR"/redis-ndau
export TENDERMINT_CHAOS_DATA_DIR="$ROOT_DATA_DIR"/tendermint-chaos
export TENDERMINT_NDAU_DATA_DIR="$ROOT_DATA_DIR"/tendermint-ndau

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

# The localnet data directory is created by setup.sh and is not modified by any other script.
# We use it for storing meta info about the local nodes we manage.
export NODE_COUNT_FILE="$LOCALNET_DIR"/node_count
if [ -e "$NODE_COUNT_FILE" ]; then
    export NODE_COUNT=$(cat "$NODE_COUNT_FILE")
    export HIGH_NODE_NUM=$(expr "$NODE_COUNT" - 1)
fi

# File used by conf.sh to tell run.sh to import genesis data on first run after a reset.
export NEEDS_UPDATE_FLAG_FILE="$ROOT_DATA_DIR"/needs_update
