# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.6.0] - 2025-11-08

### Added
- **CLI Arguments Support**: Added support for passing arguments to tasks via `--` separator
  - Use `razd run <task> -- <arguments>` to pass arguments to tasks
  - Arguments are available in tasks via `{{.CLI_ARGS}}` template variable
  - Works seamlessly with taskfile.dev's standard CLI_ARGS behavior
  - Example: `razd run test -- -v -race` executes task with `-v -race` arguments
  - Full support for arguments with spaces and special characters
  - Empty arguments handled gracefully when no `--` separator provided

## [0.5.4] - 2025-11-08

### Added
- **Automation support**: Added `-y, --yes` flag for unattended execution
  - Automatically answers "yes" to all interactive prompts
  - Enables CI/CD pipeline integration without manual intervention
  - Auto-approves Razdfile.yml creation in `razd up` command
  - Auto-resolves mise sync conflicts (prefers Razdfile.yml as source of truth)
  - Automatically creates backups during sync operations when needed
  - Works as global flag with all commands: `razd --yes up`, `razd -y list`, etc.
  - Default behavior (without flag) remains unchanged - all prompts work as before

## [0.5.3] - 2025-11-08

### Added
- **Custom configuration file path support**: Added `--taskfile` and `--razdfile` flags (with short form `-t`)
  - Both flags work as synonyms, with `--razdfile` taking priority if both specified
  - Global flags work with all commands that use configuration (list, run, up, setup)
  - Support for both relative and absolute file paths
  - Full backward compatibility - default to `Razdfile.yml` when no flag specified
  - Clear error messages when specified file not found

## [0.5.2] - 2025-11-08

### Added
- **Enhanced JSON output** for `razd list --json` command with taskfile-compatible format
  - Added `task` field (mirrors `name` for taskfile compatibility)
  - Added `summary` field (empty string, placeholder for future feature)
  - Added `aliases` field (empty array, placeholder for future alias support)
  - Added `location` object with `taskfile`, `line`, and `column` for each task
  - Added root `location` field with absolute path to Razdfile.yml
  - Internal tasks now omit `internal: false` from JSON for cleaner output
  - Full backward compatibility maintained - all existing fields preserved

## [0.5.1] - 2025-11-08

### Fixed
- Fixed clippy warning about overly complex boolean expressions in unit tests

## [0.5.0] - 2025-11-08

### Added
- **Task listing enhancements**: Added `--list-all` and `--json` flags to `razd list` command
  - `--list-all` flag shows all tasks including those marked as internal
  - `--json` flag outputs task information in JSON format for scripting and automation
  - Flags can be combined: `razd list --list-all --json` shows all tasks in JSON format
  - Maintains full backward compatibility - `razd list` works exactly as before
  - Added 9 new tests (5 unit tests, 4 integration tests) to ensure reliability

### Documentation
- Added OpenSpec proposal for surgical mise.toml synchronization to preserve platform-specific commands

## [0.4.12] - 2025-11-05

### Fixed
- **Critical**: Task commands now execute in project directory instead of temp directory
  - Added `--dir` flag to task command invocation to ensure correct working directory
  - Fixes issue where file operations (cp, mv, etc.) failed because task was running in temp directory
  - Maintains temp file location in system temp directory while executing commands in project directory

## [0.4.11] - 2025-11-05

### Changed
- **Optimized temporary workflow file cleanup**: Temporary files are now deleted within ~100ms after spawning the task process, instead of waiting for task completion
  - Reduces disk usage during long-running tasks (dev servers, builds)
  - Prevents accumulation of temp files when running multiple workflows simultaneously
  - Files are now created in system temp directory (`$TEMP` on Windows, `/tmp` on Unix)
  - Added 100ms spawn delay to ensure reliable file loading across different system loads
  - Direct task execution uses early cleanup; mise exec fallback maintains traditional cleanup for compatibility

### Technical
- Refactored process execution module to separate spawn and wait operations
- Added `spawn_command()` and `wait_for_command()` functions for better control over process lifecycle
- Introduced `DEFAULT_SPAWN_DELAY_MS` constant for configurable spawn delay

## [0.4.10] - 2025-11-05

### Added
- `razd list` command to display all available tasks with descriptions
- `razd run --list` flag to list tasks instead of running
- `razd --list` global flag for task listing
- Task list output with proper alignment and alphabetical sorting
- Automatic filtering of internal tasks from list view

### Changed
- `razd run` task_name is now optional when using `--list` flag

## [0.4.9] - 2025-11-04

### Fixed
- Code formatting issues (cargo fmt)

## [0.4.8] - 2025-11-04

### Added
- Full Taskfile v3 schema support (same as 0.4.7 but with proper release)

## [0.4.7] - 2025-11-04

### Added
- **Full Taskfile v3 Schema Support**: Extended `Razdfile.yml` to support complete Taskfile v3 specification
  - `env:` - Environment variables at root and task level (supports `sh:` dynamic values)
  - `vars:` - Variables at root and task level (supports `sh:`, `ref:`, map types)
  - `deps:` - Task dependencies (simple: `- task-name` or complex: `{task, vars, silent}`)
  - `platforms:` - Platform-specific command filtering (e.g., `[windows, linux]`)
  - `silent:` - Suppress command output
  - Complex `cmd:` syntax with `cmd`, `platforms`, `silent`, `ignore_error`, `set`, `shopt` fields
  - Task reference syntax: `{task: other-task, vars: {KEY: value}, silent: true}`
  - See [Taskfile v3 Schema](https://taskfile.dev/docs/reference/schema) for full specification

### Changed
- Reverted temporary workflow files to project directory with immediate cleanup
  - Temporary `.razd-workflow-{task_name}.yml` files are now created in project directory
  - Files are deleted immediately after task process loads them (typically within milliseconds)
  - Added `.razd-workflow-*.yml` to `.gitignore` for version control safety

### Technical
- Used `serde_yaml::Value` for dynamic variable support (`sh:`, `ref:`, map types)
- All new fields are `Option<T>` with `skip_serializing_if` for backward compatibility
- Extended `Command` enum to 3 variants: `String`, `Complex`, `TaskRef`
- Added `Dependency` enum: `Simple(String)`, `Complex{task, vars, silent}`
- `IndexMap` for ordered variable maps
  - Added `.razd-workflow-*.yml` pattern to `.gitignore` to prevent Git tracking
  - Removed `--dir` parameter from task command as it's no longer needed
  - This approach keeps Git status clean while allowing proper variable resolution in workflows

### Fixed
- Environment variables and other Taskfile features now work correctly
  - Task CLI loads the workflow file immediately upon start, allowing safe deletion
  - Variables defined in `env:` and `vars:` sections are properly accessible
  - Working directory is correctly set to project root

## [0.4.6] - 2025-11-04

### Fixed
- Fixed workflow execution in CI/CD environments
  - Added `--dir` parameter to `task` command to explicitly set working directory
  - Ensures commands in workflows execute in project directory, not temp directory
  - Fixes "file not found" errors for project files like `package.json` in CI/CD

## [0.4.5] - 2025-11-03

### Fixed
- Moved temporary workflow files to system temp directory
  - Temporary `.razd-workflow-{task_name}.yml` files are now created in system temp directory instead of project directory
  - Prevents temporary files from appearing in Git status during workflow execution
  - Follows OS best practices for temporary file management
  - No changes needed to project `.gitignore` files

## [0.4.4] - 2025-11-03

### Fixed
- Fixed `mise trust` prompt not showing during interactive setup
  - Changed `mise install` and `mise trust` to use interactive execution mode
  - Users can now see and respond to mise trust prompts during `razd up`
  - Ensures proper PTY/TTY handling for all interactive mise commands
  
### Changed
- Removed non-existent `--interactive` flag from task command invocation
- Interactive mode now properly forwards stdin/stdout for all subcommands in workflow

## [0.4.3] - 2025-11-03

### Fixed
- Fixed interactive command execution hanging on Linux
  - Switched from `tokio::process::Command` to `std::process::Command` for interactive commands
  - This resolves PTY/TTY handling issues that caused commands like `npm install` to hang on Linux
  - Commands now properly inherit stdio and work correctly across all platforms

## [0.4.2] - 2025-11-03

### Changed
- Patch release with minor updates

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