#!/bin/bash

set -e

# we've previously written the node password into a special file;
# let's ensure that the database knows about it
pw_file="/image/postgres-node-pw"
if [ -s "$pw_file" ]; then
    psql -w ndau postgres --command="ALTER ROLE node WITH PASSWORD '$(cat "$pw_file")'"
fi
