#!/bin/bash

# This is the ENTRYPOINT for the tests-container in the integration Circle CI job.

export NDAUHOME=/.ndau

echo "Configuring ndau tool..."
./ndau conf http://$IP:26670
./ndau conf update-from /system_accounts.toml

echo "Running integration tests..."
cd /integration-tests
pipenv sync
pipenv run pytest -v --ip=$IP

echo "Integration script complete"
