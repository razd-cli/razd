# Proposal: Add Custom Task Runner Command

## Overview
Add `razd run <task>` command to enable users to execute any custom task defined in their Razdfile.yml, providing flexibility to create and run arbitrary workflows beyond the predefined `dev` and `build` commands.

## Problem Statement
Currently, razd only supports predefined commands like `razd dev` and `razd build`. If a user defines a custom task like `test` or `deploy` in their Razdfile.yml, there's no way to execute it via razd CLI. Running `razd test` results in no action since no such command exists.

This limits users to the predefined workflow commands and prevents them from leveraging the full task system for custom workflows.

## Proposed Solution
Introduce a new `razd run <task>` command that:
- Accepts any task name defined in the user's Razdfile.yml
- Executes the specified task using the existing task execution infrastructure
- Maintains existing behavior for `razd dev` and `razd build` as convenience shortcuts
- Provides clear error messages when tasks don't exist

### User Experience
```bash
# Execute custom test task
razd run test

# Execute custom deployment task
razd run deploy

# Execute any user-defined task
razd run my-custom-task

# Existing commands continue to work
razd dev
razd build
```

## Implementation Approach
1. Add new `Run` variant to the CLI `Commands` enum
2. Create `src/commands/run.rs` module to handle task execution
3. Reuse existing task execution logic from `get_workflow_config()` and taskfile integration
4. Maintain backward compatibility with existing commands

## Benefits
- **Flexibility**: Users can define and run any workflow tasks they need
- **Consistency**: Single interface for all task execution
- **Extensibility**: Easy to add new custom workflows without CLI changes
- **Backward Compatible**: Existing `dev` and `build` commands remain unchanged

## Affected Components
- CLI interface (`src/main.rs`)
- Commands module (`src/commands/mod.rs`, new `src/commands/run.rs`)
- Task execution infrastructure (reuses existing code)

## Success Criteria
- `razd run <task>` executes any task defined in Razdfile.yml
- Clear error message when task doesn't exist
- `razd dev` and `razd build` continue working as before
- Integration tests validate new functionality
- Documentation updated with examples
