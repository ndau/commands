#!/usr/bin/env python3

from lib.args import get_node
import string
import subprocess
import sys
import time


def run_ssh_command(node_name, command):
    """
    Helper function for constructing an array to pass to subprocess.run() for an ssh command.
    """

    # The SSH key file must be installed here from Oneiro's 1password account.
    # If this script is run on Circle, "~" resolves to "/root"; see the deploy job in config.yml.
    pem_path = "~/.ssh/sc-node-ec2.pem"

    # Username used for logging into the AWS instance through SSH.
    ec2_user = "ec2-user"

    # We build the domain name of the node by dot-separating this after the node name.
    domain_name = "ndau.tech"

    # The alias we use for devnet is devnet.ndau.tech since every node is on the same server.
    # For testnet and mainnet we use {node_name}.ndau.tech.
    devnet_name = "devnet"
    if node_name.startswith(devnet_name):
        cname = devnet_name
    else:
        cname = node_name

    return subprocess.run(
        ["ssh", "-i", pem_path,
         "-o", "StrictHostKeyChecking=no",
         f"{ec2_user}@{cname}.{domain_name}", command],
        stdout=subprocess.PIPE,
    )


def test_ssh_access(node_name):
    """
    SSH into the given node's AWS instance and perform a docker command to test access.
    """

    print(f"Testing {node_name} SSH access...")
    r = run_ssh_command(node_name, f"docker container ls -f name={node_name}")
    if r.returncode != 0:
        sys.exit(f"ssh failed to test access with code {r.returncode}")

    output = r.stdout.decode("utf-8").rstrip("\n")

    print(output)


def get_container_id(node_name):
    """
    SSH into the given node's AWS instance to get the name of its container within.
    """

    print(f"Getting {node_name} container id...")
    r = run_ssh_command(node_name, f"docker container ls -q -f name={node_name}")
    if r.returncode != 0:
        sys.exit(f"ssh failed to get container id with code {r.returncode}")

    container_id = r.stdout.decode("utf-8").rstrip("\n")

    # Make sure we got back something that looks like a container id.
    if not all(c in string.hexdigits for c in container_id):
        sys.exit(f"Invalid container id: {container_id}")

    print(container_id)

    return container_id


def snapshot_node(node_name):
    """
    Cause the given node to take a snapshot and upload it to S3 and register it as the latest.
    Only works for nodes on AWS that have been configured with S3 creds, like testnet-5 and
    mainnet-5.  Otherwise, the snapshot will get generated inside the node's container, but
    not uploaded.
    """

    container_id = get_container_id(node_name)

    print("Creating snapshot...")
    r = run_ssh_command(
        node_name,
        f"docker exec {container_id} rm -f /image/snapshot_result; "
        f"docker exec {container_id} killall -HUP procmon",
    )
    if r.returncode != 0:
        sys.exit(f"ssh failed to create snapshot with code {r.returncode}")

    # It shouldn't take more than a few seconds to generate the snapshot.
    # We do this polling loop outside the container exec command.
    # Using an "until" loop inside the container was causing the ssh command to return error 126.
    print("Waiting for snapshot...")
    for i in range(60):
        r = run_ssh_command(
            node_name, f"docker exec {container_id} test -f /image/snapshot_result"
        )

        if r.returncode == 0:
            return True

        # If the test command failed, wait a second and try again.
        # Otherwise it's an unexpected error and we'll exit.
        if r.returncode != 1:
            sys.exit(f"ssh failed to detect snapshot with code {r.returncode}")

        time.sleep(1)

    return False


def main():
    """
    Cause a node on AWS to take a snapshot and upload it to S3 and register it as the latest.
    Must have ~/.ssh/sc-node-ec2.pem present on your system.
    """

    node_name = get_node()
    if not snapshot_node(node_name):
        sys.exit(f"An error occurred attempting to take a snapshot on {node_name}")


if __name__ == "__main__":
    main()
