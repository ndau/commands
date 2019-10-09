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


def get_catchup(url):
    """
    Get the current catch up status of the node at the given RPC url.
    """

    response = fetch_url(f"{url}/status")

    try:
        sync_info = response.json()["result"]["sync_info"]
        if sync_info["catching_up"] or int(sync_info["latest_block_height"]) == 0:
            return "CATCHINGUP"
        return "COMPLETE"
    except:
        return "UNKNOWN"


def main():
    """
    Print the current catch up status of the node at the given RPC url.
    """

    url = get_url(UrlKind.RPC)
    catchup = get_catchup(url)
    print(catchup)


if __name__ == "__main__":
    main()
