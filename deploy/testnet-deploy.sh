#!/bin/bash
# This script is for manual testing/deployment/debugging.
# In real life CircleCI handles all this.

set -e # exit on errors

# redeploy testnet
export PERSISTENT_PEERS="81ce8ead16c3424e46fd4fa162482ba783a333ce@p2p.ndau.tech:31200,12f375c1216dc4e64cb5560cc0f4e8a0ddc560ce@p2p.ndau.tech:31201,95ad9834ad1bee57df601b4c63660735a2400eb6@p2p.ndau.tech:31202,f3143a8eb17e0fe6b67e8d9048678dca4e57c3a3@p2p.ndau.tech:31203,eb9870620c46cd3608b875a5daa6c64380b44aaf@p2p.ndau.tech:31204"
export SHA="9"
export NETWORK_NAME="testnet"
export CLUSTER_NAME="sc-node-cluster"
export SNAPSHOT_URL="https://s3.amazonaws.com/ndau-snapshots"
export SNAPSHOT_NAME="snapshot-testnet-47"

for i in $( seq 0 9 ); do # automatically deploy up to 10 nodes
  if [ -f "./node-identity-$i.tgz" ]; then
    ./deploy-node.sh $i $NETWORK_NAME .
  fi # no else break in case some nodes are updated and others are not
done
