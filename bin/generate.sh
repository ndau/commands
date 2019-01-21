#!/bin/bash

# Wrapper for `go generate` for the repo specified.  Usage: ./generate.sh <repo>
# Must have msgp installed.  See https://github.com/oneiro-ndev/chaos/README.md for details.

# Load our environment variables.
CMDBIN_DIR="$(go env GOPATH)/src/github.com/oneiro-ndev/commands/bin"
# shellcheck disable=SC1090
source "$CMDBIN_DIR"/env.sh

REPO="$1"
if [ -z "$REPO" ]; then
    echo Usage examples:
    echo "  ./generate.sh metanode"
    echo "  ./generate.sh chaos"
    echo "  ./generate.sh ndau"
    exit 1
fi

cd "$NDEV_DIR/$REPO" || exit 1
go generate ./...

# Strip out all the "msgp.WrapError" lines that don't compile in unit tests.
GEN=$(git diff-index HEAD --name-only)
for f in $GEN;
do
    if [[ "$f" == *_gen.go ]]; then
        grep -v "^.*msgp.WrapError.*$" "$f" > "$f-tmp" && mv "$f-tmp" "$f"
    fi
done
