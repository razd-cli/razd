# cli-interface Specification Delta

## ADDED Requirements

### Requirement: Custom task execution command
The razd CLI MUST provide a generic `run` command to execute any user-defined task from Razdfile.yml.

#### Scenario: Execute custom test task
**Given** a user has a Razdfile.yml with a `test` task defined  
**And** the user is in the project directory  
**When** they run `razd run test`  
**Then** the system should:
- Load the Razdfile.yml configuration
- Find the `test` task definition
- Execute the task commands via taskfile integration
- Display the task output
- Return success if task completes successfully

#### Scenario: Execute custom deployment task
**Given** a user has a Razdfile.yml with a `deploy` task defined  
**And** the user is in the project directory  
**When** they run `razd run deploy`  
**Then** the system should:
- Load the Razdfile.yml configuration
- Find the `deploy` task definition
- Execute the task commands via taskfile integration
- Display the task output
- Return success if task completes successfully

#### Scenario: Execute custom task with arguments
**Given** a user has a Razdfile.yml with a `test` task defined  
**And** the user is in the project directory  
**When** they run `razd run test --verbose --coverage`  
**Then** the system should:
- Load the Razdfile.yml configuration
- Find the `test` task definition
- Execute the task with arguments `--verbose --coverage`
- Display the task output with requested verbosity
- Return success if task completes successfully

#### Scenario: Task not found in Razdfile.yml
**Given** a user has a Razdfile.yml without a `nonexistent` task  
**And** the user is in the project directory  
**When** they run `razd run nonexistent`  
**Then** the system should:
- Load the Razdfile.yml configuration
- Fail to find the `nonexistent` task
- Display a clear error message: "Task 'nonexistent' not found in Razdfile.yml"
- Suggest running `task --list` to see available tasks
- Return with error exit code

#### Scenario: Execute custom task when Razdfile.yml missing
**Given** a user is in a directory without Razdfile.yml  
**When** they run `razd run test`  
**Then** the system should:
- Attempt to load Razdfile.yml configuration
- Detect that Razdfile.yml doesn't exist
- Display error message: "No workflow found. Try running 'razd init' to create a Razdfile.yml"
- Return with error exit code

## MODIFIED Requirements

### Requirement: Primary CLI Commands
The razd CLI MUST provide core commands for project lifecycle management, including custom task execution.

#### Scenario: Execute any custom user-defined task
**Given** a user has defined a custom task (e.g., `lint`, `format`, `check`) in Razdfile.yml  
**When** they run `razd run <task-name>`  
**Then** the system should execute the specified task using the same workflow infrastructure as `dev` and `build` commands

#### Scenario: Backward compatibility with convenience commands
**Given** a user has `dev` and `build` tasks defined in Razdfile.yml  
**When** they run `razd dev` or `razd build`  
**Then** the system should continue using the dedicated command handlers (not `run`)  
**And** behavior should remain identical to previous versions

## Implementation Notes

### Technical Approach
- Reuse `get_workflow_config()` function to load task configuration
- Use existing taskfile integration for task execution
- Follow same pattern as `dev.rs` and `build.rs` but accept task name as parameter
- Ensure mise sync check runs before task execution

### CLI Structure
```rust
Commands::Run { task_name, args } => {
    commands::run::execute(&task_name, &args).await?;
}
```

### Error Handling
- Task not found: Clear message with available tasks hint
- No Razdfile.yml: Suggest initialization with `razd init`
- Mise sync failure: Warning message but continue execution
- Task execution failure: Pass through task error message
