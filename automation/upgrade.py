#!/usr/bin/env python3

from get_catchup import get_catchup
from get_health import get_health
from get_sha import get_sha
from lib.args import get_net_node_sha_snapshot
from lib.services import fetch_services, parse_services
from lib.networks import NETWORK_LOCATIONS
from snapshot_node import get_container_id
from snapshot_node import snapshot_node
import json
import os
import subprocess
import sys
import time


# Number of seconds we wait between node upgrades.
# This helps stagger the daily restart tasks so that not all nodes restart near the same time.
# Some of this time is used by a node's service restarting, before procmon starts.
MIN_WAIT_BETWEEN_NODES = 120

# Repository URI for our ndauimage Docker images.
ECR_URI = "578681496768.dkr.ecr.us-east-1.amazonaws.com/sc-node"


def fetch_container_definitions(node_name, region):
    """
    Fetch the json object (list) representing the given node's container definitions (there
    should only be one) in the given region.
    Also returns the corresponding task definition arn.
    """

    r = subprocess.run(
        [
            "aws",
            "ecs",
            "describe-task-definition",
            "--region",
            region,
            "--task-definition",
            node_name,
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
        sys.exit(f"Unable to load json: {r.stdout}")

    # Key names in json.
    task_definition_name = "taskDefinition"
    container_definitions_name = "containerDefinitions"
    arn_name = "taskDefinitionArn"

    if task_definition_name not in task_definition_json:
        sys.exit(f"Cannot find {task_definition_name} in {task_definition_json}")
    task_definition_obj = task_definition_json[task_definition_name]

    if container_definitions_name not in task_definition_obj:
        sys.exit(f"Cannot find {container_definitions_name} in {task_definition_obj}")
    container_definitions = task_definition_obj[container_definitions_name]

    if arn_name not in task_definition_obj:
        sys.exit(f"Cannot find {arn_name} in {task_definition_obj}")
    task_definition_arn = task_definition_obj[arn_name]

    return container_definitions, task_definition_arn


def register_task_definition(node_name, region, container_definitions):
    """
    Register an updated version of the latest task definition for the given node in the given
    region using the given container definitions (typically a list of length one).
    """

    r = subprocess.run(
        [
            "aws",
            "ecs",
            "register-task-definition",
            "--region",
            region,
            "--family",
            node_name,
            "--container-definitions",
            json.dumps(container_definitions),
        ],
        stdout=subprocess.PIPE,
    )
    if r.returncode != 0:
        sys.exit(f"aws ecs register-task-definition failed with code {r.returncode}")

    # Print the useful-for-debugging json ourselves so we can collapse it all on one line.
    try:
        task_definition_json = json.loads(r.stdout)
    except:
        task_definition_json = None
    if not task_definition_json is None:
        print(json.dumps(task_definition_json, separators=(",", ":")))


def update_service(node_name, region, cluster):
    """
    Update the given node (cause it to restart with the latest task definition) on the given
    cluster in the given region.
    """

    r = subprocess.run(
        [
            "aws",
            "ecs",
            "update-service",
            "--cluster",
            cluster,
            "--region",
            region,
            "--service",
            node_name,
            "--task-definition",
            node_name,
        ],
        stdout=subprocess.PIPE,
    )
    if r.returncode != 0:
        sys.exit(f"ecs-cli configure failed with code {r.returncode}")

    # Print the useful-for-debugging json ourselves so we can collapse it all on one line.
    try:
        service_json = json.loads(r.stdout)
    except:
        service_json = None
    if not service_json is None:
        print(json.dumps(service_json, separators=(",", ":")))


def is_service_running(node_name, region, cluster, task_definition_arn):
    """
    Return whether the given service is currently running with the given task definition on AWS.
    """

    r = subprocess.run(
        [
            "aws",
            "ecs",
            "describe-services",
            "--cluster",
            cluster,
            "--region",
            region,
            "--services",
            node_name,
        ],
        stdout=subprocess.PIPE,
    )
    if r.returncode != 0:
        sys.exit(f"aws ecs describe-services failed with code {r.returncode}")

    try:
        services_json = json.loads(r.stdout)
    except:
        services_json = None
    if services_json is None:
        sys.exit(f"Unable to load json: {r.stdout}")

    # Key names in json.
    services_name = "services"
    service_name = "serviceName"
    task_definition_name = "taskDefinition"
    running_count_name = "runningCount"

    if services_name in services_json:
        services = services_json[services_name]
        for service in services:
            if service_name in service and service[service_name] == node_name:
                # Service was found; return whether it's currently running with the
                # desired task definition revision.
                return (
                    task_definition_name in service and
                    service[task_definition_name] == task_definition_arn and
                    running_count_name in service and
                    service[running_count_name] > 0
                )

    # The service wasn't found and so is not running.
    return False


def wait_for_service(node_name, region, cluster, sha, api_url, rpc_url, task_definition_arn):
    """
    Wait for a node's service to become healthy and fully caught up on its network.
    Uses the urls to check its health before returning.
    """

    # Wait forever.  When doing an upgrade with full reindex, each node can take a long time to
    # catch up.  The higher the network blockchain height, the longer it'll take.  It's unbounded.
    while True:
        # Wait some time between each status request, so we don't hammer the service.
        time.sleep(1)

        # Make sure we're not polling the old service that still might be draining.
        if is_service_running(node_name, region, cluster, task_definition_arn):
            break

    # Do the remaining waiting in a separate loop.  We don't need to continually check whether
    # the new service is running.  We did that above and the tests below continue to imply that.
    while True:
        time.sleep(1)

        # Once the catch up is complete, the upgraded node is happy with the network.
        if get_catchup(rpc_url) != "COMPLETE":
            continue

        time.sleep(1)

        # Once all else looks good, check the health.  It'll likely be OK at this point since
        # an unhealthy node would certainly fail the catch up test above.
        if get_health(api_url) != "OK":
            continue

        print(f"Upgrade of {node_name} is complete")
        return

    # Will never happen (but leaving it here in case we ever do impose a max wait time).
    sys.exit(f"Timed out waiting for {node_name} upgrade to complete")


def set_snapshot(snapshot, container_definition):
    """
    Set the given snapshot name into the appropriate environment variable in the given
    container definition.
    """

    # Key names in json.
    environment_name = "environment"
    key_name = "name"
    value_name = "value"
    snapshot_key = "SNAPSHOT_NAME"

    if environment_name not in container_definition:
        sys.exit(f"Cannot find {environment_name} in {container_definition}")
    environment_variables = container_definition[environment_name]

    found = False
    for environment_variable in environment_variables:
        if (
            key_name in environment_variable
            and environment_variable[key_name] == snapshot_key
        ):
            environment_variable[value_name] = snapshot
            found = (
                True
            )  # We could break, but letting the loop run handles (unlikely) dupes.
    if not found:
        environment_variable = {key_name: snapshot_key, value_name: snapshot}
        environment_variables.append(environment_variable)


def upgrade_node(node_name, region, cluster, sha, snapshot, api_url, rpc_url):
    """
    Upgrade the given node to the given SHA on the given cluster in the given region using the
    given snapshot name ("" means "latest snapshot") from which to catch up.
    Uses the urls to check its health before returning.
    Returns the amount of time that was spent waiting for the upgrade to complete after a restart.
    Also returns the container id of the node if a snapshot is to be taken after the upgrade.
    """

    container_id = None
    if len(snapshot) > 0:
        # Make sure we can SSH into the node to take a snapshot before we do anything.
        container_id = get_container_id(node_name)

    print(f"Fetching latest {node_name} task definition...")
    container_definitions, task_definition_arn = fetch_container_definitions(node_name, region)

    # Key names in json.
    image_name = "image"

    for container_definition in container_definitions:
        if image_name not in container_definition:
            sys.exit(f"Cannot find {image_name} in {container_definition}")
        container_definition[image_name] = f"{ECR_URI}:{sha}"

        # Set the specified snapshot to use.
        # If no snapshot was specified, this will ensure that the "latest snapshot" is still set.
        set_snapshot(snapshot, container_definition)

    print(f"Registering new {node_name} task definition...")
    register_task_definition(node_name, region, container_definitions)

    print(f"Updating {node_name} service...")
    update_service(node_name, region, cluster)

    # Record the time of the restart so we make sure to wait at least MIN_WAIT_BETWEEN_NODES.
    # NOTE: It would be better to detect the old service going down first.  When we support
    # config changes (e.g. environment variable updates without a new sha), we'll need this as
    # well as improved wait-for-catchup logic below, since the sha, catchup and health won't be
    # expected to change during such an upgrade.
    time_started = time.time()

    print(f"Waiting for {node_name} to restart and catch up...")
    wait_for_service(node_name, region, cluster, sha, api_url, rpc_url, task_definition_arn)

    return time.time() - time_started, container_id


def upgrade_nodes(network_name, node_name, sha, snapshot):
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
            print(
                f"Waiting {wait_seconds} more seconds before upgrading {node_name}..."
            )
            time.sleep(wait_seconds)

        time_spent_waiting, container_id = upgrade_node(
            node_name, region, cluster, sha, snapshot, api_url, rpc_url
        )

        # If we just upgraded a node with a snapshot, the node has now caught up and regenerated
        # all its data from that snapshot.  Have it generate a new snapshot and make it the new
        # latest snapshot, so that all remaining nodes can catch up from that and save time.
        if len(snapshot) > 0:
            if not snapshot_node(node_name, container_id):
                sys.exit(f"Unable to take a snapshot on {node_name}")

            # All remaining nodes can upgrade using the latest snapshot.
            snapshot = ""

            # Re-deploy the node that just deployed, so that it uses the latest snapshot.
            # That way, if the node goes down for any reason, AWS will restart it and not
            # have to catch up from the original snapshot like it just did.
            print(f"Redeploying {node_name} at the latest snapshot...")
            time_spent_waiting, container_id = upgrade_node(
                node_name, region, cluster, sha, snapshot, api_url, rpc_url
            )


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

    r = subprocess.run(
        ["aws", "s3", "cp", current_file_path, f"s3://ndau-images/{current_file_name}"]
    )

    os.remove(current_file_path)

    if r.returncode != 0:
        sys.exit(f"aws s3 cp failed with code {r.returncode}")


def main():
    """
    Upgrade one or all nodes on the given network.
    """

    network, node_name, sha, snapshot = get_net_node_sha_snapshot()
    network_name = str(network)

    # If no snapshot was given, use the latest.
    if snapshot is None:
        snapshot = ""

    # Be extra careful with mainnet.
    if network_name == "mainnet":
        if node_name is None:
            node_text = "ALL NODES"
        else:
            node_text = node_name
        print()
        print(
            f"You are about to UPGRADE {node_text} ON MAINNET to the following SHA: {sha}"
        )
        print(
            "Please be sure that this SHA has been staged and tested on testnet first."
        )
        print()
        confirm = input(
            f"Proceed with upgrading {node_text} on mainnet now? (type yes to confirm) "
        )
        if confirm != "yes":
            sys.exit("Mainnet upgrade declined")

    start_time = time.time()

    upgrade_nodes(network_name, node_name, sha, snapshot)

    # Auto-register the upgraded sha, even if only one node was upgraded.  The assumption is that
    # if we upgrade at least one node that we'll eventually upgrade all of them on the network.
    register_sha(network_name, sha)

    total_time = int(time.time() - start_time + 0.5)
    print(f"Total upgrade time: {total_time} seconds")


if __name__ == "__main__":
    main()
