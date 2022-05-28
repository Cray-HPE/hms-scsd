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

#TODO: orig
#pldx='{"Components": [ {"ID":"X_S0_HOST:XP0","Type":"NodeBMC","State":"On","Flag":"OK"}, {"ID":"X_S1_HOST:XP1","Type":"NodeBMC","State":"On","Flag":"OK"}, {"ID":"X_S2_HOST:XP2","Type":"NodeBMC","State":"On","Flag":"OK"}, {"ID":"X_S3_HOST:XP3","Type":"NodeBMC","State":"On","Flag":"OK"},{"ID":"x0c0s4b0:XP4","Type":"NodeBMC","State":"On","Flag":"OK"}, {"ID":"x0c0s5b0:XP5","Type":"NodeBMC","State":"On","Flag":"OK"},{"ID":"X_S6_HOST:XP6","Type":"NodeBMC","State":"On","Flag":"OK"},{"ID":"X_S7_HOST:XP7","Type":"NodeBMC","State":"On","Flag":"OK"} ]}'
#TODO: orig but no ports
pldx='{"Components": [ {"ID":"X_S0_HOST","Type":"NodeBMC","State":"On","Flag":"OK"}, {"ID":"X_S1_HOST","Type":"NodeBMC","State":"On","Flag":"OK"}, {"ID":"X_S2_HOST","Type":"NodeBMC","State":"On","Flag":"OK"}, {"ID":"X_S3_HOST","Type":"NodeBMC","State":"On","Flag":"OK"},{"ID":"x0c0s4b0","Type":"NodeBMC","State":"On","Flag":"OK"}, {"ID":"x0c0s5b0","Type":"NodeBMC","State":"On","Flag":"OK"},{"ID":"X_S6_HOST","Type":"NodeBMC","State":"On","Flag":"OK"},{"ID":"X_S7_HOST","Type":"NodeBMC","State":"On","Flag":"OK"} ]}'
#TODO: no ports, xnames
#pldx='{"Components": [ {"ID":"X_S0_XNAME","Type":"NodeBMC","State":"On","Flag":"OK"}, {"ID":"X_S1_XNAME","Type":"NodeBMC","State":"On","Flag":"OK"}, {"ID":"X_S2_XNAME","Type":"NodeBMC","State":"On","Flag":"OK"}, {"ID":"X_S3_XNAME","Type":"NodeBMC","State":"On","Flag":"OK"},{"ID":"x0c0s4b0","Type":"NodeBMC","State":"On","Flag":"OK"}, {"ID":"x0c0s5b0","Type":"NodeBMC","State":"On","Flag":"OK"},{"ID":"X_S6_XNAME","Type":"NodeBMC","State":"On","Flag":"OK"},{"ID":"X_S7_XNAME","Type":"NodeBMC","State":"On","Flag":"OK"} ]}'
#TODO: orig but HOST switched to XNAME
#pldx='{"Components": [ {"ID":"X_S0_XNAME:XP0","Type":"NodeBMC","State":"On","Flag":"OK"}, {"ID":"X_S1_XNAME:XP1","Type":"NodeBMC","State":"On","Flag":"OK"}, {"ID":"X_S2_XNAME:XP2","Type":"NodeBMC","State":"On","Flag":"OK"}, {"ID":"X_S3_XNAME:XP3","Type":"NodeBMC","State":"On","Flag":"OK"},{"ID":"x0c0s4b0:XP4","Type":"NodeBMC","State":"On","Flag":"OK"}, {"ID":"x0c0s5b0:XP5","Type":"NodeBMC","State":"On","Flag":"OK"},{"ID":"X_S6_XNAME:XP6","Type":"NodeBMC","State":"On","Flag":"OK"},{"ID":"X_S7_XNAME:XP7","Type":"NodeBMC","State":"On","Flag":"OK"} ]}'

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

#TODO
#pldx='[{"label":"bmcgroup","description":"group of bmcs","tags":["bmctag"],"members":{"ids":["X_S6_HOST:XP6","X_S7_HOST:XP7"]}}]'
pldx='{"label":"bmcgroup","description":"group of bmcs","tags":["bmctag"],"members":{"ids":["X_S6_HOST","X_S7_HOST"]}}'
#pldx='{"label":"bmcgroup","description":"group of bmcs","tags":["bmctag"],"members":{"ids":["X_S6_XNAME","X_S7_XNAME"]}}'
#pldx='{"label":"bmcgroup","description":"group of bmcs","tags":["bmctag"],"members":{"ids":["X_S6_XNAME:XP6","X_S7_XNAME:XP7"]}}'
pld=`portFix "$pldx"`

#TODO
echo "pld=${pld}"

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
