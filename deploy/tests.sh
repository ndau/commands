#!/bin/sh

# copy commands' vendor directory to the gopath
cp -r $GOPATH/src/github.com/oneiro-ndev/commands/vendor/* $GOPATH/src/

for oneiro_project in "$GOPATH"/src/github.com/oneiro-ndev/*; do
    (
        cd "$oneiro_project"
        pwd
        go test ./...
    )
done
