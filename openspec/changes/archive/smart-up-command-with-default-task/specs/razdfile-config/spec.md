# Spec Delta: Razdfile Configuration

## ADDED Requirements

### Requirement: DEFAULT_TASK_FORMAT - Configuration SHALL support default task semantics
The Razdfile.yml configuration format SHALL support `tasks.default` as the primary project workflow.

#### Scenario: Default task definition
**Given** a Razdfile.yml configuration file  
**When** it contains a `tasks.default` section  
**Then** the system treats this as the primary project workflow  
**And** executes it when `razd up` is called without arguments  
**And** validates the task configuration syntax  

#### Scenario: Default task with command list
**Given** a `tasks.default` configuration with multiple commands  
**When** the default task is executed  
**Then** all commands are executed in sequence  
**And** execution stops on first failure unless configured otherwise  
**And** progress is reported for each command  

```yaml
# Example configuration
tasks:
  default:
    - mise install
    - npm install
    - npm run build
```

#### Scenario: Default task with task references
**Given** a `tasks.default` configuration with task references  
**When** the default task is executed  
**Then** referenced tasks are resolved and executed  
**And** task dependencies are handled appropriately  
**And** circular references are detected and prevented  

```yaml
# Example with task references
tasks:
  default:
    - task: install
    - task: setup
  install:
    - mise install
  setup:
    - npm install
```

---

### Requirement: LEGACY_TASK_SUPPORT - Configuration SHALL support legacy task format during transition
The Razdfile.yml parser SHALL support the legacy `tasks.up` format during a migration period.

#### Scenario: Legacy task execution with warning
**Given** a Razdfile.yml contains only `tasks.up` (no `tasks.default`)  
**When** the configuration is loaded  
**Then** the system recognizes the legacy format  
**And** displays a deprecation warning to the user  
**And** treats `tasks.up` as the fallback primary task  

#### Scenario: Priority handling for mixed formats
**Given** a Razdfile.yml contains both `tasks.default` and `tasks.up`  
**When** the configuration is processed  
**Then** `tasks.default` takes priority  
**And** `tasks.up` is ignored with a notice  
**And** the user is informed about the priority selection  

#### Scenario: Migration suggestion
**Given** a configuration file with only legacy `tasks.up`  
**When** accessed multiple times  
**Then** the system suggests automatic migration  
**And** explains the benefits of the new format  
**And** offers to perform the migration with user consent  

---

### Requirement: CONFIG_TEMPLATE_GENERATION - System SHALL generate appropriate configuration templates
The system SHALL create Razdfile.yml templates based on detected project context and user preferences.

#### Scenario: Node.js project template
**Given** a directory contains package.json  
**When** interactive configuration creation is triggered  
**Then** the system generates a Node.js-appropriate template  
**And** includes common npm/yarn workflow steps  
**And** sets up appropriate mise tool configuration  

```yaml
# Generated Node.js template
tools:
  node: "20"
  npm: "latest"

tasks:
  default:
    - mise install
    - npm install
    - npm run dev
```

#### Scenario: Rust project template
**Given** a directory contains Cargo.toml  
**When** interactive configuration creation is triggered  
**Then** the system generates a Rust-appropriate template  
**And** includes cargo build and test commands  
**And** sets up appropriate rust toolchain  

```yaml
# Generated Rust template
tools:
  rust: "stable"

tasks:
  default:
    - mise install
    - cargo build
    - cargo test
```

#### Scenario: Minimal project template
**Given** no specific project files are detected  
**When** user chooses minimal template  
**Then** the system generates a basic configuration  
**And** includes placeholder commands with helpful comments  
**And** provides examples of common patterns  

```yaml
# Generated minimal template
tasks:
  default:
    # Add your project setup commands here
    - echo "Setting up project..."
    # Example: mise install
    # Example: npm install
    # Example: task setup

  dev:
    # Add your development workflow here
    - echo "Starting development..."
```

## MODIFIED Requirements

### Requirement: CONFIG_VALIDATION - Configuration validation SHALL support new task format
The configuration validation system SHALL properly validate both legacy and new task formats.

#### Scenario: Default task validation
**Given** a Razdfile.yml with `tasks.default`  
**When** the configuration is validated  
**Then** the system checks command syntax and structure  
**And** validates task reference integrity  
**And** reports specific errors with helpful suggestions  

#### Scenario: Mixed format validation warning
**Given** a Razdfile.yml with both `tasks.default` and `tasks.up`  
**When** the configuration is validated  
**Then** the system warns about redundant legacy tasks  
**And** suggests removing the legacy format  
**And** confirms that only `tasks.default` will be used  

#### Scenario: Legacy format validation with migration hint
**Given** a Razdfile.yml with only `tasks.up`  
**When** the configuration is validated  
**Then** the system validates the legacy format  
**And** includes migration recommendations in validation output  
**And** explains the advantages of migrating to new format