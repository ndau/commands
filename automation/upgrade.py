#!/usr/bin/env python3

from lib.args import get_net_node_sha
from lib.services import fetch_services, parse_services
from lib.networks import NETWORK_LOCATIONS
import json
import subprocess
import sys
import time


# Number of seconds we wait between node upgrades.
# This helps stagger the daily restart tasks so that not all nodes restart near the same time.
WAIT_BETWEEN_NODES = 60

# Repository URI for our ndauimage Docker images.
ECR_URI = "578681496768.dkr.ecr.us-east-1.amazonaws.com/sc-node"


def upgrade_node(node_name, cluster, region, sha, url):
    """
    Upgrade the given node to the given SHA on the given cluster in the given region.
    Uses the url to check its health before returning.
    """

    print("Fetching latest task definition...")
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
    
    task_definition_json = json.loads(r.stdout)

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

    print("Registering new task definition...")
    r = subprocess.run([
        "aws", "ecs", "register-task-definition",
        "--region", region,
        "--family", node_name,
        "--container-definitions",
        json.dumps(container_definitions_obj),
    ])
    if r.returncode != 0:
        sys.exit(f"aws ecs register-task-definition failed with code {r.returncode}")

    print("Updating service...")
    r = subprocess.run([
        "aws", "ecs", "update-service",
        "--cluster", cluster,
        "--region", region,
        "--service", node_name,
        "--task-definition", node_name,
    ])
    if r.returncode != 0:
        sys.exit(f"ecs-cli configure failed with code {r.returncode}")


def upgrade_nodes(network_name, node_name, sha):
    """
    Upgrade the given node (or all nodes if node_name is None) on the given network.
    """

    if not network_name in NETWORK_LOCATIONS:
        sys.exit(f"Unknown locations for network {network_name} nodes")
    node_infos = NETWORK_LOCATIONS[network_name]

    apis, rpcs = parse_services(network_name, node_name, fetch_services())

    upgraded_nodes = 0
    for node_name in sorted(apis, reverse=True):
        url = apis[node_name]

        if not node_name in node_infos:
            sys.exit(f"Unknown location for node {node_name} on network {network_name}")
        node_info = node_infos[node_name]

        cluster = node_info["cluster"]
        region = node_info["region"]

        if upgraded_nodes > 0:
            time.sleep(WAIT_BETWEEN_NODES)

        upgrade_node(node_name, cluster, region, sha, url)

        upgraded_nodes += 1


def main():
    """
    Upgrade one or all nodes on the given network.
    """

    network, node_name, sha = get_net_node_sha()
    network_name = str(network)

    upgrade_nodes(network_name, node_name, sha)


if __name__ == '__main__':
    main()
