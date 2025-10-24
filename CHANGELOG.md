# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Support for running `razd up` without URL in already-cloned projects
- Project detection validates presence of Razdfile.yml, Taskfile.yml, or mise.toml
- Clear error messages when running `razd up` in non-project directories
- Unit tests for project detection logic
- Integration tests for local project setup

### Changed
- Made URL argument optional for `razd up` command
- Updated help text to reflect both clone and local setup modes
- Improved user experience with consistent "one command setup" workflow

## [0.1.3] - 2025-10-23

### Added
- RELEASE_TOKEN support for GitHub Actions workflows
- Explicit permissions configuration in workflows
- Fallback mechanism for release creation

### Fixed
- GitHub Actions permissions issues for release creation
- Added proper error handling and manual fallback instructions

## [0.1.2] - 2025-10-23

### Fixed
- Simplified release workflow to avoid GitHub API 403 errors
- Removed complex draft/publish flow in favor of direct release creation
- Improved release notes generation

## [0.1.1] - 2025-10-23

### Added
- GitHub Actions CI/CD pipeline with cross-platform builds
- Automated releases for Windows, macOS (Intel & Apple Silicon), and Linux
- Security policy and vulnerability reporting process
- Dependabot configuration for automated dependency updates
- Contributing guidelines and development workflow documentation
- Issue and pull request templates
- Code quality checks (rustfmt, clippy, cargo-audit)
- Test coverage reporting via codecov
- Pre-built binary distribution via GitHub Releases

### Changed
- Enhanced installation documentation with pre-built binary options
- Optimized release profile with LTO for smaller binaries
- Updated README.md with CI/CD badges and status indicators
- Updated MSRV to 1.74.0 for modern dependency compatibility
- Simplified CI matrix to use only stable Rust versions

### Security
- Added cargo-audit security vulnerability scanning
- Implemented SHA256 checksums for release binaries
- Added security policy and responsible disclosure process

### Fixed
- Resolved Cargo.lock version compatibility issues with CI
- Fixed clippy warnings and code formatting issues
- Updated deprecated GitHub Actions to modern versions

## [0.1.0] - TBD

### Added
- Initial release of razd CLI tool
- Basic project setup functionality
- Configuration system with fallback chain
- Integration with git, mise, and taskfile
- Cross-platform support (Windows, macOS, Linux)

[Unreleased]: https://github.com/razd-cli/razd/compare/v0.1.3...HEAD
[0.1.3]: https://github.com/razd-cli/razd/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/razd-cli/razd/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/razd-cli/razd/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/razd-cli/razd/releases/tag/v0.1.0