#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
cd "$SCRIPT_DIR" || exit 1

IP=$(./get_ip.sh)

../bin/runcontainer.sh \
    localnet-3 26663 26673 3033 \
    "$IP:26660:26670,$IP:26661:26671,$IP:26662:26672" \
    snapshot-localnet-1 \
    ../../bin/ndau-snapshots/node-identity-3.tgz
