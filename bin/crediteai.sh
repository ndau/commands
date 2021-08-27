#! /bin/bash

# crediteai.sh - Example script for crediting EAI to accounts delegated to a node.
# Usage: ./crediteai.sh [node to submit transaction to]
# Defaults to a mainnet node if omitted.

KEYTOOL="$GOPATH/src/github.com/ndau/commands/keytool"

if [ -z $1 ]
then
    echo "Usage: crediteai.sh <full node url>"
    exit
else
    SUBMIT_TO=$1
fi

NODE_ADDRESS="`cat address`"
VALIDATION_KEY="`cat validation_private`"

SEQ=$(( `curl --silent --get $SUBMIT_TO/account/account/$NODE_ADDRESS | jq '.[].sequence'` + 1 ))
echo "{\"node\": \"$NODE_ADDRESS\", \"sequence\": $SEQ, \"signatures\": []}" > crediteai.tmp
SIG=`$KEYTOOL sign $VALIDATION_KEY --file crediteai.tmp --txtype crediteai`
TXSIGNED=`cat crediteai.tmp | sed "s/\[/\[\"$SIG\"/"`
echo $TXSIGNED > crediteai.tmp
PREVAL=`curl -s -d "$TXSIGNED" -H "Content-Type: application/json" -X POST $SUBMIT_TO/tx/prevalidate/crediteai`

if [ `echo $PREVAL | jq .code` == 0 ]
then
    curl -s -d "$TXSIGNED" -H "Content-Type: application/json" -X POST "$SUBMIT_TO/tx/submit/crediteai"
    echo
else
    echo " Transaction did not prevalidate " "$PREVAL"
fi
rm crediteai.tmp
