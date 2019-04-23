# Node Operator's Reference

## Overview

How to create a new node and connect it to the ndau mainnet.

## Setup

Ensure that all of the following are installed:

1. Docker
1. Xcode command-line tools: `xcode-select --install`
1. [Brew](https://brew.sh/)
1. Install `jq` if needed, using: `brew install jq`

The following are also required but are likely to already be installed on your system:

1. `curl`
1. `nc`

## Run

The `docker/bin/runcontainer.sh` script will create a container based off of a Docker image named `ndauimage` which will be installed into your Docker environment automatically.  Here's how to run it:

```sh
# Give your node a name.
NODENAME=my-node

# This is the ndau mainnet genesis snapshot.
SNAPSHOT=snapshot-mainnet-47

# These are the ports you would like to use for...
P2P_PORT=26665 # ...communication with other nodes on the network.
RPC_PORT=26675 # ...responding to RPC requests to your node.
API_PORT=3035  # ...responding to ndau API requests to your node.

# Create and run your node, connecting it to mainnet.
NDAU_NETWORK=mainnet docker/bin/runcontainer.sh $NODENAME $P2P_PORT $RPC_PORT $API_PORT $SNAPSHOT
```

You now have created a node (Docker container) named "my-node", running and connected to mainnet.  It will catch up to the latest block height on the network since the height found in the given snapshot.

IMPORTANT: Read the information printed by `runcontainer.sh` about the `node-identity.tgz` file that it will generate for you.  You must keep this secure and use it again (discussed below) if you ever need to run your node from scratch.  It won't be needed if you want to stop/restart your node (Docker container).  It is only needed if you lose your container, or decide to redeploy it with different ports, or for any other reason.

## Stop

To stop your node, you can use:

```sh
docker/bin/stopcontainer.sh $NODENAME
```

This will remove the node from the network, but can be restarted to rejoin the network at any time.

## Restart

To restart your stopped node, you can use:

```sh
docker/bin/restartcontainer.sh $NODENAME
```

This will spin up the node again and it will reconnect to mainnet.  It will continue where it left off, and catch up to the other nodes on the network to reach the same block height.

You can stop/restart your node as needed.  Think of `restartcontainer.sh` as the counterpart to `stopcontainer.sh`

## Remove

To remove your node from the network (and your local Docker environment), you can use:

```sh
docker/bin/removecontainer.sh $NODENAME
```

This would allow you to use `runcontainer.sh` again, perhaps with a newer snapshot, or different ports.

You can remove/run your node as needed.  Think of `removecontainer.sh` as the counterpart to `runcontainer.sh`

## Re-Run

If you lose your node, or Docker container, or want to start it from scratch, if moving it to a new deployment environment, or for any other reason, you'll want to use the `node-identity.tgz` file that your first run of `runcontainer.sh` produced.  That way, when you run your node again, it'll "be the same node" that it was before.  It'll catch up to the latest block height, and continue normally.

Follow the original "Run" steps documented earlier, but also pass in the path to your node identity file:

```sh
IDENTITY=/path/to/your/node-identity.tgz
NDAU_NETWORK=mainnet docker/bin/runcontainer.sh $NODENAME $P2P_PORT $RPC_PORT $API_PORT $SNAPSHOT $IDENTITY
```

It'll now be running and connected to mainnet, and will catch up to the latest block height from the given snapshot, but this time it'll use the given node identity for itself rather than generate a new one.
