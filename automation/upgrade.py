#!/usr/bin/env python3

from get_health import get_health
from get_sha import get_sha
from get_synced import get_synced
from lib.args import get_net_node_sha
from lib.services import fetch_services, parse_services
from lib.networks import NETWORK_LOCATIONS
import json
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

    print(f"Waiting for {node_name} to reach steady state...")
    for attempt in range(300):
        time.sleep(1)
        if get_sha(api_url) != sha:
            continue
        time.sleep(1)
        if get_health(api_url) != "OK":
            continue
        time.sleep(1)
        if get_synced(rpc_url) != "YES":
            continue
        print(f"Upgrade of {node_name} is complete")
        return time.time() - time_started

    sys.exit(f"Timed out waiting for {node_name} upgrade to complete")
    return -1


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


def main():
    """
    Upgrade one or all nodes on the given network.
    """

    network, node_name, sha = get_net_node_sha()
    network_name = str(network)

    upgrade_nodes(network_name, node_name, sha)


if __name__ == '__main__':
    main()
