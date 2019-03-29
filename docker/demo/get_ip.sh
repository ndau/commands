#!/bin/bash

# Get the IP of the machine we're running containers on.
# Can be used as peer IPs for multiple local containers connecting to each other.
ping -c 1 $(hostname) | sed -n -e 's/^PING .* (\(.*\)):.*/\1/p'
