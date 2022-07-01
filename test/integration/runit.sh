#!/bin/bash

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

# ensure the existence of:
#   SCSD
#   HSM
#   x0b0s[0-5]b0:PORT

echo "========= hosts =========="
cat /etc/hosts
echo "========= hosts =========="

if [ -z $SCSD ]; then
    echo "ENV var 'SCSD' not set, exiting."
    exit 1
fi
if [ -z $HSM ]; then
    echo "ENV var 'HSM' not set, exiting."
    exit 1
fi
if [ -z $X_S0_PORT ]; then
    echo "ENV var 'X_S0_PORT' not set, exiting."
    exit 1
fi
if [ -z $X_S1_PORT ]; then
    echo "ENV var 'X_S1_PORT' not set, exiting."
    exit 1
fi
if [ -z $X_S2_PORT ]; then
    echo "ENV var 'X_S2_PORT' not set, exiting."
    exit 1
fi
if [ -z $X_S3_PORT ]; then
    echo "ENV var 'X_S3_PORT' not set, exiting."
    exit 1
fi
if [ -z $X_S4_PORT ]; then
    echo "ENV var 'X_S4_PORT' not set, exiting."
    exit 1
fi
if [ -z $X_S5_PORT ]; then
    echo "ENV var 'X_S5_PORT' not set, exiting."
    exit 1
fi

# Make sure SCSD, HSM, and all fake RF endpoints are running

isReady() {
    url=$1

    for (( i = 0; i < 10; i ++ )); do
        echo "Looking for: ${url}..."
        curl -s $url > /dev/null 2>&1
        if [ $? -eq 0 ]; then
            echo "  ==> ${url} Running!"
            return 1
        fi
        sleep 10
    done

    echo "ERROR, ${url} never started."
    return 0
}

echo "CHECKING FOR HSM..."
isReady http://${HSM}/hsm/v2/State/Components
if [[ $? -ne 1 ]]; then
    echo "Can't continue, exiting."
    exit 1
fi
echo "CHECKING FOR SCSD..."
isReady http://${SCSD}/v1/params
if [[ $? != 1 ]]; then
    echo "Can't continue, exiting."
    exit 1
fi

echo "CHECKING FOR ${X_S0_HOST}..."
isReady http://${X_S0_HOST}:${X_S0_PORT}/redfish/v1/
if [[ $? != 1 ]]; then
    echo "Can't continue, exiting."
    exit 1
fi
echo "CHECKING FOR ${X_S1_HOST}..."
isReady http://${X_S1_HOST}:${X_S1_PORT}/redfish/v1/
if [[ $? != 1 ]]; then
    echo "Can't continue, exiting."
    exit 1
fi
echo "CHECKING FOR ${X_S2_HOST}..."
isReady http://${X_S2_HOST}:${X_S2_PORT}/redfish/v1/
if [[ $? != 1 ]]; then
    echo "Can't continue, exiting."
    exit 1
fi
echo "CHECKING FOR ${X_S3_HOST}..."
isReady http://${X_S3_HOST}:${X_S3_PORT}/redfish/v1/
if [[ $? != 1 ]]; then
    echo "Can't continue, exiting."
    exit 1
fi
echo "CHECKING FOR ${X_S4_HOST}..."
isReady http://${X_S4_HOST}:${X_S4_PORT}/redfish/v1/
if [[ $? != 1 ]]; then
    echo "Can't continue, exiting."
    exit 1
fi
echo "CHECKING FOR ${X_S5_HOST}..."
isReady http://${X_S5_HOST}:${X_S5_PORT}/redfish/v1/
if [[ $? != 1 ]]; then
    echo "Can't continue, exiting."
    exit 1
fi

echo "##################################"
echo "Loading HSM data."
echo "##################################"

hsmLoad.sh
if [ $? -ne 0 ]; then
    echo "Error loading HSM data with hsmLoad.sh."
    exit 1
fi
sleep 2

echo "###############################################"
echo "Liveness, readiness, health, version tests."
echo "###############################################"

health.sh
if [ $? -ne 0 ]; then
    echo "Error running health.sh."
    exit 1
fi

echo "##################################"
echo "Loading config data."
echo "##################################"

loadcfg.sh
if [ $? -ne 0 ]; then
    echo "Error loading config data with loadcfg.sh."
    exit 1
fi

loadcfg2.sh
if [ $? -ne 0 ]; then
    echo "Error loading config data with loadcfg2.sh."
    exit 1
fi

echo "##################################"
echo "Reading config data."
echo "##################################"

dumpcfg.sh
if [ $? -ne 0 ]; then
    echo "Error reading config data with dumpcfg.sh."
    exit 1
fi

echo "##################################"
echo "Loading single config data."
echo "##################################"

load1cfg.sh
if [ $? -ne 0 ]; then
    echo "Error reading config data with load1cfg.sh."
    exit 1
fi

echo "##################################"
echo "Dumping single config data."
echo "##################################"

dump1cfg.sh
if [ $? -ne 0 ]; then
    echo "Error reading config data with dump1cfg.sh."
    exit 1
fi

echo "##################################"
echo "Reading global config data."
echo "##################################"

dumpcfgGP.sh
if [ $? -ne 0 ]; then
    echo "Error reading global config data with dumpcfgGP.sh."
    exit 1
fi

echo "##################################"
echo "Loading multi creds."
echo "##################################"

credsMulti.sh
if [ $? -ne 0 ]; then
    echo "Error loading multi creds with credsMulti.sh."
    exit 1
fi

echo "##################################"
echo "Loading global creds."
echo "##################################"

credsGlobal.sh
if [ $? -ne 0 ]; then
    echo "Error loading global creds data with credsGlobal.sh."
    exit 1
fi

echo "##################################"
echo "Loading single cred."
echo "##################################"

credsSingle.sh
if [ $? -ne 0 ]; then
    echo "Error loading global creds data with credsSingle.sh."
    exit 1
fi

echo "##################################"
echo "Create cab-domain certs."
echo "##################################"

certsCreateCab.sh
if [ $? -ne 0 ]; then
    echo "Error creating cab-domain cert data with certsCreateCab.sh."
    exit 1
fi

echo "##################################"
echo "Create BMC list of cab-domain certs."
echo "##################################"

certsCreateList.sh
if [ $? -ne 0 ]; then
    echo "Error creating BMC list of cab-domain cert data with certsCreateList.sh."
    exit 1
fi

echo "##################################"
echo "Delete cab-domain cert."
echo "##################################"

certsDelete.sh
if [ $? -ne 0 ]; then
    echo "Error deleting cab-domain cert data with certsDelete.sh."
    exit 1
fi

echo "##################################"
echo "Fetch cab-domain certs."
echo "##################################"

certsFetch.sh
if [ $? -ne 0 ]; then
    echo "Error fetching cab-domain cert data with certsFetch.sh."
    exit 1
fi

echo "##################################"
echo "Replace BMC certs."
echo "##################################"

certsRFPost.sh
if [ $? -ne 0 ]; then
    echo "Error replacing BMC certs with certsRFPost.sh."
    exit 1
fi

echo "##################################"
echo "Replace single BMC cert."
echo "##################################"

certsRFPostSingle.sh
if [ $? -ne 0 ]; then
    echo "Error replacing single BMC cert with certsRFPostSingle.sh."
    exit 1
fi

echo "##################################"
echo "Group tests."
echo "##################################"

groupTest.sh
if [ $? -ne 0 ]; then
    echo "Error running groupTest.sh."
    exit 1
fi

exit 0

