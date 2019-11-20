#!/usr/bin/env python3

import os
import shlex
import socket
import subprocess
import sys
import tempfile
import time
from base64 import b64decode
from pathlib import Path

import requests

SCRIPT_DIR = Path(__file__).parent
assert SCRIPT_DIR.exists()

IMAGE_BASE_URL = "https://s3.amazonaws.com/ndau-images"
SERVICES_URL = "https://s3.us-east-2.amazonaws.com/ndau-json/services.json"
INTERNAL_P2P_PORT = 26660
INTERNAL_RPC_PORT = 26670
INTERNAL_API_PORT = 3030
INTERNAL_POSTGRES_PORT = 5432
GENERATED_GENESIS_SNAPSHOT = "*"
IDENTITY_ENV = "BASE64_NODE_IDENTITY"
ULI_ENV = "USE_LOCAL_IMAGE"


def bail(*args):
    for arg in args:
        print(arg)
    sys.exit(1)


def identity_default():
    "unpack the identity file from environment if present"
    try:
        data = os.environ[IDENTITY_ENV]
    except KeyError:
        return None
    return b64decode(data)


def identity_type(path):
    "load the identity file from a given path"
    if path == "":
        return None
    with open(path, "rb") as fp:
        return fp.read()


def run(cmd):
    rv = subprocess.run(shlex.split(cmd), stdout=subprocess.PIPE, check=True, text=True)
    return rv.stdout.strip()


def test_local_port(port):
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        # TODO: double-check the semantics here
        if s.connect_ex(("localhost", port)) == 0:
            bail(f"port {port} is already in use")


def test_peer(nature, ip, port):
    print(f"Testing {nature} connection @ {ip}:{port}...")
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.settimeout(5)
        if s.connect_ex((ip, int(port))) != 0:
            bail(f"could not reach peer at {ip}:{port}")


def get_peer_id(protocol, ip, port):
    url = f"{protocol}://{ip}:{port}/status"
    resp = requests.get(url, timeout=5)
    resp.raise_for_status()
    try:
        return resp.json()["result"]["node_info"]["id"]
    except KeyError:
        bail(f"could not get peer id from {url}")


def main(args):
    if args.network not in {"localnet", "devnet", "testnet", "mainnet"}:
        bail(
            f"Unsupported network: {args.network}",
            "Supported networks: localnet, devnet, testnet, mainnet",
        )

    # Validate container name (can't have slashes).
    if "/" in args.container:
        bail(f"Container name ({args.container}) cannot contain slashes")

    if run(f"docker container ls -a -q -f name='{args.container}')") != "":
        bail(
            f"Container already exists: {args.container}",
            "Use restartcontainer.sh to restart it, or ",
            "use removecontainer.sh to remove it first",
        )

    if args.snapshot is None or args.snapshot == "":
        if args.network == "localnet":
            args.snapshot = GENERATED_GENESIS_SNAPSHOT
        else:
            args.snapshot = ""

    if args.snapshot == "":
        args.snapshot = ""
        snaps = "(latest)"
    elif args.snapshot == GENERATED_GENESIS_SNAPSHOT:
        snaps = "(generated)"
    else:
        snaps = args.snapshot

    print(f"Net:       {args.network}")
    print(f"Container: {args.container}")
    print(f"P2P port:  {args.p2p}")
    print(f"RPC port:  {args.rpc}")
    print(f"API port:  {args.api}")
    print(f"PG port:   {args.pg}")
    print(f"Snapshot:  {snaps}")

    test_local_port(args.p2p)
    test_local_port(args.rpc)
    test_local_port(args.api)
    test_local_port(args.pg)

    # If no peers were given, we can get them automatically for non-localnet networks.
    # When running a localnet, the first peer can start w/o knowing any other peers.
    if args.network == "localnet":
        if args.peers_p2p is None:
            args.peers_p2p = []
        if args.peers_rpc is None:
            args.peers_rpc = []
    elif args.peers_p2p is None or args.peers_rpc is None:
        print(f"fetching {SERVICES_URL}...")
        sjr = requests.get(SERVICES_URL)
        sjr.raise_for_status()

        nodes = sjr.json()["networks"][args.network]["nodes"]
        if len(nodes) == 0:
            bail(f"no nodes published for {args.network}")

        if args.peers_p2p is None:
            args.peers_p2p = [node["p2p"] for node in nodes.values()]
        if args.peers_rpc is None:
            args.peers_rpc = [f"https://{node['rpc']}" for node in nodes.values()]

    if len(args.peers_p2p) != len(args.peers_rpc):
        bail("length of p2p and rpc peers must match")

    for p2p in args.peers_p2p:
        ip, port = p2p.split(":")
        test_peer("p2p", ip, port)

    persistent_peers = []
    for rpc in args.peers_rpc:
        protocol, ip, port = rpc.split(":")
        ip = ip.lstrip("/")
        test_peer("rpc", ip, port)
        persistent_peers.append(f"{get_peer_id(protocol, ip, port)}@{ip}:{port}")

    print("Persistent peers:")
    for peer in persistent_peers:
        print(f"  {peer}")

    # Stop the container if it's running.  We can't run or restart it otherwise.
    subprocess.run([f"{SCRIPT_DIR}/stopcontainer.sh", args.container], check=True)

    # If the image isn't present, fetch the "current" image from S3
    if args.network == "localnet" or args.use_local_image:
        image_name = "ndauimage:latest"
    else:
        images_dir = (SCRIPT_DIR / ".." / "ndau-images").resolve()
        images_dir.mkdir(parents=True, exist_ok=True)

        cur = f"current-{args.network}.txt"
        curp = images_dir / cur

        print(f"fetching {cur}...")
        resp = requests.get(f"{IMAGE_BASE_URL}/{cur}")
        resp.raise_for_status()
        curp.write_bytes(resp.content)
        sha = resp.text.strip()

        image_name = f"ndauimage:{sha}"

        if run(f"docker image ls -q {image_name}") == "":
            print(f"{image_name} not available locally; fetching...")
            image_fn = f"ndauimage-{sha}.docker"
            image_zip = f"{image_fn}.gz"
            image_path = images_dir / image_zip
            url = f"{IMAGE_BASE_URL}/{image_zip}"
            resp = requests.get(url)
            resp.raise_for_status()
            image_path.write_bytes(resp.content)

            print(f"loading {image_fn}...")
            subprocess.run(["gunzip", "-f", image_path], check=True)
            subprocess.run(["docker", "load", "-i", image_path.stem], check=True)

    print("creating container...")
    # Some notes about the params to the run command:
    # - Using --sysctl silences a warning about TCP backlog when redis runs.
    # - Set your own HONEYCOMB_* and SLACK_* env vars ahead of time
    cargs = [
        "docker",
        "create",
        "-p",
        f"{args.p2p}:{INTERNAL_P2P_PORT}",
        "-p",
        f"{args.rpc}:{INTERNAL_RPC_PORT}",
        "-p",
        f"{args.api}:{INTERNAL_API_PORT}",
        "-p",
        f"{args.pg}:{INTERNAL_POSTGRES_PORT}",
        "--name",
        args.container,
        "-e",
        f"NETWORK={args.network}",
        "-e",
        f"NODE_ID={args.container}",
        "-e",
        f"SNAPSHOT_NAME={args.snapshot}",
        "-e",
        f"PERSISTENT_PEERS={','.join(persistent_peers)}",
        "--sysctl",
        "net.core.somaxconn=511",
    ]

    for e in [
        "HONEYCOMB_DATASET",
        "HONEYCOMB_KEY",
        "SLACK_DEPLOYS_KEY",
        "AWS_ACCESS_KEY_ID",
        "SNAPSHOT_INTERVAL",
        "AWS_SECRET_ACCESS_KEY",
    ]:
        cargs.append("-e")
        cargs.append(f"{e}={os.environ.get(e, '')}")

    cargs.append(image_name)

    subprocess.run(cargs, check=True)

    # Copy the identity file into the container if one was specified
    identity_file = "node-identity.tgz"
    if args.identity is not None:
        with tempfile.NamedTemporaryFile() as tf:
            tf.write(args.identity)
            tf.flush()
            subprocess.run(
                ["docker", "cp", tf.name, f"{args.container}:/image/{identity_file}"],
                check=True,
            )

    # Copy the snapshot into the container if it exists as a local file.
    if args.snapshot != "" and Path(args.snapshot).exists():
        print(f"copying local {args.snapshot} into container...")
        # doesn't matter if the snapshot actually captures block 0; the container
        # will look for that snapshot, not others
        subprocess.run(
            [
                "docker",
                "cp",
                args.snapshot,
                f"{args.container}:/image/snapshot-{args.network}-0.tgz",
            ],
            check=True,
        )

    print("starting container...")
    subprocess.run(["docker", "start", args.container], check=True)

    # Run the hang monitor while we wait for the node to spin up.
    watcher = subprocess.Popen([f"{SCRIPT_DIR}/watchcontainer.sh", args.container])

    print("waiting for the node to fully spin up...")
    while (
        subprocess.run(
            ["docker", "exec", args.container, "test", "-f", "/image/running"],
            stdout=subprocess.DEVNULL,
        ).returncode
        != 0
    ):
        time.sleep(1)

    # Done waiting; kill the watcher.
    watcher.kill()

    print("Node is ready; dumping container logs...")
    for line in run(f"docker container logs {args.container}").splitlines():
        print(f"> {line}")

    # In the case no node identity was passed in, wait for it to generate one,
    # then copy it out.
    # It's important that node operators keep the node-identity.tgz file secure.
    if args.identity is None:
        # We can copy the file out now since we waited for the node to fully spin up
        nif = SCRIPT_DIR / f"node-identity-{args.container}.tgz"
        subprocess.run(
            ["docker", "cp", f"{args.container}:/image/{identity_file}", nif],
            check=True,
        )

        print(
            f"""
The node identity has been generated and copied out of the container here:
    {nif}

You can always get it at a later time by running the following:
    docker cp {args.container}:/image/{identity_file} {identity_file}
It can be used to restart this container with the same identity it has now
Keep it secret; keep it safe
"""
        )

    print("done")


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(
        epilog="""
Environment variables:
  BASE64_NODE_IDENTITY
             Alternate method to set the IDENTITY parameter.
             Lower priority than the argument.
             The contents of the variable are a base64 encoded tarball containing:

               - tendermint/config/priv_validator_key.json
               - tendermint/config/node_id.json
    """.strip()
    )

    parser.add_argument(
        "network", help="Which network to join: localnet, devnet, testnet, mainnet"
    )
    parser.add_argument("container", help="Name to give to the container to run")
    parser.add_argument(
        "--p2p",
        "--p2p-port",
        help="External port to map to the internal P2P port for the blockchain",
        default=INTERNAL_P2P_PORT,
        type=int,
    )
    parser.add_argument(
        "--rpc",
        "--rpc-port",
        help="External port to map to the internal RPC port for the blockchain",
        default=INTERNAL_RPC_PORT,
        type=int,
    )
    parser.add_argument(
        "--api",
        "--api-port",
        help="External port to map to the internal ndau API port",
        default=INTERNAL_API_PORT,
        type=int,
    )
    parser.add_argument(
        "--pg",
        "--pg-port",
        help="External port to map to the internal postgres port",
        default=INTERNAL_POSTGRES_PORT,
        type=int,
    )
    parser.add_argument(
        "--identity",
        help="""
node-identity.tgz file from a previous snaphot or initial container run.
If present, the node will use it to configure itself when [re]starting.
If missing, the node will generate a new identity for itself.
    """.strip(),
        type=identity_type,
        default=identity_default(),
    )
    parser.add_argument(
        "--snapshot",
        help=f"""
Name of the snapshot to use as a starting point for the node group.
If omitted, the latest $NETWORK snapshot will be used.
If it's a file, it will be used instead of pulling one from S3.
If it's '{GENERATED_GENESIS_SNAPSHOT}', genesis data is generated.
    """.strip(),
    )
    parser.add_argument(
        "--peers-p2p",
        help="""
Space-separated list of persistent peers on the network to join.
Each peer should be of the form IP_OR_DOMAIN_NAME:PORT.
If omitted, canonical peers will be fetched for non-localnet.
    """.strip(),
        nargs="+",
        default=None,
        type=list,
    )
    parser.add_argument(
        "--peers-rpc",
        help="""
Space-separated list of the same peers for RPC connections.
Each peer should be of the form PROTOCOL://IP_OR_DOMAIN_NAME:PORT.
If omitted, canonical peers will be fetched for non-localnet.
    """.strip(),
        nargs="+",
        default=None,
        type=list,
    )
    parser.add_argument(
        "--use-local-image",
        "--uli",
        action="store_true",
        default=os.environ.get(ULI_ENV, "") == "1",
        help="Use a locally-built image instead of the most recent official image.",
    )

    args = parser.parse_args()
    main(args)
