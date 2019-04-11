#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

SNAPSHOT_BASE_URL="https://s3.amazonaws.com/ndau-snapshots"
INTERNAL_P2P_PORT=26660
INTERNAL_RPC_PORT=26670
INTERNAL_NDAUAPI_PORT=3030

if [ -z "$1" ] || \
   [ -z "$2" ] || \
   [ -z "$3" ] || \
   [ -z "$4" ] || \
   # $5 can be empty
   [ -z "$6" ]
   # $7 is optional
then
    echo "Usage:"
    echo "  ./runcontainer.sh" \
         "CONTAINER P2P_PORT RPC_PORT NDAUAPI_PORT PEERS SNAPSHOT IDENTITY"
    echo
    echo "Arguments:"
    echo "  CONTAINER     Name to give to the container to run"
    echo "  P2P_PORT      External port to map to the internal P2P port for the blockchain"
    echo "  RPC_PORT      External port to map to the internal RPC port for the blockchain"
    echo "  NDAUAPI_PORT  External port to map to the internal ndauapi port"
    echo "  PEERS         Comma-separated list of persistent peers on the network to join"
    echo "                Each peer should be of the form IP:P2P:RPC"
    echo "  SNAPSHOT      Name of the snapshot to use as a starting point for the node group"
    echo
    echo "Optional:"
    echo "  IDENTITY      node-identity.tgz file from a previous snaphot or initial container run"
    echo "                If present, the node will use it to configure itself when [re]starting"
    echo "                If missing, the node will generate a new identity for itself"
    echo
    echo "  BASE64_NODE_IDENTITY (environment variable)"
    echo "                This environment variable can be set to provide an identity. If this variable"
    echo "                is supplied, the IDENTITY file above will not be used. The contents of the"
    echo "                variable are a base64 encoded tarball containing the files: "
    echo "                  - tendermint/config/priv_validator_key.json"
    echo "                  - tendermint/config/node_id.json"
    exit 1
fi
CONTAINER="$1"
P2P_PORT="$2"
RPC_PORT="$3"
NDAUAPI_PORT="$4"
PEERS="$5"
SNAPSHOT="$6"
IDENTITY="$7"

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

echo "P2P port: $P2P_PORT"
echo "RPC port: $RPC_PORT"
echo "ndauapi port: $NDAUAPI_PORT"

test_port() {
    ip="$1"
    port="$2"

    $(nc -G 1 -z "$ip" "$port" 2>/dev/null)
    if [ "$?" = 0 ]; then
        echo "Port at $ip:$port is already in use"
        exit 1
    fi
}

test_port localhost "$P2P_PORT"
test_port localhost "$RPC_PORT"
test_port localhost "$NDAUAPI_PORT"

get_peer_id() {
    ip="$1"
    p2p="$2"
    rpc="$3"

    if [ -z "$ip" ] || [ -z "$p2p" ] || [ -z "$rpc" ]; then
        echo "Missing ip or p2p or rpc: ip=($ip) p2p=($p2p) rpc=($rpc)"
        exit 1
    fi

    echo "Testing connection to peer $ip:$p2p..."
    $(nc -G 5 -z "$ip" "$p2p")
    if [ "$?" != 0 ]; then
        echo "Could not reach peer"
        exit 1
    fi

    echo "Getting peer info for $ip:$rpc..."
    PEER_ID=$(curl -s --connect-timeout 5 "http://$ip:$rpc/status" | jq -r .result.node_info.id)
    if [ -z "$PEER_ID" ]; then
        echo "Could not get peer id"
        exit 1
    fi
    echo "Peer id: $PEER_ID"
}

# Split the peers list by comma, then by colon.  Build up the "id@ip:port" persistent peer list.
persistent_peers=()
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
    get_peer_id "$peer_ip" "$peer_p2p" "$peer_rpc"
    persistent_peers+=("tcp://$PEER_ID@$peer_ip:$peer_p2p")
done

# Join array elements together by a delimiter.  e.g. `join_by , (a b c)` returns "a,b,c".
join_by() { local IFS="$1"; shift; echo "$*"; }

PERSISTENT_PEERS=$(join_by , "${persistent_peers[@]}")
echo "Persistent peers: '$PERSISTENT_PEERS'"

echo "Snapshot: $SNAPSHOT"

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
       -p "$P2P_PORT":"$INTERNAL_P2P_PORT" \
       -p "$RPC_PORT":"$INTERNAL_RPC_PORT" \
       -p "$NDAUAPI_PORT":"$INTERNAL_NDAUAPI_PORT" \
       --name "$CONTAINER" \
       -e "HONEYCOMB_DATASET=$HONEYCOMB_DATASET" \
       -e "HONEYCOMB_KEY=$HONEYCOMB_KEY" \
       -e "NODE_ID=$CONTAINER" \
       -e "PERSISTENT_PEERS=$PERSISTENT_PEERS" \
       -e "BASE64_NODE_IDENTITY=$BASE64_NODE_IDENTITY" \
       -e "SNAPSHOT_URL=$SNAPSHOT_BASE_URL/$SNAPSHOT.tgz" \
       --sysctl net.core.somaxconn=511 \
       ndauimage 

IDENTITY_FILE=node-identity.tgz
if [ ! -z "$IDENTITY" ]; then
    echo "Copying node identity file to container..."
    docker cp "$IDENTITY" "$CONTAINER:/image/$IDENTITY_FILE"
fi

echo "Starting container..."
docker start "$CONTAINER"

echo "Waiting for the node to fully spin up..."
until docker exec "$CONTAINER" test -f /image/running 2>/dev/null
do
    :
done

# In the case no node identity was passed in, wait for it to generate one then copy it out.
# It's important that node operators keep the node-identity.tgz file secure.
if [ -z "$IDENTITY" ]; then
    # We can copy the file out now since we waited for the node to full spin up above.
    OUT_FILE="$SCRIPT_DIR/node-identity-$CONTAINER.tgz"
    docker cp "$CONTAINER:/image/$IDENTITY_FILE" "$OUT_FILE"

    echo
    echo "The node identity has been generated and copied out of the container here:"
    echo "  $OUT_FILE"
    echo
    echo "You can always get it at a later time by running the following:"
    echo "  docker cp $CONTAINER:/image/$IDENTITY_FILE $IDENTITY_FILE"
    echo "It can be used to restart this container with the same identity it has now"
    echo "Keep it secret; keep it safe"
    echo
fi

echo done
