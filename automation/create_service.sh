aws ecs create-service --region $1 --cluster $2 --service-name $2 --task-definition $2 --load-balancers loadBalancerName=$2,containerName=$2,containerPort=3030 --desired-count 1 --launch-type EC2 --deployment-configuration maximumPercent=100,minimumHealthyPercent=0 --health-check-grace-period-seconds $3 --scheduling-strategy REPLICA --deployment-controller type=ECS