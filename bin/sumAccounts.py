#! /usr/bin/env python3

#  ----- ---- --- -- -
#  Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
# 
#  Licensed under the Apache License 2.0 (the "License").  You may not use
#  this file except in compliance with the License.  You can obtain a copy
#  in the file LICENSE in the source distribution or at
#  https://www.apache.org/licenses/LICENSE-2.0.txt
#  - -- --- ---- -----

import requests
import time
import sys


def getData(base, path, parms=None):
    u = base + path
    try:
        r = requests.get(u, timeout=3, params=parms)
    except requests.Timeout:
        print(f"{time.asctime()}: Timeout in {u}")
        return {}
    except Exception as e:
        print(f"{time.asctime()}: Error {e} in {u}")
        return {}
    if r.status_code == requests.codes.ok:
        return r.json()
    print(f"{time.asctime()}: Error in {u}: ({r.status_code}) {r}")
    return {}


names = {
    "local": "http://localhost:3030",
    "main": "https://node-0.main.ndau.tech",
    "mainnet": "https://node-0.main.ndau.tech",
    "dev": "https://devnet-0.api.ndau.tech",
    "devnet": "https://devnet-0.api.ndau.tech",
    "test": "https://testnet-0.api.ndau.tech",
    "testnet": "https://testnet-0.api.ndau.tech",
}

if __name__ == "__main__":
    name = "dev"
    if len(sys.argv) > 1:
        name = sys.argv[1]

    node = names[name]

    limit = 100
    after = "-"
    balances = []
    while after != "":
        qp = dict(limit=limit, after=after)
        result = getData(node, "/account/list", parms=qp)
        after = result["NextAfter"]

        accts = result["Accounts"]
        resp = requests.post(f"{node}/account/accounts", json=result["Accounts"])

        data = resp.json()
        for k in data:
#            balances.append((k, data[k]["balance"] / 100_000_000))
            balances.append((k, data[k]["balance"]))

    total = sum([b for k, b in balances])
#    print(f"total in {len(balances)} accounts is {total}")
    print(f"total in {len(balances)} accounts is {total / 100_000_000}")

    s = sorted(balances, key=lambda a: a[1])
    for t in s:
        print(f"{t[0]}: {t[1]:16.8f}")
