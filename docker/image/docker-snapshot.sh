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
TM_TEMP="$SNAPSHOT_DATA_DIR/tendermint"
mkdir -p "$TM_TEMP/config"
mkdir -p "$TM_TEMP/data"

# Copy all the data files we want into the temp dir.
# JSG don't do a copy of noms, instead do a sync which compresses the DB
# cp -r "$NOMS_DATA_DIR" "$SNAPSHOT_DATA_DIR/noms"
mkdir -p "$SNAPSHOT_DATA_DIR/noms"
"$BIN_DIR"/noms set new database "$SNAPSHOT_DATA_DIR/noms"::ndau
"$BIN_DIR"/noms sync "$NOMS_DATA_DIR"::ndau "$SNAPSHOT_DATA_DIR/noms"::ndau
# EJM delete all obsolete redis snapshot files before copying
rm "$REDIS_DATA_DIR/temp*"
cp -r "$REDIS_DATA_DIR" "$SNAPSHOT_DATA_DIR/redis"
cp "$TM_DATA_DIR/config/genesis.json" "$TM_TEMP/config"
cp -r "$TM_DATA_DIR/data/blockstore.db" "$TM_TEMP/data"
cp -r "$TM_DATA_DIR/data/state.db" "$TM_TEMP/data"

# Use the height of the ndau chain as an idenifier for what's in this snapshot.
HEIGHT=$((36#$("$BIN_DIR"/noms show "$NOMS_DATA_DIR"::ndau.value.Height | tr -d '"')))
SNAPSHOT_NAME="snapshot-$NETWORK-$HEIGHT"
SNAPSHOT_PATH="$SCRIPT_DIR/$SNAPSHOT_NAME.tgz"

# Make the tarball and remove the temp dir.
cd "$SNAPSHOT_TEMP_DIR" || exit 1
tar -czf "$SNAPSHOT_PATH" data
cd .. || exit 1
rm -rf "$SNAPSHOT_TEMP_DIR"

upload_to_s3() {
    file_name="$1"
    content_type="$2"

    aws_key="$AWS_ACCESS_KEY_ID"
    aws_secret="$AWS_SECRET_ACCESS_KEY"

    date_str=$(date -R)
    s3_path="/$SNAPSHOT_BUCKET/$file_name"
    signable_bytes="PUT\n\n$content_type\n$date_str\n$s3_path"
    signature=$(echo -en "$signable_bytes" | openssl sha1 -hmac "$aws_secret" -binary | base64)

    echo "Uploading $file_name to S3..."
    curl -s -X PUT -T "$SCRIPT_DIR/$file_name" \
         -H "Host: s3.amazonaws.com" \
         -H "Date: $date_str" \
         -H "Content-Type: $content_type" \
         -H "Authorization: AWS $aws_key:$signature" \
         "$SNAPSHOT_URL$s3_path"
}

# Optionally upload the snapshot to the S3 bucket, but only if we have the AWS credentials.
if [ "$2" = "--upload" ] && [ ! -z "$AWS_ACCESS_KEY_ID" ] && [ ! -z "$AWS_SECRET_ACCESS_KEY" ]
then
    # Make sure we don't clobber a snapshot with the same name.  This protects us against
    # multiple nodes being set up to upload snapshots.  It's something we should avoid doing.
    # But if it happens, the first node to upload a given height's snapshot "wins".
    file_name="$SNAPSHOT_NAME.tgz"
    if curl --output /dev/null --silent --head --fail "$SNAPSHOT_URL/$SNAPSHOT_BUCKET/$file_name"
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
        if ! upload_to_s3 "$file_name" "application/x-gtar"; then
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
