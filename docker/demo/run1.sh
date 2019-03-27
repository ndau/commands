#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

IP=$("$SCRIPT_DIR"/get_ip.sh)

"$SCRIPT_DIR"/../bin/runcontainer.sh demonet-1 26661 26671 3031 "$IP:26660:26670" snapshot-demonet-66 "$SCRIPT_DIR"/../../bin/ndau-snapshots/node-identity-1.tgz
