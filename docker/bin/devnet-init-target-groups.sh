#!/bin/bash
# run this script to reinitialize all the target groups.
# first time run doesn't require --force, but subsequent runs do.

# get the current directory
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# testnet has a port offset of 50, meaning all testnet nodes start at 30X5Y where X is the service and Y is the node number.
export PORT_OFFSET=0

NETWORK_NAME=devnet

START_NODE=0
END_NODE=4 # 0-4, 5 nodes

for i in $( seq $START_NODE $END_NODE ); do
    "$DIR/target-groups.sh" $NETWORK_NAME $i --force
done
