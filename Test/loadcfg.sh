#!/bin/bash

# Copyright 2020 Hewlett Packard Enterprise Development LP

if [ -z $SCSD ]; then
    echo "MISSING SCSD ENV VAR."
    exit 1
fi


# POST to get a dump of current configs

pldx='{"Force":false,"Targets":["x0c0s0b0:XP0","x0c0s1b0:XP1"],"Params":{"NTPServerInfo":{"NTPServers":["sms-nnn-www1"],"Port":234,"ProtocolEnabled":true},"SyslogServerInfo":{"SyslogServers":["sms-mmm-yyy1"],"Port":567,"ProtocolEnabled":true},"SSHKey":"aabbccdd","SSHConsoleKey":"eeddffgg"}}'

source portFix.sh
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld" http://${SCSD}/v1/bmc/loadcfg | jq > out.txt
cat out.txt
echo " "

cat hout
scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 200 )); then
	echo "Bad status code from config load: ${scode}"
	exit 1
fi

exit 0

