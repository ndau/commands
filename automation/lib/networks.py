#!/usr/bin/env python3

from enum import Enum


class Network(Enum):
    """
    Supported network names.
    """

    devnet = 'devnet'
    testnet = 'testnet'
    mainnet = 'mainnet'

    def __str__(self):
        return self.value
