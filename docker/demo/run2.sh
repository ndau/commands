#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

IP=$("$SCRIPT_DIR"/get_ip.sh)

"$SCRIPT_DIR"/../bin/runcontainer.sh demonet-2 26662 26672 3032 "$IP:26660:26670,$IP:26661:26671" snapshot-demonet-66 "$SCRIPT_DIR"/../../bin/ndau-snapshots/node-identity-2.tgz
