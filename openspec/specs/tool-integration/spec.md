# tool-integration Specification

## Purpose
TBD - created by archiving change add-rust-cli-foundation. Update Purpose after archive.
## Requirements
### Requirement: Mise integration
The system MUST integrate with mise for development tool management.

#### Scenario: Execute mise install command
**Given** a project directory contains mise configuration (.mise.toml or .tool-versions)  
**When** razd executes the install phase  
**Then** the system should:
- Run `mise install` command
- Display mise output to the user
- Handle any errors from mise installation
- Verify successful installation

#### Scenario: Handle missing mise configuration
**Given** a project directory lacks mise configuration files  
**When** razd attempts to run mise install  
**Then** the system should:
- Skip the mise installation step
- Display a warning about missing mise configuration
- Continue with remaining setup steps

#### Scenario: Mise not installed on system
**Given** mise is not installed on the user's system  
**When** razd attempts to run mise commands  
**Then** the system should:
- Display an error message about missing mise
- Provide installation instructions for mise
- Fail gracefully without continuing

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

### Requirement: Tool detection and validation
The system MUST detect and validate required tools before execution.

#### Scenario: Pre-execution tool validation
**Given** razd is about to execute commands requiring external tools  
**When** the command execution begins  
**Then** the system should:
- Check if required tools (git, mise, task) are available
- Display clear error messages for any missing tools
- Provide installation guidance for missing dependencies

#### Scenario: Tool version compatibility check
**Given** external tools are installed but may be incompatible versions  
**When** razd executes tool commands  
**Then** the system should:
- Handle version compatibility issues gracefully
- Display warnings for potentially incompatible versions
- Continue execution unless critical incompatibility is detected

### Requirement: File backup prompts during sync operations
The system SHALL prompt users to explicitly opt-out of backup creation, making backup the safe default behavior.

#### Scenario: User declines to skip backup (creates backup)
**Given** Razdfile.yml exists and needs modification  
**And** user has not set auto-approve mode  
**When** the sync operation prompts "⚠️  Razdfile.yml will be modified. Modify WITHOUT backup? [Y/n]"  
**And** user types 'n', 'N', or presses Enter  
**Then** the system should:
- Create a backup file (Razdfile.yml.backup)
- Proceed with the modification
- Display confirmation "✓ Synced mise.toml → Razdfile.yml"

#### Scenario: User explicitly opts out of backup
**Given** Razdfile.yml exists and needs modification  
**And** user has not set auto-approve mode  
**When** the sync operation prompts "⚠️  Razdfile.yml will be modified. Modify WITHOUT backup? [Y/n]"  
**And** user types 'Y' or 'y'  
**Then** the system should:
- Skip creating a backup file
- Proceed with the modification
- Display confirmation "✓ Synced mise.toml → Razdfile.yml"

#### Scenario: mise.toml overwrite prompt uses inverted logic
**Given** mise.toml exists and will be overwritten  
**And** user has not set auto-approve mode  
**When** the sync operation prompts "⚠️  mise.toml will be overwritten. Overwrite WITHOUT backup? [Y/n]"  
**And** user types 'n', 'N', or presses Enter  
**Then** the system should:
- Create a backup file (mise.toml.backup)
- Overwrite mise.toml with new content
- Update tracking metadata
- Display confirmation "✓ Synced Razdfile.yml → mise.toml"

#### Scenario: Auto-approve mode still creates backups
**Given** user has enabled auto-approve mode  
**And** create_backups config is true  
**When** a file needs modification  
**Then** the system should:
- Skip the interactive prompt
- Create backup automatically
- Proceed with modification
- Not require user interaction

#### Scenario: create_backups config disabled
**Given** user has set create_backups config to false  
**When** a file needs modification  
**Then** the system should:
- Not prompt about backups
- Not create backup files
- Proceed directly with modification

#### Scenario: Default behavior is safe (backup creation)
**Given** a user is unfamiliar with the prompt  
**And** Razdfile.yml needs modification  
**When** the user presses Enter without typing anything  
**Then** the system should:
- Interpret empty input as 'n' (decline to skip backup)
- Create a backup file
- Proceed safely with modification

