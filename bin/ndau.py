#! /usr/bin/env python3

#  ----- ---- --- -- -
#  Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
# 
#  Licensed under the Apache License 2.0 (the "License").  You may not use
#  this file except in compliance with the License.  You can obtain a copy
#  in the file LICENSE in the source distribution or at
#  https://www.apache.org/licenses/LICENSE-2.0.txt
#  - -- --- ---- -----

"""
    This file exists to support multiple python tools
    that work with the ndau blockchain.

"""

import collections
import datetime
import itertools
import random
import re
import requests
import time
import sys


def clamp(n, min, max):
    """ clamps a value between min and max """
    if n < min:
        n = min
    if n > max:
        n = max
    return n


# ----- query help -----------------
def getData(base, path, parms=None):
    """ this is a general-purpose query helper """
    u = base + path
    try:
        r = requests.get(u, timeout=3, params=parms)
        # print(r.url)
    except requests.Timeout:
        print(f"{time.asctime()}: Timeout fetching {u} {parms}")
        return {}
    except Exception as e:
        print(f"{time.asctime()}: Error {e} fetching {u} {parms}")
        return {}

    if r.status_code == requests.codes.ok:
        return r.json()
    print(f"{time.asctime()}: Error fetching {r.url}: ({r.status_code}) {r} {r.text}")
    return {}


def post(*args, **kwargs):
    """
     This lets us avoid including requests in our clients and also gives us a
     place to hang error handling if we want
    """
    return requests.post(*args, **kwargs)


# we cache block times in case of duplicates
blocktimes = {}


# fetch a block time from our base given a block number
def getBlockTime(base, blocknum):
    if blocknum in blocktimes:
        return blocktimes[blocknum]
    block = getData(base, f"/block/height/{blocknum}")
    t = block["block_meta"]["header"]["time"]
    blocktimes[blocknum] = t
    return t


# this fetches times for a list of block numbers
def cacheBlockTimes(base, blocknums):
    # we optimize the search by creating aggregating similar block numbers
    # into runs of no more than 100 blocks so we can fetch them in bulk
    ids = sorted(blocknums)
    runs = []
    run = None
    for i in ids:
        if run is None:
            run = [i, i]
            continue

        if run[0] < i - 99:
            runs.append(run)
            run = [i, i]

        run[1] = i

    for r in runs:
        block = getData(base, f"/block/range/{r[0]}/{r[1]}")
        for b in block["block_metas"]:
            h = b["header"]["height"]
            t = b["header"]["time"]
            blocktimes[h] = t


# All of the predefined network names. Note that you can also use the entire
# URL explicitly if # you have something other than one of these.
networks = {"local": "http://localhost:3030"}

try:
    r = requests.get(
        "https://s3.us-east-2.amazonaws.com/ndau-json/services.json", timeout=3
    )
    j = r.json()
    for n in j["networks"]:
        node = random.choice([i for i in j["networks"][n]["nodes"].values()])
        networks[n] = f"https://{node['api']}"
        if n.endswith("net"):
            networks[n[:-3]] = f"https://{node['api']}"
except requests.Timeout:
    print(f"{time.asctime()}: Timeout fetching services.json")
    # in case services.json doesn't work, give us a chance
    networks["main"] = "https://mainnet-0.ndau.tech:3030"
    networks["mainnet"] = "https://mainnet-0.ndau.tech:3030"
    networks["dev"] = "https://devnet.ndau.tech:3030"
    networks["devnet"] = "https://devnet.ndau.tech:3030"
    networks["test"] = "https://testnet-0.ndau.tech:3030"
    networks["testnet"] = "https://testnet-0.ndau.tech:3030"


# ----- field names for account fields

# This is a list of the main fieldnames in the account data structure, along with
# various aliases we've defined.
accountFields = {
    "balance": ("balance", "bal"),
    "validationKeys": ("validationKeys", "keys"),
    "validationScript": ("validationScript", "script"),
    "rewardsTarget": ("rewardsTarget", "rewards"),
    "incomingRewardsFrom": ("incomingRewardsFrom", "incoming"),
    "delegationNode": ("delegationNode", "delegation"),
    "islocked": ("islocked",),
    "lock": ("lock",),
    "lock.noticePeriod": ("lock.noticePeriod", "notice", "noticeperiod", "lockperiod"),
    "lock.unlocksOn": ("lock.unlocksOn", "unlock", "unlockson"),
    "lock.bonus": ("lock.bonus", "bonus"),
    "lastEAIUpdate": ("lastEAIUpdate", "lasteai"),
    "lastWAAUpdate": ("lastWAAUpdate", "lastwaa"),
    "weightedAverageAge": ("weightedAverageAge", "waa"),
    "sequence": ("sequence", "seq"),
    "stake_rules": ("stake_rules", "rules"),
    "costakers": ("costakers",),
    "holds": ("holds",),
    "hasrecourse": ("hasrecourse",),
    "recourseSettings": ("recourseSettings", "recourse"),
    "recourseSettings.period": ("recourseSettings.period", "recourseperiod"),
    "recourseSettings.changes_at": (
        "recourseSettings.changes_at",
        "recoursesettings.changesat",
        "changesat",
    ),
    "recourseSettings.next": ("recourseSettings.next", "recoursenext"),
    "currencySeatDate": ("currencySeatDate", "seat"),
    "parent": ("parent",),
    "progenitor": ("progenitor", "progen"),
    "id": ("id", "address"),
}


def accountNames():
    """ invert the accountfields list so we can use it to look up user input. """
    return dict(
        itertools.chain.from_iterable(
            ((alias.lower(), k.lower()) for alias in aliases)
            for k, aliases in accountFields.items()
        )
    )


def printAccountFieldHelp():
    """ Walk the accountfields dictionary and print a nicely formatted version """
    names = [(v[0], ", ".join(v[1:])) for v in accountFields.values()]
    print(f"{'name':28}  aliases")
    for n in names:
        print(f"{n[0]:28}  {n[1]}")


# ----- Transaction information


class Transactions(object):
    def __init__(self):
        self.transactions = []
        self._addTx("Transfer", 1, tags=["all", "account"])
        self._addTx("ChangeValidation", 2, tags=["all", "account"])
        self._addTx("ReleaseFromEndowment", 3, aliases=["rfe"], tags=["all", "system"])
        self._addTx("ChangeRecoursePeriod", 4, tags=["all", "account"])
        self._addTx("Delegate", 5, tags=["all", "account"])
        self._addTx("CreditEAI", 6, tags=["all", "nodeops"])
        self._addTx("Lock", 7, tags=["all", "account", "lock"])
        self._addTx("Notify", 8, tags=["all", "account", "lock"])
        self._addTx(
            "SetRewardsDestination", 9, aliases=["srd"], tags=["all", "account"]
        )
        self._addTx("SetValidation", 10, tags=["all", "account"])
        self._addTx("Stake", 11, tags=["all", "staking"])
        self._addTx(
            "RegisterNode", 12, aliases=["register", "reg"], tags=["all", "nodeops"]
        )
        self._addTx("NominateNodeReward", 13, aliases=["nnr"], tags=["all", "system"])
        self._addTx("ClaimNodeReward", 14, tags=["all", "nodeops"])
        self._addTx("TransferAndLock", 15, tags=["all", "account", "lock"])
        self._addTx(
            "CommandValidatorChange", 16, aliases=["cvc"], tags=["all", "system"]
        )
        self._addTx(
            "UnregisterNode",
            18,
            aliases=["unregister", "unreg"],
            tags=["all", "nodeops"],
        )
        self._addTx("Unstake", 19, tags=["all", "staking"])
        self._addTx("Issue", 20, tags=["all", "system", "price"])
        self._addTx("CreateChildAccount", 21, aliases=["cca"], tags=["all", "account"])
        self._addTx("RecordPrice", 22, tags=["all", "system", "price"])
        self._addTx("SetSysvar", 23, tags=["all", "system"])
        self._addTx("SetStakeRules", 24, aliases=["ssr"], tags=["all", "staking"])
        self._addTx(
            "RecordEndowmentNAV",
            25,
            aliases=["ren", "renav", "nav"],
            tags=["all", "system", "price"],
        )
        self._addTx("ResolveStake", 26, tags=["all", "staking"])
        self._addTx("ChangeSchema", 30, tags=["all", "system"])

    def _addTx(self, name, index, aliases=[], tags=[]):
        tx = {
            "name": name,
            "index": index,
            "aliases": aliases,
            "tags": tags,
            "_lookup": set([name.lower()]),
        }
        tx["_lookup"].update(set(aliases))
        tx["_lookup"].update(set(["." + t for t in tags]))
        self.transactions.append(tx)

    def getMatchIndexes(self, text):
        p = re.compile("[a-z.]+")
        sa = set(p.findall(text))
        results = set()
        for tx in self.transactions:
            if len(sa.intersection(tx["_lookup"])) > 0:
                results.add(tx["index"])
        return results

    def getTxFor(self, index):
        for tx in self.transactions:
            if index == tx["index"]:
                return tx
        return None

    def getTxsByName(self, names):
        r = []
        for n in names:
            for tx in self.transactions:
                if n.lower() in tx["_lookup"]:
                    r.append(tx["name"])
        return r

    def print(self):
        tags = collections.defaultdict(list)
        print("Transactions            Aliases               Tags")
        print("------------            -------               ----")
        for tx in self.transactions:
            print(
                f"{tx['name']:22}  {', '.join(tx['aliases']):20}"
                f"  {', '.join(['.'+t for t in tx['tags']])}"
            )
            for t in tx["tags"]:
                tags[t].append(tx["name"])


# ----- Duration processing ------------------------------------------

# these are the names of time units and their duration in microseconds
timeConstants = {
    "years": 365 * 24 * 60 * 60 * 1_000_000,
    "months": 30 * 24 * 60 * 60 * 1_000_000,
    "days": 24 * 60 * 60 * 1_000_000,
    "hours": 60 * 60 * 1_000_000,
    "minutes": 60 * 1_000_000,
    "seconds": 1_000_000,
    "micros": 1,
}


def usec(x, unit):
    """ helper function to look up time units and do the multiplication """
    if x is None:
        return 0
    t = int(x)
    return timeConstants[unit] * t


# This is exactly the regexp from the go code that parses durations
# https://github.com/ndau/ndaumath/blob/de5a90c45d3f079f1a263819493d2f7a70bb4b8b/pkg/constants/constants.go#L61
# see https://github.com/ndau/ndau/issues/405 and change
# the 9 to 6 below when that is fixed
durpat = re.compile(
    r"(?i)^"  # case-insensitive
    r"(?P<neg>-)?"  # durations can be negative
    r"p?"  # and may contain a leading p
    r"((?P<years>\d+)y)?"
    r"((?P<months>\d{1,2})m)?"
    r"((?P<days>\d{1,2})d)?"
    r"(t"
    r"((?P<hours>\d{1,2})h)?"
    r"((?P<minutes>\d{1,2})m)?"
    r"((?P<seconds>\d{1,2})s)?"
    r"((?P<micros>\d{1,9})[μu]s?)?"  # 9 until bug is fixed
    r")?$"
)


def isDuration(s):
    """
    see if a string can be evaluated as a duration
    """
    return durpat.match(s) is not None


def parseDuration(s):
    """
    parse a duration and return an integer number of microseconds
    """

    m = durpat.match(s)

    t = 0
    # the usec function deals with missing parts of the expression
    for u in timeConstants:
        t += usec(m.group(u), u)
    return t


def timestamp(s):
    """ validates and canonicalizes a timestamp """
    RFC3339 = "%Y-%m-%dT%H:%M:%S"
    RFC3339f = "%Y-%m-%dT%H:%M:%S.%f"
    formats = [
        "%Y-%m-%d",
        "%Y-%m-%dT%H:%M",
        RFC3339,
        RFC3339f,
        RFC3339 + "Z",
        RFC3339f + "Z",
    ]
    # remove extra digits from the time if there are any
    # python can only handle microseconds, not nanoseconds
    s = re.sub(r"(\.[0-9]{1,6})([0-9]+)", r"\1", s)
    ts = None
    for f in formats:
        try:
            ts = datetime.datetime.strptime(s, f)
            break
        except ValueError:
            continue

    if ts is None:
        print(f"{ts} is not a valid time format", file=sys.stderr)
        raise ValueError
    return ts.strftime(RFC3339 + "Z")


# ----- general-purpose functions


def comparator(field, op, value):
    """
    Returns a comparator function that binds the parameters so they can
    later be used to compare the value against the data in an account.

    This is aware of various ndau types and will coerce field and value
    so that they are correctly comparable.
    """

    def cmp(x):
        # if the field doesn't exist, it's None
        f = x.get(field, None)

        # we might have to special-case the value depending on
        # what the type of f is. Note that we don't worry about
        # timestamps because they're directly comparable as strings,
        # but we do need to convert things that look like
        # durations.

        v = value
        vl = str(value).lower()

        if isinstance(f, str):
            # if f is a duration, also parse the value as a duration
            if isDuration(f):
                f = parseDuration(f)
                v = parseDuration(vl)
        elif isinstance(f, bool):
            # for bools, we convert value strings into booleans
            if vl == "true" or vl == "t":
                v = True
            else:
                v = False
        elif isinstance(f, int):
            # for ints we convert the value into integers so we can
            # compare numerically (we never deal with floats)
            v = int(value)

        # anything that looks like None is None.
        if vl == "null" or vl == "none" or vl == "nil":
            v = None

        # special case empty field comparisons, because
        # None is special and can't be ordered.
        if f is None or v is None:
            if op == "==" or op == "=":
                return f == v
            elif op == "!=":
                return f != v
            else:
                return False

        # finally we can actually just do the comparison
        if op == "==" or op == "=":
            return f == v
        elif op == "!=":
            return f != v
        elif op == ">=":
            return f >= v
        elif op == "<=":
            return f <= v
        elif op == "<":
            return f < v
        elif op == ">":
            return f > v
        elif op == "%":
            return re.search(v, f)

    return cmp


def flatten(value, prefix=""):
    """ returns a flattened structure with dotted field names """
    result = dict()
    for k, v in value.items():
        if isinstance(v, dict):
            result.update(flatten(v, k.lower() + "."))
        else:
            result[prefix + k.lower()] = v
    return result
