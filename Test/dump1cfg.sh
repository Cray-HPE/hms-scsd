#!/bin/bash

# Copyright 2020 Cray, Inc.  All rights reserved

if [ -z $SCSD ]; then
    echo "MISSING SCSD ENV VAR."
    exit 1
fi


# POST to get a dump of current configs

curl -D hout http://${SCSD}/v1/bmc/cfg/x0c0s0b0:${X0C0S0B0_PORT}?params=NTPServerInfo+SyslogServerInfo+SSHKey+SSHConsoleKey
echo " "

cat hout
scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 200 )); then
	echo "Bad status code from single config dump: ${scode}"
	exit 1
fi

exit 0

