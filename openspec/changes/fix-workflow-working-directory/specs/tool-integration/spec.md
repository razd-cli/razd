# tool-integration Specification Delta

## MODIFIED Requirements

### Requirement: Taskfile integration
The system MUST integrate with taskfile.dev for task execution.

#### Scenario: Execute task setup command
**Given** a project directory contains Taskfile.yml or Taskfile.yaml
**When** razd executes the setup phase
**Then** the system should:
- Run `task setup` command
- Display task output to the user
- Handle any errors from task execution
- Verify successful completion

#### Scenario: Execute specific named task
**Given** a project has taskfile configuration with defined tasks
**When** a user runs `razd task build --verbose`
**Then** the system should:
- Run `task build --verbose` command
- Pass all arguments to the task command
- Display task output to the user

#### Scenario: Execute default task (dev server)
**Given** a project has taskfile configuration
**When** a user runs `razd task` with no arguments
**Then** the system should:
- Run `task` command (which typically starts default/dev task)
- Display task output to the user
- Keep the process running for long-running tasks like dev servers

#### Scenario: Handle missing taskfile configuration
**Given** a project directory lacks Taskfile.yml/Taskfile.yaml
**When** razd attempts to run task commands
**Then** the system should:
- Display an error message about missing taskfile
- Suggest creating a Taskfile.yml
- Provide link to taskfile.dev documentation

#### Scenario: Task command not installed
**Given** taskfile (task command) is not installed on the user's system
**When** razd attempts to run task commands
**Then** the system should:
- Display an error message about missing task command
- Provide installation instructions for taskfile
- Fail gracefully

#### Scenario: Workflow execution uses system temp directory with explicit working directory
**Given** razd needs to execute a workflow task with custom content
**When** the workflow execution begins
**Then** the system should:
- Create temporary workflow file in system temp directory (`std::env::temp_dir()`)
- Name the file `razd-workflow-{task_name}.yml` (without leading dot)
- Execute task using `--dir` parameter to set working directory to project directory
- Execute task using `--taskfile` parameter pointing to temp file
- Ensure project files are accessible during task execution
- Remove the temporary file after execution completes
- Never create temporary files in the project working directory

#### Scenario: Workflow commands execute in project directory
**Given** a workflow task is being executed with temp taskfile
**When** commands in the workflow run (e.g., `npm install`, `cargo build`)
**Then** the system should:
- Execute all commands in the project working directory
- Find project files (package.json, Cargo.toml, etc.) correctly
- Not attempt to find files in the temp directory
- Produce output files in the project directory

#### Scenario: Temp file cleanup on success
**Given** a workflow task executes successfully
**When** the task completes
**Then** the system should:
- Remove the temporary workflow file from system temp directory
- Leave no orphaned files in temp directory
- Leave project directory completely clean

#### Scenario: Temp file cleanup on failure
**Given** a workflow task fails during execution
**When** the task fails or errors occur
**Then** the system should:
- Attempt to remove the temporary workflow file
- Handle cleanup errors gracefully (best-effort cleanup)
- Not fail the entire operation due to cleanup issues
