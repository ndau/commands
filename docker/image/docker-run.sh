#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
source "$SCRIPT_DIR"/docker-env.sh

echo "Running $NODE_ID node group..."

# Remove this file while we're starting up.  Once it's written, it can be used as a flag
# to the outside world as to whether the container's processes are all fully running.
RUNNING_FILE="$SCRIPT_DIR/running"
rm -f "$RUNNING_FILE"

# If there's no data directory yet, it means we're starting from scratch.
if [ ! -d "$DATA_DIR" ]; then
    echo "Configuring node group..."
    "$SCRIPT_DIR"/docker-conf.sh
fi

# This is needed because in the long term, noms eats more than 256 file descriptors
ulimit -n 1024

# All commands are run out of the bin directory.
cd "$BIN_DIR" || exit 1

./procmon --configfile "$SCRIPT_DIR/docker-ndau.toml" >"$LOG_DIR/procmon.log" 2>&1 &
echo "Started procmon as PID $!"

# Block until the entire node group is running.  Do this by checking the last task: ndauapi.
echo "Waiting for node group..."
until nc -z localhost "$NDAUAPI_PORT" 2>/dev/null
do
    :
done

# Generate the node-identity file if one wasn't passed in.
IDENTITY_FILE="$SCRIPT_DIR"/node-identity.tgz
if [ ! -f "$IDENTITY_FILE" ]; then
    echo "Generating identity file..."

    cd "$DATA_DIR" || exit 1
    tar -czf "$IDENTITY_FILE" \
        tendermint/config/node_key.json \
        tendermint/config/priv_validator_key.json
fi

# Everything's up and running.  The outside world can poll for this file to know this.
touch "$RUNNING_FILE"

echo "Node group $NODE_ID is now running"

# Wait forever to keep the container alive.
while true; do sleep 86400; done
