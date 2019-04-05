# Multi-stage Dockerfile
# For securely using the machine ssh key to access the private oneiro-ndev repos.
FROM golang:1.11-alpine3.8 AS intermediate

# Install extra tools.
# TODO: Consider combining/caching this with the above FROM to make the intermediate image.
RUN apk add --no-cache bash git openssh

# Arguments passed via --build-arg.
ARG SSH_PRIVATE_KEY

# Add credentials on build.
RUN mkdir /root/.ssh/ && \
    echo "${SSH_PRIVATE_KEY}" > /root/.ssh/id_rsa && \
    chmod 600 /root/.ssh/id_rsa

# Make sure the domain is accepted.
RUN touch /root/.ssh/known_hosts && \
    ssh-keyscan github.com >> /root/.ssh/known_hosts

# Perform custom setup steps from inside the intermediate image.
# TODO: Consider caching the setup results (binary executables) in a separate image.
COPY . /image
ENV TENDERMINT_VER=v0.31.0
ENV COMMANDS_BRANCH=master
RUN /image/docker-setup.sh

# Build the final image.  This will abandon the previous image and leave no trace of the ssh key.
FROM alpine:3.8 AS build

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

# Install extra tools.
# TODO: Consider combining/caching this with the above FROM to make the build image.
RUN apk add --no-cache bash ca-certificates sed && update-ca-certificates 2>/dev/null

# Copy the output from the intermediate image.
COPY --from=intermediate /image/bin /image/bin

# Copy image support files needed for running the node group inside the container.
COPY ./docker-conf.sh /image
COPY ./docker-env.sh /image
COPY ./docker-run.sh /image
COPY ./docker-ndau.toml /image

# We only need to expose Tendermint and ndauapi ports.
# The outside world will communicate with the container through the TM RPC ports and ndauapi.
# Tendermint itself will communicate with other containers through the P2P ports.
# All other processes in the container will communicate with each other through internal ports.
EXPOSE 26660 26670 3030

CMD ["/image/docker-run.sh"]
