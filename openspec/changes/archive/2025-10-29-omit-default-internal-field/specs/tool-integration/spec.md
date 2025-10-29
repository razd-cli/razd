# tool-integration Spec Delta

## MODIFIED Requirements

### Requirement: Razdfile.yml task configuration
The system MUST serialize task configurations without unnecessary default values.

#### Scenario: Serialize task without internal field when false
**Given** a task configuration with `internal: false` (default value)  
**When** the task is serialized to Razdfile.yml  
**Then** the system should:
- Omit the `internal` field from YAML output
- Produce clean, minimal YAML
- Maintain semantic equivalence (default is still false)

#### Scenario: Serialize task with internal field when true
**Given** a task configuration with `internal: true`  
**When** the task is serialized to Razdfile.yml  
**Then** the system should:
- Include the `internal: true` field in YAML output
- Mark the task as internal/hidden

#### Scenario: Deserialize task with explicit internal false
**Given** a Razdfile.yml contains `internal: false` explicitly  
**When** the configuration is loaded  
**Then** the system should:
- Parse the field correctly
- Set task.internal to false
- Maintain backwards compatibility with existing files

#### Scenario: Deserialize task without internal field
**Given** a Razdfile.yml task definition lacks the `internal` field  
**When** the configuration is loaded  
**Then** the system should:
- Default internal to false
- Parse the task correctly
- Treat as a public/visible task

#### Scenario: Sync preserves clean task configuration
**Given** a Razdfile.yml with tasks containing no `internal` fields  
**When** mise sync operation serializes tasks back to YAML  
**Then** the system should:
- Not add `internal: false` lines
- Maintain clean, minimal YAML structure
- Preserve user's original formatting intent
