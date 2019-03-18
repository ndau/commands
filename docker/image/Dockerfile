# Multi-stage Dockerfile
# For securely using the machine ssh key to access the private oneiro-ndev repos.
FROM golang:1.11-alpine3.8 AS intermediate

# Install extra tools.
RUN apk add --no-cache bash git openssh

# Add credentials on build.
ARG SSH_PRIVATE_KEY
RUN mkdir /root/.ssh/
RUN echo "${SSH_PRIVATE_KEY}" > /root/.ssh/id_rsa
RUN chmod 600 /root/.ssh/id_rsa

# Make sure the domain is accepted.
RUN touch /root/.ssh/known_hosts
RUN ssh-keyscan github.com >> /root/.ssh/known_hosts

# Perform custom setup steps from inside the intermediate image.
RUN mkdir /image
COPY ./*.patch /image/
COPY ./docker-setup.sh /image
RUN /bin/bash /image/docker-setup.sh

# Build the final image.  This will abandon the previous image and leave no trace of the ssh key.
FROM golang:1.11-alpine3.8 AS build

# Copy the output from the intermediate image.
COPY --from=intermediate /go /go

# Copy image support files needed for running the node group inside the container.
RUN mkdir /image
COPY ./docker-run.sh /image/docker-run.sh

# We only need to expose Tendermint's ports.
# The outside world will communicate with the container through the RPC ports.
# Tendermint itself will communicate with other containers through the P2P ports.
# All other processes in the container will communicate with each other through internal ports.
EXPOSE 26660-26661 26670-26671

CMD ["/bin/bash", "docker-run.sh"]
