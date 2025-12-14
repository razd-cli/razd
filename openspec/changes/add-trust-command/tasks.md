# Tasks: Add Trust Command

## Phase 0: Dependencies

### Task 0.1: Add dialoguer dependency

**File**: `Cargo.toml`

- [ ] Add `dialoguer = "0.11"` to dependencies for interactive prompts with arrow key navigation

**Validation**: `cargo build` succeeds

---

## Phase 1: Trust Storage Infrastructure

### Task 1.1: Create trust module

**File**: `src/core/trust.rs`

- [ ] Define `TrustStore` struct
- [ ] Define `TrustedEntry` and `IgnoredEntry` structs with timestamps
- [ ] Implement `TrustStore::get_path()` - returns cache directory path
  - Unix: `~/.cache/razd/trusted.json`
  - Windows: `%LOCALAPPDATA%\razd\trusted.json`
- [ ] Implement `TrustStore::load()` - load from JSON file
- [ ] Implement `TrustStore::save()` - save to JSON file
- [ ] Implement path normalization (canonicalize, lowercase on Windows)

**Validation**: Unit tests for load/save/path operations

### Task 1.2: Implement trust query methods

**File**: `src/core/trust.rs`

- [ ] Implement `is_trusted(path: &Path) -> bool`
- [ ] Implement `is_ignored(path: &Path) -> bool`
- [ ] Implement `add_trusted(path: &Path) -> Result<()>`
- [ ] Implement `remove_trusted(path: &Path) -> Result<()>`
- [ ] Implement `add_ignored(path: &Path) -> Result<()>`
- [ ] Implement `remove_ignored(path: &Path) -> Result<()>`
- [ ] Implement `get_status(path: &Path) -> TrustStatus` enum

**Validation**: Unit tests for all methods

### Task 1.3: Export trust module

**File**: `src/core/mod.rs`

- [ ] Add `pub mod trust;`
- [ ] Export `TrustStore`, `TrustStatus`

**Validation**: `cargo build` succeeds

---

## Phase 2: Trust Command Implementation

### Task 2.1: Create trust command

**File**: `src/commands/trust.rs`

- [ ] Implement `execute(path: Option<&str>, untrust: bool, show: bool, all: bool, ignore: bool) -> Result<()>`
- [ ] Default behavior: trust current directory + run `mise trust`
- [ ] `--untrust`: remove from trusted list
- [ ] `--show`: display trust status
- [ ] `--ignore`: add to ignored list
- [ ] `--all`: trust all parent directories with config

**Validation**: Manual test of each flag

### Task 2.2: Add trust subcommand to CLI

**File**: `src/main.rs`

- [ ] Add `Trust` variant to `Commands` enum with flags:
  - `path: Option<String>` - optional path argument
  - `--untrust` flag
  - `--show` flag
  - `--all` flag
  - `--ignore` flag
- [ ] Add match arm in main to call `commands::trust::execute()`

**Validation**: `razd trust --help` shows all options

### Task 2.3: Export trust command module

**File**: `src/commands/mod.rs`

- [ ] Add `pub mod trust;`

**Validation**: `cargo build` succeeds

---

## Phase 3: Trust Guard Integration

### Task 3.1: Create trust guard function

**File**: `src/core/trust.rs`

- [ ] Implement `ensure_trusted(path: &Path, auto_yes: bool) -> Result<()>`
  - Check if path is trusted
  - If ignored, return error
  - If auto_yes, auto-trust and continue
  - Otherwise, show interactive prompt using `dialoguer::Select`:

    ```
    razd config files in /path/to/project are not trusted. Trust them?

       [ Yes ]    [ No ]    [ Ignore ]

    ←/→ toggle • y/n/i/enter submit
    ```

  - Add to appropriate list based on response
  - If trusted, also run `mise trust` if mise config exists

**Validation**: Manual test with prompts

### Task 3.2: Integrate trust check into up command

**File**: `src/commands/up.rs`

- [ ] Add trust check at start of `execute_local_project()`
- [ ] Add trust check at start of `execute_with_clone()` (after clone)
- [ ] Skip trust check for `execute_init()` (no execution)
- [ ] Pass `auto_yes` flag from environment or CLI

**Validation**: `razd up` prompts for trust on new project

### Task 3.3: Integrate trust check into other commands

**Files**: `src/commands/{install,setup,dev,build,run}.rs`

- [ ] Add trust check at start of each execute function
- [ ] Use `ensure_trusted()` from trust module

**Validation**: All commands prompt for trust

### Task 3.4: Pass --yes flag to trust system

**File**: `src/main.rs`

- [ ] Set `RAZD_AUTO_YES=1` env var when `--yes` flag is present (if not already)
- [ ] Ensure trust guard reads this variable

**Validation**: `razd --yes up` auto-trusts

---

## Phase 4: Testing

### Task 4.1: Unit tests for TrustStore

**File**: `src/core/trust.rs` (tests module)

- [ ] Test load empty/missing file
- [ ] Test save and reload
- [ ] Test add/remove trusted
- [ ] Test add/remove ignored
- [ ] Test is_trusted/is_ignored
- [ ] Test path normalization

**Validation**: `cargo test trust`

### Task 4.2: Integration tests for trust command

**File**: `tests/trust_integration_tests.rs`

- [ ] Test `razd trust` adds to trusted list
- [ ] Test `razd trust --untrust` removes from list
- [ ] Test `razd trust --show` displays status
- [ ] Test `razd trust --ignore` adds to ignored list
- [ ] Test `razd --yes trust` works

**Validation**: `cargo test --test trust_integration_tests`

### Task 4.3: Integration tests for trust guard

**File**: `tests/trust_integration_tests.rs`

- [ ] Test untrusted project blocks execution
- [ ] Test trusted project allows execution
- [ ] Test `--yes` auto-trusts

**Validation**: `cargo test --test trust_integration_tests`

---

## Phase 5: Documentation

### Task 5.1: Update README

**File**: `README.md`

- [ ] Add section about trust system
- [ ] Document `razd trust` command
- [ ] Explain `--yes` behavior with trust

**Validation**: README is clear and accurate

### Task 5.2: Update CHANGELOG

**File**: `CHANGELOG.md`

- [ ] Add entry for trust command feature

**Validation**: CHANGELOG entry exists

---

## Verification Checklist

- [ ] `cargo build` succeeds
- [ ] `cargo test` passes
- [ ] `cargo clippy` has no warnings
- [ ] `cargo fmt --check` passes
- [ ] Manual test: `razd trust` works
- [ ] Manual test: trust prompt appears on first run
- [ ] Manual test: `--yes` auto-trusts
