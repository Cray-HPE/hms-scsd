NAME ?= hms-scsd
VERSION ?= $(shell cat .version)

all: image unittest coverage integration

image:
	docker build --pull ${DOCKER_ARGS} --tag '${NAME}:${VERSION}' .

unittest:
	./runUnitTest.sh

coverage:
	./runCoverage.sh

integration:
	./runIntegration.sh

buildbase:
	docker build -t cray/hms-scsd-build-base -f Dockerfile.build-base .

