#!/bin/bash
# This script is for manual testing/deployment/debugging.
# In real life CircleCI handles all this.
# See the commands in ./circleci/config.yml under the "configure ecs" comment for ecs-cli prereqs.

set -e # exit on errors

export SHA="36f7c63"
export NETWORK_NAME="devnet"
export PERSISTENT_PEERS="c8e98c9e80b497b79b5f8a09190f35472444556b@devnet.ndau.tech:26660,baeef050d0fe5286360e55d6a37dda916d491ff8@devnet.ndau.tech:26661,a354751ca164d047b83760843b742052b9d0dd47@devnet.ndau.tech:26662,43396b8eade8f0977088330fe27d3f7f548761a2@devnet.ndau.tech:26663,2e189edb1e4351cabf2ea5e00491cf9eb0e7278e@devnet.ndau.tech:26664"
export CLUSTER_NAME="devnet"

for i in $( seq 0 9 ); do # automatically deploy up to 10 nodes
  if [ -f "./node-identity-$i.tgz" ]; then
    ./deploy-node.sh $i $NETWORK_NAME .
  fi # no else break in case some nodes are updated and others are not
done
