#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

"$SCRIPT_DIR"/../bin/runcontainer.sh easynet-0 26660 26670 26661 26671 3030 "" snapshot-easynet-2 "$SCRIPT_DIR"/../../bin/ndau-snapshots/node-identity-0.tgz
