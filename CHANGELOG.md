# Changelog

All notable changes to kubectl-reboot will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.3.0] - 2025-09-25

### Added
- Changelog extraction script (`scripts/extract-changelog.sh`) for automated release notes
- Version-specific release notes generation from CHANGELOG.md

### Changed
- GitHub Actions release workflow now extracts detailed release notes from CHANGELOG.md instead of generating generic "Initial release" messages
- Release notes now show organized feature lists with Added/Changed/Fixed categories
- Improved release documentation with proper versioning

### Fixed
- Resolved "Initial release" issue in GitHub releases by implementing proper changelog parsing
- Enhanced release workflow to provide meaningful, detailed release notes

## [1.2.0] - 2025-09-25

### Added
- Enhanced GitHub Actions CI/CD pipelines with comprehensive testing
- Security scanning with Gosec and Trivy
- Multi-platform testing (Ubuntu, macOS)
- Comprehensive lint configuration with golangci-lint
- Automated release scripts for streamlined releases
- Improved Krew plugin manifest with detailed descriptions

### Changed
- Updated GitHub Actions workflows to use latest action versions
- Enhanced Krew manifest with better descriptions and caveats
- Improved error handling and logging throughout codebase
- Simplified logging function parameters and updated mock public key verification method
- Set GOFLAGS to disable VCS information in CI workflows and updated golangci-lint configuration

### Fixed
- Resolved potential security vulnerabilities identified by static analysis
- Fixed formatting issues across all Go source files

## [1.1.0] - 2024-09-24

### Changed
- Set GOFLAGS to disable VCS information in CI workflows
- Updated golangci-lint configuration for better code quality
- Improved build process with disabled VCS information

### Fixed
- Streamlined CI workflows by removing redundant steps
- Enhanced build output formatting

## [1.0.0] - 2024-09-24

### Added
- Initial release of kubectl-reboot plugin
- Safe node restart functionality with cordon/drain/reboot/uncordon workflow
- SSH integration for remote node rebooting
- Boot ID monitoring for reboot verification
- Support for cluster-wide node operations
- Dry-run mode for safe planning
- Extensive configuration options:
  - Custom SSH users and connection options
  - Configurable reboot commands
  - Adjustable timeouts and polling intervals
  - Node filtering and exclusion options
- Multiple installation methods:
  - Krew plugin manager (recommended)
  - Direct binary download
  - Build from source
- Comprehensive documentation and examples
- Support for multiple platforms:
  - Linux (amd64, arm64)
  - macOS (amd64, arm64)
- Cloud provider examples for AWS EKS, Google GKE, and Azure AKS
- Rich emoji-based logging for better user experience
- RBAC permission examples and security considerations

### Features
- 🔄 **Safe Node Restart**: Automated workflow ensuring minimal disruption
- 🚀 **SSH Integration**: Flexible SSH configuration with custom commands
- 🔍 **Reboot Verification**: Boot ID change monitoring for reliability
- 🌐 **Cluster Operations**: Support for single nodes or entire clusters
- 🧪 **Dry-run Mode**: Preview operations without making changes
- ⚡ **Configuration**: Extensive customization for various environments
- 📋 **Logging**: Detailed, colorized output with progress indicators

### Documentation
- Complete README with installation and usage instructions
- Cloud provider specific examples and configurations
- Troubleshooting guide for common issues
- Security best practices and RBAC examples
- Contributing guidelines for developers
- Comprehensive help system with examples

[Unreleased]: https://github.com/ayetkin/kubectl-reboot/compare/v1.3.0...HEAD
[1.3.0]: https://github.com/ayetkin/kubectl-reboot/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/ayetkin/kubectl-reboot/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/ayetkin/kubectl-reboot/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/ayetkin/kubectl-reboot/releases/tag/v1.0.0