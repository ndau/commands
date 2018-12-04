#!/bin/bash

# Tendermint version to use.
TENDERMINT_VER=v0.25.0

# Port numbers.
REDIS_CHAOS_PORT=6379
REDIS_NDAU_PORT=6380
NOMS_CHAOS_PORT=8000
NOMS_NDAU_PORT=8001
NODE_CHAOS_PORT=26658
NODE_NDAU_PORT=26663
TM_CHAOS_RPC_PORT=26657
TM_NDAU_RPC_PORT=26662
TM_CHAOS_P2P_PORT=26656
TM_NDAU_P2P_PORT=26661

# Go source path.
GO_DIR=`go env GOPATH`

# Repository locations.
ATTICLABS_DIR=$GO_DIR/src/github.com/attic-labs
TM_DIR=$GO_DIR/src/github.com/tendermint
NDEV_DIR=$GO_DIR/src/github.com/oneiro-ndev

# Build locations.
NOMS_DIR=$ATTICLABS_DIR/noms
TENDERMINT_DIR=$TM_DIR/tendermint
CHAOS_DIR=$NDEV_DIR/chaos
NDAU_DIR=$NDEV_DIR/ndau
CHAOS_GENESIS_DIR=$NDEV_DIR/chaos_genesis

# Data directories.
REDIS_CHAOS_DATA_DIR=~/.redis-chaos
REDIS_NDAU_DATA_DIR=~/.redis-ndau
NOMS_CHAOS_DATA_DIR=~/.noms-chaos
NOMS_NDAU_DATA_DIR=~/.noms-ndau
NODE_DATA_DIR=~/.ndau
TENDERMINT_CHAOS_DATA_DIR=~/.tendermint-chaos
TENDERMINT_NDAU_DATA_DIR=~/.tendermint-ndau

# Command source subdirectories.  We build all tools in their respective repo roots, though.
NOMS_CMD=cmd/noms
TENDERMINT_CMD=cmd/tendermint
CHAOS_CMD=cmd/chaos
CHAOSNODE_CMD=cmd/chaosnode
NDAU_CMD=cmd/ndau
NDAUNODE_CMD=cmd/ndaunode
NDAUAPI_CMD=cmd/ndauapi
GENERATE_CMD=cmd/generate
GENESIS_CMD=cmd/genesis

# File used by conf.sh to tell run.sh to import genesis data on first run after a reset.
NEEDS_UPDATE_FLAG_FILE=$NODE_DATA_DIR/needsupdate
