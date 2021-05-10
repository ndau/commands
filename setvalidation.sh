#!/bin/bash
TXTYPE=SetValidation

TARGET_ADDRESS="ndad3uj3kjhgg9wbe8ic7w8ak9wubvspmcxjmcykbrm4tfeh"

OWNERSHIP_PRIVATE_KEY="npvtayjadtcbia78jebf4kx7rhq9x7zhayp8dipggyghvbfpbxy8vd6ar7pweze36ut6hwq4jz7us476b8cjq848dvfnrsucj4g74tesfsjuv7t76bw95ruw3xib"

OWNERSHIP_PUBLIC_KEY="npuba8jadtbbebfd2ri7wvr5fbx52d6eu77x6hgk29beevwp5xcjamavfh5d52dj9venggnjhugy"

VALIDATION_PUBLIC_KEY="npuba8jadtbbeb2dg3ejh7aie6qpwty48w86zf8z8z2myutsdrmgagfjq9tgup65bmi882v7ktrd"

VALIDATION_PRIVATE_KEY="npvtayjadtcbiadp3pnpbkhvmsits7tpfcqghtteqpiy3vkr992p78ddq4rmkbvp26bvnuev8secqhg4i5prkrqmu9m9m6f5ji2bzxvadcwzr2vjg8psd7i37m2u"

SEQUENCE=1

read -d '' TX << EOF
{
  "target": "$TARGET_ADDRESS",
  "ownership": "$OWNERSHIP_PUBLIC_KEY",
  "validation_keys": [
    "$VALIDATION_PUBLIC_KEY"
  ],
  "sequence": $SEQUENCE
}
EOF

# create a b64 encoded string of signable bytes to be signed externally
SIGNABLE_BYTES=$(echo $TX | ./ndau signable-bytes "$TXTYPE")

# sign bytes of TX with ownership private key 
SIGNATURE_1=$(./keytool sign $OWNERSHIP_PRIVATE_KEY "$SIGNABLE_BYTES" -b)
SIGNED_TX=$(echo $TX | jq '.signature="'$SIGNATURE_1'"')

curl -H "Content-Type: application/json" -d "$SIGNED_TX" https://testnet-1.ndau.tech:3030/tx/submit/$TXTYPE

