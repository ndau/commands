FROM golang:1.18.2-alpine3.16 AS build

ENV BUILD_DIR=/build
RUN mkdir "$BUILD_DIR"

ENV VERSION=v1.5.1-20-g76755ca
ENV VERSION_PKG=github.com/ndau/commands/vendor/github.com/ndau/ndau/pkg/version
ENV LDFLAGS=github.com/ndau/commands/vendor/github.com/ndau/ndau/pkg/version.version=v1.5.1-20-g76755ca
#ENV VERSION_PKG="$NDEV_SUBDIR/commands/vendor/$NDEV_SUBDIR/ndau/pkg/version"

#RUN echo "export VERSION=$(git describe --long --tags --match="v*")" >> ~/.profile
#RUN echo $VERSION
#RUN echo "export VERSION_PKG=github.com/ndau/commands/vendor/github.com/ndau/ndau/pkg/version" >> ~/.profile
#RUN echo $VERSION_PKG


RUN mkdir -p "/go/src/github.com/ndau/commands"

WORKDIR /go/src/github.com/ndau/commands

COPY . .

#RUN go mod download
RUN go build -ldflags "-X $LDFLAGS" -o $BUILD_DIR ./cmd/ndaunode
RUN go build -ldflags "-X $LDFLAGS" -o $BUILD_DIR ./cmd/ndauapi

RUN go build -o $BUILD_DIR ./cmd/generate
RUN go build -o $BUILD_DIR ./cmd/keytool
RUN go build -o $BUILD_DIR ./cmd/ndau
RUN go build -o $BUILD_DIR ./cmd/claimer

COPY ./docker/image/docker-conf.next.sh \
    ./docker/image/docker-config-default.toml \
    ./docker/image/docker-config-mainnet.toml \
    ./docker/image/docker-config-testnet.toml \
    ./docker/image/docker-dns.next.sh \
    ./docker/image/docker-env.next.sh \
    ./docker/image/docker-procmon-claimer.toml \
    ./docker/image/docker-procmon-noclaimer.toml \
    ./docker/image/docker-run.next.sh \
    ./docker/image/docker-snapshot.sh \
    ./docker/image/s3-multipart-upload.sh \
    /build/

FROM alpine:3.16.0

# Install extra tools.
RUN apk add --no-cache bash bind-tools ca-certificates curl openssl sed jq python3 py3-pip && \
    python3 -m pip install remarshal && \
    python3 -m pip install awscli && \
    update-ca-certificates 2>/dev/null

COPY --from=build /build/* /bin/

# We only need to expose Tendermint and ndauapi ports.
# The outside world will communicate with the container through the TM RPC ports and ndauapi.
# Tendermint itself will communicate with other containers through the P2P ports.
# All other processes in the container will communicate with each other through internal ports.
EXPOSE 26660 26670 3030

# CMD ["/bin/docker-run.next.sh"]