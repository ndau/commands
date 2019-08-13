#!/bin/sh

# This is copied into the deps container and run in the build task.

set -e # exit for any command that returns non-0

# copy commands' vendor directory to the gopath
# Go won't test from a vendor directory. But it will test if you copy the contents
# to the regular go path.
cp -r /go/src/github.com/oneiro-ndev/commands/vendor/* /go/src/

for oneiro_project in /go/src/github.com/oneiro-ndev/*; do
    (
        cd "$oneiro_project"
        pwd
        # we don't want to test commands, just binary packages
        if [ "$(basename "$(pwd)")" != commands ] && [ -d cmd ]; then
          rm -rf cmd >/dev/null 2>&1
        fi
        go build ./...
    )
done
