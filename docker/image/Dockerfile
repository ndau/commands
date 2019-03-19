# Multi-stage Dockerfile
# For securely using the machine ssh key to access the private oneiro-ndev repos.
FROM golang:1.11-alpine3.8 AS intermediate

# Install extra tools.
RUN apk add --no-cache bash git openssh sed

# Arguments passed via --build-arg.
ARG SSH_PRIVATE_KEY
ARG GENESIS_TOML

# Add credentials on build.
RUN mkdir /root/.ssh/ && \
    echo "${SSH_PRIVATE_KEY}" > /root/.ssh/id_rsa && \
    chmod 600 /root/.ssh/id_rsa

# Make sure the domain is accepted.
RUN touch /root/.ssh/known_hosts && \
    ssh-keyscan github.com >> /root/.ssh/known_hosts

# Perform custom setup steps from inside the intermediate image.
COPY . /image
RUN echo "${GENESIS_TOML}" > /image/genesis.toml && \
    /bin/bash /image/docker-setup.sh && \
    /bin/bash /image/docker-conf.sh

# Build the final image.  This will abandon the previous image and leave no trace of the ssh key.
FROM alpine:3.8 AS build

# Copy the output from the intermediate image.
COPY --from=intermediate /image/bin /image/bin
COPY --from=intermediate /image/data /image/data

# Copy image support files needed for running the node group inside the container.
COPY ./docker-env.sh /image
COPY ./docker-run.sh /image

# Install extra tools.
RUN apk add --no-cache bash redis=4.0.11-r0 sed

# We only need to expose Tendermint's ports.
# The outside world will communicate with the container through the RPC ports.
# Tendermint itself will communicate with other containers through the P2P ports.
# All other processes in the container will communicate with each other through internal ports.
EXPOSE 26660 26661 26670 26671 3030

CMD ["/bin/bash", "/image/docker-run.sh"]
