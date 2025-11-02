# Implementation Tasks

## Prerequisites
- [ ] Review existing task execution code in `src/commands/dev.rs` and `src/commands/build.rs`
- [ ] Understand `get_workflow_config()` function behavior
- [ ] Review taskfile integration module

## Core Implementation
- [ ] Add `Run { task_name: String, args: Vec<String> }` variant to `Commands` enum in `src/main.rs`
- [ ] Create `src/commands/run.rs` with `execute()` function
- [ ] Implement task execution logic in `run.rs` using `get_workflow_config()` and taskfile integration
- [ ] Add command routing in `src/main.rs` to call `commands::run::execute()`
- [ ] Export new module in `src/commands/mod.rs`

## Error Handling
- [ ] Add clear error message when task is not found in Razdfile.yml
- [ ] Ensure mise sync check runs before task execution
- [ ] Handle case when Razdfile.yml doesn't exist

## Testing
- [ ] Write integration test for `razd run test` with custom test task
- [ ] Write integration test for `razd run deploy` with custom deploy task
- [ ] Test error handling when task doesn't exist
- [ ] Verify `razd dev` still works (regression test)
- [ ] Verify `razd build` still works (regression test)
- [ ] Test with arguments: `razd run test --verbose`

## Documentation
- [ ] Update README.md with `razd run` examples
- [ ] Add examples to Razdfile.yml template showing custom tasks
- [ ] Update CLI help text to mention `razd run` command

## Validation
- [ ] Run `openspec validate add-custom-task-runner --strict`
- [ ] Resolve any validation errors
- [ ] Confirm all tests pass
- [ ] Test on both Windows and Unix (if available)
