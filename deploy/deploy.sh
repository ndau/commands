#!/bin/bash

if [ "${CIRCLE_BRANCH}" == "master" ]; then
  # Redeploy nodegroup testnet
  # get current container versions
  TM_CONTAINER_VERSION=$(grep -e 'TM_VERSION_TAG' /root/chaos/tm-docker/Dockerfile -m 1 | cut -f3 -d ' ')
  NDAU_CONTAINER_VERSION=$(git rev-parse --short "$CIRCLE_SHA1")
  CHAOS_CONTAINER_VERSION=$(git rev-parse --short "$CIRCLE_SHA1")

  # Clone the automation repo master branch
  git clone git@github.com:oneiro-ndev/automation.git /root/automation

  # Remove old test net
  KUBECONFIG=/root/kubeconfig \
    helm del --purge $NODE_NAMES --tls ||\
      echo "Releases: $NODE_NAMES could not be deleted" >&2

  cd /root/automation/testnet

  # create new multinode test net
  KUBECONFIG=/root/kubeconfig \
  CHAOS_NOMS_TAG=$NOMS_CONTAINER_VERSION \
  CHAOS_TM_TAG=$TM_CONTAINER_VERSION \
  CHAOSNODE_TAG=$CHAOS_CONTAINER_VERSION \
  NDAU_NOMS_TAG=$NOMS_CONTAINER_VERSION \
  NDAU_TM_TAG=$TM_CONTAINER_VERSION \
  NDAUNODE_TAG=$NDAU_CONTAINER_VERSION \
  RELEASE=$RELEASE_NAME ELB_SUBDOMAIN=$ELB_SUBDOMAIN \
    ./gen_node_groups.py $NODE_NUM $STARTING_PORT

  # Run integration tests
  # get address and port of devnet0 RPC
  NODE_IP_ADDRESS=$(kubectl get nodes -o jsonpath='{.items[*].status.addresses[?(@.type=="ExternalIP")].address}' | cut -d " " -f1)
  NODE_PORT0=$(kubectl get service --namespace default -o jsonpath='{.spec.ports[?(@.name=="rpc")].nodePort}' devnet-0-nodegroup-ndau-tendermint-service)
  # get address and port of devnet1 RPC
  NODE_PORT1=$(kubectl get service --namespace default -o jsonpath='{.spec.ports[?(@.name=="rpc")].nodePort}' devnet-1-nodegroup-ndau-tendermint-service)
  echo $NODE_IP_ADDRESS:$NODE_PORT0
  # loop and curl until devnet0 RPC is up and running, or 50 times
  for i in {1..50}; do  if curl -v http://$NODE_IP_ADDRESS:$NODE_PORT0/status --connect-timeout 5; then break; fi; echo $i; if [ "$i" == "50" ]; then exit 1; fi; sleep 5; done
  # loop and curl until devnet1 RPC is up and running, or 50 times
  for i in {1..50}; do  if curl -v http://$NODE_IP_ADDRESS:$NODE_PORT1/status --connect-timeout 5; then break; fi; echo $i; if [ "$i" == "50" ]; then exit 1; fi; sleep 5; done
  mkdir -p $GOPATH/src/github.com/oneiro-ndev
  cd $GOPATH/src/github.com/oneiro-ndev
  echo -e "$gomu" > ~/.ssh/id_rsa
  chmod 600 ~/.ssh/id_rsa
  git clone git@github.com:oneiro-ndev/chaos-integration-tests.git -b jsg-unified-nodes-update
  cd chaos-integration-tests
  pipenv sync
  pipenv run pytest -v --run_kub src/meta_test_ndau.py src/single_validator_test_ndau.py

fi
