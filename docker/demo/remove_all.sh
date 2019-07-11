#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

cd "$SCRIPT_DIR"/../bin || exit 1

./removecontainer.sh localnet-5
./removecontainer.sh localnet-4
./removecontainer.sh localnet-3
./removecontainer.sh localnet-2
./removecontainer.sh localnet-1
./removecontainer.sh localnet-0
