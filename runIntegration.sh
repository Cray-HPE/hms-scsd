#!/bin/bash

#
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
#

# This is a functional test for SCSD.  It will create a composed environment
# containing a fake HSM, 6 fake Redfish endpoints, and SCSD itself.
#
# Once created, a container will be created which contains test scripts.
# The scripts use CURL calls into SCSD to perform the various endpoint 
# operations.
#
# Once complete, the results of the test suite is printed to stdout and this
# script will exit with 0 on success, non-zero on failure.

testtag=scsdtest
tagsuffix=${RANDOM}_${RANDOM}
brnet_suffix=ttest
export PROJ=scsd_${tagsuffix}
export HTAG=${testtag}
export HSUFFIX=${tagsuffix}

# get safe ports to use, avoiding others using them.  This method is not
# perfect but should work 99.9% of the time.

BASEPORT=25100
nrut=`pgrep -c -f runIntegrationTest.sh`
(( PORTBASE = BASEPORT + nrut ))

(( SCSD_PORT     = PORTBASE ))
(( FAKE_SM_PORT  = PORTBASE+10 ))
(( X0C0S0B0_PORT = PORTBASE+100 ))
(( X0C0S1B0_PORT = PORTBASE+101 ))
(( X0C0S2B0_PORT = PORTBASE+102 ))
(( X0C0S3B0_PORT = PORTBASE+103 ))
(( X0C0S6B0_PORT = PORTBASE+106 ))
(( X0C0S7B0_PORT = PORTBASE+107 ))


export SCSD_PORT
export FAKE_SM_PORT
export X0C0S0B0_PORT
export X0C0S1B0_PORT
export X0C0S2B0_PORT
export X0C0S3B0_PORT
export X0C0S6B0_PORT
export X0C0S7B0_PORT

export CRAY_VAULT_JWT_FILE=/tmp/k8stoken
export CRAY_VAULT_ROLE_FILE=/tmp/k8stoken
export SCSD_TEST_K8S_AUTH_URL="http://vault:8200/v1/auth/kubernetes/login"
export SCSD_TEST_VAULT_PKI_URL="http://vault:8200/v1/pki_common/issue/pki-common"
export SCSD_TEST_VAULT_CA_URL="http://vault:8200/v1/pki_common/ca_chain"

logfilename=scsdtest_${PROJ}.logs

cleanup_containers() {
    # Get rid of the interim containers

    echo " "
    echo "=============== > Deleting temporary containers..."
    echo " "

    for fff in `docker images | grep ${HTAG}_${HSUFFIX} | awk '{printf("%s\n",$3)}'`; do
        docker image rm -f ${fff}
    done
}

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

docker build --no-cache -f Dockerfile.scsd_functest \
             --tag scsd_functest:runme \
             --network=scsd_${HSUFFIX}_${brnet_suffix} \
             $addhosts \
             --build-arg IN_SCSD=scsd:${SCSD_PORT} \
             --build-arg IN_HSM=fake_hsm:${FAKE_SM_PORT} \
             --build-arg IN_X0C0S0B0_PORT=${X0C0S0B0_PORT} \
             --build-arg IN_X0C0S1B0_PORT=${X0C0S1B0_PORT} \
             --build-arg IN_X0C0S2B0_PORT=${X0C0S2B0_PORT} \
             --build-arg IN_X0C0S3B0_PORT=${X0C0S3B0_PORT} \
             --build-arg IN_X0C0S6B0_PORT=${X0C0S6B0_PORT} \
             --build-arg IN_X0C0S7B0_PORT=${X0C0S7B0_PORT} \
             --build-arg SCSD_VERSION=`cat .version` . > ${logfilename}.buildit 2>&1

docker run --attach STDOUT --attach STDERR \
             --network=scsd_${HSUFFIX}_${brnet_suffix} \
             --env IN_SCSD=scsd:${SCSD_PORT} \
             --env IN_HSM=fake_hsm:${FAKE_SM_PORT} \
             --env IN_X0C0S0B0_PORT=${X0C0S0B0_PORT} \
             --env IN_X0C0S1B0_PORT=${X0C0S1B0_PORT} \
             --env IN_X0C0S2B0_PORT=${X0C0S2B0_PORT} \
             --env IN_X0C0S3B0_PORT=${X0C0S3B0_PORT} \
             --env IN_X0C0S6B0_PORT=${X0C0S6B0_PORT} \
             --env IN_X0C0S7B0_PORT=${X0C0S7B0_PORT} \
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

