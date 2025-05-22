# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).\

## [1.23.0] - 2025-05-28

### Updated

- Updated image and module dependencies
- Explicitly closed all request and response bodies using hms-base functions
- Internal tracking ticket: CASMHMS-6398

## [1.22.0] - 2025-04-03

### Added

- Added support for ppprof builds

## [1.21.0] - 2025-03-25

### Security

- Updated image and module dependencies for security updates
- Various code changes to accomodate module updates
- Updated Go vo v1.24
- Resolved build warnings in Dockerfiles and docker compose files
- Stopped using legacy Bocker builder
- Internal tracking ticket: CASMHMS-6418

## [1.20.0] - 2023-12-03

### Changed

- Updated go to 1.23

## [1.19.0] - 2023-01-10

### Changed

- Changed to return bmc creds with no vault entry with 404 status code

## [1.18.0] - 2023-01-26

### Fixed

- Language linting of API spec file (no content changes)

## [1.17.0] - 2022-11-14

### Changed

- Added TPM State BIOS support for cray servers

## [1.16.0] - 2022-07-19

### Changed

- Updated CT tests to hms-test:3.2.0 image to pick up Helm test enhancements and CVE fixes.

## [1.15.0] - 2022-07-01

### Changed

- Replaced hsm v1 to hsm v2.

## [1.14.0] - 2022-06-27

### Added

- Added TPM State BIOS interface

## [1.13.0] - 2022-06-23

### Changed

- Updated CT and integration tests to hms-test:3.1.0 image as part of Helm test coordination.
- Replaced fake Vault with real Vault in the integration test environment.

## [1.12.0] - 2022-06-03

### Changed

- Refactored runIntegration.sh and the docker-compose environment for integration testing.
- Replaced fake HSM with real HSM for integration testing.
- Cleaned up old CT test files that are no longer needed.

## [1.11.0] - 2022-05-16

### Changed

- Builds with github actions - removed Jenkins builds

## [1.10.0] - 2021-11-16

### Added

- Added HSM re-discovery after credential changes.

## [1.9.0] - 2021-11-16

### Added

- Added creds-fetch API endpoint.

## [1.8.0] - 2021-10-27

### Added

- CASMHMS-5055 - Added SCSD CT test RPM.

## [1.7.8] - 2021-10-20

### Changed

- Workaround for iLO URI bug with trailing slashes.

## [1.7.7] - 2021-09-22

### Changed

- Changed cray-service version to ~6.0.0

## [1.7.6] - 2021-09-08

### Changed

- Changed the docker image to run as the user nobody

## [1.7.5] - 2021-08-19

### Changed

- Updated hms-base version to v1.15.1

## [1.7.4] - 2021-08-10

### Changed

- Added GitHub config files
- Fixed snyk issue in Dockerfiles

## [1.7.3] - 2021-07-27

### Changed

- Converted stash module references to github

## [1.7.2] - 2021-07-22

### Added

- Github migration

## [1.7.1] - 2021-07-12

### Security

- CASMHMS-4933 - Updated base container images for security updates.

## [1.7.0] - 2021-06-18

### Changed

- Bump minor version for CSM 1.2 release branch

## [1.6.0] - 2021-06-18

### Changed

- Bump minor version for CSM 1.1 release branch

## [1.5.1] - 2021-04-19

### Changed

- Updated Dockerfiles to pull base images from Artifactory instead of DTR.

## [1.5.0] - 2021-02-02

### Changed

- Update Copyright/License and re-vendor go packages.

## [1.4.1] - 2021-01-20

### Added

- Added User-Agent headers to all outbound HTTP reqs.

## [1.4.0] - 2021-01-14

### Changed

- Updated license file.

## [1.3.3] - 2020-11-13

- CASMHMS-4217 - Added final CA bundle configmap handling to Helm chart.

## [1.3.2] - 2020-10-21

- CASMHMS-4105 - Updated base Golang Alpine image to resolve libcrypto vulnerability.

## [1.3.1] - 2020-10-06

- Added APIs and supporting code for TLS cert management.

## [1.3.0] - 2020-09-15
These are changes to charts in support of:

- moving to Helm v1/Loftsman v1
- the newest 2.x cray-service base chart
  - upgraded to support Helm v3
  - modified containers/init containers, volume, and persistent volume claim value definitions to be objects instead of arrays
- the newest 0.2.x cray-jobs base chart
  - upgraded to support Helm v3

## [1.2.1] - 2020-07-21

- CASMHMS-3406 - Updated SCSD build to use trusted baseOS.

## [1.2.0] - 2020-06-30

- CASMHMS-3631 - Updated SCSD CT smoke test with new API test cases.

## [1.1.4] - 2020-06-22

- Fixed Etag formatting/handling, fixed use of the Force flag.

## [1.1.3] - 2020-06-03

- CASMHMS-3535 - Updated SCSD CT tests to pull BMC credentials from new location.

## [1.1.2] - 2020-05-20

- Added initial SCSD smoke, functional, and destructive CT tests.

## [1.1.1]

- Added ability to handle Etags in Redfish account interactions.

## [1.1.0] - 2020-05-11

- Support for online upgrade and rollbacks.  Updated cray-service base chart to 1.4.0

## [1.0.1]

- Added missing /version API.

## [1.0.0]

- Initial implementation.
