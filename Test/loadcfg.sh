#!/bin/bash

# MIT License
#
# (C) Copyright [2020-2021] Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.

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

