# Proposal: Add CLI Arguments Support

## Status
Proposed

## Summary
Add support for passing CLI arguments to tasks via `--` separator (e.g., `razd run hello -- -v -race`). Arguments should be available in tasks via the `CLI_ARGS` variable for interpolation, matching taskfile.dev's standard behavior.

## Motivation
Users need to pass dynamic arguments to tasks without hardcoding them in the Razdfile. This is a common pattern in task runners like taskfile.dev and make, enabling flexible task execution such as:
- Passing test flags: `razd run test -- -v -race`
- Forwarding build options: `razd run build -- --release`
- Providing runtime parameters: `razd run start -- --port=8080`

Without this feature, users must create separate tasks for each variation or modify the Razdfile every time they need different arguments.

## Goals
- Support `--` separator for CLI argument forwarding
- Make arguments available via `CLI_ARGS` variable in task commands
- Maintain consistency with taskfile.dev's CLI_ARGS behavior
- Work seamlessly with existing workflow execution

## Non-goals
- Supporting `CLI_ARGS_LIST` (shell parsed array) - only string interpolation needed
- Modifying task definition syntax
- Supporting arguments without the `--` separator
- Adding argument validation or parsing logic

## Design Overview

### Changes Required
1. **CLI Parsing** (already exists): The `run` command already captures trailing args via `#[arg(trailing_var_arg = true)]`
2. **Argument Passing**: Pass args from `commands/run.rs` to taskfile execution
3. **Variable Injection**: Inject `CLI_ARGS` into the Razdfile YAML before execution

### Implementation Approach
When executing a task:
1. Collect arguments after `--` separator (already captured by clap)
2. Join arguments into a space-separated string
3. Add `CLI_ARGS` to the Razdfile's global `vars` section
4. Execute the task with the modified YAML

### Example Usage
```yaml
# Razdfile.yml
tasks:
  test:
    cmds:
      - go test {{.CLI_ARGS}}

  hello:
    cmds:
      - go test {{.CLI_ARGS}}
```

```bash
# CLI usage
$ razd run hello -- -v -race
# Executes: go test -v -race

$ razd run test -- -timeout=30s -run=TestExample
# Executes: go test -timeout=30s -run=TestExample
```

## Alternatives Considered
1. **Environment variables**: Less ergonomic and doesn't match taskfile.dev convention
2. **Task parameters**: Would require Razdfile syntax changes and break compatibility
3. **Named arguments**: More complex parsing, less flexible than `--` separator

## Dependencies
- No external dependencies
- Builds on existing argument parsing in clap
- Uses existing variable interpolation in taskfile execution

## Risks and Mitigations
- **Risk**: Arguments with special characters might cause YAML serialization issues
  - **Mitigation**: Properly escape/quote values during YAML generation
- **Risk**: Users might expect shell parsing behavior (CLI_ARGS_LIST)
  - **Mitigation**: Document that only string interpolation is supported

## Open Questions
None - the design is straightforward and follows established patterns.
