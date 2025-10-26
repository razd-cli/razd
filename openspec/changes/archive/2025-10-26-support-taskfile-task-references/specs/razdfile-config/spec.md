# Razdfile Configuration Capability

## MODIFIED Requirements

### Requirement: RAZDFILE_COMMAND_TYPES - System SHALL support multiple command types in Razdfile.yml
The system SHALL parse and execute both simple string commands and task reference commands in Razdfile.yml configurations.

#### Scenario: Simple string commands in task
**Given** a Razdfile.yml with task containing only string commands:
```yaml
version: '3'
tasks:
  install:
    desc: "Install tools"
    cmds:
      - echo "Installing..."
      - mise install
```
**When** user executes `razd up` or references this task  
**Then** the system should parse the commands successfully  
**And** execute each string command in sequence  
**And** complete without errors

#### Scenario: Task reference commands
**Given** a Razdfile.yml with task references:
```yaml
version: '3'
tasks:
  up:
    desc: "Setup project"
    cmds:
      - task: install
      - task: setup
  install:
    cmds:
      - mise install
```
**When** user executes `razd up`  
**Then** the system should parse the task references successfully  
**And** execute the referenced `install` task  
**And** execute the referenced `setup` task  
**And** complete the workflow without errors

#### Scenario: Mixed command types in single task
**Given** a Razdfile.yml with both string commands and task references:
```yaml
version: '3'
tasks:
  up:
    cmds:
      - echo "Starting setup..."
      - task: install
      - echo "Installation complete"
      - task: configure
```
**When** user executes `razd up`  
**Then** the system should parse all commands correctly  
**And** execute string commands and task references in order  
**And** maintain proper execution sequence  
**And** complete successfully

#### Scenario: Task reference with variables
**Given** a Razdfile.yml with task reference including variables:
```yaml
version: '3'
tasks:
  deploy:
    cmds:
      - task: build
        vars:
          ENV: production
          VERSION: v1.0.0
```
**When** user executes `razd deploy`  
**Then** the system should parse the task reference with variables  
**And** pass variables to the referenced task  
**And** execute with correct variable context

### Requirement: RAZDFILE_BACKWARD_COMPAT - System SHALL maintain backward compatibility with existing configs
The system SHALL continue to support existing Razdfile.yml files that only use string commands without requiring modifications.

#### Scenario: Existing config with string commands only
**Given** an existing Razdfile.yml created before task reference support:
```yaml
version: '3'
tasks:
  dev:
    cmds:
      - echo "Starting..."
      - npm run dev
```
**When** user executes `razd dev`  
**Then** the system should parse and execute without errors  
**And** behavior should be identical to previous version  
**And** no migration or config changes required

#### Scenario: Empty command list
**Given** a Razdfile.yml with empty commands:
```yaml
version: '3'
tasks:
  placeholder:
    desc: "Future task"
    cmds: []
```
**When** the system loads this configuration  
**Then** it should parse successfully  
**And** handle empty command lists gracefully

### Requirement: RAZDFILE_ERROR_HANDLING - System SHALL provide clear errors for invalid command formats
The system SHALL validate command structure and provide helpful error messages when commands are malformed.

#### Scenario: Invalid command structure
**Given** a Razdfile.yml with malformed command:
```yaml
version: '3'
tasks:
  bad:
    cmds:
      - invalid: structure
        unknown: field
```
**When** the system attempts to parse this configuration  
**Then** it should detect the invalid command format  
**And** display error message: "Configuration error: Invalid command format"  
**And** indicate which task contains the invalid command  
**And** suggest correct formats (string or task reference)

#### Scenario: Task reference without task field
**Given** a Razdfile.yml with incomplete task reference:
```yaml
version: '3'
tasks:
  bad:
    cmds:
      - vars:
          KEY: value
```
**When** the system attempts to parse this configuration  
**Then** it should detect the missing task field  
**And** display error message indicating task field is required  
**And** provide example of correct task reference format

### Requirement: RAZDFILE_YAML_SERIALIZATION - System SHALL serialize commands correctly for taskfile execution
The system SHALL convert parsed command structures back to valid YAML format for taskfile.dev execution.

#### Scenario: Serialize string commands to YAML
**Given** parsed commands containing string values  
**When** the system generates YAML for taskfile execution  
**Then** string commands should be serialized as:
```yaml
cmds:
  - echo "test"
  - mise install
```
**And** maintain exact string values  
**And** preserve command order

#### Scenario: Serialize task references to YAML
**Given** parsed commands containing task references  
**When** the system generates YAML for taskfile execution  
**Then** task references should be serialized as:
```yaml
cmds:
  - task: install
  - task: setup
```
**And** maintain proper YAML object structure  
**And** preserve task names and variables

#### Scenario: Serialize mixed commands to YAML
**Given** parsed commands with both strings and task references  
**When** the system generates YAML for taskfile execution  
**Then** the output should correctly serialize both types:
```yaml
cmds:
  - echo "Starting..."
  - task: install
  - echo "Done"
```
**And** preserve command order  
**And** maintain proper YAML formatting for taskfile compatibility

## ADDED Requirements

### Requirement: RAZDFILE_COMMAND_ENUM - System SHALL define flexible command representation
The system SHALL implement a Command enum type that supports multiple command formats using Rust's serde untagged deserialization.

#### Scenario: Command enum parses string variant
**Given** YAML input with string command: `- echo "test"`  
**When** serde deserializes the command  
**Then** it should create Command::String("echo \"test\"")  
**And** store the exact string value  
**And** complete without errors

#### Scenario: Command enum parses task reference variant
**Given** YAML input with task reference: `- task: install`  
**When** serde deserializes the command  
**Then** it should create Command::TaskRef with task="install"  
**And** vars should be None  
**And** complete without errors

#### Scenario: Command enum parses task reference with vars
**Given** YAML input:
```yaml
- task: deploy
  vars:
    ENV: prod
```
**When** serde deserializes the command  
**Then** it should create Command::TaskRef with task="deploy"  
**And** vars should contain HashMap with ENV=prod  
**And** complete without errors
