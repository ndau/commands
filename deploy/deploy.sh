#!/bin/bash

if [ "${CIRCLE_BRANCH}" == "prgn-fix-commands" ]; then #"master" ]; then
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

  # Run integration tests
  # get address and port of devnet0 RPC
  NODE_IP_ADDRESS=$(kubectl get nodes -o jsonpath='{.items[*].status.addresses[?(@.type=="ExternalIP")].address}' | cut -d " " -f1)
  NODE_PORT_0=$(kubectl get service --namespace default -o jsonpath='{.spec.ports[?(@.name=="rpc")].nodePort}' devnet-0-nodegroup-ndau-tendermint-service)
  # get address and port of devnet1 RPC
  NODE_PORT_1=$(kubectl get service --namespace default -o jsonpath='{.spec.ports[?(@.name=="rpc")].nodePort}' devnet-1-nodegroup-ndau-tendermint-service)

  URL_0=http://$NODE_IP_ADDRESS:$NODE_PORT_0/node/status
  URL_1=http://$NODE_IP_ADDRESS:$NODE_PORT_1/node/status

  # curl retry options
  CURL_CONNECT_TIMEOUT=5  # how long each try waits
  CURL_RETRY_MAX=50       # retry this many times
  CURL_RETRY_TOTAL=1000   # arbitrary high number, it will timeout first.
  CURL_RETRY_DELAY=10     # try every X seconds
  CURL_TOTAL_TIMEOUT=420  # total time before it fails (420s=7min)

  echo "Trying to connect to $URL_0"
  # curl until devnet-0 RPC is up and running, or CURL_TOTAL_TIMEOUT passes
  curl --connect-timeout $CURL_CONNECT_TIMEOUT \
    --max-time $CURL_RETRY_MAX \
    --retry $CURL_RETRY_TOTAL \
    --retry-delay $CURL_RETRY_DELAY \
    --retry-max-time $CURL_TOTAL_TIMEOUT \
    $URL_0

  echo "Trying to connect to $URL_1"
  # curl until devnet-1 RPC is up and running, or CURL_TOTAL_TIMEOUT passes
  curl --connect-timeout $CURL_CONNECT_TIMEOUT \
    --max-time $CURL_RETRY_MAX \
    --retry $CURL_RETRY_TOTAL \
    --retry-delay $CURL_RETRY_DELAY \
    --retry-max-time $CURL_TOTAL_TIMEOUT \
    $URL_1

  # ensure go path location
  mkdir -p $GOPATH/src/github.com/oneiro-ndev
  cd $GOPATH/src/github.com/oneiro-ndev

  # clone integration tests
  git clone git@github.com:oneiro-ndev/chaos-integration-tests.git -b jsg-unified-nodes-update
  cd chaos-integration-tests

  # run tests
  pipenv sync
  pipenv run pytest -v --run_kub src/meta_test_ndau.py src/single_validator_test_ndau.py

else
  echo "Not deploying for non-master branch."
fi
