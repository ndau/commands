#! /usr/bin/env python3

# system imports
#  ----- ---- --- -- -
#  Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
# 
#  Licensed under the Apache License 2.0 (the "License").  You may not use
#  this file except in compliance with the License.  You can obtain a copy
#  in the file LICENSE in the source distribution or at
#  https://www.apache.org/licenses/LICENSE-2.0.txt
#  - -- --- ---- -----

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
        transactions it finds there. By default, it reads all transactions and returns
        a csv file containing the full details from all of them.
        However, it also supports the ability to generate JSON output, to
        select a subset of fields for each transaction, and to select a subset of
        transactions according to the values of their fields or their metadata.

         If multiple constraints are applied they must all be satisfied.

         Note that compared to the ndau API, this application:
         * flattens the transaction data so that there are no nested structures
         * injects some additional fields:

    """
        ),
        epilog=textwrap.dedent(
            """
        Examples:
            # show the block heights for all Issue and RFE transactions
            bin/forAllTransactions.py --network testnet --txtypes RFE Issue --fields blockheight txtype

            # count the number of transfer transactions between blocks 20,000 and 30,000
            bin/forAllTransactions.py --network testnet --txtypes Transfer --blockrange 20000 30000 --count

            # show the timestamps, blockheight, and quantity of all transfers between blocks 20,000 and 30,000
            bin/forAllTransactions.py --network testnet --txtypes Transfer --fields blockheight txdata.qty blocktime --blockrange 20000 30000

            # show the height and type of all price-related transactions
            bin/forAllTransactions.py --network testnet --txtypes .price --fields blockheight txtype

            # show the source accounts for all transfer transactions that paid SIB (and their fees)
            bin/forAllTransactions.py --network testnet --txtypes Transfer  --fields txdata.source fee sib --constraint "sib>0"

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
        "--txtypes",
        nargs="*",
        default=[],
        help="Accepts a set of names or aliases to limit the set of transaction types"
        ", or tags prefixed with '.'. See --names for a list.",
    )
    parser.add_argument(
        "--fields",
        default=[],
        nargs="*",
        help="fields to send to the output (default all)",
    )
    parser.add_argument(
        "--blockrange",
        type=int,
        nargs=2,
        metavar="HEIGHT",
        help="range of blocks to search",
    )
    parser.add_argument(
        "--timespan",
        nargs=2,
        metavar="TIMESTAMP",
        help="start and end times of the span of time to search (like "
        "'2019-07-18T12:34:56Z' , just a date like 2019-08-01, "
        "or a word like 'first' or 'now')",
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
        "--sum",
        dest="format",
        action="store_const",
        const="sum",
        help="print only the sum of the specified field (no other output)",
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
        help="print the list of valid transaction names, aliases, and tags",
    )
    return parser


if __name__ == "__main__":
    parser = setupArgs()
    args = parser.parse_args()

    transactions = ndau.Transactions()

    # if they just want a list of names, oblige them and quit
    if args.names:
        transactions.print()
        exit(0)

    # look up the network; if we don't find it, assume that the
    # network value is an API URL (but if it doesn't start with http, fail)
    name = args.network
    if name in ndau.networks:
        node = ndau.networks[name]
    else:
        if not name.startswith("http"):
            print("network name must start with http or https", file=sys.stderr)
            exit(1)
        node = name

    genesis = "2019-05-11T00:00:00Z"
    timeNow = datetime.datetime.now(datetime.UTC).isoformat("T") + "Z"
    times = {
        "start": genesis,
        "first": genesis,
        "genesis": genesis,
        "now": timeNow,
        "last": timeNow,
        "end": timeNow,
    }

    # parse all the constraints, building the comparators we will use to
    # evaluate them as we walk through the data
    constraints = []

    # we treat timespan and blockrange as constraints until the API gets smarter
    if args.timespan:
        st, et = [t.lower() for t in args.timespan]
        if st in times:
            startts = times[st]
        else:
            startts = ndau.timestamp(st)
        if et in times:
            endts = times[et]
        else:
            endts = ndau.timestamp(et)

        if startts > endts:
            startts, endts = endts, startts
        constraints.append(ndau.comparator("blocktime", ">=", startts))
        constraints.append(ndau.comparator("blocktime", "<=", endts))

    if args.blockrange:
        lastblock = ndau.getData(node, "/block/current")
        lastheight = lastblock["block_meta"]["header"]["height"]
        sb, eb = args.blockrange
        sb = ndau.clamp(sb, 1, lastheight)
        eb = ndau.clamp(eb, 1, lastheight)
        if sb > eb:
            sb, eb = eb, sb
        constraints.append(ndau.comparator("blockheight", ">=", sb))
        constraints.append(ndau.comparator("blockheight", "<=", eb))

    for c in args.constraints:
        pat = re.compile("([a-zA-Z0-9._-]+) *(>=|<=|>|<|==|=|!=|%) *([^ ]+)")
        m = pat.match(c)
        if not m:
            print(f"couldn't parse constraint '{c}'", file=sys.stderr)
            continue
        name, op, value = m.groups()

        constraints.append(ndau.comparator(name.lower(), op, value))

    # limit is the number of accounts in a single query -- this is limited by the
    # blockchain API and so we have to do a set of requests to get all the data
    limit = 100
    output = []
    types = transactions.getTxsByName(args.txtypes)
    nexthash = "start"
    # we need the current time to evaluate "islocked"
    while nexthash != "":
        url = f"/transaction/before/{nexthash}"
        qp = dict(limit=limit, type=types)
        resp = ndau.getData(node, url, parms=qp)
        if not resp:
            print("Failure fetching data.", file=sys.stderr)
            exit(1)

        nexthash = resp["NextTxHash"]
        data = resp["Txs"]
        if data is None:
            break

        # ok, now we can iterate through the batch of data, flatten it,
        # and evaluate constraints
        for i in range(len(data)):
            # now flatten it
            flat = ndau.flatten(data[i])

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

    # if they just want a count, we don't need to sort and filter fields,
    # and max is meaningless.
    if args.format == "count":
        print(f"{len(output)}", file=args.output)
        exit(0)
    
    if args.format == "sum":
        sum = 0
        for i in output:
            sum += i[args.sum]
        print(f"{sum}", file=args.output)
        exit(0)

    # if there are no fields specified, use all of the fields specified
    # in the unified collection of txs
    if len(args.fields) == 0:
        fields = set()
        for o in output:
            fields.update(set([k.lower() for k in o.keys()]))
        outputfields = sorted(
            list(fields), key=lambda x: "zzz" + x if x.startswith("txdata") else x
        )
    else:
        outputfields = [f.lower() for f in args.fields]

    # 1) sort the result
    if args.sort != "":
        field = args.sort
        reverse = False
        if field[0] == "/":
            field = field[1:]
            reverse = True
        f = field.lower()
        output = sorted(output, key=lambda x: x[f], reverse=reverse)

    # 2) truncate the result set if desired
    if args.max > 0 and args.max < len(output):
        output = output[: args.max]

    # 3) filter the set of fields
    result = []
    for o in output:
        i = dict([(k, v) for k, v in o.items() if k in outputfields])
        result.append(i)

    if args.format == "json":
        j = json.dump(result, args.output)
    elif args.format == "csv":
        w = csv.DictWriter(args.output, outputfields)
        w.writeheader()
        for i in result:
            w.writerow(i)
