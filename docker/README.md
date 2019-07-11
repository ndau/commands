# Single Docker Container

## Overview

How to build and run an ndau node using a single Docker container.  The Docker container contains all of our processes that make a node group: `redis`, `noms`, `ndaunode`, `tendermint` and `ndauapi`, all driven by `procmon`.  Running multiple instances of the container is how we make an ndau network.

This page outlines all of the features of the ndau Docker container, useful to Oneiro developers.  If you just want to get a node up and running, then the [Node Operator's Reference](node_operator.md) would be a good place to start.

## Build

1. Install Docker
1. Put `machine_user_key` from 1password into the `commands` repo root directory to gain access to private oneiro-ndev repos at image build time
1. Run `docker/bin/buildimage.sh` to build the `ndauimage` locally
1. Optionally you can upload the image to S3; instructions provided by `buildimge.sh` output

## Run

In order to run a container, we feed the following arguments to `docker/bin/runcontainer.sh`:
- Container name
- Tendermint P2P port
- Tendermint RPC port
- ndauapi port
- Node identity file (Optional)
- Snapshot name (Optional)
- Initial persistent peers (Optional)

For example:
```sh
IP=52.90.26.139
SNAPSHOT=snapshot-devnet-12345.tgz
docker/bin/runcontainer.sh \
    devnet-2 \
    26662 26672 3032 \
    node-identity-2.tgz \
    $SNAPSHOT \
    "$IP:26660,$IP:26661" \
    "$IP:26670,$IP:26671"
```

Below we look closer at each of these arguments.

### Container name

This can be anything you want.  For example, "devnet-0", "devnet-1", etc.  This name should be passed to other scripts in the `docker` directory to stop/restart/explore/remove the running container.

### Ports

Pass in the ports you'd like to use for communicating with the running node.  These will be mapped to the containers internal ports it uses for:

- Tendermint P2P port (needed for multiple containers to communicate with each other on the network)
- Tendermint RPC port (needed for our external tools, like the `ndau` tool, to send commands to the node)
- ndauapi port (needed for applications such as the Blockchain Explorer to have access to the ndauapi endpoints)

### Initial persistent peers

This can be `""` for the first node that is spun up on a network.

Currently, the persistent peers must be up and running before you start a new node that references them.  We could consider loosening this restriction, but it would require us to pass in "Tendermint node ids" otherwise.  Currently, `runcontainer.sh` requires the peers to be running and healthy as a way to verify that they are valid before attempting to run a container.

That means the first node takes no peers.  The 2nd node points to the first peer.  The 3rd node points to the first two peers, and so on.  After the initial N validator peers fire up, new verifier peers can start up and point to any subset of other currently running nodes on the network.

The format of this argument is a comma-separted list of peer IP, P2P and RPC ports: `"IP_ONE:P2P_ONE:RPC_ONE,IP_TWO:P2P_TWO:RPC_TWO,..."`

### Snapshot name

This is the name (minus `.tgz` extension) of the snapshot file to use for giving the new running node a blockchain starting point.  It's required.  We use localnet to generate the genesis snapshot.

1. Run an N-node localnet (see the [README](../README.md) at the root of the commands repo for how to set one up)
1. Run `<commands>/bin/snapshot.sh`
1. The snapshot will generate the following files in the `<commands>/bin/ndau-snapshots` directory:
    - `snapshot-devnet-12345.tgz` (This is a "snapshot of devnet at block height 12345")
    - `node-identity-0.tgz`, `node-identity-1.tgz`, ..., `node_identity-<N-1>.tgz`

The `snapshot.sh` script will print instructions on what to do with these files.  The snapshot file should be uploaded to S3.  The `node-identity-*.tgz` files should be kept secure by the node operators.  In this case, it's assumed that Oneiro will operate the first N nodes.  Oneiro must save the node identity files for each node somewhere in case any of the nodes need their containers re-run, possibly on a different server, or for any other reason.

In the above example, the snapshot name would be `snapshot-devnet-12345` and when the container runs, it'll pull down that snapshot from S3 and extract it to where it needs inside the container.

### Node identity file

This is a `node-identity.tgz` file generated either by the initial `snapshot.sh` or, more commonly, when new nodes are spun up for the first time by 3rd parties.

If you have a node identity file, pass it to `runcontainer.sh`.  If you don't have a node identity file, you can omit this argument and `runcontainer.sh` will generate one for you.  You will see a message printed from `runcontainer.sh` that will tell you where it is, and what to do with it.

You can also get at your `node-identity.tgz` file by running `docker cp <container>:/image/node-identity.tgz node-identity.tgz` to pull it out of your container at any time.  This works whether you passed one in originally, or if the container generated one for you.

## Scripts

In the `docker/bin` directory there are the following additional scripts:

- `stopcontainer.sh`
- `restartcontainer.sh`
- `exploreimage.sh`
- `explorecontainer.sh`
- `removecontainer.sh`
- `cleanupdocker.sh`

### `stopcontainer.sh`

This script takes the name of the container you'd like to stop.  The container can then be restarted at any time.

### `restartcontainer.sh`

This script takes the name of the container you'd like to restart.  The container should be currently stopped when you run this.  The container must have been started originally using `runcontainer.sh`.

A restart is similar to a run, but it doesn't need anything more than the container name to run.  All of environment variables, port mappings, and other config used at initial run time are still there in the container and will be reused.  e.g. It won't pull down the snapshot again; it'll just continue where it left off before the last time it was stopped.  When it reconnects to the network, it'll automatically catch up to the current height of the blockchain.

### `removecontainer.sh`

This script takes the name of the container you'd like to remove.  This does a `stopcontainer.sh` and then runs Docker commands to remove the container completely.  It's like a "delete node" command.

If you remove a container from Docker, you'll have to re-`runcontainer.sh` to start it back up again.  At that point, you may want to use the latest snapshot available, so that it doesn't have to catch up to the current blockchain height in as many steps as would be necessary if you re-run with an earlier (genesis) snapshot.

In this case, you should use your original `node-identity.tgz` file when you `runcontainer.sh` again, so that the re-run node will act like it was the same node it was before on the network.

### `exploreimage.sh`

This script takes no parameters as it assumes the image name is `ndauimage`.  It'll start a shell inside the image where you can poke around and see what's there.  `cd /image` once you get inside to see the ndau-specific stuff.

### `explorecontainer.sh`

This script takes the name of the container you'd like to explore.  It'll start a shell from inside the container.  `cd /image` once inside to see the node `logs` and `data` directories.

You can also run `docker container logs <container>` to see anything that's been printed to stdout when the node starts up.  This can be useful as a quick test to see if it's in good shape after just running/restarting your container.

### `cleanupdocker.sh`

Useful to run periodically to wipe out any old local docker images and containers that aren't currently running.  They pile up after awhile and you might run out of Docker space without periodic clean ups.

If you have other non-ndau-related containers in your Docker system, **you might not want to use this script**.  It'll remove any containers not currently running, ndau containers or otherwise.

## Demo

There's a `docker/demo` directory that can be used to fire up a 5-node network locally.  Once done, you can even run integration tests against this "Docker localnet" of containers, using `pytest -v --net=localnet` from the `integration-tests` repo root directory.

You can edit these files how you want, to test things out.  But if you'd like to just see it running with the fewest amount of steps, do the following:

1. `cd ~/go/src/github.com/oneiro-ndev/commands`
1. Run `bin/setup.sh 5 localnet` to set up a 5-validator-node localnet network
1. Run `bin/run.sh` and fill the blockchain with any transactions you want
1. Run `bin/snapshot.sh` to stop localnet and generate a snapshot
    - Upload it to S3 using the instructions printed by `snapshot.sh`
    - Edit `docker/demo/get_snapshot.sh` to make it return the name of the snapshot you just created
1. Run `docker/bin/buildimage.sh` if you haven't already
1. Run `docker/demo/run_all.sh` to run a 5-node network (4 original validators and 1 new verifier)
1. Run integration tests against it, or anything else you'd like to do on the local Docker network
1. Run `docker/demo/remove_all.sh` to stop and remove all 5 containers

## Honeycomb

To enable logging to honeycomb, simply have your `HONEYCOMB_DATASET` and `HONEYCOMB_KEY` environment variables exported before running a container.  Any restart of that container will preserve and reuse these settings.

## Validators and Verifiers

The nodes that exist when the genesis snapshot is taken on localnet will all be validator nodes.  Nodes that spin up and connect to a network after the initial validators are running will be verifier nodes by default.  We can control the power a node has (e.g. change a node from a verifier to a validator, and vice-versa) by using the ndau tool's `cvc` command.  We describe how to do that here.

This is currently not a one-button-push operation, but could be.  For now, the first thing you do is find out the public key of the node for which you want to change power.  That can be done using the following python script:

```python
#!/usr/bin/env python3

import base64
import json
import subprocess

info = subprocess.check_output(["./ndau", "info"])
info = json.loads(info)
pubkey_bytes = bytes(info["validator_info"]["pub_key"])
pubkey = base64.b64encode(pubkey_bytes).decode("utf-8").rstrip("=")
print(pubkey)
```

This uses the `ndau` tool get the non-padded base64 encoding of the public key for the node that is cofigured by the `ndautool.toml`.  Its location is determined by the current value of your `NDAUHOME` environment variable.

For example, if your `NDAUHOME` is `~/.ndau` and you have an `ndautool.toml` sitting in `$NDAUHOME/ndau/ndautool.toml` that points to the devnet-2 node on devnet, then a `cvc` command will affect that node.

So, using the above script (call it `pubkey.py`), you can then change the node's power to 10 using the following commands:

```sh
PUBKEY=$(./pubkey.py)
./ndau cvc "$PUBKEY" 10
```

By default, the initial 5 Oneiro-managed nodes on a network will all have power 10 (all validators).  Any new nodes that join a network will have power 0 (verifier).  The above steps can be used to grant or revoke validator status of any node at any time.

Once a verifier becomes a validator, transactions will round-robin with it to decide on which blocks can be added to the blockchain.
