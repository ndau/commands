#! /usr/bin/env python3

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


class FooAction(argparse.Action):
    def __init__(self, option_strings, dest, nargs=None, **kwargs):
        if nargs is not None:
            raise ValueError("nargs not allowed")
        super(FooAction, self).__init__(option_strings, dest, **kwargs)

    def __call__(self, parser, namespace, values, option_string=None):
        print("%r %r %r" % (namespace, values, option_string))
        setattr(namespace, self.dest, values)


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
        "--transactions",
        nargs="*",
        default=[],
        help="Accepts a set of names or aliases to limit the set of transactions"
        ", or tags prefixed with #. See --names for a list.",
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
        action=FooAction,
        metavar="HEIGHT",
        help="range of blocks to search",
    )
    parser.add_argument(
        "--timespan",
        type=ndau.timestamp,
        nargs=2,
        metavar="TIMESTAMP",
        help="span of time to search (like '2019-07-18T12:34:56Z')",
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
        help="print the list of valid transaction names, aliases, and tags",
    )
    return parser


if __name__ == "__main__":
    parser = setupArgs()
    args = parser.parse_args()

    transactions = ndau.Transactions()

    print(args.blockrange)
    print(args.timespan)

    # if they just want a list of names, oblige them and quit
    if args.names:
        transactions.print()
        exit(1)

    # look up the network; if we don't find it, assume that the
    # network value is an API URL
    name = args.network
    if name in ndau.networks:
        node = ndau.networks[name]
    else:
        node = name

    # if there are no fields specified, use all of them
    if len(args.fields) == 0:
        outputfields = ndau.accountFields.keys()
    else:
        outputfields = [ndau.fieldnames()[f.lower()] for f in args.fields]

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

        if name.lower() not in ndau.fieldnames():
            print(f"no known field called '{name}'", file=sys.stderr)
            continue

        constraints.append(ndau.comparator(ndau.fieldnames()[name.lower()], op, value))

    # limit is the number of accounts in a single query -- this is limited by the
    # blockchain API and so we have to do a set of requests to get all the data
    limit = 100
    after = "-"
    output = []
    # we need the current time to evaluate "islocked"
    timeNow = datetime.datetime.now().isoformat("T")
    while after != "":
        qp = dict(limit=limit, after=after)
        result = ndau.getData(node, "/account/list", parms=qp)
        after = result["NextAfter"]

        accts = result["Accounts"]
        resp = ndau.post(f"{node}/account/accounts", json=result["Accounts"])

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
        f = ndau.fieldnames()[field.lower()]
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
