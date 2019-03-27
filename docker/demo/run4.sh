#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

IP=$("$SCRIPT_DIR"/get_ip.sh)

"$SCRIPT_DIR"/../bin/runcontainer.sh demonet-4 26664 26674 3034 "$IP:26660:26670,$IP:26661:26671,$IP:26662:26672,$IP:26663:26673" snapshot-demonet-66
