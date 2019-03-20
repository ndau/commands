#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

INTERNAL_CHAOS_P2P=26660
INTERNAL_CHAOS_RPC=26670
INTERNAL_NDAU_P2P=26661
INTERNAL_NDAU_RPC=26671
INTERNAL_NDAUAPI=3030

if [ -z "$1" ]||[ -z "$2" ]||[ -z "$3" ]||[ -z "$4" ]||[ -z "$5" ]||[ -z "$6" ]||[ -z "$7" ]; then
    echo "Usage:"
    echo "  ./runcontainer.sh CONTAINER SNAPSHOT CHAOS_P2P CHAOS_RPC NDAU_P2P NDAU_RPC NDAUAPI"
    echo
    echo "Arguments:"
    echo "  CONTAINER   Name to give to the container to run"
    echo "  SNAPSHOT    Path to snapshot data with which to start the node group"
    echo "  CHAOS_P2P   External port to map to the internal P2P port for the chaos chain"
    echo "  CHAOS_RPC   External port to map to the internal RPC port for the chaos chain"
    echo "  NDAU_P2P    External port to map to the internal P2P port for the ndau chain"
    echo "  NDAU_RPC    External port to map to the internal RPC port for the ndau chain"
    echo "  NDAUAPI     External port to map to the internal ndauapi port"
    exit 1
fi
CONTAINER="$1"
SNAPSHOT="$2"
CHAOS_P2P="$3"
CHAOS_RPC="$4"
NDAU_P2P="$5"
NDAU_RPC="$6"
NDAUAPI="$7"

if [ ! -z "$(docker container ls -a -q -f name=$CONTAINER)" ]; then
    echo "Container already exists: $CONTAINER"
    echo "Use restartcontainer.sh to restart it, or use removecontainer.sh to remove it first"
    exit 1
fi    
echo "Container: $CONTAINER"

if [ ! -d "$SNAPSHOT" ]; then
    echo "Could not find snapshot directory: $SNAPSHOT"
    exit 1
fi

SVI_NAMESPACE_FILE="$SNAPSHOT/svi-namespace"
if [ ! -f "$SVI_NAMESPACE_FILE" ]; then
    echo "Could not find svi namespace file: $SVI_NAMESPACE_FILE"
    exit 1
fi
SVI_NAMESPACE=$(cat "$SVI_NAMESPACE_FILE")
echo "SVI Namespace: $SVI_NAMESPACE"

echo "chaos P2P port: $CHAOS_P2P"
echo "chaos RPC port: $CHAOS_RPC"
echo "ndau P2P port: $NDAU_P2P"
echo "ndau RPC port: $NDAU_RPC"
echo "ndauapi port: $NDAUAPI"

DATA_DIR="$SNAPSHOT/data"
if [ ! -d "$DATA_DIR" ]; then
    echo "Could not find data directory: $DATA_DIR"
    exit 1
fi

# Check for the existence of required data files in the snapshot.
# TODO: Add support for starting fresh with a genesis snapshot, which doesn't include everything.
#       We will have to set genesis.json app_hash inside docker-run.sh when to support this.
#       Currently we rely on it already being there, and can only connect to existing networks.
CHAOS_DATA_DIR="$DATA_DIR/chaos"
if [ ! -d "$CHAOS_DATA_DIR" ]; then
    echo "Could not find chaos data directory: $CHAOS_DATA_DIR"
    exit 1
fi
CHAOS_NOMS_DATA_DIR="$CHAOS_DATA_DIR/noms"
if [ ! -d "$CHAOS_NOMS_DATA_DIR" ]; then
    echo "Could not find chaos noms data directory: $CHAOS_NOMS_DATA_DIR"
    exit 1
fi
CHAOS_REDIS_DATA_DIR="$CHAOS_DATA_DIR/redis"
if [ ! -d "$CHAOS_REDIS_DATA_DIR" ]; then
    echo "Could not find chaos redis data directory: $CHAOS_REDIS_DATA_DIR"
    exit 1
fi
CHAOS_TENDERMINT_DATA_DIR="$CHAOS_DATA_DIR/tendermint"
if [ ! -d "$CHAOS_TENDERMINT_DATA_DIR" ]; then
    echo "Could not find chaos tendermint data directory: $CHAOS_TENDERMINT_DATA_DIR"
    exit 1
fi
CHAOS_TENDERMINT_GENESIS_FILE="$CHAOS_TENDERMINT_DATA_DIR/config/genesis.json"
if [ ! -f "$CHAOS_TENDERMINT_GENESIS_FILE" ]; then
    echo "Could not find chaos tendermint genesis file: $CHAOS_TENDERMINT_GENESIS_FILE"
    exit 1
fi
NDAU_DATA_DIR="$DATA_DIR/ndau"
if [ ! -d "$NDAU_DATA_DIR" ]; then
    echo "Could not find ndau data directory: $NDAU_DATA_DIR"
    exit 1
fi
NDAU_NOMS_DATA_DIR="$NDAU_DATA_DIR/noms"
if [ ! -d "$NDAU_NOMS_DATA_DIR" ]; then
    echo "Could not find ndau noms data directory: $NDAU_NOMS_DATA_DIR"
    exit 1
fi
NDAU_REDIS_DATA_DIR="$NDAU_DATA_DIR/redis"
if [ ! -d "$NDAU_REDIS_DATA_DIR" ]; then
    echo "Could not find ndau redis data directory: $NDAU_REDIS_DATA_DIR"
    exit 1
fi
NDAU_TENDERMINT_DATA_DIR="$NDAU_DATA_DIR/tendermint"
if [ ! -d "$NDAU_TENDERMINT_DATA_DIR" ]; then
    echo "Could not find ndau tendermint data directory: $NDAU_TENDERMINT_DATA_DIR"
    exit 1
fi
NDAU_TENDERMINT_GENESIS_FILE="$NDAU_TENDERMINT_DATA_DIR/config/genesis.json"
if [ ! -f "$NDAU_TENDERMINT_GENESIS_FILE" ]; then
    echo "Could not find ndau tendermint genesis file: $NDAU_TENDERMINT_GENESIS_FILE"
    exit 1
fi

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
       --sysctl net.core.somaxconn=511 \
       ndauimage 

echo "Copying snapshot to container..."
docker cp "$SNAPSHOT/." "$CONTAINER":/image/

echo "Starting container..."
docker start "$CONTAINER"

echo done
