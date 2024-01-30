#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
source "$SCRIPT_DIR"/docker-env.sh

# The name of the snapshot file (or an error message) will be written to this file
# for the outside world to access.
SNAPSHOT_RESULT="$SCRIPT_DIR/snapshot_result"

# To start a snapshot, run `docker exec <container> /image/docker-snapshot.sh` from the outside.
# Then procmon will pass in --generate as the flag to do the actual snapshot at the right time.
if [ "$1" != "--generate" ]; then
    rm -f "$SNAPSHOT_RESULT"
    killall -HUP procmon
    exit 0
fi

echo "Generating $NETWORK snapshot..."

# Remove any existing snapshot from the container.  The user should copy each one out every time.
rm -rf "$SCRIPT_DIR"/snapshot-*.tgz

# Make a temp dir for copying data files into for tar'ing up later in this script.
SNAPSHOT_TEMP_DIR="$SCRIPT_DIR"/snapshot-temp
rm -rf "$SNAPSHOT_TEMP_DIR"
mkdir -p "$SNAPSHOT_TEMP_DIR"
SNAPSHOT_DATA_DIR="$SNAPSHOT_TEMP_DIR/data"

# Use the deep tendermint data directories to create all the parent subdirectories we need.
# Copy all the data files we want into the temp dir.

TM_TEMP="$SNAPSHOT_DATA_DIR/tendermint"
mkdir -p "$TM_TEMP/config"
mkdir -p "$TM_TEMP/data"
cp "$TM_DATA_DIR/config/genesis.json" "$TM_TEMP/config"
cp -r "$TM_DATA_DIR/data/blockstore.db" "$TM_TEMP/data"
cp -r "$TM_DATA_DIR/data/state.db" "$TM_TEMP/data"

# EJM delete all obsolete redis snapshot files before copying
rm "$REDIS_DATA_DIR"/temp-*

# We're only using that part of the TM data that we just copied above. We're otherwise
# making a tarball of the entire data directory so we don't need to copy the whole
# thing (since it's getting very big).

mv "$TM_DATA_DIR" "$SCRIPT_DIR/tendermint"
mv "$TM_TEMP" "$TM_DATA_DIR"

# Use the height of the ndau chain as an idenifier for what's in this snapshot.
HEIGHT=$((36#$("$BIN_DIR"/noms show "$NOMS_DATA_DIR"::ndau.value.Height | tr -d '"')))
SNAPSHOT_NAME="snapshot-$NETWORK-$HEIGHT"
SNAPSHOT_PATH="$SCRIPT_DIR/$SNAPSHOT_NAME.tgz"

cd "$ROOT_DIR" || exit 1
# Use noms sync to clean up (and shrink) the database before the snapshot
cd data
mv noms noms-presync
/image/bin/noms sync noms-presync::ndau noms::ndau
rm -rf noms-presync
cd ..
# Make the snapshot tarball
tar -czf "$SNAPSHOT_PATH" data

# Put the TM live data directory back where it belongs

rm -rf "$TM_DATA_DIR"
mv "$SCRIPT_DIR/tendermint" "$DATA_DIR/tendermint"
rm -rf "$SNAPSHOT_TEMP_DIR"

upload_to_s3() {
    file_name="$1"

    # AWS credentials are stored in the appropriate environment variables,
    # so we can create the configuration file

    aws configure set aws_access_key_id "$AWS_ACCESS_KEY_ID"
    aws configure set aws_secret_access_key "$AWS_SECRET_ACCESS_KEY"

    cd "$SCRIPT_DIR"
    ./s3-multipart-upload.sh "$file_name" "$SNAPSHOT_BUCKET"
}

# Optionally upload the snapshot to the S3 bucket, but only if we have the AWS credentials.
if [ "$2" = "--upload" ] && [ ! -z "$AWS_ACCESS_KEY_ID" ] && [ ! -z "$AWS_SECRET_ACCESS_KEY" ]
then
    # Make sure we don't clobber a snapshot with the same name.  This protects us against
    # multiple nodes being set up to upload snapshots.  It's something we should avoid doing.
    # But if it happens, the first node to upload a given height's snapshot "wins".

    file_name="$SNAPSHOT_NAME.tgz"
    if curl --output /dev/null --silent --head --fail "$SNAPSHOT_BUCKET/$file_name"
    then
        if [ "$3" = "--force" ]; then
            # Give it a new "forced" name so that we don't clobber what's up there.
            # If there is a forced one, though, it will get clobbered by this.
            # Forcing the snapshot is used primarily when doing an upgrade-with-full-reindex.
            # We need the new snapshot with reindexed data to be uploaded even if there's a
            # snapshot already there with the same name (height).
            # NOTE: It would be better if we instead incorporated the SHA into the snapshot name.
            #       Then when we're doing an upgrade, the snapshot names can't possibly collide.
            SNAPSHOT_NAME="$SNAPSHOT_NAME-forced"
            file_name="$SNAPSHOT_NAME.tgz"
            new_path="$SCRIPT_DIR/$file_name"
            mv "$SNAPSHOT_PATH" "$new_path"
            SNAPSHOT_PATH="$new_path"
        else
            echo "Snapshot $file_name already exists on S3"
            file_name=""
        fi
    fi
    
    # If a defective 0-length snapshot was created, don't upload it.

    if [ ! -s "$file_name" ]; then
        file_name=""
        echo "Snapshot tarball exists but is empty, upload canceled."
    fi

    if [ -n "$file_name" ]; then
        if ! upload_to_s3 "$file_name"; then
            echo "Failed to upload snapshot"
        else
            LATEST_FILE="latest-$NETWORK.txt"
            LATEST_PATH="$SCRIPT_DIR/$LATEST_FILE"
            echo "$SNAPSHOT_NAME" > "$LATEST_PATH"
            upload_to_s3 "$LATEST_FILE" "text/plain"
        fi
    fi
fi

# Flag the snapshot as ready to be copied out of the container.
echo "$SNAPSHOT_NAME.tgz" > "$SNAPSHOT_RESULT"

echo "Snapshot ready: $SNAPSHOT_PATH"
