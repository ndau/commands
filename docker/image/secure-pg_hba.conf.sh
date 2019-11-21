#!/bin/bash

set -e -x

# shellcheck source=docker-env.sh
source /image/docker-env.sh

authfile="$PGDATA"/pg_hba.conf

# the final line of the authfile currently reads:
#
#   host all all all md5
#
# that's a useful default, but we want two particular rules to take
# precendence:
#
# 1. the superuser account (postgres) cannot log in via ssh at all
# 2. the guest account (guest) requires no password
#
# https://www.postgresql.org/docs/current/auth-pg-hba-conf.html states:
#
# The first record with a matching connection type, client address, requested
# database, and user name is used to perform authentication. There is no
# “fall-through” or “backup”: if one record is chosen and the authentication
# fails, subsequent records are not considered.
#
# Therefore, we need to insert appropriate rules before that default line.

sed -Ee '/^host\s+all\s+all\s+all/{
    i host all postgres all reject
    i host all guest all trust
}' "$authfile" > "$authfile.new" && mv "$authfile.new" "$authfile"
