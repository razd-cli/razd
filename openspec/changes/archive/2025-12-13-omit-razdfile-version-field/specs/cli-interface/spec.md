# cli-interface Specification Delta

## MODIFIED Requirements

### Requirement: Razdfile.yml Configuration Format
The Razdfile.yml configuration file MUST support an optional `version` field with automatic defaults.

**Previous behavior:** The `version: '3'` field was implicitly required because serialized YAML was passed directly to taskfile.

**New behavior:** The `version` field becomes optional in user-facing Razdfile.yml files. When omitted, razd automatically defaults to version '3' during taskfile execution.

#### Scenario: User creates Razdfile.yml without version field
**Given** a user initializes a new Razdfile.yml  
**When** they define tasks without specifying a `version` field:
```yaml
tasks:
  default:
    desc: "Set up project"
    cmds:
      - mise install
      - echo "Done"
```
**Then** the system should:
- Successfully parse the configuration
- Automatically inject `version: '3'` when serializing for taskfile execution
- Execute tasks without errors

#### Scenario: User provides explicit version field (backward compatibility)
**Given** a user has an existing Razdfile.yml with explicit version  
**When** the configuration contains:
```yaml
version: '3'
tasks:
  default:
    desc: "Set up project"
    cmds:
      - mise install
```
**Then** the system should:
- Parse and respect the explicit version value
- Continue working exactly as before
- Maintain full backward compatibility

#### Scenario: Razdfile initialization omits version field
**Given** a user runs `razd init` to create a new Razdfile.yml  
**When** the configuration file is generated  
**Then** the system should:
- Create a Razdfile.yml without the `version: '3'` field
- Include only meaningful configuration (tasks, mise settings)
- Generate clean, minimal boilerplate

#### Scenario: Version field serialization for taskfile
**Given** razd needs to execute a task using taskfile  
**When** the Razdfile.yml is serialized to YAML for taskfile execution  
**Then** the system should:
- Ensure `version: '3'` is present in the serialized YAML
- Insert the field automatically if not present in the original config
- Pass valid Taskfile v3 format to the taskfile tool
