#!/bin/bash

# Copyright 2020 Hewlett Packard Enterprise Development LP

# This script creates 2 cabinet-level certs.

pldx='{"Domain":"Cabinet","DomainIDs":["x0","x1"]}'

source portFix.sh
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld"  http://${SCSD}/v1/bmc/createcerts | jq > out.txt
cat out.txt
echo " "

scode=`cat hout | grep HTTP | awk '{print $2}'`
scode2=`cat out.txt | grep StatusCode | grep -v 200`
if [[ $scode -ne 200 || "${scode2}" != "" ]]; then
	echo "Bad status code from cabinet-domain cert create: ${scode}"
	exit 1
fi

exit 0

