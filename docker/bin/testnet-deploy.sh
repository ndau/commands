#!/bin/bash
# This script is for manual testing/deployment/debugging.
# In real life CircleCI handles all this.

set -e # exit on errors

# redeploy testnet
export PERSISTENT_PEERS="1e1c860b9c3b65fd155fe63e96482f71967f7c99@_IP_:31200,940a6e3f071da7644f7f9a6b53edd99699bb9460@_IP_:31201,778a87a4537a4dd87acd37b1e5c6c458f2b414c3@_IP_:31202,2a171821c9855c85b3c50aa4eef79ad539b5d563@_IP_:31203,a11fa11b65f1c898ddf66d5b5446ec07e655e144@_IP_:31204" # _IP_ gets s/_IP_/real_ip/g 'd
STATIC_IPS="50.17.109.111 54.196.108.229"
export SHA="9"
export NETWORK_NAME="testnet"
export STATIC_IPS="50.17.109.111 54.196.108.229"
export CLUSTER_NAME="sc-node-cluster"
export SNAPSHOT_URL="https://s3.amazonaws.com/ndau-snapshots/snapshot-testnet-47.tgz"

STATIC_IPS="50.17.109.111 54.196.108.229"
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
