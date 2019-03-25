#!/bin/bash

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

# Kill everything gracefully first, so that all data files are dumped by node group processes.
"$CMDBIN_DIR"/kill.sh 1>/dev/null

# Prepare the output directory for the snapshot.
NDAU_SNAPSHOTS_SUBDIR=ndau-snapshots
NDAU_SNAPSHOTS_DIR="$CMDBIN_DIR/$NDAU_SNAPSHOTS_SUBDIR"
rm -rf "$NDAU_SNAPSHOTS_DIR"
mkdir -p "$NDAU_SNAPSHOTS_DIR"

# Make the private config tarball(s).
for node_num in $(seq 0 "$HIGH_NODE_NUM");
do
    echo "Bundling private-chaos-$node_num..."
    cd "$TENDERMINT_CHAOS_DATA_DIR-$node_num/config" || exit 1
    tar -czf "$NDAU_SNAPSHOTS_DIR/private-chaos-$node_num.tgz" *_key.json

    echo "Bundling private-ndau-$node_num..."
    cd "$TENDERMINT_NDAU_DATA_DIR-$node_num/config" || exit 1
    tar -czf "$NDAU_SNAPSHOTS_DIR/private-ndau-$node_num.tgz" *_key.json
done

# Make a temp dir for copying data files into for tar'ing up later in this script.
SNAPSHOT_TEMP_DIR="$CMDBIN_DIR"/snapshot-temp
rm -rf "$SNAPSHOT_TEMP_DIR"
mkdir -p "$SNAPSHOT_TEMP_DIR"
SNAPSHOT_DATA_DIR="$SNAPSHOT_TEMP_DIR/data"

# Use the deep tendermint data directories to create all the parent subdirectories we need.
TM_CHAOS_TEMP="$SNAPSHOT_DATA_DIR/chaos/tendermint"
TM_NDAU_TEMP="$SNAPSHOT_DATA_DIR/ndau/tendermint"
mkdir -p "$TM_CHAOS_TEMP/config"
mkdir -p "$TM_CHAOS_TEMP/data"
mkdir -p "$TM_NDAU_TEMP/config"
mkdir -p "$TM_NDAU_TEMP/data"

# Get the SVI Namespace from the ndau config.toml file.
SVI_NAMESPACE=$(awk '/^  Namespace/{print $NF}' "$NODE_DATA_DIR-0"/ndau/config.toml | tr -d '"')
echo "$SVI_NAMESPACE" > "$SNAPSHOT_TEMP_DIR/svi-namespace"

# Copy all the data files we want into the temp dir.
cp -r "$NOMS_CHAOS_DATA_DIR-0" "$SNAPSHOT_DATA_DIR/chaos/noms"
cp -r "$NOMS_NDAU_DATA_DIR-0" "$SNAPSHOT_DATA_DIR/ndau/noms"
cp -r "$REDIS_CHAOS_DATA_DIR-0" "$SNAPSHOT_DATA_DIR/chaos/redis"
cp -r "$REDIS_NDAU_DATA_DIR-0" "$SNAPSHOT_DATA_DIR/ndau/redis"
cp "$TENDERMINT_CHAOS_DATA_DIR-0/config/genesis.json" "$TM_CHAOS_TEMP/config"
cp -r "$TENDERMINT_CHAOS_DATA_DIR-0/data/blockstore.db" "$TM_CHAOS_TEMP/data"
cp -r "$TENDERMINT_CHAOS_DATA_DIR-0/data/state.db" "$TM_CHAOS_TEMP/data"
cp "$TENDERMINT_NDAU_DATA_DIR-0/config/genesis.json" "$TM_NDAU_TEMP/config"
cp -r "$TENDERMINT_NDAU_DATA_DIR-0/data/blockstore.db" "$TM_NDAU_TEMP/data"
cp -r "$TENDERMINT_NDAU_DATA_DIR-0/data/state.db" "$TM_NDAU_TEMP/data"

# Use the height of the ndau chain as an idenifier for what's in this snapshot.
HEIGHT=$((36#$("$NOMS_DIR"/noms show "$NOMS_NDAU_DATA_DIR-0"::ndau.value.Height | tr -d '"')))
SNAPSHOT_NAME=snapshot-$NETWORK-$HEIGHT
SNAPSHOT_FILE="$CMDBIN_DIR/$NDAU_SNAPSHOTS_SUBDIR/$SNAPSHOT_NAME.tgz"

# Make the tarball and remove the temp dir.
echo "Bundling $SNAPSHOT_NAME..."
cd "$SNAPSHOT_TEMP_DIR" || exit 1
tar -czf "$SNAPSHOT_FILE" svi-namespace data
rm -rf "$SNAPSHOT_TEMP_DIR"

# These can be used for uploading the snapshot to S3.
S3URI="s3://$NDAU_SNAPSHOTS_SUBDIR/$SNAPSHOT_NAME.tgz"
UPLOAD_CMD="aws s3 cp $SNAPSHOT_FILE $S3URI"

echo
echo "SNAPSHOT CREATED: $SNAPSHOT_FILE"
echo "PRIVATE FILES CREATED: $NDAU_SNAPSHOTS_DIR/private-*.tgz"
echo "Next steps:"
echo "  1. Upload the snapshot to S3 using: $UPLOAD_CMD"
echo "  2. Make its name \"$SNAPSHOT_NAME\" known to people wanting to run a node with this snapshot on $NETWORK"
echo "  3. Back up the private-*.tgz file(s) and keep them secure; use them to start the first $NODE_COUNT nodes on $NETWORK"
