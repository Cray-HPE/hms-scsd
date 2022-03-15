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
if [ -z $X_S0_PORT ]; then
    echo "MISSING X_S0_PORT ENV VAR."
    exit 1
fi
if [ -z $X_S1_PORT ]; then
    echo "MISSING X_S1_PORT ENV VAR."
    exit 1
fi
if [ -z $X_S2_PORT ]; then
    echo "MISSING X_S2_PORT ENV VAR."
    exit 1
fi
if [ -z $X_S3_PORT ]; then
    echo "MISSING X_S3_PORT ENV VAR."
    exit 1
fi
if [ -z $X_S6_PORT ]; then
    echo "MISSING X_S6_PORT ENV VAR."
    exit 1
fi
if [ -z $X_S7_PORT ]; then
    echo "MISSING X_S7_PORT ENV VAR."
    exit 1
fi
if [ -z $X_S0_HOST ]; then
    echo "MISSING X_S0_HOST ENV VAR."
    exit 1
fi
if [ -z $X_S1_HOST ]; then
    echo "MISSING X_S1_HOST ENV VAR."
    exit 1
fi
if [ -z $X_S2_HOST ]; then
    echo "MISSING X_S2_HOST ENV VAR."
    exit 1
fi
if [ -z $X_S3_HOST ]; then
    echo "MISSING X_S3_HOST ENV VAR."
    exit 1
fi
if [ -z $X_S6_HOST ]; then
    echo "MISSING X_S6_HOST ENV VAR."
    exit 1
fi
if [ -z $X_S7_HOST ]; then
    echo "MISSING X_S7_HOST ENV VAR."
    exit 1
fi

portFix () {
    local pld
	pld=`echo $1 | sed "s/XP0/$X_S0_PORT/g" | \
                   sed "s/XP1/$X_S1_PORT/g" | \
                   sed "s/XP2/$X_S2_PORT/g" | \
                   sed "s/XP3/$X_S3_PORT/g" | \
                   sed "s/XP6/$X_S6_PORT/g" | \
                   sed "s/XP7/$X_S7_PORT/g" | \
                   sed "s/X_S0_HOST/$X_S0_HOST/g" | \
                   sed "s/X_S1_HOST/$X_S1_HOST/g" | \
                   sed "s/X_S2_HOST/$X_S2_HOST/g" | \
                   sed "s/X_S3_HOST/$X_S3_HOST/g" | \
                   sed "s/X_S6_HOST/$X_S6_HOST/g" | \
                   sed "s/X_S7_HOST/$X_S7_HOST/g" | \
                   sed 's/"/\\"/g'`
    echo $pld
}
