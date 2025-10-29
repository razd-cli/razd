# Design: Mise Configuration Integration

## Architecture Overview

This design describes the technical implementation of mise configuration management in razd, including the data structures, file synchronization logic, and integration points.

## Component Design

### 1. Data Structures

#### Razdfile Extension
Extend the existing `RazdfileConfig` structure to include mise configuration:

```rust
// src/config/razdfile.rs
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RazdfileConfig {
    pub version: String,
    pub tasks: HashMap<String, TaskConfig>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub mise: Option<MiseConfig>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MiseConfig {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub tools: Option<HashMap<String, ToolConfig>>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub plugins: Option<HashMap<String, String>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(untagged)]
pub enum ToolConfig {
    /// Simple version string: "22"
    Simple(String),
    /// Complex configuration with options
    Complex {
        version: String,
        #[serde(skip_serializing_if = "Option::is_none")]
        postinstall: Option<String>,
        #[serde(skip_serializing_if = "Option::is_none")]
        os: Option<Vec<String>>,
        #[serde(skip_serializing_if = "Option::is_none")]
        install_env: Option<HashMap<String, String>>,
    },
}
```

#### File Tracking Metadata
Store file modification tracking in user data directory:

```rust
// src/config/file_tracker.rs
#[derive(Debug, Serialize, Deserialize)]
pub struct FileTrackingState {
    pub razdfile_modified: SystemTime,
    pub mise_toml_modified: SystemTime,
    pub last_sync_time: SystemTime,
}
```

### 2. File Synchronization Logic

#### Sync Strategy Decision Tree
```
On razd command start:
  ├─ Load tracking metadata
  ├─ Check Razdfile.yml modification time
  ├─ Check mise.toml modification time
  │
  ├─ If Razdfile.yml changed:
  │   ├─ Generate/update mise.toml from Razdfile
  │   ├─ Update tracking metadata
  │   └─ Continue command execution
  │
  ├─ If mise.toml changed (but not Razdfile.yml):
  │   ├─ Show warning: "mise.toml was modified manually"
  │   ├─ Prompt: "Sync changes to Razdfile.yml? [y/N]"
  │   ├─ If yes:
  │   │   ├─ Parse mise.toml
  │   │   ├─ Update Razdfile.yml mise section
  │   │   └─ Update tracking metadata
  │   └─ Continue command execution
  │
  └─ If no changes:
      └─ Continue command execution
```

#### File Storage Location
- **Tracking metadata**: `$RAZD_DATA_DIR/file_tracking/<project_hash>.json`
  - Project hash: SHA256 of absolute project path
  - Prevents conflicts between multiple projects
  - Platform-specific data directories:
    - Windows: `%LOCALAPPDATA%\razd\file_tracking\`
    - Unix: `~/.local/share/razd/file_tracking/`

### 3. TOML Generation

#### Template-based Generation
Use a TOML serialization library with structured generation:

```rust
// src/config/mise_generator.rs
pub fn generate_mise_toml(mise_config: &MiseConfig) -> Result<String> {
    let mut toml_doc = toml_edit::Document::new();
    
    // Add tools section
    if let Some(tools) = &mise_config.tools {
        let mut tools_table = toml_edit::Table::new();
        for (name, config) in tools {
            match config {
                ToolConfig::Simple(version) => {
                    tools_table.insert(name, toml_edit::value(version));
                }
                ToolConfig::Complex { version, postinstall, os, install_env } => {
                    let mut tool_table = toml_edit::InlineTable::new();
                    tool_table.insert("version", version.into());
                    if let Some(cmd) = postinstall {
                        tool_table.insert("postinstall", cmd.into());
                    }
                    if let Some(os_list) = os {
                        tool_table.insert("os", os_list.clone().into());
                    }
                    if let Some(env) = install_env {
                        tool_table.insert("install_env", env.clone().into());
                    }
                    tools_table.insert(name, toml_edit::value(tool_table));
                }
            }
        }
        toml_doc.insert("tools", toml_edit::Item::Table(tools_table));
    }
    
    // Add plugins section
    if let Some(plugins) = &mise_config.plugins {
        let mut plugins_table = toml_edit::Table::new();
        for (name, url) in plugins {
            plugins_table.insert(name, toml_edit::value(url));
        }
        toml_doc.insert("plugins", toml_edit::Item::Table(plugins_table));
    }
    
    Ok(toml_doc.to_string())
}
```

### 4. Integration Points

#### Command Interceptor
Add pre-command hook to check for sync needs:

```rust
// src/main.rs or src/commands/mod.rs
pub async fn run_command_with_sync_check<F, Fut>(
    working_dir: &Path,
    command_fn: F,
) -> Result<()>
where
    F: FnOnce() -> Fut,
    Fut: Future<Output = Result<()>>,
{
    // Check and handle sync before running actual command
    file_sync::check_and_sync_if_needed(working_dir).await?;
    
    // Execute the actual command
    command_fn().await
}
```

#### Backward Compatibility
- If no `mise` section in Razdfile.yml, use existing behavior
- Standalone `mise.toml` continues to work without Razdfile
- No forced migration - opt-in feature

### 5. Error Handling

#### Error Scenarios
1. **Parse errors in Razdfile mise section**: Show clear error with line numbers
2. **TOML generation errors**: Validate before writing, rollback on failure
3. **File permission errors**: Clear message about file access issues
4. **Concurrent modification**: Detect and warn about race conditions
5. **Invalid tool/plugin names**: Validate against mise requirements

#### Recovery Strategy
- Never overwrite files without backup
- Store previous mise.toml as `mise.toml.backup` before sync
- Allow manual resolution of conflicts
- Provide `razd sync --force` flag to override prompts

## Performance Considerations

### File I/O Optimization
- **Lazy loading**: Only check files when necessary
- **Caching**: Cache tracking metadata in memory during command execution
- **Minimal writes**: Only update files when content actually changes
- **Fast path**: Skip all checks if both files haven't changed

### Expected Performance Impact
- File modification check: < 1ms (metadata only)
- TOML generation: < 10ms for typical configs
- User prompt: 0ms if no sync needed
- Total overhead: < 15ms in worst case, ~0ms typical case

## Testing Strategy

### Unit Tests
- TOML generation for all ToolConfig variants
- File tracking metadata serialization
- Sync decision logic for all scenarios
- Error handling for each failure mode

### Integration Tests
- End-to-end sync from Razdfile to mise.toml
- Manual mise.toml edit detection
- Multi-project tracking isolation
- Concurrent razd command handling

### Test Data
Create fixture directories with:
- Razdfile.yml with various mise configs
- Pre-existing mise.toml files
- Modified file scenarios
- Edge cases (empty configs, invalid syntax)

## Security Considerations

- **File path validation**: Prevent directory traversal attacks
- **Sanitize tool names**: Validate against mise naming rules
- **URL validation**: Check plugin URLs before writing to TOML
- **Permission checks**: Respect file system permissions

## Dependencies

### New Crates
- `toml_edit = "0.22"` - For preserving TOML formatting and comments
- `sha2 = "0.10"` - For project path hashing

### Existing Crates
- `serde_yaml` - Already used for Razdfile parsing
- `tokio` - Async file operations
- `std::time::SystemTime` - File modification tracking

## Migration Path

### Phase 1: Foundation (Current Change)
- Implement core sync logic
- Support tools and plugins only
- Add file tracking infrastructure

### Phase 2: Enhanced Features (Future)
- Support additional mise.toml sections (env, settings)
- IDE integration with JSON schema
- Visual diff tool for sync conflicts

### Phase 3: Advanced (Future)
- Automatic migration tool for existing mise.toml
- Team-wide sync strategies
- Configuration templates and presets

## Open Questions

1. **Sync confirmation**: Should first-time sync require explicit confirmation?
   - **Decision**: Yes, show one-time educational message

2. **Conflict resolution**: What if both files changed simultaneously?
   - **Decision**: Show diff, ask user to choose or merge manually

3. **Empty mise section**: Should empty `mise: {}` generate mise.toml?
   - **Decision**: No, only generate if tools or plugins are defined

4. **Comment preservation**: Should we preserve comments in generated mise.toml?
   - **Decision**: No for v1, use toml_edit for future enhancement
