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

