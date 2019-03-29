#!/bin/bash

# Wrapper for `go generate` for the repo specified.  Usage: ./generate.sh <repo>
# Must have msgp installed.

# Load our environment variables.
CMDBIN_DIR="$(go env GOPATH)/src/github.com/oneiro-ndev/commands/bin"
# shellcheck disable=SC1090
source "$CMDBIN_DIR"/env.sh

REPO="$1"
if [ -z "$REPO" ]; then
    echo Usage examples:
    echo "  ./generate.sh metanode"
    echo "  ./generate.sh ndau"
    exit 1
fi

cd "$NDEV_DIR/$REPO" || exit 1
go generate ./...

# At this point we can do optional post-processing if ever needed.
