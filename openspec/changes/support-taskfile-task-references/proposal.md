# Support Taskfile Task References in Razdfile.yml

## Summary

Add support for Taskfile.dev's native task reference syntax (`- task: task-name`) in Razdfile.yml command lists. Currently, only simple string commands are supported, which causes parsing errors when users try to reference other tasks using the standard Taskfile format.

## Motivation

Users expect Razdfile.yml to support Taskfile.dev's standard command syntax, including task references. Currently, when users write:

```yaml
tasks:
  up:
    desc: "Clone repository and set up project"
    cmds:
      - task: install  # This fails with "invalid type: map, expected a string"
```

The parser fails with:
```
Error: Configuration error: Failed to parse Razdfile.yml: tasks.up.cmds[0]: invalid type: map, expected a string
```

This happens because the current `TaskConfig` struct only accepts `Vec<String>` for commands:

```rust
pub struct TaskConfig {
    pub desc: Option<String>,
    pub cmds: Vec<String>,  // Only accepts strings
    #[serde(default)]
    pub internal: bool,
}
```

However, Taskfile.dev supports multiple command types:
1. **Simple string commands**: `- echo "hello"`
2. **Task references**: `- task: other-task`
3. **Complex commands with options**: `- cmd: echo "test"\n  silent: true`

Users naturally expect to compose tasks by referencing other tasks, which is a fundamental Taskfile.dev feature. Without this support, Razdfile.yml becomes less flexible and forces users to duplicate command sequences or use workarounds.

## Proposed Changes

### 1. Flexible Command Representation

Introduce a new `Command` enum that supports both simple strings and task references:

```rust
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(untagged)]
pub enum Command {
    /// Simple string command (e.g., "echo hello")
    String(String),
    /// Task reference with optional parameters
    TaskRef {
        task: String,
        #[serde(skip_serializing_if = "Option::is_none")]
        vars: Option<HashMap<String, String>>,
    },
}
```

Update `TaskConfig` to use the new enum:

```rust
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TaskConfig {
    pub desc: Option<String>,
    pub cmds: Vec<Command>,  // Changed from Vec<String>
    #[serde(default)]
    pub internal: bool,
}
```

### 2. YAML Serialization Format

When converting back to YAML for taskfile execution, serialize commands appropriately:

- **String commands**: `- echo "hello"`
- **Task references**: `- task: install`

Example Razdfile.yml that will now work:

```yaml
version: '3'

tasks:
  up:
    desc: "Clone repository and set up project"
    cmds:
      - task: install
      - task: setup
      
  install:
    desc: "Install development tools"
    cmds:
      - echo "📦 Installing tools..."
      - mise install
      
  setup:
    desc: "Setup project dependencies"
    cmds:
      - echo "⚙️ Setting up..."
      - npm install
```

### 3. Backward Compatibility

The `#[serde(untagged)]` attribute ensures:
- Existing Razdfile.yml files with only string commands continue to work
- No changes required to existing configurations
- Parsing automatically detects string vs. object format

## Implementation Approach

1. **Define Command enum**: Create `Command` type with `String` and `TaskRef` variants
2. **Update TaskConfig**: Change `cmds` field from `Vec<String>` to `Vec<Command>`
3. **Implement Serialize/Deserialize**: Use serde's `#[serde(untagged)]` for automatic parsing
4. **Add tests**: Cover string commands, task references, and mixed command lists
5. **Update error messages**: Provide clear feedback if command format is invalid

## Impact Assessment

### Benefits
- **Standard Taskfile syntax**: Users can use familiar `task:` references
- **Task composition**: Enable building complex workflows from simple tasks
- **Backward compatible**: Existing configs continue to work
- **Better error messages**: Clear guidance when command format is incorrect

### Risks
- **Serialization changes**: YAML output format may differ slightly
- **Test coverage**: Need comprehensive tests for all command variants

### Migration Path
- **No migration needed**: Existing Razdfile.yml files remain valid
- **Optional adoption**: Users can add task references incrementally
- **Documentation update**: Add examples of task reference syntax

## Timeline

- **Implementation**: 1-2 hours
- **Testing**: 1 hour
- **Documentation**: 30 minutes

## Open Questions

1. Should we support additional Taskfile command properties like `silent`, `ignore_error`, `vars`?
   - **Decision**: Start with basic `task:` and `vars:` support, add more properties in future iterations
2. How should we handle invalid task references?
   - **Decision**: Let taskfile handle validation; razd only ensures valid YAML structure
3. Should we validate that referenced tasks exist?
   - **Decision**: No, defer to taskfile for task existence validation during execution
