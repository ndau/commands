#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
source "$SCRIPT_DIR"/docker-env.sh

echo "Running $NODE_ID node group..."

# Remove this file while we're starting up.  Once it's written, it can be used as a flag
# to the outside world as to whether the container's processes are all fully running.
RUNNING_FILE="$SCRIPT_DIR/running"
rm -f "$RUNNING_FILE"

# This is needed because in the long term, noms eats more than 256 file descriptors.
ulimit -n 1024

# If there's no data directory yet, it means we're starting from scratch.
if [ ! -d "$DATA_DIR" ]; then
    echo "Configuring node group..."
    "$SCRIPT_DIR"/docker-conf.sh
fi

# Start procmon, which will launch and manage all processes in the node group.
cd "$BIN_DIR" || exit 1
./procmon "$SCRIPT_DIR/docker-procmon.toml" >"$LOG_DIR/procmon.log" 2>&1 &
procmon_pid="$!"
echo "Started procmon as PID $procmon_pid"

# This will gracefully shut down all running processes through procmon when the container stops.
on_sigterm() {
    echo "Received SIGTERM; shutting down node group..."

    # Gracefully exit all processes in the node group through procmon.
    kill "$procmon_pid"
    wait "$procmon_pid"

    # Logs start over next time.  Save a copy of them all.  Having the "last run" might be useful.
    lastrun_dir="$LOG_DIR/lastrun"
    rm -rf "$lastrun_dir"
    mkdir -p "$lastrun_dir"
    mv "$LOG_DIR"/*.log "$lastrun_dir"

    # For completeness, mark the container as not running.
    rm -f "$RUNNING_FILE"

    # SIGTERM = 128 + 15
    exit 143;
}

# Kill the last background process (`tail -f /dev/null`) then execute the specified handler.
trap 'kill ${!}; on_sigterm' SIGTERM

# Block until the entire node group is running.  Do this by checking the last task (ndauapi) port.
echo "Waiting for node group..."
until nc -z localhost "$NDAUAPI_PORT" 2>/dev/null
do
    :
done

# Now that we know all data files are in place and the node group is running,
# we can generate the node-identity file if one wasn't passed in.
IDENTITY_FILE="$SCRIPT_DIR"/node-identity.tgz
if [ ! -f "$IDENTITY_FILE" ] && [ -z "$BASE64_NODE_IDENTITY" ]; then
    echo "Generating identity file..."

    cd "$DATA_DIR" || exit 1
    tar -czf "$IDENTITY_FILE" \
        tendermint/config/node_key.json \
        tendermint/config/priv_validator_key.json
fi

# Everything's up and running.  The outside world can poll for this file to know this.
touch "$RUNNING_FILE"

# If the INSANE_LOGGING environment varibale is set, tail will dump all output to stdout.
if [ !-z "$INSANE_LOGGING" ]; then
  tail "$LOG_DIR/*.log"-f &
fi

echo "Node group $NODE_ID is now running"

# Wait forever to keep the container alive.
while true
do
    tail -f /dev/null & wait ${!}
done
