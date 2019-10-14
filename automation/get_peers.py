#!/usr/bin/env python3

#  ----- ---- --- -- -
#  Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
# 
#  Licensed under the Apache License 2.0 (the "License").  You may not use
#  this file except in compliance with the License.  You can obtain a copy
#  in the file LICENSE in the source distribution or at
#  https://www.apache.org/licenses/LICENSE-2.0.txt
#  - -- --- ---- -----

from lib.args import get_url, UrlKind
from lib.fetch import fetch_url


def get_peers(url):
    """
    Get the peer count of the node at the given RPC url.
    """

    response = fetch_url(f"{url}/net_info")

    try:
        return response.json()["result"]["n_peers"]
    except:
        # Return an invalid peer count to signal failure.
        return -1


def main():
    """
    Print the peer count for the node at the given RPC url.
    """

    url = get_url(UrlKind.RPC)
    peers = get_peers(url)
    print(peers)


if __name__ == "__main__":
    main()
