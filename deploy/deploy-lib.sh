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
get_network_number() {
  case "$1" in
    devnet) echo 0
        ;;
    *) err "Network name: $1. Not recognized."
        ;;
  esac
}

# looks up and returns a port for a service
get_service_port() {
  case "$1" in
    rpc) echo 26670
        ;;
    p2p) echo 26660
        ;;
    ndauapi) echo 3030
        ;;
    *) err "service name: $1. Not recognized."
        ;;
  esac
}

# calculates the port based on a formula
calc_port() {
  local service_name=$1
  local service_port=$(get_service_port $service_name)
  local node_number=$2
  echo $((service_port + node_number))
}
