#!/bin/bash

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

# This is a functional test for SCSD.  It will create a composed environment
# containing a fake HSM, 6 fake Redfish endpoints, and SCSD itself.
#
# Once created, a container will be created which contains test scripts.
# The scripts use CURL calls into SCSD to perform the various endpoint 
# operations.
#
# Once complete, the results of the test suite is printed to stdout and this
# script will exit with 0 on success, non-zero on failure.

set_properties() {
    testtag=scsdtest
    tagsuffix=${RANDOM}_${RANDOM}
    brnet_suffix=ttest
    export PROJ=scsd_${tagsuffix}
    export HTAG=${testtag}
    export HSUFFIX=${tagsuffix}

    # get safe ports to use, avoiding others using them.  This method is not
    # perfect but should work 99.9% of the time.

    export SCSD_RUN_ID=$(($RANDOM % 100))
    BASEPORT=$(($SCSD_RUN_ID * 200 + 25100))
    if [[ "$SCSD_RUN_ID" == "0" ]]; then
        export S_PREFIX=""
    else
        export S_PREFIX=$SCSD_RUN_ID
    fi

    nrut=`pgrep -c -f runIntegrationTest.sh`
    (( PORTBASE = BASEPORT + nrut ))

    (( SCSD_PORT     = PORTBASE ))
    (( FAKE_SM_PORT  = PORTBASE+10 ))
    (( X_S0_PORT = PORTBASE+100 ))
    (( X_S1_PORT = PORTBASE+101 ))
    (( X_S2_PORT = PORTBASE+102 ))
    (( X_S3_PORT = PORTBASE+103 ))
    (( X_S6_PORT = PORTBASE+106 ))
    (( X_S7_PORT = PORTBASE+107 ))

    export SCSD_PORT
    export FAKE_SM_PORT
    export X_S0_PORT
    export X_S1_PORT
    export X_S2_PORT
    export X_S3_PORT
    export X_S6_PORT
    export X_S7_PORT

    export SCSD_HOST="scsd_"$SCSD_RUN_ID
    export FAKE_SM_HOST="fake_hsm_"$SCSD_RUN_ID
    export VAULT_HOST="vault_"$SCSD_RUN_ID
    export X_S0_HOST="x0c0s"$S_PREFIX"0b0"
    export X_S1_HOST="x0c0s"$S_PREFIX"1b0"
    export X_S2_HOST="x0c0s"$S_PREFIX"2b0"
    export X_S3_HOST="x0c0s"$S_PREFIX"3b0"
    export X_S6_HOST="x0c0s"$S_PREFIX"6b0"
    export X_S7_HOST="x0c0s"$S_PREFIX"7b0"

    export CRAY_VAULT_JWT_FILE=/tmp/k8stoken
    export CRAY_VAULT_ROLE_FILE=/tmp/k8stoken
    export SCSD_TEST_K8S_AUTH_URL="http://${VAULT_HOST}:8200/v1/auth/kubernetes/login"
    export SCSD_TEST_VAULT_PKI_URL="http://${VAULT_HOST}:8200/v1/pki_common/issue/pki-common"
    export SCSD_TEST_VAULT_CA_URL="http://${VAULT_HOST}:8200/v1/pki_common/ca_chain"

    logfilename=scsdtest_${PROJ}.logs
}

print_properties() {
    echo "PROJ=$PROJ"
    echo "SCSD_RUN_ID=$SCSD_RUN_ID"
    echo "BASEPORT=$BASEPORT"
    echo "S_PREFIX=$S_PREFIX"
    echo "SCSD_HOST=$SCSD_HOST"
    echo "FAKE_SM_HOST=$FAKE_SM_HOST"
    echo "VAULT_HOST=$VAULT_HOST"
    echo "X_S0_HOST=$X_S0_HOST"
    echo "X_S1_HOST=$X_S1_HOST"
    echo "X_S2_HOST=$X_S2_HOST"
    echo "X_S3_HOST=$X_S3_HOST"
    echo "X_S6_HOST=$X_S6_HOST"
    echo "X_S7_HOST=$X_S7_HOST"
}

start_marker_container () {
    max_expected_run_time=600 # 10 minutes
    export MARKER_CONTAINER=$(\
        docker run -d -it --rm \
            --label "scsd_integration_project=$PROJ" \
            --label "scsd_run_id=$SCSD_RUN_ID" \
            --name "scsd-test-$PROJ" \
            arti.dev.cray.com/baseos-docker-master-local/alpine:3.13 \
            sleep $max_expected_run_time
    )
}

cleanup_containers() {
    # Get rid of the interim containers

    echo " "
    echo "=============== > Deleting temporary containers..."
    echo " "

    for fff in `docker images | grep ${HTAG}_${HSUFFIX} | awk '{printf("%s\n",$3)}'`; do
        docker image rm -f ${fff}
    done

    docker container stop --time 1 $MARKER_CONTAINER
    echo "Stopped marker container: $MARKER_CONTAINER"
}

set_properties
found_unique_ids="false"
for ((i=0; i<5; i++)); do
    if [[ $(docker ps -q --filter label="scsd_integration_project=$PROJ") ]] ||
        [[ $(docker ps -q --filter label="scsd_run_id=$SCSD_RUN_ID") ]]; then
        # The ids are being used by a different run of the tests.
        # Try picking different ids
        echo "WARNING: IDs in use: proj=$PROJ and scsd_run_id=$SCSD_RUN_ID"
        echo "Getting a new set of IDs"
        set_properties
    else
        found_unique_ids="true"
        break;
    fi
done
if [[ "$found_unique_ids" == "false" ]]; then
    echo "WARNING: Failed to find unique IDs. Continuing with the tests, but they may fail with conflicts in the docker names, hostnames, and port numbers."
fi
start_marker_container
print_properties

# It's possible we don't have docker-compose, so if necessary bring our own.

docker_compose_exe=$(command -v docker-compose)
docker_compose_file=docker-compose-functest.yaml

if ! [[ -x "$docker_compose_exe" ]]; then
    if ! [[ -x "./docker-compose" ]]; then
        echo "Getting docker-compose..."
        curl -L "https://github.com/docker/compose/releases/download/1.23.2/docker-compose-$(uname -s)-$(uname -m)" \
        -o ./docker-compose

        if [[ $? -ne 0 ]]; then
            echo "Failed to fetch docker-compose!"
            exit 1
        fi

        chmod +x docker-compose
    fi
    docker_compose_exe="./docker-compose"
fi



DCOMPOSE="${docker_compose_exe} -p ${PROJ} -f ${docker_compose_file}"

# for test/debug of this script, won't normally be used.

if [ $# -ge 1 -a "$1" == "log" ]; then
    PROJ=`docker network ls | grep scsd | grep ttest | awk '{print $2}' | sed 's/_ttest//'`
    DCOMPOSE="${docker_compose_exe} -p ${PROJ} -f ${docker_compose_file}"
    ${DCOMPOSE} logs > ${logfilename} 2>&1
    echo "Logs in ${logfilename}"
    exit 0
fi

if [ $# -ge 1 -a "$1" == "down" ]; then
    PROJ=`docker network ls | grep scsd | grep ttest | awk '{print $2}' | sed 's/_ttest//'`
    DCOMPOSE="${docker_compose_exe} -p ${PROJ} -f ${docker_compose_file}"
    ${DCOMPOSE} logs > ${logfilename} 2>&1
    ${DCOMPOSE} down
    echo "Logs in ${logfilename}"
    exit 0
fi

echo " "
echo "=============> Building sym links..."
echo " "

for fff in Dockerfile.testscsd Dockerfile.fake-hsm Dockerfile.fake-rfep Dockerfile.scsd_functest Dockerfile.fake-vault docker-compose-functest.yaml; do
    echo "Linking: ${fff}..."
    ln -s Test/${fff}
done


echo " "
echo "=============> Building docker composed environment..."
echo " "

${DCOMPOSE} build  > ${logfilename}.dcbuild 2>&1
if [[ $? -ne 0 ]]; then
    echo "Docker compose build FAILED, exiting."
    exit 1
fi

${DCOMPOSE} up -d
if [[ $? -ne 0 ]]; then
    ${DCOMPOSE} logs > ${logfilename} 2>&1
    ${DCOMPOSE} down
    echo "Docker compose up FAILED, exiting."
    echo "Logs in ${logfilename}"
    exit 1
fi

container_network=`docker network ls --filter "name=${HSUFFIX}_${brnet_suffix}" --format "{{.Name}}"`
echo "Bridge network name: ${container_network}"
addhosts=`docker network inspect ${container_network} | Test/getnets.py`

if [[ "${addhosts}" == "" ]]; then
    ${DCOMPOSE} logs > ${logfilename} 2>&1
    ${DCOMPOSE} down
    echo "No containers/network data found in docker network for our services, exiting."
    echo "Logs in ${logfilename}"
    cleanup_containers
    exit 1
fi

echo " "
echo "=============> Adding hosts:"
echo $addhosts
echo " "

echo " "
echo "=============> Building test script container..."
echo " "

DOCKER_BUILDKIT=0 docker build --no-cache -f Dockerfile.scsd_functest \
             --tag scsd_functest:runme \
             --network=scsd_${HSUFFIX}_${brnet_suffix} \
             $addhosts \
             --build-arg IN_SCSD=${SCSD_HOST}:${SCSD_PORT} \
             --build-arg IN_HSM=${FAKE_SM_HOST}:${FAKE_SM_PORT} \
             --build-arg IN_X_S0_PORT=${X_S0_PORT} \
             --build-arg IN_X_S1_PORT=${X_S1_PORT} \
             --build-arg IN_X_S2_PORT=${X_S2_PORT} \
             --build-arg IN_X_S3_PORT=${X_S3_PORT} \
             --build-arg IN_X_S6_PORT=${X_S6_PORT} \
             --build-arg IN_X_S7_PORT=${X_S7_PORT} \
             --build-arg SCSD_HOST=${SCSD_HOST} \
             --build-arg FAKE_SM_HOST=${FAKE_SM_HOST} \
             --build-arg VAULT_HOST=${VAULT_HOST} \
             --build-arg X_S0_HOST=${X_S0_HOST} \
             --build-arg X_S1_HOST=${X_S1_HOST} \
             --build-arg X_S2_HOST=${X_S2_HOST} \
             --build-arg X_S3_HOST=${X_S3_HOST} \
             --build-arg X_S6_HOST=${X_S6_HOST} \
             --build-arg X_S7_HOST=${X_S7_HOST} \
             --build-arg SCSD_VERSION=`cat .version` . > ${logfilename}.buildit 2>&1

docker run --rm --attach STDOUT --attach STDERR \
             --network=scsd_${HSUFFIX}_${brnet_suffix} \
             --env IN_SCSD=${SCSD_HOST}:${SCSD_PORT} \
             --env IN_HSM=${FAKE_SM_HOST}:${FAKE_SM_PORT} \
             --env IN_X_S0_PORT=${X_S0_PORT} \
             --env IN_X_S1_PORT=${X_S1_PORT} \
             --env IN_X_S2_PORT=${X_S2_PORT} \
             --env IN_X_S3_PORT=${X_S3_PORT} \
             --env IN_X_S6_PORT=${X_S6_PORT} \
             --env IN_X_S7_PORT=${X_S7_PORT} \
             --env SCSD_HOST=${SCSD_HOST} \
             --env FAKE_SM_HOST=${FAKE_SM_HOST} \
             --env VAULT_HOST=${VAULT_HOST} \
             --env X_S0_HOST=${X_S0_HOST} \
             --env X_S1_HOST=${X_S1_HOST} \
             --env X_S2_HOST=${X_S2_HOST} \
             --env X_S3_HOST=${X_S3_HOST} \
             --env X_S6_HOST=${X_S6_HOST} \
             --env X_S7_HOST=${X_S7_HOST} \
             --env SCSD_VERSION=`cat .version` scsd_functest:runme > ${logfilename}.runit 2>&1

test_rslt=$?

# Shut down, clean up

echo " "
echo "=============> Shutting down container sets..."
echo " "

${DCOMPOSE} logs > ${logfilename} 2>&1
echo "========== Build log ===============" >> ${logfilename}
cat ${logfilename}.dcbuild >> ${logfilename}
cat ${logfilename}.buildit >> ${logfilename}
echo "========== Run log ===============" >> ${logfilename}
cat ${logfilename}.runit >> ${logfilename}
${DCOMPOSE} down
cleanup_containers

echo " "
echo " See ${logfilename} for container set logs."
echo " "

echo " "
echo "================================================="
if [[ ${test_rslt} -ne 0 ]]; then
    echo "SCSD test(s) FAILED."
    echo " "
    echo "LOGS:"
    cat ${logfilename}
    echo " "
else
    echo "SCSD test SUCCESS!"
fi
echo "================================================="

exit ${test_rslt}

