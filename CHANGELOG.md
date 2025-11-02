# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.1] - 2025-11-03

### Removed
- **Breaking Change**: Removed `razd task` command in favor of `razd run`
  - The `razd task` command has been removed to simplify the CLI interface
  - Users should now use `razd run <name>` to execute tasks defined in Razdfile.yml
  - This change eliminates confusion between two similar commands and provides a clearer mental model

### Migration Guide
Replace all instances of `razd task` with `razd run`:

**Before (0.4.0)**:
```bash
razd task build
razd task test
razd task dev
```

**After (0.4.1)**:
```bash
razd run build
razd run test
razd run dev
```

### Changed
- Updated error messages and help text to reference `razd run` instead of `razd task`
- Updated all documentation and examples to use the new command

## [0.4.0] - 2025-11-03

### Fixed
- **Critical fix**: Version field injection now works correctly when omitted from Razdfile.yml
  - Fixed bug where Razdfile.yml was passed directly to task command without serialization
  - razd now always serializes configuration through temporary files to ensure version field is injected
  - Users can now reliably omit `version: '3'` from their Razdfile.yml files

### Changed
- Removed unused `has_razdfile_config()` function from taskfile integration
- Workflow execution always uses serialized YAML to maintain consistency

## [0.3.2] - 2025-11-03

### Changed
- Simplified template generation to use a single universal template for all project types
- Removed technology-specific templates (Node.js, Python, Rust, Go)
- All projects now use the same minimal, flexible Razdfile.yml template

### Improved
- Consistent experience across all project types
- Easier maintenance with single template approach
- Users can customize the generated template for their specific needs

## [0.3.1] - 2025-10-30

### Changed
- **Razdfile.yml format simplification**: The `version: '3'` field is now optional
  - Users can omit the version field for cleaner configuration files
  - razd automatically injects `version: '3'` when executing taskfile commands
  - Full backward compatibility: existing Razdfile.yml with explicit version continue to work
  - All templates and examples updated to omit the version field

### Improved
- Cleaner configuration files with less boilerplate
- Better user experience for new projects (via `razd up --init`)
- Maintained compatibility with Taskfile v3 format under the hood

## [0.3.0] - 2025-10-29

### Changed
- **BREAKING**: Backup prompts now ask to opt-out instead of opt-in for safer defaults
  - Old prompt: "Create backup? [Y/n]" (Y creates backup, n/Enter skips)
  - New prompt: "Modify WITHOUT backup? [Y/n]" (Y skips backup, n/Enter creates backup)
  - Pressing Enter or 'n' now creates a backup (safe default)
  - Typing 'Y' explicitly opts out of backup creation
  - Rationale: Makes backup creation the path of least resistance, improving data safety

### Migration Notes
- Users familiar with the old prompts should note the inverted behavior
- The new prompt wording clearly states what 'Y' does ("WITHOUT backup")
- Auto-approve mode continues to create backups by default

## [0.2.7] - 2025-10-29

### Changed
- Task configurations no longer serialize `internal: false` (default value) to YAML
- Cleaner Razdfile.yml output with only non-default values shown
- `internal: true` continues to be serialized for internal tasks

## [0.2.6] - 2025-10-29

### Added
- **Semantic change detection**: Mise sync now ignores formatting-only changes (whitespace, blank lines, comments, key order)
- Only semantic content changes (tool versions, task definitions, plugin URLs) trigger synchronization
- Canonical form serialization for both Razdfile.yml and mise.toml before comparison
- `format_version` field in tracking state for future migration support

### Changed
- File tracking system now uses semantic hashes instead of modification timestamps
- Formatting changes in YAML/TOML files no longer prompt for unnecessary syncs
- More intelligent change detection reduces false positives

### Fixed
- Formatting tools (Prettier, YAML formatters) no longer disrupt mise sync workflow
- Manual whitespace adjustments don't trigger sync prompts

## [0.2.5] - 2025-10-29

### Changed
- **Improved Razdfile.yml structure**: `mise:` section now always appears before `tasks:` section
- **Task ordering**: Tasks are now sorted in a logical order: `default`, `install`, `dev`, `build`, then alphabetically
- Better consistency in YAML serialization using IndexMap for ordered fields
- Enhanced synchronization to preserve preferred task order

### Added
- `indexmap` dependency with serde support for ordered configuration maps

## [0.2.4] - 2025-10-28

### Added
- **Interactive backup prompts**: Ask user before creating backups during mise sync operations
- User confirmation for backup when modifying Razdfile.yml or mise.toml
- Better user control over backup file creation

### Changed
- Backup creation now requires user confirmation in interactive mode (Y/n)
- Auto-approve mode still creates backups automatically without prompts

## [0.2.3] - 2025-10-28

### Added
- **`razd up --init` command**: Initialize new Razdfile.yml with project-specific template
- Automatic project type detection (Node.js, Rust, Python, Go, Generic)
- Template generation based on detected project type with appropriate tasks and commands

### Changed
- Improved user experience for project initialization
- Added helpful next steps after Razdfile.yml creation

## [0.2.2] - 2025-10-28

### Fixed
- **Improved YAML formatting**: Razdfile.yml now has better spacing between sections and tasks
- **Auto-sync mise.toml → Razdfile.yml**: When Razdfile has no mise section but mise.toml exists, automatically syncs configuration with user prompt
- Better prompt messaging for mise configuration sync scenarios

### Changed
- Enhanced YAML output formatting with blank lines between top-level sections and task definitions

## [0.2.1] - 2025-10-28

### Added
- **Mise configuration integration in Razdfile.yml**: Define mise tool versions and plugins directly in Razdfile.yml
- **Bidirectional synchronization**: Automatic sync between Razdfile.yml and mise.toml
- **Conflict detection**: Interactive prompts when both files are modified with diff display
- **Automatic backups**: `.backup` files created before sync operations
- **Global `--no-sync` flag**: Skip synchronization for any command
- **`RAZD_NO_SYNC` environment variable**: Disable sync via environment
- Cross-platform file tracking for modification detection (Windows/Linux/macOS)
- Comprehensive test suite with 123 tests covering all sync scenarios
- Documentation for mise configuration in README and examples

### Changed
- All commands (up, dev, build, task, install, setup) now check for mise sync automatically
- `mise.rs` integration prioritizes Razdfile.yml over standalone mise.toml
- Updated examples to demonstrate mise configuration in Razdfile.yml

### Fixed
- Improved error handling for mise configuration parsing
- Better user feedback during sync operations with clear status messages

## [0.1.3] - 2025-10-23

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