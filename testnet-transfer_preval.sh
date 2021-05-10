#!/bin/bash
NDAUTOOL="$GOPATH/src/github.com/ndau/commands/ndau"
KEYTOOL="$GOPATH/src/github.com/ndau/commands/keytool"
NETWORK="https://testnet-2.ndau.tech:3030"
ACTION=prevalidate
TXTYPE=transfer

do-transfer1() {
    read -d '' TX << EOF
    {
        "source": "$SOURCE",
        "destination": "$DESTINATION",
        "qty": $AMOUNT,
        "sequence": $SEQUENCE
    }
EOF

    # create a b64 encoded string of signable bytes to be signed externally

    SIGNABLE_BYTES=$(echo $TX | $NDAUTOOL signable-bytes "transfer")
    # sign bytes of TX with private validation key
    SIGNATURE_1=$($KEYTOOL sign "$KEY1" "$SIGNABLE_BYTES" -b)
    SIGNED_TX=$(echo $TX | jq '.signatures=["'$SIGNATURE_1'"]')

    RESULT=`curl -s -H "Content-Type: application/json" -d "$SIGNED_TX" https://testnet-1.ndau.tech:3030/tx/$ACTION/$TXTYPE`
}

do-transfer2() {
    read -d '' TX << EOF
    {
        "source": "$SOURCE",
        "destination": "$DESTINATION",
        "qty": $AMOUNT,
        "sequence": $SEQUENCE
    }
EOF

    # create a b64 encoded string of signable bytes to be signed externally

    SIGNABLE_BYTES=$(echo $TX | $NDAUTOOL signable-bytes "transfer")

    # sign bytes of TX with private validation keys
    SIGNATURE_1=$($KEYTOOL sign "$KEY1" "$SIGNABLE_BYTES" -b)
    SIGNATURE_2=$($KEYTOOL sign "$KEY2" "$SIGNABLE_BYTES" -b)
    SIGNED_TX=$(echo $TX | jq '.signatures=["'$SIGNATURE_1'","'$SIGNATURE_2'"]')

    RESULT=`curl -s -H "Content-Type: application/json" -d "$SIGNED_TX" https://testnet-1.ndau.tech:3030/tx/$ACTION/$TXTYPE`
}

setup-transfer() {
    SEQUENCE=$(( `curl --silent --get $NETWORK/account/account/${SOURCE} | jq '.[].sequence'` + 1 ))
    BALANCE=$(( `curl --silent --get $NETWORK/account/account/${SOURCE} | jq '.[].balance'` + 1 ))
    AMOUNT=$(( $BALANCE / 2))
}

get-sib() {
    SIB=$(( `curl --silent --get $NETWORK/price/current | jq '.sib'` ))
    SIBAMT=$(( $SIB * $AMOUNT / 1000000000000 ))
}

# Exchange account 1

NDX1=ndxc99guypcqdvyjwbvpnq9xvh7uvb29xmmqf5iui3wrngut
NDX1PK1=npvtayjadtcbiaup4vcnvn3s3vskpmitt32eb89bjs7xzyrc5h8pzpqqg7eh9spidsm7jvqe3xitvzrkrw8i3zv4hx4c4zhajp8qhtdvzafbu9uvj453me3u43tn
NDX1PK2=npvtayjadtcbicryzeqw4ctn7wdj7zzeejg9aqxqadk964cbgevbnv4zekz8x43kuwdcqviq9y33xzpqndg8sdq2bpadme8d3uaynum78b92pgje9p27b5puvjk4

# Exchange account 2 (Oneiro)

NDX2=ndxnxrhi6e65tkzvbijzzjaia4x8xrrt6rt6p8rg5xsrfa2y
NDX2PK1=npvtayjadtcbibyzkjmtqb5rusc9szgkrhcku36ygpidpy79rgckhrqpzxefikdzf6r9jt9c6dmm9szhv5fs8cfpeggtb33knax9s7avfqj9m2uwiq492w75kd7z
NDX2PK2=npvtayjadtcbiauri3f9xbb52tfiqfur5jj4v7kkefvqih3v8a4mpbuayvtt255s8trr49yk8ic9r72f5geqjgq3z2gazcme5tcge7rtw6ixgdp3y5xtji6ajjx8

# Standard account 1

NDA1=ndaccku9ymmsjn4qhbwym89kdkekqegf7hnckpw9yz7guzvp
NDA1PK1=npvtayjadtcbia4iqujjykdersfygbxrxirhz8x2z44qmnme879tm7npir2k6px7hvki772dbg2rkkewjz98jrv9gz2675iaus6i2mgbkdh2k97ttws56ab9488q

# Standard account 2

NDA2=ndam6pxhem8gp5hgnduyk3m2w4i5in6h42k8d2apesvfvi46
NDA2PK1=npvtayjadtcbib83ihmv9k9twwxt94e7s8iayxztj68n8xfewhtujmkf259bb4zh9vcde6ruxt6kfw24b7qptzsqtdpbmtbbu633hsjr3ax8za5r78zf5cbie8wq

# Transfer - exchange to exchange

SOURCE=$NDX1
DESTINATION=$NDX2
KEY1=$NDX1PK1
KEY2=$NDX1PK2

setup-transfer
do-transfer2
get-sib
echo "Expected SIB: 0"
echo "Exchange to exchange:" $RESULT
echo


# Transfer - exchange to standard

SOURCE=$NDX2
DESTINATION=$NDA1
KEY1=$NDX2PK1
KEY2=$NDX2PK2

setup-transfer
do-transfer2
get-sib
echo "Expected SIB: 0"
echo "Exchange to standard:" $RESULT
echo


# Transfer - standard to standard

SOURCE=$NDA1
DESTINATION=$NDA2
KEY1=$NDA1PK1

setup-transfer
do-transfer1
get-sib
echo "Expected SIB: 0"
echo "Standard to standard:" $RESULT
echo


# Transfer - standard to exchange

SOURCE=$NDA2
DESTINATION=$NDX1
KEY1=$NDA2PK1

setup-transfer
do-transfer1
get-sib
echo "Expected SIB amount:" $SIBAMT
echo "Expected SIB value:" $SIB
echo "Quantity:" $AMOUNT
echo "Standard to exchange:" $RESULT
