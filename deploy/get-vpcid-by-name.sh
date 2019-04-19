#!/bin/bash

# gets vpc by name

name=$1

aws ec2 describe-vpcs |
jq -r " \
  .Vpcs[] as \$parent | \
  \$parent.Tags[]?.Value | \
  select (.==\"${name}\") | \
  \$parent | \
  .VpcId"
