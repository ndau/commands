#!/usr/bin/env python3

#  ----- ---- --- -- -
#  Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
# 
#  Licensed under the Apache License 2.0 (the "License").  You may not use
#  this file except in compliance with the License.  You can obtain a copy
#  in the file LICENSE in the source distribution or at
#  https://www.apache.org/licenses/LICENSE-2.0.txt
#  - -- --- ---- -----

import requests


def fetch_url(url):
    """
    Return requests.get(url).  Return None if it times out.
    """

    try:
        response = requests.get(url, timeout=3)
    except KeyboardInterrupt:
        raise
    except:
        response = None

    return response
