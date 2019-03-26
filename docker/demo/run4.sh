#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

IP=$("$SCRIPT_DIR"/get_ip.sh)

"$SCRIPT_DIR"/../bin/runcontainer.sh easynet-4 26668 26678 26669 26679 3034 "$IP:26660:26670:26661:26671,$IP:26662:26672:26663:26673,$IP:26664:26674:26665:26675,$IP:26666:26676:26667:26677" snapshot-easynet-2
