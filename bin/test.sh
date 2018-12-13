#!/bin/bash

# Save the arguments for use within the functions below.
ARGS=("$@")

# Supported arguments.
ARG_INTEGRATION=-i # Runs integration tests in addition to regular unit tests.

initialize() {
    CMDBIN_DIR="$(go env GOPATH)/src/github.com/oneiro-ndev/commands/bin"
    # shellcheck disable=SC1090
    source "$CMDBIN_DIR"/env.sh
}

link_vendor_in_cwd() {
    REPO="$1"

    unlink_vendor_in_cwd "$REPO"

    # Allow go test to find all the dependencies in the commands vendor directory.
    echo linking vendor directory in "$REPO"
    ln -s "$COMMANDS_DIR"/vendor vendor

    # Avoid circular references when testing the given repo.
    DIR="$COMMANDS_DIR"/vendor/github.com/oneiro-ndev/"$REPO"
    if [ -e "$DIR" ]; then
        echo moving away "$REPO" subdirectory in vendor directory
        rm -rf "$DIR-backup"
        mv "$DIR" "$DIR-backup"
    fi
}

unlink_vendor_in_cwd() {
    REPO="$1"

    DIR="$COMMANDS_DIR"/vendor/github.com/oneiro-ndev/"$REPO"
    if [ -e "$DIR-backup" ]; then
        echo moving back "$REPO" subdirectory in vendor directory
        rm -rf "$DIR"
        mv "$DIR-backup" "$DIR"
    fi

    if [ -e vendor ]; then
        echo removing vendor from "$REPO"
        if [ -d vendor ]; then
            # In case someone has the old glide vendor directory lying around.
            rm -rf vendor
        else
            # Ensure there is no vendor file or symbolic link here.
            rm -f vendor
        fi
    fi
}

test_chaos() {
    cd "$CHAOS_DIR"
    link_vendor_in_cwd chaos

    chaosintegration=0
    for arg in "${ARGS[@]}"; do
        if [[ "$arg" == "$ARG_INTEGRATION" ]]; then
            chaosintegration=1
            break
        fi
    done
    if [ "$chaosintegration" != 1 ]; then
        echo
        echo testing chaos
        go test ./...
        echo
    fi

    unlink_vendor_in_cwd chaos
}

test_ndau() {
    cd "$NDAU_DIR"
    link_vendor_in_cwd ndau

    ndauintegration=0
    for arg in "${ARGS[@]}"; do
        if [[ "$arg" == "$ARG_INTEGRATION" ]]; then
            ndauintegration=1
            break
        fi
    done
    if [ "$ndauintegration" != 1 ]; then
        echo
        echo testing ndau
        go test ./...
        echo
    else
        # Integration tests require that the node group is running.
        "$CMDBIN_DIR"/run.sh

        # Sleep one more second so that tendermint has a chance to become ready.
        sleep 1

        echo
        echo testing ndau integration
        NDAU_RPC=http://localhost:$TM_NDAU_RPC_PORT
        CHAOS_RPC=http://localhost:$TM_CHAOS_RPC_PORT
        go test ./pkg/ndauapi/routes/... -integration -ndaurpc="$NDAU_RPC" -chaosrpc="$CHAOS_RPC"
        echo

        # We forced-ran for integration tests, so we might as well kill automatically too.
        "$CMDBIN_DIR"/kill.sh
    fi

    unlink_vendor_in_cwd ndau
}

test_all() {
    initialize
    test_chaos
    test_ndau
}

test_all
echo "done."
