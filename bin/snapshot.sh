#!/bin/bash

# Halt on any error.
set -e

# Load our environment variables.
CMDBIN_DIR="$(go env GOPATH)/src/github.com/oneiro-ndev/commands/bin"
# shellcheck disable=SC1090
source "$CMDBIN_DIR"/env.sh

NETWORK="$1"
if [ -z "$NETWORK" ]; then
    echo "Usage:"
    echo "  ./snapshot.sh NETWORK"
    echo "Examples:"
    echo "  ./snapshot.sh devnet"
    echo "  ./snapshot.sh mainnet"
    exit 1
fi

# Check a common case: not having run localnet at least once.  The redis directory 
if [ ! -d "$REDIS_NDAU_DATA_DIR-0" ]; then
    echo "Must ./run.sh localnet at least once before generating a snapshot"
    exit 1
fi

# Kill everything gracefully first, so that all data files are dumped by node group processes.
echo "Ensuring localnet is not running..."
"$CMDBIN_DIR"/kill.sh 1>/dev/null

# Prepare the output directory for the snapshot.
echo "Removing any old snapshots..."
NDAU_SNAPSHOTS_SUBDIR=ndau-snapshots
NDAU_SNAPSHOTS_DIR="$CMDBIN_DIR/$NDAU_SNAPSHOTS_SUBDIR"
rm -rf "$NDAU_SNAPSHOTS_DIR"
mkdir -p "$NDAU_SNAPSHOTS_DIR"

# Make a temp dir for copying data files into for tar'ing up later in this script.
SNAPSHOT_TEMP_DIR="$CMDBIN_DIR"/snapshot-temp
rm -rf "$SNAPSHOT_TEMP_DIR"
mkdir -p "$SNAPSHOT_TEMP_DIR"
SNAPSHOT_DATA_DIR="$SNAPSHOT_TEMP_DIR/data"

# Use the deep tendermint data directories to create all the parent subdirectories we need.
TM_TEMP="$SNAPSHOT_DATA_DIR/tendermint"
mkdir -p "$TM_TEMP/config"
mkdir -p "$TM_TEMP/data"

# Make the node identity tarball(s) first.
echo "Building node identity files..."
for node_num in $(seq 0 "$HIGH_NODE_NUM");
do
    name="node-identity-$node_num"

    echo "  bundling $name..."

    cd "$TENDERMINT_NDAU_DATA_DIR-$node_num" || exit 1
    cp config/node_key.json "$TM_TEMP/config"
    cp config/priv_validator_key.json "$TM_TEMP/config"

    cd "$SNAPSHOT_DATA_DIR"
    tar -czf "$NDAU_SNAPSHOTS_DIR/$name.tgz" \
        tendermint/config/node_key.json \
        tendermint/config/priv_validator_key.json

    # Get rid of these files so they're not part of the snapshot.
    rm -rf "$TM_TEMP"/config/*
    rm -rf "$TM_TEMP"/data/*
done

# Copy all the data files we want into the temp dir.
cp -r "$NOMS_NDAU_DATA_DIR-0" "$SNAPSHOT_DATA_DIR/noms"
cp -r "$REDIS_NDAU_DATA_DIR-0" "$SNAPSHOT_DATA_DIR/redis"
cp "$TENDERMINT_NDAU_DATA_DIR-0/config/genesis.json" "$TM_TEMP/config"
cp -r "$TENDERMINT_NDAU_DATA_DIR-0/data/blockstore.db" "$TM_TEMP/data"
cp -r "$TENDERMINT_NDAU_DATA_DIR-0/data/state.db" "$TM_TEMP/data"

# Use something better than "test-chain-..." for the chain_id.
genesis_file="$TM_TEMP/config/genesis.json"
jq ".chain_id=\"$NETWORK\"" "$genesis_file" > "$genesis_file.temp"
mv "$genesis_file.temp" "$genesis_file"

# Use the height of the ndau chain as an idenifier for what's in this snapshot.
HEIGHT=$((36#$("$NOMS_DIR"/noms show "$NOMS_NDAU_DATA_DIR-0"::ndau.value.Height | tr -d '"')))
SNAPSHOT_NAME=snapshot-$NETWORK-$HEIGHT
SNAPSHOT_FILE="$CMDBIN_DIR/$NDAU_SNAPSHOTS_SUBDIR/$SNAPSHOT_NAME.tgz"

# Make the tarball and remove the temp dir.
echo "  bundling $SNAPSHOT_NAME..."
cd "$SNAPSHOT_TEMP_DIR" || exit 1
tar -czf "$SNAPSHOT_FILE" data
rm -rf "$SNAPSHOT_TEMP_DIR"

# These can be used for uploading the snapshot to S3.
S3URI="s3://$NDAU_SNAPSHOTS_SUBDIR/$SNAPSHOT_NAME.tgz"
UPLOAD_CMD="aws s3 cp $SNAPSHOT_FILE $S3URI"

echo
echo "SNAPSHOT CREATED: $SNAPSHOT_FILE"
echo "NODE IDENTITY FILES CREATED: $NDAU_SNAPSHOTS_DIR/node-identity-*.tgz"
echo
echo "Next steps:"
echo "  1. Upload the snapshot to S3 using:"
echo "       $UPLOAD_CMD"
echo "  2. Make its name \"$SNAPSHOT_NAME\" known to people wanting to run a node with this snapshot on $NETWORK"
echo "  3. Back up the node-identity-*.tgz file(s) and keep them secure; use them to start (and restart any of) the first $NODE_COUNT nodes on $NETWORK"
echo
