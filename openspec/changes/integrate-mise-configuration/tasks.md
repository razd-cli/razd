# Implementation Tasks

## Overview
This document outlines the ordered implementation tasks for integrating mise configuration into Razdfile.yml with intelligent file synchronization.

## Task Breakdown

### Phase 1: Foundation - Data Structures and Parsing (Days 1-3)

#### Task 1.1: Extend Razdfile config structures
**Objective**: Add mise configuration support to existing Razdfile structures

**Steps**:
1. Add `MiseConfig`, `ToolConfig` structs to `src/config/razdfile.rs`
2. Add `mise: Option<MiseConfig>` field to `RazdfileConfig`
3. Implement `#[serde(untagged)]` for `ToolConfig` enum (Simple/Complex variants)
4. Add unit tests for struct serialization/deserialization

**Validation**:
- [x] Unit tests pass for parsing simple tool versions
- [x] Unit tests pass for parsing complex tool configurations
- [x] Unit tests pass for parsing plugin URLs
- [x] cargo clippy shows no warnings

**Estimated effort**: 4 hours

---

#### Task 1.2: Add tool and plugin name validation
**Objective**: Validate mise naming rules for tools and plugins

**Steps**:
1. Create `src/config/mise_validator.rs` module
2. Implement `validate_tool_name()` function (alphanumeric, hyphens, underscores, prefixes)
3. Implement `validate_plugin_url()` function (valid URL format, git refs)
4. Add validation to Razdfile parsing
5. Add error messages with examples of valid names

**Validation**:
- [x] Valid names pass validation
- [x] Invalid names fail with clear error messages
- [x] Plugin URL validation catches malformed URLs
- [x] Error messages include helpful examples

**Estimated effort**: 3 hours

---

### Phase 2: TOML Generation (Days 3-5)

#### Task 2.1: Add toml_edit dependency
**Objective**: Add and configure TOML generation library

**Steps**:
1. Add `toml_edit = "0.22"` to Cargo.toml dependencies
2. Add `sha2 = "0.10"` for project path hashing
3. Run `cargo build` to verify dependencies
4. Create basic test to verify toml_edit usage

**Validation**:
- [x] Dependencies resolve without conflicts
- [x] cargo build succeeds
- [x] Basic TOML generation test passes

**Estimated effort**: 1 hour

---

#### Task 2.2: Implement TOML generator module
**Objective**: Generate valid mise.toml from MiseConfig

**Steps**:
1. Create `src/config/mise_generator.rs` module
2. Implement `generate_mise_toml(config: &MiseConfig) -> Result<String>`
3. Handle simple tool versions as TOML strings
4. Handle complex tool configs as TOML inline tables
5. Generate [tools] section with proper formatting
6. Generate [plugins] section with proper formatting
7. Add comprehensive unit tests for all config variants

**Validation**:
- [x] Generated TOML parses correctly with mise
- [x] Simple tool versions generate correct format
- [x] Complex tool configs with all options generate correctly
- [x] Plugin URLs with git refs preserve formatting
- [x] Empty mise config generates nothing
- [x] Unit tests cover all ToolConfig variants

**Estimated effort**: 6 hours

---

#### Task 2.3: Add TOML generation integration tests
**Objective**: Verify end-to-end TOML generation from Razdfile

**Steps**:
1. Create test fixtures in `tests/fixtures/mise-config/`
2. Create Razdfile.yml samples with various mise configs
3. Implement integration test that parses Razdfile and generates TOML
4. Verify generated TOML is parseable by mise
5. Test with actual `mise config ls` command (if mise available)

**Validation**:
- [x] Integration tests pass for all fixture variants
- [x] Generated TOML validated by mise (if available)
- [x] Test coverage > 90% for generator module

**Estimated effort**: 4 hours

---

### Phase 3: File Tracking Infrastructure (Days 5-7)

#### Task 3.1: Implement file tracking metadata structure
**Objective**: Create data structures for tracking file modifications

**Steps**:
1. Create `src/config/file_tracker.rs` module
2. Implement `FileTrackingState` struct with SystemTime fields
3. Implement `get_tracking_file_path(project_dir: &Path) -> PathBuf`
4. Use SHA256 to hash absolute project path for unique tracking
5. Implement platform-specific data directory paths (Windows/Unix)
6. Add serialization/deserialization tests

**Validation**:
- [x] Tracking path generation is deterministic for same project
- [x] Different projects generate different tracking paths
- [x] Platform-specific paths use correct directories
- [x] Serialization/deserialization round-trips correctly

**Estimated effort**: 4 hours

---

#### Task 3.2: Implement file modification detection
**Objective**: Detect when files have changed since last sync

**Steps**:
1. Implement `load_tracking_state(project_dir: &Path) -> Result<Option<FileTrackingState>>`
2. Implement `save_tracking_state(project_dir: &Path, state: &FileTrackingState) -> Result<()>`
3. Implement `check_file_changes(project_dir: &Path) -> Result<ChangeDetection>`
4. Create `ChangeDetection` enum (NoChanges, RazdfileChanged, MiseTomlChanged, BothChanged)
5. Use std::fs::metadata to get modification times
6. Compare with stored tracking state
7. Handle missing files gracefully

**Validation**:
- [x] Detects Razdfile.yml modifications correctly
- [x] Detects mise.toml modifications correctly
- [x] Handles missing tracking state (first run)
- [x] Handles missing files without panicking
- [x] Unit tests cover all detection scenarios

**Estimated effort**: 5 hours

---

#### Task 3.3: Add atomic file write operations
**Objective**: Ensure safe concurrent file operations

**Steps**:
1. Implement `atomic_write_file(path: &Path, content: &str) -> Result<()>`
2. Use temp file + rename pattern for atomicity
3. Handle platform-specific temp file creation
4. Add error handling for permission issues
5. Test concurrent write scenarios

**Validation**:
- [x] Atomic writes complete successfully
- [x] Concurrent writes don't corrupt files
- [x] Partial writes are prevented
- [x] Error handling provides clear messages

**Estimated effort**: 3 hours

---

### Phase 4: Synchronization Logic (Days 7-10)

#### Task 4.1: Implement Razdfile → mise.toml sync
**Objective**: Auto-generate mise.toml when Razdfile changes

**Steps**:
1. Create `src/config/file_sync.rs` module
2. Implement `sync_razdfile_to_mise(project_dir: &Path) -> Result<()>`
3. Parse Razdfile.yml mise config
4. Backup existing mise.toml if present
5. Generate and write new mise.toml
6. Update tracking metadata
7. Display user-friendly messages
8. Add comprehensive tests

**Validation**:
- [x] mise.toml correctly generated from Razdfile
- [x] Backup created before overwrite
- [x] Tracking metadata updated after sync
- [x] User messages are clear and helpful
- [x] Integration tests verify full sync flow

**Estimated effort**: 5 hours

---

#### Task 4.2: Implement mise.toml → Razdfile sync
**Objective**: Parse mise.toml and update Razdfile.yml mise section

**Steps**:
1. Implement `parse_mise_toml(path: &Path) -> Result<MiseConfig>`
2. Use toml parsing to extract [tools] and [plugins] sections
3. Implement `update_razdfile_mise_section(razdfile_path: &Path, mise_config: MiseConfig) -> Result<()>`
4. Use serde_yaml to preserve Razdfile formatting
5. Update only mise section without affecting tasks/version
6. Add tests for partial updates

**Validation**:
- [x] mise.toml correctly parsed to MiseConfig
- [x] Razdfile.yml mise section updated correctly
- [x] Other Razdfile sections preserved
- [x] YAML formatting maintained
- [x] Tests cover edge cases (missing sections, etc.)

**Estimated effort**: 6 hours

---

#### Task 4.3: Add user prompt for mise.toml changes
**Objective**: Prompt user to sync manual mise.toml edits

**Steps**:
1. Implement `prompt_user_for_sync() -> Result<bool>` with stdin/stdout
2. Add non-interactive mode detection (CI environments)
3. Implement `sync_mise_to_razdfile_with_prompt(project_dir: &Path) -> Result<()>`
4. Display warning about manual edit
5. Show what will change
6. Handle user response (y/n/dismiss)
7. Update tracking metadata regardless of response

**Validation**:
- [x] Interactive prompt works in terminal
- [x] Non-interactive mode skips prompt
- [x] User can accept or decline sync
- [x] Metadata updated to prevent repeated prompts
- [x] Clear messages explain consequences

**Estimated effort**: 4 hours

---

#### Task 4.4: Implement sync check on command start
**Objective**: Check for sync needs before every razd command

**Steps**:
1. Create `check_and_sync_if_needed(project_dir: &Path) -> Result<()>`
2. Call from command pre-execution hook
3. Load tracking state
4. Check file modifications
5. Handle each change scenario:
   - Razdfile changed: auto-sync to mise.toml
   - mise.toml changed: prompt user
   - Both changed: show conflict message
   - No changes: skip silently
6. Add integration tests

**Validation**:
- [x] All change scenarios handled correctly
- [x] Minimal overhead when no changes detected
- [x] Clear messages for each scenario
- [x] Integration tests cover all paths
- [x] Performance acceptable (< 15ms overhead)

**Estimated effort**: 5 hours

---

### Phase 5: Integration with Existing Commands (Days 10-12)

#### Task 5.1: Add sync check to all commands
**Objective**: Integrate file sync into command execution flow

**Steps**:
1. Modify `src/main.rs` or command dispatcher
2. Add `check_and_sync_if_needed()` before command execution
3. Update all command modules (up, dev, build, task, install, setup)
4. Handle sync errors gracefully
5. Add `--no-sync` flag to bypass sync (for debugging)
6. Test each command with sync integration

**Validation**:
- [x] All commands perform sync check
- [x] Sync errors don't crash commands
- [x] `--no-sync` flag works correctly
- [x] No performance regression in commands
- [x] Integration tests updated for all commands

**Estimated effort**: 4 hours

---

#### Task 5.2: Update mise integration to use Razdfile config
**Objective**: Prefer Razdfile mise config over standalone mise.toml

**Steps**:
1. Update `src/integrations/mise.rs`
2. Modify `has_mise_config()` to check Razdfile mise section first
3. Update error messages to mention both config sources
4. Ensure mise install works with generated mise.toml
5. Add tests for both config sources

**Validation**:
- [x] Razdfile mise config takes precedence
- [x] Standalone mise.toml still supported
- [x] Error messages updated appropriately
- [x] Backward compatibility maintained
- [x] Tests verify both paths

**Estimated effort**: 3 hours

---

### Phase 6: Testing and Documentation (Days 12-14)

#### Task 6.1: Comprehensive integration testing
**Objective**: Ensure end-to-end functionality works correctly

**Steps**:
1. Create test project fixtures with various scenarios
2. Test first-time sync (no tracking metadata)
3. Test Razdfile modification → auto-sync
4. Test mise.toml manual edit → prompt
5. Test concurrent command execution
6. Test error handling (invalid YAML, invalid TOML, etc.)
7. Test cross-platform behavior (Windows/Unix)
8. Run tests in CI environment

**Validation**:
- [x] All integration tests pass on Windows
- [x] All integration tests pass on Linux/macOS
- [x] CI tests pass in non-interactive mode
- [x] Error scenarios handled gracefully
- [x] Test coverage > 85% overall

**Estimated effort**: 8 hours

---

#### Task 6.2: Update documentation and examples
**Objective**: Document the new feature for users

**Steps**:
1. Update README.md with mise config section
2. Update example Razdfile.yml in `examples/`
3. Create migration guide for existing projects
4. Add troubleshooting section for sync issues
5. Update CLI help text for affected commands
6. Add `--no-sync` flag documentation

**Validation**:
- [x] README clearly explains mise integration
- [x] Examples demonstrate common use cases
- [x] Migration guide tested with real project
- [x] Help text accurate and helpful

**Estimated effort**: 4 hours

---

#### Task 6.3: Manual testing and validation
**Objective**: Verify feature works in real-world scenarios

**Steps**:
1. Test with real projects using mise
2. Test editing Razdfile.yml mise section
3. Test manual mise.toml edits
4. Test with different tool configurations
5. Test with plugins and git refs
6. Verify performance with large configs
7. Test sync prompt in different terminals

**Validation**:
- [x] Feature works smoothly in manual testing
- [x] User experience is intuitive
- [x] Performance is acceptable
- [x] No unexpected edge cases discovered
- [x] Sync prompt displays correctly

**Estimated effort**: 4 hours

---

### Phase 7: Polish and Release (Days 14-15)

#### Task 7.1: Code review and cleanup
**Objective**: Ensure code quality and maintainability

**Steps**:
1. Run `cargo clippy` and fix all warnings
2. Run `cargo fmt` for consistent formatting
3. Review all error messages for clarity
4. Refactor any duplicated code
5. Add missing doc comments
6. Review test coverage and add missing tests

**Validation**:
- [x] No clippy warnings
- [x] Code formatted consistently
- [x] All public APIs documented
- [x] No code duplication
- [x] Test coverage > 85%

**Estimated effort**: 4 hours

---

#### Task 7.2: Performance profiling and optimization
**Objective**: Ensure minimal overhead from sync checks

**Steps**:
1. Profile sync check performance
2. Optimize file metadata reading
3. Optimize tracking state loading
4. Add caching if beneficial
5. Verify < 15ms overhead target
6. Test with large projects

**Validation**:
- [x] Sync check completes in < 15ms
- [x] No memory leaks
- [x] Performance acceptable with large configs
- [x] Caching improves repeated checks

**Estimated effort**: 3 hours

---

#### Task 7.3: Final integration testing and release prep
**Objective**: Prepare for release

**Steps**:
1. Run full test suite on all platforms
2. Test in CI/CD environments
3. Update CHANGELOG.md
4. Tag release version
5. Verify backward compatibility
6. Create release notes

**Validation**:
- [x] All tests pass on Windows, Linux, macOS
- [x] CI/CD tests pass
- [x] CHANGELOG complete and accurate
- [x] No breaking changes introduced
- [x] Release notes clear and comprehensive

**Estimated effort**: 3 hours

---

## Summary

**Total estimated effort**: 76 hours (approximately 2 weeks with full-time development)

**Dependencies**:
- Phases must be completed sequentially
- Some tasks within phases can be parallelized
- Testing tasks depend on implementation completion

**Risk areas**:
- YAML formatting preservation (Task 4.2) - may require iteration
- Cross-platform testing (Task 6.1) - ensure Windows compatibility
- User prompt UX (Task 4.3) - may need refinement based on feedback

**Success metrics**:
- All integration tests pass
- Test coverage > 85%
- No performance regression
- User feedback positive
- Zero breaking changes
