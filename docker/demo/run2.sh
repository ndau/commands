#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

source "$SCRIPT_DIR"/get_ip.sh

"$SCRIPT_DIR"/../bin/runcontainer.sh easynet-2 26664 26674 26665 26675 3032 "$IP:26660:26670:26661:26671,$IP:26662:26672:26663:26673" snapshot-easynet-2 "$SCRIPT_DIR"/../../bin/ndau-snapshots/node-identity-2.tgz
