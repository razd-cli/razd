# Design: Interactive Yes Flag

## Overview

Add a global `-y, --yes` flag to automatically approve all interactive prompts, enabling unattended execution for automation workflows.

## Architecture

### Option 1: Environment Variable Pattern (RECOMMENDED)

**Approach:** Follow existing `RAZD_NO_SYNC` pattern using environment variable for cross-module communication.

**Flow:**
```
CLI args → Set RAZD_AUTO_YES env var → Commands read env var → Auto-approve prompts
```

**Pros:**
- Consistent with existing `--no-sync` flag implementation
- Simple to implement across all modules
- No function signature changes needed
- Easy to test (set env var in tests)

**Cons:**
- Global state via environment variable
- Not pure functional approach

**Implementation:**
1. Add `yes: bool` to `Cli` struct with `#[arg(short = 'y', long, global = true)]`
2. Set `std::env::set_var("RAZD_AUTO_YES", "1"|"0")` in main.rs
3. Update `SyncConfig` initialization to read env var
4. Update `prompt_yes_no()` to check env var
5. Update conflict resolution to auto-select Option 1

### Option 2: Pass Flag Through Function Parameters

**Approach:** Add `auto_approve: bool` parameter to all functions that prompt users.

**Pros:**
- Explicit dependencies
- No global state
- Pure functional approach

**Cons:**
- Changes many function signatures
- Cascading changes through call stack
- More complex refactoring
- Breaks API for any external callers

### Option 3: Context Object Pattern

**Approach:** Create a `Context` struct containing all global flags, pass through all commands.

**Pros:**
- Centralized configuration
- Easier to add more flags later
- Clear dependency injection

**Cons:**
- Major refactoring required
- Changes many existing signatures
- Overkill for single flag addition
- Future work, not immediate need

## Decision

**Choose Option 1: Environment Variable Pattern**

**Rationale:**
- Matches existing architecture (`RAZD_NO_SYNC` precedent)
- Minimal code changes required
- Maintains backward compatibility
- Simple to understand and maintain
- Works well with current codebase structure

## Implementation Details

### CLI Layer (main.rs)
```rust
#[derive(Parser)]
struct Cli {
    /// Automatically answer "yes" to all prompts
    #[arg(short = 'y', long, global = true)]
    yes: bool,
    // ... existing fields
}

fn main() {
    // Set environment variable based on flag
    std::env::set_var("RAZD_AUTO_YES", if cli.yes { "1" } else { "0" });
}
```

### Config Layer (config/mod.rs)
```rust
pub fn check_and_sync_mise(project_dir: &Path) -> Result<()> {
    let auto_yes = env::var("RAZD_AUTO_YES").unwrap_or_default() == "1";
    
    let config = SyncConfig {
        no_sync,
        auto_approve: auto_yes,  // Use flag value
        create_backups: true,
    };
    // ...
}
```

### Sync Layer (config/mise_sync.rs)
```rust
fn handle_sync_conflict(&self) -> Result<SyncResult> {
    if self.config.auto_approve {
        // Auto-select Option 1 (Razdfile priority)
        println!("Auto-approved: Using Razdfile.yml (overwriting mise.toml)");
        return self.sync_razdfile_to_mise();
    }
    // ... existing prompt logic
}
```

### Command Layer (commands/up.rs)
```rust
fn prompt_yes_no(message: &str, default: bool) -> Result<bool> {
    // Check auto-approve first
    let auto_yes = env::var("RAZD_AUTO_YES").unwrap_or_default() == "1";
    if auto_yes {
        return Ok(true);
    }
    // ... existing prompt logic
}
```

## Testing Strategy

1. **Unit tests**: Mock environment variable in tests
2. **Integration tests**: Test full commands with `--yes` flag
3. **Scenario coverage**:
   - Up command creates Razdfile without prompt
   - Mise sync resolves conflicts automatically
   - Short form `-y` works identically
   - No flag maintains interactive behavior

## Backward Compatibility

- Default behavior (no flag): **unchanged** - all prompts work as before
- Environment variable not set: defaults to interactive mode
- No breaking changes to existing commands
- Help text updated to document new flag

## Future Considerations

If more global flags are added later, consider refactoring to Option 3 (Context Object Pattern). For now, environment variable pattern is sufficient and consistent with project conventions.
