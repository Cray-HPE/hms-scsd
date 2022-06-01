#!/usr/bin/env bash

# MIT License
#
# (C) Copyright [2022] Hewlett Packard Enterprise Development LP
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

set -x

# Configure docker compose
export COMPOSE_PROJECT_NAME=${RANDOM}
export COMPOSE_FILE="docker-compose.test.integration.yaml"
export SCSD_VERSION=$(cat .version)
export LOG_FILE="scsd-integration-test-${COMPOSE_PROJECT_NAME}.logs"


echo "COMPOSE_PROJECT_NAME: ${COMPOSE_PROJECT_NAME}"
echo "COMPOSE_FILE: ${COMPOSE_FILE}"

#TODO
#export SCSD_TEST_K8S_AUTH_URL="http://${COMPOSE_PROJECT_NAME}_vault_1:8200/v1/auth/kubernetes/login"
export SCSD_TEST_K8S_AUTH_URL="http://${COMPOSE_PROJECT_NAME}_fake-vault_1:8200/v1/auth/kubernetes/login"
#export SCSD_TEST_VAULT_PKI_URL="http://${COMPOSE_PROJECT_NAME}_vault_1:8200/v1/pki_common/issue/pki-common"
export SCSD_TEST_VAULT_PKI_URL="http://${COMPOSE_PROJECT_NAME}_fake-vault_1:8200/v1/pki_common/issue/pki-common"
#export SCSD_TEST_VAULT_CA_URL="http://${COMPOSE_PROJECT_NAME}_vault_1:8200/v1/pki_common/ca_chain"
export SCSD_TEST_VAULT_CA_URL="http://${COMPOSE_PROJECT_NAME}_fake-vault_1:8200/v1/pki_common/ca_chain"
export CRAY_VAULT_JWT_FILE="/tmp/k8stoken"
export CRAY_VAULT_ROLE_FILE="/tmp/k8stoken"


function cleanup() {
    docker-compose down
    if [[ $? -ne 0 ]]; then
        echo "Failed to decompose environment!"
        exit 1
    fi
    exit $1
}


# Get the base containers running
echo "Starting containers..."
docker-compose build --build-arg SCSD_VERSION=${SCSD_VERSION}
docker-compose up -d cray-scsd
docker-compose up --exit-code-from integration-tests integration-tests

test_result=$?

# Clean up
echo "Cleaning up containers..."
if [[ $test_result -ne 0 ]]; then
    docker-compose logs > ${LOG_FILE} 2>&1
    echo "================================================="
    echo "LOGS:"
    cat ${LOG_FILE}
    echo "================================================="
    echo
    echo "Integration tests FAILED!"
    cleanup 1
fi

echo "Integration tests PASSED!"
cleanup 0
