#!/bin/bash

# Stop testing later steps if an earlier step fails.
set -e

# Save the arguments for use within the functions below.
ARGS=("$@")

# Supported arguments.
ARG_INTEGRATION=-i # Runs integration tests in addition to regular unit tests.

initialize() {
    CMDBIN_DIR="$(go env GOPATH)/src/github.com/oneiro-ndev/commands/bin"
    # shellcheck disable=SC1090
    source "$CMDBIN_DIR"/env.sh
}

test_chaos() {
    cd "$CHAOS_DIR"

    chaosintegration=0
    for arg in "${ARGS[@]}"; do
        if [[ "$arg" == "$ARG_INTEGRATION" ]]; then
            chaosintegration=1
            break
        fi
    done
    if [ "$chaosintegration" != 1 ]; then
        echo testing chaos
        go test ./...
    fi
}

test_ndau() {
    cd "$NDAU_DIR"

    ndauintegration=0
    for arg in "${ARGS[@]}"; do
        if [[ "$arg" == "$ARG_INTEGRATION" ]]; then
            ndauintegration=1
            break
        fi
    done
    if [ "$ndauintegration" != 1 ]; then
        echo testing ndau
        go test ./...
    else
        # Integration tests require that the node group is running.
        "$CMDBIN_DIR"/run.sh

        # Sleep one more second so that tendermint has a chance to become ready.
        sleep 1

        echo testing ndau integration
        NDAU_RPC=http://localhost:$TM_NDAU_RPC_PORT
        CHAOS_RPC=http://localhost:$TM_CHAOS_RPC_PORT
        go test ./pkg/ndauapi/routes/... -integration -ndaurpc="$NDAU_RPC" -chaosrpc="$CHAOS_RPC"

        # We forced-ran for integration tests, so we might as well kill automatically too.
        "$CMDBIN_DIR"/kill.sh
    fi
}

test_all() {
    initialize
    test_chaos
    test_ndau
}

test_all
echo "done."
