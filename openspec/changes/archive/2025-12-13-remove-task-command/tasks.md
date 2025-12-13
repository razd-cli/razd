# Implementation Tasks: Remove `razd task` Command

## Phase 1: Remove Command Infrastructure
- [x] Remove `Commands::Task` variant from the enum in src/main.rs
- [x] Remove the match arm for `Commands::Task` in src/main.rs
- [x] Remove `pub mod task;` from src/commands/mod.rs
- [x] Delete src/commands/task.rs file

## Phase 2: Update Error Messages and Help Text
- [x] Update NoDefaultTask error message in src/core/error.rs to reference `razd run` instead of `razd task`
- [x] Update help text in src/commands/up.rs that shows `razd task <name>` to use `razd run <name>`
- [x] Scan for any other error messages or comments referencing `razd task` and update them

## Phase 3: Update Documentation
- [x] Update README.md to replace all `razd task` references with `razd run`
- [x] Update examples/nodejs-project/README.md to use `razd run` instead of `razd task`
- [x] Review any other markdown files in the repository for `razd task` references

## Phase 4: Update Changelog and Version
- [x] Add entry to CHANGELOG.md under version 0.4.1 explaining the breaking change
- [x] Include migration guidance in CHANGELOG: `razd task <name>` → `razd run <name>`
- [x] Update version in Cargo.toml from "0.4.0" to "0.4.1"

## Phase 5: Verification and Testing
- [x] Build the project: `cargo build --release`
- [x] Test that `razd task` is no longer recognized: `cargo run -- task` should show "unexpected argument" error
- [x] Test that `razd run` works correctly with an example task
- [x] Verify help output: `cargo run -- --help` should not mention `task` command
- [x] Run full test suite: `cargo test`

## Phase 6: Documentation Review
- [x] Run `rg -i "razd task"` to find any remaining references in the codebase
- [x] Fix any remaining references found in the search
- [x] Verify all examples work with updated commands

## Validation
- [x] All tasks above completed and marked with [x]
- [x] No compilation errors or warnings
- [x] All tests pass
- [x] CLI help output correct
- [x] Examples directory working correctly

## Bonus Tasks Completed
- [x] Removed unused `execute_task` function from src/integrations/taskfile.rs that was only used by the deleted command

## Dependencies
None - all tasks can be completed sequentially in the order listed.

## Estimated Effort
~1-2 hours for implementation and testing.

