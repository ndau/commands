#! /usr/bin/env python3

"""
    This file exists to support multiple python tools
    that work with the ndau blockchain.

"""

import collections
import datetime
import itertools
import re
import requests
import time


# ----- query help -----------------
def getData(base, path, parms=None):
    """ this is a general-purpose query helper """
    u = base + path
    try:
        r = requests.get(u, timeout=3, params=parms)
    except requests.Timeout:
        print(f"{time.asctime()}: Timeout in {u}")
        return {}
    except Exception as e:
        print(f"{time.asctime()}: Error {e} in {u}")
        return {}
    if r.status_code == requests.codes.ok:
        return r.json()
    print(f"{time.asctime()}: Error in {u}: ({r.status_code}) {r} {r.text}")
    return {}


def post(*args, **kwargs):
    """
     This lets us avoid including requests in our clients and also gives us a
     place to hang error handling if we want
    """
    return requests.post(*args, **kwargs)


# All of the predefined network names. Note that you can also use the entire
# URL explicitly if # you have something other than one of these.
networks = {
    "local": "http://localhost:3030",
    "main": "https://mainnet-0.ndau.tech:3030",
    "mainnet": "https://mainnet-0.ndau.tech:3030",
    "dev": "https://devnet.ndau.tech:3030",
    "devnet": "https://devnet.ndau.tech:3030",
    "test": "https://testnet-0.ndau.tech:3030",
    "testnet": "https://testnet-0.ndau.tech:3030",
}

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
            ((alias.lower(), k) for alias in aliases)
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
        tx["_lookup"].update(set(["#" + t for t in tags]))
        self.transactions.append(tx)

    def getMatchIndexes(self, text):
        p = re.compile("[a-z#]+")
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

    def print(self):
        tags = collections.defaultdict(list)
        print("Transactions            Aliases               Tags")
        print("------------            -------               ----")
        for tx in self.transactions:
            print(
                f"{tx['name']:22}  {', '.join(tx['aliases']):20}"
                f"  {', '.join(['#'+t for t in tx['tags']])}"
            )
            for t in tx["tags"]:
                tags[t].append(tx["name"])


# ----- Duration processing ------------------------------------------

# these are the names of time units and their duration in microseconds
timeConstants = {
    "year": 365 * 24 * 60 * 60 * 1_000_000,
    "month": 30 * 24 * 60 * 60 * 1_000_000,
    "day": 24 * 60 * 60 * 1_000_000,
    "hour": 60 * 60 * 1_000_000,
    "min": 60 * 1_000_000,
    "sec": 1_000_000,
    "usec": 1,
}


def usec(x, unit):
    """ helper function to look up time units and do the multiplication """
    if x is None:
        return 0
    t = int(x)
    return timeConstants[unit] * t


def parseDuration(s):
    """
    parse a duration and return an integer number of microseconds
    """

    # this uses a tagged regexp to parse strings that we believe are already
    # durations (this isn't perfect because the t is optional and without it
    # the m is ambiguous).
    p = re.compile(
        "("
        "((?P<year>[0-9]+)y)?"
        "((?P<month>[0-9]+)m)?"
        "((?P<day>[0-9]+)d)?)?"
        "t?"
        "(((?P<hour>[0-9]+)h)?"
        "((?P<min>[0-9]+)m)?"
        "((?P<sec>[0-9]+)s)?"
        "((?P<usec>[0-9]+)us?"
        ")?"
        ")?"
    )
    m = p.match(s)

    t = 0
    # the usec function deals with missing parts of the expression
    for u in timeConstants:
        t += usec(m.group(u), u)
    return t


def timestamp(s):
    """ validates and canonicalizes a timestamp """
    RFC3339 = "%Y-%m-%dT%H:%M:%S.%f"
    formats = [
        "%Y-%m-%d",
        "%Y-%m-%dT%H:%M",
        "%Y-%m-%dT%H:%M:%S",
        RFC3339,
        RFC3339 + "Z",
    ]
    ts = None
    for f in formats:
        try:
            ts = datetime.datetime.strptime(s, f)
            break
        except ValueError:
            continue

        if ts is None:
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

    durpat = re.compile(
        "(([0-9+][ymd])*(t((h|m|s|us)[0-9]+)+)+)|"
        "(([0-9+][ymd])+(t((h|m|s|us)[0-9]+)+)*)"
    )

    def cmp(x):
        # if the field doesn't exist, it's None
        f = x.get(field, None)

        # we might have to special-case the value depending on
        # what the type of f is. Note that we don't worry about
        # timestamps because they're directly comparable as strings,
        # but we do need to convert things that look like
        # durations.

        v = value
        vl = value.lower()

        if isinstance(f, str):
            m = durpat.match(f)
            # if f is a duration, also parse the value as a duration
            if m:
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
            result.update(flatten(v, k + "."))
        else:
            result[prefix + k] = v
    return result
