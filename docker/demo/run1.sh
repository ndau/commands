#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
cd "$SCRIPT_DIR" || exit 1

IP=$(./get_ip.sh)

../bin/runcontainer.sh \
    localnet-1 26661 26671 3031 \
    "$IP:26660:26670" \
    snapshot-localnet-1 \
    ../../bin/ndau-snapshots/node-identity-1.tgz
