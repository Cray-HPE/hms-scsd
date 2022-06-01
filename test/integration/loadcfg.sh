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

# POST to get a dump of current configs

#TODO: orig
#cray-scsd_1             | time="2022-05-31T21:36:10Z" level=trace msg="  [ 0]: 0xc0002420f0 tg: 'x0c0s480b0:34800', gp: '', gpm: false, st: Unknown, isM: false sc: 0 err: <nil>"
#cray-scsd_1             | time="2022-05-31T21:36:10Z" level=trace msg="  [ 1]: 0xc000242168 tg: 'x0c0s481b0:34801', gp: '', gpm: false, st: Unknown, isM: false sc: 0 err: <nil>"
#cray-scsd_1             | time="2022-05-31T21:36:10Z" level=trace msg="getRvMt(0), targ: 'x0c0s480b0:34800', state: Unknown"
#cray-scsd_1             | time="2022-05-31T21:36:10Z" level=trace msg="getRvMt(0), targ: 'x0c0s481b0:34801', state: Unknown"
#cray-scsd_1             | time="2022-05-31T21:36:10Z" level=error msg="getRvMt(): no valid targets."
pldx='{"Force":false,"Targets":["X_S0_HOST:XP0","X_S1_HOST:XP1"],"Params":{"NTPServerInfo":{"NTPServers":["sms-nnn-www1"],"Port":234,"ProtocolEnabled":true},"SyslogServerInfo":{"SyslogServers":["sms-mmm-yyy1"],"Port":567,"ProtocolEnabled":true},"SSHKey":"aabbccdd","SSHConsoleKey":"eeddffgg"}}'
#TODO: orig without ports
#cray-scsd_1             | time="2022-05-31T21:38:39Z" level=trace msg="CHECKLIST:"
#cray-scsd_1             | time="2022-05-31T21:38:39Z" level=trace msg="  [ 0]: 0xc0002de0f0 tg: 'x0c0s480b0', gp: '', gpm: false, st: On, isM: false sc: 0 err: <nil>"
#cray-scsd_1             | time="2022-05-31T21:38:39Z" level=trace msg="  [ 1]: 0xc0002de168 tg: 'x0c0s481b0', gp: '', gpm: false, st: On, isM: false sc: 0 err: <nil>"
#cray-scsd_1             | time="2022-05-31T21:38:39Z" level=trace msg="getRvMt(0), targ: 'x0c0s480b0', state: On"
#cray-scsd_1             | time="2022-05-31T21:38:39Z" level=trace msg="getRvMt(0), targ: 'x0c0s481b0', state: On"
#cray-scsd_1             | time="2022-05-31T21:38:39Z" level=trace msg="popRFCreds(), Vault is disabled, no creds."
#cray-scsd_1             | time="2022-05-31T21:38:39Z" level=trace msg="popRFCreds(), Vault is disabled, no creds."
#cray-scsd_1             | time="2022-05-31T21:38:39Z" level=trace msg="setting up context for request"
#cray-scsd_1             | time="2022-05-31T21:38:39Z" level=trace msg="setting up context for request"
#cray-scsd_1             | time="2022-05-31T21:38:46Z" level=trace msg="Err: GET https://x0c0s481b0/redfish/v1/ giving up after 4 attempt(s): Get \"https://x0c0s481b0/redfish/v1/\": dial tcp 192.168.128.13:443: connect: connection refused"
#cray-scsd_1             | time="2022-05-31T21:38:46Z" level=trace msg="getStatusCode, no response, err: 'GET https://x0c0s481b0/redfish/v1/ giving up after 4 attempt(s): Get \"https://x0c0s481b0/redfish/v1/\": dial tcp 192.168.128.13:443: connect: connection refused'"
#cray-scsd_1             | time="2022-05-31T21:38:46Z" level=debug msg="Task complete, URL: '/redfish/v1/', status code: 500"
#cray-scsd_1             | time="2022-05-31T21:38:46Z" level=trace msg="Err: GET https://x0c0s480b0/redfish/v1/ giving up after 4 attempt(s): Get \"https://x0c0s480b0/redfish/v1/\": dial tcp 192.168.128.15:443: connect: connection refused"
#cray-scsd_1             | time="2022-05-31T21:38:46Z" level=trace msg="getStatusCode, no response, err: 'GET https://x0c0s480b0/redfish/v1/ giving up after 4 attempt(s): Get \"https://x0c0s480b0/redfish/v1/\": dial tcp 192.168.128.15:443: connect: connection refused'"
#cray-scsd_1             | time="2022-05-31T21:38:46Z" level=debug msg="Task complete, URL: '/redfish/v1/', status code: 500"
#cray-scsd_1             | time="2022-05-31T21:38:46Z" level=trace msg="getStatusCode, no response, err: 'GET https://x0c0s480b0/redfish/v1/ giving up after 4 attempt(s): Get \"https://x0c0s480b0/redfish/v1/\": dial tcp 192.168.128.15:443: connect: connection refused'"
#cray-scsd_1             | time="2022-05-31T21:38:46Z" level=trace msg="getStatusCode, no response, err: 'GET https://x0c0s481b0/redfish/v1/ giving up after 4 attempt(s): Get \"https://x0c0s481b0/redfish/v1/\": dial tcp 192.168.128.13:443: connect: connection refused'"
#cray-scsd_1             | time="2022-05-31T21:38:46Z" level=trace msg="getStatusCode, no response, err: 'GET https://x0c0s480b0/redfish/v1/ giving up after 4 attempt(s): Get \"https://x0c0s480b0/redfish/v1/\": dial tcp 192.168.128.15:443: connect: connection refused'"
#cray-scsd_1             | time="2022-05-31T21:38:46Z" level=trace msg="getStatusCode, no response, err: 'GET https://x0c0s481b0/redfish/v1/ giving up after 4 attempt(s): Get \"https://x0c0s481b0/redfish/v1/\": dial tcp 192.168.128.13:443: connect: connection refused'"
#cray-scsd_1             | time="2022-05-31T21:38:46Z" level=debug msg="doOp(): No tasks to perform, all ignored."
#cray-scsd_1             | time="2022-05-31T21:38:46Z" level=trace msg="setNWP() NWP info: '{\"Oem\":{\"Syslog\":{\"ProtocolEnabled\":true,\"SyslogServers\":[\"sms-mmm-yyy1\"],\"Port\":567},\"SSHAdmin\":{\"AuthorizedKeys\":\"aabbccdd\"},\"SSHConsole\":{\"AuthorizedKeys\":\"eeddffgg\"}},\"NTP\":{\"NTPServers\":[\"sms-nnn-www1\"],\"ProtocolEnabled\":true,\"Port\":234}}'"
#cray-scsd_1             | time="2022-05-31T21:38:46Z" level=error msg="setNWP(): ERROR: No valid targets."
#cray-scsd_1             | time="2022-05-31T21:38:46Z" level=error msg="ERROR: problem loading NWP data: ERROR: No valid targets.\n"
#pldx='{"Force":false,"Targets":["X_S0_HOST","X_S1_HOST"],"Params":{"NTPServerInfo":{"NTPServers":["sms-nnn-www1"],"Port":234,"ProtocolEnabled":true},"SyslogServerInfo":{"SyslogServers":["sms-mmm-yyy1"],"Port":567,"ProtocolEnabled":true},"SSHKey":"aabbccdd","SSHConsoleKey":"eeddffgg"}}'

#TODO: FAIL SCSD loadcfg.sh logs w/ ports
#cray-scsd_1             | time="2022-05-28T01:42:37Z" level=info msg="Service Instance Name: '23195_cray-scsd'"
#cray-scsd_1             | time="2022-05-28T01:42:37Z" level=info msg="Overriding k8s auth url with: 'http://23195_vault_1:8200/v1/auth/kubernetes/login'"
#cray-scsd_1             | time="2022-05-28T01:42:37Z" level=info msg="Overriding PKI url with: 'http://23195_vault_1:8200/v1/pki_common/issue/pki-common'"
#cray-scsd_1             | time="2022-05-28T01:42:37Z" level=info msg="Overriding CA url with: 'http://23195_vault_1:8200/v1/pki_common/ca_chain'"
#cray-scsd_1             | time="2022-05-28T01:42:37Z" level=info msg="Overriding JWT file with: '/tmp/k8stoken'"
#cray-scsd_1             | time="2022-05-28T01:42:37Z" level=info msg="Overriding ROLE file with: '/tmp/k8stoken'"
#cray-scsd_1             | time="2022-05-28T01:42:37Z" level=info msg="23195_cray-scsd UUID: 0"
#cray-scsd_1             | time="2022-05-28T01:42:37Z" level=info msg="HTTP Listen port: 25309"
#cray-scsd_1             | time="2022-05-28T01:42:37Z" level=info msg="HTTP retries:     5"
#cray-scsd_1             | time="2022-05-28T01:42:37Z" level=info msg="Log level:        TRACE"
#cray-scsd_1             | time="2022-05-28T01:42:37Z" level=info msg="TRS mode local:   true"
#cray-scsd_1             | time="2022-05-28T01:42:37Z" level=info msg="TRS kafka URL:    ''"
#cray-scsd_1             | time="2022-05-28T01:42:37Z" level=info msg="Vault enabled:    false"
#cray-scsd_1             | time="2022-05-28T01:42:37Z" level=info msg="Vault keypath:    ''"
#cray-scsd_1             | time="2022-05-28T01:42:37Z" level=info msg="Starting up HTTP server."
#cray-scsd_1             | time="2022-05-28T01:42:55Z" level=trace msg="hsmVerify() xn: 'x0c0s480b0', targ: 'x0c0s480b0:34800'"
#cray-scsd_1             | time="2022-05-28T01:42:55Z" level=trace msg="Added 0xc0000d0000 'x0c0s480b0:34800' to checkList"
#cray-scsd_1             | time="2022-05-28T01:42:55Z" level=trace msg="hsmVerify() xn: 'x0c0s481b0', targ: 'x0c0s481b0:34801'"
#cray-scsd_1             | time="2022-05-28T01:42:55Z" level=trace msg="Added 0xc0000d0078 'x0c0s481b0:34801' to checkList"
#cray-scsd_1             | time="2022-05-28T01:42:55Z" level=trace msg="Mapping checklist member 0: '{x0c0s480b0:34800  false Unknown   false 0 <nil>}'"
#cray-scsd_1             | time="2022-05-28T01:42:55Z" level=trace msg="Mapping checklist member 1: '{x0c0s481b0:34801  false Unknown   false 0 <nil>}'"
#cray-scsd_1             | time="2022-05-28T01:42:55Z" level=trace msg="CHECKLIST:"
#cray-scsd_1             | time="2022-05-28T01:42:55Z" level=trace msg="  [ 0]: 0xc0000d00f0 tg: 'x0c0s480b0:34800', gp: '', gpm: false, st: Unknown, isM: false sc: 0 err: <nil>"
#cray-scsd_1             | time="2022-05-28T01:42:55Z" level=trace msg="  [ 1]: 0xc0000d0168 tg: 'x0c0s481b0:34801', gp: '', gpm: false, st: Unknown, isM: false sc: 0 err: <nil>"
#cray-scsd_1             | time="2022-05-28T01:42:55Z" level=trace msg="getRvMt(0), targ: 'x0c0s480b0:34800', state: Unknown"
#cray-scsd_1             | time="2022-05-28T01:42:55Z" level=trace msg="getRvMt(0), targ: 'x0c0s481b0:34801', state: Unknown"
#cray-scsd_1             | time="2022-05-28T01:42:55Z" level=error msg="getRvMt(): no valid targets."
#cray-scsd_1             | time="2022-05-28T01:42:55Z" level=error msg="ERROR: Problem determining target architectures: ERROR: No valid targets.."

#TODO: PASS SCSD loadcfg.sh logs
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:02Z" level=info msg="Service Instance Name: 'scsd_96'"
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:02Z" level=info msg="Overriding k8s auth url with: 'http://vault_96:8200/v1/auth/kubernetes/login'"
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:02Z" level=info msg="Overriding PKI url with: 'http://vault_96:8200/v1/pki_common/issue/pki-common'"
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:02Z" level=info msg="Overriding CA url with: 'http://vault_96:8200/v1/pki_common/ca_chain'"
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:02Z" level=info msg="Overriding JWT file with: '/tmp/k8stoken'"
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:02Z" level=info msg="Overriding ROLE file with: '/tmp/k8stoken'"
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:02Z" level=info msg="scsd_96     UUID: 0"
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:02Z" level=info msg="HTTP Listen port: :44300"
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:02Z" level=info msg="HTTP retries:     5"
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:02Z" level=info msg="Log level:        TRACE"
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:02Z" level=info msg="TRS mode local:   true"
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:02Z" level=info msg="TRS kafka URL:    ''"
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:02Z" level=info msg="Vault enabled:    false"
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:02Z" level=info msg="Vault keypath:    ''"
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:02Z" level=info msg="Starting up HTTP server."
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:25Z" level=trace msg="hsmVerify() xn: 'x0c0s960b0', targ: 'x0c0s960b0:44400'"
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:25Z" level=trace msg="Added 0xc000238000 'x0c0s960b0:44400' to checkList"
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:25Z" level=trace msg="hsmVerify() xn: 'x0c0s961b0', targ: 'x0c0s961b0:44401'"
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:25Z" level=trace msg="Added 0xc000238078 'x0c0s961b0:44401' to checkList"
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:25Z" level=trace msg="Mapping checklist member 0: '{x0c0s960b0:44400  false Unknown   false 0 <nil>}'"
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:25Z" level=trace msg="Mapping checklist member 1: '{x0c0s961b0:44401  false Unknown   false 0 <nil>}'"
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:25Z" level=trace msg="CHECKLIST:"
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:25Z" level=trace msg="  [ 0]: 0xc0002380f0 tg: 'x0c0s960b0:44400', gp: '', gpm: false, st: On, isM: false sc: 0 err: <nil>"
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:25Z" level=trace msg="  [ 1]: 0xc000238168 tg: 'x0c0s961b0:44401', gp: '', gpm: false, st: On, isM: false sc: 0 err: <nil>"
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:25Z" level=trace msg="getRvMt(0), targ: 'x0c0s960b0:44400', state: On"
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:25Z" level=trace msg="getRvMt(0), targ: 'x0c0s961b0:44401', state: On"
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:25Z" level=trace msg="popRFCreds(), Vault is disabled, no creds."
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:25Z" level=trace msg="popRFCreds(), Vault is disabled, no creds."
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:25Z" level=trace msg="setting up context for request"
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:25Z" level=trace msg="setting up context for request"
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:25Z" level=trace msg="Response: 200"
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:25Z" level=debug msg="Task complete, URL: '/redfish/v1/', status code: 200"
#^[[33mscsd_96   |^[[0m time="2022-05-27T19:02:25Z" level=trace msg="Response: 200"

#TODO: FAIL SCSD loadcfg.sh logs w/o ports
#cray-scsd_1             | time="2022-05-28T01:31:15Z" level=info msg="Service Instance Name: '12811_cray-scsd'"
#cray-scsd_1             | time="2022-05-28T01:31:15Z" level=info msg="Overriding k8s auth url with: 'http://12811_vault_1:8200/v1/auth/kubernetes/login'"
#cray-scsd_1             | time="2022-05-28T01:31:15Z" level=info msg="Overriding PKI url with: 'http://12811_vault_1:8200/v1/pki_common/issue/pki-common'"
#cray-scsd_1             | time="2022-05-28T01:31:15Z" level=info msg="Overriding CA url with: 'http://12811_vault_1:8200/v1/pki_common/ca_chain'"
#cray-scsd_1             | time="2022-05-28T01:31:15Z" level=info msg="Overriding JWT file with: '/tmp/k8stoken'"
#cray-scsd_1             | time="2022-05-28T01:31:15Z" level=info msg="Overriding ROLE file with: '/tmp/k8stoken'"
#cray-scsd_1             | time="2022-05-28T01:31:15Z" level=info msg="12811_cray-scsd UUID: 0"
#cray-scsd_1             | time="2022-05-28T01:31:15Z" level=info msg="HTTP Listen port: 25309"
#cray-scsd_1             | time="2022-05-28T01:31:15Z" level=info msg="HTTP retries:     5"
#cray-scsd_1             | time="2022-05-28T01:31:15Z" level=info msg="Log level:        TRACE"
#cray-scsd_1             | time="2022-05-28T01:31:15Z" level=info msg="TRS mode local:   true"
#cray-scsd_1             | time="2022-05-28T01:31:15Z" level=info msg="TRS kafka URL:    ''"
#cray-scsd_1             | time="2022-05-28T01:31:15Z" level=info msg="Vault enabled:    false"
#cray-scsd_1             | time="2022-05-28T01:31:15Z" level=info msg="Vault keypath:    ''"
#cray-scsd_1             | time="2022-05-28T01:31:15Z" level=info msg="Starting up HTTP server."
#cray-scsd_1             | time="2022-05-28T01:31:22Z" level=trace msg="hsmVerify() xn: 'x0c0s480b0', targ: 'x0c0s480b0'"
#cray-scsd_1             | time="2022-05-28T01:31:22Z" level=trace msg="Added 0xc00032c000 'x0c0s480b0' to checkList"
#cray-scsd_1             | time="2022-05-28T01:31:22Z" level=trace msg="hsmVerify() xn: 'x0c0s481b0', targ: 'x0c0s481b0'"
#cray-scsd_1             | time="2022-05-28T01:31:22Z" level=trace msg="Added 0xc00032c078 'x0c0s481b0' to checkList"
#cray-scsd_1             | time="2022-05-28T01:31:22Z" level=trace msg="Mapping checklist member 0: '{x0c0s480b0  false Unknown   false 0 <nil>}'"
#cray-scsd_1             | time="2022-05-28T01:31:22Z" level=trace msg="Mapping checklist member 1: '{x0c0s481b0  false Unknown   false 0 <nil>}'"
#cray-scsd_1             | time="2022-05-28T01:31:22Z" level=trace msg="CHECKLIST:"
#cray-scsd_1             | time="2022-05-28T01:31:22Z" level=trace msg="  [ 0]: 0xc00032c0f0 tg: 'x0c0s480b0', gp: '', gpm: false, st: On, isM: false sc: 0 err: <nil>"
#cray-scsd_1             | time="2022-05-28T01:31:22Z" level=trace msg="  [ 1]: 0xc00032c168 tg: 'x0c0s481b0', gp: '', gpm: false, st: On, isM: false sc: 0 err: <nil>"
#cray-scsd_1             | time="2022-05-28T01:31:22Z" level=trace msg="getRvMt(0), targ: 'x0c0s480b0', state: On"
#cray-scsd_1             | time="2022-05-28T01:31:22Z" level=trace msg="getRvMt(0), targ: 'x0c0s481b0', state: On"
#cray-scsd_1             | time="2022-05-28T01:31:22Z" level=trace msg="popRFCreds(), Vault is disabled, no creds."
#cray-scsd_1             | time="2022-05-28T01:31:22Z" level=trace msg="popRFCreds(), Vault is disabled, no creds."
#cray-scsd_1             | time="2022-05-28T01:31:22Z" level=trace msg="setting up context for request"
#cray-scsd_1             | time="2022-05-28T01:31:22Z" level=trace msg="setting up context for request"
#cray-scsd_1             | time="2022-05-28T01:31:29Z" level=trace msg="Err: GET https://x0c0s480b0/redfish/v1/ giving up after 4 attempt(s): Get \"https://x0c0s480b0/redfish/v1/\": dial tcp 192.168.224.17:443: connect: connection refused"
#cray-scsd_1             | time="2022-05-28T01:31:29Z" level=trace msg="Err: GET https://x0c0s481b0/redfish/v1/ giving up after 4 attempt(s): Get \"https://x0c0s481b0/redfish/v1/\": dial tcp 192.168.224.16:443: connect: connection refused"
#cray-scsd_1             | time="2022-05-28T01:31:29Z" level=trace msg="getStatusCode, no response, err: 'GET https://x0c0s480b0/redfish/v1/ giving up after 4 attempt(s): Get \"https://x0c0s480b0/redfish/v1/\": dial tcp 192.168.224.17:443: connect: connection refused'"
#cray-scsd_1             | time="2022-05-28T01:31:29Z" level=debug msg="Task complete, URL: '/redfish/v1/', status code: 500"
#cray-scsd_1             | time="2022-05-28T01:31:29Z" level=trace msg="getStatusCode, no response, err: 'GET https://x0c0s481b0/redfish/v1/ giving up after 4 attempt(s): Get \"https://x0c0s481b0/redfish/v1/\": dial tcp 192.168.224.16:443: connect: connection refused'"
#cray-scsd_1             | time="2022-05-28T01:31:29Z" level=debug msg="Task complete, URL: '/redfish/v1/', status code: 500"
#cray-scsd_1             | time="2022-05-28T01:31:29Z" level=trace msg="getStatusCode, no response, err: 'GET https://x0c0s480b0/redfish/v1/ giving up after 4 attempt(s): Get \"https://x0c0s480b0/redfish/v1/\": dial tcp 192.168.224.17:443: connect: connection refused'"
#cray-scsd_1             | time="2022-05-28T01:31:29Z" level=trace msg="getStatusCode, no response, err: 'GET https://x0c0s481b0/redfish/v1/ giving up after 4 attempt(s): Get \"https://x0c0s481b0/redfish/v1/\": dial tcp 192.168.224.16:443: connect: connection refused'"
#cray-scsd_1             | time="2022-05-28T01:31:29Z" level=trace msg="getStatusCode, no response, err: 'GET https://x0c0s480b0/redfish/v1/ giving up after 4 attempt(s): Get \"https://x0c0s480b0/redfish/v1/\": dial tcp 192.168.224.17:443: connect: connection refused'"
#cray-scsd_1             | time="2022-05-28T01:31:29Z" level=trace msg="getStatusCode, no response, err: 'GET https://x0c0s481b0/redfish/v1/ giving up after 4 attempt(s): Get \"https://x0c0s481b0/redfish/v1/\": dial tcp 192.168.224.16:443: connect: connection refused'"
#cray-scsd_1             | time="2022-05-28T01:31:29Z" level=debug msg="doOp(): No tasks to perform, all ignored."
#cray-scsd_1             | time="2022-05-28T01:31:29Z" level=trace msg="setNWP() NWP info: '{\"Oem\":{\"Syslog\":{\"ProtocolEnabled\":true,\"SyslogServers\":[\"sms-mmm-yyy1\"],\"Port\":567},\"SSHAdmin\":{\"AuthorizedKeys\":\"aabbccdd\"},\"SSHConsole\":{\"AuthorizedKeys\":\"eeddffgg\"}},\"NTP\":{\"NTPServers\":[\"sms-nnn-www1\"],\"ProtocolEnabled\":true,\"Port\":234}}'"
#cray-scsd_1             | time="2022-05-28T01:31:29Z" level=error msg="setNWP(): ERROR: No valid targets."
#cray-scsd_1             | time="2022-05-28T01:31:29Z" level=error msg="ERROR: problem loading NWP data: ERROR: No valid targets.\n"

source portFix.sh
pld=$(portFix "$pldx")

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

