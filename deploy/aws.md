# ECS Deployment

Resources are scattered around a lot of different places and subscreens in AWS. To help recognize these resources, most of them have the prefix `sc-node` in either the name or description field. `ecs-cli` sets up a cloud formation stack, which is helpful, but other resources, such as load balancers are not included in that stack.

NOTE: Some of this is old and only partially updated to deploy to us-west-1.  See [this](aws_node_setup.md) for an alternate approach to setting things up on AWS.

# Setting up the cluster

First go to the EC2 console and create a keypair https://console.aws.amazon.com/ec2/v2/home?region=us-west-1#KeyPairs:sort=keyName

The keypair used for the single container nodes (sc-node) is `sc-node-ec2`.  If there is one up there ending in `-mainnet`, we'll use that one instead of creating one.

ecs-cli up --keypair sc-node-ec2-mainnet --capability-iam --size 1 --instance-type m5.large --cluster-config sc-node

The security group that ecs-cli sets up should be changed to allow incoming traffic on the following ports:
  - 22 ssh
  - 3030 ndauapi
  - 26660 tendermint p2p
  - 26670 tendermint rpc

Name the VPC `sc-node-ecs`. It'll be easier to look up and if you name it something else.

Create load balancer
    - Classic ELB
        name: sc-node
        health check: just make it TCP port 22 to keep it happy.

Go to Route53 and add records the load balancers
    - `devnet.ndau.tech` for `sc-node`.

The user that's running ecs-cli needs the following permissions:
    "ecs:DescribeServices",
    "ecs:DescribeClusters",
    "ecs:ListTasks",
    "ecs:DescribeTasks",
    "ecs:DescribeTaskDefinition",
    "ecs:DescribeContainerInstances",
    "ec2:Describe*",
    "ecs:RegisterTaskDefinition",
    "logs:CreateLogGroup",
    "ecs:UpdateService",
    "ecs:ListAccountSettings",
    "ecs:CreateService"

# Port scheme

Moving things to AWS has led to some differences in the way our main services are addressed. Two load balancers, reachable at `api.ndau.tech` and `p2p.ndau.tech` serve different kinds of traffic. `api` is for HTTP, and `p2p` is for TCP. Therefore, ndauapi and Tendermint's RPC port are accessed through `api` and Tendermint's P2P is reachable from `p2p`. This would be simple enough, except that port numbers also change for each service, ie. subdomains (devnet-0.api.ndau.tech) are no longer used due to limitations in AWS routing vs Kubernetes routing.

This is an example of node 0 on devnet:

 * devnet.ndau.tech:26660 P2P
 * devnet.ndau.tech:26670 RPC
 * devnet.ndau.tech:3030 ndauapi

For nodes 1 through 4, add the node number to the above ports.  e.g. 26663 is node 3's P2P port.

# Other resources

## Node identities

An S3 bucket called `ndau-deploy-secrets` holds tarballs containing private keys specific to each node. For example `s3://ndau-deploy-secrets/node-identities-devnet.tgz` will contain a set of tarballs, one for each node. Each node identity tarball contains two files needed by Tendermint to establish its identity on the network.

## Load balancers

This deployment uses a combination of 1 ALB for http traffic and 1 classic ELB for p2p traffic: `sc-node-http` (accessible through `api.ndau.tech`) and `sc-node-p2p` (accessible through `p2p.ndau.tech`) respectively.

## Manual deployment

This should only be done for debugging situations and CircleCI relied on for all normal deployments. `devnet-deploy.sh`. They contain the variables necessary to deploy a node, however, the variables are meant to be changed to match your configuration and situation. Internally both scripts use `deploy-node.sh` to deploy a series of nodes one at a time.

## Scripts

* `deploy-node.sh` deploys a new single node.
* `process-ids.sh` takes a directory with node identities and make a tarball and persistent peers list for the PERSISTENT_PEERS variable in CircleCI which need to be copied and pasted into config.yml.

### Debugging

* `devnet-liveness.sh` tests that ports are available on devnet.
* `devnet-deploy.sh` manually deploys devnet (do not run unless manually configured).

