# Multi-stage Dockerfile
# For securely using the machine ssh key to access the private oneiro-ndev repos.

# This stage has all purely external dependencies we'll want in the final image.
# Nothing here will ever change based on ndev code changes, so it's safe to
# build/cache it all right up top.
FROM alpine:3.9 AS combined_node_base

# Install extra tools.
RUN apk add --no-cache bash bind-tools ca-certificates curl openssl sed jq python3 su-exec && \
    python3 -m pip install remarshal && \
    update-ca-certificates 2>/dev/null

# Use a specific version of redis.
COPY --from=redis:5.0.4-alpine [ \
    "usr/local/bin/redis-server", \
    "usr/local/bin/redis-sentinel", \
    "usr/local/bin/redis-cli", \
    "usr/local/bin/redis-benchmark", \
    "usr/local/bin/redis-check-rdb", \
    "usr/local/bin/redis-check-aof", \
    "/usr/local/bin/" \
    ]

COPY --from=postgres:12-alpine \
    /usr/local/share/postgresql/* \
    /usr/local/share/postgresql/
COPY --from=postgres:12-alpine \
    /usr/local/bin/postgres \
    /usr/local/bin/pg_ctl \
    /usr/local/bin/docker-entrypoint.sh \
    /usr/local/bin/
RUN mv /usr/local/bin/docker-entrypoint.sh /usr/local/bin/postgres-docker-entrypoint.sh

# The go_build stage has all the tooling necessary to clone repos and build
# go programs.
FROM golang:1.12-alpine3.9 AS go_build

# Install extra tools.
RUN apk add --no-cache bash git openssh make


# Arguments passed via --build-arg.
ARG SSH_PRIVATE_KEY

# Add credentials on build.
RUN mkdir /root/.ssh/ && \
    echo "${SSH_PRIVATE_KEY}" > /root/.ssh/id_rsa && \
    chmod 600 /root/.ssh/id_rsa

# Make sure the domain is accepted.
RUN touch /root/.ssh/known_hosts && \
    ssh-keyscan github.com >> /root/.ssh/known_hosts


# The build_noms stage builds noms, caching results when posssible
FROM go_build AS build_noms

ENV ATTICLABS_DIR=$GOPATH/src/github.com/attic-labs
RUN mkdir -p "$ATTICLABS_DIR"
# This copy exists to bust the cache if new commits exist
COPY ./noms_sha /
RUN git clone git@github.com:attic-labs/noms.git "$ATTICLABS_DIR"/noms
RUN go get -u "$ATTICLABS_DIR"/noms/...
RUN cd /bin && go build "$ATTICLABS_DIR"/noms/cmd/noms

# The build_tendermint stage builds tendermint, caching results when posssible
FROM go_build AS build_tendermint
# Use pre-built tendermint image
# COPY --from=tendermint/tendermint:v0.32.3 /usr/bin/tendermint /image/bin/
# build tendermint because pre-built 0.32.5 doesn't exist
ENV TENDERMINT_DIR=$GOPATH/src/github.com/tendermint
ENV TENDERMINT_VER=v0.32.6
RUN mkdir -p "$TENDERMINT_DIR"
RUN cd "$TENDERMINT_DIR" && git clone https://github.com/tendermint/tendermint.git
RUN cd "$TENDERMINT_DIR"/tendermint && git checkout "$TENDERMINT_VER"
RUN cd "$TENDERMINT_DIR"/tendermint && GO111MODULE=on make build
RUN cp "$TENDERMINT_DIR"/tendermint/build/tendermint /bin

# The get_commands_deps stage gets the dependencies used to build the commands
# repo. It's in a separate stage for caching: often, the commands repo's sha
# will change without actually changing the dependencies.
FROM go_build as get_commands_deps
# ensure we have dep
RUN go get -u github.com/golang/dep/...
ENV ONEIRO_DIR=$GOPATH/src/github.com/oneiro-ndev
ENV COMMANDS_DIR=${ONEIRO_DIR}/commands
RUN mkdir -p ${COMMANDS_DIR}
# bust cache only when deps are updated
COPY ./Gopkg.* "$COMMANDS_DIR"/
RUN cd ${COMMANDS_DIR} && dep ensure --vendor-only


# The build_commands stage builds ndev programs from the commands repo,
# caching results when possible.
FROM go_build AS build_commands

# Arguments passed via --build-arg.
ARG COMMANDS_BRANCH
ENV COMMANDS_BRANCH=${COMMANDS_BRANCH}
ENV ONEIRO_DIR=$GOPATH/src/github.com/oneiro-ndev
ENV COMMANDS_DIR=${ONEIRO_DIR}/commands

# this shouldn't normally bust the cache
COPY docker-build.sh /image/

# bust the cache when the commands repo has changed
COPY commands_sha /image/
RUN git clone git@github.com:oneiro-ndev/commands.git \
    --branch "$COMMANDS_BRANCH" \
    "${COMMANDS_DIR}"

# get the vendor folder
COPY --from=get_commands_deps "${COMMANDS_DIR}/vendor" "${COMMANDS_DIR}/vendor"

# Perform custom setup steps from inside the intermediate image.
ARG RUN_UNIT_TESTS
ENV RUN_UNIT_TESTS=${RUN_UNIT_TESTS}
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
    ./docker-procmon.toml \
    ./docker-run.sh \
    ./docker-snapshot.sh \
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
