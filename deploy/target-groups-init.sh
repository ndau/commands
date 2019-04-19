#!/bin/bash
# run this script to reinitialize all the target groups.

# get the current directory
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

NETWORK_NAME=${1:-devnet}

START_NODE=0
END_NODE=4 # 0-4, 5 nodes

for i in $( seq $START_NODE $END_NODE ); do
    "$DIR/target-groups.sh" $NETWORK_NAME $i --force
done
