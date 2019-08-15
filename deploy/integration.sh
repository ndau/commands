#!/bin/bash

# This is the ENTRYPOINT for the tests-container in the integration Circle CI job.

set -e

ndev_dir=/go/src/github.com/oneiro-ndev
commands_dir="$ndev_dir"/commands

export GOPATH=/go
export NDAUHOME=/.ndau

echo "Configuring ndau tool..."
mkdir -p "$commands_dir"
cp /image/bin/keytool "$commands_dir"
cp /image/bin/ndau "$commands_dir"
"$commands_dir"/ndau conf http://"$IP":26670
"$commands_dir"/ndau conf update-from /system_accounts.toml

echo "Setting up python environment..."
python3 -m ensurepip
rm -r /usr/lib/python*/ensurepip
pip3 install --upgrade pip setuptools
pip3 install pytest pipenv
if [ ! -e /usr/bin/pip ]; then ln -s pip3 /usr/bin/pip; fi
if [ ! -e /usr/bin/python ]; then ln -sf /usr/bin/python3 /usr/bin/python; fi

echo "Running integration tests..."
mv /integration-tests "$ndev_dir"
cd "$ndev_dir"/integration-tests
pipenv sync
pipenv run pytest -v --ip="$IP"

echo "Integration script complete"
