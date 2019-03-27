#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

IP=$("$SCRIPT_DIR"/get_ip.sh)

"$SCRIPT_DIR"/../bin/runcontainer.sh demonet-3 26663 26673 3033 "$IP:26660:26670,$IP:26661:26671,$IP:26662:26672" snapshot-demonet-66 "$SCRIPT_DIR"/../../bin/ndau-snapshots/node-identity-3.tgz
