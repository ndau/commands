# Multi-stage Dockerfile
# For securely using the machine ssh key to access the private ndau repos.

# This stage has all purely external dependencies we'll want in the final image.
# Nothing here will ever change based on ndev code changes, so it's safe to
# build/cache it all right up top.

FROM alpine:3.19.1 AS combined_node_base

# Install extra tools.
RUN apk add --no-cache bash bind-tools ca-certificates curl openssl sed jq python3 py3-pip && \
#    python3 -m pip install remarshal && \
#    python3 -m pip install awscli && \
    update-ca-certificates 2>/dev/null

RUN python3 --break-system-packages -m pip install remarshal
RUN python3 --break-system-packages -m pip install awscli

# Use a specific version of redis.
COPY --from=redis:7.2.4-alpine3.19 [ \
    "usr/local/bin/redis-server", \
    "usr/local/bin/redis-sentinel", \
    "usr/local/bin/redis-cli", \
    "usr/local/bin/redis-benchmark", \
    "usr/local/bin/redis-check-rdb", \
    "usr/local/bin/redis-check-aof", \
    "/usr/local/bin/" \
    ]

# The go_build stage has all the tooling necessary to clone repos and build
# go programs.
FROM golang:1.21.6-alpine3.19 AS go_build

# Install extra tools.
RUN apk add --no-cache bash git openssh make gcc libc-dev

# The build_noms stage builds noms, caching results when posssible
FROM go_build AS build_noms
# We're going to use our own noms.....
ENV NOMS_DIR=$GOPATH/src/github.com/ndau
RUN mkdir -p "$NOMS_DIR"
# This copy exists to bust the cache if new commits exist
COPY ./noms_sha /
RUN git clone https://github.com/ndau/noms.git "$NOMS_DIR"/noms
RUN cd "$NOMS_DIR"/noms && go get -u "$NOMS_DIR"/noms/...
RUN cd "$NOMS_DIR"/noms && go build "$NOMS_DIR"/noms/cmd/noms && mv noms /bin

# The build_tendermint stage builds tendermint, caching results when posssible
FROM go_build AS build_tendermint
# Use pre-built tendermint image
# build tendermint because pre-built 0.32.6 doesn't exist
ENV TENDERMINT_DIR=$GOPATH/src/github.com/tendermint
ENV TENDERMINT_VER=v0.32.6
RUN mkdir -p "$TENDERMINT_DIR"
RUN cd "$TENDERMINT_DIR" && git clone https://github.com/tendermint/tendermint.git
RUN cd "$TENDERMINT_DIR"/tendermint && git checkout "$TENDERMINT_VER"
RUN cd "$TENDERMINT_DIR"/tendermint && GO111MODULE=on make build
RUN cp "$TENDERMINT_DIR"/tendermint/build/tendermint /bin

# Set up commands directories

FROM go_build as get_commands_deps
ENV ONEIRO_DIR=$GOPATH/src/github.com/ndau
ENV COMMANDS_DIR=${ONEIRO_DIR}/commands
RUN mkdir -p ${COMMANDS_DIR}

# The build_commands stage builds ndev programs from the commands repo,
# caching results when possible.
FROM go_build AS build_commands

# Arguments passed via --build-arg.
ARG COMMANDS_BRANCH
ENV COMMANDS_BRANCH=${COMMANDS_BRANCH}
ENV ONEIRO_DIR=$GOPATH/src/github.com/ndau
ENV COMMANDS_DIR=${ONEIRO_DIR}/commands

# this shouldn't normally bust the cache
COPY docker-build.sh /image/

# bust the cache when the commands repo has changed
COPY commands_sha /image/
RUN git clone https://github.com/ndau/commands.git \
    --branch "$COMMANDS_BRANCH" \
    "${COMMANDS_DIR}"

# Perform custom setup steps from inside the intermediate image.
ARG RUN_UNIT_TESTS
ENV RUN_UNIT_TESTS=${RUN_UNIT_TESTS}
RUN apk add yarn
RUN /image/docker-build.sh

# Build the final image.  If published, this should leave no trace of earlier stages.
FROM combined_node_base AS combined_node

# Copy image support files needed for running the node group inside the container.
COPY ./docker-conf.sh \
    ./docker-config-default.toml \
    ./docker-config-mainnet.toml \
    ./docker-config-testnet.toml \
    ./docker-dns.sh \
    ./docker-env.sh \
    ./docker-procmon-claimer.toml \
    ./docker-procmon-noclaimer.toml \
    ./docker-run.sh \
    ./docker-snapshot.sh \
    ./s3-multipart-upload.sh \
    /image/

# Copy the noms image we built
COPY --from=build_noms /bin/noms /image/bin/

# Copy the tendermint image we built
COPY --from=build_tendermint /bin/tendermint /image/bin/

# Copy the output from the intermediate image.
COPY --from=build_commands /image/bin/* /image/bin/

# We only need to expose Tendermint and ndauapi ports.
# The outside world will communicate with the container through the TM RPC ports and ndauapi.
# Tendermint itself will communicate with other containers through the P2P ports.
# All other processes in the container will communicate with each other through internal ports.
EXPOSE 26660 26670 3030

CMD ["/image/docker-run.sh"]
