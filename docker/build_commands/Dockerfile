# The get_deps stage gets the dependencies used to build the commands
# repo. It's in a separate stage for caching: often, the commands repo's sha
# will change without actually changing the dependencies.
FROM go_build as get_deps
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
COPY --from=get_deps "${COMMANDS_DIR}/vendor" "${COMMANDS_DIR}/vendor"

# Perform custom setup steps from inside the intermediate image.
ARG RUN_UNIT_TESTS
ENV RUN_UNIT_TESTS=${RUN_UNIT_TESTS}
RUN /image/docker-build.sh

