#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

cd "$SCRIPT_DIR"/../bin || exit 1

./removecontainer.sh easynet-4
./removecontainer.sh easynet-3
./removecontainer.sh easynet-2
./removecontainer.sh easynet-1
./removecontainer.sh easynet-0
