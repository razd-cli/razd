# Design: Smart Up Command with Default Task

## Context

This change transforms the `razd up` command from a simple workflow executor into an intelligent, context-aware project setup tool. The change addresses user confusion around command responsibilities and introduces a more intuitive configuration format.

## Goals / Non-Goals

### Goals
- Unify project setup workflows under a single intuitive command
- Provide intelligent behavior based on directory context
- Smooth migration path for existing users
- Improved configuration naming conventions
- Reduced cognitive load for new users

### Non-Goals
- Complete rewrite of task execution system
- Changes to underlying task runner integrations (mise, taskfile)
- Breaking changes to tool integrations
- Performance optimization (unless regression occurs)

## Technical Decisions

### 1. Context Detection Strategy

**Decision**: Use filesystem scanning with caching for context detection

**Implementation**:
```rust
struct ProjectContext {
    has_razdfile: bool,
    razdfile_path: Option<PathBuf>,
    detected_tools: Vec<ProjectTool>,
    suggested_template: TemplateType,
}

enum ProjectTool {
    Node(NodeConfig),
    Rust(RustConfig),
    Python(PythonConfig),
    Generic,
}
```

**Rationale**: 
- Fast detection without external dependencies
- Extensible for future project types
- Minimal performance impact through caching
- Clear separation of concerns

**Alternatives Considered**:
- External tool detection services: Rejected (adds dependencies)
- Database-backed context caching: Rejected (overcomplicated)
- Always prompt user: Rejected (poor UX for obvious cases)

### 2. Configuration Format Migration

**Decision**: Support both formats during transition with priority system

**Implementation**:
```rust
#[derive(Deserialize)]
struct TaskConfig {
    #[serde(rename = "default")]
    default_task: Option<Vec<Command>>,
    
    #[serde(rename = "up")]
    up_task: Option<Vec<Command>>, // Legacy support
    
    // other tasks...
}

impl TaskConfig {
    fn get_primary_task(&self) -> Option<&Vec<Command>> {
        self.default_task.as_ref()
            .or_else(|| {
                if self.up_task.is_some() {
                    warn!("Using legacy 'up' task. Consider migrating to 'default'");
                }
                self.up_task.as_ref()
            })
    }
}
```

**Rationale**:
- Backward compatibility during transition
- Clear migration path with helpful warnings
- Zero-downtime migration for existing projects
- Future removal of legacy support is straightforward

### 3. Interactive Configuration Creation

**Decision**: Template-based generation with intelligent defaults

**Implementation**:
```rust
struct ConfigGenerator {
    detected_context: ProjectContext,
    user_preferences: UserPreferences,
}

impl ConfigGenerator {
    fn generate_template(&self) -> Result<RazdfileConfig> {
        match self.detected_context.suggested_template {
            TemplateType::Node => self.generate_node_template(),
            TemplateType::Rust => self.generate_rust_template(),
            TemplateType::Python => self.generate_python_template(),
            TemplateType::Minimal => self.generate_minimal_template(),
        }
    }
}
```

**Rationale**:
- Reduces configuration burden for common project types
- Extensible template system for future project types
- Maintains user control over generated configuration
- Provides good defaults while allowing customization

### 4. Command Structure Changes

**Decision**: Remove `razd init` entirely, enhance `razd up` with smart behavior

**Implementation**:
```rust
// Before
match command {
    Commands::Up { url } => execute_up(url).await,
    Commands::Init { .. } => execute_init().await,
    // ...
}

// After  
match command {
    Commands::Up { url } => execute_smart_up(url).await,
    // Commands::Init removed
    // ...
}

async fn execute_smart_up(url: Option<String>) -> Result<()> {
    match url {
        Some(repo_url) => clone_and_setup(repo_url).await,
        None => setup_local_project().await,
    }
}
```

**Rationale**:
- Reduces command surface area
- Eliminates user decision about which command to use
- More intuitive behavior matches user expectations
- Single responsibility: "bring up the project"

## Architecture Impact

### Module Structure
```
src/
├── commands/
│   ├── up.rs           # Enhanced with smart behavior
│   ├── init.rs         # Removed
│   └── ...
├── config/
│   ├── detection.rs    # New: Project context detection
│   ├── templates.rs    # New: Configuration templates
│   ├── migration.rs    # New: Legacy format handling
│   └── razdfile.rs     # Enhanced: Dual format support
└── ...
```

### Data Flow
1. **Context Detection**: Scan directory for project indicators
2. **Decision Logic**: Choose behavior based on context and arguments
3. **Action Execution**: Execute appropriate workflow
4. **Feedback Loop**: Provide user guidance and migration suggestions

## Migration Plan

### Phase 1: Foundation (Week 1)
- Add `tasks.default` support alongside `tasks.up`
- Implement context detection system
- Create template generation system

### Phase 2: Smart Behavior (Week 2)
- Enhance `razd up` with context-aware logic
- Add interactive configuration creation
- Implement migration helpers

### Phase 3: Deprecation (Week 3)
- Add deprecation warnings for `razd init`
- Add migration suggestions for `tasks.up`
- Update documentation and examples

### Phase 4: Breaking Changes (Week 4)
- Remove `razd init` command
- Remove `tasks.up` support
- Clean up legacy code

## Risks / Trade-offs

### Risks
1. **Migration Friction**: Existing users need to update workflows
   - **Mitigation**: Clear migration guide, automated tools, gradual deprecation

2. **Increased Complexity**: Smart behavior adds complexity to up command
   - **Mitigation**: Well-tested logic, clear error messages, comprehensive docs

3. **Context Detection Failures**: May not detect project type correctly
   - **Mitigation**: Conservative defaults, user override options, learning from feedback

### Trade-offs
1. **Breaking Changes vs UX**: Accepting breaking changes for better long-term UX
2. **Simplicity vs Intelligence**: Adding complexity to provide smarter defaults
3. **Backward Compatibility vs Clean Design**: Time-limited compatibility for smooth transition

## Performance Considerations

- **Context Detection**: O(1) directory scan with caching
- **Template Generation**: Pre-computed templates, minimal runtime cost
- **Migration Logic**: Only active during transition period
- **Memory Usage**: Minimal increase for context caching

## Security Considerations

- **Template Generation**: No user input injection in generated configs
- **File Operations**: Safe file writing with atomic operations and backups
- **URL Handling**: Existing git clone security applies unchanged

## Testing Strategy

1. **Unit Tests**: Context detection, template generation, migration logic
2. **Integration Tests**: End-to-end workflows for all scenarios
3. **Migration Tests**: Backward compatibility during transition
4. **Performance Tests**: Context detection and template generation speed
5. **User Acceptance Tests**: Real-world workflow validation

## Open Questions

1. **Template Customization**: How much customization should be allowed during interactive setup?
2. **Context Caching**: Should context detection results be cached between runs?
3. **Migration Timeline**: How long should backward compatibility be maintained?
4. **Error Recovery**: How should the system handle partial migrations or setup failures?

These will be resolved during implementation based on user feedback and technical constraints.