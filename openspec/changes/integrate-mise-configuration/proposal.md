# Proposal: Integrate mise Configuration into Razdfile.yml

## Overview
This change proposes integrating mise tool and plugin configuration directly into `Razdfile.yml`, allowing developers to manage their development tools in a single configuration file. The system will automatically generate and synchronize `mise.toml` from the Razdfile configuration, with intelligent change detection to prevent configuration drift.

## Motivation
Currently, projects using razd and mise require two separate configuration files:
- `Razdfile.yml` for task definitions
- `mise.toml` or `.tool-versions` for tool management

This separation creates several issues:
- **Configuration fragmentation**: Developers must manage multiple files for project setup
- **Manual synchronization**: Changes to tool requirements must be manually reflected in both files
- **Onboarding friction**: New team members need to understand multiple configuration formats
- **Version drift**: Direct edits to `mise.toml` can diverge from intended configuration

## Proposed Solution
Extend `Razdfile.yml` to include mise configuration sections (`mise.tools` and `mise.plugins`) that mirror the functionality of `mise.toml`. razd will:

1. **Read mise configuration from Razdfile.yml**: Parse tool and plugin definitions in YAML format
2. **Generate mise.toml**: Automatically create/update `mise.toml` when Razdfile changes
3. **Track file modifications**: Store metadata about file modification times to detect changes
4. **Prompt for synchronization**: When `mise.toml` is manually edited, offer to sync back to Razdfile on next razd command
5. **Maintain backward compatibility**: Continue supporting standalone `mise.toml` files

### Example Configuration
```yaml
version: '3'

mise:
  tools:
    node:
      version: "22"
      postinstall: "corepack enable"
    python: "3.11"
    rust: "latest"
    go:
      version: "1.21"
      os: ["linux", "macos"]
  
  plugins:
    elixir: "https://github.com/my-org/mise-elixir.git"
    node: "https://github.com/my-org/mise-node.git#DEADBEEF"

tasks:
  default:
    desc: "Setup and start development"
    cmds:
      - mise install
      - npm install
      - npm run dev
```

## Benefits
- **Single source of truth**: All project configuration in one file
- **Automatic synchronization**: No manual mise.toml management
- **Change detection**: Intelligent handling of manual mise.toml edits
- **Better developer experience**: Simpler project setup and maintenance
- **Team consistency**: Easier to enforce tool versions across team

## Risks and Mitigations
| Risk | Mitigation |
|------|------------|
| Breaking changes for existing projects | Maintain full backward compatibility with standalone mise.toml |
| Sync conflicts between files | Clear precedence rules and user prompts before overwriting |
| Complex edge cases in mise.toml format | Start with core features (tools, plugins) and iterate |
| Performance overhead from file tracking | Use efficient file metadata storage in user cache directory |

## Out of Scope
- Syncing other mise.toml sections (env, tasks, settings, aliases) - focus on tools and plugins only
- IDE integration and schema validation - can be added in future iterations
- Automatic migration of existing mise.toml files - manual adoption required

## Success Criteria
1. Razdfile.yml can fully express tool and plugin configuration
2. mise.toml is automatically generated with correct TOML syntax
3. File change detection accurately identifies modifications
4. User prompts are clear and non-intrusive
5. No breaking changes to existing razd functionality
6. Integration tests cover all synchronization scenarios

## Related Work
- Existing tool-integration spec: defines current mise integration
- task-auto-installation spec: establishes pattern for auto-installing tools
- Razdfile.yml structure: already uses Taskfile v3 format

## Timeline
Estimated implementation: 2-3 weeks
- Week 1: File tracking and sync logic
- Week 2: Razdfile.yml parsing and mise.toml generation
- Week 3: Integration, testing, and documentation
