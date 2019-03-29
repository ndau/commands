#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

cd "$SCRIPT_DIR" || exit 1

./run0.sh
./run1.sh
./run2.sh
./run3.sh
./run4.sh
