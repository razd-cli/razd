# mise-configuration Capability Specification

## Purpose
Enable razd to parse, validate, and manage mise tool and plugin configurations defined in Razdfile.yml, providing a unified configuration experience for development tool management.

## Requirements

### ADDED Requirement: Parse mise tools from Razdfile.yml
The system SHALL parse mise tool configurations from the `mise.tools` section of Razdfile.yml.

#### Scenario: Parse simple tool version
**Given** a Razdfile.yml contains:
```yaml
mise:
  tools:
    node: "22"
    python: "3.11"
```
**When** razd parses the configuration  
**Then** the system should:
- Extract tool name "node" with version "22"
- Extract tool name "python" with version "3.11"
- Store configurations for mise.toml generation

#### Scenario: Parse complex tool configuration
**Given** a Razdfile.yml contains:
```yaml
mise:
  tools:
    node:
      version: "22"
      postinstall: "corepack enable"
      os: ["linux", "macos"]
    go:
      version: "1.21"
      install_env:
        CGO_ENABLED: "1"
```
**When** razd parses the configuration  
**Then** the system should:
- Extract node version "22" with postinstall command
- Extract node OS restrictions ["linux", "macos"]
- Extract go version "1.21" with install_env variable
- Preserve all configuration options for TOML generation

#### Scenario: Handle missing mise section
**Given** a Razdfile.yml contains no `mise` section  
**When** razd parses the configuration  
**Then** the system should:
- Continue without error
- Not generate or modify mise.toml
- Use existing mise.toml if present

### ADDED Requirement: Parse mise plugins from Razdfile.yml
The system SHALL parse mise plugin configurations from the `mise.plugins` section of Razdfile.yml.

#### Scenario: Parse plugin URLs
**Given** a Razdfile.yml contains:
```yaml
mise:
  plugins:
    elixir: "https://github.com/my-org/mise-elixir.git"
    node: "https://github.com/my-org/mise-node.git#DEADBEEF"
    "vfox-backend:myplugin": "https://github.com/jdx/vfox-npm"
```
**When** razd parses the configuration  
**Then** the system should:
- Extract plugin name "elixir" with repository URL
- Extract plugin name "node" with URL and git ref (DEADBEEF)
- Extract plugin name with type prefix "vfox-backend:myplugin"
- Validate URL format before storage

#### Scenario: Handle invalid plugin URLs
**Given** a Razdfile.yml contains:
```yaml
mise:
  plugins:
    invalid: "not-a-url"
```
**When** razd parses the configuration  
**Then** the system should:
- Display validation error with line number
- Indicate invalid URL format
- Fail configuration parsing with clear error message

### ADDED Requirement: Generate mise.toml from Razdfile configuration
The system SHALL generate a valid mise.toml file from Razdfile mise configuration.

#### Scenario: Generate tools section
**Given** a Razdfile.yml with mise tools configuration  
**When** razd generates mise.toml  
**Then** the system should:
- Create [tools] section in TOML format
- Convert simple versions to TOML string values
- Convert complex configurations to TOML inline tables
- Ensure valid TOML syntax for all tool configurations

#### Scenario: Generate plugins section
**Given** a Razdfile.yml with mise plugins configuration  
**When** razd generates mise.toml  
**Then** the system should:
- Create [plugins] section in TOML format
- Map plugin names to repository URLs
- Preserve git refs and type prefixes
- Quote URLs properly for TOML format

#### Scenario: Generate complete mise.toml
**Given** a Razdfile.yml with both tools and plugins:
```yaml
mise:
  tools:
    node: "22"
    python:
      version: "3.11"
      os: ["linux"]
  plugins:
    elixir: "https://github.com/my-org/mise-elixir.git"
```
**When** razd generates mise.toml  
**Then** the generated file should contain:
```toml
[tools]
node = "22"
python = { version = "3.11", os = ["linux"] }

[plugins]
elixir = "https://github.com/my-org/mise-elixir.git"
```

#### Scenario: Handle empty mise configuration
**Given** a Razdfile.yml with empty mise section:
```yaml
mise:
  tools: {}
```
**When** razd checks for mise.toml generation  
**Then** the system should:
- Not generate or modify mise.toml
- Skip mise.toml generation step
- Continue without warnings

### ADDED Requirement: Validate tool and plugin names
The system SHALL validate tool and plugin names against mise naming rules.

#### Scenario: Validate valid tool names
**Given** tool names contain alphanumeric characters, hyphens, and underscores  
**When** razd validates the configuration  
**Then** the system should:
- Accept names like "node", "python-3", "my_tool"
- Allow plugin type prefixes like "asdf:", "vfox:", "vfox-backend:"
- Continue configuration processing

#### Scenario: Reject invalid tool names
**Given** tool names contain spaces or special characters  
**When** razd validates the configuration  
**Then** the system should:
- Display validation error for invalid names
- Show which tool names are problematic
- Provide guidance on valid naming patterns
- Fail configuration parsing

### ADDED Requirement: Support OS-specific tool configurations
The system SHALL respect OS restrictions defined in tool configurations.

#### Scenario: Parse OS restrictions
**Given** a tool configuration with OS restrictions:
```yaml
mise:
  tools:
    go:
      version: "1.21"
      os: ["linux", "macos"]
```
**When** razd parses the configuration  
**Then** the system should:
- Extract OS list ["linux", "macos"]
- Store OS restrictions for TOML generation
- Preserve exact OS names

#### Scenario: Generate OS restrictions in TOML
**Given** a tool with OS restrictions  
**When** razd generates mise.toml  
**Then** the generated TOML should contain:
```toml
[tools]
go = { version = "1.21", os = ["linux", "macos"] }
```

### MODIFIED Requirement: Mise integration (from tool-integration spec)
The system MUST integrate with mise for development tool management, now supporting both standalone mise.toml and Razdfile.yml mise configuration.

#### Scenario: Execute mise install with Razdfile mise config
**Given** a Razdfile.yml contains mise configuration  
**And** mise.toml exists and is up-to-date  
**When** razd executes the install phase  
**Then** the system should:
- Run `mise install` command using the generated mise.toml
- Display mise output to the user
- Handle any errors from mise installation

#### Scenario: Prefer Razdfile mise config over standalone mise.toml
**Given** both Razdfile.yml with mise config and standalone mise.toml exist  
**And** Razdfile.yml mise config is more recent  
**When** razd processes configuration  
**Then** the system should:
- Use Razdfile.yml as the source of truth
- Regenerate mise.toml from Razdfile
- Update file tracking metadata

### ADDED Requirement: Support install_env for tools
The system SHALL support environment variables during tool installation via install_env.

#### Scenario: Parse install_env configuration
**Given** a Razdfile.yml contains:
```yaml
mise:
  tools:
    python:
      version: "3.11"
      install_env:
        PYTHON_CONFIGURE_OPTS: "--enable-optimizations"
        CFLAGS: "-O2"
```
**When** razd parses the configuration  
**Then** the system should:
- Extract install_env as a map of key-value pairs
- Preserve all environment variable definitions
- Store for TOML generation

#### Scenario: Generate install_env in TOML
**Given** a tool with install_env configuration  
**When** razd generates mise.toml  
**Then** the generated TOML should contain:
```toml
[tools]
python = { version = "3.11", install_env = { PYTHON_CONFIGURE_OPTS = "--enable-optimizations", CFLAGS = "-O2" } }
```
