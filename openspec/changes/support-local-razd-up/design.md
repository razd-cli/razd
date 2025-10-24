# Technical Design: Support Local Project Setup with `razd up`

## Overview
This document details the technical implementation for making `razd up` work without a URL argument, enabling setup of already-cloned projects from within their directory.

## Architecture

### Current Flow
```
razd up <url>
  ├─> parse required URL argument
  ├─> git::clone_repository(url, name)
  ├─> env::set_current_dir(cloned_path)
  └─> execute up workflow
```

### New Flow
```
razd up [url]
  ├─> if URL provided:
  │     ├─> git::clone_repository(url, name)
  │     ├─> env::set_current_dir(cloned_path)
  │     └─> execute up workflow
  └─> if URL not provided:
        ├─> validate current directory is a project
        └─> execute up workflow in current directory
```

## Implementation Details

### 1. CLI Argument Parsing

**File**: `src/main.rs`

Change `Commands::Up` enum variant:
```rust
// Current:
Up {
    /// Git repository URL to clone
    url: String,
    /// Directory name (defaults to repository name)
    #[arg(short, long)]
    name: Option<String>,
},

// New:
Up {
    /// Git repository URL to clone (optional for local projects)
    url: Option<String>,
    /// Directory name (defaults to repository name)
    #[arg(short, long)]
    name: Option<String>,
},
```

Update `run()` function to handle optional URL:
```rust
Commands::Up { url, name } => {
    commands::up::execute(url.as_deref(), name.as_deref()).await?;
}
```

### 2. Up Command Logic

**File**: `src/commands/up.rs`

Restructure `execute()` function:

```rust
pub async fn execute(url: Option<&str>, name: Option<&str>) -> Result<()> {
    if let Some(url_str) = url {
        // Clone mode: existing behavior
        execute_with_clone(url_str, name).await
    } else {
        // Local mode: new behavior
        execute_local_project().await
    }
}

async fn execute_with_clone(url: &str, name: Option<&str>) -> Result<()> {
    output::info(&format!("Setting up project from {}", url));
    
    // Step 1: Clone the repository
    let repo_path = git::clone_repository(url, name).await?;
    
    // Step 2: Change to the repository directory
    let absolute_repo_path = env::current_dir()?.join(&repo_path);
    env::set_current_dir(&absolute_repo_path)?;
    output::info(&format!("Working in directory: {}", absolute_repo_path.display()));
    
    // Step 3: Execute up workflow
    execute_up_workflow().await?;
    
    show_success_message()?;
    Ok(())
}

async fn execute_local_project() -> Result<()> {
    output::info("Setting up local project...");
    
    // Step 1: Validate we're in a project directory
    let current_dir = env::current_dir()?;
    validate_project_directory(&current_dir)?;
    
    output::info(&format!("Working in directory: {}", current_dir.display()));
    
    // Step 2: Execute up workflow
    execute_up_workflow().await?;
    
    show_success_message()?;
    Ok(())
}

async fn execute_up_workflow() -> Result<()> {
    if let Some(workflow_content) = get_workflow_config("up")? {
        output::step("Executing up workflow...");
        taskfile::execute_workflow_task("up", &workflow_content).await?;
    } else {
        output::warning("No up workflow found, falling back to legacy setup");
        let current_dir = env::current_dir()?;
        mise::install_tools(&current_dir).await?;
        taskfile::setup_project(&current_dir).await?;
    }
    Ok(())
}

fn validate_project_directory(dir: &Path) -> Result<()> {
    // Check for at least one project indicator file
    let has_razdfile = dir.join("Razdfile.yml").exists();
    let has_taskfile = dir.join("Taskfile.yml").exists();
    let has_mise = dir.join("mise.toml").exists() || dir.join(".mise.toml").exists();
    
    if !has_razdfile && !has_taskfile && !has_mise {
        return Err(RazdError::command(
            "No project detected in current directory. Expected one of: Razdfile.yml, Taskfile.yml, or mise.toml\n\
             Hint: Run 'razd up <url>' to clone a repository, or 'razd init' to initialize a new project."
        ));
    }
    
    Ok(())
}

fn show_success_message() -> Result<()> {
    output::success("Project setup completed successfully!");
    output::info("Next steps:");
    output::info("  razd dev            # Start development workflow");
    output::info("  razd build          # Build project");
    output::info("  razd task <name>    # Run specific task");
    Ok(())
}
```

### 3. Project Detection Logic

The `validate_project_directory()` function checks for project indicator files:

**Priority order:**
1. `Razdfile.yml` - Primary razd configuration
2. `Taskfile.yml` - Task runner configuration
3. `mise.toml` or `.mise.toml` - Tool version management

**Rationale:**
- At least one of these files indicates an existing project that razd can work with
- All three tools are already integrated into razd workflows
- Checking for these files is lightweight and reliable

**Error handling:**
- If none found: clear error message with hints
- Suggests either cloning with URL or running `razd init`

### 4. Help Text Updates

**File**: `src/main.rs`

Update command description:
```rust
/// Clone repository and set up project, or set up local project
Up {
    /// Git repository URL to clone (optional for local projects)
    url: Option<String>,
    // ...
}
```

## Edge Cases and Error Handling

### Case 1: No URL and No Project Files
**Scenario**: User runs `razd up` in empty or non-project directory
**Handling**: Return error with clear guidance
```
Error: No project detected in current directory. Expected one of: Razdfile.yml, Taskfile.yml, or mise.toml
Hint: Run 'razd up <url>' to clone a repository, or 'razd init' to initialize a new project.
```

### Case 2: URL Provided but Name Conflict with `--name`
**Scenario**: User provides URL with `--name` flag
**Handling**: Existing behavior - `--name` only applies when URL is provided
**Note**: When no URL, `--name` flag is ignored (no cloning happens)

### Case 3: Git Repository Detection
**Scenario**: Directory has `.git` but no project files
**Handling**: Treat as "no project detected" - presence of `.git` alone isn't enough
**Rationale**: User might be in wrong subdirectory or incomplete clone

### Case 4: Partial Project Setup
**Scenario**: Project has some but not all expected files (e.g., Taskfile.yml but no mise.toml)
**Handling**: Still valid - workflow will handle missing tools gracefully
**Rationale**: Not all projects use all tools (mise is optional)

## Testing Strategy

### Unit Tests
- Test `validate_project_directory()` with various file combinations
- Test argument parsing with and without URL

### Integration Tests
- Test `razd up <url>` (existing behavior)
- Test `razd up` in project directory (new behavior)
- Test `razd up` in non-project directory (error case)
- Test with various project file combinations

### Manual Testing
- Real project setup scenarios
- Cross-platform validation (Windows/Linux/macOS)
- Various shell environments

## Performance Considerations
- File existence checks are O(1) operations
- No performance impact on existing clone-based workflow
- Local mode is actually faster (skips git clone)

## Security Considerations
- No security implications - only reads project indicator files
- No new file permissions or execution risks
- Maintains existing security model

## Backward Compatibility
- ✅ Existing `razd up <url>` behavior unchanged
- ✅ All existing scripts and documentation continue to work
- ✅ Optional argument is additive, not breaking

## Migration Path
- No migration needed - change is backward compatible
- Users can adopt new behavior organically
- Documentation update recommended but not required

## Future Enhancements
- Potential to add `--force` flag to skip project validation
- Could detect git remotes and suggest clone commands
- Might add interactive mode to guide users through setup options
