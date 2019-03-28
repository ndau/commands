#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
cd "$SCRIPT_DIR" || exit 1

IP=$(./get_ip.sh)

../bin/runcontainer.sh \
    localnet-4 26664 26674 3034 \
    "$IP:26660:26670,$IP:26661:26671,$IP:26662:26672,$IP:26663:26673" \
    snapshot-localnet-1
# This last node is not one of the initial validators, so there's no node-identity.tgz passed in.
