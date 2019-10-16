#! /usr/bin/env python3

#  ----- ---- --- -- -
#  Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
#
#  Licensed under the Apache License 2.0 (the "License").  You may not use
#  this file except in compliance with the License.  You can obtain a copy
#  in the file LICENSE in the source distribution or at
#  https://www.apache.org/licenses/LICENSE-2.0.txt
#  - -- --- ---- -----

# system imports
import argparse
import csv
import datetime
import json
import re
import sys
import textwrap

# local import (ndau.py must be found locally)
import ndau


def setupArgs():
    """ sets up the argument parser with all of our options """
    parser = argparse.ArgumentParser(
        formatter_class=argparse.RawDescriptionHelpFormatter,
        description=textwrap.dedent(
            """
        This program reads an ndau blockchain and returns information from the
        accounts it finds there. By default, it reads all accounts and returns
        a csv file containing the full details from all of the accounts.
        However, it also supports the ability to generate JSON output, to
        select a subset of fields for each account, and to select a subset of
        accounts according to the values of their fields.

         If multiple constraints are applied they must all be satisfied.

         Note that compared to the ndau API, this application:
         * flattens the account data so that there are no nested structures
         * injects three additional fields:
             * id (the ndau account address)
             * islocked (true if the lock exists and has not yet expired)
             * hasrecourse (true if recourseSettings is non-empty)

    """
        ),
        epilog=textwrap.dedent(
            """
        Examples:
            # count the number of accounts with more than 1000 ndau that are unlocked
            forAllAccounts.py --network=test --count --constraints "balance>=100000000000" "islocked == false"

            # print the account IDs and balances of accounts with a balance of less than 10000 napu
            forAllAccounts.py --network=test --constraints "balance<10000" --fields id balance

            # print the account IDs and balances of the top 10 largest accounts as JSON
            forAllAccounts.py --network=test --json --fields id balance --sort /bal --max 10

            # print the 3 largest accounts that are delegated to a specific node that ends with a given string
            forAllAccounts.py --network=test --constraint "delegation%vuzhqxa"  --fields id  --sort /bal --max 3

            # count the number of accounts that are locked with the maximum lock bonus of 5%
            forAllAccounts.py --network=test --count --constraints "islocked==true" "bonus=50000000000"

            # count the number of accounts that are locked for less than one year
            forAllAccounts.py --network=test --count --constraints "islocked==true" "notice<1y"

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
        "--csv",
        dest="format",
        action="store_const",
        const="csv",
        default="csv",
        help="emit the output as CSV (default)",
    )
    parser.add_argument(
        "--json",
        dest="format",
        action="store_const",
        const="json",
        help="emit the output as JSON",
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

    # if they just want a list of names, oblige them and quit
    if args.names:
        ndau.printAccountFieldHelp()
        exit(1)

    # look up the network; if we don't find it, assume that the
    # network value is an API URL
    name = args.network
    if name in ndau.networks:
        node = ndau.networks[name]
    else:
        if not name.startswith("http"):
            print("network name must start with http or https", file=sys.stderr)
            exit(1)
        node = name

    # if there are no fields specified, use all of them
    if len(args.fields) == 0:
        outputfields = ndau.accountFields.keys()
    else:
        outputfields = [ndau.accountNames()[f.lower()] for f in args.fields]

    # force all to lowercase
    outputfields = [f.lower() for f in outputfields]

    # parse all the constraints, building the comparators we will use to
    # evaluate them as we walk through the data
    constraints = []
    for c in args.constraints:
        pat = re.compile("([a-zA-Z0-9._-]+) *(>=|<=|>|<|==|=|!=|%) *([^ ]+)")
        m = pat.match(c)
        if not m:
            print(f"couldn't parse constraint '{c}'", file=sys.stderr)
            continue
        name, op, value = m.groups()

        if name.lower() not in ndau.accountNames():
            print(f"no known field called '{name}'", file=sys.stderr)
            continue

        constraints.append(
            ndau.comparator(ndau.accountNames()[name.lower()], op, value)
        )

    # limit is the number of accounts in a single query -- this is limited by the
    # blockchain API and so we have to do a set of requests to get all the data
    limit = 100
    after = "-"
    output = []
    # we need the current time to evaluate "islocked"
    timeNow = datetime.datetime.utcnow().isoformat("T")
    while after != "":
        qp = dict(limit=limit, after=after)
        result = ndau.getData(node, "/account/list", parms=qp)
        after = result["NextAfter"]

        accts = result["Accounts"]
        failcount = 0
        while failcount < 5:
            resp = ndau.post(f"{node}/account/accounts", json=result["Accounts"])
            if resp.status_code == 200:
                break
            else:
                print(f"Error from {resp.url}: {resp.text}", file=sys.stderr)
                failcount += 1

        data = resp.json()
        # ok, now we can iterate through the batch of data
        for k in data:
            # add some manufactured fields to the account data
            data[k]["id"] = k
            # we're unlocked if there's no lock object, OR if
            # the current time is after the "unlocksOn" time.
            unlocked = data[k].get("lock") is None or (
                data[k]["lock"].get("unlocksOn") is not None
                and data[k]["lock"]["unlocksOn"] < timeNow
            )
            data[k]["islocked"] = not unlocked
            data[k]["hasrecourse"] = data[k]["recourseSettings"] is not None
            # now flatten it
            flat = ndau.flatten(data[k])

            # by default we keep it but if it fails any constraint we drop it
            keep = True
            for c in constraints:
                # try:
                if not c(flat):
                    keep = False
            if keep:
                output.append(flat)

        # this can make testing faster because we only do one batch of data
        if args.once:
            break

    # These next steps are carefully ordered because we want to sort before truncation
    # and we might want to sort on data that's not included in the output set

    # 1) sort the result
    if args.sort != "":
        field = args.sort
        reverse = False
        if field[0] == "/":
            field = field[1:]
            reverse = True
        f = ndau.accountNames()[field.lower()]
        output = sorted(output, key=lambda x: x[f], reverse=reverse)

    # 2) truncate the result set if desired
    if args.max > 0 and args.max < len(output):
        output = output[: args.max]

    # 3) filter the set of fields
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
