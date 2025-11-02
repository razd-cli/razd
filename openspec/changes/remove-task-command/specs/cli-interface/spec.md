# cli-interface Specification Delta

## REMOVED Requirements

### Requirement: Task execution via razd task command
The `razd task` command has been removed in favor of `razd run` for task execution.

#### Scenario: Task execution with specific task name
**Given** a user is in a project directory with a Razdfile.yml containing tasks  
**When** they run `razd task build --verbose`  
**Then** the system should respond with "unexpected argument" error indicating the command is not recognized

**REMOVED**: This scenario is no longer valid. Users should use `razd run build --verbose` instead.

#### Scenario: Task execution with no arguments (dev server fallback)
**Given** a user is in a project directory with a Taskfile.yml  
**When** they run `razd task` with no arguments  
**Then** the system should respond with "unexpected argument" error

**REMOVED**: This scenario is no longer valid. Users should use `razd run` with an appropriate task name instead.

## MODIFIED Requirements

### Requirement: Primary CLI Commands
The razd CLI MUST provide core commands for project lifecycle management.

#### Scenario: Task execution with razd run
**Given** a user is in a project directory with a Razdfile.yml containing tasks  
**When** they run `razd run build --verbose`  
**Then** the system should:
- Execute the "build" task defined in Razdfile.yml
- Pass the `--verbose` flag to the underlying task
- Display execution output and results

**RATIONALE**: This replaces the `razd task` command scenarios, consolidating task execution under the more intuitive `razd run` command.

## MODIFIED Requirements

### Requirement: Help and documentation
The CLI MUST provide comprehensive help information.

#### Scenario: General help display shows run command only
**Given** a user wants to learn about razd commands  
**When** they run `razd --help` or `razd -h`  
**Then** the system should:
- Display all available commands with brief descriptions
- Show `razd run` as the command for executing custom tasks
- NOT show `razd task` in the command list

**RATIONALE**: Help output must reflect that only `razd run` is available for task execution.

## Migration Notes

### For Users
Replace all instances of `razd task <name>` with `razd run <name>`:

**Before (0.4.0 and earlier)**:
```bash
razd task build
razd task test
razd task dev
```

**After (0.4.1 and later)**:
```bash
razd run build
razd run test
razd run dev
```

### For Documentation
All references to `razd task` in error messages, help text, and examples must be updated to `razd run`.
