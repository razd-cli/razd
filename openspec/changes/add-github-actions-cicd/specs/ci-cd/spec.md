# Spec: CI/CD

Continuous Integration and Continuous Deployment infrastructure for razd.

## ADDED Requirements

### Requirement: Lint Workflow

The CI system SHALL run golangci-lint on every push to main and on pull requests.

#### Scenario: Lint passes on clean code
- **GIVEN** a push to main branch or a pull request
- **WHEN** the lint workflow runs
- **THEN** golangci-lint executes without errors
- **AND** the workflow completes successfully

#### Scenario: Lint fails on problematic code
- **GIVEN** code with linting issues
- **WHEN** the lint workflow runs
- **THEN** golangci-lint reports the issues
- **AND** the workflow fails with appropriate error message

### Requirement: Cross-platform Testing

The CI system SHALL run tests on Linux, macOS, and Windows.

#### Scenario: Tests run on all platforms
- **GIVEN** a push to main branch or a pull request
- **WHEN** the test workflow runs
- **THEN** tests execute on ubuntu-latest, macos-latest, and windows-latest
- **AND** tests pass on all platforms

#### Scenario: Test matrix includes multiple Go versions
- **GIVEN** the test workflow configuration
- **WHEN** tests are triggered
- **THEN** tests run on Go 1.24.x and 1.25.x
- **AND** all matrix combinations are tested

### Requirement: Automated Release

The CI system SHALL create releases when version tags are pushed.

#### Scenario: Release on version tag
- **GIVEN** a tag matching pattern `v*` is pushed
- **WHEN** the release workflow runs
- **THEN** GoReleaser builds binaries for all platforms
- **AND** a GitHub Release is created with artifacts

#### Scenario: Release includes all platforms
- **GIVEN** a release build
- **WHEN** GoReleaser completes
- **THEN** binaries exist for Linux (amd64, arm64)
- **AND** binaries exist for macOS (amd64, arm64)
- **AND** binaries exist for Windows (amd64, arm64)

### Requirement: Version Information

The binary SHALL include version, commit, and build date information.

#### Scenario: Version from release tag
- **GIVEN** a binary built via GoReleaser from tag v1.2.3
- **WHEN** user runs `razd --version`
- **THEN** version displays as "1.2.3"
- **AND** commit hash is included
- **AND** build date is included

#### Scenario: Dev version for local builds
- **GIVEN** a binary built locally without ldflags
- **WHEN** user runs `razd --version`
- **THEN** version displays as "dev"

### Requirement: Checksums

The release SHALL include SHA256 checksums for all artifacts.

#### Scenario: Checksum file generated
- **GIVEN** a release is created
- **WHEN** artifacts are published
- **THEN** a `checksums.txt` file is included
- **AND** it contains SHA256 hashes for all archives
