# Node Operator's Reference

## Overview

These instructions allow a node operation to start a new ndau node and connect it to the ndau mainnet. Mainnet nodes may be run using Docker on macOS and Ubuntu hosts. Other Linux distros may also be usable but are untested.

## Setup

Ensure that all of the following are installed:

1. [Docker Desktop](https://docker.com)
1. `curl` (`sudo apt install curl` on Linux)
1. `jq` (`brew install jq` on macOS)

The following are also required but are likely to already be installed on your system:

1. `ping`
1. `nc`

### Ubuntu Notes

Ubuntu restricts access to TCP ports for non-root users. To run Docker as a non-root user you must further configure your Docker installation. Please consult the [official Docker documentation for instructions](https://docs.docker.com/install/linux/linux-postinstall).

## Build

There is nothing to build.  The `ndauimage` Docker image is served on S3 and `runcontainer.sh` will pull down the proper version if it's not found locally.

## Run

The `docker/bin/runcontainer.sh` script will create a container based off of a Docker image named `ndauimage` which will be installed into your Docker environment automatically.  Here's how to run it:

```sh
# Give your node a name.
NODENAME=my-node
# Choose the network - mainnet, testnet, devnet, localnet
NDAU_NETWORK=mainnet

# Choose the ports you would like to use for...
# ...communication with other nodes on the network.
# ...responding to RPC requests to your node.
# ...and responding to ndau API requests to your node.

P2P_PORT=26665
RPC_PORT=26675
API_PORT=3035

# Create and run your node, connecting it to $NDAU_NETWORK
docker/bin/runcontainer.sh $NDAU_NETWORK $NODENAME $P2P_PORT $RPC_PORT $API_PORT
```

You now have created a node (Docker container) named "my-node", running and connected to mainnet.  It will catch up to the latest block height on the network since the height found in the latest snapshot on S3.

IMPORTANT: Read the information printed by `runcontainer.sh` about the `node-identity.tgz` file that it will generate for you.  **You must keep this secure and use it again (discussed below) if you ever need to run your node from scratch.**  It won't be needed if you want to stop/restart your node (Docker container).  It is only needed if you lose your container, or decide to redeploy it with different ports, or for any other reason. **If you intend to stake ndau and register your node as a validator node, you must retain the same node identity across restarts.**

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

This would allow you to use `runcontainer.sh` again, to allow you to change which ports you expose from your container, for example.

You can remove/run your node as needed.  Think of `removecontainer.sh` as the counterpart to `runcontainer.sh`

## Re-Run

If you lose your node or Docker container, want to start it from scratch, move it to a new deployment environment, or for any other reason, you'll want to use the `node-identity.tgz` file that your first run of `runcontainer.sh` produced.  That way, when you run your node again, it'll "be the same node" that it was before.  It'll catch up to the latest block height, and continue normally.

Follow the original "Run" steps documented earlier, but also pass in the path to your node identity file:

```sh
IDENTITY=/path/to/your/node-identity.tgz
docker/bin/runcontainer.sh $NDAU_NETWORK $NODENAME $P2P_PORT $RPC_PORT $API_PORT $IDENTITY
```

It'll now be running and connected to mainnet, and will catch up to the latest block height from the latest snapshot on S3, but this time it'll use the given node identity for itself rather than generate a new one.
