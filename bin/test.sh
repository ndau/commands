#!/bin/bash

initialize() {
    CMDBIN_DIR="$(go env GOPATH)/src/github.com/oneiro-ndev/commands/bin"
    # shellcheck disable=SC1090
    source "$CMDBIN_DIR"/env.sh
}

test_ndau() {
    cd "$NDAU_DIR" || exit 1
    link_vendor_for_test ndau

    echo
    echo testing ndau
    go test ./...
    echo

    unlink_vendor_for_test ndau
}

test_all() {
    initialize
    test_ndau
}

test_all
echo "done."
