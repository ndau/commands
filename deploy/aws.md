# ECS Deployment

Resources are scattered around a lot of different places and subscreens in AWS. To help recognize these resources, most of them have the prefix `sc-node` in either the name or description field. `ecs-cli` sets up a cloud formation stack, which is helpful, but other resources, such as load balancers are not included in that stack.

# Setting up the cluster

First go to the Ec2 console and create a keypair https://console.aws.amazon.com/ec2/v2/home?region=us-east-1#KeyPairs:sort=keyName

The keypair used for the single container nodes (sc-node) is `sc-node-ec2`.

ecs-cli up --keypair sc-node-ec2 --capability-iam --size 2 --instance-type m5a.xlarge --cluster-config sc-node --azs=us-east-1a,us-east-1b

> note: Be sure to get `ecs-cli` version 1.14.1 or later, for `m5a.xlarge` instance type support.

The security group that ecs-cli sets up should be changed to allow incoming traffic on the following ports:
  - 22 ssh
  - 3030 ndauapi
  - 26660 tendermint p2p
  - 26670 tendermint rpc
  - 30000-30400 load balancers

Name the VPC `sc-node-ecs`. It'll be easier to look up and if you name it something else, you'll have to set the `VPC` environment variable when you run `target-groups.sh`.

Create 2 load balancers
    - ALB named sc-node-http. Health checks are installed with `target-groups.sh`.
    - ELB
        name: sc-node-p2p
        health check: just make it TCP port 22 to keep it happy.

Run `target-groups-init.sh devnet`.

Go to Route53 and add records the load balancers
    - `api.ndau.tech` for `sc-node-http` and
    - `p2p.ndau.tech` for `sc-node-p2p`.

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

This is an example of node-0 on devnet:

 * api.ndau.tech:30100 RPC
 * p2p.ndau.tech:30200 P2P
 * api.ndau.tech:30300 ndauapi

Or basically the following: `3XYZZ`
  * where 3 is by convention,
  * X is the network (0=devnet),
  * Y is the service (1=rpc, 2=p2p, 3=ndauapi),
  * and ZZ is the node number.

## Target groups

Setting up target groups, listeners, connecting them to instances is tedious and error prone. Once the target groups have been established however, they likely will not need to be touched for some time. `target-groups.sh` helps by setting up a single node's target groups, attaching them to the ec2 instances, and setting up listeners on the correct ports. `target-groups-init.sh` will run `target-groups.sh` multiple times to set up either devnet (e.g. `./target-groups-init.sh devnet`).

# Other resources

## Node identities

An S3 bucket called `ndau-deploy-secrets` holds tarballs containing private keys specific to each node. For example `s3://ndau-deploy-secrets/node-identities-devnet.tgz` will contain a set of tarballs, one for each node. Each node identity tarball contains two files needed by Tendermint to establish its identity on the network.

## Load balancers

This deployment uses a combination of 1 ALB for http traffic and 1 classic ELB for p2p traffic: `sc-node-http` (accessible through `api.ndau.tech`) and `sc-node-p2p` (accessible through `p2p.ndau.tech`) respectively.

## Elastic IPs

There are two elastic IPs that are connected to each instance in the sc-node-cluster. This allows Tendermint's p2p to function with static ips.

## Manual deployment

This should only be done for debugging situations and CircleCI relied on for all normal deployments. `devnet-deploy.sh`. They contain the variables necessary to deploy a node, however, the variables are meant to be changed to match your configuration and situation. Internally both scripts use `deploy-node.sh` to deploy a series of nodes one at a time.

## Scripts

* `deploy-node.sh` deploys a new single node.
* `process-ids.sh` takes a directory with node identities and make a tarball and persistent peers list for the PERSISTENT_PEERS variable in CircleCI which need to be copied and pasted into config.yml.
* `target-groups.sh` creates target groups and listeners for a node's rpc, p2p and ndauapi ports.
* `target-groups-init.sh` creates target groups for a network.

### Debugging

* `devnet-liveness.sh` tests that ports are available on devnet.
* `devnet-deploy.sh` manually deploys devnet (do not run unless manually configured).

