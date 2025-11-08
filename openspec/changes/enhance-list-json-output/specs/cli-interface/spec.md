# cli-interface Specification Deltas

## MODIFIED Requirements

### Requirement: JSON Output Format Enhancement
The `razd list --json` command MUST output taskfile-compatible JSON format including source location metadata and standard fields.

#### Scenario: Basic JSON output with taskfile-compatible fields
```gherkin
Given a Razdfile.yml with a simple task:
  """yaml
  version: '3'
  tasks:
    build:
      desc: Build the project
      cmds:
        - cargo build
  """
When the user runs `razd list --json`
Then the output MUST be valid JSON containing:
  - A "tasks" array with task objects
  - Each task MUST have: name, task, desc, summary, aliases, location, internal
  - A root "location" field with the Razdfile.yml absolute path
And the task object MUST match this structure:
  """json
  {
    "tasks": [
      {
        "name": "build",
        "task": "build",
        "desc": "Build the project",
        "summary": "",
        "aliases": [],
        "location": {
          "taskfile": "/absolute/path/to/Razdfile.yml",
          "line": 4,
          "column": 3
        },
        "internal": false
      }
    ],
    "location": "/absolute/path/to/Razdfile.yml"
  }
  """
```

#### Scenario: JSON output with internal tasks filtered
```gherkin
Given a Razdfile.yml with internal and public tasks:
  """yaml
  version: '3'
  tasks:
    public:
      desc: Public task
    _internal:
      desc: Internal task
      internal: true
  """
When the user runs `razd list --json` (without --list-all)
Then the output MUST contain only the public task
And internal tasks MUST be excluded from the tasks array
```

#### Scenario: JSON output with all tasks including internal
```gherkin
Given a Razdfile.yml with internal and public tasks
When the user runs `razd list --list-all --json`
Then the output MUST contain both public and internal tasks
And tasks with internal: true MUST have the internal field set to true
And tasks without internal field MUST have internal field set to false
```

#### Scenario: Source location accuracy
```gherkin
Given a Razdfile.yml at "/home/user/project/Razdfile.yml":
  """yaml
  version: '3'
  
  tasks:
    first:
      desc: First task
    second:
      desc: Second task
  """
When the user runs `razd list --json`
Then each task's location.taskfile MUST be "/home/user/project/Razdfile.yml"
And the first task's location.line MUST be 4 (or close to it)
And the second task's location.line MUST be greater than the first task's line
And the root location field MUST be "/home/user/project/Razdfile.yml"
```

#### Scenario: Empty Razdfile handling
```gherkin
Given a Razdfile.yml with no tasks:
  """yaml
  version: '3'
  tasks: {}
  """
When the user runs `razd list --json`
Then the output MUST be valid JSON
And the tasks array MUST be empty
And the location field MUST still contain the Razdfile.yml path
```

#### Scenario: Task field conventions
```gherkin
Given any Razdfile.yml with tasks
When the user runs `razd list --json`
Then for each task object:
  - The "task" field MUST equal the "name" field
  - The "summary" field MUST be an empty string (future: extended description)
  - The "aliases" field MUST be an empty array (future: alias support)
  - The "up_to_date" field MUST be omitted (future: runtime status)
  - The "internal" field MUST be a boolean (default false)
```

## ADDED Requirements

### Requirement: Backward Compatibility for JSON Output
Changes to JSON output format MUST maintain backward compatibility with existing consumers.

#### Scenario: Existing fields preserved
```gherkin
Given an existing tool that parses razd JSON output expecting name, desc, internal
When razd is updated with enhanced JSON output
Then all existing fields (name, desc, internal) MUST remain present
And the field types MUST not change
And the field order MAY change but MUST not break strict parsers
```

#### Scenario: Additive changes only
```gherkin
Given the enhanced JSON output
Then new fields (task, summary, aliases, location) MUST be additions
And no existing fields MUST be removed or renamed
And existing boolean fields MUST retain their true/false values
```

### Requirement: Path Handling Cross-Platform
File paths in JSON output MUST be absolute and platform-appropriate.

#### Scenario: Windows path format
```gherkin
Given a Windows system with Razdfile at "C:\Users\dev\project\Razdfile.yml"
When the user runs `razd list --json`
Then the location fields MUST use Windows path format
And backslashes MUST be properly escaped in JSON strings
Example: "C:\\Users\\dev\\project\\Razdfile.yml"
```

#### Scenario: Unix path format
```gherkin
Given a Unix system with Razdfile at "/home/dev/project/Razdfile.yml"
When the user runs `razd list --json`
Then the location fields MUST use Unix path format
And paths MUST be absolute (starting with /)
Example: "/home/dev/project/Razdfile.yml"
```

#### Scenario: Relative to absolute path resolution
```gherkin
Given the user is in a subdirectory of the project
When the user runs `razd list --json`
Then the location paths MUST be absolute
And the paths MUST be resolved from the current working directory
And symbolic links SHOULD be resolved to canonical paths
```
