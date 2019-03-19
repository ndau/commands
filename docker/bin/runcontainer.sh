#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

INTERNAL_CHAOS_P2P=26660
INTERNAL_CHAOS_RPC=26670
INTERNAL_NDAU_P2P=26661
INTERNAL_NDAU_RPC=26671

exit_with_usage() {
    echo "Usage:"
    echo "Run a container with a given name and public tendermint ports for chaos and ndau:"
    echo "  ./runcontainer.sh CONTAINER CHAOS_P2P CHAOS_RPC NDAU_P2P NDAU_RPC"
    echo "Restart an existing (stopped) container with a given name:"
    echo "  ./runcontainer.sh CONTAINER"
    exit 1
}

CONTAINER="$1"
if [ -z "$CONTAINER" ]; then
    CONTAINER=demonet-0
    echo "No container specified; using default: $CONTAINER"
fi

# Stop the container if it's running.  We can't run or restart it otherwise.
"$SCRIPT_DIR"/stopcontainer.sh "$CONTAINER"

if [ -z "$(docker container ls -a -q -f name=$CONTAINER)" ]; then
    if [ -z "$2" ] || [ -z "$3" ] || [ -z "$4" ] || [ -z "$5" ]; then
        exit_with_usage
    fi
    CHAOS_P2P="$2"
    CHAOS_RPC="$3"
    NDAU_P2P="$4"
    NDAU_RPC="$5"

    echo Silencing warning about Transparent Huge Pages when redis-server runs...
    docker run -it --privileged --pid=host debian nsenter -t 1 -m -u -n -i \
           sh -c 'echo never > /sys/kernel/mm/transparent_hugepage/enabled'
    docker run -it --privileged --pid=host debian nsenter -t 1 -m -u -n -i \
           sh -c 'echo never > /sys/kernel/mm/transparent_hugepage/defrag'

    echo "Running ndauimage as $CONTAINER..."
    # Some notes about the params to the run command:
    # - Using --sysctl silences a warning about TCP backlog when redis runs.
    # - Set your own HONEYCOMB_* env vars ahead of time, if desired.
    docker run -d \
           -p "$CHAOS_P2P":"$INTERNAL_CHAOS_P2P" \
           -p "$CHAOS_RPC":"$INTERNAL_CHAOS_RPC" \
           -p "$NDAU_P2P":"$INTERNAL_NDAU_P2P" \
           -p "$NDAU_RPC":"$INTERNAL_NDAU_RPC" \
           -e "HONEYCOMB_DATASET=$HONEYCOMB_DATASET" \
           -e "HONEYCOMB_KEY=$HONEYCOMB_KEY" \
           -e "NODE_ID=$CONTAINER" \
           --name="$CONTAINER" \
           --sysctl net.core.somaxconn=511 \
           ndauimage 
else
    echo "Restarting $CONTAINER..."
    docker restart "$CONTAINER"
fi
echo done
