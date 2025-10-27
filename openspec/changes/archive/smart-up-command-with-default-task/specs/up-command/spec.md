# Spec Delta: Up Command

## MODIFIED Requirements

### Requirement: UP_CMD_CONTEXT - Command SHALL provide context-aware project setup
The `razd up` command SHALL intelligently adapt its behavior based on the current directory context and provided arguments.

#### Scenario: Setup with existing configuration
**Given** the current directory contains a Razdfile.yml with `tasks.default`  
**When** user runs `razd up` without arguments  
**Then** the system executes the default task workflow  
**And** shows progress feedback during execution  
**And** reports success or failure status  

#### Scenario: Setup with legacy configuration
**Given** the current directory contains a Razdfile.yml with only `tasks.up`  
**When** user runs `razd up` without arguments  
**Then** the system executes the up task workflow  
**And** displays a deprecation warning about `tasks.up`  
**And** suggests migrating to `tasks.default`  

#### Scenario: Setup without configuration
**Given** the current directory does not contain a Razdfile.yml  
**When** user runs `razd up` without arguments  
**Then** the system offers to create a basic Razdfile.yml interactively  
**And** prompts for project type selection  
**And** generates appropriate configuration based on detected tools  
**And** optionally executes the default task after creation  

#### Scenario: Repository cloning and setup
**Given** a valid git repository URL  
**When** user runs `razd up <url>`  
**Then** the system clones the repository to a local directory  
**And** changes to the cloned directory  
**And** automatically detects and executes the appropriate setup workflow  
**And** prioritizes `tasks.default` over `tasks.up` if both exist  

#### Scenario: Error handling for ambiguous context
**Given** the current directory contains project files but no clear configuration  
**When** user runs `razd up` without arguments  
**Then** the system provides helpful guidance on next steps  
**And** suggests appropriate commands based on detected project type  

---

### Requirement: UP_CMD_INTERACTIVE - Command SHALL support interactive configuration creation
When no Razdfile.yml is found, the `razd up` command SHALL provide an interactive workflow to create project configuration.

#### Scenario: Interactive project type selection
**Given** no Razdfile.yml exists in current directory  
**When** user runs `razd up` and chooses to create configuration  
**Then** the system presents a list of common project types  
**And** allows custom configuration options  
**And** detects existing tools (package.json, Cargo.toml, etc.) to suggest templates  

#### Scenario: Tool-based configuration generation
**Given** the current directory contains a package.json file  
**When** user chooses Node.js project type during interactive setup  
**Then** the system generates a Razdfile.yml with appropriate Node.js tasks  
**And** includes common development workflow steps  
**And** uses `tasks.default` as the primary task  

#### Scenario: Minimal configuration creation
**Given** user chooses minimal setup during interactive configuration  
**When** the configuration is generated  
**Then** the system creates a basic Razdfile.yml with placeholder tasks  
**And** includes helpful comments explaining task structure  
**And** provides examples of common task patterns  

---

### Requirement: UP_CMD_MIGRATION - Command SHALL support smooth migration from legacy formats
The `razd up` command SHALL handle legacy configuration formats during a transition period.

#### Scenario: Legacy format deprecation warning
**Given** a Razdfile.yml contains `tasks.up` but no `tasks.default`  
**When** user runs `razd up`  
**Then** the system shows a clear deprecation warning  
**And** explains the benefits of migrating to `tasks.default`  
**And** provides guidance on how to update the configuration  
**And** continues execution with the legacy task  

#### Scenario: Migration assistance offer
**Given** a Razdfile.yml with only legacy `tasks.up`  
**When** user runs `razd up` multiple times  
**Then** the system periodically offers automated migration  
**And** explains what changes would be made  
**And** provides option to create backup before migration  

## ADDED Requirements

### Requirement: UP_CMD_DEFAULT_TASK - Command SHALL prioritize default task execution
The `razd up` command SHALL execute `tasks.default` as the primary project workflow.

#### Scenario: Default task priority
**Given** a Razdfile.yml contains both `tasks.default` and `tasks.up`  
**When** user runs `razd up`  
**Then** the system executes `tasks.default`  
**And** ignores the legacy `tasks.up` task  
**And** logs which task is being executed  

#### Scenario: Default task validation
**Given** a Razdfile.yml contains `tasks.default`  
**When** the system attempts to execute the default task  
**Then** it validates the task configuration before execution  
**And** provides clear error messages for invalid configurations  
**And** suggests fixes for common configuration issues  

---

### Requirement: UP_CMD_SMART_DETECTION - Command SHALL intelligently detect project context
The `razd up` command SHALL analyze the current directory to determine the most appropriate setup behavior.

#### Scenario: Multi-format project detection
**Given** a directory contains multiple project indicators (package.json, Cargo.toml, Taskfile.yml)  
**When** user runs `razd up` without Razdfile.yml  
**Then** the system detects all relevant project types  
**And** offers configuration templates that integrate multiple tools  
**And** prioritizes the most common project patterns  

#### Scenario: Tool version detection and setup
**Given** a directory contains tool configuration files  
**When** interactive setup is triggered  
**Then** the system detects required tool versions  
**And** suggests appropriate mise configuration  
**And** includes tool installation in the generated default task  

## REMOVED Requirements

### Requirement: INIT_CMD_SEPARATE - Separate initialization command
**Reason**: Functionality merged into smart `razd up` command for better UX  
**Migration**: Users should use `razd up` in empty directories for initialization