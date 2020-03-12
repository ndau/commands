#!/bin/bash

# 3rd party repos and version to use.
export NOMS_REPO=https://github.com/attic-labs/noms.git
export TENDERMINT_REPO=https://github.com/tendermint/tendermint.git
export TENDERMINT_VER=v0.32.6

# For multi-node support.
export MAX_NODE_COUNT=10

# Port numbers.
# For example, here are the noms ports for each node:
#   node 0: 8000
#   node 1: 8001
#   node N: 8000 + N
# Therefore, we must leave room for MAX_NODE_COUNT values in each port number space.
export NODE_PORT=26650
export NOMS_PORT=8000
export REDIS_PORT=6379
export TM_P2P_PORT=26660
export TM_RPC_PORT=26670
export NDAUAPI_PORT=3030
export CLAIMER_PORT=3000

# Redis can't have more clients than 32 less than the ulimit amount we use.
export ULIMIT_AMOUNT=1024
export REDIS_MAX_CLIENTS=$((ULIMIT_AMOUNT - 32))

export GO111MODULE=off
# Go source path.
GO_DIR=$(go env GOPATH)
if [[ "$GO_DIR" == *":"* ]]; then
    echo Multiple Go paths not supported
    exit 1
fi
export GO_DIR

# Prefix of node names in the network, e.g. localnet-0, localnet-1, ...
export MONIKER_PREFIX=localnet

# Repository locations.
export ATTICLABS_DIR="$GO_DIR"/src/github.com/attic-labs
export NDEV_SUBDIR=github.com/ndau
export NDEV_DIR="$GO_DIR/src/$NDEV_SUBDIR"
export TM_DIR="$GO_DIR"/src/github.com/tendermint

# Build locations.
export COMMANDS_DIR="$NDEV_DIR"/commands
export NDAU_DIR="$NDEV_DIR"/ndau
export NOMS_DIR="$ATTICLABS_DIR"/noms
export TENDERMINT_DIR="$TM_DIR"/tendermint

# Localnet directories common to all nodes.  The data dir is deleted and recreated by reset.sh.
export LOCALNET_DIR=~/.localnet
export ROOT_DATA_DIR="$LOCALNET_DIR"/data
export GENESIS_FILES_DIR="$LOCALNET_DIR"/genesis_files
export SYSTEM_VARS_TOML="$GENESIS_FILES_DIR/system_vars.toml"
export SYSTEM_ACCOUNTS_TOML="$GENESIS_FILES_DIR/system_accounts.toml"

# Data directories.  These get "-$node_num" appended to them when they are used.
export NODE_DATA_DIR="$ROOT_DATA_DIR"/ndau
export NOMS_NDAU_DATA_DIR="$ROOT_DATA_DIR"/noms-ndau
export REDIS_NDAU_DATA_DIR="$ROOT_DATA_DIR"/redis-ndau
export TENDERMINT_NDAU_DATA_DIR="$ROOT_DATA_DIR"/tendermint-ndau

# Command source subdirectories.  We build all tools in their respective repo roots, though.
export GENERATE_CMD=cmd/generate
export NDAU_CMD=cmd/ndau
export NDAUAPI_CMD=cmd/ndauapi
export NDAUNODE_CMD=cmd/ndaunode
export NDSH_CMD=cmd/ndsh
export NOMS_CMD=cmd/noms
export TENDERMINT_CMD=cmd/tendermint
export ETL_CMD=cmd/etl
export KEYTOOL_CMD=cmd/keytool
export PROCMON_CMD=cmd/procmon
export CLAIMER_CMD=cmd/claimer

# The localnet data directory is created by setup.sh and is not modified by any other script.
# We use it for storing meta info about the local nodes we manage.
export NODE_COUNT_FILE="$LOCALNET_DIR"/node_count
if [ -f "$NODE_COUNT_FILE" ]; then
    # cat can't fail in this situation
    # shellcheck disable=SC2155
    export NODE_COUNT=$(cat "$NODE_COUNT_FILE")
    export HIGH_NODE_NUM=$((NODE_COUNT - 1))
fi
export CHAIN_ID_FILE="$LOCALNET_DIR"/chain_id
if [ -f "$CHAIN_ID_FILE" ]; then
    # cat can't fail in this situation
    # shellcheck disable=SC2155
    export CHAIN_ID=$(cat "$CHAIN_ID_FILE")
fi

# Join array elements together by a delimiter.  e.g. `join_by , (a b c)` returns "a,b,c".
join_by() { local IFS="$1"; shift; echo "$*"; }

# Common steps to be done prior to linking for building or testing.
prepare_for_linking() {
    REPO="$1"

    # Ensure the given repo source dir exists to be linked to or from.
    SOURCE_DIR="$NDEV_DIR/$REPO"
    if [ ! -d "$SOURCE_DIR" ]; then
        echo Must clone "$REPO" into "$SOURCE_DIR" first
        exit 1
    fi

    # Ensure we start with a clean setup before linking, in case we linkdep'd for building and
    # testing (which would create a circular vendor referencing situation).  Only one link kind
    # can be active at any given time.
    unlink_vendor_for_build "$REPO"
    unlink_vendor_for_test "$REPO"
}

# Move away the commands vendor subdirectory for the given repo.
backup_vendor_subdir() {
    REPO="$1"

    VENDOR_DIR="$COMMANDS_DIR/vendor/$NDEV_SUBDIR/$REPO"
    if [ -d "$VENDOR_DIR" ] && [ ! -L "$VENDOR_DIR" ]; then
        echo moving away "$REPO" subdirectory in vendor directory
        rm -rf "$VENDOR_DIR-backup"
        mv "$VENDOR_DIR" "$VENDOR_DIR-backup"
    fi
}

# Put back the commands vendor subdirectory for the given repo.
restore_vendor_subdir() {
    REPO="$1"

    VENDOR_DIR="$COMMANDS_DIR/vendor/$NDEV_SUBDIR/$REPO"
    if [ -d "$VENDOR_DIR-backup" ]; then
        echo moving back "$REPO" subdirectory in vendor directory
        rm -rf "$VENDOR_DIR"
        mv "$VENDOR_DIR-backup" "$VENDOR_DIR"
    fi
}

# Make the commands vendor subdirectory for a given repo point to the local copy of that repo.
link_vendor_for_build() {
    REPO="$1"

    prepare_for_linking "$REPO"

    # Move the vendor directory away before linking from it.
    backup_vendor_subdir "$REPO"

    echo linking vendor directory to "$REPO"
    VENDOR_DIR="$COMMANDS_DIR/vendor/$NDEV_SUBDIR/$REPO"
    rm -rf "$VENDOR_DIR"
    ln -s "$NDEV_DIR/$REPO" "$VENDOR_DIR"
}

# Undo what link_vendor_for_build() did for a given repo.
unlink_vendor_for_build() {
    REPO="$1"

    restore_vendor_subdir "$REPO"
}

# Make the vendor directory for a given repo point to the commands vendor directory.
link_vendor_for_test() {
    REPO="$1"

    prepare_for_linking "$REPO"

    # Allow go test to find all the dependencies in the commands vendor directory.
    echo linking vendor directory in "$REPO"
    SOURCE_VENDOR="$NDEV_DIR/$REPO/vendor"
    rm -rf "$SOURCE_VENDOR"
    ln -s "$COMMANDS_DIR"/vendor "$SOURCE_VENDOR"

    # Avoid circular references when testing the given repo.
    backup_vendor_subdir "$REPO"
}

# Undo what link_vendor_for_test() did for a given repo.
unlink_vendor_for_test() {
    REPO="$1"

    restore_vendor_subdir "$REPO"

    SOURCE_VENDOR="$NDEV_DIR/$REPO/vendor"
    # Ensure there is no vendor file or symbolic link here.
    # Also, in case someone has the old glide vendor directory lying around.
    if [ -e "$SOURCE_VENDOR" ]; then
        echo removing vendor from "$REPO"
        rm -rf "$SOURCE_VENDOR"
    fi
}
