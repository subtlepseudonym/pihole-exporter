# Changelog
## [0.0.9] - 2022-11-04
### Added
- Include static curl binary in docker image for healthchecks
- Default tag value when vtag not available

### Fixed
- Fix numeric bounds on queryType

## [0.0.8] - 2022-02-02
### Added
- Support query types MX, DS, RRSIG, DNSKEY, NS, OTHER, SVCB, HTTPS

### Fixed
- Prevent panic on unknown status codes
- Fixed changelog

## [0.0.7] - 2022-02-01
### Added
- Support status codes 12, 13, and 14

## [0.0.6] - 2020-10-28
### Fixed
- Prevent counting requests for all time on restart

## [0.0.5] - 2020-10-28
### Added
- Handler label to http request duration

### Changed
- Change http request duration to counter from gauge

## [0.0.4] - 2020-07-13
### Fixed
- Removed dockerfile binary compression

## [0.0.3] - 2020-07-14
### Added
- Binary compression in docker image
- Changelog
- mattn/go-sqlite3 dependency
- Add stat for metrics handler duration

### Changed
- Use pihole database rather than API
- Alter exposed metrics to better fit database source

## [0.0.2] - 2020-07-07
### Changed
- Reduce unnecessary logging
- Stop building API token into docker image

## [0.0.1] - 2020-07-07
### Added
- Counters for "daily" metrics
- Gauge vectors for "top" metrics
- Docker image
