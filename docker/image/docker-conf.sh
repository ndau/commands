#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
# shellcheck source=docker-env.sh
source "$SCRIPT_DIR"/docker-env.sh

if [ "$SNAPSHOT_NAME" = "$GENERATED_GENESIS_SNAPSHOT" ]; then
    # Generate a new genesis snapshot with this node as its only validator.
    echo Generating genesis...
    "$BIN_DIR"/generate -v -g "$SYSTEM_VARS_TOML" -a "$SYSTEM_ACCOUNTS_TOML"
    mkdir -p "$DATA_DIR"
else
    SNAPSHOT_DIR="$SCRIPT_DIR/snapshot"
    mkdir -p "$SNAPSHOT_DIR"

    # If a local snapshot file exists, it must have been copied in by the outside world; use it.
    if [ -f "$LOCAL_SNAPSHOT" ]; then
        SNAPSHOT_FILE=$(basename "$LOCAL_SNAPSHOT")

        echo "Moving $SNAPSHOT_FILE..."
        mv "$LOCAL_SNAPSHOT" "$SNAPSHOT_DIR/$SNAPSHOT_FILE"
    else
        # No snapshot given means "use the latest".
        if [ -z "$SNAPSHOT_NAME" ]; then
            LATEST_FILE="latest-$NETWORK.txt"
            LATEST_PATH="$SNAPSHOT_DIR/$LATEST_FILE"

            echo "Fetching $LATEST_FILE..."
            rm -f "$LATEST_PATH"
            curl -s -o "$LATEST_PATH" "$SNAPSHOT_URL/$LATEST_FILE"
            if [ ! -f "$LATEST_PATH" ]; then
                echo "Unable to fetch $SNAPSHOT_URL/$LATEST_FILE"
                exit 1
            fi

            SNAPSHOT_NAME=$(cat $LATEST_PATH)
        fi

        SNAPSHOT_FILE="$SNAPSHOT_NAME.tgz"

        echo "Fetching $SNAPSHOT_FILE..."
        curl -s -o "$SNAPSHOT_DIR/$SNAPSHOT_FILE" "$SNAPSHOT_URL/$SNAPSHOT_FILE"
    fi

    echo "Extracting $SNAPSHOT_FILE..."
    cd "$SNAPSHOT_DIR" || exit 1
    tar -xf "$SNAPSHOT_FILE"

    echo "Validating $SNAPSHOT_DIR..."
    if [ ! -d "$SNAPSHOT_DIR" ]; then
        echo "Could not find snapshot directory: $SNAPSHOT_DIR"
        exit 1
    fi
    SNAPSHOT_DATA_DIR="$SNAPSHOT_DIR/data"
    if [ ! -d "$SNAPSHOT_DATA_DIR" ]; then
        echo "Could not find data directory: $SNAPSHOT_DATA_DIR"
        exit 1
    fi
    SNAPSHOT_NOMS_DATA_DIR="$SNAPSHOT_DATA_DIR/noms"
    if [ ! -d "$SNAPSHOT_NOMS_DATA_DIR" ]; then
        echo "Could not find noms data directory: $SNAPSHOT_NOMS_DATA_DIR"
        exit 1
    fi
    SNAPSHOT_REDIS_DATA_DIR="$SNAPSHOT_DATA_DIR/redis"
    if [ ! -d "$SNAPSHOT_REDIS_DATA_DIR" ]; then
        echo "Could not find redis data directory: $SNAPSHOT_REDIS_DATA_DIR"
        exit 1
    fi
    SNAPSHOT_TENDERMINT_HOME_DIR="$SNAPSHOT_DATA_DIR/tendermint"
    if [ ! -d "$SNAPSHOT_TENDERMINT_HOME_DIR" ]; then
        echo "Could not find tendermint home directory: $SNAPSHOT_TENDERMINT_HOME_DIR"
        exit 1
    fi
    SNAPSHOT_TENDERMINT_CONFIG_DIR="$SNAPSHOT_TENDERMINT_HOME_DIR/config"
    if [ ! -d "$SNAPSHOT_TENDERMINT_CONFIG_DIR" ]; then
        echo "Could not find tendermint config directory: $SNAPSHOT_TENDERMINT_CONFIG_DIR"
        exit 1
    fi
    SNAPSHOT_TENDERMINT_GENESIS_FILE="$SNAPSHOT_TENDERMINT_CONFIG_DIR/genesis.json"
    if [ ! -f "$SNAPSHOT_TENDERMINT_GENESIS_FILE" ]; then
        echo "Could not find tendermint genesis file: $SNAPSHOT_TENDERMINT_GENESIS_FILE"
        exit 1
    fi

    # Move the snapshot data dir where applications expect it, then remove the temp snapshot dir.
    mv "$SNAPSHOT_DATA_DIR" "$DATA_DIR"
    rm -rf $SNAPSHOT_DIR
fi

check_identity_files() {
  # check to make sure the identity files are where they shold be.
  if [ ! -f "$DATA_DIR/tendermint/config/priv_validator_key.json" ] || \
     [ ! -f "$DATA_DIR/tendermint/config/node_key.json" ]; then
     echo "Identity files were not found in the expected location"
     find "$DATA_DIR/tendermint/config"
     exit 1
  fi
}

# If we have an environment variable that defines identities, do not use an identity file.
if [ ! -z "$BASE64_NODE_IDENTITY" ]; then
  # echo the environment variable, decode the base64, and unzip into the files.
  echo "Using node identity environment variables"
  cd "$DATA_DIR" || exit 1
  echo -n "$BASE64_NODE_IDENTITY" | base64 -d | tar xfvz -
  check_identity_files
else
  # If we have a node identity file, extract its contents to the data dir.
  # It'll blend with other files already there from the snapshot.
  IDENTITY_FILE=node-identity.tgz
  if [ -f "$SCRIPT_DIR/$IDENTITY_FILE" ]; then
      echo "Using existing node identity..."
      # Copy, don't move, in case the node operator wants to copy it out again later.
      # Its presence also prevents us from generating it later.
      cp "$SCRIPT_DIR/$IDENTITY_FILE" "$DATA_DIR"
      cd "$DATA_DIR" || exit 1
      tar -xf "$IDENTITY_FILE"
      check_identity_files
  else
      # When we start without a node identity, we generate one so the node operator can restart
      # this node later, having the same identity every time.
      echo "No node identity found; a new node identity will be generated"
  fi
fi
# If present, expand the claimer config into the image directory
if [ ! -z "$BASE64_CLAIMER_CONFIG" ]; then
    echo "Generating claimer config ..."
    cd "$BIN_DIR" || exit 1
    echo -n "$BASE64_CLAIMER_CONFIG" | base64 -d | tar xfvz -
else
    echo "No '$BASE64_CLAIMER_CONFIG' found; claimer process will not run."
fi

# Tendermint complains if this file isn't here, but it can be empty json.
pvs_dir="$DATA_DIR/tendermint/data"
pvs_file="$pvs_dir/priv_validator_state.json"
if [ ! -f "$pvs_file" ]; then
  mkdir -p "$pvs_dir"
  echo "{}" > "$pvs_file"
fi

# Make directories that don't get created elsewhere.
mkdir -p "$NODE_DATA_DIR"
mkdir -p "$LOG_DIR"

# Choose the appropriate ndau config file for the current network.
NDAU_CONFIG_TOML="$SCRIPT_DIR/docker-config-$NETWORK.toml"
if [ ! -f "$NDAU_CONFIG_TOML" ]; then
    NDAU_CONFIG_TOML="$SCRIPT_DIR/docker-config-default.toml"
fi
echo "Using ndau config file: $NDAU_CONFIG_TOML"

echo "Webhook config pre-copy:"
sed -e 's/^/  /' <(grep Webhook "$NDAU_CONFIG_TOML")

# Now that we have our ndau data directory (ndau home dir), move the config file into it,
# injecting the appropriate node webhook if so configured
mkdir -p "$NDAUHOME/ndau"
if [ -z "$WEBHOOK_URL" ]; then
    mv "$NDAU_CONFIG_TOML" "$NDAUHOME/ndau/config.toml"
else
    toml2json "$NDAU_CONFIG_TOML" |\
    jq ". + {\"NodeRewardWebhook\": \"$WEBHOOK_URL\"}" |\
    json2toml > "$NDAUHOME/ndau/config.toml"
fi

echo "Webhook config post-copy:"
sed -e 's/^/  /' <(grep Webhook "$NDAUHOME/ndau/config.toml")

if grep -q '^\s*NodeRewardWebhookDelay' "$NDAUHOME/ndau/config.toml"; then
    # configured value is present
    if ! grep -q '^\s*NodeRewardWebhookDelay.*\.0$' "$NDAUHOME/ndau/config.toml"; then
        # configured value is not a float
        sed -i -e '/NodeRewardWebhookDelay.*/s//&.0/' "$NDAUHOME/ndau/config.toml"
        echo "Webhook config post-sed:"
        sed -e 's/^/  /' <(grep Webhook "$NDAUHOME/ndau/config.toml")
    fi
fi

cd "$BIN_DIR" || exit 1

# If we're running on AWS, get our public IP address. If not do nothing - the
# variable will be already set or undefined, both of which are OK.

if [ ! -z "$AWS" ]; then
    PUBLIC_IP_ADDRESS=`curl -s http://169.254.169.254/latest/meta-data/public-ipv4`
    TM_EXTERNAL_ADDRESS="tcp:\\/\\/$PUBLIC_IP_ADDRESS:$TM_P2P_PORT"
    TM_LADDR="tcp:\\/\\/$PUBLIC_IP_ADDRESS:$TM_RPC_PORT"
fi

echo Configuring tendermint...
# This will init all the config for the current container.
# It will leave genesis.json alone, or create one if we're generating a genesis snapshot.
./tendermint init --home "$TM_DATA_DIR"
sed -i -E \
    -e 's/^(create_empty_blocks =) (.*)/\1 false/' \
    -e 's/^(addr_book_strict =) (.*)/\1 false/' \
    -e 's/^(pex =) (.*)/\1 '"$PEX"'/' \
    -e 's/^(seeds =) (.*)/\1 "'"$SEEDS"'"/' \
    -e 's/^(seed_mode =) (.*)/\1 '"$SEED_MODE"'/' \
    -e 's/^(external_address =) (.*)/\1 "'"$TM_EXTERNAL_ADDRESS"'"/' \
    -e 's/^(laddr =) (.*)/\1 "'"$TM_LADDR"'"/' \
    -e 's/^(allow_duplicate_ip =) (.*)/\1 true/' \
    -e 's/^(log_format =) (.*)/\1 "json"/' \
    -e 's/^(log_level =) (.*)/\1 "'"$TM_LOG_LEVEL"'"/' \
    -e 's/^(timeout_prevote =) (.*)/\1 "60s"/' \
    -e 's/^(timeout_precommit =) (.*)/\1 "60s"/' \
    -e 's/^(timeout_commit =) (.*)/\1 "60s"/' \
    -e 's/^(timeout_broadcast_tx_commit =) (.*)/\1 "60s"/' \
    -e 's/^(moniker =) (.*)/\1 "'"$NODE_ID"'"/' \
    "$TM_DATA_DIR/config/config.toml"

if [ "$SNAPSHOT_NAME" = "$GENERATED_GENESIS_SNAPSHOT" ]; then
    echo "Generating genesis noms data..."
    ./ndaunode -use-ndauhome -genesisfile "$SYSTEM_VARS_TOML" -asscfile "$SYSTEM_ACCOUNTS_TOML"
    mv "$NDAUHOME/ndau/noms" "$NOMS_DATA_DIR"

    echo "Starting noms..."
    ./noms serve --port="$NOMS_PORT" "$NOMS_DATA_DIR" 2>&1 &
    sleep 1

    echo "Getting app hash..."
    app_hash=$(./ndaunode -spec http://localhost:"$NOMS_PORT" -echo-hash 2>/dev/null)
    echo "app_hash: $app_hash"

    echo "Killing noms..."
    killall noms
    sleep 1

    echo "Configuring app hash in tendermint..."
    sed -i -E \
        -e 's/"app_hash": ""/"app_hash": "'"$app_hash"'"/' \
        "$TM_DATA_DIR/config/genesis.json"
fi

echo Configuration complete
