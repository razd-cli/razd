# Proposal: Add List Command

## Problem Statement

Users need a way to see all available tasks in their `Razdfile.yml` without opening the file. Currently, the only way to discover available tasks is to manually inspect the Razdfile or run `task list` directly (which bypasses razd's workflow).

## Proposed Solution

Add a `list` command to razd that displays all available tasks with their descriptions, similar to `task list`.

### Command Variants

Support multiple command patterns for user convenience:
- `razd list` - Primary command
- `razd run --list` - Alternative flag syntax
- `razd --list` - Short global flag

### Output Format

Display tasks in a format similar to `task list`:

```
task: Available tasks for this project:
* add-extensions-to-product:       Add all extensions to builtInExtensions in product.json (auto-detected)
* build:                           Build project
* build-zip:                       Create ZIP installer package (recommended, always works)
* change-exe-icon:                 Change VS Code executable icon using rcedit
```

## Implementation Details

### Command Structure

1. **`razd list`** - Standalone subcommand
   - Parse `Razdfile.yml`
   - Extract all tasks and their descriptions
   - Format and display output

2. **`razd run --list`** - Flag on run command
   - Add `--list` flag to `run` subcommand
   - Same output as `razd list`
   - Does not execute any task

3. **`razd --list`** - Global flag
   - Add global flag to main CLI
   - Same output as `razd list`

### Task Filtering

- Show all tasks by default
- Optionally hide `internal: true` tasks (configurable)
- Sort alphabetically by task name
- Handle tasks with missing descriptions gracefully

### Output Details

For each task, display:
- Task name (left-aligned)
- Description (right-aligned with padding)
- Mark default task with special indicator (optional)

## Benefits

1. **Discoverability**: Users can quickly see what tasks are available
2. **Convenience**: No need to open `Razdfile.yml` to see task list
3. **Consistency**: Matches `task list` behavior that users expect
4. **Documentation**: Task descriptions serve as inline documentation

## Alternatives Considered

1. **Only `razd list`**: Simplest, but users might expect `--list` flag
2. **Only `razd run --list`**: Consistent with other CLIs, but less discoverable
3. **Shell completion**: Could provide task names, but doesn't show descriptions

## Migration Path

This is a purely additive change - no breaking changes to existing functionality.

## Example Usage

```bash
# List all tasks
razd list

# List tasks using flag
razd run --list

# List tasks using global flag
razd --list
```

## Success Criteria

- [x] `razd list` displays all non-internal tasks with descriptions
- [x] `razd run --list` works identically to `razd list`
- [x] `razd --list` works identically to `razd list`
- [x] Output formatting matches or improves upon `task list`
- [x] Works correctly when no Razdfile.yml exists
- [x] Works correctly when Razdfile.yml has no tasks
