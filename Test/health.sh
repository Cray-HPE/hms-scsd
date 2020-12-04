#!/bin/bash

# Copyright 2020 Cray, Inc.  All rights reserved

if [ -z $SCSD ]; then
    echo "MISSING SCSD ENV VAR."
    exit 1
fi

# GET /liveness

echo "====================================================================="
echo "Checking /liveness"
echo "====================================================================="

curl -D hout http://${SCSD}/v1/liveness
echo " "

cat hout
scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 204 )); then
    echo "Bad status code from liveness check"
    exit 1
fi
echo "+++++++++++++++++++++"
echo "LIVENESS CHECK OK"
echo "+++++++++++++++++++++"
echo " "

# GET /readiness

echo "====================================================================="
echo "Checking readiness"
echo "====================================================================="

curl -D hout http://${SCSD}/v1/readiness
echo " "

cat hout
scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 204 )); then
    echo "Bad status code from readiness check"
    exit 1
fi
echo "+++++++++++++++++++++"
echo "READINESS CHECK OK"
echo "+++++++++++++++++++++"
echo " "

# Get /version

echo "====================================================================="
echo "Checking /version"
echo "====================================================================="

curl -D hout http://${SCSD}/v1/version > /tmp/out 2>&1
echo " "
if [ -f /tmp/out ]; then
    echo /tmp/out
fi

cat hout
scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 200 )); then
    echo "Bad status code from version check"
    exit 1
fi

ver=`cat /var/run/scsd_version.txt`
vok=`cat /tmp/out | grep ${ver}`
vout=`cat /tmp/out`

if [ "${vok}" == "" ]; then
    echo "Bad version found in version check"
    echo "Expecting: ${ver}"
    echo "Got:       ${vout}"
    exit 1
fi
echo "+++++++++++++++++++++++++++++++++++++++++++"
echo "VERSION CHECK OK (${vok})"
echo "+++++++++++++++++++++++++++++++++++++++++++"
echo " "

exit 0

