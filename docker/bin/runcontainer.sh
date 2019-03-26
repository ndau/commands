#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

SNAPSHOT_BASE_URL="https://s3.amazonaws.com/ndau-snapshots"
INTERNAL_CHAOS_P2P=26660
INTERNAL_CHAOS_RPC=26670
INTERNAL_NDAU_P2P=26661
INTERNAL_NDAU_RPC=26671
INTERNAL_NDAUAPI=3030

if [ -z "$1" ] || \
   [ -z "$2" ] || \
   [ -z "$3" ] || \
   [ -z "$4" ] || \
   [ -z "$5" ] || \
   [ -z "$6" ] || \
   # $7 can be empty
   [ -z "$8" ]
   # $9 is optional
then
    echo "Usage:"
    echo "  ./runcontainer.sh" \
         "CONTAINER CHAOS_P2P CHAOS_RPC NDAU_P2P NDAU_RPC NDAUAPI PEERS SNAPSHOT IDENTITY"
    echo
    echo "Arguments:"
    echo "  CONTAINER  Name to give to the container to run"
    echo "  CHAOS_P2P  External port to map to the internal P2P port for the chaos chain"
    echo "  CHAOS_RPC  External port to map to the internal RPC port for the chaos chain"
    echo "  NDAU_P2P   External port to map to the internal P2P port for the ndau chain"
    echo "  NDAU_RPC   External port to map to the internal RPC port for the ndau chain"
    echo "  NDAUAPI    External port to map to the internal ndauapi port"
    echo "  PEERS      Comma-separated list of persistent peers on the network to join"
    echo "             Each peer should be of the form IP:CHAOS_P2P:CHAOS_RPC:NDAU_P2P:NDAU_RPC"
    echo "  SNAPSHOT   Name of the snapshot to use as a starting point for the node group"
    echo
    echo "Optionsl:"
    echo "  IDENTITY   node-identity.tgz file from a previous snaphot or initial container run"
    echo "             If present, the node will use it to configure itself when [re]starting"
    echo "             If missing, the node will generate one; keep it secret; keep it safe"
    exit 1
fi
CONTAINER="$1"
CHAOS_P2P="$2"
CHAOS_RPC="$3"
NDAU_P2P="$4"
NDAU_RPC="$5"
NDAUAPI="$6"
PEERS="$7"
SNAPSHOT="$8"
IDENTITY="$9"

if [[ "$CONTAINER" == *"/"* ]]; then
    # This is because we use a sed command inside the container and slashes confuse it.
    echo "Container name $CONTAINER cannot contain slashes"
    exit 1
fi

echo "Container: $CONTAINER"

if [ ! -z "$(docker container ls -a -q -f name=$CONTAINER)" ]; then
    echo "Container already exists: $CONTAINER"
    echo "Use restartcontainer.sh to restart it, or use removecontainer.sh to remove it first"
    exit 1
fi    

if [ ! -z "$IDENTITY" ] && [ ! -f "$IDENTITY" ] ; then
    echo "Cannot find node identity file: $IDENTITY"
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

CHAOS_PEERS=()
NDAU_PEERS=()

get_peer_id() {
    chain="$1"
    ip="$2"
    p2p="$3"
    rpc="$4"

    echo "Testing connection to $chain peer $ip:$p2p..."
    $(nc -G 5 -z "$ip" "$p2p")
    if [ "$?" != 0 ]; then
        echo "Could not reach $chain peer"
        exit 1
    fi

    echo "Getting $chain peer info for $ip:$rpc..."
    PEER_ID=$(curl -s --connect-timeout 5 "http://$ip:$rpc/status" | jq -r .result.node_info.id)
    if [ -z "$PEER_ID" ]; then
        echo "Could not get $chain peer id"
        exit 1
    fi
    echo "$chain peer id: $PEER_ID"
}

# Split the peers list by comma, then by colon.  Build up the "id@ip:port" persistent peer list.
peers=()
IFS=',' read -ra PEER <<< "$PEERS"
for i in "${PEER[@]}"; do
    peers+=("$i")
done
for peer in "${peers[@]}"; do
    peer_pieces=()

    IFS=':' read -ra pair <<< "$peer"
    for i in "${pair[@]}"; do
        peer_pieces+=("$i")
    done

    peer_ip=${peer_pieces[0]}

    peer_p2p=${peer_pieces[1]}
    peer_rpc=${peer_pieces[2]}
    PEER_ID=""
    get_peer_id chaos "$peer_ip" "$peer_p2p" "$peer_rpc"
    CHAOS_PEERS+=("tcp://$PEER_ID@$peer_ip:$peer_p2p")

    peer_p2p=${peer_pieces[3]}
    peer_rpc=${peer_pieces[4]}
    PEER_ID=""
    get_peer_id ndau "$peer_ip" "$peer_p2p" "$peer_rpc"
    NDAU_PEERS+=("tcp://$PEER_ID@$peer_ip:$peer_p2p")
done

# Join array elements together by a delimiter.  e.g. `join_by , (a b c)` returns "a,b,c".
join_by() { local IFS="$1"; shift; echo "$*"; }

CHAOS_PERSISTENT_PEERS=$(join_by , "${CHAOS_PEERS[@]}")
echo "chaos persistent peers: '$CHAOS_PERSISTENT_PEERS'"
NDAU_PERSISTENT_PEERS=$(join_by , "${NDAU_PEERS[@]}")
echo "ndau persistent peers: '$NDAU_PERSISTENT_PEERS'"

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
docker create \
       -p "$CHAOS_P2P":"$INTERNAL_CHAOS_P2P" \
       -p "$CHAOS_RPC":"$INTERNAL_CHAOS_RPC" \
       -p "$NDAU_P2P":"$INTERNAL_NDAU_P2P" \
       -p "$NDAU_RPC":"$INTERNAL_NDAU_RPC" \
       -p "$NDAUAPI":"$INTERNAL_NDAUAPI" \
       --name "$CONTAINER" \
       -e "HONEYCOMB_DATASET=$HONEYCOMB_DATASET" \
       -e "HONEYCOMB_KEY=$HONEYCOMB_KEY" \
       -e "NODE_ID=$CONTAINER" \
       -e "CHAOS_PERSISTENT_PEERS=$CHAOS_PERSISTENT_PEERS" \
       -e "NDAU_PERSISTENT_PEERS=$NDAU_PERSISTENT_PEERS" \
       -e "SNAPSHOT_URL=$SNAPSHOT_BASE_URL/$SNAPSHOT.tgz" \
       --sysctl net.core.somaxconn=511 \
       ndauimage 

if [ ! -z "$IDENTITY" ]; then
    echo "Copying node identity file to container..."
    docker cp "$IDENTITY" "$CONTAINER":/image/node-identity.tgz
fi

echo "Starting container..."
docker start "$CONTAINER"

echo done
