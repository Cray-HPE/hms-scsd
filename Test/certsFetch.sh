#!/bin/bash

# Copyright 2020 Hewlett Packard Enterprise Development LP

# This script fetches certs for 2 BMCs in the x0 domain

pldx='{"Domain":"Cabinet","DomainIDs":["x0c0s0b0","x0c7s7b1"]}'

source portFix.sh
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld"  http://${SCSD}/v1/bmc/fetchcerts | jq > out.txt
cat out.txt
echo " "

scode=`cat hout | grep HTTP | awk '{print $2}'`
scode2=`cat out.txt | grep StatusCode | grep -v 200`
if [[ $scode -ne 200 || "${scode2}" != "" ]]; then
	echo "Bad status code from cert fetch: ${scode}"
	exit 1
fi

crt=`cat out.txt | grep BEGIN`
if [[ "${crt}" == "" ]]; then
	echo "No certficate info seen in fetched data."
	exit 1
fi

exit 0

