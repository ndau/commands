#!/bin/bash
# deploy single node

set -e # exit on errors

# get the directory of this file
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# args
node_number=$1
network_name=$2
identity_folder=$3
[ -z "$identity_folder" ] && identity_folder=.

# consts
TEMPLATE_FILE="$DIR/node-template.yml"
IDENTITY_FILE="$identity_folder/node-identity-${node_number}.tgz"

RND=$(dd if=/dev/urandom count=8 bs=1 2> /dev/null | base64 | tr -dc 0-9a-zA-Z | head -c8)
TMP_FILE="$DIR/temp-docker-compose-$RND.yml"

source "$DIR/deploy-lib.sh"

usage() {
  errcho "$0 is mainly meant to be run in the CircleCI environment where most of these variables are suppplied."
  errcho "Usage: $0 node_number network_name identity_folder"
  errcho "  e.g.: $0 0 devnet ./ids"
  errcho "    node_number is the number of the node (e.g. 0 or 4)"
  errcho "    network_name is the name of the network (e.g. devnet)"
  errcho "    identity_folder must contain files named 'node-identity-X.tgz' where X matches a node_number."
  errcho ""
  errcho "  environment variables"
  errcho "    SHA is the 7-digit sha1 that matches a tag in ECR."
  errcho "    CLUSTER_NAME is the name of the cluster to deploy to."
  errcho "    [HONEYCOMB_KEY] is the honeycomb key to log to."
  errcho "    [SLACK_DEPLOYS_KEY] is the slack key to send deploy-related notifcations to."
  errcho "    [PERSISTENT_PEERS] is a comma separated list of peers for Tendermint (id@IP:port)."
}

if [ "$#" -ne 3 ]; then
    errcho "Error: need more arguments"
    usage
    exit 1
fi

# Warn if things didn't close down properly last execution
if [ -f "$TMP_FILE" ]; then
  errcho "Error: temp docker-compose file already exists: $TMP_FILE"
  exit 1
fi

# Make sure template file is there
if [ ! -f "$TEMPLATE_FILE" ]; then
  errcho "Error: template file not found: $TEMPLATE_FILE"
  exit 1
fi

# Make sure identity file is there
if [ ! -f "$IDENTITY_FILE" ]; then
  errcho "Error: Identity file not found: $IDENTITY_FILE"
  exit 1
fi

rpc_port=$(calc_port rpc $node_number)
p2p_port=$(calc_port p2p $node_number)
ndauapi_port=$(calc_port ndauapi $node_number)

# Make devnet-4 the one that takes periodic snapshots.
snapshot_interval=""
aws_access_key_id=""
aws_secret_access_key=""
if [ "$network_name" = "devnet" ] && [ "$node_number" = "4" ]; then
    # These environment variables are defined on Circle.
    aws_access_key_id="$AWS_ACCESS_KEY_ID"
    aws_secret_access_key="$AWS_SECRET_ACCESS_KEY"

    # If they aren't set, log a warning and continue without snapshots set up on the node.
    if [ -z "$aws_access_key_id" ] || [ -z "$aws_secret_access_key" ]; then
        # Make sure they're both unset.
        aws_access_key_id=""
        aws_secret_access_key=""
        echo "Unable to find AWS env vars for taking snapshots on $network_name-$node_number"
    else
        snapshot_interval="12h"
        echo "Snapshots every $snapshot_interval will be done on $network_name-$node_number"
    fi
else
    echo "Snapshots are disabled on $network_name-$node_number"
fi

# Some versions of base64 inject newlines; strip them.
# Doing it this way works with more versions of base64 than using -w0, for example.
BASE64_NODE_IDENTITY=$(cat "$IDENTITY_FILE" | base64 | tr -d \\n)

cat "$TEMPLATE_FILE" | \
  sed \
    -e "s/{{TAG}}/${SHA}/g" \
    -e "s/{{NODE_NUMBER}}/${node_number}/g" \
    -e "s%{{BASE64_NODE_IDENTITY}}%${BASE64_NODE_IDENTITY}%g" \
    -e "s/{{PERSISTENT_PEERS}}/${PERSISTENT_PEERS}/g" \
    -e "s/{{HONEYCOMB_KEY}}/${HONEYCOMB_KEY}/g" \
    -e "s/{{SLACK_DEPLOYS_KEY}}/${SLACK_DEPLOYS_KEY}/g" \
    -e "s/{{RPC_PORT}}/${rpc_port}/g" \
    -e "s/{{P2P_PORT}}/${p2p_port}/g" \
    -e "s/{{NDAUAPI_PORT}}/${ndauapi_port}/g" \
    -e "s/{{NETWORK_NAME}}/${network_name}/g" \
    -e "s/{{AWS_ACCESS_KEY_ID}}/${aws_access_key_id}/g" \
    -e "s/{{AWS_SECRET_ACCESS_KEY}}/${aws_secret_access_key}/g" \
    -e "s/{{SNAPSHOT_INTERVAL}}/${snapshot_interval}/g" \
  > "$TMP_FILE"
cat "$TMP_FILE"

# Send the new task definition to AWS and update the service for the node.
echo "Updating $network_name-$node_number..."
ecs-cli compose \
        --project-name ${network_name}-${node_number} \
        -f ${TMP_FILE} \
        service up \
        --cluster-config "$CLUSTER_NAME"

# clean up
rm "$TMP_FILE"

# Wait for the node to become healthy.
echo "Waiting for $network_name-$node_number to become healthy..."
# Only devnet nodes are all on the same instance.
if [ "$network_name" = "devnet" ]; then
    cname="$network_name"
    port="303$node_number"
else
    cname="$network_name-$node_number"
    port="3030"
fi

# Give the old service some time to shut down.  If we check the health too soon after updating
# the service, it'll come back as "OK" but it will be the old service responding.
sleep 20

for i in {1..60}; do
    printf "$node_number" # Use the number we're waiting for since this script is backgrounded.
    health=$(curl -s "https://$cname.ndau.tech:$port/health")
    if [ "$health" = '"OK"' ]; then
        printf "\n"
        echo "$network_name-$node_number is healthy; its deploy is complete"
        exit 0
    fi
    sleep 2
done

echo "Timed out waiting for $network_name-$node_number to become healthy"
exit 1
