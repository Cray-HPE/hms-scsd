#!/bin/bash

# Copyright 2020 Hewlett Packard Enterprise Development LP


pldx='{"Force":false, "Username":"root", "Password":"zzaabb", "Targets":["x0c0s0b0:XP0","x0c0s1b0:XP1","x0c0s2b0:XP2","x0c0s3b0:XP3"]}'

source portFix.sh
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld"  http://${SCSD}/v1/bmc/globalcreds | jq > out.txt
echo " "

scode=`cat hout | grep HTTP | awk '{print $2}'`
scode2=`cat out.txt | grep StatusCode | grep -v 200`
if [[ $scode -ne 200 || "${scode2}" != "" ]]; then
	echo "Bad status code from global creds load: ${scode}"
	exit 1
fi

exit 0

