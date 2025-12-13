# tool-integration Specification Delta

## MODIFIED Requirements

### Requirement: Taskfile Integration with Automatic Version Handling
The system MUST automatically inject the required `version` field when executing taskfile commands, even when the field is omitted from Razdfile.yml.

**Context:** Taskfile.dev requires a `version: '3'` field in its YAML configuration. However, since razd manages the interaction with taskfile, this implementation detail can be hidden from users by automatically injecting the field during serialization.

#### Scenario: Execute task from Razdfile without version field
**Given** a Razdfile.yml exists without an explicit version field:
```yaml
tasks:
  build:
    desc: "Build the project"
    cmds:
      - cargo build
```
**When** a user runs `razd task build`  
**Then** the system should:
- Load the Razdfile.yml configuration
- Serialize it to YAML with `version: '3'` automatically injected at the top
- Pass the complete YAML to taskfile for execution
- Execute the build task successfully

#### Scenario: Workflow execution defaults version field
**Given** razd needs to execute a workflow task (default, dev, build, install)  
**When** `get_workflow_config()` serializes the Razdfile to YAML  
**Then** the system should:
- Ensure the serialized YAML contains `version: '3'` as the first field
- Generate valid Taskfile v3 format regardless of source configuration
- Support both temporary workflow files and direct Razdfile.yml usage

#### Scenario: Default workflows continue to work
**Given** a project has no Razdfile.yml and relies on built-in defaults  
**When** razd falls back to `DEFAULT_WORKFLOWS` constant  
**Then** the system should:
- Continue using the embedded workflow YAML that includes `version: '3'`
- Maintain backward compatibility with existing default behavior
- Execute default, install, dev, and build workflows successfully

#### Scenario: Preserve explicit version when present
**Given** a Razdfile.yml explicitly specifies `version: '3'`  
**When** razd serializes the configuration for taskfile execution  
**Then** the system should:
- Respect and preserve the user-specified version value
- Not override or duplicate the version field
- Maintain full backward compatibility
