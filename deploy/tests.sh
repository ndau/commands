#!/bin/sh

# copy commands' vendor directory to the gopath
# Go won't test from a vendor directory. But it will test if you copy the contents
# to the regular go path.
cp -r /go/src/github.com/oneiro-ndev/commands/vendor/* /go/src/

for oneiro_project in /go/src/github.com/oneiro-ndev/*; do
    (
        cd "$oneiro_project"
        pwd
        go test ./...
    )
done
