#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

SNAPSHOT_BASE_URL="https://s3.amazonaws.com/ndau-snapshots"
INTERNAL_CHAOS_P2P=26660
INTERNAL_CHAOS_RPC=26670
INTERNAL_NDAU_P2P=26661
INTERNAL_NDAU_RPC=26671
INTERNAL_NDAUAPI=3030

if [ -z "$1" ] || [ -z "$2" ] || [ -z "$3" ] || [ -z "$4" ] || [ -z "$5" ] || [ -z "$6" ] || \
       [ -z "$7" ] || [ -z "$8" ] || [ -z "$9" ]
then
    echo "Usage:"
    echo "  ./runcontainer.sh" \
         "CONTAINER CHAOS_P2P CHAOS_RPC NDAU_P2P NDAU_RPC NDAUAPI PEER_IP PEER_RPC SNAPSHOT"
    echo
    echo "Arguments:"
    echo "  CONTAINER   Name to give to the container to run"
    echo "  CHAOS_P2P   External port to map to the internal P2P port for the chaos chain"
    echo "  CHAOS_RPC   External port to map to the internal RPC port for the chaos chain"
    echo "  NDAU_P2P    External port to map to the internal P2P port for the ndau chain"
    echo "  NDAU_RPC    External port to map to the internal RPC port for the ndau chain"
    echo "  NDAUAPI     External port to map to the internal ndauapi port"
    echo "  PEER_IP     IP of an ndau chain peer on the network to join"
    echo "  PEER_RPC    RPC port of the ndau chain peer"
    echo "  SNAPSHOT    Name of the snapshot to use as a starting point for the node group"
    exit 1
fi
CONTAINER="$1"
CHAOS_P2P="$2"
CHAOS_RPC="$3"
NDAU_P2P="$4"
NDAU_RPC="$5"
NDAUAPI="$6"
PEER_IP="$7"
PEER_RPC="$8"
SNAPSHOT="$9"

echo "Container: $CONTAINER"

if [ ! -z "$(docker container ls -a -q -f name=$CONTAINER)" ]; then
    echo "Container already exists: $CONTAINER"
    echo "Use restartcontainer.sh to restart it, or use removecontainer.sh to remove it first"
    exit 1
fi    

echo "chaos P2P port: $CHAOS_P2P"
echo "chaos RPC port: $CHAOS_RPC"
echo "ndau P2P port: $NDAU_P2P"
echo "ndau RPC port: $NDAU_RPC"
echo "ndauapi port: $NDAUAPI"

test_port() {
    ip="$1"
    port="$2"

    $(nc -G 1 -z "$ip" "$port" 2>/dev/null)
    if [ "$?" = 0 ]; then
        echo "Port at $ip:$port is already in use"
        exit 1
    fi
}

test_port localhost "$CHAOS_P2P"
test_port localhost "$CHAOS_RPC"
test_port localhost "$NDAU_P2P"
test_port localhost "$NDAU_RPC"
test_port localhost "$NDAUAPI"

test_peer() {
    chain="$1"
    ip="$2"
    port="$3"

    echo "Getting $chain peer info..."
    PEER_ID=$(curl -s --connect-timeout 5 "http://$ip:$port/status" | jq -r .result.node_info.id)
    if [ -z "$PEER_ID" ]; then
        echo "Could not get $chain peer id"
        exit 1
    fi
    echo "$chain peer id: $PEER_ID"

    PEER_P2P=$((port - 1))
    echo "Testing connection to $chain peer..."
    $(nc -G 5 -z "$ip" "$PEER_P2P")
    if [ "$?" != 0 ]; then
        echo "Could not reach $chain peer"
        exit 1
    fi
}

test_peer ndau "$PEER_IP" "$PEER_RPC"
NDAU_PEER_ID="$PEER_ID"
NDAU_PEER_P2P="$PEER_P2P"

test_peer chaos "$PEER_IP" $((PEER_RPC - 2))
CHAOS_PEER_ID="$PEER_ID"
CHAOS_PEER_P2P="$PEER_P2P"

# Stop the container if it's running.  We can't run or restart it otherwise.
"$SCRIPT_DIR"/stopcontainer.sh "$CONTAINER"

echo Silencing warning about Transparent Huge Pages when redis-server runs...
docker run --rm -it --privileged --pid=host debian nsenter -t 1 -m -u -n -i \
       sh -c 'echo never > /sys/kernel/mm/transparent_hugepage/enabled'
docker run --rm -it --privileged --pid=host debian nsenter -t 1 -m -u -n -i \
       sh -c 'echo never > /sys/kernel/mm/transparent_hugepage/defrag'

echo "Creating container..."
# Some notes about the params to the run command:
# - Using --sysctl silences a warning about TCP backlog when redis runs.
# - Set your own HONEYCOMB_* env vars ahead of time to enable honeycomb logging.
docker run -d \
       -p "$CHAOS_P2P":"$INTERNAL_CHAOS_P2P" \
       -p "$CHAOS_RPC":"$INTERNAL_CHAOS_RPC" \
       -p "$NDAU_P2P":"$INTERNAL_NDAU_P2P" \
       -p "$NDAU_RPC":"$INTERNAL_NDAU_RPC" \
       -p "$NDAUAPI":"$INTERNAL_NDAUAPI" \
       --name "$CONTAINER" \
       -e "HONEYCOMB_DATASET=$HONEYCOMB_DATASET" \
       -e "HONEYCOMB_KEY=$HONEYCOMB_KEY" \
       -e "NODE_ID=$CONTAINER" \
       -e "CHAOS_PEER=$CHAOS_PEER_ID@$PEER_IP:$CHAOS_PEER_P2P" \
       -e "NDAU_PEER=$NDAU_PEER_ID@$PEER_IP:$NDAU_PEER_P2P" \
       -e "SNAPSHOT_URL=$SNAPSHOT_BASE_URL/$SNAPSHOT.tgz" \
       --sysctl net.core.somaxconn=511 \
       ndauimage 

echo done
