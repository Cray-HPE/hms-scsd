# MIT License
#
# (C) Copyright [2020-2022] Hewlett Packard Enterprise Development LP
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

# Dockerfile for running API/integration tests.

# Author: mpkelly
# Date: 26-February 2020

FROM artifactory.algol60.net/docker.io/library/golang:1.16-alpine AS builder
LABEL maintainer="Hewlett Packard Enterprise"

RUN go env -w GO111MODULE=auto

COPY *.sh /usr/local/bin/

# Run the main test script

ARG IN_SCSD
ARG IN_HSM
ARG IN_X_S0_PORT
ARG IN_X_S1_PORT
ARG IN_X_S2_PORT
ARG IN_X_S3_PORT
ARG IN_X_S6_PORT
ARG IN_X_S7_PORT
ARG SCSD_VERSION
ARG CRAY_VAULT_JWT_FILE
ARG CRAY_VAULT_ROLE_FILE
ARG SCSD_TEST_K8S_AUTH_URL
ARG SCSD_TEST_VAULT_PKI_URL
ARG SCSD_TEST_VAULT_CA_URL

RUN apk -U upgrade && \
    apk add bash curl jq

RUN mkdir -p /var/run && \
    echo ${SCSD_VERSION} > /var/run/scsd_version.txt

ENV SCSD=${IN_SCSD}
ENV HSM=${IN_HSM}
ENV X_S0_PORT=${IN_X_S0_PORT}
ENV X_S1_PORT=${IN_X_S1_PORT}
ENV X_S2_PORT=${IN_X_S2_PORT}
ENV X_S3_PORT=${IN_X_S3_PORT}
ENV X_S6_PORT=${IN_X_S6_PORT}
ENV X_S7_PORT=${IN_X_S7_PORT}

CMD ["sh", "-c", "runit.sh"]
