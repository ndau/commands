#!/bin/bash

# echos to stderr
errcho() { >&2 echo -e "$@"; }

# echos to stderr if VERBOSE is set to true
verrcho() { [ "$VERBOSE" == "true" ] && errcho "$@"; }

# echos to stderr and quits
err() {
    errcho "$@"
    exit 1
}

# if the first argument is empty, will print second argument to stderr and exit with an error
err_if_empty() {
  if [ -z "$1" ]; then
    err $2
  fi
}

# if the first argument is empty, will print second argument to stderr and exit with an error
check_empty() {
  if [ -z "$1" ]; then
    errcho $2
  fi
}


# looks up and returns a number for a network
# This is used to determine the network digit of a port assignment.
get_network_number() {
  case "$1" in
    devnet) echo 0
        ;;
    testnet) echo 1
        ;;
    *) err "Network name: $1. Not recognized."
        ;;
  esac
}

# looks up and returns a number for a service
# This is used to determine the service digit of a port assignment.
get_service_number() {
  case "$1" in
    rpc) echo 1
        ;;
    p2p) echo 2
        ;;
    ndauapi) echo 3
        ;;
    *) err "service name: $1. Not recognized."
        ;;
  esac
}

# calculates the port based on a formula
calc_port() {
  local network_name=$1
  local service_name=$2
  local network_number=$(get_network_number $network_name)
  local service_number=$(get_service_number $service_name)
  local node_number=$3
  echo $((BASE_PORT + (1000*network_number) + (100*service_number) + node_number))
}
