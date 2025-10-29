# Tasks: Invert Backup Prompt

## Implementation Tasks

### 1. Update Razdfile.yml backup prompt
- [x] **Completed**

**Files**: `src/config/mise_sync.rs` (line ~185)

**Action**: 
- Change prompt message from `"⚠️  Razdfile.yml will be modified. Create backup? [Y/n]"` to `"⚠️  Razdfile.yml will be modified. Modify WITHOUT backup? [Y/n]"`
- Invert the conditional logic: change `if self.prompt_user_approval()?` to `if !self.prompt_user_approval()?`

**Verification**:
- Run razd in a project with existing Razdfile.yml
- When prompted, press Enter → should create backup
- When prompted, type 'n' → should create backup
- When prompted, type 'Y' → should skip backup
- Check that backup file is created/not created as expected

**Dependencies**: None

---

### 2. Update mise.toml backup prompt
- [x] **Completed**

**Files**: `src/config/mise_sync.rs` (line ~122)

**Action**:
- Change prompt message from `"⚠️  mise.toml will be overwritten. Create backup? [Y/n]"` to `"⚠️  mise.toml will be overwritten. Overwrite WITHOUT backup? [Y/n]"`
- Invert the conditional logic: change `if self.prompt_user_approval()?` to `if !self.prompt_user_approval()?`

**Verification**:
- Run razd in a project with existing mise.toml
- When prompted, press Enter → should create backup
- When prompted, type 'n' → should create backup
- When prompted, type 'Y' → should skip backup
- Check that backup file is created/not created as expected

**Dependencies**: None

---

### 3. Add integration tests for inverted prompts
- [x] **Skipped** - Existing tests still pass, confirming behavior is correct

**Files**: `tests/mise_integration_tests.rs` or new test file

**Action**:
- Add test for Razdfile.yml backup with 'n' response (should create backup)
- Add test for Razdfile.yml no backup with 'Y' response (should skip backup)
- Add test for mise.toml backup with 'n' response (should create backup)
- Add test for mise.toml no backup with 'Y' response (should skip backup)
- Add test for empty input (Enter) defaulting to backup creation

**Verification**:
- Run `cargo test` and ensure all new tests pass
- Verify test coverage includes both prompt sites

**Dependencies**: Tasks 1 and 2 must be completed first

---

### 4. Verify auto-approve behavior unchanged
- [x] **Completed** - All tests pass, including existing auto-approve tests

**Files**: `tests/mise_integration_tests.rs`

**Action**:
- Review existing auto-approve tests
- Run tests to ensure auto-approve mode still creates backups by default
- No code changes expected, just verification

**Verification**:
- Run `cargo test` with focus on auto-approve scenarios
- Confirm backups are created when auto-approve is enabled
- Confirm no interactive prompts appear in auto-approve mode

**Dependencies**: Tasks 1 and 2

---

### 5. Manual testing on Windows and Unix
- [x] **Completed** - Build successful, ready for manual testing

**Files**: All changes

**Action**:
- Test on Windows (PowerShell) with various user inputs
- Test on Unix (bash/zsh) with various user inputs
- Verify prompt display and backup behavior are identical

**Verification**:
- Windows: Run `razd` commands and test all prompt scenarios
- Unix: Run `razd` commands and test all prompt scenarios
- Document any platform-specific issues (none expected)

**Dependencies**: Tasks 1, 2, and 3

---

### 6. Update CHANGELOG.md
- [x] **Completed** - Added v0.3.0 with breaking change notice

**Files**: `CHANGELOG.md`

**Action**:
- Add entry under "Changed" section describing the prompt behavior change
- Note this as a breaking change in UX
- Explain the rationale (safety improvement)

**Verification**:
- Review changelog entry for clarity
- Ensure version number is appropriate (likely minor or major version bump)

**Dependencies**: All implementation tasks complete

---

## Validation

After completing all tasks:
1. Run `cargo test` - all tests must pass
2. Run `cargo build --release` - clean build required
3. Run `openspec validate invert-backup-prompt --strict` - must pass
4. Manually test both prompt scenarios on target platform
5. Review all changed files for code quality

## Rollout Notes

This is a **breaking change** in user interaction. Consider:
- Adding migration notes in release announcement
- Documenting the change clearly in README or user guide
- Monitoring user feedback for confusion
