#!/bin/bash

errcho() { >&2 echo -e "$@"; }

# test the liveness of testnet
errcho "\nTesting tendermint RPC"
curl -s https://api.ndau.tech:3110{0..4}/status | jq ".result.node_info.id"

errcho "\nTesting ndauapi"
curl -s https://api.ndau.tech:3130{0..4}/node/health -w "\n" | jq -c

errcho "\nTesting Tendermint p2p"
nc -z 50.17.109.111 31200-31204
nc -z 54.196.108.229 31200-31204
