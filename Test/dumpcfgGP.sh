#!/bin/bash

# Copyright 2020 Hewlett Packard Enterprise Development LP

if [ -z $SCSD ]; then
    echo "MISSING SCSD ENV VAR."
    exit 1
fi

pldx='{"Force": false, "Targets":["bmcgroup","x0c0s1b0:XP1","x0c0s2b0:XP2","x0c0s3b0:XP3"],"Params":["NTPServerInfo","SyslogServerInfo","SSHKey","SSHConsoleKey"]}'

source portFix.sh
pld=`portFix "$pldx"`

# POST to get a dump of current configs

curl -D hout -X POST -d "$pld" http://${SCSD}/v1/bmc/dumpcfg | jq > out.txt
cat out.txt
echo " "

cat hout
scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 200 )); then
	echo "Bad status code from HSM global config dump: ${scode}"
	exit 1
fi

exit 0

