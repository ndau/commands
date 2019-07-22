#!/usr/bin/env bash

set -e

cd "$(go env GOPATH)/src/github.com/oneiro-ndev/commands"
GOOS=linux go build ./cmd/claimer-lambda
zf="cmd/claimer-lambda/claimer-lambda.zip"
zip "$zf" claimer-lambda
rm claimer-lambda # get rid of linux binary
aws lambda update-function-code \
    --function-name "claimer-service" \
    --zip-file "fileb://$zf" \
    --publish
