#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

IMAGE_BASE_URL="https://s3.amazonaws.com/ndau-images"
SERVICES_URL="https://s3.us-east-2.amazonaws.com/ndau-json/services.json"
INTERNAL_P2P_PORT=26660
INTERNAL_RPC_PORT=26670
INTERNAL_API_PORT=3030
GENERATED_GENESIS_SNAPSHOT="*"

# Leave this blank/unset to disable periodic snapshot creation.
# Set to "4h", for example, to generate a snapshot every 4 hours.
# Only the latest snapshot will exist in the container at a time, and the AWS_* env vars
# must be set in order for each snapshot to be uploaded to the ndau-snapshots S3 bucket.
#SNAPSHOT_INTERVAL="4h"
#AWS_ACCESS_KEY_ID=""
#AWS_SECRET_ACCESS_KEY=""

if [ -z "$1" ] || \
   [ -z "$2" ] || \
   [ -z "$3" ] || \
   [ -z "$4" ] || \
   [ -z "$5" ]
   # $6 through $10 are optional and "" can be used for any of them.
then
    echo "Usage:"
    echo "  ./runcontainer.sh" \
         "NETWORK CONTAINER P2P_PORT RPC_PORT API_PORT" \
         "[IDENTITY] [SNAPSHOT] [PEERS_P2P] [PEERS_RPC]"
    echo
    echo "Arguments:"
    echo "  NETWORK    Which network to join: localnet, devnet, testnet, mainnet"
    echo "  CONTAINER  Name to give to the container to run"
    echo "  P2P_PORT   External port to map to the internal P2P port for the blockchain"
    echo "  RPC_PORT   External port to map to the internal RPC port for the blockchain"
    echo "  API_PORT   External port to map to the internal ndau API port"
    echo
    echo "Optional:"
    echo "  IDENTITY   node-identity.tgz file from a previous snaphot or initial container run"
    echo "             If present, the node will use it to configure itself when [re]starting"
    echo "             If missing, the node will generate a new identity for itself"
    echo "  SNAPSHOT   Name of the snapshot to use as a starting point for the node group"
    echo "               If omitted, the latest $NETWORK snapshot will be used"
    echo "               If it's a file, it will be used instead of pulling one from S3"
    echo "               If it's $GENERATED_GENESIS_SNAPSHOT, genesis data is generated"
    echo "  PEERS_P2P  Comma-separated list of persistent peers on the network to join"
    echo "               Each peer should be of the form IP_OR_DOMAIN_NAME:PORT"
    echo "               If omitted, peers will be gotten from $NETWORK for non-localnet"
    echo "  PEERS_RPC  Comma-separated list of the same peers for RPC connections"
    echo "               Each peer should be of the form PROTOCOL://IP_OR_DOMAIN_NAME:PORT"
    echo "               If omitted, peers will be gotten from $NETWORK for non-localnet"
    echo
    echo "Required if the node is to become a validator node at some point:"
    echo "  WEBHOOK_URL The URL to be called by ndaunode when the node is nominated for"
    echo "                a node reward. This usually points to a port listening on"
    echo "                locahost if the claimer process is running, but may point to"
    echo "                any external URL"
    echo
    echo "Required if the node reward claimer process is to be run locally by the node:"
    echo "  CLAIMER_PORT The localhost port number used by the claimer process, if running. This value"
    echo "                 should match the port number specified in the claimer_config.toml file"
    echo "                 defined by the BASE64_CLAIMER_CONFIG file (see below)."
    echo
    echo "Environment variables:"
    echo "  BASE64_NODE_IDENTITY"
    echo "             Set to override the IDENTITY parameter"
    echo "             The contents of the variable are a base64 encoded tarball containing:"
    echo "               - tendermint/config/priv_validator_key.json"
    echo "               - tendermint/config/node_id.json"
    echo
    echo "  BASE64_CLAIMER_CONFIG"
    echo "             Provides configuration information for automated claiming of"
    echo "             node rewards; there are no default values. If this variable is not set,"
    echo "             no node reward claimer process will be run locally. The contents of the"
    echo "             variable are a base64 encoded tarball containing"
    echo "               - claimer_config.toml"
    echo
    exit 1
fi

NETWORK="$1"
CONTAINER="$2"
P2P_PORT="$3"
RPC_PORT="$4"
API_PORT="$5"
IDENTITY="$6"
SNAPSHOT="$7"
PEERS_P2P="$8"
PEERS_RPC="$9"
WEBHOOK_URL="$10"
CLAIMER_PORT="$11"

if [ "$NETWORK" != "localnet" ] && \
   [ "$NETWORK" != "devnet" ] && \
   [ "$NETWORK" != "testnet" ] && \
   [ "$NETWORK" != "mainnet" ]; then
    echo "Unsupported network: $NETWORK"
    echo "Supported networks: localnet, devnet, testnet, mainnet"
    exit 1
fi

echo "Network: $NETWORK"

# Validate container name (can't have slashes).
if [[ "$CONTAINER" == *"/"* ]]; then
    # This is because we use a sed command inside the container and slashes confuse it.
    echo "Container name $CONTAINER cannot contain slashes"
    exit 1
fi

echo "Container: $CONTAINER"

if [ -n "$(docker container ls -a -q -f name="$CONTAINER")" ]; then
    echo "Container already exists: $CONTAINER"
    echo "Use restartcontainer.sh to restart it, or use removecontainer.sh to remove it first"
    exit 1
fi

# If we're not overriding the identity parameter,
# and an identity file was specified,
# but the file doesn't exist...
if [ -z "$BASE64_NODE_IDENTITY" ] && [ -n "$IDENTITY" ] && [ ! -f "$IDENTITY" ]; then
    echo "Cannot find node identity file: $IDENTITY"
    exit 1
fi

echo "P2P port: $P2P_PORT"
echo "RPC port: $RPC_PORT"
echo "API port: $API_PORT"

if [ -z "$SNAPSHOT" ]; then
    echo "Snapshot: (latest)"
elif [ "$SNAPSHOT" = "$GENERATED_GENESIS_SNAPSHOT" ]; then
    echo "Snapshot: (generated)"
else
    echo "Snapshot: $SNAPSHOT"
fi

# The timeout flag on linux differs from mac.
if [[ "$OSTYPE" == *"darwin"* ]]; then
    # Use -G on macOS; there is no -G option on linux.
    NC_TIMEOUT_FLAG="-G"
else
    # Use -w on linux; the -w option does not work on macOS.
    NC_TIMEOUT_FLAG="-w"
fi

test_local_port() {
    port="$1"

    if nc "$NC_TIMEOUT_FLAG" 5 -z localhost "$port" 2>/dev/null; then
        echo "Port $port is already in use"
        exit 1
    fi
}

test_local_port "$P2P_PORT"
test_local_port "$RPC_PORT"
test_local_port "$API_PORT"

test_peer() {
    ip="$1"
    port="$2"

    if [ -z "$ip" ] || [ -z "$port" ]; then
        echo "Missing p2p ip or port: ip=($ip) port=($port)"
        exit 1
    fi

    echo "Testing connection to peer $ip:$port..."
    if ! nc "$NC_TIMEOUT_FLAG" 5 -z "$ip" "$port"; then
        echo "Could not reach peer"
        exit 1
    fi
}

get_peer_id() {
    protocol="$1"
    ip="$2"
    port="$3"

    if [ -z "$protocol" ] || [ -z "$ip" ] || [ -z "$port" ]; then
        echo "Missing rpc protocol, ip or port: protocol=($protocol) ip=($ip) port=($port)"
        exit 1
    fi

    url="$protocol://$ip:$port"
    echo "Getting peer info for $url..."
    PEER_ID=$(curl -s --connect-timeout 5 "$url/status" | jq -r .result.node_info.id)
    if [ -z "$PEER_ID" ]; then
        echo "Could not get peer id"
        exit 1
    fi
    echo "Peer id: $PEER_ID"
}

# Join array elements together by a delimiter.  e.g. `join_by , (a b c)` returns "a,b,c".
join_by() { local IFS="$1"; shift; echo "$*"; }

# If no peers were given, we can get them automatically for non-localnet networks.
# When running a localnet, the first peer can start w/o knowing any other peers.
if [ -z "$PEERS_P2P" ] && [ -z "$PEERS_RPC" ] && [ "$NETWORK" != "localnet" ]; then
    echo "Fetching $SERVICES_URL..."
    services_json=$(curl -s "$SERVICES_URL")
    # shellcheck disable=SC2207
    # it works well enough for now
    p2ps=($(echo "$services_json" | jq -r ".networks.$NETWORK.nodes[].p2p"))
    # shellcheck disable=SC2207
    # it works well enough for now
    rpcs=($(echo "$services_json" | jq -r ".networks.$NETWORK.nodes[].rpc"))

    len="${#rpcs[@]}"
    if [ "$len" = 0 ]; then
        echo "No nodes published for network: $NETWORK"
        exit 1
    fi

    # The RPC connections must be made through https.
    for node in $(seq 0 $((len - 1))); do
        rpcs[$node]="https://${rpcs[$node]}"
    done

    PEERS_P2P=$(join_by , "${p2ps[@]}")
    PEERS_RPC=$(join_by , "${rpcs[@]}")
fi

# Split the peers list by comma, then by colon.  Build up the "id@ip:port" persistent peer list.
persistent_peers=()
IFS=',' read -ra peers_p2p <<< "$PEERS_P2P"
IFS=',' read -ra peers_rpc <<< "$PEERS_RPC"
len="${#peers_p2p[@]}"
if [ "$len" != "${#peers_rpc[@]}" ]; then
    echo "The length of P2P and RPC peers must match"
    exit 1
fi
if [ "$len" -gt 0 ]; then
    for peer in $(seq 0 $((len - 1))); do
        IFS=':' read -ra pieces <<< "${peers_p2p[$peer]}"
        p2p_ip="${pieces[0]}"
        p2p_port="${pieces[1]}"

        test_peer "$p2p_ip" "$p2p_port"

        IFS=':' read -ra pieces <<< "${peers_rpc[$peer]}"
        rpc_protocol="${pieces[0]}"
        rpc_ip="${pieces[1]}"
        rpc_port="${pieces[2]}"

        # Since we split on colon, the double-slash is stuck to the ip.  Remove it.
        rpc_ip="${rpc_ip:2}"

        PEER_ID=""
        get_peer_id "$rpc_protocol" "$rpc_ip" "$rpc_port"
        persistent_peers+=("$PEER_ID@$p2p_ip:$p2p_port")
    done
fi

PERSISTENT_PEERS=$(join_by , "${persistent_peers[@]}")
echo "Persistent peers: '$PERSISTENT_PEERS'"

# Stop the container if it's running.  We can't run or restart it otherwise.
"$SCRIPT_DIR"/stopcontainer.sh "$CONTAINER"

# If the image isn't present, fetch the "current" image from S3 for the given network.
if [ "$NETWORK" = "localnet" ] || [ "$USE_LOCAL_IMAGE" = 1 ]; then
    NDAU_IMAGE_NAME="ndauimage:latest"
else
    NDAU_IMAGES_SUBDIR="ndau-images"
    NDAU_IMAGES_DIR="$SCRIPT_DIR/../$NDAU_IMAGES_SUBDIR"
    mkdir -p "$NDAU_IMAGES_DIR"

    CURRENT_FILE="current-$NETWORK.txt"
    CURRENT_PATH="$NDAU_IMAGES_DIR/$CURRENT_FILE"
    echo "Fetching $CURRENT_FILE..."
    curl -o "$CURRENT_PATH" "$IMAGE_BASE_URL/$CURRENT_FILE"
    if [ ! -f "$CURRENT_PATH" ]; then
        echo "Unable to fetch $IMAGE_BASE_URL/$CURRENT_FILE"
        exit 1
    fi
    # Use the specified image SHA if one is given
    if [ -z $SHA ]; then
        SHA=$(cat "$CURRENT_PATH")
    fi
    NDAU_IMAGE_NAME="ndauimage:$SHA"

    if [ -z "$(docker image ls -q "$NDAU_IMAGE_NAME")" ]; then
        echo "Unable to find $NDAU_IMAGE_NAME locally; fetching..."

        IMAGE_NAME="ndauimage-$SHA"
        IMAGE_ZIP="$IMAGE_NAME.docker.gz"
        IMAGE_PATH="$NDAU_IMAGES_DIR/$IMAGE_NAME.docker"
        echo "Fetching $IMAGE_ZIP..."
        curl -o "$IMAGE_PATH.gz" "$IMAGE_BASE_URL/$IMAGE_ZIP"
        if [ ! -f "$IMAGE_PATH.gz" ]; then
            echo "Unable to fetch $IMAGE_BASE_URL/$IMAGE_ZIP"
            exit 1
        fi

        echo "Loading $NDAU_IMAGE_NAME..."
        gunzip -f "$IMAGE_PATH.gz"
        docker load -i "$IMAGE_PATH"
        if [ -z "$(docker image ls -q "$NDAU_IMAGE_NAME")" ]; then
            echo "Unable to load $NDAU_IMAGE_NAME"
            exit 1
        fi
    fi
fi

# Supply Tendermint configuration defaults if not provided. No default is needed
# for SEEDS because "seeds = " is acceptable syntax in config.toml

if [ -z $TM_LOG_LEVEL ]; then
    TM_LOG_LEVEL="p2p:none,rpc-server:none,*:info"
    echo "Setting TM_LOG_LEVEL default to $TM_LOG_LEVEL"
fi
if [ -z $PEX ]; then
    PEX=true          # Turn on PEX peer reactor by default
    echo "Setting PEX default to $PEX"
fi
if [ -z $SEED_MODE ]; then
    SEED_MODE=false   # Don't run in seed mode
    echo "Setting SEED_MODE default to $SEED_MODE"
fi

echo "Creating container..."
# Some notes about the params to the run command:
# - Using --sysctl silences a warning about TCP backlog when redis runs.
# - Set your own HONEYCOMB_* and SLACK_* env vars ahead of time to enable honeycomb logging.
docker create \
       -p "$P2P_PORT":"$INTERNAL_P2P_PORT" \
       -p "$RPC_PORT":"$INTERNAL_RPC_PORT" \
       -p "$API_PORT":"$INTERNAL_API_PORT" \
       --name "$CONTAINER" \
       -e "NETWORK=$NETWORK" \
       -e "HONEYCOMB_DATASET=$HONEYCOMB_DATASET" \
       -e "HONEYCOMB_KEY=$HONEYCOMB_KEY" \
       -e "SLACK_DEPLOYS_KEY=$SLACK_DEPLOYS_KEY" \
       -e "AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID" \
       -e "AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY" \
       -e "SNAPSHOT_INTERVAL=$SNAPSHOT_INTERVAL" \
       -e "NODE_ID=$CONTAINER" \
       -e "PERSISTENT_PEERS=$PERSISTENT_PEERS" \
       -e "BASE64_NODE_IDENTITY=$BASE64_NODE_IDENTITY" \
       -e "BASE64_CLAIMER_CONFIG=$BASE64_CLAIMER_CONFIG" \
       -e "CLAIMER_PORT=$CLAIMER_PORT" \
       -e "WEBHOOK_URL=$WEBHOOK_URL" \
       -e "SNAPSHOT_NAME=$SNAPSHOT" \
       -e "TM_LOG_LEVEL=$TM_LOG_LEVEL" \
       -e "PEX=$PEX" \
       -e "SEEDS=$SEEDS" \
       -e "SEED_MODE=$SEED_MODE" \
       --sysctl net.core.somaxconn=511 \
       "$NDAU_IMAGE_NAME"

IDENTITY_FILE="node-identity.tgz"
# Copy the identity file into the container if one was specified,
# but not if the base64 environment variable is being used to effectively override the file.
if [ -n "$IDENTITY" ] && [ -z "$BASE64_NODE_IDENTITY" ]; then
    echo "Copying node identity file to container..."
    docker cp "$IDENTITY" "$CONTAINER:/image/$IDENTITY_FILE"
fi

# Copy the snapshot into the container if it exists as a local file.
if [ -f "$SNAPSHOT" ]; then
    echo "Copying local snapshot file to container..."
    docker cp "$SNAPSHOT" "$CONTAINER:/image/snapshot-$NETWORK-0.tgz"
fi

echo "Starting container..."
docker start "$CONTAINER"

# Run the hang monitor while we wait for the node to spin up.
"$SCRIPT_DIR"/watchcontainer.sh "$CONTAINER" &
watcher_pid="$!"

echo "Waiting for the node to fully spin up..."
until docker exec "$CONTAINER" test -f /image/running 2>/dev/null
do
    # It usually takes a second or two to start up, so checking once per second doesn't cause too
    # much extra wait time and it also frees up CPU for the node to consume while starting up.
    sleep 1
done

# Done waiting; kill the watcher.
kill "$watcher_pid" && wait "$watcher_pid" 2>/dev/null

echo "Node is ready; dumping container logs..."
docker container logs "$CONTAINER" 2>/dev/null | sed -e 's/^/> /'

# In the case no node identity was passed in, wait for it to generate one then copy it out.
# It's important that node operators keep the node-identity.tgz file secure.
if [ -z "$IDENTITY" ] && [ -z "$BASE64_NODE_IDENTITY" ]; then
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

echo "done"
