#! /bin/bash

NDAUTOOL="$GOPATH/src/github.com/oneiro-ndev/commands/ndau"
KEYTOOL="$GOPATH/src/github.com/oneiro-ndev/commands/keytool"

NETWORK='https://testnet-1.ndau.tech:3030'

NODES=(ndarw5i7rmqtqstw4mtnchmfvxnrq4k3e2ytsyvsc7nxt2y7 \
        ndaq3nqhez3vvxn8rx4m6s6n3kv7k9js8i3xw8hqnwvi2ete \
        ndahnsxr8zh7r6u685ka865wz77wb78xcn45rgskpeyiwuza \
        ndam75fnjn7cdues7ivi7ccfq8f534quieaccqibrvuzhqxa \
        ndaekyty73hd56gynsswuj5q9em68tp6ed5v7tpft872hvuc)

PKS=(npvtayjadtcbid5jfgyaxfdqmjagy3wcks2i9ugs8q37t5xy69uydj953dmjd4y3ufyrrv75vh7tfwvdbkfbds2dhdu7x85s79kuqz4hkj5wc6nuauzmszk3u74w \
    npvtayjadtcbidmztx5nhij7d4msqhjjizzrnuz8j2yjpa279kz9segh3diiyub7j8dmk7gykebpmds7ryjf4v46xebdp8ftsz7gagrpp2hdp9yizjz9dmag6mem \
    npvtayjadtcbibcdyjat959u82s242m84phfdmrf655s8xwtewhvqx5zssc892bfukcqwumevsnguwpd4e7kgw5wici7je5mbnkh46kbv534kaytrszb3u2yi35y \
    npvtayjadtcbic5m757sf4vr6v7bxjdysjf2htib7bs9uua23nmsxkgx93ybgs422r8stm7j6bpditgi476tdey396d2bf97vdnf8iyi4miakkxadxicr8nq57gx \
    npvtayjadtcbid77ujrn4atvry38ukaqxk6g4a54nw3tcf8ukktn4gi7ecvg35jswmywgp6g3mxbwmdw2nftf4ewqd3fxvw4mqws3gsumwyaiystnqse5tbz4c96)

for N in {0..4}; do
    SEQ=$(( `curl --silent --get $NETWORK/account/account/${NODES[$N]} | jq '.[].sequence'` + 1 ))
    echo "{\"node\": \"${NODES[$N]}\", \"sequence\": $SEQ, \"signatures\": []}" > crediteai.tmp
    SIG=`$KEYTOOL sign ${PKS[$N]} --file crediteai.tmp --txtype crediteai`
    TXSIGNED=`cat crediteai.tmp | sed "s/\[/\[\"$SIG\"/"`
    echo $TXSIGNED > crediteai.tmp
    PREVAL=`curl -s -d "$TXSIGNED" -H "Content-Type: application/json" -X POST $NETWORK/tx/prevalidate/crediteai`

    echo "Node" $N

    echo $PREVAL

    if [ `echo $PREVAL | jq .code` == 0 ]
    then
        curl -s -d "$TXSIGNED" -H "Content-Type: application/json" -X POST $NETWORK/tx/submit/crediteai
        echo
    else
        echo $PREVAL
    fi
    rm crediteai.tmp
done