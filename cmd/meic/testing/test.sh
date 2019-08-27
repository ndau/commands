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
# 3. Run the uat bitmart API
# 4. Run the issuance service
#    - it must wait for a connection for the as-yet-not-running signing service
#    - after receiving the connection, it must create exactly one Issue tx on the blockchain
# 5. Run the signing service
#    - This connects to the issuance service and kicks everything off
# 6. Notice a new Issue tx on the blockchain, or time out
# 7. Shut down everything
#    - bring down the issuance service first
#    - this should automatically bring down the signing service
#    - shut down the uat bitmart API
#
# IMPORTANT: this assumes a localnet is running, the ndau tool is properly
# configured to talk to it, and appropriate keys are assigned. The tooling
# tests its assumptions and tries to give good errors when required, but
# it is outside the scope of this test script to build, configure, or bring up
# or down a localnet.
#
# Required tools:
# - jq          https://github.com/stedolan/jq
# - msgpack-cli https://github.com/jakm/msgpack-cli
# - remarshal   https://github.com/dbohdan/remarshal (toml2json, json2toml)
# - wsta        https://github.com/esphen/wsta

commands_path="$GOPATH/src/github.com/oneiro-ndev/commands"
meic_path="$commands_path/cmd/meic"
testing_path="$meic_path/testing"
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
sa=~/.localnet/genesis_files/system_vars.toml
if [ ! -f "$sa" ]; then
    echo "$sa not found; aborting"
    exit 1
fi
rfea_local=$(toml2json $sa | jq .ReleaseFromEndowmentAddress.data --raw-output)
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

echo "$NDAU" account query -a "$rfea_chain"
"$NDAU" account query -a "$rfea_chain" | jq .
"$NDAU" account query -a "$rfea_chain" | jq '.validationKeys'

if [ "$rfe_validation_public_local" != "$rfe_validation_public_chain" ]; then
    echo "ReleaseFromEndowment validation keys mismatch:"
    echo "  local rfe: $rfea_local"
    echo "  chain rfe: $rfea_chain"
    echo "  local: $rfe_validation_public_local"
    echo "  chain: $rfe_validation_public_chain"
#    exit 1
fi
rfe_validation_public=$(echo "$rfe_validation_public_chain" | jq '.[0]' --raw-output)
# rfe_validation_private=npvtayjadtcbick79eu599f8w5aaeg8iqnj8bxmpztgghnbkzatsw4be27f2at43n93pxk6e6f5ckuvpycvm8fhvk8wumnptf57esd23b75vximaknveqsipy9ev
rfe_validation_private=npvtayjadtcbib2u2rnem47vgqqifsbt8r8y2ktdygtfkaw98imgpjw25mqztt72b7k2scgjyhypk4twibph5m88sw3xmg3xe77m9vmsyfhq32ywn7z63bmm4zix
# rfe_validation_private=$(
#     toml2json "$sa" |\
#     jq '.ReleaseFromEndowmentValidationPrivate' --raw-output
# )

# inject the validation keys into sigconfig.toml
sigconfig="$testing_path/sigconfig.toml"
sigconfig_json=$(toml2json "$sigconfig")
sigconfig_json=$(
    echo "$sigconfig_json" |\
    jq ".virtual |= . + {R1: {
        pub_key: $(echo "$rfe_validation_public_chain" | jq ".[0]"),
        priv_key: \"$rfe_validation_private\"
    }}"
)
echo "$sigconfig_json" | json2toml > "$sigconfig"

# generate a public/private keypair for this instance of the issuance service
issuance_service_private=$("$KEYTOOL" hd new)
issuance_service_public=$("$KEYTOOL" hd public "$issuance_service_private")
# issuance_service_private=npvtayjadtcbia9vpqqci2hxb2msrmh44kqpmcqj937tsc3bpx22a6tbmzrw99rikd6quz5755e99ng3u27ff7gj7pf6ttqe23zxhyqprum6a94bpr3ji84tcfu2
# issuance_service_public=npuba8jadtbbeah27frz5zyj982pvft4km4nv44m3dc6jtvrkrn649ez2b9wc49uvuuzgctk93y9

# get the RPC address the ndau tool is using
rpc=$(toml2json "$("$NDAU" conf-path)" | jq .node --raw-output)

# get the websocket address to which the client signature service will connect
ws_addr=$(echo "$sigconfig_json" | jq .connections.local.url --raw-output)

# ensure there's plenty of un-issued RFE'd ndau floating around
# this is an arbitrary address; nobody's expected to have access to it
# "$NDAU" rfe 50000 --address ndaaiz75f4ejxp3gdxb7eqct4wuyukrj36epf245qaeifcw2

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

# echo starting mock bitmart api
# echo go run "$testing_path/mock_bitmart" &
# go run "$testing_path/mock_bitmart" &
# sleep 2

echo starting issuance service
echo go run "$meic_path" \
    --server-addr "$ws_addr" \
    --node-addr "$rpc" \
    --server-pub-key "$issuance_service_public" \
    --server-pvt-key "$issuance_service_private" \
    -c ./config.toml
go run "$meic_path" \
    --server-addr "$ws_addr" \
    --node-addr "$rpc" \
    --server-pub-key "$issuance_service_public" \
    --server-pvt-key "$issuance_service_private" \
    -c ./config.toml \
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
echo wsta "$wrpc/websocket" "$subscribe_to_txs"
wsta "$wrpc/websocket" "$subscribe_to_txs" > "$datafile" &
sleep 2


# start the signing service client pointing to our configuration
echo starting signing service client
echo go run "$signing_service_path" -c "$sigconfig"
go run "$signing_service_path" -c "$sigconfig" &

# give everything a bit to get settled
echo "processing, please wait"
sleep 600

# The sequence of events when we went to sleep just then:
#
# - Signing service starts up, connects to all its defined connections:
#   just the issuance service, as defined in sigconfig.toml
# - Issuance service notices the incoming connection, starts its loop.
# - Issuance service sends a request to bitmart asking for the list of all
#   new trades. Because there is an "endpoint" parameter in test.apikey.json,
#   it actually sends that request to the uat api.
# - uat api says, yup, here are some new trades.
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

# if the datafile contains no tx with a hash, then this definitely failed
if ! grep -q "tx.hash" "$datafile"; then
    echo "no transaction appeared in websocket tx stream ($datafile)"
    exit 1
fi

# ensure the tx was an Issue tx
# it's possible that external users submitted transactions; let's protect against
# that case.
foundissue=false
transactions=$(jq --raw-output '.result.data.value.TxResult.tx | values' "$datafile")
# this may produce multiple lines, so loop
for tx in $transactions; do
    txid=$(echo "$tx" | base64 -D | msgpack-cli decode | jq .TransactableID)
    if [ "$txid" == 20 ]; then
        # issue txid is 20: https://github.com/oneiro-ndev/ndau/blob/370c97d92f6cff6aa04f123819d232fb8a2dfd27/pkg/ndau/transactions.go#L36
        foundissue=true
        break
    fi
done

# we can't know for sure whether or not we submitted that issue tx, or whether some
# external party submitted it, but it's pretty suggestive evidence. Let's just
# treat it as if it were accurate:
if "$foundissue"; then
    echo "found an issue tx; assuming this means success"
else
    echo "no issue tx found in transactions; something went wrong :("
fi

sleep 600
# run the program again to generate the right return code.
"$foundissue"
