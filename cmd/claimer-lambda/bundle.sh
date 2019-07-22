#!/usr/bin/env bash

set -e

cd "$(go env GOPATH)/src/github.com/oneiro-ndev/commands"
GOOS=linux go build ./cmd/claimer-lambda
zip cmd/claimer-lambda/claimer-lambda.zip claimer-lambda
