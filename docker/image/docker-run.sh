#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
# shellcheck source=docker-env.sh
source "$SCRIPT_DIR"/docker-env.sh

# Log startup notification locally.
notif_msg="$NODE_ID is starting up"
echo "$notif_msg"

# Place startup marker in honeycomb.
if [ -n "$HONEYCOMB_KEY" ] && [ -n "$HONEYCOMB_DATASET" ]; then
    notif_data='{"message":"'"$notif_msg"'","type":"deploy"}'
    curl -X POST -H "X-Honeycomb-Team: $HONEYCOMB_KEY" -d "$notif_data" \
         "https://api.honeycomb.io/1/markers/$HONEYCOMB_DATASET"
fi

# Send startup message to slack.
if [ -n "$SLACK_DEPLOYS_KEY" ]; then
    notif_data='{"text":"'"$notif_msg"'"}'
    curl -X POST -H "Content-type: application/json" -d "$notif_data" \
         "https://hooks.slack.com/services/$SLACK_DEPLOYS_KEY"
fi

# Remove this file while we're starting up.  Once it's written, it can be used as a flag
# to the outside world as to whether the container's processes are all fully running.
RUNNING_FILE="$SCRIPT_DIR/running"
rm -f "$RUNNING_FILE"

# This is needed because in the long term, noms eats more than 256 file descriptors.
ulimit -n "$ULIMIT_AMOUNT"

# If there's no data directory yet, it means we're starting from scratch.
if [ ! -d "$DATA_DIR" ]; then
    echo "Configuring node group..."
    "$SCRIPT_DIR"/docker-conf.sh
fi

# Every time the node group launches, replace persistent peer domain names with IP addresses.
# echo "Converting persistent peer domain names to IP addresses..."
# "$SCRIPT_DIR"/docker-dns.sh  <-- Don't do this, let AWS do smart DNS name resolution

# ensure the log directory exists
mkdir -p "$LOG_DIR"

# Start procmon, which will launch and manage all processes in the node group.
cd "$BIN_DIR" || exit 1
if [ -z "$HONEYCOMB_KEY" ] || [ -z "$HONEYCOMB_DATASET" ]; then
    # Honeycomb not configured, we'll dump everything locally from procmon itself.
    echo "logging to $LOG_DIR"
    ./procmon "$SCRIPT_DIR/docker-procmon.toml" >"$LOG_DIR/procmon.log" 2>&1 &
else
    # Honeycomb takes care of logging, we'll log nothing locally from procmon in this case.
    echo "logging to honeycomb: $HONEYCOMB_DATASET"
    ./procmon "$SCRIPT_DIR/docker-procmon.toml" &
fi
procmon_pid="$!"
echo "Started procmon as PID $procmon_pid"

# This will gracefully shut down all running processes through procmon when the container stops.
on_sigterm() {
    echo "Received SIGTERM; shutting down node group..."

    # Gracefully exit all processes in the node group through procmon.
    kill "$procmon_pid"
    wait "$procmon_pid"

    # For completeness, mark the container as not running.
    rm -f "$RUNNING_FILE"

    # SIGTERM = 128 + 15
    exit 143;
}

# Execute the specified handler when SIGTERM is received.
trap 'on_sigterm' SIGTERM

# Block until the entire node group is running.  Do this by checking the last task (ndauapi) port.
echo "Waiting for node group..."
until nc -z localhost "$NDAUAPI_PORT" 2>/dev/null
do
    # It usually takes a second or two to start up, so checking once per second doesn't cause too
    # much extra wait time and it also frees up CPU for the node to consume while starting up.
    sleep 1
done

# Block until we have a block height of at least 1.
# Useful for taking snapshots immediately (and safely) after genesis snapshot has been generated.
echo "Waiting for valid block height..."
while :
do
    response=$(curl -s "http://localhost:$NDAUAPI_PORT/block/height/1")
    if [ -n "$response" ] && [[ "$response" != *"could not get block"* ]]; then
        break
    fi
    sleep 1
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

echo "Node group $NODE_ID is now running"

# Keep the container alive for as long as procmon is alive.  We want the container to stop
# running if procmon dies for any reason.  One use case is for handling the schema change tx.
wait "$procmon_pid"
