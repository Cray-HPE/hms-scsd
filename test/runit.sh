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

# insure the existence of:
#   SCSD
#   HSM
#   X0B0S[0-7]B0_PORT

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
if [ -z $X_S6_PORT ]; then
    echo "ENV var 'X_S6_PORT' not set, exiting."
    exit 1
fi
if [ -z $X_S7_PORT ]; then
    echo "ENV var 'X_S7_PORT' not set, exiting."
    exit 1
fi

# Make sure HSM is running and all the fake RF endpoints

isReady () {
    url=$1

    for (( i = 0; i < 10; i ++ )); do
        echo "Looking for: ${url}..."
        curl -s $url > /dev/null 2>&1
        #curl -D hout  $url 
        if [ $? -eq 0 ]; then
            echo "  ==> ${url} Running!"
            return 1
        fi
        #cat hout
        sleep 10
    done

    echo "ERROR, ${url} never started."
    return 0
}

echo "CHECKING FOR HSM..."
isReady http://${HSM}/hsm/v1/State/Components
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
echo "CHECKING FOR ${X_S6_HOST}..."
isReady http://${X_S6_HOST}:${X_S6_PORT}/redfish/v1/
if [[ $? != 1 ]]; then
    echo "Can't continue, exiting."
    exit 1
fi
echo "CHECKING FOR ${X_S7_HOST}..."
isReady http://${X_S7_HOST}:${X_S7_PORT}/redfish/v1/
if [[ $? != 1 ]]; then
    echo "Can't continue, exiting."
    exit 1
fi


echo "##################################"
echo "Loading HSM data"
echo "##################################"
hsmLoad.sh
if [ $? -ne 0 ]; then
    echo "Error loading HSM data."
	exit 1
fi
sleep 2



echo "##################################"
echo "Loading config data."
echo "##################################"

loadcfg.sh
if [ $? -ne 0 ]; then
    echo "Error loading config data."
	exit 1
fi

loadcfg2.sh
if [ $? -ne 0 ]; then
    echo "Error loading config data."
	exit 1
fi

echo "##################################"
echo "Reading config data."
echo "##################################"

dumpcfg.sh
if [ $? -ne 0 ]; then
    echo "Error reading config data."
	exit 1
fi

echo "##################################"
echo "Loading single config data."
echo "##################################"

load1cfg.sh
if [ $? -ne 0 ]; then
    echo "Error reading config data."
	exit 1
fi

echo "##################################"
echo "Dumping single config data."
echo "##################################"

dump1cfg.sh
if [ $? -ne 0 ]; then
    echo "Error reading config data."
	exit 1
fi

echo "##################################"
echo "Reading global config data."
echo "##################################"

dumpcfgGP.sh
if [ $? -ne 0 ]; then
    echo "Error reading global config data."
	exit 1
fi

echo "##################################"
echo "Loading multi creds."
echo "##################################"

credsMulti.sh
if [ $? -ne 0 ]; then
    echo "Error loading multi creds."
	exit 1
fi

echo "##################################"
echo "Loading global creds."
echo "##################################"

credsGlobal.sh
if [ $? -ne 0 ]; then
    echo "Error loading global creds data."
	exit 1
fi

echo "##################################"
echo "Loading single cred."
echo "##################################"

credsSingle.sh
if [ $? -ne 0 ]; then
    echo "Error loading global creds data."
	exit 1
fi

echo "##################################"
echo "Create Cab-domain Certs"
echo "##################################"

certsCreateCab.sh
if [ $? -ne 0 ]; then
    echo "Error creating cab-domain cert data."
	exit 1
fi

echo "##################################"
echo "Create BMC list of Cab-domain Certs"
echo "##################################"

certsCreateList.sh
if [ $? -ne 0 ]; then
    echo "Error creating BMC list of cab-domain cert data."
	exit 1
fi

echo "##################################"
echo "Delete Cab-domain Cert"
echo "##################################"

certsDelete.sh
if [ $? -ne 0 ]; then
    echo "Error deleting cab-domain cert data."
	exit 1
fi

echo "##################################"
echo "Fetch Cab-domain Certs"
echo "##################################"

certsFetch.sh
if [ $? -ne 0 ]; then
    echo "Error fetching cab-domain cert data."
	exit 1
fi

echo "##################################"
echo "Replace BMC Certs"
echo "##################################"

certsRFPost.sh
if [ $? -ne 0 ]; then
    echo "Error replacing BMC certs."
	exit 1
fi

echo "##################################"
echo "Replace single BMC Cert"
echo "##################################"

certsRFPostSingle.sh
if [ $? -ne 0 ]; then
    echo "Error replacing single BMC cert."
	exit 1
fi



echo "##################################"
echo "Group tests."
echo "##################################"

groupTest.sh
if [ $? -ne 0 ]; then
    echo "Error running group tests."
	exit 1
fi

echo "###############################################"
echo "Liveness, readiness, health, version tests."
echo "###############################################"

health.sh
if [ $? -ne 0 ]; then
    echo "Error running health tests."
	exit 1
fi

exit 0

