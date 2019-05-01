# Demo Docker Scripts

## Overview

This directory contains examples demonstrating how to use `runcontainer.sh` to create and run nodes on a network.

You'll need to be set up with a localnet before any of the demo scripts will work.  See the README at the root of the [commands](https://github.com/oneiro-ndev/commands) repo for how to set up a localnet.

## Run All

You can use the `run_all.sh` script to run a 6-node network (5 validators and 1 verifier) locally.  Before you can invoke it, though, you need to have the 5 validator `node-identity-*.tgz` files present where each individual `run*.sh` script is looking for them.

The following steps generate the local `node-identity-*.tgz` files in the right place:

1. `cd ~/go/src/github.com/oneiro-ndev/commands`
1. `bin/reset.sh 5 localnet`
1. `bin/run.sh`
1. `bin/snapshot.sh`

At this point you must upload the localnet snapshot to S3 using:

```sh
aws s3 cp \
    docker/ndau-snapshots/snapshot-localnet-1.tgz \
    s3://ndau-snapshots/snapshot-localnet-demo.tgz
```

Finally, `docker/demo/run_all.sh` can be executed.

The containers that start up will then use the snapshot you uploaded to S3 together with the local `node-identity-*.tgz` files on disk.  The snapshot contains the `genesis.json` file matching the 5 validator node ids in the local node identity files.

We currently do not support running with a local snapshot.  The snapshots are always pulled down from S3 from inside the container.  Conversely, we don't publish node identity files because those are secret and maintained privately by node operators.

To shut everything down, use `docker/demo/remove_all.sh`

## Add a Verifier Node to an Existing Network

You can use `docker/demo/add_node_to_network.sh testnet`, for example, to create a node named "testnet-test" locally and have it join testnet.  Other supported networks are "devnet" and "mainnet".

Use `docker/bin/removecontainer.sh testnet-test` (or `devnet-test` or `mainnet-test`) to stop and remove the new node to disconnect it from the network.

## Integration Tests

If you'd like to run [Integration Tests](https://github.com/oneiro-ndev/integration-tests) against a "local Docker net", it can be done similar to running them against a normal localnet.

Normally on a localnet, you'd run integration tests as follows:

1. `cd ~/go/src/github.com/oneiro-ndev/commands`
1. `bin/reset.sh 2 localnet`
1. `bin/run.sh`
1. `cd ../integration-tests`
1. `pytest -v`

It's similar for running integration tests against a local Docker net, because it'll use the same ndautool.toml file as a localnet.  The integration tests won't know any difference.  The local Docker containers will expose the same ports that a localnet uses so that a localnet config matches a local Docker net config, effectively.  Here are the steps:

1. `cd ~/go/src/github.com/oneiro-ndev/commands`
1. `bin/reset.sh 5 localnet`
    - We need at least two local nodes for integration testing, but the demo scripts assume 5 nodes
1. `bin/run.sh`
    - It needs to run once before snapshotting
1. `bin/snapshot.sh`
    - Be sure to upload it to S3 as `snapshot-localnet-demo.tgz` as described earlier
    - We need a snapshot so that local `node-identity-*.tgz` files are generated
1. `docker/demo/run_all.sh`
    - This runs all 5 local nodes that use the local node-identity files, plus a 6th node
1. `cd ../integration-tests`
1. `pytest -v`
