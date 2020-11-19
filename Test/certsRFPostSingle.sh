#!/bin/bash

# Copyright 2020 Hewlett Packard Enterprise Development LP

# This script applies a cert to a BMC using the {xname} in URL.

source portFix.sh

curl -D hout -X POST  http://${SCSD}/v1/bmc/setcert/x0c0s0b0:${X0C0S0B0_PORT}?Force=false\&Domain=cabinet | jq > out.txt
cat out.txt
echo " "

scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 200 )); then
	echo "Bad status code from BMC single cert replace: ${scode}"
	exit 1
fi

exit 0

