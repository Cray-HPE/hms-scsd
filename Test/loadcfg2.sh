#!/bin/bash

# Copyright 2020 Hewlett Packard Enterprise Development LP

if [ -z $SCSD ]; then
    echo "MISSING SCSD ENV VAR."
    exit 1
fi


# POST to get a dump of current configs

pldx='{"Force":true,"Targets":["x0c0s0b0:XP0","x0c0s1b0:XP1"],"Params":{"NTPServerInfo":{"NTPServers":["sms-mmm-xxx10"],"Port":345,"ProtocolEnabled":true},"SyslogServerInfo":{"SyslogServers":["sms-ppp-zzz10"],"Port":567,"ProtocolEnabled":true},"SSHKey":"wwwwxxxx","SSHConsoleKey":"yyyyzzzz"}}'

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

