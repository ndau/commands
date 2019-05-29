# Automation

## Overview

This directory contains scripts for managing nodes on the following ndau networks:

* devnet
* testnet
* mainnet

The scripts assume that the network is already set up.  See the `.md` files in the `deploy` directory for for further details and instructions on setting up a node manually.

## Automated Tasks

There are scripts for different kinds of node management task.

### Status

To get the status of nodes on a network, run: `./status.sh`

It will prompt you for which network's status to display.

### Stop

To stop a node (so that it won't automatically get restarted by AWS), run: `./stop.sh`

It will prompt you for which network's node(s) to stop, one or all.  Stopping all nodes is useful for re-deploying a network, for example, with a new genesis snapshot or an updated (non backwards compatible) image.

### Start

To start a node, run: `./start.sh`

It will prompt you for which network's node(s) to start, one or all.

### Upgrade

To upgrade a node, run: `./upgrade.sh`

It will prompt you for which network's node(s) to upgrade.

It will prompt you with choices available for which SHA to use for the new `ndauimage` Docker image.  If you don't see it listed, then it's probably not pushed to ECR.  Do a tagged build (e.g. `git tag your-tag-push`) to push up the desired image from the branch you want.

It will then prompt you for which nodes to upgrade, one or all.  When upgrading all nodes, it does a rolling upgrade, starting with the hightest node number and working backward.  There is a deliberate delay between each node's upgrade so that the daily restart timers of each node don't are somewhat staggered.

## Manual Tasks

If there are things you'd like to change about a node that the above scripts don't support, they can be done manually through the AWS Management Console.

1. Sign on to AWS
1. Choose a region in the upper right corner
    - testnet-0 and mainnet-0 are on `us-east-1` (N. Virginia)
    - testnet-1, testnet-5, mainnet-1 and mainnet-5 are on `us-east-2` (Ohio)
    - devnet (all nodes), testnet-2 and mainnet-2 are on `us-west-1` (N. California)
    - testnet-3 and mainnet-3 are on `us-west-2` (Oregon)
    - testnet-4 and mainnet-4 are on `ap-southeast-1` (Singapore)
1. ECS > Task Definitions
1. Select the task definition to alter (e.g. `testnet-1`)
1. Select the highest task definition (e.g. `testnet-1:30`)
1. Create new revision
1. Configure via JSON
1. Edit the JSON
    - Edit environment variables
    - Select which image to use
    - Make any other desired modifications
    - Save
    - Create
1. ECS > Clusters
1. Select the cluster to update (e.g. `testnet-1`)
1. Select the Service (e.g. `testnet-1`)

At this point, choose whether to update the node vs deleting and recreating it.

### Update the node

1. Update
1. Select the latest Task Definition Revision
1. Skip to review
1. Update Service

After a few seconds (where the old docker container exits and a new one starts), the edits you made to the new "latest" Task Definition will be active for the given node.

### Delete and Recreate the node

This approach shouldn't be needed.  The Update approach is preferred.  But it's an option and likely doesn't result in too much more down time for the node compared to the Update approach.

1. Delete
1. Create
    - Launch type: EC2
    - Select the Task Definition Family and Revision (latest)
    - Set the Service name (make it match the Cluster name)
    - Number of tasks: 1
    - Minimum healthy percent: `0`
    - Maximum percent: `100`
    - Placement Templates: One Task Per Host
    - Next step
    - Uncheck "Enable service discovery integration"
    - Next step
    - Next step (again)
    - Create Service (if it gives an error about the old service still draining, click Back, then Create Service again until it works)
