#! /bin/bash

NDAUTOOL="$GOPATH/src/github.com/ndau/commands/ndau"
KEYTOOL="$GOPATH/src/github.com/ndau/commands/keytool"
SIGN="$GOPATH/src/github.com/ndau/commands/bin/YubiHSM/sign.py"

NODE='https://mainnet-2.ndau.tech:3030'
RFEACCT='ndeeh86uun6us9cenuck2uur679e37uczsmys33794gnvtfz'

SEQ=$(( `curl --silent --get $NODE/account/account/$RFEACCT | jq '.[].sequence'` + 1 ))
NAPU=`echo "scale=0; $2 * 100000000 / 1"| bc -l`
TX="{\"destination\": \"$1\", \"qty\": $NAPU, \"sequence\": $SEQ, \"signatures\": []}"

SB=`$NDAUTOOL signable-bytes rfe <<< $TX`
echo $SB > temp.SB

echo "Releasing" `printf "%'f\n" $2` "ndau from the Endowment to account" $1
echo "ARE YOU SURE?"
read -p "Insert Axiom YubiHSM key and press [Enter] to continue"

SIG=`$SIGN temp.SB 2001 | $KEYTOOL ed raw signature --stdin -b`
TXSIGNED=`echo $TX | sed "s/\[/\[\"$SIG\"/"`
curl -s -d "$TXSIGNED" -H "Content-Type: application/json" -X POST $NODE/tx/prevalidate/rfe | jq .

read -p "\nPress [Enter] to submit"
curl -s -d "$TXSIGNED" -H "Content-Type: application/json" -X POST $NODE/tx/submit/rfe | jq .

rm temp.SB