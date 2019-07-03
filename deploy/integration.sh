#!/bin/bash

# This is the ENTRYPOINT for the tests-container in the integration Circle CI job.

ndev_dir=/go/src/github.com/oneiro-ndev

export GOPATH=/go
export NDAUHOME=/.ndau

echo "Building ndev tools..."
cd $ndev_dir/commands
go build ./cmd/ndau
go build ./cmd/keytool

echo "Configuring ndau tool..."
./ndau conf http://$IP:26670
./ndau conf update-from /system_accounts.toml

echo "Running integration tests..."
mv /integration-tests $ndev_dir
cd $ndev_dir/integration-tests
pipenv sync
pipenv run pytest -v --ip=$IP

echo "Integration script complete"
