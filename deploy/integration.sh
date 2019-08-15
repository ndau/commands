#!/bin/bash

# This is the ENTRYPOINT for the tests-container in the integration Circle CI job.

set -e

export NDAUHOME=/.ndau

echo "Configuring ndau tool..."
/image/bin/ndau conf http://$IP:26670
/image/bin/ndau conf update-from /system_accounts.toml

echo "Setting up python environment..."
python3 -m ensurepip
rm -r /usr/lib/python*/ensurepip
pip3 install --upgrade pip setuptools
pip3 install pytest pipenv
if [ ! -e /usr/bin/pip ]; then ln -s pip3 /usr/bin/pip; fi
if [ ! -e /usr/bin/python ]; then ln -sf /usr/bin/python3 /usr/bin/python; fi

echo "Running integration tests..."
cd /integration-tests
pipenv sync
pipenv run pytest -v --ip=$IP

echo "Integration script complete"
