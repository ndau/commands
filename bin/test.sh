#!/bin/bash

# Save the arguments for use within the functions below.
ARGS=("$@")

ARG_INTEGRATION=-i # Pass this in to run integration tests as opposed to unit tests.
RUN_INTEGRATION=0  # Default to running unit tests if the argument is missing.

initialize() {
    CMDBIN_DIR="$(go env GOPATH)/src/github.com/oneiro-ndev/commands/bin"
    # shellcheck disable=SC1090
    source "$CMDBIN_DIR"/env.sh

    # Process command line arguments.
    for arg in "${ARGS[@]}"; do
        if [[ "$arg" == "$ARG_INTEGRATION" ]]; then
            RUN_INTEGRATION=1
            break
        fi
    done
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

    if [ "$RUN_INTEGRATION" == 0 ]; then
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

    if [ "$RUN_INTEGRATION" == 0 ]; then
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

        # We use the ports of the 0'th node, even in a multi-node localnet.
        chaos_rpc_port="$TM_RPC_PORT"
        ndau_rpc_port=$(expr "$TM_RPC_PORT" + 1)

        chaos_rpc=http://localhost:"$chaos_rpc_port"
        ndau_rpc=http://localhost:"$ndau_rpc_port"
        go test ./pkg/ndauapi/routes/... -integration -ndaurpc="$ndau_rpc" -chaosrpc="$chaos_rpc"

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
