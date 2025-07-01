# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Significantly improved test coverage from 17.4% to 39.3%
  - Client package coverage increased from 62.2% to 88.3%
  - Provider package coverage increased from 0% to 93.9%
  - Data sources package coverage increased from 0% to 27.8%
  - Resources package coverage increased from 1.8% to 8.8%
- Comprehensive rate limiting handling with support for Retry-After headers
- Exponential backoff retry mechanism for transient errors
- GitHub workflow to automatically generate CONTRIBUTORS.md file
- Extensive error handling tests for HTTP status codes (400, 401, 403, 404, 409, 429, 500)
- Tests for request timeout and context cancellation scenarios
- Tests for JSON unmarshaling errors and network failures
- Coverage report documentation (COVERAGE_REPORT.md)
- Test analytics support with Codecov integration
  - JUnit XML test result reporting for failed test analysis
  - Enhanced CI workflow with gotestsum for better test output
  - Codecov configuration for test analytics and coverage tracking
  - New make targets: `test-junit` and `test-ci` for JUnit output

### Improved
- Enhanced error handling throughout the client package
- Better handling of edge cases in API responses
- More robust retry logic for API requests

## [0.1.1] - 2025-07-01

### Fixed
- Fixed groups ordering issue in `pocketid_user` resource where Terraform would detect phantom changes when group IDs were in different order. Changed `groups` attribute from List to Set type to properly handle unordered collections.
- Fixed attribute names in example files - changed `given_name`/`family_name` to `first_name`/`last_name` and `enabled` to `disabled` to match actual schema.

### Added
- Added acceptance tests for user resource with groups
- Added CHANGELOG.md to track changes

## [0.1.0] - 2025-07-01

### Added
- Initial release of the Terraform Provider for PocketID
- `pocketid_client` resource for managing OAuth2/OIDC clients
- `pocketid_user` resource for managing users
- `pocketid_group` resource for managing user groups
- `pocketid_client` data source for reading client information
- `pocketid_clients` data source for listing all clients
- `pocketid_user` data source for reading user information
- `pocketid_users` data source for listing and filtering users
- `pocketid_group` data source for reading group information
- `pocketid_groups` data source for listing and filtering groups
- Provider configuration with base URL and API token authentication
- Support for TLS certificate verification skip option
- Comprehensive documentation and examples
- Full test coverage for all resources and data sources
- CI/CD pipeline with GitHub Actions
- GoReleaser configuration for automated releases

[0.1.1]: https://github.com/Trozz/terraform-provider-pocket
