# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).\

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

