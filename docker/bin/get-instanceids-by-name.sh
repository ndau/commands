#!/bin/bash

# returns instance ids by name

name=$1 #"sc-node-cluster$"

aws ec2 describe-instances | jq -r " \
  .Reservations[].Instances[] as \$parent | \
  \$parent.Tags[]?.Value | \
  select (. | match(\"${name}\") ) | \
  \$parent | \
  .InstanceId" | uniq # Don't know why uniq is required, but doubles appear.
