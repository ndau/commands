#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
cd "$SCRIPT_DIR" || exit 1

# We don't get IP here like the other demo scripts since we're first and know no peers yet.

SNAPSHOT=$(./get_snapshot.sh)
IDENTITY=../../bin/ndau-snapshots/node-identity-0.tgz

../bin/runcontainer.sh \
    localnet-0 26660 26670 3030 \
    "$SNAPSHOT" \
    "$IDENTITY"
