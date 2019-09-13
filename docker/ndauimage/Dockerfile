# Build the final image. This should contain no trace of earlier stages.
FROM alpine:3.9 AS ndauimage

# our scripts require bash as opposed to bare sh
# bind-tools are required for dig, for manual DNS
RUN apk add --no-cache bash bind-tools ca-certificates && \
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

# Use pre-built tendermint image
COPY --from=tendermint/tendermint:v0.32.3 /usr/bin/tendermint /image/bin/

# Copy the noms image we built
COPY --from=noms /bin/noms /image/bin/

# Copy the output from the intermediate image.
COPY --from=build_commands /image/bin/ndaunode /image/bin/ndauapi /image/bin/procmon /image/bin/

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


# We only need to expose Tendermint and ndauapi ports.
# The outside world will communicate with the container through the TM RPC ports and ndauapi.
# Tendermint itself will communicate with other containers through the P2P ports.
# All other processes in the container will communicate with each other through internal ports.
EXPOSE 26660 26670 3030

CMD ["/image/docker-run.sh"]
