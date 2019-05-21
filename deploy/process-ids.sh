#!/bin/bash
# generates a persistent peers list and uploads a tarball full of identities to our S3 bucket

# exit on errors
set -e

# get the directory of this file
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

PEERS=()
tarballs=()

source "$DIR/deploy-lib.sh"

network_name=$1
id_dir=$2

if [ "$#" -lt 2 ]; then
  echo "Usage: $0 network_name identity-directory [--upload]"
  echo "  --upload is an optional flag that will use AWS creds to upload to S3"
  exit 1
fi

for node_number in $( seq 0 9 ); do # automatically deploy up to 10 nodes
    echo "Attempting node: $node_number"
    tar_filename="node-identity-$node_number.tgz"
    tarball="$id_dir/$tar_filename"
    if [ -f  "$tarball" ]; then
        tarballs+=("$tar_filename")
        echo "Found node identity at: $tarball"
        rnd_dir=$(dd if=/dev/urandom count=8 bs=1 2> /dev/null | base64 | tr -dc 0-9a-zA-Z | head -c8)
        mkdir "$rnd_dir"
        tar xzvf "$tarball" -C "$rnd_dir" 2> /dev/null
        node_id=$(TMHOME="$rnd_dir/tendermint" tendermint show_node_id)
        PEERS+=("$node_id@_IP_:$(calc_port p2p $node_number)")
        echo "node_id $node_number: $node_id"
        rm -rf "$rnd_dir"
    fi
done

echo "PERSISTENT_PEERS: $(IFS=,; echo "${PEERS[*]}") "


big_tarball="$DIR/node-identities-${network_name}.tgz"
(
  cd "$id_dir"
  tar zcvf "$big_tarball" ${tarballs[*]}
)
upload=false
while test $# -gt 0; do
    case "$1" in
        --upload) upload=true
            ;;
    esac
    shift
done

if [ "$upload" == true ]; then
  aws s3 cp "$big_tarball" "s3://ndau-deploy-secrets/$(basename $big_tarball)"
fi

