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

if [ -z $HSM ]; then
    echo "MISSING HSM HOSTNAME:PORT (e.g. fake_hsm_zzzz:27999)"
    exit 1
fi

pldx='{"Components": [ {"ID":"X_S0_HOST","Type":"NodeBMC","State":"On","Flag":"OK"}, {"ID":"X_S1_HOST","Type":"NodeBMC","State":"On","Flag":"OK"}, {"ID":"X_S2_HOST","Type":"NodeBMC","State":"On","Flag":"OK"}, {"ID":"X_S3_HOST","Type":"NodeBMC","State":"On","Flag":"OK"}, {"ID":"X_S4_HOST","Type":"NodeBMC","State":"On","Flag":"OK"}, {"ID":"X_S5_HOST","Type":"NodeBMC","State":"On","Flag":"OK"}, {"ID":"x0c0s6b0","Type":"NodeBMC","State":"On","Flag":"OK"}, {"ID":"x0c0s7b0","Type":"NodeBMC","State":"On","Flag":"OK"}]}'

source portFix.sh
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld" http://${HSM}/hsm/v2/State/Components
echo " "

echo "Components:"
cat hout
scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 204 )); then
    echo "Bad status code from HSM component load: ${scode}"
    exit 1
fi

pldx='{"label":"bmcgroup","description":"group of bmcs","tags":["bmctag"],"members":{"ids":["X_S4_HOST","X_S5_HOST"]}}'
pld=`portFix "$pldx"`

curl -D hout -X POST -d "$pld" http://${HSM}/hsm/v2/groups
echo " "

echo "Groups:"
cat hout

scode=`cat hout | grep HTTP | awk '{print $2}'`
if (( scode != 201 )); then
    echo "Bad status code from HSM group load: ${scode}"
    exit 1
fi

exit 0

