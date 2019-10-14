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


def get_sha(url):
    """
    Get the SHA of the node at the given API url.
    """

    response = fetch_url(f"{url}/version")

    try:
        return response.json()["NdauSha"]
    except:
        return "UNKNOWN"


def main():
    """
    Print the SHA for the node at the given API url.
    """

    url = get_url(UrlKind.API)
    sha = get_sha(url)
    print(sha)


if __name__ == "__main__":
    main()
