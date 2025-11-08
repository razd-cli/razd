# Design: Add --taskfile Flag Support

## Architecture Overview

This change introduces a global CLI flag mechanism for specifying custom configuration file paths, affecting the CLI layer, configuration loading layer, and all commands that depend on Razdfile configuration.

## Implementation Options

### Option 1: Global Flag with Thread-Local Storage (Recommended)

**Approach**: Add global `--taskfile`/`--razdfile` flags to `Cli` struct, pass custom path through function parameters to all commands.

**Pros**:
- Explicit data flow (no hidden state)
- Easy to test and debug
- Type-safe at compile time
- Follows Rust best practices

**Cons**:
- Requires updating all command signatures
- More verbose parameter passing

**Implementation**:
```rust
// main.rs
struct Cli {
    #[arg(short = 't', long, global = true)]
    taskfile: Option<String>,
    
    #[arg(long, global = true)]
    razdfile: Option<String>,
    // ...
}

// Derive final path
fn get_config_path(cli: &Cli) -> Option<PathBuf> {
    cli.razdfile.as_ref()
        .or(cli.taskfile.as_ref())
        .map(PathBuf::from)
}
```

### Option 2: Environment Variable Fallback

**Approach**: Support CLI flag + environment variable `RAZD_TASKFILE`.

**Pros**:
- Useful for CI/CD environments
- Can set per-session default

**Cons**:
- Hidden state (harder to debug)
- Rejected per user requirements (answer #4: Нет)

### Option 3: Configuration Context Object

**Approach**: Create `ConfigContext` struct to carry configuration metadata.

**Pros**:
- Encapsulates all config-related state
- Easier to extend with new metadata

**Cons**:
- More complex abstraction
- Overkill for single field

**Implementation**:
```rust
struct ConfigContext {
    custom_path: Option<PathBuf>,
    // Future: working_dir, env_overrides, etc.
}
```

## Chosen Approach: Option 1 (Global Flag with Explicit Passing)

**Rationale**:
- Aligns with Rust idioms (explicit over implicit)
- Minimal code changes to existing structure
- Clear data flow for debugging
- No hidden global state

## Component Changes

### 1. CLI Layer (`src/main.rs`)

```rust
#[derive(Parser)]
struct Cli {
    /// Specify custom taskfile/razdfile path
    #[arg(short = 't', long, global = true, value_name = "FILE")]
    taskfile: Option<String>,
    
    /// Specify custom razdfile path (overrides --taskfile)
    #[arg(long, global = true, value_name = "FILE")]
    razdfile: Option<String>,
    
    // ... existing fields
}

// Helper to resolve final path
fn resolve_config_path(cli: &Cli) -> Option<PathBuf> {
    cli.razdfile.as_ref()
        .or(cli.taskfile.as_ref())
        .map(|s| PathBuf::from(s))
}
```

### 2. Config Layer (`src/config/razdfile.rs`)

```rust
impl RazdfileConfig {
    /// Load from custom path or default Razdfile.yml
    pub fn load_with_path(custom_path: Option<PathBuf>) -> Result<Option<Self>> {
        let path = match custom_path {
            Some(p) => {
                if !p.exists() {
                    return Err(RazdError::config(format!(
                        "Specified configuration file not found: {}",
                        p.display()
                    )));
                }
                p
            }
            None => {
                let default = env::current_dir()?.join("Razdfile.yml");
                if !default.exists() {
                    return Ok(None);
                }
                default
            }
        };
        
        Self::load_from_path(path)
    }
}
```

### 3. Commands Update

Update all commands to accept `custom_path: Option<PathBuf>`:
- `commands::list::execute(list_all, json, custom_path)`
- `commands::run::execute(task_name, args, custom_path)`
- `commands::setup::execute(custom_path)`
- `commands::up::execute(url, name, init, custom_path)`

### 4. Integration Points

**Taskfile integration** (`src/integrations/taskfile.rs`):
- Already uses `--taskfile` flag when calling external `task` command
- Add custom path to taskfile invocations when specified

**Mise integration** (`src/integrations/mise.rs`):
- Update `has_mise_config()` to check custom path
- Sync logic needs custom path awareness

## Data Flow

```
CLI Parsing (main.rs)
    ↓
Resolve config path (razdfile > taskfile > default)
    ↓
Pass to command::execute(custom_path)
    ↓
RazdfileConfig::load_with_path(custom_path)
    ↓
Use config in command logic
```

## Error Handling

### File Not Found
- **Custom path**: Error immediately with clear message
- **Default path**: Return `None` (backward compatible)

### Invalid YAML
- Same behavior regardless of custom/default path
- Show file path in error message

### Permission Errors
- Show OS error with file path context

## Testing Strategy

### Unit Tests
- Flag parsing (both long and short forms)
- Priority logic (razdfile > taskfile)
- Path resolution (relative, absolute)
- Error cases (file not found, invalid path)

### Integration Tests
- Commands with custom paths
- Default behavior unchanged
- Cross-platform path handling (Windows/Unix)

### Edge Cases
- Empty path string
- Path with spaces
- Path with special characters
- Non-existent parent directory
- Circular symlinks

## Backward Compatibility

**Breaking changes**: None

**Behavior changes**:
- New flags available (opt-in)
- Default behavior unchanged (uses `Razdfile.yml`)
- Existing commands work without modification

## Migration Path

No migration needed - this is an additive change. Users can:
1. Continue using default `Razdfile.yml` (no changes)
2. Start using `--taskfile`/`--razdfile` flags when needed
3. Mix and match on per-command basis

## Future Extensions

This design enables future enhancements:
- Environment variable support (if requirements change)
- Multiple config file merging
- Config file discovery (search parent directories)
- Default config path in `.razd/config`
