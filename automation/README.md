# Automation

## Overview

This directory contains scripts for managing nodes on the following ndau networks:

* devnet
* testnet
* mainnet

The scripts assume that the network is already set up.  See the `.md` files in the `deploy` directory for further details and instructions on setting up nodes on a network manually.

## Install

You will need to have the `requests` module installed into your Python 3 environment in order to run the scripts here.

In order to use the "rolling upgrade with reindex" feature, you must have the `sc-node-ec2.pem` file installed in your `~/.ssh` directory and `chmod 400` it.  Get it from the Oneiro 1password account.

You can optionally set and export the `SLACK_DEPLOYS_KEY` environment variable before running the upgrde script.  Get the key from someone on the team.  It'll allow the scripts to post notifications to the ndev `#deploys` slack channel when upgrades complete.

If we ever implement the upgrade script as a Circle job, we won't need to do any of the above locally.  Our `commands/.circleci/config.yml` and environment vairables up on the Circle server are already configured to have what we need.

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

First, find the SHA you want to upgrade to on [ECR](https://console.aws.amazon.com/ecr/repositories/ndauimage/?region=us-east-1).  The image revisions that show up there come from devnet master deploys and tagged builds from a branch (e.g. `git tag your-tag-push`).  Only SHAs that are listed here are allowed to be used when upgrading.

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

### Upgrade with full reindex

Sometimes we need to do a rolling upgrade, but force the nodes to wipe and reindex Redis data.  This is required if we change the format of an index, and is useful when we add new indexes that we want to have filled with everything that's on the blockchain already.

To do this, we can choose a snapshot to start with.  This is more general than it needs to be for this use case, because here we'd always want to start with the genesis snapshot `snapshot-mainnet-1`:

```sh
./upgrade.py testnet --sha badf00d --snapshot snapshot-mainnet-1
```

This will cause testnet to upgrade to the given SHA, but the `testnet-backup` node will be the only one that reindexes from genesis.  (The testnet network is a fork of mainnet, so we use mainnet's genesis snapshot for both).  The upgrade-with-reindex process will then take a new snapshot and upload it to S3.  Then the rest of the network's nodes will do a rolling upgrade as usual, this time with the new "latest" snapshot with new Redis data.

NOTE: There is a private bucket at `s3://ndau-snapshots/old/<NETWORK>-vX.Y.Z` you can move old snapshots to.  You should move away all old snapshots on the network you're upgrading-with-reindex because after the new snapshot is taken, all old ones become obsolete (since they have old Redis data in them).  The only one we keep forever is `snapshot-mainnet-1.tgz`.  That, too, has obsolete Redis data in it, but generally on an upgrade-with-reindex, we bump the `indexVersion` in Go code, to force a reindex of the sysvars present from noms in that snapshot.  So the obsolete Redis data gets regenerated automatically, and relatively quickly, when the backup node first starts up using that snapshot.

### Snapshot

We can cause a node to take a snapshot of its data files then upload it to S3 and register it as the latest snapshot for its network.

To cause the backup (snapshot-taking) node on testnet to take a snapshot:
```sh
./snapshot_node.py testnet-backup
```

### Configure

This is not yet automated.

Currently, upgrading (above) only supports changing which image SHA to use.  We may want to update some of the environment variables in the Task Definitions as well, regardless of whether we're changing the SHA within it.

We could fold this into the current upgrade script, or make a new one for changing non-SHA config (environment variables).

For now, we can use the Manual Steps (below) for editing a node's Task Definition and Updating its service.

### Genesis

Sometimes we want to reset devnet.  Currently when we land to `commands` master, we redeploy without resetting blockchain data.  We can use the `*-jobs_reset` tag to force a resetting deploy, though.  There is no way to reset testnet or mainnet.  Of course, we never want to reset mainnet.  But sometimes we want to "repave" testnet with the latest blockchain data from mainnet.  That's covered below.

## Manual Steps

If there are things you'd like to change about a node that the above scripts don't support, they can be done manually through the AWS Management Console.

1. Sign on to AWS
1. Choose a region in the upper right corner
    - testnet-0 and mainnet-0 are on `us-east-1` (N. Virginia)
    - testnet-1, testnet-backup, mainnet-1 and mainnet-backup are on `us-east-2` (Ohio)
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

## Repaving testnet

Sometimes we want to repave testnet with the latest blockchain snapshot from mainnet.  Any time we do this, testnet becomes a fork of mainnet.  When we add big features like new or changed transactions, we need to practice using them on testnet before we upgrade mainnet and submit them there.

Repaving testnet is currently a manual process.  Conceptually, it's simple:

1. Take down all testnet nodes
1. Copy `latest-mainnet.txt` to `latest-testnet.txt` on S3
1. Fire up all testnet nodes

We can't do a "rolling repave" because blockchain data differs between testnet and mainnet, and so catchup will fail for a restarted node while the others are still running on old testnet data.  This is why we do a full take-down and fire-up.

Here are the steps for accomplishing this:

1. Optional take a manual snapshot of testnet
    - This is if you want a backup of old testnet before repaving it
    - run: `./snapshot_node.py testnet-backup`
    - It'll get uploaded to `s3://ndau-snapshots` for safe keeping
1. Optional: take a manual snapshot of mainnet
    - This is if you want to ensure you're using the absolute latest state of mainnet to repave testnet with; mainnet takes snapshots every hour, so this might not be necessary
    - run: `./snapshot_node.py mainnet-backup`
    - It'll upload to S3 and update `mainnet-latest.txt` to point to it
1. In the AWS Management Console, navigate to ECS > Clusters
    - "Delete" services `testnet-0` through `testnet-4` and `testnet-backup`
    - Instructions for this (including which regions they are on) are documented above in the Manual Steps section
1. Once all testnet nodes are down, run `aws s3 cp s3://ndau-snapshots/latest-mainnet.txt s3://ndau-snapshots/latest-testnet.txt`
1. Optionally: move away all existing `snapshot-testnet-*.tgz` files since they are not valid history in the repaved world
    - There is a per-network-and-version place to store them if you want, e.g. `s3://ndau-snapshots/old/testnet-v1.2.0`
    - Once we start using the `commands` SHA as part of the snapshot name, this is likely not going to be worth the trouble; we can keep all old snapshots in a flat list at the root when that happens without worrying about blockchain compatibility problems
1. Back in the AWS Console, navigate to ECS > Clusters
    - "Create" all of the testnet services in the same regions they were before
    - Instructions for this are documented above in the Manual Steps section
    - Wait a minute or two between firing up each successive node, so that the daily restart timer in procmon doesn't cause them all to restart at near the same time
1. Testnet is now repaved and is an effective copy of recent mainnet
