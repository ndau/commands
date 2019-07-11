#!/bin/bash

errcho() { >&2 echo -e "$@"; }

# test the liveness of devnet
errcho "\nTesting tendermint RPC"
curl -s https://devnet.ndau.tech:2667{0..4}/status | jq ".result.node_info.id"

errcho "\nTesting ndauapi"
curl -s https://devnet.ndau.tech:303{0..4}/node/health -w "\n" | jq -c

errcho "\nTesting Tendermint p2p"
nc -z 54.183.173.106 26660-26664
nc -z 54.153.92.31 26660-26664
