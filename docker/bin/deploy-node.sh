#!/bin/bash
# deploy single devnet node

# get the directory of this file
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# args
node_number=$1
network_name=$2
identity_folder=$3
[ -z "$identity_folder" ] && identity_folder=.

# consts
TMP_FILE="$DIR/temp-docker-compose.yml"
TEMPLATE_FILE="$DIR/node-template.yml"
IDENTITY_FILE="$identity_folder/node-identity-${node_number}.tgz"
ECS_PARAMS_FILE="$DIR/ecs-params.yml"

CPU_SHARES_DEFAULT=150 # 250 = 25% of a vcpu
MEM_LIMIT_DEFAULT=2000000000 # 2GB

ecs_params() {
  cat <<EOF
version: 1
task_definition:
  services:
    {{NETWORK_NAME}}-{{NODE_NUMBER}}:
      cpu_shares: {{CPU_SHARES}}
      mem_limit: {{MEM_LIMIT}}
EOF
}

errcho() { >&2 echo -e "$@"; }
usage() {
  errcho "Usage: $0 node_number network_name identity_folder"
  errcho "  e.g.: $0 0 devnet https://a.b.c ."
  errcho "    node_number is the number of the node (e.g. 0 or 4)"
  errcho "    network_name is the name of the network (e.g. devnet, testnet, mainnet)"
  errcho "    identity_folder must contain files named 'node-identity-X.tgz' where X matches a node_number."
  errcho ""
  errcho "  environment variables"
  errcho "    SNAPSHOT_URL is the url of a snapshot to restore from."
}

# Warn if things didn't close down properly last execution
if [ -f "$TMP_FILE" ]; then
  errcho "temp docker-compose file already exists: $TMP_FILE"
  exit 1
fi

# Warn if things didn't close down properly last execution
if [ -f "$ECS_PARAMS_FILE" ]; then
  errcho "temp ecs_params file already exists: $ECS_PARAMS_FILE"
  exit 1
fi

# Make sure template file is there
if [ ! -f "$TEMPLATE_FILE" ]; then
  errcho "template file not found: $TEMPLATE_FILE"
  exit 1
fi

# Make sure identity file is there
if [ ! -f "$IDENTITY_FILE" ]; then
  errcho "Identity file not found: $IDENTITY_FILE"
  exit 1
fi


CPU_SHARES=${CPU_SHARES:-$CPU_SHARES_DEFAULT}
MEM_LIMIT=${MEM_LIMIT:-$MEM_LIMIT_DEFAULT}

# Test to see if the snapshot url exists.
if ! curl --output /dev/null --silent --head --fail "$SNAPSHOT_URL"; then
  errcho "Snapshot URL doesn't exist: $SNAPSHOT_URL"
fi

port_offset=$PORT_OFFSET
[ -z "$port_offset" ] && port_offset=0

base_port=30000
rpc_port=$((base_port     + 100 + node_number + port_offset))
p2p_port=$((base_port     + 200 + node_number + port_offset))
ndauapi_port=$((base_port + 300 + node_number + port_offset))

cat "$TEMPLATE_FILE" | \
  sed \
    -e "s/{{TAG}}/${SHA}/g" \
    -e "s/{{NODE_NUMBER}}/${node_number}/g" \
    -e "s%{{BASE64_NODE_IDENTITY}}%$(cat "$IDENTITY_FILE" | base64 -w0)%g" \
    -e "s*{{SNAPSHOT_URL}}*${SNAPSHOT_URL}*g" \
    -e "s/{{PERSISTENT_PEERS}}/${PERSISTENT_PEERS}/g" \
    -e "s/{{RPC_PORT}}/${rpc_port}/g" \
    -e "s/{{P2P_PORT}}/${p2p_port}/g" \
    -e "s/{{NDAUAPI_PORT}}/${ndauapi_port}/g" \
    -e "s/{{NETWORK_NAME}}/${network_name}/g" \
  > "$TMP_FILE"
cat "$TMP_FILE"

ecs_params | sed \
    -e "s/{{NETWORK_NAME}}/${network_name}/g" \
    -e "s/{{NODE_NUMBER}}/${node_number}/g" \
    -e "s/{{CPU_SHARES}}/${CPU_SHARES}/g" \
    -e "s/{{MEM_LIMIT}}/${MEM_LIMIT}/g" \
  > "$ECS_PARAMS_FILE"

# Send it to AWS
ecs-cli compose \
  --verbose \
  --project-name ${network_name}-${node_number} \
  -f ${TMP_FILE} \
  service up \
  --create-log-groups \
  --cluster-config "$CLUSTER_NAME"

# clean up
rm "$TMP_FILE" "$ECS_PARAMS_FILE"
