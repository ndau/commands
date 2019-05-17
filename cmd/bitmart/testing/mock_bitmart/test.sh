#!/usr/bin/env bash

set -e -x

# This test script attempts to get the issuance service up and running in a
# test environment. Fundamentally, this requires these steps:
#
# 1. Assume that the ndau tool is connected to a running localnet
#    - verify that the ndau tool exists
#    - verify that the ndau tool is pointing to a running blockchain
# 2. Configure the signing service
#    - check for the existence of `system_accounts.toml`
#    - ensure the issuance address in `system_accounts.toml` matches the running sysvar
#    - copy the appropriate keys from `system_accounts.toml` into `sigconfig.toml`
# 3. Run the mock bitmart API
# 4. Run the issuance service
#    - it must wait for a connection for the as-yet-not-running signing service
#    - after receiving the connection, it must create exactly one Issue tx on the blockchain
# 5. Run the signing service
#    - This connects to the issuance service and kicks everything off
# 6. Notice a new Issue tx on the blockchain, or time out
# 7. Shut down everything
#    - bring down the issuance service first
#    - this should automatically bring down the signing service
#    - shut down the mock bitmart API
#
# IMPORTANT: this assumes a localnet is running, the ndau tool is properly
# configured to talk to it, and appropriate keys are assigned. The tooling
# tests its assumptions and tries to give good errors when required, but
# it is outside the scope of this test script to build, configure, or bring up
# or down a localnet.
#
# Required tools:
# - jq        https://github.com/stedolan/jq
# - remarshal https://github.com/dbohdan/remarshal (toml2json, json2toml)
# - wsta      https://github.com/esphen/wsta

commands_path="$GOPATH/src/github.com/oneiro-ndev/commands"
bitmart_path="$commands_path/cmd/bitmart"
testing_path="$bitmart_path/testing"
mock_bitmart_path="$testing_path/mock_bitmart"

# set up the ndau tool
if [ -z "$GOPATH" ]; then
    echo "GOPATH empty; aborting"
    exit 1
fi
maybe_ndaupath=(
    "$commands_path/ndau"
    "$commands_path/cmd/ndau/ndau"
)
for ndaupath in "${maybe_ndaupath[@]}"; do
    if [ -x "$ndaupath" ]; then
        NDAU="$ndaupath"
        break
    fi
done
if [ -z "$NDAU" ]; then
    echo "ndau tool not found; aborting"
    exit 1
fi
if ! "$NDAU" info > /dev/null; then
    echo "could not connect to blockchain"
    exit 1
fi

# set up keytool
maybe_ktpath=(
    "$commands_path/keytool"
    "$commands_path/cmd/keytool/keytool"
)
for ktpath in "${maybe_ktpath[@]}"; do
    if [ -x "$ktpath" ]; then
        KEYTOOL="$ktpath"
        break
    fi
done
if [ -z "$KEYTOOL" ]; then
    echo "keytool not found; aborting"
    exit 1
fi

# configure the signing service
sa=~/.localnet/genesis_files/system_accounts.toml
if [ ! -f "$sa" ]; then
    echo "$sa not found; aborting"
    exit 1
fi
rfea_local=$(toml2json $sa | jq .ReleaseFromEndowmentAddress --raw-output)
rfea_chain=$(
    "$NDAU" sysvar get ReleaseFromEndowmentAddress |\
    jq .ReleaseFromEndowmentAddress[0]? --raw-output
)
if [ "$rfea_local" != "$rfea_chain" ]; then
    echo "ReleaseFromEndowmentAddress mismatch:"
    echo "  local: $rfea_local"
    echo "  chain: $rfea_chain"
    exit 1
fi

# ensure we have the correct and only validation key
rfe_validation_public_local=$(
    toml2json "$sa" |\
    jq '[.ReleaseFromEndowmentValidation]' --compact-output
)
rfe_validation_public_chain=$(
    "$NDAU" account query -a "$rfea_chain" |\
    jq '.validationKeys' --compact-output
)
if [ "$rfe_validation_public_local" != "$rfe_validation_public_chain" ]; then
    echo "ReleaseFromEndowment validation keys mismatch:"
    echo "  local: $rfe_validation_public_local"
    echo "  chain: $rfe_validation_public_chain"
    exit 1
fi
rfe_validation_private=$(
    toml2json "$sa" |\
    jq '.ReleaseFromEndowmentValidationPrivate' --raw-output
)

# inject the validation keys into sigconfig.toml
sigconfig="$mock_bitmart_path/sigconfig.toml"
sigconfig_json=$(toml2json "$sigconfig")
sigconfig_json=$(
    echo "$sigconfig_json" |\
    jq ".keys |= . + {issue: {
        type: \"virtual\",
        pub_key: $(echo "$rfe_validation_public_local" | jq ".[0]"),
        priv_key: \"$rfe_validation_private\"
    }}"
)
echo "$sigconfig_json" | json2toml > "$sigconfig"

# generate a public/private keypair for this instance of the issuance service
issuance_service_private=$("$KEYTOOL" hd new)
issuance_service_public=$("$KEYTOOL" hd public "$issuance_service_private")

# get the RPC address the ndau tool is using
rpc=$(toml2json "$("$NDAU" conf-path)" | jq .node | sed -E 's/https?/ws/')

# get the websocket address to which the client signature service will connect
ws_addr=$(echo "$sigconfig_json" | jq .connections.local.url --raw-output)

# ensure there's plenty of un-issued RFE'd ndau floating around
# this is an arbitrary address; nobody's expected to have access to it
"$NDAU" rfe 1000 --address ndaaiz75f4ejxp3gdxb7eqct4wuyukrj36epf245qaeifcw2

# let's start running things!
# mock bitmart api
go run "$testing_path/mock_bitmart" &
mock_bitmart_pid="$!"

# issuance service
go run "$bitmart_path" \
    "$mock_bitmart_path/test.apikey.json" \
    "$rpc" \
    "$issuance_service_public" \
    --priv-key "$issuance_service_private" \
    --serve "$ws_addr"
    "$rfe_validation_public_chain" \
    &
issuance_service_pid="$!"

# we're going to use this message to subscribe to tendermint's tx notification,
# so that we can detect whether or not an issue tx went through
subscribe_to_txs="{
    \"jsonrpc\":\"2.0\",
    \"method\":\"subscribe\",
    \"id\":\"0\",
    \"params\":{
        \"query\":\"tm.event='Tx'\"
    }
}"

# start watching for transactions
wsta "$rpc/websocket" "$subscribe_to_txs" > "$testing_path/websocket.data"
wsta_pid="$!"

# TODO:
# - start the signing service client pointing to the appropriate sigconfig.toml,
# - notice an issue tx in websocket.data or timeout
# - shut down everything
