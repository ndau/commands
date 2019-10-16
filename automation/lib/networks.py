#!/usr/bin/env python3

#  ----- ---- --- -- -
#  Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
# 
#  Licensed under the Apache License 2.0 (the "License").  You may not use
#  this file except in compliance with the License.  You can obtain a copy
#  in the file LICENSE in the source distribution or at
#  https://www.apache.org/licenses/LICENSE-2.0.txt
#  - -- --- ---- -----

from enum import Enum


class Network(Enum):
    """
    Supported network names.
    """

    devnet = "devnet"
    testnet = "testnet"
    mainnet = "mainnet"

    def __str__(self):
        return self.value


NETWORK_LOCATIONS = {
    "devnet": {
        "devnet-0": {"cluster": "devnet", "region": "us-west-1"},
        "devnet-1": {"cluster": "devnet", "region": "us-west-1"},
        "devnet-2": {"cluster": "devnet", "region": "us-west-1"},
        "devnet-3": {"cluster": "devnet", "region": "us-west-1"},
        "devnet-4": {"cluster": "devnet", "region": "us-west-1"},
    },
    "testnet": {
        "testnet-0": {"cluster": "testnet-0", "region": "us-east-1"},
        "testnet-1": {"cluster": "testnet-1", "region": "us-east-2"},
        "testnet-2": {"cluster": "testnet-2", "region": "us-west-1"},
        "testnet-3": {"cluster": "testnet-3", "region": "us-west-2"},
        "testnet-4": {"cluster": "testnet-4", "region": "ap-southeast-1"},
        "testnet-backup": {"cluster": "testnet-backup", "region": "us-east-2"},
    },
    "mainnet": {
        "mainnet-0": {"cluster": "mainnet-0", "region": "us-east-1"},
        "mainnet-1": {"cluster": "mainnet-1", "region": "us-east-2"},
        "mainnet-2": {"cluster": "mainnet-2", "region": "us-west-1"},
        "mainnet-3": {"cluster": "mainnet-3", "region": "us-west-2"},
        "mainnet-4": {"cluster": "mainnet-4", "region": "ap-southeast-1"},
        "mainnet-backup": {"cluster": "mainnet-backup", "region": "us-east-2"},
    },
}
