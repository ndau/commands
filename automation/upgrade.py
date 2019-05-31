#!/usr/bin/env python3

from get_health import get_health
from get_sha import get_sha
from get_sync import get_sync
from lib.args import get_net_node_sha
from lib.services import fetch_services, parse_services
from lib.networks import NETWORK_LOCATIONS
import json
import os
import subprocess
import sys
import time


# Number of seconds we wait between node upgrades.
# This helps stagger the daily restart tasks so that not all nodes restart near the same time.
MIN_WAIT_BETWEEN_NODES = 60

# Repository URI for our ndauimage Docker images.
ECR_URI = "578681496768.dkr.ecr.us-east-1.amazonaws.com/sc-node"


def upgrade_node(node_name, cluster, region, sha, api_url, rpc_url):
    """
    Upgrade the given node to the given SHA on the given cluster in the given region.
    Uses the urls to check its health before returning.
    Returns the amount of time that was spent waiting for the upgrade to complete after a restart.
    """

    print(f"Fetching latest {node_name} task definition...")
    r = subprocess.run(
        [
            "aws", "ecs", "describe-task-definition",
            "--region", region,
            "--task-definition", node_name,
        ],
        stdout=subprocess.PIPE,
    )
    if r.returncode != 0:
        sys.exit(f"aws ecs describe-task-definition failed with code {r.returncode}")
    
    try:
        task_definition_json = json.loads(r.stdout)
    except:
        task_definition_json = None
    if task_definition_json is None:
        sys.exit(f"Unable to load json")

    # Key names in json.
    task_definition_name = "taskDefinition"
    container_definitions_name = "containerDefinitions"
    image_name = "image"

    if task_definition_name not in task_definition_json:
        sys.exit(f"Cannot find {task_definition_name} in {task_definition_json}")
    task_definition_obj = task_definition_json[task_definition_name]

    if container_definitions_name not in task_definition_obj:
        sys.exit(f"Cannot find {container_definitions_name} in {task_definition_obj}")
    container_definitions_obj = task_definition_obj[container_definitions_name]

    for container_definition in container_definitions_obj:
        if image_name not in container_definition:
            sys.exit(f"Cannot find {image_name} in {container_definition}")
        container_definition[image_name] = f"{ECR_URI}:{sha}"

    print(f"Registering new {node_name} task definition...")
    r = subprocess.run([
        "aws", "ecs", "register-task-definition",
        "--region", region,
        "--family", node_name,
        "--container-definitions",
        json.dumps(container_definitions_obj),
    ])
    if r.returncode != 0:
        sys.exit(f"aws ecs register-task-definition failed with code {r.returncode}")

    print(f"Updating {node_name} service...")
    r = subprocess.run([
        "aws", "ecs", "update-service",
        "--cluster", cluster,
        "--region", region,
        "--service", node_name,
        "--task-definition", node_name,
    ])
    if r.returncode != 0:
        sys.exit(f"ecs-cli configure failed with code {r.returncode}")

    time_started = time.time()

    print(f"Waiting for {node_name} to catch up...")
    for attempt in range(300):
        # Wait some time between each status request, so we don't hammer the service.
        time.sleep(1)

        # Check the sha first since that's the one that'll fail the fastest, as the first few
        # attempts will still be polling the old service that's currently being restarted.
        if get_sha(api_url) != sha:
            continue

        time.sleep(1)

        # Once the sync (catch up) is complete, the upgraded node is happy with the network.
        if get_sync(rpc_url) != "COMPLETE":
            continue

        time.sleep(1)

        # Once all else looks good, check the health.  It'll likely be OK at this point since
        # an unhealthy node would certainly fail the sha and sync tests above.
        if get_health(api_url) != "OK":
            continue

        print(f"Upgrade of {node_name} is complete")
        return time.time() - time_started

    sys.exit(f"Timed out waiting for {node_name} upgrade to complete")


def upgrade_nodes(network_name, node_name, sha):
    """
    Upgrade the given node (or all nodes if node_name is None) on the given network.
    """

    if not network_name in NETWORK_LOCATIONS:
        sys.exit(f"Unknown locations for network {network_name} nodes")
    node_infos = NETWORK_LOCATIONS[network_name]

    apis, rpcs = parse_services(network_name, node_name, fetch_services())

    time_spent_waiting = -1
    for node_name in sorted(apis, reverse=True):
        api_url = apis[node_name]
        rpc_url = rpcs[node_name]

        if not node_name in node_infos:
            sys.exit(f"Unknown location for node {node_name} on network {network_name}")
        node_info = node_infos[node_name]

        cluster = node_info["cluster"]
        region = node_info["region"]

        if time_spent_waiting >= 0 and time_spent_waiting < MIN_WAIT_BETWEEN_NODES:
            wait_seconds = int(MIN_WAIT_BETWEEN_NODES - time_spent_waiting + 0.5)
            print(f"Waiting {wait_seconds} more seconds before upgrading {node_name}...")
            time.sleep(wait_seconds)

        time_spent_waiting = upgrade_node(node_name, cluster, region, sha, api_url, rpc_url)


def register_sha(network_name, sha):
    """
    Upload a new current-<network>.txt to S3 that points to the given SHA.
    This allows our local docker scripts to know which SHA to use when connecting to the network.
    """

    print(f"Registering {sha} as the current one in use on {network_name}...")

    current_file_name = f"current-{network_name}.txt"
    current_file_path = f"./{current_file_name}"

    with open(current_file_path, "w") as f:
        f.write(f"{sha}\n")

    r = subprocess.run([
        "aws", "s3", "cp",
        current_file_path, f"s3://ndau-images/{current_file_name}",
    ])

    os.remove(current_file_path)

    if r.returncode != 0:
        sys.exit(f"aws s3 cp failed with code {r.returncode}")


def main():
    """
    Upgrade one or all nodes on the given network.
    """

    start_time = time.time()

    network, node_name, sha = get_net_node_sha()
    network_name = str(network)

    upgrade_nodes(network_name, node_name, sha)

    # Auto-register the upgraded sha, even if only one node was upgraded.  The assumption is that
    # if we upgrade at least one node that we'll eventually upgrade all of them on the network.
    register_sha(network_name, sha)

    total_time = int(time.time() - start_time + 0.5)
    print(f"Total upgrade time: {total_time} seconds")


if __name__ == '__main__':
    main()
