#!/bin/bash
# deploy single devnet node

# get the directory of this file
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# args
node_number=$1
network_name=$2

# consts
TMP_FILE="$DIR/node-docker-compose.yml"
TEMPLATE_FILE="$DIR/node-template.yml"
IDENTITY_FILE="$DIR/node-identity-${node_number}.tgz"

if [ -f "$TMP_FILE" ]; then
  >&2 echo "temp file already exists: $TMP_FILE"
  exit 1
fi

# Make sure template file is there
if [ ! -f "$TEMPLATE_FILE" ]; then
  >&2 echo "template file not found: $TEMPLATE_FILE"
  exit 1
fi

# Make sure identity file is there
if [ ! -f "$IDENTITY_FILE" ]; then
  >&2 echo "Identity file not found: $IDENTITY_FILE"
  exit 1
fi

port_offset=$PORT_OFFSET
[ -z "$port_offset" ] && port_offset=0

base_port=30000
rpc_port=$((base_port + 100 + node_number + port_offset))
p2p_port=$((base_port + 200 + node_number + port_offset))
ndauapi_port=$((base_port + 300 + node_number + port_offset))

cat "$TEMPLATE_FILE" | \
  sed \
    -e "s/{{NODE_NUMBER}}/${node_number}/g" \
    -e "s%{{BASE64_NODE_IDENTITY}}%$(cat "$IDENTITY_FILE" | base64)%g" \
    -e "s/{{PERSISTENT_PEERS}}/${PERSISTENT_PEERS}/g" \
    -e "s/{{RPC_PORT}}/${rpc_port}/g" \
    -e "s/{{P2P_PORT}}/${p2p_port}/g" \
    -e "s/{{NDAUAPI_PORT}}/${ndauapi_port}/g" \
    -e "s/{{NETWORK_NAME}}/${network_name}/g" \
  > "$TMP_FILE"
cat "$TMP_FILE"

# Send it to AWS
ecs-cli compose --verbose --project-name ${network_name}-${node_number} -f ${TMP_FILE} service up --create-log-groups --cluster-config sc-node

# clean up
rm "$TMP_FILE"
