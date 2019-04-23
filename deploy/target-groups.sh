#!/bin/bash

# makes target groups for different ips and ports

# get the directory of this file
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "$DIR/deploy-lib.sh"

CERT_ARN="CertificateArn=arn:aws:acm:us-east-1:578681496768:certificate/2e669f22-adf7-44eb-ac7b-fa45a85503d7"
DEFAULT_VPC="sc-node-ecs"
BASE_PORT=30000

vpc=${VPC:-$DEFAULT_VPC}

network_name=$1
node_number=$2

p2p_elb_name="sc-node-p2p"
p2p_elb_arn=$(printf "arn:aws:elasticloadbalancing:%s:%s:loadbalancer/%s" us-east-1 578681496768 $p2p_elb_name)
err_if_empty "$p2p_elb_arn" "Could not get p2p ARN for $p2p_elb_name"

http_alb_name="sc-node-http"
http_alb_arn=$(aws elbv2 describe-load-balancers | jq -r ".LoadBalancers[] | select (.LoadBalancerName==\"$http_alb_name\") | .LoadBalancerArn")
err_if_empty "$http_alb_arn" "Could not get http load balancer ARN for $http_alb_name"

usage() {
  errcho "Usage: $0 network_name node_number"
}

if [ -z "$network_name" ]; then
  errcho "Missing network_name"
  usage
  exit 1
fi

if [ -z "$node_number" ]; then
  errcho "Missing node_number"
  usage
  exit 1
fi

FORCE=false
VERBOSE=false

while test $# -gt 0; do
    case "$1" in
        --force) FORCE=true
            ;;
        --verbose) VERBOSE=true
            ;;
    esac
    shift
done

verrcho "FORCE=$FORCE"
verrcho "VERBOSE=$VERBOSE"

rpc_port=$(calc_port $network_name rpc $node_number)
rpc_node_name="${network_name}-${node_number}-rpc"

p2p_port=$(calc_port $network_name p2p $node_number)
p2p_node_name="${network_name}-${node_number}-p2p"

ndauapi_port=$(calc_port $network_name ndauapi $node_number)
ndauapi_node_name="${network_name}-${node_number}-ndauapi"

verrcho "ports"
verrcho $rpc_node_name $rpc_port
verrcho $p2p_node_name $p2p_port
verrcho $ndauapi_node_name $ndauapi_port



# returns instance ids by name
get_instanceids_by_name() {
  name=$1 #"sc-node-cluster$"
  aws ec2 describe-instances | jq -r " \
    .Reservations[].Instances[] as \$parent | \
    \$parent.Tags[]?.Value | \
    select (. | match(\"${name}\") ) | \
    \$parent | \
    .InstanceId" | uniq # Don't know why uniq is required, but doubles appear.
}

# gets vpc by name
get_vpcid_by_name() {
  name=$1

  aws ec2 describe-vpcs |
  jq -r " \
    .Vpcs[] as \$parent | \
    \$parent.Tags[]?.Value | \
    select (.==\"${name}\") | \
    \$parent | \
    .VpcId"
}

reg_targets() {
  verrcho "\nreg_targets($@)"
  local name_match=$1
  local tg_arn=$2
  instances=$(get_instanceids_by_name $name_match )
  err_if_empty "$instances" "Could not fetch instances with the name $name_match"
  errcho "Registering: $instances"
  ids=$(for id in $instances; do echo "Id=$id"; done)
  id_string=$(sed 's/\n/ /g' <<< $ids)
  aws elbv2 register-targets --target-group-arn "$tg_arn" --targets $id_string # leave id_string unquoted
}

get_tg_arn_by_name() {
  verrcho "\nget_tg_arn_by_name($@)"
  aws elbv2 describe-target-groups | jq -r ".TargetGroups[] | select( .TargetGroupName==\"$1\" ) | .TargetGroupArn"
}

get_load_balancer_arn_by_name() {
  verrcho "\nget_load_balancer_arn_by_name($@)"
  aws elbv2 describe-load-balancers | jq -r ".LoadBalancers[] | select (.LoadBalancerName==\"$1\") | .LoadBalancerArn"
}

get_listener_arn() {
  verrcho "\nget_listener_arn($@)"
  lb_name=$1
  tg_name=$2
  verrcho "lb_name: $lb_name"
  verrcho "tg_name: $tg_name"
  lb_arn=$(get_load_balancer_arn_by_name $lb_name)
  err_if_empty "$lb_arn" "Could not get load balancer ARN with name: $lb_name"
  verrcho "lb_arn:  $lb_arn"
  aws elbv2 describe-listeners \
    --load-balancer $lb_arn | \
    jq -r ".Listeners[] as \$parent | \
        \$parent.DefaultActions[]?.TargetGroupArn | \
        select(. | match(\"$tg_name\") ) | \
        \$parent.ListenerArn" \
    | uniq
}

get_listener_arn_by_port() {
  verrcho "\nget_listener_arn_by_port($@)"
  lb_name=$1
  port=$2
  lb_arn=$(get_load_balancer_arn_by_name $lb_name)
  err_if_empty "$lb_arn" "Could not get load balancer ARN with name: $lb_name"
  verrcho "lb_name: $lb_name"
  verrcho "lb_arn:  $lb_arn"
  verrcho "port:    $port"
  aws elbv2 describe-listeners --load-balancer $lb_arn | jq -r ".Listeners[]? | select ( .Port==$port) | .ListenerArn" | uniq
}

# If force is set checks for a listener with a certain target group or port and deletes the listener and the target group.
delete_if_forced() {
  verrcho "\ndelete_if_forced($@)"
  tg_name=$1
  port=$2
  if [ "$FORCE" == "true" ]; then
    verrcho "redundantly attempting to delete previous listeners and target groups"
    listener_arn=$(get_listener_arn $http_alb_name $tg_name)
    check_empty "$listener_arn" "Could not get listener ARN with: $http_alb_name $tg_name"
    verrcho "listener_arn $listener_arn"
    [ ! -z "$listener_arn" ] && aws elbv2 delete-listener --listener-arn $listener_arn --region us-east-1

    same_port_arn=$(get_listener_arn_by_port $http_alb_name $port)
    check_empty "$same_port_arn" "Could not get listener ARN with: $http_alb_name $port"
    verrcho "same_port_listener_arn $same_port_arn"
    [ ! -z "$same_port_arn" ] && aws elbv2 delete-listener --listener-arn $same_port_arn

    tg_arn=$(get_tg_arn_by_name $tg_name)
    check_empty "$tg_arn" "Could not get listener ARN with: $tg_name"
    verrcho "tg_arn $tg_arn"
    [ ! -z "$tg_arn" ] && aws elbv2 delete-target-group --target-group-arn $tg_arn
  fi
}

delete_p2p_if_forced() {
  verrcho "\ndelete_p2p_if_forced($@)"
  lb_name=$1
  port=$2
  tg_name=$3
  if [ "$FORCE" == "true" ]; then
    errcho "Attempting to delete p2p listener"
    aws elb delete-load-balancer-listeners --load-balancer-name $lb_name --load-balancer-ports $port

    tg_arn=$(get_tg_arn_by_name $tg_name)
    check_empty "$tg_arn" "Could not get target group ARN with: $tg_name"

    verrcho "tg_arn $tg_arn"
    [ ! -z "$tg_arn" ] && aws elbv2 delete-target-group --target-group-arn $tg_arn
  fi
}

# RPC
errcho "${network_name}-${node_number}: rpc"
delete_if_forced $rpc_node_name $rpc_port
resp=$(aws elbv2 create-target-group \
--name $rpc_node_name \
--protocol HTTP \
--vpc-id "$(get_vpcid_by_name $vpc)" \
--port ${rpc_port})
if ! grep "An error occurred" <<< $resp; then
  tg_arn=$(jq -r '.TargetGroups[0].TargetGroupArn' <<< $resp)
  err_if_empty "$tg_arn" "Could not get target group ARN with name $rpc_node_name on port $rpc_port"
  errcho "Created RPC target group for ${network_name}-${node_number}"
  reg_targets sc-node-cluster\$ $tg_arn
  aws elbv2 create-listener \
    --load-balancer-arn $http_alb_arn \
    --certificates "$CERT_ARN" \
    --protocol HTTPS \
    --port $rpc_port \
    --default-actions Type=forward,TargetGroupArn=$tg_arn
fi

# P2P
errcho "${network_name}-${node_number}: p2p"
delete_p2p_if_forced $p2p_elb_name $p2p_port $p2p_node_name
resp=$(aws elbv2 create-target-group \
--name $p2p_node_name \
--protocol TCP \
--vpc-id "$(get_vpcid_by_name $vpc)" \
--port ${p2p_port})
if ! grep "An error occurred" <<< $resp; then
  tg_arn=$(jq -r '.TargetGroups[0].TargetGroupArn' <<< $resp)
  err_if_empty "$tg_arn" "Could not get target group ARN with name $p2p_node_name on port $p2p_port"
  errcho "Created P2P target group for ${network_name}-${node_number}"
  reg_targets sc-node-cluster\$ $tg_arn
  aws elb create-load-balancer-listeners --load-balancer-name $p2p_elb_name --listeners "Protocol=TCP,LoadBalancerPort=$p2p_port,InstanceProtocol=TCP,InstancePort=$p2p_port"
fi

# ndauapi
errcho "${network_name}-${node_number}: ndauapi"
delete_if_forced $ndauapi_node_name $ndauapi_port
resp=$(aws elbv2 create-target-group \
--name $ndauapi_node_name \
--protocol HTTP \
--vpc-id "$(get_vpcid_by_name $vpc)" \
--health-check-path "/node/health" \
--port ${ndauapi_port})
if ! grep "An error occurred" <<< $resp; then
  tg_arn=$(jq -r '.TargetGroups[0].TargetGroupArn' <<< $resp)
  err_if_empty "$tg_arn" "Could not get target group ARN with: name:$ndauapi_port on port $ndauapi_port"

  errcho "Created ndauapi target group for ${network_name}-${node_number}"
  reg_targets sc-node-cluster\$ $tg_arn
  aws elbv2 create-listener \
    --load-balancer-arn $http_alb_arn \
    --certificates "$CERT_ARN" \
    --protocol HTTPS \
    --port $ndauapi_port \
    --default-actions Type=forward,TargetGroupArn=$tg_arn
fi
