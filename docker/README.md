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
      svi-namespace
      data/
        chaos/
          noms/...
          redis/dump.rdb
          tendermint/
            config/genesis.json
            data/...
        ndau/
          noms/...
          redis/dump.rdb
          tendermint/
            config/genesis.json
            data/...
    ```
1. `cd <commands_repo>/docker`
1. `bin/buildimage.sh`
1. `bin/runcontainer.sh devnet-5 26660 26670 26661 26671 3030 52.90.26.139 30054 /path/to/snapshot/directory`

* Make sure the IP address and port `52.90.26.139:30054` is where `devnet-0` is currently
* After you run the above, you'll have a new node called `devnet-5` talking to `devnet`
  - It'll be stuck, though, if devnet has any sidechain tx's on it.  Still, it's a fully running node, doing the best it can before we have the ndau-only blockchain working.  You can shell into the container and see what it's doing.  In the container, the `/image` dir has all the interesting stuff.
