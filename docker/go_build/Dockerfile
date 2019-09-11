# The go_build stage has all the tooling necessary to clone repos and build
# go programs.
#
# WARNING: this stage contains private data!
# THIS CONTAINER MUST NEVER BE UPLOADED TO ANY CONTAINER REGISTRY
# NO CONTAINER DERIVING FROM THIS MAY EVER BE UPLOADED TO ANY CONTAINER REGISTRY
#
# Intended usage: built locally, this container provides caching and git access
# to private repos. Layer-2 containers derived from this one should be used
# to produce golang executables without code duplication.
# Layer-3 containers deriving from a bare Alpine may copy artifacts from this
# container or descendants; those may safely be uploaded to registries.
FROM golang:1.12-alpine3.9 AS go_build

# Install extra tools.
RUN apk add --no-cache bash git openssh build-base

# Arguments passed via --build-arg.
ARG SSH_PRIVATE_KEY

# Add credentials.
RUN mkdir /root/.ssh/ && \
    echo "${SSH_PRIVATE_KEY}" > /root/.ssh/id_rsa && \
    chmod 600 /root/.ssh/id_rsa

# Make sure the domain is accepted.
RUN touch /root/.ssh/known_hosts && \
    ssh-keyscan github.com >> /root/.ssh/known_hosts
