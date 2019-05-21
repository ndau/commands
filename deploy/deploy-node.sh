#!/bin/bash
# deploy single node

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

# test base64 capibilities
if echo "A" | base64 -w0 2> /dev/null; then
  b64_opts="-w0"
else
  b64_opts=""
fi

cat "$TEMPLATE_FILE" | \
  sed \
    -e "s/{{TAG}}/${SHA}/g" \
    -e "s/{{NODE_NUMBER}}/${node_number}/g" \
    -e "s%{{BASE64_NODE_IDENTITY}}%$(cat "$IDENTITY_FILE" | base64 $b64_opts)%g" \
    -e "s/{{PERSISTENT_PEERS}}/${PERSISTENT_PEERS}/g" \
    -e "s/{{HONEYCOMB_KEY}}/${HONEYCOMB_KEY}/g" \
    -e "s/{{RPC_PORT}}/${rpc_port}/g" \
    -e "s/{{P2P_PORT}}/${p2p_port}/g" \
    -e "s/{{NDAUAPI_PORT}}/${ndauapi_port}/g" \
    -e "s/{{NETWORK_NAME}}/${network_name}/g" \
  > "$TMP_FILE"
cat "$TMP_FILE"

# Send it to AWS
ecs-cli compose \
  --verbose \
  --project-name ${network_name}-${node_number} \
  -f ${TMP_FILE} \
  service up \
  --force-deployment true \
  --cluster-config "$CLUSTER_NAME"

# clean up
rm "$TMP_FILE"
