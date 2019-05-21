# Devnet ECS Deployment

## Overview

Resources are scattered around a lot of different places and subscreens in AWS. `ecs-cli` sets up a cloud formation stack, which is helpful, but other resources, such as load balancers are not included in that stack.

## Setting up the cluster

See [this](aws_node_setup.md) for the manual steps we used for creating mainnet.  Use similar steps to set up devnet.  The main difference is that we'll use a single EC2 instance on us-west-1 for all 5 devnet nodes.  Most things named `mainnet-<N>` will instead be named `devnet`.  We'll still use 5 different ECS Task Definitions, named `devnet-0` through `devnet-4`.

The user that's running `ecs-cli` needs the following permissions:
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

## Port scheme

Moving things to AWS has led to some differences in the way our main services are addressed. One load balancer, reachable at `devnet.ndau.tech` routes all traffic.

This is an example of node 0 on devnet:

 * devnet.ndau.tech:26660 P2P
 * devnet.ndau.tech:26670 RPC
 * devnet.ndau.tech:3030 ndauapi

For nodes 1 through 4, add the node number to the above ports.  e.g. 26663 is node 3's P2P port.

## Node identities

An S3 bucket called `ndau-deploy-secrets` holds tarballs containing private keys specific to each node. For example `s3://ndau-deploy-secrets/node-identities-devnet.tgz` will contain a set of tarballs, one for each node. Each node identity tarball contains two files needed by Tendermint to establish its identity on the network.

## Manual deployment

This should only be done for debugging situations and CircleCI relied on for all normal deployments. `devnet-deploy.sh`. They contain the variables necessary to deploy a node, however, the variables are meant to be changed to match your configuration and situation. Internally both scripts use `deploy-node.sh` to deploy a series of nodes one at a time.

## Scripts

* `deploy-node.sh` deploys a new single node.
* `process-ids.sh` takes a directory with node identities and make a tarball and persistent peers list for the PERSISTENT_PEERS variable in CircleCI which need to be copied and pasted into config.yml.

## Debugging

* `devnet-liveness.sh` tests that ports are available on devnet.
* `devnet-deploy.sh` manually deploys devnet (do not run unless manually configured).
