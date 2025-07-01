# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed
- Fixed groups ordering issue in `pocketid_user` resource where Terraform would detect phantom changes when group IDs were in different order. Changed `groups` attribute from List to Set type to properly handle unordered collections.
- Fixed attribute names in example files - changed `given_name`/`family_name` to `first_name`/`last_name` and `enabled` to `disabled` to match actual schema.

## [0.1.0] - 2024-01-09

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

[Unreleased]: https://github.com/Trozz/terraform-provider-pocketid/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/Trozz/terraform-provider-pocketid/releases/tag/v0.1.0