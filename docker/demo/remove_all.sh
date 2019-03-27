#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

cd "$SCRIPT_DIR"/../bin || exit 1

./removecontainer.sh demonet-4
./removecontainer.sh demonet-3
./removecontainer.sh demonet-2
./removecontainer.sh demonet-1
./removecontainer.sh demonet-0
