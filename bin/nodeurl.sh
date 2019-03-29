#!/bin/bash

NODE_NET=$1 # devnet or testnet
NODE_NUM=$2 # 0-based node number
CHAIN="ndau"

if [ -z "$CHAIN" ]; then
    echo nodeurl: Echo the URL for a tendermint node.
    echo Must have kubectl set up.  See integration-tests repo README.md for details.
    echo Usage:
    echo "  ./nodeurl.sh NODE_NET NODE_NUM"
    echo Example:
    echo "  ./nodeurl.sh devnet 0"
    exit 1
fi

# These commands were adapted from integration-tests/conftest.py:
ADDRESS=$(kubectl get nodes -o jsonpath='{.items[*].status.addresses[?(@.type=="ExternalIP")].address}' | tr " " "\n" | head -n 1 | tr -d "[:space:]")
PORT=$(kubectl get service --namespace default -o jsonpath='{.spec.ports[?(@.name=="rpc")].nodePort}' "$NODE_NET-$NODE_NUM-nodegroup-$CHAIN-tendermint-service")

# User can use this URL in curl commands.
echo "http://$ADDRESS:$PORT"
