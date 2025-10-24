# Design: Task Auto-Installation via Mise

## Overview

This design document outlines the architectural approach for automatically installing the `task` tool via mise during `razd up` execution, ensuring reliable taskfile operations without manual dependency management.

## Architecture

### Component Interaction

```
razd up
    ↓
execute_up_workflow()
    ↓
taskfile operations
    ↓
ensure_task_available() ← NEW
    ↓
[task available?] → YES → continue
    ↓ NO
mise::install_specific_tool("task", "latest") ← NEW
    ↓
verify_tool_installation() ← NEW
    ↓ 
[verified?] → YES → continue
    ↓ NO
error with guidance
```

### Key Design Decisions

#### 1. Integration Point Selection
**Decision**: Enhance taskfile integration functions rather than modifying up command directly
**Rationale**: 
- Keeps tool installation logic close to where tools are used
- Ensures consistency across all taskfile operations (not just `razd up`)
- Maintains separation of concerns between command orchestration and tool management

#### 2. Tool Installation Strategy
**Decision**: Use mise for automatic tool installation
**Rationale**:
- Leverages existing mise integration in razd
- Provides consistent tool versions across team members
- Aligns with project's tool management philosophy
- Works cross-platform (Windows/Unix)

#### 3. Installation Timing
**Decision**: Install tools lazily (just-in-time) rather than eagerly
**Rationale**:
- Minimizes overhead when tools are already available
- Avoids unnecessary installations for projects that don't use taskfiles
- Provides faster feedback for successful cases

#### 4. Error Handling Strategy
**Decision**: Fail fast with clear guidance rather than silent fallbacks
**Rationale**:
- Users get clear understanding of what went wrong
- Maintains explicit dependency management expectations
- Avoids hard-to-debug hidden state issues

## Implementation Details

### New Functions

#### `mise::install_specific_tool(tool: &str, version: &str)`
```rust
pub async fn install_specific_tool(tool: &str, version: &str, working_dir: &Path) -> Result<()> {
    // Check if mise is available
    if !process::check_command_available("mise").await {
        return Err(RazdError::missing_tool("mise", "https://mise.jdx.dev/getting-started.html"));
    }

    output::step(&format!("Installing {} via mise...", tool));
    
    let args = vec!["install", &format!("{}@{}", tool, version)];
    process::execute_command("mise", &args, Some(working_dir)).await
        .map_err(|e| RazdError::mise(format!("Failed to install {}: {}", tool, e)))?;

    output::success(&format!("✓ {} installed successfully", tool));
    Ok(())
}
```

#### `mise::ensure_tool_available(tool: &str, version: &str)`
```rust
pub async fn ensure_tool_available(tool: &str, version: &str, working_dir: &Path) -> Result<()> {
    // Fast path: check if tool is already available
    if process::check_command_available(tool).await {
        return Ok(());
    }

    // Install tool via mise
    install_specific_tool(tool, version, working_dir).await?;
    
    // Verify installation
    if !process::check_command_available(tool).await {
        return Err(RazdError::tool(format!(
            "Tool '{}' was installed but is not accessible. Please check your PATH configuration.",
            tool
        )));
    }
    
    Ok(())
}
```

#### Enhanced `taskfile::setup_project()`
```rust
pub async fn setup_project(working_dir: &Path) -> Result<()> {
    // Ensure task tool is available
    mise::ensure_tool_available("task", "latest", working_dir).await?;
    
    // Check if Taskfile exists
    if !has_taskfile_config(working_dir) {
        output::warning("No Taskfile found (Taskfile.yml or Taskfile.yaml), skipping project setup");
        return Ok(());
    }

    output::step("Setting up project dependencies with task");
    // ... rest of existing implementation
}
```

### Error Handling Patterns

#### Tool Installation Failure
```rust
// When mise install fails
Err(RazdError::mise(format!(
    "Failed to install task via mise: {}\n\
     Please install task manually: https://taskfile.dev/installation/",
    error_details
)))
```

#### Tool Verification Failure
```rust
// When installed tool is not accessible
Err(RazdError::tool(format!(
    "Task tool was installed but is not accessible.\n\
     This might be a PATH configuration issue.\n\
     Try running: mise reshim\n\
     Or install manually: https://taskfile.dev/installation/"
)))
```

#### Mise Not Available
```rust
// When mise itself is missing
Err(RazdError::missing_tool(
    "mise", 
    "Install mise to enable automatic tool installation: https://mise.jdx.dev/getting-started.html"
))
```

### Performance Considerations

#### Fast Path Optimization
- Use `process::check_command_available()` first before attempting installation
- This avoids expensive mise operations when tools are already present
- Typical case (tool already installed) has minimal overhead

#### Installation Caching
- Rely on mise's built-in caching and tool management
- Don't implement additional caching layers in razd
- Keeps implementation simple and leverages proven tool management

## Cross-Platform Considerations

### Windows Specifics
- PowerShell execution environment
- Windows PATH handling for installed tools  
- File system permissions and paths
- Use tokio::process for consistent async execution

### Unix Specifics  
- bash/zsh execution environment
- Unix PATH handling for installed tools
- File system permissions and executable bits
- Consistent error handling across shells

### Shared Implementation
- Use Rust std library and tokio for cross-platform process execution
- Rely on mise's cross-platform tool installation
- Consistent error message formatting and user guidance

## Testing Strategy

### Unit Tests
- Test tool detection logic (available/not available)
- Test installation success/failure scenarios
- Test verification success/failure scenarios
- Mock process execution for deterministic testing

### Integration Tests
- Test full workflow with real mise installation
- Test cross-platform behavior (Windows/Unix)
- Test various failure scenarios (network issues, permissions, etc.)
- Test existing workflows remain unchanged

### Error Scenario Testing
- No internet connectivity during installation
- Mise not installed or not functional
- Tool installation succeeds but tool not accessible
- Permission issues during tool installation

## Migration and Rollback

### Backward Compatibility
- All existing workflows continue to work unchanged
- Users who already have task installed see no behavior change
- No breaking changes to existing APIs or configurations

### Rollback Strategy
- If automatic installation causes issues, users can:
  1. Install task manually (existing workflow)
  2. Use environments where task is pre-installed
  3. Disable automatic installation (future enhancement)

### Gradual Rollout
- Feature activates automatically when conditions are met (task missing + mise available)
- No configuration changes required for users
- Clear feedback when automatic installation occurs

## Security Considerations

### Tool Installation Security
- Rely on mise's security model for tool installation
- Use latest version rather than pinned versions for security updates
- No credential storage or network authentication in razd

### Process Execution Security
- Use tokio::process for secure process execution
- Validate tool availability after installation
- Provide clear audit trail of installed tools via output messages

## Future Enhancements

### Configurable Tool Versions
- Allow projects to specify required task versions in Razdfile.yml
- Support version constraints and compatibility checking

### Additional Tool Support  
- Extend pattern to other tools (git, etc.) if needed
- Generalize tool installation framework

### Installation Policies
- Configuration option to disable automatic installation
- Support for organization-specific tool installation policies