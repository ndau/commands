#! /usr/bin/env python3

import argparse
import base64
import getpass
import json
import textwrap
import time
import subprocess
import sys

import ndau

from yubihsm import YubiHsm
from yubihsm.objects import AsymmetricKey


def setupArgs():
    """ sets up the argument parser with all of our options """
    parser = argparse.ArgumentParser(
        formatter_class=argparse.RawDescriptionHelpFormatter,
        description=textwrap.dedent(
            """
            Reads a file of transactions and signs them.

            Generally, create a file of transaction data (format below). Leave
            off the signatures, but indicate which signatures each tx needs
            with the sigsneeded field.

            Then run this program once for each signer with a yubikey,
            specifying the initials and password for their yubikey, and feed
            each session the output from the previous session. The end result
            will be a json data file with all the transactions, signed, and
            decorated with the signable bytes as well.

            Finally, use this program with the --prevalidate and/or --submit
            flags to submit the tx to the blockchain.

            To use a yubikey, the yubihsm-connector must be running on port 12345.
            """
        ),
        epilog=textwrap.dedent(
            """
            The input format is a JSON array of transactions enclosed in [].
            Each individual transaction should look something like this:

            {
                "comment": "set up the an account with 5 ndau",
                "txtype": "transfer",
                "txbody": {
                    "source": "ndaekyty73hd56gynsswuj5q9em68tp6ed5v7tpft872hvuc",
                    "destination": "ndakfmgvxdm5rzenjabtt85mbma7aas2bhfzksqpegdu33v4",
                    "qty": 500000000,
                    "sequence": 24,
                    "signatures": []
                },
                "sigsneeded": [
                    "kq",
                    "cq"
                ]
            }

            The txbody is the actual transaction, and txtype is the name of the transaction.
            The sigsneeded field can contain any number of initials (something you will use
            in the --initials parameter) or private keys. If the latter, the tx will be signed
            with those keys. If the former, the tx will be signed with the Yubikey when those
            initials are specified.
            """  # noqa: E501
        ),
    )
    parser.add_argument(
        "--input",
        type=argparse.FileType("r"),
        default=sys.stdin,
        help="get the json input from this file (default stdin)",
    )
    parser.add_argument(
        "--output",
        type=argparse.FileType("w"),
        default=sys.stdout,
        help="send the json output to this file (default stdout)",
    )
    parser.add_argument(
        "--initials",
        default="",
        help="use a yubikey to sign transactions marked with these initials",
    )
    parser.add_argument(
        "--password",
        default="",
        help="use this password for the yubikey; prompted if not supplied",
    )
    parser.add_argument(
        "--keynum",
        type=int,
        default=0,
        help="use a yubikey to sign transactions marked with these initials",
    )
    parser.add_argument(
        "--skip",
        type=int,
        default=0,
        help="skip this many transactions (used for submit)",
    )
    parser.add_argument(
        "--network",
        default="test",
        help="specify network name (dev/local/main/test/full url)",
    )
    parser.add_argument(
        "--prevalidate",
        default=False,
        action="store_true",
        help="try to prevalidate the resulting tx on the network",
    )
    parser.add_argument(
        "--submit",
        default=False,
        action="store_true",
        help="try to submit the resulting tx on the network "
        "(if it prevalidated without error)",
    )
    return parser


class YubiSession(object):
    def __init__(self, pw, authkeyID=101, keynum=-1):
        # print(f"authkeyID = {authkeyID}")
        self.pw = pw
        self.authkeyID = authkeyID
        self.keynum = keynum
        self.connect()

    def connect(self):
        self.hsm = YubiHsm.connect("http://localhost:12345/connector/api")
        self.session = self.hsm.create_session_derived(self.authkeyID, self.pw)

    def close(self):
        self.session.close()
        self.hsm.close()

    def sign(self, bytes, keynum=-1):
        if keynum == -1:
            keynum = self.keynum
        key = AsymmetricKey(self.session, keynum)
        sig = key.sign_eddsa(bytes)
        encoded_sig = base64.standard_b64encode(sig).decode()
        return encoded_sig


def getSignableBytes(txtype, tx):
    txbytes = json.dumps(tx).encode()
    proc = subprocess.run(
        ["./ndau", "--json", "signable-bytes", "--strip", txtype],
        input=txbytes,
        capture_output=True,
    )
    return proc.stdout.decode("utf-8")


def getNdauSignature(signable):
    b64 = signable.encode()
    print(signable)
    proc = subprocess.run(
        ["./keytool", "ed", "raw", "signature", "--stdin", "-b"],
        input=b64,
        capture_output=True,
    )
    return proc.stdout.decode("utf-8").strip()


def ndauSign(key, signable):
    proc = subprocess.run(
        ["./keytool", "sign", key, signable, "-b"], capture_output=True
    )
    return proc.stdout.decode("utf-8").strip()


def prepTx(tx):
    try:
        # cmt = tx["comment"]
        txtype = tx["txtype"]
        body = tx["txbody"]
        sigs = tx["sigsneeded"]
    except ValueError as e:
        print(f"{tx} is missing field: {e}", file=sys.stderr)
        return None, None, None
    # print(f"---Working on {cmt}", file=sys.stderr)
    return txtype, body, sigs


if __name__ == "__main__":
    parser = setupArgs()
    args = parser.parse_args()

    # look up the network; if we don't find it, assume that the
    # network value is an API URL
    name = args.network
    if name in ndau.networks:
        node = ndau.networks[name]
    else:
        node = name

    password = args.password
    yubi = None
    if args.initials != "":
        if password == "":
            print(f"Enter yubikey password for '{args.initials}'.", file=sys.stderr)
            password = getpass.getpass()
            if password == "":
                print("password was blank!", file=sys.stderr)
                exit(1)
        yubi = YubiSession(password, keynum=args.keynum)

    txs = json.load(args.input)
    output_txs = []
    n = 0
    for t in txs:
        n += 1
        if n <= args.skip:
            continue
        txtype, body, sigs = prepTx(t)
        if txtype is None:
            print("skipping", file=sys.stderr)
            continue
        sb = getSignableBytes(txtype, body)
        t["signable_bytes"] = sb
        print(n, file=sys.stderr)
        for s in sigs:
            ndausig = None
            if s[:4] == "npvt":
                # it's a private key, so sign with it
                ndausig = ndauSign(s, sb)
            if yubi and s == args.initials:
                b = base64.b64decode(sb)
                sig = yubi.sign(b)
                ndausig = getNdauSignature(sig)
            if ndausig:
                if t["txtype"] == "setvalidation":
                    t["txbody"]["signature"] = ndausig
                elif ndausig not in body["signatures"]:
                    t["txbody"]["signatures"].append(ndausig)
        output_txs.append(t)

        if args.prevalidate or args.submit:
            print(f"prevalidating {t['txtype']}")
            resp = ndau.post(f"{node}/tx/prevalidate/{txtype}", json=t["txbody"])
            if resp.status_code != 200:
                print(
                    f"{t['txtype']} prevalidate failed with "
                    f"{resp.status_code}: {resp.text}"
                )
                exit(1)
            print(f"{resp.text}")
            print(f"submitting {t['txtype']}")
            if args.submit:
                resp = ndau.post(f"{node}/tx/submit/{txtype}", json=t["txbody"])
                if resp.status_code != 200:
                    print(
                        f"{t['txtype']} submit failed with "
                        f"{resp.status_code}: {resp.text}"
                    )
                    exit(1)
            print(f"{resp.text}")

            time.sleep(2)

    print(" done", file=sys.stderr)
    json.dump(output_txs, args.output)
    args.output.close()
