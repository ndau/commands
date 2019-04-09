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

errcho() { >&2 echo "$@"; }

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

while test $# -gt 0
do
    case "$1" in
        --force) FORCE=true
            ;;
    esac
    shift
done

echo "should force? $FORCE"

base_port=30000
rpc_port=$((base_port + 100 + node_number + port_offset))
rpc_node_name="${network_name}-${node_number}-rpc"

p2p_port=$((base_port + 200 + node_number + port_offset))
p2p_node_name="${network_name}-${node_number}-p2p"

ndauapi_port=$((base_port + 300 + node_number + port_offset))
ndauapi_node_name="${network_name}-${node_number}-ndauapi"

reg_targets() {
  >&2 echo -e "\nreg_targets($@)"
  local name_match=$1
  local tg_arn=$2
  instances=$(./get-instanceids-by-name.sh $name_match )
  echo "Registering: $instances"
  ids=$(for id in $instances; do echo "Id=$id"; done)
  id_string=$(echo $ids | sed 's/\n/ /g')
  aws elbv2 register-targets --target-group-arn $tg_arn --targets $id_string
}

get_tg_arn_by_name() {
  >&2 echo -e "\nget_tg_arn_by_name($@)"
  >&2 echo "getting tg arn by name $1"
  aws elbv2 describe-target-groups | jq -r ".TargetGroups[] | select( .TargetGroupName==\"$1\" ) | .TargetGroupArn"
}

get_load_balancer_arn() {
  >&2 echo -e "\nget_load_balancer_arn($@)"
  >&2 echo "getting load balancer arn: $1"
  resp=$(aws elbv2 describe-load-balancers | jq -r ".LoadBalancers[] | select (.LoadBalancerName==\"$1\") | .LoadBalancerArn")
  >&2 echo "got this back: $resp"
  echo $resp


}

get_listener_arn() {
  >&2 echo -e "\nget_listener_arn($@)"
  lb_name=$1
  tg_name=$2
  >&2 echo "lb_name: $lb_name"
  >&2 echo "tg_name: $tg_name"
  lb_arn=$(get_load_balancer_arn $lb_name)
  >&2 echo "lb_name: $lb_name"
  >&2 echo "lb_arn:  $lb_arn"
  aws elbv2 describe-listeners \
    --load-balancer $lb_arn | \
    jq -r ".Listeners[] as \$parent | \
        .Listeners[].DefaultActions[]?.TargetGroupArn | \
        select(. | match(\"$tg_name\") ) | ." \
    | uniq
}


get_listener_arn_by_port() {
  >&2 echo -e "\nget_listener_arn_by_port($@)"
  lb_name=$1
  port=$2
  lb_arn=$(get_load_balancer_arn $lb_name)
  >&2 echo "lb name: $lb_name"
  >&2 echo "lb arn:  $lb_arn"
  >&2 echo "port:    $port"
  aws elbv2 describe-listeners --load-balancer $lb_arn | jq -r ".Listeners[]? | select ( .Port==$port) | .ListenerArn" | uniq
}

force_check() {
  >&2 echo -e "\nforce_check($@)"
  tg_name=$1
  port=$2
  if [ $FORCE == "true" ]; then
    echo "forcing"
    arn=$(get_listener_arn $http_alb_name $tg_name)
    echo "listener_arn $arn"
    if [ ! -z "$arn" ]; then
        aws elbv2 delete-listener --listener-arn $arn
    fi
    same_port_arn=$(get_listener_arn_by_port $http_alb_name $port)
    echo "same_port_listener_arn $same_port_arn"
    if [ ! -z "$same_port_arn" ]; then
        aws elbv2 delete-listener --listener-arn $same_port_arn
    fi
    arn=$(get_tg_arn_by_name $tg_name)
    echo "tg_arn $arn"
    if [ ! -z "$arn" ]; then
      aws elbv2 delete-target-group --target-group-arn $arn
    fi
  fi
}

force_p2p_check() {
  >&2 echo -e "\nforce_p2p_check($@)"
  lb_name=$1
  port=$2
  if [ $FORCE == "true" ]; then
    echo "forcing"
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
  echo "Created RPC target group for ${network_name}-${node_number}"
  reg_targets sc-node-cluster\$ $tg_arn
  aws elbv2 create-listener \
    --load-balancer-arn $http_alb_arn \
    --protocol HTTP \
    --port $rpc_port \
    --default-actions Type=forward,TargetGroupArn=$tg_arn
fi

# P2P
force_p2p_check $p2p_node_name $p2p_port
resp=$(aws elbv2 create-target-group \
--name $p2p_node_name \
--protocol TCP \
--vpc-id "$(./get-vpcid-by-name.sh $vpc)" \
--port ${p2p_port})
if ! grep "An error occurred" <<< $resp; then
  tg_arn=$(jq -r '.TargetGroups[0].TargetGroupArn' <<< $resp)
  echo "Created P2P target group for ${network_name}-${node_number}"
  reg_targets sc-node-cluster\$ $tg_arn
  aws elb create-load-balancer-listeners --load-balancer-name $p2p_elb_name --listeners "Protocol=TCP,LoadBalancerPort=$p2p_port,InstanceProtocol=TCP,InstancePort=$p2p_port"
fi

# ndauapi
force_check $ndauapi_node_name $ndauapi_port
resp=$(aws elbv2 create-target-group \
--name $ndauapi_node_name \
--protocol HTTP \
--vpc-id "$(./get-vpcid-by-name.sh $vpc)" \
--port ${ndauapi_port})
if ! grep "An error occurred" <<< $resp; then
  tg_arn=$(jq -r '.TargetGroups[0].TargetGroupArn' <<< $resp)
  echo "Created ndauapi target group for ${network_name}-${node_number}"
  reg_targets sc-node-cluster\$ $tg_arn
  aws elbv2 create-listener \
    --load-balancer-arn $http_alb_arn \
    --protocol HTTP \
    --port $ndauapi_port \
    --default-actions Type=forward,TargetGroupArn=$tg_arn
fi
