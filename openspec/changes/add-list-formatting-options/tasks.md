## Tasks

### Phase 1: CLI Interface Updates
- [x] **T1.1** Add `--list-all` flag to `List` command in `src/main.rs`
- [x] **T1.2** Add `--json` flag to `List` command in `src/main.rs`
- [x] **T1.3** Update `commands::list::execute()` signature to accept `list_all` and `json` parameters
- [x] **T1.4** Pass flag values from CLI parser to `list::execute()` function

### Phase 2: List Command Logic
- [x] **T2.1** Modify task filtering in `list.rs` to include internal tasks when `--list-all` is true
- [x] **T2.2** Create JSON serialization structure for task information
- [x] **T2.3** Implement JSON output formatter that outputs task data as valid JSON
- [x] **T2.4** Add conditional logic to switch between text and JSON output based on flag

### Phase 3: Testing
- [x] **T3.1** Add unit test for `--list-all` flag (verifies internal tasks are shown)
- [x] **T3.2** Add unit test for `--json` flag (verifies valid JSON output)
- [x] **T3.3** Add unit test for combined `--list-all --json` flags
- [x] **T3.4** Add integration test with real Razdfile containing internal tasks
- [x] **T3.5** Verify backward compatibility (no flags = current behavior)

### Phase 4: Documentation
- [x] **T4.1** Update CLI help text for `list` command to document new flags
- [x] **T4.2** Add examples to README showing `--list-all` usage
- [x] **T4.3** Add examples to README showing `--json` usage
- [x] **T4.4** Document JSON output schema

## Dependencies
- **T1.3** depends on **T1.1** and **T1.2** (must define CLI flags first)
- **T2.x** depends on **T1.4** (need parameters passed to function)
- **T3.x** depends on **T2.x** (implementation must be complete)
- **T4.x** can run in parallel after **T2.x** (documentation after implementation)

## Validation Criteria
- `razd list` with no flags continues to work exactly as before
- `razd list --list-all` shows internal tasks marked with `internal: true`
- `razd list --json` outputs valid, parseable JSON
- `razd list --list-all --json` combines both behaviors correctly
- All existing tests continue to pass
- New tests verify the new functionality