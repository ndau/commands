#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

"$SCRIPT_DIR"/../bin/runcontainer.sh demonet-0 26660 26670 3030 "" snapshot-demonet-66 "$SCRIPT_DIR"/../../bin/ndau-snapshots/node-identity-0.tgz
