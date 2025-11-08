# cli-interface Specification Deltas

## ADDED Requirements

### Requirement: List all tasks including internal ones
The `razd list` command MUST support a `--list-all` flag to show all tasks including those marked as internal.

#### Scenario: Show all tasks with --list-all flag
**Given** a Razdfile.yml with both public and internal tasks:
```yaml
version: '3'
tasks:
  build:
    desc: Build project
  test:
    desc: Run tests
  internal-setup:
    desc: Internal setup task
    internal: true
```
**When** the user runs `razd list --list-all`  
**Then** the system should:
- Load the Razdfile.yml configuration
- Extract all tasks including internal ones
- Display all tasks with their descriptions
- Include the internal-setup task in the output

#### Scenario: Default behavior without --list-all
**Given** a Razdfile.yml with internal tasks  
**When** the user runs `razd list` without the `--list-all` flag  
**Then** the system should:
- Load the Razdfile.yml configuration
- Filter out tasks marked with `internal: true`
- Display only public tasks (current behavior)
- Not show internal-setup in the output

#### Scenario: All tasks are internal
**Given** a Razdfile.yml where all tasks are marked internal  
**When** the user runs `razd list`  
**Then** the system should display "No tasks found in Razdfile.yml"  
**And** when the user runs `razd list --list-all`  
**Then** the system should display all internal tasks

### Requirement: JSON output format for task lists
The `razd list` command MUST support a `--json` flag to output task information in JSON format for machine parsing and automation.

#### Scenario: Output tasks as JSON
**Given** a Razdfile.yml with multiple tasks:
```yaml
version: '3'
tasks:
  build:
    desc: Build project
  test:
    desc: Run tests
  deploy:
    desc: Deploy application
```
**When** the user runs `razd list --json`  
**Then** the system should:
- Load the Razdfile.yml configuration
- Extract task information
- Output valid JSON with the structure:
```json
{
  "tasks": [
    {
      "name": "build",
      "desc": "Build project",
      "internal": false
    },
    {
      "name": "deploy",
      "desc": "Deploy application",
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
- Tasks should be sorted alphabetically by name

#### Scenario: JSON output with empty description
**Given** a Razdfile.yml with a task without description:
```yaml
version: '3'
tasks:
  build:
    cmds:
      - echo "Building"
```
**When** the user runs `razd list --json`  
**Then** the JSON output should include the task with empty description:
```json
{
  "tasks": [
    {
      "name": "build",
      "desc": "",
      "internal": false
    }
  ]
}
```

#### Scenario: JSON output is valid and parseable
**Given** any Razdfile.yml with tasks  
**When** the user runs `razd list --json`  
**Then** the output should be valid JSON that can be parsed by standard JSON parsers  
**And** the output should use consistent formatting (pretty-printed with 2-space indentation)

### Requirement: Combined --list-all and --json flags
The `razd list` command MUST support using both `--list-all` and `--json` flags together.

#### Scenario: JSON output showing all tasks including internal
**Given** a Razdfile.yml with internal tasks:
```yaml
version: '3'
tasks:
  build:
    desc: Build project
  internal-setup:
    desc: Internal setup
    internal: true
```
**When** the user runs `razd list --list-all --json`  
**Then** the system should output JSON including internal tasks:
```json
{
  "tasks": [
    {
      "name": "build",
      "desc": "Build project",
      "internal": false
    },
    {
      "name": "internal-setup",
      "desc": "Internal setup",
      "internal": true
    }
  ]
}
```

#### Scenario: Flag order independence
**Given** a Razdfile.yml with tasks  
**When** the user runs `razd list --json --list-all`  
**Or** the user runs `razd list --list-all --json`  
**Then** both commands should produce identical output

## MODIFIED Requirements

### Requirement: Task listing command
The `razd list` command MUST display available tasks from Razdfile.yml with support for filtering and output format options.

#### Scenario: Basic task listing (backward compatibility)
**Given** a Razdfile.yml with tasks  
**When** the user runs `razd list` with no additional flags  
**Then** the system should:
- Display only non-internal tasks
- Use text format output
- Maintain exact current behavior for backward compatibility

#### Scenario: Global --list flag compatibility
**Given** a user runs `razd --list` (global flag)  
**When** the command executes  
**Then** it should behave identically to `razd list` with no additional flags  
**And** should not include internal tasks
**And** should use text format output

#### Scenario: Error handling with formatting flags
**Given** no Razdfile.yml exists in the current directory  
**When** the user runs `razd list --json`  
**Then** the system should output an error in JSON format:
```json
{
  "error": "Razdfile.yml not found in current directory"
}
```
**And** when the user runs `razd list --list-all` (without --json)  
**Then** the system should display text error message as currently implemented