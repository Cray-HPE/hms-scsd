# Copyright 2020 Hewlett Packard Enterprise Development LP

# Dockerfile for building HMS SCSD.

FROM dtr.dev.cray.com/baseos/golang:1.14-alpine3.12 AS build-base

RUN set -ex \
    && apk update \
    && apk add build-base

FROM build-base AS base

# Copy all the necessary files to the image.
COPY cmd $GOPATH/src/stash.us.cray.com/HMS/hms-scsd/cmd
COPY vendor $GOPATH/src/stash.us.cray.com/HMS/hms-scsd/vendor

### Build Stage ###

FROM base AS builder

# Now build
RUN set -ex \
    && go build -v -i -o scsd stash.us.cray.com/HMS/hms-scsd/cmd/scsd

### Final Stage ###

FROM dtr.dev.cray.com/baseos/alpine:3.12
LABEL maintainer="Cray, Inc."
STOPSIGNAL SIGTERM
EXPOSE 25309
STOPSIGNAL SIGTERM

RUN set -ex \
    && apk update \
    && apk add --no-cache curl

# Setup environment variables.
# ENV VAULT_ENABLED="true"
ENV VAULT_ADDR="http://cray-vault.vault:8200"
ENV VAULT_SKIP_VERIFY="true"
# ENV VAULT_KEYPATH="secret/hms-creds"

# Get scsd from the builder stage.
COPY --from=builder /go/scsd /usr/local/bin
COPY .version /var/run/scsd_version.txt

# Set up the command to start the service, the run the init script.
CMD scsd
