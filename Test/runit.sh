#!/bin/bash

# Copyright 2020 Hewlett Packard Enterprise Development LP

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
if [ -z $X0C0S0B0_PORT ]; then
    echo "ENV var 'X0C0S0B0_PORT' not set, exiting."
    exit 1
fi
if [ -z $X0C0S1B0_PORT ]; then
    echo "ENV var 'X0C0S1B0_PORT' not set, exiting."
    exit 1
fi
if [ -z $X0C0S2B0_PORT ]; then
    echo "ENV var 'X0C0S2B0_PORT' not set, exiting."
    exit 1
fi
if [ -z $X0C0S3B0_PORT ]; then
    echo "ENV var 'X0C0S3B0_PORT' not set, exiting."
    exit 1
fi
if [ -z $X0C0S6B0_PORT ]; then
    echo "ENV var 'X0C0S6B0_PORT' not set, exiting."
    exit 1
fi
if [ -z $X0C0S7B0_PORT ]; then
    echo "ENV var 'X0C0S7B0_PORT' not set, exiting."
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

echo "CHECKING FOR x0c0s0b0..."
isReady http://x0c0s0b0:${X0C0S0B0_PORT}/redfish/v1/
if [[ $? != 1 ]]; then
    echo "Can't continue, exiting."
    exit 1
fi
echo "CHECKING FOR x0c0s1b0..."
isReady http://x0c0s1b0:${X0C0S1B0_PORT}/redfish/v1/
if [[ $? != 1 ]]; then
    echo "Can't continue, exiting."
    exit 1
fi
echo "CHECKING FOR x0c0s2b0..."
isReady http://x0c0s2b0:${X0C0S2B0_PORT}/redfish/v1/
if [[ $? != 1 ]]; then
    echo "Can't continue, exiting."
    exit 1
fi
echo "CHECKING FOR x0c0s3b0..."
isReady http://x0c0s3b0:${X0C0S3B0_PORT}/redfish/v1/
if [[ $? != 1 ]]; then
    echo "Can't continue, exiting."
    exit 1
fi
echo "CHECKING FOR x0c0s6b0..."
isReady http://x0c0s6b0:${X0C0S6B0_PORT}/redfish/v1/
if [[ $? != 1 ]]; then
    echo "Can't continue, exiting."
    exit 1
fi
echo "CHECKING FOR x0c0s7b0..."
isReady http://x0c0s7b0:${X0C0S7B0_PORT}/redfish/v1/
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

