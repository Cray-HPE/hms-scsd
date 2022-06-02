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

# This script applies a cert to 2 different BMCs, one Cray and one HPE.

pldx='{"Force":false,"CertDomain":"Cabinet","Targets":["X_S5_HOST:XP5","X_S1_HOST:XP1"]}'

source portFix.sh
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld"  http://${SCSD}/v1/bmc/setcerts | jq > out.txt
cat out.txt
echo " "

scode=`cat hout | grep HTTP | awk '{print $2}'`
scode2=`cat out.txt | grep StatusCode | grep -v 200`
if [[ $scode -ne 200 || "${scode2}" != "" ]]; then
	echo "Bad status code from BMC cert replace: ${scode}"
	exit 1
fi

exit 0

