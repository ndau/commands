#! /usr/bin/env python3

import argparse
import json
import requests
import itertools
import textwrap
import time
import re
import sys
import csv


def getData(base, path, parms=None):
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


names = {
    "local": "http://localhost:3030",
    "main": "https://mainnet-0.ndau.tech:3030",
    "mainnet": "https://mainnet-0.ndau.tech:3030",
    "dev": "https://devnet.ndau.tech:3030",
    "devnet": "https://devnet.ndau.tech:3030",
    "test": "https://testnet-0.ndau.tech:3030",
    "testnet": "https://testnet-0.ndau.tech:3030",
}

allFields = {
    "balance": ("balance", "bal"),
    "validationKeys": ("validationKeys", "keys"),
    "validationScript": ("validationScript", "script"),
    "rewardsTarget": ("rewardsTarget", "rewards"),
    "incomingRewardsFrom": ("incomingRewardsFrom", "incoming"),
    "delegationNode": ("delegationNode", "delegation"),
    "haslock": ("haslock",),
    "lock": ("lock",),
    "lock.noticePeriod": ("lock.noticePeriod", "notice", "noticeperiod"),
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

# invert the allfields list into something we can look up
fieldnames = dict(
    [
        x
        for x in itertools.chain(
            *[[(s.lower(), k) for s in v] for k, v in allFields.items()]
        )
    ]
)


def printFieldNames():
    names = [(v[0], ", ".join(v[1:])) for v in allFields.values()]
    print(f"{'name':28}  aliases")
    for n in names:
        print(f"{n[0]:28}  {n[1]}")


def comparator(field, op, value):
    def cmp(x):
        f = x.get(field, None)
        v = value
        if isinstance(f, bool):
            if value.lower() == "true" or value.lower() == "t":
                v = True
            else:
                v = False
        elif isinstance(f, int):
            v = int(value)

        if value == "null" or value == "None" or value == "nil":
            v = None

        # special case empty fields
        if f is None or v is None:
            if op == "==" or op == "=":
                return f == v
            elif op == "!=":
                return f != v
            else:
                return False

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
            return re.match(v, f)

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


def setupArgs():
    """ sets up the argument parser with all of our options """
    parser = argparse.ArgumentParser(
        formatter_class=argparse.RawDescriptionHelpFormatter,
        description=textwrap.dedent(
            """
         This program reads an ndau blockchain and returns information from the
         accounts it finds there. By default, it reads all accounts and returns
         a large blob of JSON containing the full details from all of the
         accounts. However, it also supports the ability to generate CSV files,
         to select a subset of fields for each account, and to select a subset
         of accounts according to the values of their fields.

         Note that compared to the ndau API, this application:
         * flattens the account data so that there are no nested structures
         * injects three additional fields:
             * id (the ndau account address)
             * haslock (true if the lock value is non-empty)
             * hasrecourse (true if recourseSettings is non-empty)

    """
        ),
        epilog=textwrap.dedent(
            """
        Examples:
            # print the number of accounts with more than 1000 ndau that are unlocked
            forAllAccounts.py --network=test --count --constraints "balance>=100000000000" "haslock == false"

            # print the account IDs and balances of accounts with a balance of less than 10000 napu
            forAllAccounts.py --network=test --csv --constraints "balance<10000" --fields id balance

            # print the account IDs and balances of the top 10 largest accounts
            forAllAccounts.py --network=test --csv  --fields id balance --sort /bal --max 10

            # print the 3 largest accounts that are delegated
            forAllAccounts.py --network=test --csv --constraint "delegation=ndam75fnjn7cdues7ivi7ccfq8f534quieaccqibrvuzhqxa"  --fields id  --sort /bal --max 3

            # count the number of accounts that are locked with the maximum lock bonus of 5%
            forAllAccounts.py --network=test --count --fields id bonus --constraints "haslock==true" "bonus=50000000000"

    """  # noqa: E501
        ),
    )
    parser.add_argument(
        "--network",
        default="dev",
        help="specify network name (dev/local/main/test/full url)",
    )
    parser.add_argument(
        "--constraints",
        nargs="*",
        default=[],
        help="each constraint should be 'fieldname op value', "
        "where op is one of '%%' '>' '<' '==' '!=' '>=' or '<='. "
        "The %% is for pattern match, where patterns are python regular expressions. "
        "Use quotes around each individual constraint.",
    )
    parser.add_argument(
        "--fields",
        default=[],
        nargs="*",
        help="fields to send to the output (default all)",
    )
    parser.add_argument(
        "--sort",
        default="",
        help="field used to sort the result; use '/field' to sort descending",
    )
    parser.add_argument(
        "--max",
        type=int,
        default=0,
        help="return at most this many results (default no limit)",
    )
    parser.add_argument(
        "--output",
        type=argparse.FileType("w"),
        default=sys.stdout,
        help="send the output to this file (default stdout)",
    )
    parser.add_argument(
        "--json",
        dest="format",
        action="store_const",
        const="json",
        default="json",
        help="emit the output as JSON (default)",
    )
    parser.add_argument(
        "--csv",
        dest="format",
        action="store_const",
        const="csv",
        help="emit the output as CSV",
    )
    parser.add_argument(
        "--count",
        dest="format",
        action="store_const",
        const="count",
        help="print the number of matching results and exit (no other output)",
    )
    parser.add_argument(
        "--once",
        default=False,
        action="store_true",
        help="do a single query for 100 accounts and exit (speeds up testing)",
    )
    parser.add_argument(
        "--names",
        default=False,
        action="store_true",
        help="print the list of valid field names and exit",
    )
    return parser


if __name__ == "__main__":
    parser = setupArgs()
    args = parser.parse_args()

    if args.names:
        printFieldNames()
        exit(1)

    name = args.network
    if name in names:
        node = names[name]
    else:
        node = name

    if len(args.fields) == 0:
        outputfields = allFields.keys()
    else:
        outputfields = [fieldnames[f.lower()] for f in args.fields]

    constraints = []
    for c in args.constraints:
        pat = re.compile("([a-zA-Z0-9._-]+) *(>=|<=|>|<|==|=|!=|%) *([^ ]+)")
        m = pat.match(c)
        if not m:
            print(f"couldn't parse constraint '{c}'", file=sys.stderr)
            continue
        name, op, value = m.groups()

        if name.lower() not in fieldnames:
            print(f"no known field called '{name}'", file=sys.stderr)
            continue

        constraints.append(comparator(fieldnames[name.lower()], op, value))

    limit = 100
    after = "-"
    output = []
    while after != "":
        qp = dict(limit=limit, after=after)
        result = getData(node, "/account/list", parms=qp)
        after = result["NextAfter"]

        accts = result["Accounts"]
        resp = requests.post(f"{node}/account/accounts", json=result["Accounts"])

        data = resp.json()
        for k in data:
            # add some manufactured fields to the account data
            data[k]["id"] = k
            data[k]["haslock"] = data[k]["lock"] is not None
            data[k]["hasrecourse"] = data[k]["recourseSettings"] is not None
            # now flatten it
            flat = flatten(data[k])
            keep = True
            for c in constraints:
                # try:
                if not c(flat):
                    keep = False
            if keep:
                output.append(flat)

        if args.once:
            break

    # sort the result
    if args.sort != "":
        field = args.sort
        reverse = False
        if field[0] == "/":
            field = field[1:]
            reverse = True
        f = fieldnames[field.lower()]
        output = sorted(output, key=lambda x: x[f], reverse=reverse)

    # truncate the result set if desired
    if args.max > 0 and args.max < len(output):
        output = output[: args.max]

    # filter the fields
    result = []
    for o in output:
        i = dict([(k, v) for k, v in o.items() if k in outputfields])
        result.append(i)

    if args.format == "count":
        print(f"{len(result)}", file=args.output)
    elif args.format == "json":
        j = json.dump(result, args.output)
    elif args.format == "csv":
        w = csv.DictWriter(args.output, outputfields)
        w.writeheader()
        for i in result:
            w.writerow(i)
