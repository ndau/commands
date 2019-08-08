# Automation

## Overview

This directory contains scripts for managing nodes on the following ndau networks:

* devnet
* testnet
* mainnet

The scripts assume that the network is already set up.  See the `.md` files in the `deploy` directory for further details and instructions on setting up nodes on a network manually.

## Node Status

The `get_*.py` scripts poll status info from nodes on a network.

It's up to the caller to know the API and RPC URLs to use.  The [services.json](https://s3.us-east-2.amazonaws.com/ndau-json/services.json) file contains the URLs available.  We don't fetch it ourselves internally since these tools are meant to be executed in loops over nodes in a network and over multiple networks.  The caller can get `services.json` once and then run those loops.

To get the health of node 3 on testnet:
```sh
./get_health.py https://testnet-3.ndau.tech:3030
```

To get the SHA of node 3 on testnet:
```sh
./get_sha.py https://testnet-3.ndau.tech:3030
```

To get the height of node 3 on testnet:
```sh
./get_height.py https://testnet-3.ndau.tech:3030
```

To get the number of peers of node 3 on testnet:
```sh
./get_peers.py https://testnet-3.ndau.tech:26670
```

To get the catch up status of node 3 on testnet:
```sh
./get_catchup.py https://testnet-3.ndau.tech:26670
```

To get the version tag of node 3 on testnet:
```sh
./get_version.py https://testnet-3.ndau.tech:26670
```

## HUD

Ultimately we'll want a GUI published somewhere for point-and-click node management.  For now, we have a text-based HUD.  This script demonstrates fetching `services.json` once and looping over its entries to display status information about every Oneiro node on every network.

To see an auto-updating HUD:
```sh
./hud.py
```

To see only testnet info in the HUD:
```sh
./hud.py testnet
```

It can be useful when doing a rolling upgrade (below) to keep an eye on every node's status on that network as each node is restarted.

## Node Control

The following scripts make modifications to nodes on a network.

### Upgrade

This should be the most common thing we need to do when controlling nodes on a network.  As long as we have backward-compatible changes, we can do a rolling upgrade of a network's nodes.

First, find the SHA you want to upgrade to on [ECR](https://console.aws.amazon.com/ecr/repositories/sc-node/?region=us-east-1).  The image revisions that show up there come from devnet master deploys and tagged builds from a branch (e.g. `git tag your-tag-push`).  Only SHAs that are listed here are allowed to be used when upgrading.

To upgrade all nodes on testnet to the `badf00d` SHA:
```sh
./upgrade.py testnet --sha badf00d
```

This does a rolling upgrade, starting with the highest node number and working backward.  There is a deliberate delay between each node's upgrade so that the daily restart timers of each node are somewhat staggered.

To upgrade node 3 on testnet to the `badf00d` SHA:
```sh
./upgrade.py testnet --sha badf00d --node testnet-3
```

Single node upgrades are useful if you would like more control over the timing and order of node upgrades on a network.  It's also useful if a rolling upgrade was interrupted for any reason.

### Snapshot

We can cause a node to take a snapshot of its data files then upload it to S3 and register it as the latest snapshot for its network.

To cause node 5 on testnet to take a snapshot:
```sh
./snapshot_node.py testnet-5
```

### Configure

This is not yet automated.

Currently, upgrading (above) only supports changing which image SHA to use.  We may want to update some of the environment variables in the Task Definitions as well, regardless of whether we're changing the SHA within it.

We could fold this into the current upgrade script, or make a new one for changing non-SHA config (environment variables).

For now, we can use the Manual Steps (below) for editing a node's Task Definition and Updating its service.

### Schema Change

This is not yet automated.

We currently have a schema change transaction, but it has never been tested on one of our networks.  We'll add support for this once we have a need for it.  We'll need to stop all validator nodes and restart them after an upgrade.

Our current rolling upgrade process (above) is not suitable for a schema change, since the node software is presumably not backward compatible and we therefore cannot upgrade one node without upgrading all of them simultaneously.

### Genesis

This is not yet automated.

Currently, devnet does a re-genesis on every deploy, so that's taken care of.  Mainnet will never start over from genesis, so that's not something we need to support.

That just leaves testnet.  Since testnet is a staging network for mainnet, it should be a rare occurrence to have to start it over with a new genesis snapshot.

For now, we continue to do that manually (taking a snapshot, uploading it to S3, getting the node identities and persistent peers and editing the Task Definitions on ECS, then Updating the testnet services via the AWS Management Console).

## Manual Steps

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

### Option 1: Update the node

1. Update
1. Select the latest Task Definition Revision
1. Skip to review
1. Update Service

After a few seconds (where the old docker container exits and a new one starts), the edits you made to the new "latest" Task Definition will be active for the given node.

### Option 2: Delete and Recreate the node

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
