#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
source "$SCRIPT_DIR"/docker-env.sh

# The PERSISTENT_PEERS may contain domain names.  Convert them to IPs.
PERSISTENT_PEERS_WITH_IPS="$PERSISTENT_PEERS"

sed -i -E \
    -e 's|^(persistent_peers =) (.*)|\1 "'"$PERSISTENT_PEERS_WITH_IPS"'"|' \
    "$TM_DATA_DIR/config/config.toml"
