#!/bin/bash
TXTYPE=Stake

TARGET_ADDRESS="ndad3uj3kjhgg9wbe8ic7w8ak9wubvspmcxjmcykbrm4tfeh"

RULES_ADDRESS="ndakfmgvxdm5rzenjabtt85mbma7aas2bhfzksqpegdu33v4"

STAKE_TO_ADDRESS="ndakfmgvxdm5rzenjabtt85mbma7aas2bhfzksqpegdu33v4"

VALIDATION_PRIVATE_KEY="npvtayjadtcbiadp3pnpbkhvmsits7tpfcqghtteqpiy3vkr992p78ddq4rmkbvp26bvnuev8secqhg4i5prkrqmu9m9m6f5ji2bzxvadcwzr2vjg8psd7i37m2u"

QTY_OF_STAKE=100000000000

SEQUENCE=2

read -d '' TX << EOF

{
  "target": "$TARGET_ADDRESS",
  "rules": "$RULES_ADDRESS",
  "stake_to": "$STAKE_TO_ADDRESS",
  "qty": $QTY_OF_STAKE,
  "sequence": $SEQUENCE
}
EOF

# create a b64 encoded string of signable bytes to be signed externally
SIGNABLE_BYTES=$(echo $TX | ./ndau signable-bytes "$TXTYPE")

# sign bytes of TX with validation private key 
SIGNATURE_1=$(./keytool sign $VALIDATION_PRIVATE_KEY "$SIGNABLE_BYTES" -b)
SIGNED_TX=$(echo $TX | jq '.signatures=["'$SIGNATURE_1'"]')

curl -H "Content-Type: application/json" -d "$SIGNED_TX" https://testnet-1.ndau.tech:3030/tx/submit/$TXTYPE
