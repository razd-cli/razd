# Implementation Tasks

## Phase 1: CLI Interface (2 tasks)

### Task 1.1: Add --yes flag to CLI
- [x] Add `yes: bool` field to `Cli` struct in `src/main.rs`
- [x] Use clap attributes: `#[arg(short = 'y', long, global = true)]`
- [x] Add help text: "Automatically answer 'yes' to all prompts"
- [x] Verify flag appears in `razd --help` output

### Task 1.2: Store yes flag in environment
- [x] Add `std::env::set_var("RAZD_AUTO_YES", ...)` after parsing CLI args
- [x] Follow same pattern as existing `RAZD_NO_SYNC` variable
- [x] Set to "1" when flag present, "0" otherwise

## Phase 2: Configuration Layer (3 tasks)

### Task 2.1: Update SyncConfig initialization
- [x] Modify `check_and_sync_mise()` in `src/config/mod.rs`
- [x] Read `RAZD_AUTO_YES` environment variable
- [x] Set `SyncConfig.auto_approve = true` when env var is "1"
- [x] Ensure default remains `false` for backward compatibility

### Task 2.2: Update conflict resolution
- [x] Modify `handle_sync_conflict()` in `src/config/mise_sync.rs`
- [x] When `auto_approve` is true, automatically choose Option 1 (Razdfile.yml priority)
- [x] Skip the three-option prompt entirely
- [x] Print info message about automatic choice

### Task 2.3: Verify auto_approve propagation
- [x] Confirm all `prompt_user_approval()` calls respect `self.config.auto_approve`
- [x] Check backup prompts work correctly with auto_approve
- [x] Ensure sync direction prompts are skipped when auto_approve is true

## Phase 3: Commands Update (2 tasks)

### Task 3.1: Update up command
- [x] Modify `prompt_yes_no()` in `src/commands/up.rs`
- [x] Read `RAZD_AUTO_YES` environment variable
- [x] Return `true` immediately if env var is "1"
- [x] Skip stdin reading when auto-approve enabled

### Task 3.2: Update execute_local_project
- [x] Ensure Razdfile creation prompt respects yes flag
- [x] Verify workflow execution continues without prompts
- [x] Test interactive creation skips all input requests

## Phase 4: Testing (4 tasks)

### Task 4.1: Add integration test for --yes with up command
- [x] Create test: `test_yes_flag_in_help`
- [x] Verify flag appears in help output
- [x] Ensure correct description is shown
- [x] Check exit code is success

### Task 4.2: Add integration test for -y short form
- [x] Create test: `test_short_yes_flag_works`
- [x] Verify `-y` equivalent to `--yes`
- [x] Test with different commands (up, list)

### Task 4.3: Add test for mise sync auto-approve
- [x] Create test: `test_yes_flag_auto_approves_mise_sync`
- [x] Set up scenario with mise.toml conflict
- [x] Run with `--yes` flag
- [x] Verify automatic resolution to Razdfile priority
- [x] Check no stdin reading occurs

### Task 4.4: Add test for conflict resolution
- [x] Create test: `test_yes_flag_resolves_conflicts_automatically`
- [x] Create both mise.toml and Razdfile.yml with different content
- [x] Run sync with `--yes` flag
- [x] Verify Option 1 (Razdfile → mise.toml) chosen
- [x] Check no user prompts displayed

## Phase 5: Documentation (3 tasks)

### Task 5.1: Update CHANGELOG.md
- [x] Add entry under `[Unreleased]` section
- [x] Document `-y, --yes` flag addition
- [x] Explain automatic approval behavior
- [x] Note default behavior unchanged

### Task 5.2: Update README or docs
- [x] Document `--yes` flag usage
- [x] Provide automation/CI examples
- [x] Explain conflict resolution defaults
- [x] Add warning about backups with auto-approve

### Task 5.3: Mark tasks complete
- [x] Mark all tasks as completed in this file
- [x] Run `openspec validate add-yes-flag --strict`
- [x] Fix any validation errors
- [x] Commit OpenSpec changes

## Dependencies

- Task 2.1 depends on Task 1.2 (environment variable must exist)
- Phase 3 depends on Phase 2 completion (config changes must be in place)
- Phase 4 depends on Phase 1-3 completion (implementation must be done)
- Phase 5 can start after Phase 4 (documentation after implementation)

## Validation

After each phase:
1. Run `cargo test` to ensure no regressions
2. Run `cargo clippy` to check code quality
3. Test manually with `razd --yes up` in a test directory
4. Verify `razd --help` shows the new flag

Final validation:
1. All 14 tasks marked complete
2. `openspec validate add-yes-flag --strict` passes
3. All tests passing (cargo test)
4. Code formatted (cargo fmt --all)
5. No clippy warnings
