#! /bin/bash

# Usage: process-tx.sh <address> <txtype> <json file> <'main' or 'test'> <number of signatures>

NDAUTOOL="$GOPATH/src/github.com/oneiro-ndev/commands/ndau"
KEYTOOL="$GOPATH/src/github.com/oneiro-ndev/commands/keytool"
SIGN="$GOPATH/src/github.com/oneiro-ndev/commands/bin/YubiHSM/sign.py"
YUBIHSM="/usr/local/bin/yubihsm-shell"

NODE='https://'$4'net-2.ndau.tech:3030'

SEQ=$(( `curl --silent --get $NODE/account/account/$1 | jq ".$1.sequence"` + 1))
TX=`cat $3 | jq ".sequence = $SEQ"`
echo $TX

SB=`echo $TX | $NDAUTOOL signable-bytes $2`

for i in $(seq 1 1 $5);
do
    read -p "Insert YubiHSM key $i - enter object ID for signing: " OBJID
    SIG=`echo $SB | $YUBIHSM --connector=yhusb:// --action sign-eddsa --authkey 101 --object-id $OBJID --algorithm ed25519 --informat base64 --outformat base64 | $KEYTOOL ed raw signature --stdin -b`
    TX=`echo $TX | jq ".signatures += [\"$SIG\"]"`
done

echo $TX | jq .

read -p "Press [Enter] to prevalidate"
curl -s -d "$TX" -H "Content-Type: application/json" -X POST $NODE/tx/prevalidate/$2 | jq .

read -p "Press [Enter] to submit"
curl -s -d "$TX" -H "Content-Type: application/json" -X POST $NODE/tx/submit/$2 | jq .
