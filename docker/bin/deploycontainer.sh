#!/bin/bash

# This script deploys a single container node using AWS Fargate.
# The structure of this script is meant to be a customized version of a more generic script that can be used to deploy microservices with some modification for each project.
# Before a Fargate service can be created/updated the following AWS infrastructure needs to be configured.
# - [ ] A VPC for a group of nodes.
# - [ ] A subnet. This should be the APP_BASE_NAME suffixed with -public. The subnet must have a route table configuration that allows outgoing requests to 0.0.0.0/0 to go to an internet gateway. Otherwise anything in ECS will not have access to the internet. Containers will also not be able to be downloaded and will fail with a container pull timeout.
# A network ACL. This should allow traffic to your application ports and any outgoing connections.
# An ECS cluster. This should be the APP_BASE_NAME suffixed with -cluster.
# A security group. This should be the APP_BASE_NAME suffixed with -sg. This should allow traffic to your application ports and any outgoig connection.

# Required variables
# APP_BASE_NAME the base name for this application (e.g. sc-node)
# NODE_NAME the name of the node (e.g. sc-node-0). Used to identify the node in logs. `node_id` in Honeycomb.
# NODE_IDENTITY a base64 encoded tar ball containing two files `tendermint/config/priv_validator_key.json` and `tendermint/config/node_id.json`.
# SNAPSHOT to use to start the node's databases. (e.g. snapshot-devnet-g)
# SHA is the 7-digit SHA1 that the image is tagged in ECR.
# PERSISTENT_PEERS a formatted list of peers to connect to separated by commas (e.g ca195d91c051d91c451483b08bdea0059e39fd34@100.24.11.77:30051,...)


# project constants
export AWS_ACCOUNT_ID=578681496768
export INTERNAL_P2P_PORT=26660
export INTERNAL_RPC_PORT=26670
export INTERNAL_NDAUAPI_PORT=3030
export HONEYCOMB_KEY=b5d540e08c05885849ae13cd7886df04
export SNAPSHOT_BASE_URL="https://s3.amazonaws.com/ndau-snapshots"

# generated constants
export SERVICE_NAME="${APP_BASE_NAME}-service"
export CLUSTER_NAME="${APP_BASE_NAME}-cluster"
export HONEYCOMB_DATASET="${APP_BASE_NAME}-devnet"

# outputs to stderr
errcho() {
  >&2 echo -e "$@"
}

# This outputs a string containing the
task_def(){

  cat << EOF
{
    "family": "${SERVICE_NAME}",
    "networkMode": "awsvpc",
    "executionRoleArn": "arn:aws:iam::${AWS_ACCOUNT_ID}:role/ecsTaskExecutionRole",
    "containerDefinitions": [
        {
            "essential": true,
            "name": "${APP_BASE_NAME}",
            "image": "${AWS_ACCOUNT_ID}.dkr.ecr.us-east-1.amazonaws.com/${APP_BASE_NAME}:${SHA}",
            "environment": [
              { "name" : "HONEYCOMB_DATASET",    "value" : "${HONEYCOMB_DATASET}" },
              { "name" : "HONEYCOMB_KEY",        "value" : "${HONEYCOMB_KEY}" },
              { "name" : "NODE_ID",              "value" : "${NODE_NAME}" },
              { "name" : "PERSISTENT_PEERS",     "value" : "${PERSISTENT_PEERS}" },
              { "name" : "BASE64_NODE_IDENTITY", "value" : "${NODE_IDENTITY}"},
              { "name" : "SNAPSHOT_URL",         "value" : "${SNAPSHOT_BASE_URL}/${SNAPSHOT}.tgz" }
            ],
            "portMappings": [
                {
                    "containerPort": ${INTERNAL_P2P_PORT},
                    "hostPort" : ${INTERNAL_P2P_PORT}
                },
                {
                    "containerPort": ${INTERNAL_RPC_PORT},
                    "hostPort" : ${INTERNAL_RPC_PORT}
                },
                {
                    "containerPort": ${INTERNAL_NDAUAPI_PORT},
                    "hostPort" : ${INTERNAL_NDAUAPI_PORT}
                }
            ],
            "logConfiguration": {
                "logDriver": "awslogs",
                "options": {
                    "awslogs-group": "${APP_BASE_NAME}",
                    "awslogs-region": "us-east-1",
                    "awslogs-stream-prefix": "${NODE_NAME}"
                }
            }
        }
    ],
    "requiresCompatibilities": [
        "FARGATE"
    ],
    "cpu": "256",
    "memory": "512"
}
EOF
# TODO come up with some reasonable settings for cpu and memory.
}

# wrtite a temporary task definition file and register it
tmp_file="$(pwd)/tmp-task-def.json"
task_def > "$tmp_file"
echo "Using task definition: $(cat "$tmp_file")"
resp=$(aws ecs register-task-definition --cli-input-json "file://$tmp_file")

echo "AWS response: $resp"

# Get the task definition ARN from the response
td_name=$(jq -r '.taskDefinition.taskDefinitionArn' <<< $resp)
echo "Task definition ARN: $td_name"

# get the security group id
security_group_id=$(aws ec2 describe-security-groups | jq -r ".SecurityGroups[] | select(.GroupName==\"${APP_BASE_NAME}-sg\") .GroupId")
echo "Fetched security group ID: $security_group_id"

# get the subnet id
subnet_id=$(aws ec2 describe-subnets | \
  jq -r " \
    .Subnets[] as \$parent | \
    \$parent.Tags[]?.Value | \
    select (.==\"sc-node-public\") | \
    \$parent | \
    .SubnetId"
)
echo "Fetched subnet ID: $subnet_id"

# update the service
echo "Updating the service..."
aws ecs update-service --cluster "$CLUSTER_NAME" --service "$SERVICE_NAME" --task-definition "$td_name" --network-configuration "awsvpcConfiguration={assignPublicIp=ENABLED,subnets=[$subnet_id],securityGroups=[$security_group_id]}"

echo "All done"
