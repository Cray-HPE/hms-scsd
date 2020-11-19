#!/bin/bash

# Copyright 2020 Cray, Inc.  All rights reserved

if [ -z $SCSD ]; then
    echo "MISSING SCSD ENV VAR."
    exit 1
fi

pldx='{"Force":false, "Creds":{"Username":"root", "Password":"zzaabb"}}'

source portFix.sh
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld" http://${SCSD}/v1/bmc/creds/x0c0s0b0:${X0C0S0B0_PORT}
echo " "

scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 200 )); then
	echo "Bad status code from global creds load: ${scode}"
	exit 1
fi

exit 0

