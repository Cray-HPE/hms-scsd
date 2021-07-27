# MIT License
#
# (C) Copyright [2020-2021] Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.

# Dockerfile for building HMS SCSD.

FROM arti.dev.cray.com/baseos-docker-master-local/golang:1.16-alpine3.13 AS build-base

RUN set -ex \
    && apk update \
    && apk add build-base

FROM build-base AS base

RUN go env -w GO111MODULE=auto

# Copy all the necessary files to the image.
COPY cmd $GOPATH/src/github.com/Cray-HPE/hms-scsd/cmd
COPY vendor $GOPATH/src/github.com/Cray-HPE/hms-scsd/vendor

### Build Stage ###

FROM base AS builder

# Now build
RUN set -ex \
    && go build -v -i -o scsd github.com/Cray-HPE/hms-scsd/cmd/scsd

### Final Stage ###

FROM arti.dev.cray.com/baseos-docker-master-local/alpine:3.13
LABEL maintainer="Hewlett Packard Enterprise"
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
