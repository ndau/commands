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


def get_health(url):
    """
    Get the health of the node at the given API url.
    """

    response = fetch_url(f"{url}/health")

    try:
        health = response.content.decode("utf-8").strip('"').rstrip('"\n')
        if len(health) != 0:
            return health
        # Blank health isn't really "healthy", but there was a response, so it's not "bad" either.
        return "ILL"
    except:
        return "BAD"


def main():
    """
    Print the health for the node at the given API url.
    """

    url = get_url(UrlKind.API)
    health = get_health(url)
    print(health)


if __name__ == "__main__":
    main()
