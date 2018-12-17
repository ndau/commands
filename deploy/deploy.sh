#!/bin/bash

if [ "${CIRCLE_BRANCH}" == "master" ]; then
  # Redeploy nodegroup testnet

  # Clone the automation repo master branch
  git clone git@github.com:oneiro-ndev/automation.git /root/automation

  # Remove old test net
  helm del --purge $NODE_NAMES --tls ||\
    echo "Releases: $NODE_NAMES could not be deleted" >&2

  cd /root/automation/testnet

  # create new multinode test net
  RELEASE=$RELEASE_NAME \
  ELB_SUBDOMAIN=$ELB_SUBDOMAIN \
    ./gen_node_groups.py $NODE_NUM $STARTING_PORT

else
  echo "Not deploying for non-master branch."
fi
