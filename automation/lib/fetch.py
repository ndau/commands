#!/usr/bin/env python3

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
