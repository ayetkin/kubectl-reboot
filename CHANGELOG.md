# Changelog

All notable changes to kubectl-reboot will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

### Fixed
- Resolved potential security vulnerabilities identified by static analysis
- Fixed formatting issues across all Go source files

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

[Unreleased]: https://github.com/ayetkin/kubectl-reboot/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/ayetkin/kubectl-reboot/releases/tag/v1.0.0