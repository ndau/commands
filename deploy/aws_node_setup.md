# Mainnet Node AWS Setup

## Overview

Steps for setting up a mainnet node on a given region.

## Key

* `<N>` - zero-based node number on mainnet
* `<XXX>` - IP field = `100 + <N>`

## Steps

In all cases, leave default settings unless specified below.

1. Log into the AWS Management Console
1. Choose a Region in the upper right for where you want to set up the new node
1. EC2 > Key Pairs
    - Import Key Pair
    - Name it `sc-node-ec2-mainnet`
    - Public key contents:
    ```
    ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC55zKlUU5P+iUVR++59SoPM3PKjSvVnA06swqdLc3UiNK7dun6crh3VT/8O66lOil/+LcsfYDbFeKkXRl8LYqcs/BrCZMVa0exJrcb/iUOlGKgmxkQYx0+x1+WdtEIdn/5RozdYZC7lmOMvpPD/Sg4OeqF6/kM/crdKWEYWbPEZmVFzZeSCh9ln0cqTceMCPx7NwaZki3k3ldy9rmeY6hkBa0QaqZ38aosgQJ9NNs/ls6O9WbXnhCgKP9km6GlYBkIcTBBD1za0qAzUN3s3v3ldcGSrkwwF76gLcGjoQTzmpnI+bP0u/ByJaqgZ0u6oOpDuRShUHRb7wPbA6Vyk1QH
    ```
    - Grab the `sc-node-ec2.pem` file from 1password for later use if you want to shell into the node instance.  Use the Connect button on the EC2 > Instances page to get the ssh command to use.  You'll want to remove the `-mainnet` portion of the `.pem` file name in the command, and use `ec2-user` instead of `root`.
1. VPC > Your VPCs
    - Create VPC
        - Name tag: `mainnet-<N>`
        - IPv4 CIDR block: `<XXX>.0.0.0/16`
        - Create
1. VPC > Security Groups
    - Find the "default" Security Group in the list associated with the new VPC
    - Click the pencil next to the Name field and give it the name: `mainnet-<N>`
    - Actions > Edit inbound rules
    - Leave the "All traffic" rule alone
    - Add Rule: `Custom TCP - TCP - 26660 - 0.0.0.0/0 - P2P`
    - Add Rule: `Custom TCP - TCP - 26670 - 0.0.0.0/0 - RPC`
    - Add Rule: `Custom TCP - TCP - 3030 -- 0.0.0.0/0 - API`
    - Add Rule: `SSH -------- TCP - 22 ---- 0.0.0.0/0 - SSH`
    - Save rules
1. VPC > Subnets
    - Create subnet
    - Name tag: `mainnet-<N>`
    - VPC: `mainnet-<N>`
    - Availability Zone: `No preference`
    - IPv4 CIDR block: `<XXX>.0.0.0/24`
    - Create
1. VPC > Internet Gateways
    - Create internet gateway
    - Name tg: `mainnet-<N>`
    - Create
    - Select it in the list
    - Actions > Attach to VPC
    - VPC: `mainnet-<N>`
    - Attach
1. VPC > Route Tables
    - Find the route table in the list for the new VPC; give it the name `mainnet-<N>-main`
    - Create route table
    - Name tag: `mainnet-<N>`
    - VPC: `mainnet-<N>`
    - Create
    - Select it in the list
    - Actions > Edit routes
    - Leave the "local" route alone
    - Add route
    - Destination: `0.0.0.0/0`
    - Target: `Internet Gateway` > `mainnet-<N>`
    - Save routes
    - Actions > Edit subnet associations
    - Select `mainnet-<N>`
    - Save
1. ECS > Clusters
    - Create Cluster
    - EC2 Linux + Networking
    - Next step
    - Cluster name: `mainnet-<N>`
    - Key pair: `sc-node-ec2-mainnet`
    - VPC: `mainnet-<N>`
    - Subnets: `mainnet-<N>` (there should only be one choice)
    - Security group: `(default)` (there should only be one choice)
    - Create
    - View Cluster
1. EC2 > Load Balancers
    - Create Load Balancer
    - Select "Classic Load Balancer"
    - Name: `mainnet-<N>`
    - Create LB Inside: `mainnet-<N>`
    - Remove the "HTTP/80" entry
    - Add: `TCP --- 26660 - TCP -- 26660`
    - Add: `HTTPS - 26670 - HTTP - 26670`
    - Add: `HTTPS - 3030 -- HTTP - 3030 `
    - Select available subnets: (there should only be one choice; select it)
    - Next: Assign Security Groups
    - Select `default` (it should already be selected)
    - Next: Configure Security Settings
    - Certificate type: `Choose a certificate from ACM (recommended)`
    - Certificate: `*.ndau.tech` (use the AWS Certificate Manager to request it if needed)
    - Next: Configure Health Check
    - Ping Protocol: `HTTP`
    - Ping Port: `3030`
    - Ping Path: `/health`
    - Change the Healthy interval from `10` to `2`
    - Next: Add EC2 Instances
    - Select `ECS Instance - EC2ContainerService-mainnet-<N>`
    - Next: Add Tags
    - Review and Create
    - Create
    - If there is an "unknown error" reported, click "Review and resolve", then "Create" again
    - Ignore failing health checks until the end of the remaining steps
1. EC2 > Network Interfaces
    - Filter by VPC ID to find all the Network Interfaces created by the Load Balancer
    - Name them all appropriately: `mainnet-<N>:{0,1,2}`
    - The two with Description `ELB mainnet-<N>` can be left alone; AWS clears their names periodically.
1. EC2 > Load Balancers
    - Select `mainnet-<N>`
    - Find the "DNS name" at the bottom under the "Description" tab, select the text (up to the .com) and copy it to the clipboard
    - Go to the AWS Route 53 page
    - Hosted zones
    - `ndau.tech.`
    - Create Record Set
    - Name: `mainet-<N>` (.ndau.tech)
    - Type: `A - IPv4 address`
    - Alias: `Yes`
    - Alias Target: (paste the DNS name from the clipboard)
        - It'll get a `dualstack.` prepended automatically, leave it
        - Add a `.` to the end of it (might not matter)
    - Create
    - Run `dig +noall +answer mainnet-<N>.ndau.tech` and make sure the `A` records that come back mention `mainnet-<N>.ndau.tech` in them.  If not, they haven't propogated and the nodes won't work yet.  Wait for this to happen before moving on.
1. ECS > Task Definitions
    - Create new Task Definition
    - Select "EC2"
    - Next step
    - Task Definition Name: `mainnet-<N>`
    - Task Role: `None`
    - Skip to the bottom and click "Configure via JSON"
    - Set the JSON how you want (sample below)
    - Save
    - Create
1. ECS > Clusters
    - Click the `mainnet-<N>` cluster link
    - Services > Create
    - Launch type: `EC2`
    - Task Definition
         - Family: `mainnet-<N>`
         - Revision: `(latest)`
    - Service name: `mainnet-<N>`
    - Number of tasks: `1`
    - Minimum healthy percent: `0`
    - Maximum percent: `100`
    - Placement Templates: `One Task Per Host`
    - Next step
    - Uncheck "Enable service discovery integration"
    - Next step
    - Next step (again)
    - Create Service
    - View Service

At this point, the node is up and running and ready to use.

## Sample Task Definition JSON

Here is the Task Definition JSON for a `mainnet-<N>` node.

1. Copy/paste it into the JSON box when setting up the Task Definition.
1. Replace all occurrences of `mainnet-<N>` with the desired node name.  e.g. `mainnet-6`
1. Configure snapshot environment variables
    - Leave the snapshot name blank for it to use the latest
    - Set `SNAPSHOT_INTERVAL` (e.g. "4h") and the `AWS_*` variables to have periodic backups uploaded to S3
1. Set the `BASE64_NODE_IDENTITY` and `PERSISTENT_PEERS` environment variable values (beyond the scope of this document)

NOTE: If you change the image used, you must do a rolling restart of mainnet nodes (upgrade one at a time, letting it rejoin the network before restarting the next) and update `s3://ndau-images/current-mainnet.txt` to reference the new SHA (in this example, it's "cb8e545").

```json
{
    "ipcMode": null,
    "executionRoleArn": null,
    "containerDefinitions": [
        {
            "dnsSearchDomains": null,
            "logConfiguration": null,
            "entryPoint": null,
            "portMappings": [
                {
                    "hostPort": 26660,
                    "protocol": "tcp",
                    "containerPort": 26660
                },
                {
                    "hostPort": 26670,
                    "protocol": "tcp",
                    "containerPort": 26670
                },
                {
                    "hostPort": 3030,
                    "protocol": "tcp",
                    "containerPort": 3030
                }
            ],
            "command": [
                "/image/docker-run.sh"
            ],
            "linuxParameters": null,
            "cpu": 512,
            "environment": [
                {
                    "name": "NETWORK",
                    "value": "mainnet"
                },
                {
                    "name": "NODE_ID",
                    "value": "mainnet-<N>"
                },
                {
                    "name": "SNAPSHOT_NAME",
                    "value": ""
                },
                {
                    "name": "AWS_ACCESS_KEY_ID",
                    "value": ""
                },
                {
                    "name": "AWS_SECRET_ACCESS_KEY",
                    "value": ""
                },
                {
                    "name": "SNAPSHOT_INTERVAL",
                    "value": ""
                },
                {
                    "name": "BASE64_NODE_IDENTITY",
                    "value": ""
                },
                {
                    "name": "PERSISTENT_PEERS",
                    "value": ""
                },
                {
                    "name": "HONEYCOMB_KEY",
                    "value": "b5d540e08c05885849ae13cd7886df04"
                },
                {
                    "name": "HONEYCOMB_DATASET",
                    "value": "sc-node-mainnet"
                }
            ],
            "resourceRequirements": null,
            "ulimits": null,
            "dnsServers": null,
            "mountPoints": [],
            "workingDirectory": null,
            "secrets": null,
            "dockerSecurityOptions": null,
            "memory": 1024,
            "memoryReservation": 512,
            "volumesFrom": [],
            "stopTimeout": null,
            "image": "578681496768.dkr.ecr.us-east-1.amazonaws.com/sc-node:cb8e545",
            "startTimeout": null,
            "dependsOn": null,
            "disableNetworking": null,
            "interactive": null,
            "healthCheck": null,
            "essential": true,
            "links": null,
            "hostname": null,
            "extraHosts": null,
            "pseudoTerminal": null,
            "user": null,
            "readonlyRootFilesystem": null,
            "dockerLabels": null,
            "systemControls": null,
            "privileged": null,
            "name": "mainnet-<N>"
        }
    ],
    "memory": null,
    "taskRoleArn": "",
    "family": "mainnet-<N>",
    "pidMode": null,
    "requiresCompatibilities": [
        "EC2"
    ],
    "networkMode": null,
    "cpu": null,
    "proxyConfiguration": null,
    "volumes": [],
    "placementConstraints": []
}
```
