#!/bin/bash

# MIT License
#
# (C) Copyright [2020-2022] Hewlett Packard Enterprise Development LP
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

source portFix.sh

echo "====================================================================="
echo "Group test: global creds, valid group only"
echo "====================================================================="

pldx='{"Force":false, "Username":"root", "Password":"zzaabb", "Targets":["bmcgroup"]}'
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld" http://${SCSD}/v1/bmc/globalcreds | jq > out.txt
cat out.txt
echo " "

scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 200 )); then
	echo "Bad status code from global creds load: ${scode}"
	exit 1
fi

echo "====================================================================="
echo "Group test: global creds, valid group with 4 other targs"
echo "====================================================================="

pldx='{"Force":false, "Username":"root", "Password":"zzaabb", "Targets":["X_S0_HOST","X_S1_HOST","X_S2_HOST","X_S3_HOST","bmcgroup"]}'
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld" http://${SCSD}/v1/bmc/globalcreds | jq > out.txt
cat out.txt
echo " "

scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 200 )); then
	echo "Bad status code from global creds load: ${scode}"
	exit 1
fi

echo "====================================================================="
echo "Group test: global creds, invalid group only"
echo "====================================================================="

pldx='{"Force":false, "Username":"root", "Password":"zzaabb", "Targets":["bmcgroupx"]}'
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld" http://${SCSD}/v1/bmc/globalcreds | jq > out.txt
cat out.txt
echo " "

scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode == 200 )); then
	echo "Bad status code from global creds load (should fail): ${scode}"
	exit 1
fi

echo "====================================================================="
echo "Group test: global creds, bad group with 4 valid targs"
echo "====================================================================="

pldx='{"Force":false, "Username":"root", "Password":"zzaabb", "Targets":["bmcgroupx","X_S0_HOST","X_S1_HOST","X_S2_HOST","X_S3_HOST"]}'
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld" http://${SCSD}/v1/bmc/globalcreds | jq > out.txt
cat out.txt
echo " "

scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 200 )); then
	echo "Bad status code from global creds load: ${scode}"
	exit 1
fi

echo "====================================================================="
echo "Group test: global creds, bad group with 3 valid targs, 1 bad targ"
echo "====================================================================="

pldx='{"Force":false, "Username":"root", "Password":"zzaabb", "Targets":["bmcgroupx","x0c0s0bx","X_S1_HOST","X_S2_HOST","X_S3_HOST"]}'
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld" http://${SCSD}/v1/bmc/globalcreds | jq > out.txt
cat out.txt
echo " "

scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 200 )); then
	echo "Bad status code from global creds load: ${scode}"
	exit 1
fi


echo "====================================================================="
echo "Group test: discreet creds, valid group only (should fail)"
echo "====================================================================="

pldx='{ "Force":false, "Targets": [ { "Xname": "bmcgroup", "Creds": { "Username":"root", "Password":"aaaaaa" } } ] }'
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld" http://${SCSD}/v1/bmc/discreetcreds | jq > out.txt
cat out.txt
echo " "

scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode == 200 )); then
	echo "Bad status code from global creds load (should fail): ${scode}"
	exit 1
fi

echo "====================================================================="
echo "Group test: discreet creds, valid group with 3 valid targs"
echo "====================================================================="

pldx='{ "Force":false, "Targets": [ { "Xname": "bmcgroup", "Creds": { "Username":"root", "Password":"aaaaaa" } }, { "Xname": "X_S1_HOST", "Creds": { "Username":"root", "Password":"bbbbbb" } }, { "Xname": "X_S2_HOST", "Creds": { "Username":"root", "Password":"cccccc" } }, { "Xname": "X_S3_HOST", "Creds": { "Username":"root", "Password":"dddddd" } } ] }'
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld" http://${SCSD}/v1/bmc/discreetcreds | jq > out.txt
cat out.txt
echo " "

scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 200 )); then
	echo "Bad status code from global creds load: ${scode}"
	exit 1
fi

echo "====================================================================="
echo "Group test: discreet creds, invalid group only"
echo "====================================================================="

pldx='{ "Force":false, "Targets": [ { "Xname": "bmcgroupx", "Creds": { "Username":"root", "Password":"aaaaaa" } } ] }'
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld" http://${SCSD}/v1/bmc/discreetcreds | jq > out.txt
cat out.txt
echo " "

scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode == 200 )); then
	echo "Bad status code from global creds load (should fail): ${scode}"
	exit 1
fi

echo "====================================================================="
echo "Group test: discreet creds, bad group with 3 valid targs"
echo "====================================================================="

pldx='{ "Force":false, "Targets": [ { "Xname": "bmcgroupx", "Creds": { "Username":"root", "Password":"aaaaaa" } }, { "Xname": "X_S1_HOST", "Creds": { "Username":"root", "Password":"bbbbbb" } }, { "Xname": "X_S2_HOST", "Creds": { "Username":"root", "Password":"cccccc" } }, { "Xname": "X_S3_HOST", "Creds": { "Username":"root", "Password":"dddddd" } } ] }'
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld" http://${SCSD}/v1/bmc/discreetcreds | jq > out.txt
cat out.txt
echo " "

scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 200 )); then
	echo "Bad status code from global creds load: ${scode}"
	exit 1
fi

echo "====================================================================="
echo "Group test: discreet creds, bad group with 2 valid targs, 1 bad targ (fail)"
echo "====================================================================="

pldx='{ "Force":false, "Targets": [ { "Xname": "X_S0_HOSTx", "Creds": { "Username":"root", "Password":"aaaaaa" } }, { "Xname": "bmcgroupx", "Creds": { "Username":"root", "Password":"bbbbbb" } }, { "Xname": "X_S2_HOST", "Creds": { "Username":"root", "Password":"cccccc" } }, { "Xname": "X_S3_HOST", "Creds": { "Username":"root", "Password":"dddddd" } } ] }'
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld" http://${SCSD}/v1/bmc/discreetcreds | jq > out.txt
cat out.txt
echo " "

scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 200 )); then
	echo "Bad status code from global creds load: ${scode}"
	exit 1
fi

echo "====================================================================="
echo "Group test: load config, valid group only"
echo "====================================================================="

pldx='{"Force":false,"Targets":["bmcgroup"],"Params":{"NTPServerInfo":{"NTPServers":["sms-nnn-www1"],"Port":234,"ProtocolEnabled":true},"SyslogServerInfo":{"SyslogServers":["sms-mmm-yyy1"],"Port":567,"ProtocolEnabled":true},"SSHKey":"aabbccdd","SSHConsoleKey":"eeddffgg"}}'
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld" http://${SCSD}/v1/bmc/loadcfg | jq > out.txt
cat out.txt
echo " "

scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 200 )); then
	echo "Bad status code from global creds load: ${scode}"
	exit 1
fi

echo "====================================================================="
echo "Group test: load config, valid group with other valid targs"
echo "====================================================================="

pldx='{"Force":false,"Targets":["bmcgroup","X_S0_HOST","X_S1_HOST"],"Params":{"NTPServerInfo":{"NTPServers":["sms-nnn-www1"],"Port":234,"ProtocolEnabled":true},"SyslogServerInfo":{"SyslogServers":["sms-mmm-yyy1"],"Port":567,"ProtocolEnabled":true},"SSHKey":"aabbccdd","SSHConsoleKey":"eeddffgg"}}'
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld" http://${SCSD}/v1/bmc/loadcfg | jq > out.txt
cat out.txt
echo " "

scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 200 )); then
	echo "Bad status code from global creds load: ${scode}"
	exit 1
fi

echo "====================================================================="
echo "Group test: load config, bad group only"
echo "====================================================================="

pldx='{"Force":false,"Targets":["bmcgroupx"],"Params":{"NTPServerInfo":{"NTPServers":["sms-nnn-www1"],"Port":234,"ProtocolEnabled":true},"SyslogServerInfo":{"SyslogServers":["sms-mmm-yyy1"],"Port":567,"ProtocolEnabled":true},"SSHKey":"aabbccdd","SSHConsoleKey":"eeddffgg"}}'
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld" http://${SCSD}/v1/bmc/loadcfg | jq > out.txt
cat out.txt
echo " "

scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode == 200 )); then
	echo "Bad status code from global creds load (should fail): ${scode}"
	exit 1
fi

echo "====================================================================="
echo "Group test: load config, bad group with other valid targs"
echo "====================================================================="

pldx='{"Force":false,"Targets":["bmcgroupx","X_S0_HOST","X_S1_HOST"],"Params":{"NTPServerInfo":{"NTPServers":["sms-nnn-www1"],"Port":234,"ProtocolEnabled":true},"SyslogServerInfo":{"SyslogServers":["sms-mmm-yyy1"],"Port":567,"ProtocolEnabled":true},"SSHKey":"aabbccdd","SSHConsoleKey":"eeddffgg"}}'
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld" http://${SCSD}/v1/bmc/loadcfg | jq > out.txt
cat out.txt
echo " "

scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 200 )); then
	echo "Bad status code from global creds load: ${scode}"
	exit 1
fi

echo "====================================================================="
echo "Group test: load config, bad group with bad and valid targs"
echo "====================================================================="

pldx='{"Force":false,"Targets":["bmcgroupx","X_S0_HOST","X_S1_HOSTx"],"Params":{"NTPServerInfo":{"NTPServers":["sms-nnn-www1"],"Port":234,"ProtocolEnabled":true},"SyslogServerInfo":{"SyslogServers":["sms-mmm-yyy1"],"Port":567,"ProtocolEnabled":true},"SSHKey":"aabbccdd","SSHConsoleKey":"eeddffgg"}}'
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld" http://${SCSD}/v1/bmc/loadcfg | jq > out.txt
cat out.txt
echo " "

scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 200 )); then
	echo "Bad status code from global creds load: ${scode}"
	exit 1
fi

echo "====================================================================="
echo "Group test: load config, with bad and valid targs"
echo "====================================================================="

pldx='{"Force":false,"Targets":["X_S0_HOST","X_S1_HOSTx","X_S2_HOST"],"Params":{"NTPServerInfo":{"NTPServers":["sms-nnn-www1"],"Port":234,"ProtocolEnabled":true},"SyslogServerInfo":{"SyslogServers":["sms-mmm-yyy1"],"Port":567,"ProtocolEnabled":true},"SSHKey":"aabbccdd","SSHConsoleKey":"eeddffgg"}}'
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld" http://${SCSD}/v1/bmc/loadcfg | jq > out.txt
cat out.txt
echo " "

scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 200 )); then
	echo "Bad status code from global creds load: ${scode}"
	exit 1
fi

echo "====================================================================="
echo "Group test: dump config, valid group only"
echo "====================================================================="

pldx='{"Force": false, "Targets":["bmcgroup"],"Params":["NTPServerInfo","SyslogServerInfo","SSHKey","SSHConsoleKey"]}'
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld" http://${SCSD}/v1/bmc/dumpcfg | jq > out.txt
cat out.txt
echo " "

scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 200 )); then
	echo "Bad status code from global creds load: ${scode}"
	exit 1
fi

echo "====================================================================="
echo "Group test: load config, valid group with other valid targs"
echo "====================================================================="

pldx='{"Force": false, "Targets":["bmcgroup","X_S0_HOST","X_S1_HOST"],"Params":["NTPServerInfo","SyslogServerInfo","SSHKey","SSHConsoleKey"]}'
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld" http://${SCSD}/v1/bmc/dumpcfg | jq > out.txt
cat out.txt
echo " "


scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 200 )); then
	echo "Bad status code from global creds load: ${scode}"
	exit 1
fi

echo "====================================================================="
echo "Group test: load config, bad group only"
echo "====================================================================="

pldx='{"Force": false, "Targets":["bmcgroupx"],"Params":["NTPServerInfo","SyslogServerInfo","SSHKey","SSHConsoleKey"]}'
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld" http://${SCSD}/v1/bmc/dumpcfg | jq > out.txt
cat out.txt
echo " "

scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode == 200 )); then
	echo "Bad status code from global creds load (should fail): ${scode}"
	exit 1
fi

echo "====================================================================="
echo "Group test: load config, bad group with other valid targs"
echo "====================================================================="

pldx='{"Force": false, "Targets":["bmcgroupx","X_S1_HOST","X_S2_HOST","X_S3_HOST"],"Params":["NTPServerInfo","SyslogServerInfo","SSHKey","SSHConsoleKey"]}'
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld" http://${SCSD}/v1/bmc/dumpcfg | jq > out.txt
cat out.txt
echo " "

scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 200 )); then
	echo "Bad status code from global creds load: ${scode}"
	exit 1
fi

echo "====================================================================="
echo "Group test: load config, bad group with bad and valid targs"
echo "====================================================================="

pldx='{"Force": false, "Targets":["bmcgroupx","X_S1_HOST","x0c0s2bx","X_S3_HOST"],"Params":["NTPServerInfo","SyslogServerInfo","SSHKey","SSHConsoleKey"]}'
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld" http://${SCSD}/v1/bmc/dumpcfg | jq > out.txt
cat out.txt
echo " "

scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 200 )); then
	echo "Bad status code from global creds load: ${scode}"
	exit 1
fi

echo "====================================================================="
echo "Group test: load config, with bad and valid targs"
echo "====================================================================="

pldx='{"Force": false, "Targets":["X_S1_HOST","x0c0s2bx","X_S3_HOST"],"Params":["NTPServerInfo","SyslogServerInfo","SSHKey","SSHConsoleKey"]}'
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld" http://${SCSD}/v1/bmc/dumpcfg | jq > out.txt
cat out.txt
echo " "

scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 200 )); then
	echo "Bad status code from global creds load: ${scode}"
	exit 1
fi

exit 0

