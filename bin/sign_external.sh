#!/bin/bash
# generate a ndau transaction and sign it externally, then send it
# this is intended to demo how to use a hardware device to sign transactions

# initialize
CMD_DIR="$(go env GOPATH)/src/github.com/oneiro-ndev/commands/"
cd "$CMD_DIR" || exit 1

./ndau -jk0 "$@" |\
    tee tx.json |\
    # fragile: this works for some common txs, but not all
    ./ndau signable-bytes -r "$1" > signable_bytes

# here we fake things: we want to sign with an external key, but
# I don't have a hardware key to play with, and don't know how its CLI
# works. Instead, we'll use keytool to generate and sign this tx.
# For real usage, ensure you replace this with a real yubikey cmd
signature=$(./keytool hd new | ./keytool sign -S --file signable_bytes)
# note that as we're sending a signature that we just invented right now,
# we can always expect to get an error of the form
# invalid signature(s): [0]

# ok, now just insert the signature into the transaction
# and send it
jq ".signatures += [\"$signature\"]" tx.json |\
    # fragile for the same reason as above: not all txs map to ndautool cmds
    ./ndau -v send "$1"


# clean up
rm -f tx.json signable_bytes
