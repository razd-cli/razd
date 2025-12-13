# tool-integration Specification Delta

## MODIFIED Requirements

### Requirement: Validate tool and plugin names

The system SHALL validate tool and plugin names against mise naming rules, supporting backend-specific syntax.

#### Scenario: Validate valid tool names without prefix

**Given** tool names without a backend prefix  
**When** razd validates the configuration  
**Then** the system should:

- Accept names containing alphanumeric characters, hyphens, and underscores
- Accept names like "node", "python-3", "my_tool", "rust-1-74"
- Reject names with spaces or special characters

#### Scenario: Validate tool names with backend prefix

**Given** tool names with a recognized backend prefix  
**When** razd validates the configuration  
**Then** the system should:

- Accept prefixes: `npm:`, `pipx:`, `cargo:`, `go:`, `gem:`, `aqua:`, `ubi:`, `github:`, `gitlab:`, `spm:`, `asdf:`, `vfox:`, `vfox-backend:`, `http:`, `core:`
- Apply backend-specific validation rules after the prefix
- Continue configuration processing

#### Scenario: Validate scoped npm packages

**Given** tool names using npm backend with scoped packages  
**When** razd validates names like:

```yaml
mise:
  tools:
    "npm:@fission-ai/openspec": latest
    "npm:@babel/cli": "7.0.0"
    "npm:cowsay": latest
```

**Then** the system should:

- Accept `@` for npm scope prefix
- Accept `/` to separate scope from package name
- Accept alphanumeric, hyphens, underscores in scope and package names
- Pass validation successfully

#### Scenario: Validate aqua/github repository format

**Given** tool names using aqua or github backends  
**When** razd validates names like:

```yaml
mise:
  tools:
    "aqua:cli/cli": latest
    "github:jdx/mise": "2024.0.0"
    "ubi:sharkdp/fd": latest
```

**Then** the system should:

- Accept `owner/repo` format
- Accept alphanumeric, hyphens, underscores, dots in owner and repo
- Pass validation successfully

#### Scenario: Validate go module paths

**Given** tool names using go backend  
**When** razd validates names like:

```yaml
mise:
  tools:
    "go:github.com/golangci/golangci-lint/cmd/golangci-lint": latest
    "go:golang.org/x/tools/gopls": latest
```

**Then** the system should:

- Accept full go module paths with domain and path components
- Accept dots, slashes, hyphens, underscores
- Pass validation successfully

#### Scenario: Reject invalid tool names

**Given** tool names that don't match any valid pattern  
**When** razd validates names like:

- Tool without prefix containing spaces: "my tool"
- Tool without prefix containing `@`: "my@tool"
- Empty tool name: ""
  **Then** the system should:
- Display validation error for invalid names
- Show which tool names are problematic
- Provide guidance on valid naming patterns for the appropriate backend
- Fail configuration parsing
