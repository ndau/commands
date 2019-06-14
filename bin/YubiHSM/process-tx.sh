#! /bin/bash

# Usage: process-tx.sh <address> <txtype> <json> <YubiHSM object ID> <'main' or 'test'>
NDAUTOOL="$GOPATH/src/github.com/oneiro-ndev/commands/ndau"
KEYTOOL="$GOPATH/src/github.com/oneiro-ndev/commands/keytool"
SIGN="$GOPATH/src/github.com/oneiro-ndev/commands/bin/YubiHSM/sign.py"
YUBIHSM="/usr/local/bin/yubihsm-shell"

NODE='https://'$5'net-2.ndau.tech:3030'

export SEQ=$(( `curl --silent --get $NODE/account/account/$1 | jq '.[].sequence'` + 1 ))
TX=`cat $3 | jq '.sequence = (env.SEQ | tonumber)'`
SB=`$NDAUTOOL signable-bytes $2 <<< $TX`

read -p "Insert YubiHSM key and press [Enter] to continue"

SIG=`$YUBIHSM --connector=yhusb:// --action sign-eddsa -authkey 101 --object-id $4 --algorithm ed25519 --informat base64 --outformat base64 | $KEYTOOL ed raw signature --stdin -b`
TXSIGNED=`cat $3 | jq '.signatures += ["$SIG"]'`

read -p "\nPress [Enter] to prevalidate"
curl -s -d "$TXSIGNED" -H "Content-Type: application/json" -X POST $NODE/tx/prevalidate/rfe | jq .

read -p "\nPress [Enter] to submit"
curl -s -d "$TXSIGNED" -H "Content-Type: application/json" -X POST $NODE/tx/submit/rfe | jq .
