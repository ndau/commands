#!/bin/bash

# Stop testing later steps if an earlier step fails.
set -e

# Save the arguments for use within the functions below.
ARGS=$@

# Supported arguments.
ARG_INTEGRATION=-i # Runs integration tests in addition to regular unit tests.

initialize() {
    SETUP_DIR="$( cd "$(dirname "$0")" ; pwd -P )"
    source $SETUP_DIR/env.sh
}

test_chaos() {
    cd $CHAOS_DIR

    # Move the vendor directory back if we moved it away in the ndau build and somehow didn't
    # move it back.
    if [ -e "$CHAOS_DIR/vendor-backup" ]; then
        mv $CHAOS_DIR/vendor-backup $CHAOS_DIR/vendor
    fi

    if [[ "$ARGS" != *"$ARG_INTEGRATION"* ]]; then
        echo testing chaos
        go test ./...
    fi
}

test_ndau() {
    cd $NDAU_DIR

    # If we're trying to build with local changes, and we've used linkdep.sh, we need to move
    # away the chaos vendor directory to prevent ndau from thinking there are mismatches.
    if [ -L "$NDAU_DIR/vendor/github.com/oneiro-ndev/chaos" ]; then
        mv $CHAOS_DIR/vendor $CHAOS_DIR/vendor-backup
    fi

    if [[ "$ARGS" != *"$ARG_INTEGRATION"* ]]; then
        echo testing ndau
        go test ./...
    else
        # Integration tests require that the node group is running.
        $SETUP_DIR/run.sh

        # Sleep one more second so that tendermint has a chance to become ready.
        sleep 1

        echo testing ndau integration
        NDAU_RPC=http://localhost:$TM_NDAU_RPC_PORT
        CHAOS_RPC=http://localhost:$TM_CHAOS_RPC_PORT
        go test ./pkg/ndauapi/routes/... -integration -ndaurpc=$NDAU_RPC -chaosrpc=$CHAOS_RPC

        # We forced-ran for integration tests, so we might as well kill automatically too.
        $SETUP_DIR/kill.sh
    fi

    # Move the vendor directory back if we moved it away above.
    if [ -e "$CHAOS_DIR/vendor-backup" ]; then
        mv $CHAOS_DIR/vendor-backup $CHAOS_DIR/vendor
    fi
}

test_all() {
    initialize
    test_chaos
    test_ndau
}

test_all
echo "done."
