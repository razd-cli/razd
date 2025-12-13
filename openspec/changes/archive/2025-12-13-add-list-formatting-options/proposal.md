# Proposal: Add List Formatting Options

## Problem Statement

Currently, `razd list` only shows non-internal tasks in a human-readable text format. Users need additional formatting options to:

1. **See all tasks including internal ones** (like `task --list-all`)
2. **Get machine-readable JSON output** for scripting and automation (like `task --json`)

This limits automation capabilities and prevents users from inspecting internal tasks when needed.

### Current Limitations

```bash
# Current: Only shows non-internal tasks in text format
$ razd list
task: Available tasks for this project:
* build:    Build project
* test:     Run tests

# Cannot see internal tasks
# Cannot get JSON output for scripting
```

## Proposed Solution

Add two command-line flags to `razd list` that mirror taskfile's functionality:

### 1. `--list-all` Flag

Show all tasks including internal ones, matching `task --list-all` behavior.

```bash
$ razd list --list-all
task: Available tasks for this project:
* build:            Build project
* test:             Run tests
* internal-setup:   Internal setup task
```

### 2. `--json` Flag

Output tasks in JSON format for scripting and automation, matching `task --json` behavior.

```bash
$ razd list --json
{
  "tasks": [
    {
      "name": "build",
      "desc": "Build project",
      "internal": false
    },
    {
      "name": "test", 
      "desc": "Run tests",
      "internal": false
    }
  ]
}
```

### Combined Usage

Both flags can be combined:

```bash
$ razd list --list-all --json
{
  "tasks": [
    {
      "name": "build",
      "desc": "Build project",
      "internal": false
    },
    {
      "name": "internal-setup",
      "desc": "Internal setup task",
      "internal": true
    }
  ]
}
```

## Why

This change aligns razd with taskfile's capabilities and enables:

1. **Automation and scripting** - JSON output allows tools to parse task information
2. **Complete task inspection** - Users can see internal tasks for debugging
3. **Consistent user experience** - razd behaves like task, reducing cognitive load
4. **CI/CD integration** - Scripts can programmatically discover available tasks

Without these flags, users must:
- Open Razdfile.yml manually to see internal tasks
- Write custom parsers for task information
- Switch between `razd` and `task` commands

## Implementation Strategy

Extend the existing `list` command with two optional boolean flags that modify output behavior:

1. Add `--list-all` and `--json` flags to CLI parser
2. Modify `commands::list::execute()` to accept these parameters
3. Adjust task filtering logic based on `--list-all`
4. Add JSON serialization output path for `--json`
5. Maintain backward compatibility (no flags = current behavior)