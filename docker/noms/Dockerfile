# The noms stage builds noms
FROM go_build:latest AS noms

# This copy exists to bust the cache if new commits exist
COPY ./noms_sha /
RUN go get github.com/attic-labs/noms/...
RUN cd /bin && go build "$GOPATH"/src/github.com/attic-labs/noms/cmd/noms
