#!/bin/bash

# Halt on any error.
set -e

# Load our environment variables.
CMDBIN_DIR="$(go env GOPATH)/src/github.com/oneiro-ndev/commands/bin"
# shellcheck disable=SC1090
source "$CMDBIN_DIR"/env.sh

# Check a common case: not having run localnet at least once.  The redis directory 
if [ ! -d "$REDIS_NDAU_DATA_DIR-0" ]; then
    echo "Must ./run.sh localnet at least once before generating a snapshot"
    exit 1
fi

# Make sure you're running the right version of redis, to match what the container is using.
REDIS_VERSION_EXPECTED=5.0
REDIS_VERSION_ACTUAL=$(redis-server --version | \
                           sed -n -e 's/^Redis server v=\([0-9]*\.[0-9]*\)\.[0-9]* .*/\1/p')
if [ "$REDIS_VERSION_ACTUAL" != "$REDIS_VERSION_EXPECTED" ]; then
    echo "Must have Redis version $REDIS_VERSION_EXPECTED installed to generate a valid snapshot"
    exit 1
fi

# Since the chain_id was specified at setup-time, we make sure the user really wants to use it
# as the network name.  Otherwise they'll get the default snapshot for localnet.
NETWORK="$CHAIN_ID"
echo "Generating snapshot for the network: $NETWORK"
printf "Is this the right network? [y|n]: "
read CONFIRM
if [ "$CONFIRM" != "y" ]; then
    echo "You can change the network name by running setup.sh or reset.sh with a new name"
    exit 1
fi

# Kill everything gracefully first, so that all data files are dumped by node group processes.
echo "Ensuring localnet is not running..."
"$CMDBIN_DIR"/kill.sh 1>/dev/null

# Prepare the output directory for the snapshot under the docker directory.
echo "Removing any old snapshots..."
NDAU_SNAPSHOTS_SUBDIR=ndau-snapshots
NDAU_SNAPSHOTS_DIR="$CMDBIN_DIR/../docker/$NDAU_SNAPSHOTS_SUBDIR"
rm -rf "$CMDBIN_DIR/$NDAU_SNAPSHOTS_SUBDIR" # Remove the old location, too.
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

# Prepare for modifying peer ids in circle config.
echo "Processing peers..."
MODIFY_CONFIG_YML=false
PERSISTENT_PEERS=()
PEER_HOSTS=()
PEER_PORTS=()
CONFIG_YML_NAME=".circleci/config.yml"
CONFIG_YML_PATH="$CMDBIN_DIR/../$CONFIG_YML_NAME"
# CircleCI isn't used for localnet (not applicable) or mainnet (manual deploy).
if [ "$NETWORK" != "localnet" ] && [ "$NETWORK" != "mainnet" ]; then
    set +e
    grep '^ *PERSISTENT_PEERS: .* # '"$NETWORK"'$' "$CONFIG_YML_PATH" > /dev/null
    GREP_RESULT="$?"
    set -e
    if [ "$GREP_RESULT" = 0 ]; then
        name="PERSISTENT_PEERS"
        p=$(sed -n -e 's|^\( *'"$name"': \)\(.*\)\( # '"$NETWORK"'\)$|\2|p' "$CONFIG_YML_PATH")
        IFS=',' read -ra peers <<< "$p"
        for peer in "${peers[@]}"; do
            IFS='@' read -ra split <<< "$peer"
            host_and_port="${split[1]}"

            IFS=':' read -ra split <<< "$host_and_port"
            PEER_HOSTS+=("${split[0]}")
            PEER_PORTS+=("${split[1]}")
        done
        # The number of peers listed in the yml must match the number of nodes in the snapshot.
        if [ "${#PEER_HOSTS[@]}" = "$NODE_COUNT" ]; then
            MODIFY_CONFIG_YML=true
        fi
    fi
fi

# Make the node identity tarball(s) first.
echo "Building node identity files..."
NODE_IDENTITY_NAME="node-identity"
for node_num in $(seq 0 "$HIGH_NODE_NUM");
do
    name="$NODE_IDENTITY_NAME-$node_num"
    tm_home="$TENDERMINT_NDAU_DATA_DIR-$node_num"
    tm_config="$TM_TEMP/config"

    echo "  bundling $name..."

    cd "$tm_home" || exit 1
    cp config/node_key.json "$tm_config"
    cp config/priv_validator_key.json "$tm_config"

    cd "$SNAPSHOT_DATA_DIR" || exit 1
    tar -czf "$NDAU_SNAPSHOTS_DIR/$name.tgz" \
        tendermint/config/node_key.json \
        tendermint/config/priv_validator_key.json

    peer_id=$("$TENDERMINT_DIR/tendermint" show_node_id --home "$tm_home")
    echo "    peer id: $peer_id"
    if [ "$MODIFY_CONFIG_YML" = true ]; then
        peer_host=${PEER_HOSTS[$node_num]}
        peer_port=${PEER_PORTS[$node_num]}
        PERSISTENT_PEERS+=("$peer_id@$peer_host:$peer_port")
    fi

    # Get rid of these files so they're not part of the snapshot.
    rm -rf "$tm_config"/*
done

# Make the all-in-one node identities file, preserving the individual ones for local testing.
cd "$NDAU_SNAPSHOTS_DIR" || exit 1
IDENTITIES_FILE=node-identities-$NETWORK.tgz
IDENTITIES_PATH="$NDAU_SNAPSHOTS_DIR/$IDENTITIES_FILE"
mkdir -p tmp
cp "$NODE_IDENTITY_NAME"-*.tgz tmp
tar -czf "$IDENTITIES_FILE" "$NODE_IDENTITY_NAME"-*.tgz
mv tmp/"$NODE_IDENTITY_NAME"-*.tgz .
rm -rf tmp

# Copy all the data files we want into the temp dir.
echo "Building snapshot file..."
cp -r "$NOMS_NDAU_DATA_DIR-0" "$SNAPSHOT_DATA_DIR/noms"
cp -r "$REDIS_NDAU_DATA_DIR-0" "$SNAPSHOT_DATA_DIR/redis"
cp "$TENDERMINT_NDAU_DATA_DIR-0/config/genesis.json" "$TM_TEMP/config"
cp -r "$TENDERMINT_NDAU_DATA_DIR-0/data/blockstore.db" "$TM_TEMP/data"
cp -r "$TENDERMINT_NDAU_DATA_DIR-0/data/state.db" "$TM_TEMP/data"

# Use the height of the ndau chain as an idenifier for what's in this snapshot.
HEIGHT=$((36#$("$NOMS_DIR"/noms show "$NOMS_NDAU_DATA_DIR-0"::ndau.value.Height | tr -d '"')))
SNAPSHOT_NAME=snapshot-$NETWORK-$HEIGHT
SNAPSHOT_PATH="$NDAU_SNAPSHOTS_DIR/$SNAPSHOT_NAME.tgz"

# Make the tarball and remove the temp dir.
echo "  bundling $SNAPSHOT_NAME..."
cd "$SNAPSHOT_TEMP_DIR" || exit 1
tar -czf "$SNAPSHOT_PATH" data
cd .. || exit 1
rm -rf "$SNAPSHOT_TEMP_DIR"

# Make the "latest" file.
LATEST_FILE="latest-$NETWORK.txt"
LATEST_PATH="$NDAU_SNAPSHOTS_DIR/$LATEST_FILE"
echo "$SNAPSHOT_NAME" > "$LATEST_PATH"

# Update the circle config with the new peer ids and snapshot name.
if [ "$MODIFY_CONFIG_YML" = true ]; then
    echo "Modifying $CONFIG_YML_NAME..."
    persistent_peers=$(join_by , "${PERSISTENT_PEERS[@]}")
    sed -i '' -E \
        -e 's|^( *PERSISTENT_PEERS: )(.*)( # '"$NETWORK"')$|\1'"$persistent_peers"'\3|' \
        "$CONFIG_YML_PATH"
else
    # If this happens, it could mean that the anchor comment is missing, e.g. "... # mainnet",
    # or the number of nodes in the snapshot differs from the number of peers found in the yml.
    # It's non-fatal; anyone wanting to re-deploy will have to take care of it manually.
    echo "Unable to modify PERSISTENT_PEERS for $NETWORK in $CONFIG_YML_PATH"
fi

# These can be used for uploading the snapshot to S3.
S3_SNAPSHOT_URI="s3://$NDAU_SNAPSHOTS_SUBDIR"
S3_IDENTITIES_URI="s3://ndau-deploy-secrets"
UPLOAD_SNAPSHOT_CMD="aws s3 cp $SNAPSHOT_PATH $S3_SNAPSHOT_URI/$SNAPSHOT_NAME.tgz"
UPLOAD_IDENTITIES_CMD="aws s3 cp $IDENTITIES_PATH $S3_IDENTITIES_URI/$IDENTITIES_FILE"
UPLOAD_LATEST_CMD="aws s3 cp $LATEST_PATH $S3_SNAPSHOT_URI/$LATEST_FILE"

echo
echo "SNAPSHOT CREATED: $SNAPSHOT_PATH"
echo "NODE IDENTITY FILES CREATED (separate): $NDAU_SNAPSHOTS_DIR/$NODE_IDENTITY_NAME-*.tgz"
echo "NODE IDENTITY FILES CREATED (all in 1): $NDAU_SNAPSHOTS_DIR/$IDENTITIES_FILE"
echo "LATEST FILE CREATED: $LATEST_PATH"
echo
echo "Next steps:"
echo "  1. Upload the snapshot to S3 using:"
echo "       $UPLOAD_SNAPSHOT_CMD"
echo "  2. Upload the node identities file to S3 using:"
echo "       $UPLOAD_IDENTITIES_CMD"
echo "  3. Use this to mark it as the latest snapshot on $NETWORK if desired:"
echo "       $UPLOAD_LATEST_CMD"

if [ "$MODIFY_CONFIG_YML" = true ]; then
    echo "  4. Your copy of $CONFIG_YML_NAME has been modified; commit it if desired"
else
    echo "  4. The $CONFIG_YML_NAME file might need PERSISTENT_PEERS updated"
fi

echo
