#!/usr/bin/env python3

from lib.args import get_url, UrlKind
from lib.fetch import fetch_url
import json


def get_version(url):
    """
    Get the version of the node at the given API url.
    """

    # Key names in response json.
    version_name = "NdauVersion"

    response = fetch_url(f"{url}/version")

    if not response is None:
        try:
            version_obj = json.loads(response.content)
        except:
            version_obj = None
        if not version_obj is None and version_name in version_obj:
            return version_obj[version_name]

    return "UNKNOWN"


def main():
    """
    Print the version for the node at the given API url.
    """

    url = get_url(UrlKind.API)
    version = get_version(url)
    print(version)


if __name__ == "__main__":
    main()
