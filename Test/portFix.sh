#!/bin/bash

# Copyright 2020 Cray, Inc.  All rights reserved

if [ -z $SCSD ]; then
    echo "MISSING SCSD ENV VAR."
    exit 1
fi
if [ -z $X0C0S0B0_PORT ]; then
    echo "MISSING X0C0S0B0_PORT ENV VAR."
    exit 1
fi
if [ -z $X0C0S1B0_PORT ]; then
    echo "MISSING X0C0S1B0_PORT ENV VAR."
    exit 1
fi
if [ -z $X0C0S2B0_PORT ]; then
    echo "MISSING X0C0S2B0_PORT ENV VAR."
    exit 1
fi
if [ -z $X0C0S3B0_PORT ]; then
    echo "MISSING X0C0S3B0_PORT ENV VAR."
    exit 1
fi
if [ -z $X0C0S6B0_PORT ]; then
    echo "MISSING X0C0S6B0_PORT ENV VAR."
    exit 1
fi
if [ -z $X0C0S7B0_PORT ]; then
    echo "MISSING X0C0S7B0_PORT ENV VAR."
    exit 1
fi

portFix () {
    local pld
	pld=`echo $1 | sed "s/XP0/$X0C0S0B0_PORT/g" | \
                   sed "s/XP1/$X0C0S1B0_PORT/g" | \
                   sed "s/XP2/$X0C0S2B0_PORT/g" | \
                   sed "s/XP3/$X0C0S3B0_PORT/g" | \
                   sed "s/XP6/$X0C0S6B0_PORT/g" | \
                   sed "s/XP7/$X0C0S7B0_PORT/g" | \
                   sed 's/"/\\"/g'`
    echo $pld
}

