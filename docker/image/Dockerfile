FROM alpine:3.8

WORKDIR /nodegroup
COPY . /nodegroup

RUN apk add --no-cache bash
RUN /bin/bash docker-setup.sh

# We only need to expose Tendermint's ports.
# The outside world will communicate with the container through the RPC ports.
# Tendermint itself will communicate with other containers through the P2P ports.
# All other processes in the container will communicate with each other through internal ports.
EXPOSE 26660-26661 26670-26671

CMD ["/bin/bash", "docker-run.sh"]
