# Implementation Tasks

## Phase 1: Core Implementation
- [x] 1.1 Modify `execute()` in `src/commands/run.rs` to accept and use `args` parameter (remove `_` prefix)
- [x] 1.2 Update taskfile execution to pass CLI args with `--` separator to task command
- [x] 1.3 Add `execute_workflow_task_with_args()` function to handle CLI arguments
- [x] 1.4 Handle empty args case (when no `--` is provided)

## Phase 2: Testing
- [x] 2.1 Add unit test for workflow config loading
- [x] 2.2 Verify all existing tests still pass
- [x] 2.3 Add integration test: `razd run test -- -v -race` should pass args correctly
- [x] 2.4 Integration test verifies CLI_ARGS is available in task commands
- [x] 2.5 Test edge cases: empty args, args with spaces, args with special chars

## Phase 3: Documentation
- [x] 3.1 README.md points to external docs (no changes needed)
- [x] 3.2 Add CLI_ARGS example to Razdfile.yml in examples/nodejs-project
- [x] 3.3 CLI help text already shows trailing args support

## Validation
- [x] All tests pass: `cargo test` - 177 tests passing
- [x] Manual testing with example tasks - verified with test-razdfile.yml
- [x] OpenSpec validation passes: `openspec validate add-cli-args-support --strict`
