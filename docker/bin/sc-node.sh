#!/bin/bash
# deploy single devnet node

# get the directory of this file
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"


# args
node_number=$1

# consts
TMP_FILE="$DIR/sc-node-docker-compose.yml"
TEMPLATE_FILE="$DIR/sc-node-template.yml"
IDENTITY_FILE="sc-node-${node_number}.tgz"

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

cat "$TEMPLATE_FILE" | \
  sed \
    -e "s/{{NODE_NUMBER}}/${node_number}/g" \
    -e "s%{{BASE64_NODE_IDENTITY}}%$(cat "$IDENTITY_FILE" | base64)%g" \
  > "$TMP_FILE"
cat "$TMP_FILE"

# Send it to AWS
ecs-cli compose --verbose --project-name sc-node-${node_number} -f ${TMP_FILE} service up --create-log-groups --cluster-config sc-node

# clean up
rm "$TMP_FILE"


#ecs-cli compose --verbose  -f nginx-compose.yml service up --target-group-arn arn:aws:elasticloadbalancing:us-east-1:aws_account_id:targetgroup/ecs-cli-alb/9856106fcc5d4be8 --container-name nginx --container-port 80 --role ecsSeviceRole
#aws elb create-load-balancer --load-balancer-name "$CLUSTER_NAME" --listeners Protocol="TCP,LoadBalancerPort=30000,InstanceProtocol=TCP,InstancePort=8080" --subnets "$SUBNET_ID" --security-groups "$ECS_SECURITY_GROUP" --scheme internal
