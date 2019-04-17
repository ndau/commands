#!/bin/bash
# This script is for manual testing/deployment/debugging.
# In real life CircleCI handles all this.

set -e # exit on errors

export SHA="9"
export NETWORK_NAME="devnet"
export PERSISTENT_PEERS="93ae89195d0b2f798c91cd7d1fc96062d4f72791@_IP_:30200,a02408086b9688e2f8bda083e1940d010e371627@_IP_:30201,950efa3bee5f90d442d67751969ec976eb529fdd@_IP_:30202,a6bdfb3f423ec31b6fd8ea72e60e901095d0649c@_IP_:30203,c80cdd1daeb7f65c15258831fc1dab4aacf98069@_IP_:30204" # _IP_ gets s/_IP_/real_ip/g 'd
export STATIC_IPS="50.17.109.111 54.196.108.229"
export CLUSTER="sc-node-cluster"
export SNAPSHOT_URL="https://s3.amazonaws.com/ndau-snapshots/snapshot-testnet-47.tgz"

# get the p2p load balancer's ip address
PEERS=()
for IP in $STATIC_IPS; do
  PEERS+=($(sed "s/_IP_/$IP/g" <<< $PERSISTENT_PEERS))
done
export PERSISTENT_PEERS=$(IFS=,; echo "${PEERS[*]}")

for i in $( seq 0 9 ); do # automatically deploy up to 10 nodes
  if [ -f "./node-identity-$i.tgz" ]; then
    ./deploy-node.sh $i $NETWORK_NAME .
  fi # no else break in case some nodes are updated and others are not
done
