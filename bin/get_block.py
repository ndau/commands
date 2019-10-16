#!/usr/bin/env python3

#  ----- ---- --- -- -
#  Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
# 
#  Licensed under the Apache License 2.0 (the "License").  You may not use
#  this file except in compliance with the License.  You can obtain a copy
#  in the file LICENSE in the source distribution or at
#  https://www.apache.org/licenses/LICENSE-2.0.txt
#  - -- --- ---- -----

import re
from base64 import b64decode
from pathlib import Path
from pprint import pformat
from urllib.parse import urljoin

import requests

import msgpack

SERVICES = "https://s3.us-east-2.amazonaws.com/ndau-json/services.json"
TXS = re.compile(r"TxID\((?P<txid>\d+)\):\s*&(?P<name>\w+)\{\},")

# TX names
tx_file = Path(__file__).parent / Path(
    "../vendor/github.com/oneiro-ndev/ndau/pkg/ndau/transactions.go"
)
TX_NAMES = {}
with tx_file.open() as f:
    for line in f:
        m = TXS.search(line)
        if m:
            txid, name = m.group("txid", "name")
            TX_NAMES[int(txid)] = name


class Transaction:
    def __init__(self, b64_data):
        data = b64decode(b64_data)
        self.raw = msgpack.loads(data)
        self.tx = self.raw[b"Transactable"]
        try:
            self.name = TX_NAMES[self.raw[b"TransactableID"]]
        except KeyError:
            self.name = "unknown"

    def __str__(self):
        return f"{self.name}: {pformat(self.tx)}".strip()


def get_net_url(netname):
    netname = netname.lower()
    nets = {"local", "dev", "test", "main"}
    if netname.endswith("net"):
        netname = netname[:-3]
    if netname not in nets:
        return netname

    if netname == "local":
        return "http://localhost:3030"

    netname += "net"

    resp = requests.get(SERVICES)
    resp.raise_for_status()

    return "https://" + resp.json()["networks"][netname]["nodes"][f"{netname}-0"]["api"]


def get_txs(url, block):
    resp = requests.get(urljoin(url, f"/block/height/{block}"))
    resp.raise_for_status()
    return [Transaction(data) for data in resp.json()["block"]["data"]["txs"]]


def main(args):
    url = get_net_url(args.net)
    for tx in get_txs(url, args.block):
        print(tx)


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser()

    parser.add_argument("block", type=int, help="block number from which to get txs")
    parser.add_argument("-n", "--net", default="main", help="net name or ndauapi URL")

    args = parser.parse_args()
    main(args)
