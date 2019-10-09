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


def get_height(url):
    """
    Get the current height of the node at the given API url.
    """

    response = fetch_url(f"{url}/block/current")

    try:
        return response.json()["block_meta"]["header"]["height"]
    except:
        # Return an invalid height to signal failure.
        return 0


def main():
    """
    Print the height for the node at the given API url.
    """

    url = get_url(UrlKind.API)
    height = get_height(url)
    print(height)


if __name__ == "__main__":
    main()
