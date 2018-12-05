#!/bin/bash

# Tendermint version to use.
export TENDERMINT_VER=v0.25.0

# Port numbers.
export REDIS_CHAOS_PORT=6379
export REDIS_NDAU_PORT=6380
export NOMS_CHAOS_PORT=8000
export NOMS_NDAU_PORT=8001
export NODE_CHAOS_PORT=26658
export NODE_NDAU_PORT=26663
export TM_CHAOS_RPC_PORT=26657
export TM_NDAU_RPC_PORT=26662
export TM_CHAOS_P2P_PORT=26656
export TM_NDAU_P2P_PORT=26661

# Go source path.
GO_DIR=$(go env GOPATH)
export GO_DIR

# Repository locations.
export ATTICLABS_DIR=$GO_DIR/src/github.com/attic-labs
export TM_DIR=$GO_DIR/src/github.com/tendermint
export NDEV_DIR=$GO_DIR/src/github.com/oneiro-ndev

# Build locations.
export NOMS_DIR=$ATTICLABS_DIR/noms
export TENDERMINT_DIR=$TM_DIR/tendermint
export CHAOS_DIR=$NDEV_DIR/chaos
export NDAU_DIR=$NDEV_DIR/ndau
export CHAOS_GENESIS_DIR=$NDEV_DIR/chaos_genesis

# Data directories.
export REDIS_CHAOS_DATA_DIR=~/.redis-chaos
export REDIS_NDAU_DATA_DIR=~/.redis-ndau
export NOMS_CHAOS_DATA_DIR=~/.noms-chaos
export NOMS_NDAU_DATA_DIR=~/.noms-ndau
export NODE_DATA_DIR=~/.ndau
export TENDERMINT_CHAOS_DATA_DIR=~/.tendermint-chaos
export TENDERMINT_NDAU_DATA_DIR=~/.tendermint-ndau

# Command source subdirectories.  We build all tools in their respective repo roots, though.
export NOMS_CMD=cmd/noms
export TENDERMINT_CMD=cmd/tendermint
export CHAOS_CMD=cmd/chaos
export CHAOSNODE_CMD=cmd/chaosnode
export NDAU_CMD=cmd/ndau
export NDAUNODE_CMD=cmd/ndaunode
export NDAUAPI_CMD=cmd/ndauapi
export GENERATE_CMD=cmd/generate
export GENESIS_CMD=cmd/genesis

# File used by conf.sh to tell run.sh to import genesis data on first run after a reset.
export NEEDS_UPDATE_FLAG_FILE=$NODE_DATA_DIR/needsupdate
