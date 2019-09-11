FROM ndauimage AS integration_tests

# Install extra tools.
RUN apk add --no-cache bash bind-tools ca-certificates curl openssl sed jq python3 && \
    python3 -m pip install remarshal && \
    update-ca-certificates 2>/dev/null

# Copy the output from the intermediate image.
COPY --from=build_commands /image/bin/* /image/bin/
