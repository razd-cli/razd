# Tasks: Add --taskfile Flag Support

## Phase 1: CLI Interface (4 tasks)

### T1.1: Add global flags to Cli struct ✅ completed
**Description**: Add `--taskfile` and `--razdfile` global flags to the main `Cli` struct in `src/main.rs`
- Add `taskfile: Option<String>` field with `#[arg(short = 't', long, global = true)]`
- Add `razdfile: Option<String>` field with `#[arg(long, global = true)]`
- Add help text for both flags

**Validation**: `cargo build` succeeds, `razd --help` shows new flags

**Dependencies**: None

---

### T1.2: Implement config path resolution helper ✅ completed
**Description**: Create helper function to resolve final config path based on flag priority
- Add `resolve_config_path(cli: &Cli) -> Option<PathBuf>` function
- Implement priority logic: `razdfile` > `taskfile` > `None`
- Handle empty string cases

**Validation**: Unit test for priority logic

**Dependencies**: T1.1

---

### T1.3: Update command invocations to pass custom path ✅ completed
**Description**: Modify all command calls in `main.rs` to pass resolved config path
- Update `list` command invocation
- Update `run` command invocation
- Update `setup` command invocation
- Update `up` command invocation

**Validation**: `cargo build` succeeds (may have compile errors in commands)

**Dependencies**: T1.2

---

### T1.4: Add CLI flag parsing tests ✅ completed
**Description**: Add unit tests for CLI argument parsing
- Test `--taskfile` flag parsing
- Test `--razdfile` flag parsing
- Test short form `-t` parsing
- Test priority when both flags specified
- Test with various path formats

**Validation**: `cargo test` for new tests passes

**Dependencies**: T1.3

---

## Phase 2: Configuration Layer (4 tasks)

### T2.1: Add load_with_path method to RazdfileConfig ✅ completed
**Description**: Extend `RazdfileConfig` with custom path support
- Add `pub fn load_with_path(custom_path: Option<PathBuf>) -> Result<Option<Self>>`
- Implement path resolution logic (custom vs default)
- Add error handling for custom path not found

**Validation**: Unit test for path resolution

**Dependencies**: None (can parallelize with Phase 1)

---

### T2.2: Update existing load() method ✅ completed
**Description**: Refactor existing `load()` to use `load_with_path(None)`
- Replace `load()` implementation with call to `load_with_path(None)`
- Ensure backward compatibility

**Validation**: Existing tests still pass

**Dependencies**: T2.1

---

### T2.3: Add config path validation ✅ completed
**Description**: Implement path validation and error messages
- Check file existence for custom paths
- Provide clear error messages with file path
- Handle relative vs absolute paths correctly

**Validation**: Test error messages for various invalid paths

**Dependencies**: T2.1

---

### T2.4: Add configuration layer tests ✅ completed
**Description**: Add comprehensive tests for config loading
- Test loading from custom path
- Test loading from default path
- Test custom path not found error
- Test invalid YAML in custom file
- Test relative and absolute paths
- Test cross-platform paths (Windows/Unix)

**Validation**: `cargo test` shows all config tests passing

**Dependencies**: T2.3

---

## Phase 3: Commands Update (6 tasks)

### T3.1: Update list command signature ✅ completed
**Description**: Add custom_path parameter to list command
- Update `pub async fn execute(list_all: bool, json: bool, custom_path: Option<PathBuf>)`
- Replace `RazdfileConfig::load()` with `RazdfileConfig::load_with_path(custom_path)`
- Update error messages to show actual path used

**Validation**: `cargo build` for list command

**Dependencies**: T1.3, T2.1

---

### T3.2: Update run command signature ✅ completed
**Description**: Add custom_path parameter to run command
- Update `pub async fn execute(task_name: &str, args: &[String], custom_path: Option<PathBuf>)`
- Update config loading logic
- Update error messages

**Validation**: `cargo build` for run command

**Dependencies**: T1.3, T2.1

---

### T3.3: Update setup command signature ✅ completed
**Description**: Add custom_path parameter to setup command
- Update `pub async fn execute(custom_path: Option<PathBuf>)`
- Update config loading logic
- Update error messages

**Validation**: `cargo build` for setup command

**Dependencies**: T1.3, T2.1

---

### T3.4: Update up command signature ✅ completed
**Description**: Add custom_path parameter to up command
- Update `pub async fn execute(url: Option<String>, name: Option<String>, init: bool, custom_path: Option<PathBuf>)`
- Update config loading logic
- Update error messages

**Validation**: `cargo build` for up command

**Dependencies**: T1.3, T2.1

---

### T3.5: Update taskfile integration for custom paths ✅ completed
**Description**: Pass custom path to taskfile commands when specified
- Update `execute_workflow_task_interactive` to accept custom path
- Add `--taskfile` flag to task command invocations when custom path provided
- Ensure path is properly formatted for taskfile

**Validation**: Manual test with `task` command

**Dependencies**: T3.2

---

### T3.6: Update mise integration for custom paths ✅ completed
**Description**: Update mise integration to check custom config paths
- Update `has_mise_config()` to check custom path location
- Update sync logic to handle custom paths
- Ensure mise.toml sync works with custom Razdfile locations

**Validation**: Manual test with mise sync

**Dependencies**: T3.4

---

## Phase 4: Integration Testing (5 tasks)

### T4.1: Add list command integration tests ✅ completed
**Description**: Test list command with custom paths
- Test `razd list --taskfile custom.yml`
- Test `razd list --razdfile custom.yml`
- Test `razd list -t custom.yml`
- Test error when custom file not found
- Test JSON output with custom path

**Validation**: All integration tests pass

**Dependencies**: T3.1

---

### T4.2: Add run command integration tests ✅ completed
**Description**: Test run command with custom paths
- Test `razd run build --taskfile custom.yml`
- Test `razd run build --razdfile custom.yml`
- Test task execution from custom file
- Test error messages

**Validation**: All integration tests pass

**Dependencies**: T3.2

---

### T4.3: Add up command integration tests ✅ completed
**Description**: Test up command with custom paths
- Test local setup with custom taskfile
- Test init with custom filename
- Test error cases

**Validation**: All integration tests pass

**Dependencies**: T3.4

---

### T4.4: Test flag priority behavior ✅ completed
**Description**: Integration tests for flag priority
- Test `--razdfile` overrides `--taskfile`
- Test short form `-t` works correctly
- Test default behavior when no flags specified

**Validation**: All priority tests pass

**Dependencies**: T4.1, T4.2, T4.3

---

### T4.5: Cross-platform path testing ✅ completed
**Description**: Test path handling on Windows and Unix
- Test relative paths (./config.yml)
- Test absolute paths (C:\path\config.yml on Windows, /path/config.yml on Unix)
- Test paths with spaces
- Test paths with special characters

**Validation**: Tests pass on both Windows and Unix CI

**Dependencies**: T4.4

---

## Phase 5: Documentation and Validation (4 tasks)

### T5.1: Update CLI documentation ✅ completed
**Description**: Update help text and documentation
- Update command help messages
- Ensure flag descriptions are clear
- Add examples to help text

**Validation**: `razd --help` shows accurate information

**Dependencies**: T4.5

---

### T5.2: Update CHANGELOG ✅ completed
**Description**: Document new feature in CHANGELOG.md
- Add entry in [Unreleased] section
- Document both `--taskfile` and `--razdfile` flags
- Include usage examples
- Note backward compatibility

**Validation**: CHANGELOG follows Keep a Changelog format

**Dependencies**: T5.1

---

### T5.3: Update spec deltas ✅ completed
**Description**: Finalize spec changes for cli-interface
- Ensure all scenarios are accurate
- Update requirement descriptions
- Cross-reference related specs

**Validation**: Specs match implementation

**Dependencies**: T5.2

---

### T5.4: Run OpenSpec validation ✅ completed
**Description**: Validate proposal with openspec tool
- Run `openspec validate add-taskfile-flag --strict`
- Fix any validation errors
- Ensure all requirements have scenarios

**Validation**: `openspec validate` passes with no errors

**Dependencies**: T5.3

---

## Summary

**Total tasks**: 23
- Phase 1 (CLI Interface): 4 tasks ✅
- Phase 2 (Configuration Layer): 4 tasks ✅
- Phase 3 (Commands Update): 6 tasks ✅
- Phase 4 (Integration Testing): 5 tasks ✅
- Phase 5 (Documentation and Validation): 4 tasks ✅

**All tasks completed** ✅

**Parallelizable work**:
- Phase 1 and Phase 2 can be done in parallel
- Tests in Phase 4 can be done in parallel after their dependencies

**Critical path**: T1.1 → T1.2 → T1.3 → T3.1/T3.2/T3.3/T3.4 → T4.1-T4.5 → T5.4

**Estimated effort**: Medium (touches CLI, config, and all commands but changes are straightforward)
**Description**: Add `--taskfile` and `--razdfile` global flags to the main `Cli` struct in `src/main.rs`
- Add `taskfile: Option<String>` field with `#[arg(short = 't', long, global = true)]`
- Add `razdfile: Option<String>` field with `#[arg(long, global = true)]`
- Add help text for both flags

**Validation**: `cargo build` succeeds, `razd --help` shows new flags

**Dependencies**: None

---

### T1.2: Implement config path resolution helper ⬜ not-started
**Description**: Create helper function to resolve final config path based on flag priority
- Add `resolve_config_path(cli: &Cli) -> Option<PathBuf>` function
- Implement priority logic: `razdfile` > `taskfile` > `None`
- Handle empty string cases

**Validation**: Unit test for priority logic

**Dependencies**: T1.1

---

### T1.3: Update command invocations to pass custom path ⬜ not-started
**Description**: Modify all command calls in `main.rs` to pass resolved config path
- Update `list` command invocation
- Update `run` command invocation
- Update `setup` command invocation
- Update `up` command invocation

**Validation**: `cargo build` succeeds (may have compile errors in commands)

**Dependencies**: T1.2

---

### T1.4: Add CLI flag parsing tests ⬜ not-started
**Description**: Add unit tests for CLI argument parsing
- Test `--taskfile` flag parsing
- Test `--razdfile` flag parsing
- Test short form `-t` parsing
- Test priority when both flags specified
- Test with various path formats

**Validation**: `cargo test` for new tests passes

**Dependencies**: T1.3

---

## Phase 2: Configuration Layer (4 tasks)

### T2.1: Add load_with_path method to RazdfileConfig ⬜ not-started
**Description**: Extend `RazdfileConfig` with custom path support
- Add `pub fn load_with_path(custom_path: Option<PathBuf>) -> Result<Option<Self>>`
- Implement path resolution logic (custom vs default)
- Add error handling for custom path not found

**Validation**: Unit test for path resolution

**Dependencies**: None (can parallelize with Phase 1)

---

### T2.2: Update existing load() method ⬜ not-started
**Description**: Refactor existing `load()` to use `load_with_path(None)`
- Replace `load()` implementation with call to `load_with_path(None)`
- Ensure backward compatibility

**Validation**: Existing tests still pass

**Dependencies**: T2.1

---

### T2.3: Add config path validation ⬜ not-started
**Description**: Implement path validation and error messages
- Check file existence for custom paths
- Provide clear error messages with file path
- Handle relative vs absolute paths correctly

**Validation**: Test error messages for various invalid paths

**Dependencies**: T2.1

---

### T2.4: Add configuration layer tests ⬜ not-started
**Description**: Add comprehensive tests for config loading
- Test loading from custom path
- Test loading from default path
- Test custom path not found error
- Test invalid YAML in custom file
- Test relative and absolute paths
- Test cross-platform paths (Windows/Unix)

**Validation**: `cargo test` shows all config tests passing

**Dependencies**: T2.3

---

## Phase 3: Commands Update (6 tasks)

### T3.1: Update list command signature ⬜ not-started
**Description**: Add custom_path parameter to list command
- Update `pub async fn execute(list_all: bool, json: bool, custom_path: Option<PathBuf>)`
- Replace `RazdfileConfig::load()` with `RazdfileConfig::load_with_path(custom_path)`
- Update error messages to show actual path used

**Validation**: `cargo build` for list command

**Dependencies**: T1.3, T2.1

---

### T3.2: Update run command signature ⬜ not-started
**Description**: Add custom_path parameter to run command
- Update `pub async fn execute(task_name: &str, args: &[String], custom_path: Option<PathBuf>)`
- Update config loading logic
- Update error messages

**Validation**: `cargo build` for run command

**Dependencies**: T1.3, T2.1

---

### T3.3: Update setup command signature ⬜ not-started
**Description**: Add custom_path parameter to setup command
- Update `pub async fn execute(custom_path: Option<PathBuf>)`
- Update config loading logic
- Update error messages

**Validation**: `cargo build` for setup command

**Dependencies**: T1.3, T2.1

---

### T3.4: Update up command signature ⬜ not-started
**Description**: Add custom_path parameter to up command
- Update `pub async fn execute(url: Option<String>, name: Option<String>, init: bool, custom_path: Option<PathBuf>)`
- Update config loading logic
- Update error messages

**Validation**: `cargo build` for up command

**Dependencies**: T1.3, T2.1

---

### T3.5: Update taskfile integration for custom paths ⬜ not-started
**Description**: Pass custom path to taskfile commands when specified
- Update `execute_workflow_task_interactive` to accept custom path
- Add `--taskfile` flag to task command invocations when custom path provided
- Ensure path is properly formatted for taskfile

**Validation**: Manual test with `task` command

**Dependencies**: T3.2

---

### T3.6: Update mise integration for custom paths ⬜ not-started
**Description**: Update mise integration to check custom config paths
- Update `has_mise_config()` to check custom path location
- Update sync logic to handle custom paths
- Ensure mise.toml sync works with custom Razdfile locations

**Validation**: Manual test with mise sync

**Dependencies**: T3.4

---

## Phase 4: Integration Testing (5 tasks)

### T4.1: Add list command integration tests ⬜ not-started
**Description**: Test list command with custom paths
- Test `razd list --taskfile custom.yml`
- Test `razd list --razdfile custom.yml`
- Test `razd list -t custom.yml`
- Test error when custom file not found
- Test JSON output with custom path

**Validation**: All integration tests pass

**Dependencies**: T3.1

---

### T4.2: Add run command integration tests ⬜ not-started
**Description**: Test run command with custom paths
- Test `razd run build --taskfile custom.yml`
- Test `razd run build --razdfile custom.yml`
- Test task execution from custom file
- Test error messages

**Validation**: All integration tests pass

**Dependencies**: T3.2

---

### T4.3: Add up command integration tests ⬜ not-started
**Description**: Test up command with custom paths
- Test local setup with custom taskfile
- Test init with custom filename
- Test error cases

**Validation**: All integration tests pass

**Dependencies**: T3.4

---

### T4.4: Test flag priority behavior ⬜ not-started
**Description**: Integration tests for flag priority
- Test `--razdfile` overrides `--taskfile`
- Test short form `-t` works correctly
- Test default behavior when no flags specified

**Validation**: All priority tests pass

**Dependencies**: T4.1, T4.2, T4.3

---

### T4.5: Cross-platform path testing ⬜ not-started
**Description**: Test path handling on Windows and Unix
- Test relative paths (./config.yml)
- Test absolute paths (C:\path\config.yml on Windows, /path/config.yml on Unix)
- Test paths with spaces
- Test paths with special characters

**Validation**: Tests pass on both Windows and Unix CI

**Dependencies**: T4.4

---

## Phase 5: Documentation and Validation (4 tasks)

### T5.1: Update CLI documentation ⬜ not-started
**Description**: Update help text and documentation
- Update command help messages
- Ensure flag descriptions are clear
- Add examples to help text

**Validation**: `razd --help` shows accurate information

**Dependencies**: T4.5

---

### T5.2: Update CHANGELOG ⬜ not-started
**Description**: Document new feature in CHANGELOG.md
- Add entry in [Unreleased] section
- Document both `--taskfile` and `--razdfile` flags
- Include usage examples
- Note backward compatibility

**Validation**: CHANGELOG follows Keep a Changelog format

**Dependencies**: T5.1

---

### T5.3: Update spec deltas ⬜ not-started
**Description**: Finalize spec changes for cli-interface
- Ensure all scenarios are accurate
- Update requirement descriptions
- Cross-reference related specs

**Validation**: Specs match implementation

**Dependencies**: T5.2

---

### T5.4: Run OpenSpec validation ⬜ not-started
**Description**: Validate proposal with openspec tool
- Run `openspec validate add-taskfile-flag --strict`
- Fix any validation errors
- Ensure all requirements have scenarios

**Validation**: `openspec validate` passes with no errors

**Dependencies**: T5.3

---

## Summary

**Total tasks**: 23
- Phase 1 (CLI Interface): 4 tasks
- Phase 2 (Configuration Layer): 4 tasks  
- Phase 3 (Commands Update): 6 tasks
- Phase 4 (Integration Testing): 5 tasks
- Phase 5 (Documentation and Validation): 4 tasks

**Parallelizable work**:
- Phase 1 and Phase 2 can be done in parallel
- Tests in Phase 4 can be done in parallel after their dependencies

**Critical path**: T1.1 → T1.2 → T1.3 → T3.1/T3.2/T3.3/T3.4 → T4.1-T4.5 → T5.4

**Estimated effort**: Medium (touches CLI, config, and all commands but changes are straightforward)
