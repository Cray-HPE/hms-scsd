#!/bin/bash

# Copyright 2020 Hewlett Packard Enterprise Development LP

# This script applies a cert to 2 different BMCs, one Cray and one HPE.

pldx='{"Force":false,"CertDomain":"Cabinet","Targets":["x0c0s7b0:XP7","x0c0s1b0:XP1"]}'

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

