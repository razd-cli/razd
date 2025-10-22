# Add razd.yml Configuration System

## Summary

Implement `razd.yml` configuration system as a **template generator** for quickly setting up `Taskfile.yml` and `mise.toml` files. The goal is not to replace taskfile.dev, but to provide a faster way to bootstrap standard project configurations with `razd init`.

## Motivation

Currently, developers need to maintain separate configuration files:
- **mise.toml/.tool-versions**: Tool and runtime versions
- **Taskfile.yml**: Project tasks and build commands
- **Project-specific configs**: Package.json, requirements.txt, etc.

This leads to:
1. **Information duplication**: Same project metadata scattered across files
2. **Maintenance overhead**: Updates require changes in multiple places  
3. **Inconsistency risk**: Versions and settings can get out of sync
4. **Complex onboarding**: New developers need to understand multiple config formats

The `razd.yml` configuration will provide:
- **Single source of truth**: All project metadata in one place
- **Template generation**: Automatic creation of tool-specific configs
- **Consistency guarantee**: Generated files always match the central config
- **Simplified maintenance**: One file to update, all configs stay in sync

## Proposed Changes

### Core Configuration Format (First Concept)

```yaml
# Razdfile.yml - razd's own taskfile (executed via task --taskfile Razdfile.yml)

version: '3'

tasks:
  # razd up - полная настройка проекта
  up:
    desc: "Clone repository and set up project"
    cmds:
      - echo "🚀 Setting up project..."
      - task: clone-repo
      - task: install-tools
      - task: setup-deps
      
  # razd install - установка mise tools
  install:
    desc: "Install development tools via mise"
    cmds:
      - echo "📦 Installing tools..."
      - mise install
      
  # razd setup - установка зависимостей
  setup:
    desc: "Install project dependencies"
    cmds:
      - echo "⚙️ Setting up dependencies..."
      - task: "setup" --taskfile Taskfile.yml  # запускаем setup из проектного Taskfile.yml
      
  # razd task - proxy to project Taskfile.yml
  dev:
    desc: "Start development server"
    cmds:
      - task: "default" --taskfile Taskfile.yml
      
  # Внутренние задачи
  clone-repo:
    internal: true
    cmds:
      - echo "Cloning repository..."
      
  install-tools:
    internal: true
    cmds:
      - mise install
      
  setup-deps:
    internal: true
    cmds:
      - task: "setup" --taskfile Taskfile.yml
```

### Enhanced razd init Command

- Create `Razdfile.yml` with project-specific tasks
- Detect project type and generate appropriate templates
- Generate `Taskfile.yml` and `mise.toml` if missing
- Interactive prompts for project configuration

### Smart Task Execution

- **Razdfile.yml**: razd's own taskfile with project-specific logic
- **Task delegation**: razd tasks call project's Taskfile.yml when needed
- **Standard tooling**: Uses `task --taskfile Razdfile.yml` under the hood

## Implementation Approach

1. **Configuration parsing**: YAML deserialization with serde
2. **Template engine**: Simple string templating for file generation
3. **File detection**: Analyze existing project structure
4. **Interactive prompts**: User-friendly configuration wizard
5. **Validation**: Ensure generated configs are valid

## Impact Assessment

### Benefits
- **Centralized metadata**: Project info and tool versions in one place
- **Smart task mapping**: razd knows which tasks to run for each operation
- **Reduced configuration**: Generate mise.toml automatically
- **Flexible settings**: Control razd behavior per project
- **No duplication**: Taskfile.yml remains the source of truth for tasks

### Risks
- **Additional file**: One more config file to maintain
- **Learning curve**: Users need to understand task mapping concept
- **Complexity**: Mapping between razd commands and task names

### Migration Strategy
- **Backward compatibility**: Continue supporting existing mise.toml/Taskfile.yml
- **Gradual adoption**: `razd.yml` is optional, enhances existing workflow
- **Import command**: Generate `razd.yml` from existing configurations

## Success Criteria

- `razd init` creates minimal `razd.yml` with task mappings
- `razd task` uses task mapping to run correct Taskfile tasks
- `razd task` (без аргументов) запускает default_task из настроек
- `razd up` respects auto-install and auto-setup settings
- Existing Taskfile.yml projects work without modification
- Configuration is simple and focused only on razd behavior