SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
source "$SCRIPT_DIR"/docker-env.sh

SNAPSHOT_DIR="$SCRIPT_DIR/snapshot"
mkdir -p "$SNAPSHOT_DIR"
SNAPSHOT_FILE="${SNAPSHOT_URL##*/}"

echo "Fetching $SNAPSHOT_FILE..."
wget -O "$SNAPSHOT_DIR/$SNAPSHOT_FILE" "$SNAPSHOT_URL"

echo "Extracting $SNAPSHOT_FILE..."
cd "$SNAPSHOT_DIR" || exit 1
tar -xf "$SNAPSHOT_FILE"

echo "Validating $SNAPSHOT_DIR..."
if [ ! -d "$SNAPSHOT_DIR" ]; then
    echo "Could not find snapshot directory: $SNAPSHOT_DIR"
    exit 1
fi

SVI_NAMESPACE_FILE="$SNAPSHOT_DIR/svi-namespace"
if [ ! -f "$SVI_NAMESPACE_FILE" ]; then
    echo "Could not find svi namespace file: $SVI_NAMESPACE_FILE"
    exit 1
fi
SVI_NAMESPACE=$(cat "$SVI_NAMESPACE_FILE")
echo "SVI Namespace: $SVI_NAMESPACE"

SNAPSHOT_DATA_DIR="$SNAPSHOT_DIR/data"
if [ ! -d "$SNAPSHOT_DATA_DIR" ]; then
    echo "Could not find data directory: $SNAPSHOT_DATA_DIR"
    exit 1
fi

CHAOS_DATA_DIR="$SNAPSHOT_DATA_DIR/chaos"
if [ ! -d "$CHAOS_DATA_DIR" ]; then
    echo "Could not find chaos data directory: $CHAOS_DATA_DIR"
    exit 1
fi
CHAOS_NOMS_DATA_DIR="$CHAOS_DATA_DIR/noms"
if [ ! -d "$CHAOS_NOMS_DATA_DIR" ]; then
    echo "Could not find chaos noms data directory: $CHAOS_NOMS_DATA_DIR"
    exit 1
fi
CHAOS_REDIS_DATA_DIR="$CHAOS_DATA_DIR/redis"
if [ ! -d "$CHAOS_REDIS_DATA_DIR" ]; then
    echo "Could not find chaos redis data directory: $CHAOS_REDIS_DATA_DIR"
    exit 1
fi
CHAOS_TENDERMINT_DATA_DIR="$CHAOS_DATA_DIR/tendermint"
if [ ! -d "$CHAOS_TENDERMINT_DATA_DIR" ]; then
    echo "Could not find chaos tendermint data directory: $CHAOS_TENDERMINT_DATA_DIR"
    exit 1
fi
CHAOS_TENDERMINT_GENESIS_FILE="$CHAOS_TENDERMINT_DATA_DIR/config/genesis.json"
if [ ! -f "$CHAOS_TENDERMINT_GENESIS_FILE" ]; then
    echo "Could not find chaos tendermint genesis file: $CHAOS_TENDERMINT_GENESIS_FILE"
    exit 1
fi
NDAU_DATA_DIR="$SNAPSHOT_DATA_DIR/ndau"
if [ ! -d "$NDAU_DATA_DIR" ]; then
    echo "Could not find ndau data directory: $NDAU_DATA_DIR"
    exit 1
fi
NDAU_NOMS_DATA_DIR="$NDAU_DATA_DIR/noms"
if [ ! -d "$NDAU_NOMS_DATA_DIR" ]; then
    echo "Could not find ndau noms data directory: $NDAU_NOMS_DATA_DIR"
    exit 1
fi
NDAU_REDIS_DATA_DIR="$NDAU_DATA_DIR/redis"
if [ ! -d "$NDAU_REDIS_DATA_DIR" ]; then
    echo "Could not find ndau redis data directory: $NDAU_REDIS_DATA_DIR"
    exit 1
fi
NDAU_TENDERMINT_DATA_DIR="$NDAU_DATA_DIR/tendermint"
if [ ! -d "$NDAU_TENDERMINT_DATA_DIR" ]; then
    echo "Could not find ndau tendermint data directory: $NDAU_TENDERMINT_DATA_DIR"
    exit 1
fi
NDAU_TENDERMINT_GENESIS_FILE="$NDAU_TENDERMINT_DATA_DIR/config/genesis.json"
if [ ! -f "$NDAU_TENDERMINT_GENESIS_FILE" ]; then
    echo "Could not find ndau tendermint genesis file: $NDAU_TENDERMINT_GENESIS_FILE"
    exit 1
fi

# Copy the snapshot data where the applications expect it, then remove the temp snapshot dir.
mv "$SNAPSHOT_DATA_DIR" "$DATA_DIR"
rm -rf $SNAPSHOT_DIR

# Make data directories that don't get created elsewhere.
mkdir -p "$NODE_DATA_DIR"

cd "$BIN_DIR" || exit 1

echo Configuring intra-nodegroup port references for chaosnode and ndaunode...
./chaosnode --set-ndaunode "http://localhost:$TM_NDAU_RPC_PORT"
./ndaunode --set-chaosnode "http://localhost:$TM_CHAOS_RPC_PORT"

echo Importing genesis data into ndau conf...
# Build up a temporary genesis file to leverage ndaunode's updte-conf-from feature.
genesis_toml="$SCRIPT_DIR"/genesis.toml
echo '["'"$SVI_NAMESPACE"'"]' > "$genesis_toml"
echo '  ["'"$SVI_NAMESPACE"'".c3Zp]' >> "$genesis_toml"
echo '    svi_stub = true' >> "$genesis_toml"
./ndaunode -use-ndauhome -update-conf-from "$genesis_toml"
rm -f "$genesis_toml"

echo Removing ndau mock setting...
sed -i \
    -e "/UseMock/d" \
    "$NODE_DATA_DIR/ndau/config.toml"

echo Configuring tendermint...
# This will init all the config for the current container, leaving genesis.json alone.
./tendermint init --home "$TM_CHAOS_DATA_DIR"
./tendermint init --home "$TM_NDAU_DATA_DIR"
sed -i -E \
    -e 's/^(create_empty_blocks = .*)/# \1/' \
    -e 's/^(create_empty_blocks_interval =) (.*)/\1 "300s"/' \
    -e 's/^(addr_book_strict =) (.*)/\1 false/' \
    -e 's/^(allow_duplicate_ip =) (.*)/\1 true/' \
    -e 's/^(moniker =) (.*)/\1 "'"$NODE_ID"'"/' \
    "$TM_CHAOS_DATA_DIR/config/config.toml" \
    "$TM_NDAU_DATA_DIR/config/config.toml"
sed -i -E \
    -e 's/^(persistent_peers =) (.*)/\1 "'"$CHAOS_PERSISTENT_PEERS"'"/' \
    "$TM_CHAOS_DATA_DIR/config/config.toml"
sed -i -E \
    -e 's/^(persistent_peers =) (.*)/\1 "'"$NDAU_PERSISTENT_PEERS"'"/' \
    "$TM_NDAU_DATA_DIR/config/config.toml"

echo Configuration complete
