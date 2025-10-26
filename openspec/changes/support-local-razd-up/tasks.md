# Implementation Tasks: Support Local Project Setup with `razd up`

## Overview
Implement support for running `razd up` without a URL to set up already-cloned projects in the current directory.

## Prerequisites
- [x] Proposal approved
- [x] Design reviewed
- [x] Spec validated with `openspec validate support-local-razd-up --strict`

## Implementation Tasks

### Phase 1: CLI Interface Changes
- [x] **Task 1.1**: Update `Commands::Up` enum in `src/main.rs`
  - Change `url: String` to `url: Option<String>`
  - Update doc comment to indicate URL is optional
  - Update `run()` function to pass `url.as_deref()`
  - **Validation**: Code compiles without errors

- [x] **Task 1.2**: Update help text and command description
  - Change description to: "Clone repository and set up project, or set up local project"
  - Update URL argument description to indicate it's optional
  - **Validation**: `razd up --help` shows updated text

### Phase 2: Core Logic Implementation
- [x] **Task 2.1**: Refactor `commands::up::execute()` function signature
  - Change `url: &str` to `url: Option<&str>`
  - Add branching logic for URL vs no-URL scenarios
  - **Validation**: Code compiles, existing behavior unchanged

- [x] **Task 2.2**: Extract current logic into `execute_with_clone()` helper
  - Move existing clone-based logic to new private function
  - Keep exact same behavior as current implementation
  - **Validation**: Integration tests pass with URL argument

- [x] **Task 2.3**: Implement `execute_local_project()` helper
  - Get current directory with `env::current_dir()`
  - Call `validate_project_directory()`
  - Log working directory
  - Call `execute_up_workflow()`
  - Call `show_success_message()`
  - **Validation**: Manual test in real project directory

- [x] **Task 2.4**: Extract shared `execute_up_workflow()` helper
  - Move workflow execution logic to separate function
  - Handle workflow config lookup and fallback
  - Reuse in both clone and local modes
  - **Validation**: Both modes execute workflows correctly

- [x] **Task 2.5**: Extract shared `show_success_message()` helper
  - Move success output to separate function
  - Ensure consistent messaging across modes
  - **Validation**: Messages appear correctly in both modes

### Phase 3: Project Detection Logic
- [x] **Task 3.1**: Implement `validate_project_directory()` function
  - Check for `Razdfile.yml` existence
  - Check for `Taskfile.yml` existence
  - Check for `mise.toml` or `.mise.toml` existence
  - Return `Ok(())` if at least one exists
  - Return error with helpful message if none exist
  - **Validation**: Unit tests cover all file combinations

- [x] **Task 3.2**: Implement error message for validation failure
  - Include list of expected files
  - Suggest `razd up <url>` for cloning
  - Suggest `razd init` for new projects
  - **Validation**: Error message is clear and actionable

### Phase 4: Testing
- [x] **Task 4.1**: Add unit tests for `validate_project_directory()`
  - Test with Razdfile.yml present
  - Test with Taskfile.yml present
  - Test with mise.toml present
  - Test with .mise.toml present
  - Test with multiple files present
  - Test with no files present (error case)
  - Test with only .git directory (error case)
  - **Validation**: All tests pass, >90% coverage

- [x] **Task 4.2**: Add integration test for `razd up <url>`
  - Verify existing clone behavior unchanged
  - Test successful clone and setup
  - Test error handling for invalid URL
  - **Validation**: Integration tests pass

- [x] **Task 4.3**: Add integration test for `razd up` in local project
  - Create temp directory with Taskfile.yml
  - Run `razd up` without arguments
  - Verify workflow executes
  - Verify no clone attempted
  - **Validation**: Test passes

- [x] **Task 4.4**: Add integration test for `razd up` error case
  - Create empty temp directory
  - Run `razd up` without arguments
  - Verify error returned with helpful message
  - **Validation**: Test passes

### Phase 5: Documentation and Polish
- [x] **Task 5.1**: Update README.md usage examples
  - Add example for `razd up` without URL
  - Clarify when to use each mode
  - Update quick start guide
  - **Validation**: Documentation is clear and accurate

- [x] **Task 5.2**: Update CHANGELOG.md
  - Add entry under "Added" section
  - Describe new local project setup capability
  - Mention backward compatibility
  - **Validation**: Changelog entry follows project conventions

- [x] **Task 5.3**: Manual cross-platform testing
  - Test on Windows (PowerShell)
  - Test on Linux/macOS (bash/zsh) - tested on WSL Ubuntu
  - Verify file path handling
  - **Validation**: Works consistently across platforms

### Phase 6: Final Validation
- [x] **Task 6.1**: Run full test suite
  - Execute `cargo test`
  - All tests pass
  - No regressions
  - **Validation**: CI/CD passes

- [x] **Task 6.2**: Test real-world scenarios
  - Clone a real project with `razd up <url>` (verify no regression)
  - Run `razd up` in existing project (verify new feature)
  - Try in non-project directory (verify error handling)
  - **Validation**: All scenarios work as expected

- [x] **Task 6.3**: Update spec validation
  - Run `openspec validate support-local-razd-up --strict`
  - Fix any issues
  - **Validation**: Validation passes with no errors

## Post-Implementation
- [ ] **Task 7.1**: Create pull request
  - Reference change ID in PR title/description
  - Include all implementation tasks
  - Link to proposal and design docs
  - **Validation**: PR created and passes CI

- [ ] **Task 7.2**: Code review and iteration
  - Address review comments
  - Make requested changes
  - Re-test after changes
  - **Validation**: Approved by reviewer

- [ ] **Task 7.3**: Merge and deploy
  - Merge to main branch
  - Monitor deployment
  - Verify feature works in production
  - **Validation**: Feature live and working

## Dependencies and Blockers
- No external dependencies
- No blocking issues identified
- Can be implemented independently of other changes

## Estimated Effort
- Phase 1: 1-2 hours (CLI changes)
- Phase 2: 2-3 hours (Core logic)
- Phase 3: 1-2 hours (Detection logic)
- Phase 4: 3-4 hours (Testing)
- Phase 5: 1-2 hours (Documentation)
- Phase 6: 1 hour (Validation)
- **Total**: ~10-14 hours

## Success Criteria
- ✅ `razd up <url>` works exactly as before (no regressions)
- ✅ `razd up` works in project directories (new feature)
- ✅ Clear error when `razd up` run in non-project directory
- ✅ All tests pass with >80% coverage
- ✅ Documentation updated and clear
- ✅ Works consistently on Windows, Linux, and macOS
- ✅ Tested and verified on WSL Ubuntu
