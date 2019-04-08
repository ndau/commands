#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
source "$SCRIPT_DIR"/docker-env.sh

# The outside world can look for a SNAPSHOT_RESULT starting with this to handle errors.
ERROR_PREFIX="ERROR:"

# Get the network name from tendermint's chain_id.
GENESIS_JSON="$TM_DATA_DIR/config/genesis.json"
NETWORK=$(sed -n -e 's/^  "chain_id": "\(.*\)",$/\1/p' "$GENESIS_JSON")
if [ "$NETWORK" = "" ]; then
    ERROR_MSG="$ERROR_PREFIX Unable to deduce network name; cannot generate snapshot"
    echo "$ERROR_MSG"
    export SNAPSHOT_RESULT="$ERROR_MSG"
    exit 1
fi
echo "Generating $NETWORK snapshot..."

# Prepare the output directory for the snapshot.
NDAU_SNAPSHOTS_SUBDIR=ndau-snapshots
NDAU_SNAPSHOTS_DIR="$SCRIPT_DIR/$NDAU_SNAPSHOTS_SUBDIR"
rm -rf "$NDAU_SNAPSHOTS_DIR"
mkdir -p "$NDAU_SNAPSHOTS_DIR"

# Make a temp dir for copying data files into for tar'ing up later in this script.
SNAPSHOT_TEMP_DIR="$SCRIPT_DIR"/snapshot-temp
rm -rf "$SNAPSHOT_TEMP_DIR"
mkdir -p "$SNAPSHOT_TEMP_DIR"
SNAPSHOT_DATA_DIR="$SNAPSHOT_TEMP_DIR/data"

# Use the deep tendermint data directories to create all the parent subdirectories we need.
TM_TEMP="$SNAPSHOT_DATA_DIR/tendermint"
mkdir -p "$TM_TEMP/config"
mkdir -p "$TM_TEMP/data"

# Copy all the data files we want into the temp dir.
cp -r "$NOMS_DATA_DIR" "$SNAPSHOT_DATA_DIR/noms"
cp -r "$REDIS_DATA_DIR" "$SNAPSHOT_DATA_DIR/redis"
cp "$GENESIS_JSON" "$TM_TEMP/config"
cp -r "$TM_DATA_DIR/data/blockstore.db" "$TM_TEMP/data"
cp -r "$TM_DATA_DIR/data/state.db" "$TM_TEMP/data"

# Use the height of the ndau chain as an idenifier for what's in this snapshot.
HEIGHT=$((36#$("$BIN_DIR"/noms show "$NOMS_DATA_DIR"::ndau.value.Height | tr -d '"')))
SNAPSHOT_NAME=snapshot-$NETWORK-$HEIGHT
SNAPSHOT_FILE="$NDAU_SNAPSHOTS_DIR/$SNAPSHOT_NAME.tgz"

# Make the tarball and remove the temp dir.
cd "$SNAPSHOT_TEMP_DIR" || exit 1
tar -czf "$SNAPSHOT_FILE" data
rm -rf "$SNAPSHOT_TEMP_DIR"

echo "Snapshot created: $SNAPSHOT_FILE"

# Flag the snapshot as ready to be copied out of the container.
export SNAPSHOT_RESULT=$SNAPSHOT_FILE
