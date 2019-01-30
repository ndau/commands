#!/bin/bash

NET=$1   # devnet or testnet
CHAIN=$2 # ndau or chaos

if [ -z "$NET" ]; then
    echo nodenet: Echo the URL for a node net.
    echo Must have kubectl set up.  See integration-tests repo README.md for details.
    echo Usage:
    echo "  ./nodenet.sh NET CHAIN"
    echo Example:
    echo "  ./nodenet.sh devnet ndau"
    exit 1
fi

echo http://$(kubectl get nodes -o jsonpath='{.items[*].status.addresses[?(@.type=="ExternalIP")].address}' | tr " " "\n" | head -n 1 | tr -d "[:space:]"):$(kubectl get service --namespace default -o jsonpath='{.spec.ports[?(@.name=="rpc")].nodePort}' $NET-0-nodegroup-$CHAIN-tendermint-service)
