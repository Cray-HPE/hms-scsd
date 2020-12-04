#!/bin/bash

# Copyright 2020 Cray, Inc.  All rights reserved

if [ -z $SCSD ]; then
    echo "MISSING SCSD ENV VAR."
    exit 1
fi


# POST to load a single target's configs

curl -D hout -X POST -d '{"Force":false,"Params":{"NTPServerInfo":{"NTPServers":["sms-nnn-www1"],"Port":234,"ProtocolEnabled":true},"SyslogServerInfo":{"SyslogServers":["sms-mmm-yyy1"],"Port":567,"ProtocolEnabled":true},"SSHKey":"aabbccdd","SSHConsoleKey":"eeddffgg"}}' http://${SCSD}/v1/bmc/cfg/x0c0s0b0:${X0C0S0B0_PORT}
echo " "

cat hout
scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 200 )); then
	echo "Bad status code from single config load: ${scode}"
	exit 1
fi

exit 0

