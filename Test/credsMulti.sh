#!/bin/bash

# Copyright 2020 Hewlett Packard Enterprise Development LP

if [ -z $SCSD ]; then
    echo "MISSING SCSD ENV VAR."
    exit 1
fi

pldx='{ "Force":false, "Targets": [ { "Xname": "x0c0s0b0:XP0", "Creds": { "Username":"root", "Password":"aaaaaa" } }, { "Xname": "x0c0s1b0:XP1", "Creds": { "Username":"root", "Password":"bbbbbb" } }, { "Xname": "x0c0s2b0:XP2", "Creds": { "Username":"root", "Password":"cccccc" } }, { "Xname": "x0c0s3b0:XP3", "Creds": { "Username":"root", "Password":"dddddd" } }, { "Xname": "x0c0s6b0:XP6", "Creds": { "Username":"root", "Password":"eeeeee" } } ] }'

source portFix.sh
pld=`portFix "$pldx"`

rm hout
curl -D hout -X POST -d "$pld" http://${SCSD}/v1/bmc/discreetcreds | jq > out.txt
cat out.txt
echo " "

cat hout
scode=`cat hout | grep HTTP | awk '{print $2}'`
scode2=`cat out.txt | grep StatusCode | grep -v 200`
if [[ $scode -ne 200 || "${scode2}" != "" ]]; then
	echo "Bad status code from multi creds load: ${scode}"
	exit 1
fi

exit 0

