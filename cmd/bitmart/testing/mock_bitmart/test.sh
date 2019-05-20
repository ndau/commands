#!/usr/bin/env bash

set -e

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
signing_service_path="$commands_path/../recovery/cmd/signer"

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
rfe_validation_public=$(echo "$rfe_validation_public_chain" | jq '.[0]' --raw-output)
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
rpc=$(toml2json "$("$NDAU" conf-path)" | jq .node --raw-output)

# get the websocket address to which the client signature service will connect
ws_addr=$(echo "$sigconfig_json" | jq .connections.local.url --raw-output)

# ensure there's plenty of un-issued RFE'd ndau floating around
# this is an arbitrary address; nobody's expected to have access to it
"$NDAU" rfe 50000 --address ndaaiz75f4ejxp3gdxb7eqct4wuyukrj36epf245qaeifcw2

# let's start running things!
# before we start: we're going to be running several background tasks,
# and we want them to just die when we exit. Because we've set -e to quit if
# anything fails, we can't know the full list of what will be running ahead
# of time. Happily, it's not hard to kill background jobs in an exit function.

killsubp () {
    # kill direct subprocesses
    for job in $(jobs -p); do
        kill "$job"
    done
    # go run does a double-fork, orphaning the process, so we have to
    # just kill everything that's been go-built in the current process table
    pkill -f 'go-build'
}

trap killsubp EXIT

echo starting mock bitmart api
go run "$testing_path/mock_bitmart" &
sleep 2

echo starting issuance service
go run "$bitmart_path" \
    "$mock_bitmart_path/test.apikey.json" \
    "$rpc" \
    "$issuance_service_public" \
    --priv-key "$issuance_service_private" \
    --serve "$ws_addr" \
    "$rfe_validation_public" \
    &
sleep 2

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
datafile="$testing_path/websocket.data"
# wsta is picky about protocols
wrpc=$(echo "$rpc" | sed -E 's/http/ws/')
echo starting websocket transfer agent
wsta "$wrpc/websocket" "$subscribe_to_txs" > "$datafile" &
sleep 2


# start the signing service client pointing to our configuration
echo starting signing service client
go run "$signing_service_path" -c "$sigconfig" &

# give everything a bit to get settled
echo "processing, please wait"
sleep 10

# The sequence of events when we went to sleep just then:
#
# - Signing service starts up, connects to all its defined connections:
#   just the issuance service, as defined in sigconfig.toml
# - Issuance service notices the incoming connection, starts its loop.
# - Issuance service sends a request to bitmart asking for the list of all
#   new trades. Because there is an "endpoint" parameter in test.apikey.json,
#   it actually sends that request to the mock api.
# - Mock api says, yup, here are some new trades.
# - Issuance service goes back and forth establishing a list of only those
#   trades which were sales.
# - Issuance service adds up the qty of the sales and generates an appropriate
#   Issue tx. However, the tx is incomplete: the issuance service doesn't have
#   appropriate private keys to sign it with.
# - Issuance service sends the tx to the signing service and asks for it to be
#   signed with the list of signing keys from its invocation.
# - Signing service discovers that it does in fact have that key configured
#   (as a virtual key in sigconfig.toml), signs the signable bytes of the tx,
#   and returns the signature.
# - Issuance service applies that signature to the tx and sends it to the
#   blockchain, then goes to sleep for the poll interval.
# - Blockchain accepts this request, hopefully. This creates a "Tx" event on
#   its internal event loop, which in due course gets pushed to all connected
#   tx subscriptions.
# - WSTA receives the event, pretty-prints it to stdout, which is redirected
#   to a file we know to watch.
#
# That should all hopefully take much less than 5 seconds, but we don't really
# have a good way to detect when it's complete, so we just overkill a little
# waiting for it to all resolve.

cat "$datafile"
# if the datafile contains a tx with a hash, it's ours, which means that this
# succeeded, so we should return success
grep -q "tx.hash" "$datafile"
