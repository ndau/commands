#!/bin/bash

set -eo pipefail

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
cd "$SCRIPT_DIR" || exit 1

# Run a local node connected to mainnet starting from the genesis snapshot.
nodename="catchup-node-local"
snapshot="snapshot-mainnet-1"
USE_LOCAL_IMAGE=1 \
            AWS_ACCESS_KEY_ID="" \
            AWS_SECRET_ACCESS_KEY="" \
            SLACK_DEPLOYS_KEY="" \
../bin/runcontainer.sh mainnet "$nodename" 26660 26670 3030 "" "$snapshot"

echo

# Get the current height of mainnet.  We need to catch up to at least this height.
# Use mainnet-2 since that's in the same region as devnet.
status=$(curl -s https://mainnet-2.ndau.tech:26670/status)
mainnet_height=$(echo "$status" | jq -r .result.sync_info.latest_block_height)
if [ -z "$mainnet_height" ] || [ "$mainnet_height" -le 0 ]; then
    echo "Unable to get mainnet height"
    false
fi
echo "Current mainnet height: $mainnet_height"

# Catching up on mainnet will take longer and longer as the block height of mainnet
# increases over time.  This is known and we'll need to deal with it at some point.
# For now, we've agreed to wait as long as it takes.  However, it'll eventually be
# "too long".  Let's put a 20-minute cap on it for now.  If catchup does not complete
# in that amount of time, the circle job will fail and we'll be forced to deal with
# it in some way.  Either by accepting longer build times or by inventing a way to
# validate catchup compatibility another way, perhaps external to circle workflows.
printf "Catching up..."
last_height=0

# we don't need to ever use the loop index, but we can't omit it
# shellcheck disable=SC2034
while : ; do
    sleep 10
    if ! node_status=$(docker exec "$nodename" curl -s http://localhost:26670/status); then
        # The status query is what usually fails when playback of a block fails.
        echo " (ERROR: unable to catch up)"
        noms_fn=$(printf "snapshot-catchup-failure-%s.tar.bz2" "$last_height")
        echo "creating noms tarball: $noms_fn"
        docker exec "$nodename" tar c -j -f "/$noms_fn" -C "/image/data" noms
        echo "extracting noms tarball from container"
        docker cp "$nodename:/$noms_fn" ../..

        echo "attempting to find a mainnet snapshot high enough"
        snapshot_pair=$(
            aws s3 ls ndau-snapshots/snapshot-mainnet |\
            tr -s ' ' |\
            cut -d' ' -f4 |\
            sed -E -e 's/^([^[:digit:]]*([[:digit:]]+).*)$/\2 \1/g' |\
            sort -rn |\
            head -n 1
        )
        if [ -z "$snapshot_pair" ]; then
            echo "no snapshot could be found"
            break
        fi
        mainnet_snapshot_height=$(echo "$snapshot_pair" | cut -d' ' -f1)
        if [[ ! "$mainnet_snapshot_height" -gt "$last_height" ]]; then
            echo "highest mainnet snapshot was $mainnet_snapshot_height; need at least $((last_height+1))"
            break
        fi
        mainnet_snapshot_name=$(echo "$snapshot_pair" | cut -d' ' -f2)
        echo "fetching mainnet snapshot: $mainnet_snapshot_name"
        aws s3 cp "s3://ndau-snapshots/$mainnet_snapshot_name" ../..

        break
    fi

    node_height=$(echo "$node_status" | jq -r .result.sync_info.latest_block_height)
    if [ -z "$node_height" ]; then
        # If we didn't get a height back, something went wrong; assume failed catchup.
        printf " (ERROR: no height)"
        break
    fi
    printf " %s" "$node_height"

    catching_up=$(echo "$node_status" | sed -n -e 's/.*catching_up...\([a-z]\{1,\}\).*/\1/p')
    if [ "$catching_up" = "false" ] && [ "$node_height" -ge "$mainnet_height" ]; then
        caught_up=1
        printf " (caught up)"
        break
    fi

    if [ "$node_height" -le "$last_height" ]; then
        # Fail if we didn't catch up at all since the last iteration.
        # This indicates a stall, which likely means we're failing on full catchup.
        printf " (ERROR: stalled)"
        break
    fi

    last_height=$node_height
done
printf "\n"

echo

# Stop and remove the container instance for the catchup test node.
../bin/removecontainer.sh "$nodename"

echo

if [ -z "$caught_up" ]; then
    echo "Catchup failed"
    false
fi

echo "Catchup complete"
