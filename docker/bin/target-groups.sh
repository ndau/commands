#!/bin/bash

# makes target groups for different ips and ports

vpc=${VPC:-sc-node-ecs}

network_name=$1
node_number=$2
port_offset=$PORT_OFFSET
[ -z "$port_offset" ] && port_offset=0

p2p_elb_name="sc-node-p2p"
p2p_elb_arn=$(printf "arn:aws:elasticloadbalancing:%s:%s:loadbalancer/%s" us-east-1 578681496768 $p2p_elb_name)

http_alb_name="sc-node-http"
http_alb_arn=$(aws elbv2 describe-load-balancers | jq -r ".LoadBalancers[] | select (.LoadBalancerName==\"$http_alb_name\") | .LoadBalancerArn")

errcho() { >&2 echo -e "$@"; }
verrcho() { [ "$VERBOSE" == "true" ] && errcho "$@"; }

usage() {
  errcho "Usage: PORT_OFFSET=0 $0 network_name node_number"
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

if [ ! "$port_offset" -eq "$port_offset" ]; then
  errcho "port_offset not a number"
  usage
  exit 1
fi

FORCE=false
VERBOSE=false

while test $# -gt 0
do
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

base_port=30000
rpc_port=$((base_port + 100 + node_number + port_offset))
rpc_node_name="${network_name}-${node_number}-rpc"

p2p_port=$((base_port + 200 + node_number + port_offset))
p2p_node_name="${network_name}-${node_number}-p2p"

ndauapi_port=$((base_port + 300 + node_number + port_offset))
ndauapi_node_name="${network_name}-${node_number}-ndauapi"

reg_targets() {
  verrcho "\nreg_targets($@)"
  local name_match=$1
  local tg_arn=$2
  instances=$(./get-instanceids-by-name.sh $name_match )
  errcho "Registering: $instances"
  ids=$(for id in $instances; do echo "Id=$id"; done)
  id_string=$(sed 's/\n/ /g' <<< $ids)
  aws elbv2 register-targets --target-group-arn "$tg_arn" --targets "$id_string"
}

get_tg_arn_by_name() {
  verrcho "\nget_tg_arn_by_name($@)"
  verrcho "getting tg arn by name $1"
  aws elbv2 describe-target-groups | jq -r ".TargetGroups[] | select( .TargetGroupName==\"$1\" ) | .TargetGroupArn"
}

get_load_balancer_arn_by_name() {
  verrcho "\nget_load_balancer_arn_by_name($@)"
  verrcho "getting load balancer arn: $1"
  resp=$(aws elbv2 describe-load-balancers | jq -r ".LoadBalancers[] | select (.LoadBalancerName==\"$1\") | .LoadBalancerArn")
  verrcho "got this back: $resp"
  echo $resp
}

get_listener_arn() {
  verrcho "\nget_listener_arn($@)"
  lb_name=$1
  tg_name=$2
  verrcho "lb_name: $lb_name"
  verrcho "tg_name: $tg_name"
  lb_arn=$(get_load_balancer_arn_by_name $lb_name)
  verrcho "lb_arn:  $lb_arn"
  aws elbv2 describe-listeners \
    --load-balancer $lb_arn | \
    jq -r ".Listeners[] as \$parent | \
        .Listeners[].DefaultActions[]?.TargetGroupArn | \
        select(. | match(\"$tg_name\") ) | ." \
    | uniq
}


get_listener_arn_by_port() {
  verrcho "\nget_listener_arn_by_port($@)"
  lb_name=$1
  port=$2
  lb_arn=$(get_load_balancer_arn_by_name $lb_name)
  verrcho "lb_name: $lb_name"
  verrcho "lb_arn:  $lb_arn"
  verrcho "port:    $port"
  aws elbv2 describe-listeners --load-balancer $lb_arn | jq -r ".Listeners[]? | select ( .Port==$port) | .ListenerArn" | uniq
}

# If force is set checks for a listener with a certain target group or port and deletes the listener and the target group.
force_check() {
  verrcho "\nforce_check($@)"
  tg_name=$1
  port=$2
  if [ "$FORCE" == "true" ]; then
    verrcho "redundantly attempting to delete previous listeners and target groups"
    listener_arn=$(get_listener_arn $http_alb_name $tg_name)
    verrcho "listener_arn $listener_arn"
    [ ! -z "$listener_arn" ] && aws elbv2 delete-listener --listener-arn $listener_arn

    same_port_arn=$(get_listener_arn_by_port $http_alb_name $port)
    verrcho "same_port_listener_arn $same_port_arn"
    [ ! -z "$same_port_arn" ] && aws elbv2 delete-listener --listener-arn $same_port_arn

    tg_arn=$(get_tg_arn_by_name $tg_name)
    verrcho "tg_arn $tg_arn"
    [ ! -z "$tg_arn" ] && aws elbv2 delete-target-group --target-group-arn $tg_arn
  fi
}

force_p2p_check() {
  verrcho "\nforce_p2p_check($@)"
  lb_name=$1
  port=$2
  if [ "$FORCE" == "true" ]; then
    errcho "Attempting to delete p2p listener"
    aws elb delete-load-balancer-listeners --load-balancer-name $lb_name --load-balancer-ports $port
  fi
}

# RPC
force_check $rpc_node_name $rpc_port
resp=$(aws elbv2 create-target-group \
--name $rpc_node_name \
--protocol HTTP \
--vpc-id "$(./get-vpcid-by-name.sh $vpc)" \
--port ${rpc_port})
if ! grep "An error occurred" <<< $resp; then
  tg_arn=$(jq -r '.TargetGroups[0].TargetGroupArn' <<< $resp)
  errcho "Created RPC target group for ${network_name}-${node_number}"
  reg_targets sc-node-cluster\$ $tg_arn
  aws elbv2 create-listener \
    --load-balancer-arn $http_alb_arn \
    --certificates "CertificateArn=arn:aws:acm:us-east-1:578681496768:certificate/2e669f22-adf7-44eb-ac7b-fa45a85503d7" \
    --protocol HTTPS \
    --port $rpc_port \
    --default-actions Type=forward,TargetGroupArn=$tg_arn
fi

# P2P
force_p2p_check $p2p_elb_name $p2p_port
resp=$(aws elbv2 create-target-group \
--name $p2p_node_name \
--protocol TCP \
--vpc-id "$(./get-vpcid-by-name.sh $vpc)" \
--port ${p2p_port})
if ! grep "An error occurred" <<< $resp; then
  tg_arn=$(jq -r '.TargetGroups[0].TargetGroupArn' <<< $resp)
  errcho "Created P2P target group for ${network_name}-${node_number}"
  reg_targets sc-node-cluster\$ $tg_arn
  aws elb create-load-balancer-listeners --load-balancer-name $p2p_elb_name --listeners "Protocol=TCP,LoadBalancerPort=$p2p_port,InstanceProtocol=TCP,InstancePort=$p2p_port"
fi

# ndauapi
force_check $ndauapi_node_name $ndauapi_port
resp=$(aws elbv2 create-target-group \
--name $ndauapi_node_name \
--protocol HTTP \
--vpc-id "$(./get-vpcid-by-name.sh $vpc)" \
--health-check-path "/node/status" \
--port ${ndauapi_port})
if ! grep "An error occurred" <<< $resp; then
  tg_arn=$(jq -r '.TargetGroups[0].TargetGroupArn' <<< $resp)
  errcho "Created ndauapi target group for ${network_name}-${node_number}"
  reg_targets sc-node-cluster\$ $tg_arn
  aws elbv2 create-listener \
    --load-balancer-arn $http_alb_arn \
    --certificates "CertificateArn=arn:aws:acm:us-east-1:578681496768:certificate/2e669f22-adf7-44eb-ac7b-fa45a85503d7" \
    --protocol HTTPS \
    --port $ndauapi_port \
    --default-actions Type=forward,TargetGroupArn=$tg_arn
fi
