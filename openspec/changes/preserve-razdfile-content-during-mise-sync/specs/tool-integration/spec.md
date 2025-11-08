# tool-integration Specification Deltas

## ADDED Requirements

### Requirement: Surgical Mise Section Updates
When synchronizing from `mise.toml` to `Razdfile.yml`, the system MUST update only the `mise:` section while preserving all other content exactly as authored.

#### Scenario: Platform-specific commands are preserved during sync
```yaml
# Given: Razdfile.yml with platform-specific commands
version: '3'
tasks:
  install:
    desc: Install tools and dependencies
    cmds:
    - cmd: scoop install gcc make
      platform: windows
    - mise install

# And: mise.toml with new tool configuration
[tools]
node = "18"
python = "3.11"

# When: Synchronization from mise.toml to Razdfile.yml occurs
# Then: Platform command metadata is preserved exactly
version: '3'
mise:
  tools:
    node: "18"
    python: "3.11"
tasks:
  install:
    desc: Install tools and dependencies
    cmds:
    - cmd: scoop install gcc make
      platform: windows  # ✅ PRESERVED
    - mise install
```

#### Scenario: YAML formatting and structure preservation
```yaml
# Given: Razdfile.yml with specific formatting and comments
version: '3'

# Development tools configuration
mise:
  tools:
    lua: latest

tasks:
  # Main setup task
  default:
    desc: Set up project
    cmds:
    - mise install
    - echo "Project setup completed!"

# When: Synchronization updates mise section
# Then: Comments, spacing, and structure are preserved
version: '3'

# Development tools configuration
mise:
  tools:
    node: "18"        # ✅ Only this section changes
    python: "3.11"

tasks:
  # Main setup task         # ✅ Comments preserved
  default:                  # ✅ Formatting preserved
    desc: Set up project
    cmds:
    - mise install
    - echo "Project setup completed!"
```

### Requirement: Task Content Immutability During Mise Sync
When synchronizing from `mise.toml`, the system MUST NOT modify, reorder, or reformat any content in the `tasks:` section.

#### Scenario: Task order preservation
```yaml
# Given: Tasks in specific developer-intended order
tasks:
  setup:    # Listed first intentionally
    desc: Initial setup
  build:    # Listed second for workflow
    desc: Build project  
  test:     # Listed third in pipeline
    desc: Run tests

# When: Mise sync occurs
# Then: Task order remains exactly the same
tasks:
  setup:    # ✅ Still first
    desc: Initial setup
  build:    # ✅ Still second  
    desc: Build project
  test:     # ✅ Still third
    desc: Run tests
```

#### Scenario: Complex command structures preserved
```yaml
# Given: Tasks with complex command structures
tasks:
  multi-platform:
    desc: Cross-platform setup
    cmds:
    - cmd: apt-get update
      platform: linux
    - cmd: brew install node
      platform: darwin  
    - cmd: scoop install nodejs
      platform: windows
    - task: common-setup
      vars:
        ENV: development

# When: Mise sync occurs  
# Then: All command complexity is preserved exactly
tasks:
  multi-platform:
    desc: Cross-platform setup
    cmds:
    - cmd: apt-get update
      platform: linux      # ✅ PRESERVED
    - cmd: brew install node  
      platform: darwin     # ✅ PRESERVED
    - cmd: scoop install nodejs
      platform: windows    # ✅ PRESERVED  
    - task: common-setup
      vars:                 # ✅ PRESERVED
        ENV: development
```

### Requirement: Non-Mise Section Immunity
When synchronizing from `mise.toml`, the system MUST NOT modify any sections other than `mise:` including `env:`, `vars:`, `version:`, and task definitions.

#### Scenario: Environment variables untouched
```yaml
# Given: Razdfile with environment configuration
version: '3'
env:
  NODE_ENV: development
  DEBUG: "myapp:*"
vars:
  PROJECT_NAME: myproject
  VERSION: "1.0.0"

# When: Mise sync adds tools
# Then: env and vars sections remain untouched
version: '3'                    # ✅ UNCHANGED
env:                           # ✅ UNCHANGED  
  NODE_ENV: development
  DEBUG: "myapp:*"
vars:                          # ✅ UNCHANGED
  PROJECT_NAME: myproject
  VERSION: "1.0.0"
mise:                          # ✅ ONLY THIS CHANGES
  tools:
    node: "18"
```

## MODIFIED Requirements

### Requirement: Enhanced Sync Result Reporting  
The sync operation MUST provide clear feedback about what was modified, emphasizing the selective nature of the update.

#### Scenario: Informative sync messages
```bash
# When: Successful selective sync occurs
# Then: Message clarifies scope of changes
✓ Synced mise.toml → Razdfile.yml (mise section only)
  - Updated tools: node, python  
  - Preserved: 3 tasks, env variables, formatting
```

#### Scenario: Conflict detection with selective context
```bash
# When: Conflict is detected during selective sync
# Then: Messaging reflects selective update scope  
⚠️  Conflict detected in mise configuration only
   Razdfile mise section and mise.toml both changed
   Other Razdfile content will be preserved regardless of choice
```