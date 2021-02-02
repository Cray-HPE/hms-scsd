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

if [ -z $HSM ]; then
    echo "MISSING HSM HOSTNAME:PORT (e.g. fake_hsm_zzzz:27999)"
    exit 1
fi

pldx='{"Components": [ {"ID":"x0c0s0b0:XP0","Type":"NodeBMC","State":"On","Flag":"OK"}, {"ID":"x0c0s1b0:XP1","Type":"NodeBMC","State":"On","Flag":"OK"}, {"ID":"x0c0s2b0:XP2","Type":"NodeBMC","State":"On","Flag":"OK"}, {"ID":"x0c0s3b0:XP3","Type":"NodeBMC","State":"On","Flag":"OK"},{"ID":"x0c0s4b0:XP4","Type":"NodeBMC","State":"On","Flag":"OK"}, {"ID":"x0c0s5b0:XP5","Type":"NodeBMC","State":"On","Flag":"OK"},{"ID":"x0c0s6b0:XP6","Type":"NodeBMC","State":"On","Flag":"OK"},{"ID":"x0c0s7b0:XP7","Type":"NodeBMC","State":"On","Flag":"OK"} ]}'

source portFix.sh
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld" http://${HSM}/hsm/v1/State/Components
echo " "

echo "Components:"
cat hout
scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 200 )); then
	echo "Bad status code from HSM component load: ${scode}"
	exit 1
fi

pldx='[{"label":"bmcgroup","description":"group of bmcs","tags":["bmctag"],"members":{"ids":["x0c0s6b0:XP6","x0c0s7b0:XP7"]}}]'
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld" http://${HSM}/hsm/v1/groups
echo " "

echo "Groups"
cat hout

scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 200 )); then
	echo "Bad status code from HSM group load: ${scode}"
	exit 1
fi

exit 0

