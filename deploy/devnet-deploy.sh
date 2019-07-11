#!/bin/bash
# This script is for manual testing/deployment/debugging.
# In real life CircleCI handles all this.
# See the commands in ./circleci/config.yml under the "configure ecs" comment for ecs-cli prereqs.

set -e # exit on errors

export SHA="42a870a"
export NETWORK_NAME="devnet"
export PERSISTENT_PEERS="88cf98107823c1ca6621a0656daeecf731870532@devnet.ndau.tech:26660,7c7a66648ca0bf152aeee0c2d358f2d9f7b18341@devnet.ndau.tech:26661,dfa5eca4f826e977379e44d19dd606c06d8f7b7c@devnet.ndau.tech:26662,595562bf12ae2ba03d522f7026d9aa653ab9707c@devnet.ndau.tech:26663,59ed8217b8ef647b7ed1439408f3de35873e65d0@devnet.ndau.tech:26664"
export CLUSTER_NAME="devnet"

for i in $( seq 0 9 ); do # automatically deploy up to 10 nodes
  if [ -f "./node-identity-$i.tgz" ]; then
    ./deploy-node.sh $i $NETWORK_NAME . &
  fi # no else break in case some nodes are updated and others are not
done
