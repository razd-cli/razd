# Tasks

## Phase 1: Data Structure Updates
- [x] **T1.1** Extend `TaskInfo` struct in `src/commands/list.rs` to include `task`, `summary`, `aliases`, `location` fields
- [x] **T1.2** Create `TaskLocation` struct with `taskfile`, `line`, `column` fields
- [x] **T1.3** Update `TaskListOutput` struct to include root `location` field
- [x] **T1.4** Add `is_false()` helper function for conditional serialization of `internal` field

## Phase 2: Location Tracking Implementation
- [x] **T2.1** Implement `find_razdfile_path()` to get absolute path to Razdfile.yml
- [x] **T2.2** Implement `estimate_task_line()` function that scans file to find task definition line
- [x] **T2.3** Add logic to compute absolute paths cross-platform (Windows/Unix)
- [x] **T2.4** Handle edge cases: missing file, malformed YAML, tasks not found

## Phase 3: JSON Output Enhancement
- [x] **T3.1** Modify `execute()` to populate new fields when `json` flag is true
- [x] **T3.2** Set `task` field equal to `name` field for each task
- [x] **T3.3** Set `summary` to empty string (placeholder for future feature)
- [x] **T3.4** Set `aliases` to empty array (placeholder for future feature)
- [x] **T3.5** Populate `location` object with file path, line, column
- [x] **T3.6** Add root `location` field to output
- [x] **T3.7** Ensure `internal` field serialization skips false values for cleaner JSON

## Phase 4: Testing
- [x] **T4.1** Add unit test: Verify JSON structure matches taskfile format
- [x] **T4.2** Add unit test: Check that `task` equals `name` for all tasks
- [x] **T4.3** Add unit test: Verify `location` object contains all required fields
- [x] **T4.4** Add unit test: Confirm root `location` is absolute path
- [x] **T4.5** Add integration test: Test JSON output with real Razdfile.yml
- [x] **T4.6** Add integration test: Verify `--list-all --json` includes internal tasks with locations
- [x] **T4.7** Add integration test: Cross-platform path format validation (Windows/Unix)
- [x] **T4.8** Add test: Empty Razdfile produces valid JSON with empty tasks array

## Phase 5: Documentation and Validation
- [x] **T5.1** Update README examples showing enhanced JSON output format
- [x] **T5.2** Add inline documentation to new structs and functions
- [x] **T5.3** Verify text output remains unchanged (only JSON affected)
- [x] **T5.4** Run `openspec validate enhance-list-json-output --strict`
- [x] **T5.5** Update CHANGELOG with enhanced JSON output feature

## Dependencies
- **T2.1** must complete before **T3.5** (need path resolution)
- **T2.2** must complete before **T3.5** (need line number tracking)
- **T1.x** must complete before **T3.x** (data structures must exist)
- **T3.x** must complete before **T4.x** (implementation needed for testing)
- **T4.x** must pass before **T5.4** (all tests green before validation)

## Validation Criteria
- JSON output is valid and parseable by `serde_json`
- All taskfile-compatible fields are present in output
- `task` field always equals `name` field
- `location.taskfile` is an absolute path
- `location.line` points to task definition in YAML (±2 lines acceptable)
- Root `location` field contains Razdfile.yml path
- Text output (`razd list` without `--json`) is unchanged
- `--list-all --json` correctly filters/includes internal tasks
- Cross-platform: Windows paths use backslashes, Unix uses forward slashes
- All existing tests continue to pass
- New integration tests verify enhanced JSON structure
