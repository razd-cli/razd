# Spec Delta: Up Command

## MODIFIED Requirements

### Requirement: UP_CMD_ARGS - Command SHALL accept optional repository URL
The `razd up` command SHALL accept an optional repository URL argument. When the URL is provided, the command SHALL clone the repository before setup. When omitted, the command SHALL set up the project in the current directory.

#### Scenario: Clone and setup with URL (existing behavior)
**Given** a valid git repository URL  
**When** user runs `razd up <url>`  
**Then** the tool clones the repository to a local directory  
**And** changes to the cloned directory  
**And** executes the up workflow  

#### Scenario: Setup local project without URL (new behavior)
**Given** the current directory contains project files (Razdfile.yml, Taskfile.yml, or mise.toml)  
**When** user runs `razd up` without arguments  
**Then** the tool validates project files exist  
**And** executes the up workflow in the current directory  

#### Scenario: Error when no URL and no project detected
**Given** the current directory does not contain project files  
**When** user runs `razd up` without arguments  
**Then** the tool returns an error  
**And** displays a helpful message suggesting to use `razd up <url>` or `razd init`  

---

### Requirement: UP_CMD_DETECT - Command SHALL detect local project configuration
When `razd up` is run without a URL, the command SHALL detect if the current directory contains a valid project by checking for at least one project indicator file.

**Project indicator files** (checked in any order):
- `Razdfile.yml` - Primary razd configuration
- `Taskfile.yml` - Task runner configuration  
- `mise.toml` or `.mise.toml` - Tool version management

#### Scenario: Detect project with Razdfile
**Given** current directory contains `Razdfile.yml`  
**When** `razd up` is run without URL  
**Then** project detection succeeds  
**And** up workflow executes  

#### Scenario: Detect project with Taskfile only
**Given** current directory contains `Taskfile.yml` but no other project files  
**When** `razd up` is run without URL  
**Then** project detection succeeds  
**And** up workflow executes  

#### Scenario: Detect project with mise.toml only
**Given** current directory contains `mise.toml` but no other project files  
**When** `razd up` is run without URL  
**Then** project detection succeeds  
**And** up workflow executes  

#### Scenario: No detection with only .git directory
**Given** current directory contains only `.git` directory  
**When** `razd up` is run without URL  
**Then** project detection fails  
**And** error message is displayed  

---

### Requirement: UP_CMD_WORKFLOW - Command SHALL execute up workflow
After directory is determined (either cloned or current), the command SHALL execute the up workflow with the same behavior regardless of how the project was located.

#### Scenario: Execute workflow after clone
**Given** repository was cloned successfully  
**When** workflow execution begins  
**Then** checks for workflow configuration  
**And** executes workflow task or falls back to mise install + task setup  
**And** displays success message  

#### Scenario: Execute workflow in local directory
**Given** local project was detected successfully  
**When** workflow execution begins  
**Then** checks for workflow configuration  
**And** executes workflow task or falls back to mise install + task setup  
**And** displays success message  

---

## ADDED Requirements

### Requirement: UP_CMD_VALIDATE - Command SHALL validate project directory
When running `razd up` without a URL, the command SHALL validate that the current directory contains at least one project indicator file before proceeding with setup.

**Validation rules:**
- At least one of: `Razdfile.yml`, `Taskfile.yml`, `mise.toml`, or `.mise.toml` must exist
- `.git` directory alone is insufficient for validation
- Validation occurs before any setup operations begin

#### Scenario: Validation success with multiple files
**Given** current directory contains both `Taskfile.yml` and `mise.toml`  
**When** validation runs  
**Then** validation succeeds  
**And** setup proceeds  

#### Scenario: Validation failure in empty directory
**Given** current directory is empty  
**When** validation runs  
**Then** validation fails  
**And** error message includes: "No project detected in current directory"  
**And** error message suggests using `razd up <url>` or `razd init`  

#### Scenario: Validation failure with only git directory
**Given** current directory contains only `.git` subdirectory  
**When** validation runs  
**Then** validation fails  
**And** error message is displayed  

---

### Requirement: UP_CMD_HELP - Help text SHALL describe optional URL
The command help text SHALL accurately describe the optional URL parameter and both usage modes.

#### Scenario: Display help text
**Given** user runs `razd up --help`  
**Then** help text indicates URL is optional  
**And** describes clone mode with URL  
**And** describes local setup mode without URL  

---

## Implementation Notes

### Backward Compatibility
- All existing `razd up <url>` invocations continue to work unchanged
- URL argument is optional, not removed
- No breaking changes to CLI interface

### Error Messages
All error messages shall:
- Clearly indicate the problem
- Provide actionable guidance
- Suggest relevant commands when appropriate

### Cross-References
- Depends on: existing workflow execution logic in `get_workflow_config()` and `taskfile::execute_workflow_task()`
- Related to: `razd init` command for project initialization
- Related to: `razd install` and `razd setup` commands (local mode essentially combines these)
