#!/usr/bin/env bash

# Copyright 2020 Cray, Inc.  All rights reserved


# Build the build base image
docker build -t cray/hms-scsd-build-base -f Dockerfile.build-base .

docker build -t cray/hms-scsd-testing -f Dockerfile.testing .
