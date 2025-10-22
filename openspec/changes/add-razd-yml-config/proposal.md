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

### Core Configuration Format

#### Built-in Default Workflows (встроенные в razd)
```rust
// src/defaults.rs - встроенные workflows
const DEFAULT_WORKFLOWS: &str = r#"
version: '3'

tasks:
  up:
    desc: "Clone repository and set up project"
    cmds:
      - echo "🚀 Setting up project..."
      - mise install
      - task: setup --taskfile Taskfile.yml
      
  install:
    desc: "Install development tools via mise"
    cmds:
      - echo "📦 Installing tools..."
      - mise install
      
  dev:
    desc: "Start development workflow"
    cmds:
      - echo "🚀 Starting development..."
      - task: default --taskfile Taskfile.yml
      
  build:
    desc: "Build project"
    cmds:
      - task: build --taskfile Taskfile.yml
"#;
```

#### Optional Razdfile.yml (переопределяет дефолты)
```yaml
# Razdfile.yml - опциональный файл для кастомизации workflows
version: '3'

tasks:
  # Переопределяем только нужные команды
  up:
    desc: "Custom project setup"
    cmds:
      - echo "🚀 Custom setup workflow..."
      - task: clone-repo
      - task: install-tools  
      - task: setup-deps
      - task: custom-init
      
  # Добавляем новые команды
  deploy:
    desc: "Deploy to production"
    cmds:
      - task: build --taskfile Taskfile.yml
      - task: docker-push
      
  # Внутренние задачи
  clone-repo:
    internal: true
    cmds:
      - echo "Cloning repository..."
      
  custom-init:
    internal: true
    cmds:
      - echo "Running custom initialization..."
```

#### Command Delegation Logic

**Workflow Commands** (with fallback priority):
1. **`razd up/install/dev/build`** → `task --taskfile Razdfile.yml <command>` (если файл существует)
2. **`razd up/install/dev/build`** → встроенные дефолтные workflows (если Razdfile.yml отсутствует)

**Direct Task Delegation**:
- **`razd task <anything>`** → `task <anything>` (прямое делегирование на проектный Taskfile.yml)

### Enhanced razd init Command

**Optional Configuration Generation**:
- **`razd init`** - работает с встроенными дефолтами (файлы не создаются)
- **`razd init --config`** - создает `Razdfile.yml` для кастомизации workflows
- **`razd init --full`** - создает все файлы (`Razdfile.yml`, `Taskfile.yml`, `mise.toml`)

**Smart Detection**:
- Detect project type and generate appropriate templates
- Generate `Taskfile.yml` and `mise.toml` only if missing or requested
- Interactive prompts for project configuration

### Two-Level Task System with Built-in Fallbacks

**Level 1: razd workflows** (with fallback chain)
- `razd up/install/dev/build` - ищут команды в следующем порядке:
  1. **Razdfile.yml** (если существует) → `task --taskfile Razdfile.yml <command>`
  2. **Built-in defaults** (встроенные в razd) → выполняются напрямую
- **Customization**: Razdfile.yml переопределяет только нужные команды

**Level 2: Project tasks** (direct delegation)  
- `razd task <anything>` - полностью делегируется на проектный Taskfile.yml
- Выполняется: `task <anything>` (как будто просто запустили task)

**Benefits:**
- **Zero config**: razd работает из коробки без файлов конфигурации 
- **Progressive enhancement**: добавляешь Razdfile.yml только при необходимости
- **Workflows** - razd управляет сложными процессами (up, install)
- **Tasks** - прямой доступ ко всем проектным задачам  
- **Transparency** - пользователь видит что именно выполняется

## Implementation Approach

1. **Built-in defaults**: Embedded workflows in Rust code as fallback
2. **Configuration parsing**: Optional YAML deserialization with serde
3. **Fallback chain**: Razdfile.yml → built-in defaults → error
4. **Template engine**: Simple string templating for file generation
5. **File detection**: Analyze existing project structure  
6. **Interactive prompts**: User-friendly configuration wizard
7. **Validation**: Ensure generated configs are valid

## Impact Assessment

### Benefits
- **Zero configuration**: Works out-of-the-box without any config files
- **Progressive enhancement**: Add Razdfile.yml only when customization needed
- **Smart task mapping**: razd knows which tasks to run for each operation
- **Flexible overrides**: Only override workflows that need customization
- **No duplication**: Taskfile.yml remains the source of truth for tasks
- **Built-in intelligence**: Common workflows embedded in razd itself

### Risks
- **Optional complexity**: Razdfile.yml adds complexity only when needed
- **Learning curve**: Users need to understand workflow vs task concepts
- **Fallback debugging**: Need clear indication when using built-in vs custom workflows

### Migration Strategy
- **Zero migration needed**: razd works with existing projects immediately
- **Backward compatibility**: Continue supporting existing mise.toml/Taskfile.yml
- **Optional enhancement**: Razdfile.yml is purely additive, never required
- **Export command**: Generate Razdfile.yml from built-in defaults when customization needed

## Success Criteria

- **Zero config**: `razd up/install/dev/build` work without any config files
- **Optional config**: `razd init --config` creates `Razdfile.yml` for customization
- **Fallback chain**: Commands work with Razdfile.yml → built-in defaults → error
- **Direct delegation**: `razd task <anything>` delegates to `task <anything>`
- **Standard compatibility**: Generated files work with mise/taskfile tools
- **No breaking changes**: Existing projects work without modification
- **Clear feedback**: Users know when using built-in vs custom workflows