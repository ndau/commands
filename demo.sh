#!/bin/bash

set -v

REPO=$(git rev-parse --show-toplevel)
cd "$REPO" || exit 1

# clear logs
rm -f bin/*.log

# reset and rebuild
bin/reset.sh
bin/setup.sh

shutdown() {
    bin/kill.sh
    exit
}

eshutdown() {
    bin/kill.sh
    exit 1
}

trap shutdown SIGINT SIGTERM
trap eshutdown ERR

# start the nodes
bin/run.sh

# create an ndau account and give it some money
./ndau account new demo
./ndau account claim demo
./ndau rfe 10 demo

# create a chaos account associated with the ndau account
./chaos id new demo
./chaos id copy-keys-from demo

# create some transactions
./chaos set demo -k key -v val
./chaos set demo -k key -v value

# verify that the chaos chain queried the ndau chain for sidechain transactions
grep sidechain bin/ndau_tm.log

# shutdown gracefully
bin/kill.sh

# SUCCESS! To restart your nodes, just run
#   bin/run.sh
