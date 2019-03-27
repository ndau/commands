# Single Docker container nodes

## Overview

How to build and run an ndau node using a single Docker container.

## Setup

TODO: Rewrite this document to be more complete.
      Do that after we've got it working with the ndau-only blockchain.
      Until then, here's some quick and dirty documentation...

1. Install Docker for Mac
1. Put `machine_user_key` from 1password in `<commands_repo>/docker/` to gain access to private oneiro-ndev repos
1. Get a snapshot from devnet, extract all the files into a directory structure like this:
    ```
    /path/to/snapshot/directory
      data/
        noms/...
        redis/dump.rdb
        tendermint/
          config/genesis.json
          data/...
    ```
1. `cd <commands_repo>/docker`
1. `bin/buildimage.sh`
1. Check out the `demo/` subdirectory for examples on how to run containers
