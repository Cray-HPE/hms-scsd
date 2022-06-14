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

if [ -z $SCSD ]; then
    echo "MISSING SCSD ENV VAR."
    exit 1
fi

# if [ ! -r /var/run/scsd_version.txt ]; then
    # echo "MISSING SCSD VERSION FILE."
    # exit 1
# fi

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

# echo "====================================================================="
# echo "Checking /version"
# echo "====================================================================="

# curl -D hout http://${SCSD}/v1/version > /tmp/out 2>&1
# echo " "
# if [ -f /tmp/out ]; then
    # echo /tmp/out
# fi

# cat hout
# scode=`cat hout | grep HTTP | awk '{print $2}'`
# if (( scode != 200 )); then
    # echo "Bad status code from version check"
    # exit 1
# fi

# ver=`cat /var/run/scsd_version.txt`
# vok=`cat /tmp/out | grep ${ver}`
# vout=`cat /tmp/out`

# if [ "${vok}" == "" ]; then
    # echo "Bad version found in version check"
    # echo "Expecting: ${ver}"
    # echo "Got:       ${vout}"
    # exit 1
# fi
# echo "+++++++++++++++++++++++++++++++++++++++++++"
# echo "VERSION CHECK OK (${vok})"
# echo "+++++++++++++++++++++++++++++++++++++++++++"
# echo " "

exit 0

