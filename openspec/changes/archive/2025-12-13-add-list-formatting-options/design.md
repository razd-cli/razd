# Design: Add List Formatting Options

## Current Architecture

The current `razd list` command in `src/commands/list.rs` has a simple flow:

```rust
pub async fn execute() -> Result<()> {
    // 1. Load Razdfile.yml
    let razdfile = RazdfileConfig::load()?;
    
    // 2. Filter: only non-internal tasks
    let tasks: Vec<_> = razdfile.tasks
        .iter()
        .filter(|(_, config)| !config.internal)
        .collect();
    
    // 3. Sort alphabetically
    tasks.sort_by(|a, b| a.0.cmp(&b.0));
    
    // 4. Display in text format with alignment
    println!("task: Available tasks for this project:");
    for (name, task) in tasks {
        println!("* {}: {}", name, task.desc);
    }
}
```

**Limitations:**
- Hardcoded to filter out internal tasks
- Single text output format
- No parameterization for different output modes

## Proposed Architecture

### Updated Function Signature

```rust
pub async fn execute(list_all: bool, json: bool) -> Result<()> {
    // Load Razdfile.yml
    let razdfile = RazdfileConfig::load()?;
    
    // Filter based on list_all flag
    let tasks = if list_all {
        // Show all tasks
        razdfile.tasks.iter().collect()
    } else {
        // Show only non-internal tasks (current behavior)
        razdfile.tasks.iter()
            .filter(|(_, config)| !config.internal)
            .collect()
    };
    
    // Output based on json flag
    if json {
        output_json(&tasks)?;
    } else {
        output_text(&tasks);
    }
    
    Ok(())
}
```

### JSON Output Structure

```rust
use serde::{Serialize, Deserialize};

#[derive(Serialize, Deserialize)]
struct TaskListOutput {
    tasks: Vec<TaskInfo>,
}

#[derive(Serialize, Deserialize)]
struct TaskInfo {
    name: String,
    desc: String,
    #[serde(skip_serializing_if = "std::ops::Not::not")]
    internal: bool,
}

fn output_json(tasks: &[(&String, &TaskConfig)]) -> Result<()> {
    let task_infos: Vec<TaskInfo> = tasks
        .iter()
        .map(|(name, config)| TaskInfo {
            name: (*name).clone(),
            desc: config.desc.clone().unwrap_or_default(),
            internal: config.internal,
        })
        .collect();
    
    let output = TaskListOutput { tasks: task_infos };
    let json = serde_json::to_string_pretty(&output)?;
    println!("{}", json);
    Ok(())
}
```

### CLI Integration

Update `src/main.rs`:

```rust
#[derive(Subcommand)]
enum Commands {
    /// List all available tasks from Razdfile.yml
    List {
        /// List all tasks, including internal ones
        #[arg(long)]
        list_all: bool,
        
        /// Output in JSON format
        #[arg(long)]
        json: bool,
    },
    // ... other commands
}

// In match statement:
Some(Commands::List { list_all, json }) => {
    commands::list::execute(list_all, json).await?;
}
```

### Global `--list` Flag Updates

The global `--list` flag currently calls `list::execute()`. Update to pass default values:

```rust
// Handle global --list flag
if cli.list {
    return commands::list::execute(false, false).await;
}
```

## Implementation Details

### Text Output (Default)

Current behavior preserved - no changes needed to text formatting:

```
task: Available tasks for this project:
* build:    Build project
* test:     Run tests
```

When `--list-all` is added:

```
task: Available tasks for this project:
* build:            Build project
* test:             Run tests
* internal-setup:   Internal setup task (auto-detected)
```

Note: We could add `(auto-detected)` suffix or similar indicator for internal tasks in text mode.

### JSON Output Format

Following taskfile's JSON structure (simplified for razd):

```json
{
  "tasks": [
    {
      "name": "build",
      "desc": "Build project",
      "internal": false
    },
    {
      "name": "test",
      "desc": "Run tests",
      "internal": false
    }
  ]
}
```

When task has no description:

```json
{
  "name": "deploy",
  "desc": "",
  "internal": false
}
```

### Error Handling

Maintain current error behavior:
- Missing Razdfile.yml → Error message
- No tasks → "No tasks found" (in text) or empty array (in JSON)

JSON error output for consistency:

```json
{
  "error": "Razdfile.yml not found in current directory"
}
```

## Trade-offs and Considerations

### Pros
- ✅ Matches taskfile behavior (familiar to users)
- ✅ Enables automation and scripting
- ✅ Maintains backward compatibility
- ✅ Simple implementation (minimal code changes)

### Cons
- ⚠️ Adds two new flags to maintain
- ⚠️ JSON schema needs to stay stable for scripts
- ⚠️ Text output with internal tasks might be cluttered

### Alternative Approaches Considered

#### 1. Separate commands instead of flags
```bash
razd list-all
razd list-json
```
**Rejected:** Creates command sprawl, flags are more standard

#### 2. Single format flag
```bash
razd list --format=json
razd list --format=all
```
**Rejected:** `--list-all` and `--json` are taskfile standards

#### 3. Delegate to task directly
```bash
razd list → runs task --list internally
```
**Rejected:** Loses razd's filtering of internal tasks, requires Taskfile.yml

## Testing Strategy

### Unit Tests

```rust
#[tokio::test]
async fn test_list_all_includes_internal_tasks() {
    // Create Razdfile with internal task
    let config = create_test_config_with_internal_task();
    
    // Execute with list_all = true
    let output = capture_output(|| {
        execute(true, false).await
    });
    
    // Verify internal task appears
    assert!(output.contains("internal-setup"));
}

#[tokio::test]
async fn test_json_output_valid() {
    let config = create_test_config();
    
    let output = capture_output(|| {
        execute(false, true).await
    });
    
    // Verify valid JSON
    let parsed: TaskListOutput = serde_json::from_str(&output)?;
    assert_eq!(parsed.tasks.len(), 2);
}
```

### Integration Tests

Test with real Razdfile.yml in temp directory:

```rust
#[tokio::test]
async fn test_list_all_json_combined() {
    let temp_dir = create_temp_razdfile_with_internal();
    
    // Run razd list --list-all --json
    let output = run_razd_command(&["list", "--list-all", "--json"]);
    
    let json: TaskListOutput = serde_json::from_str(&output)?;
    
    // Verify all tasks present
    assert!(json.tasks.iter().any(|t| t.internal));
}
```

## Migration Path

1. **Phase 1:** Add flags, maintain backward compatibility
2. **Phase 2:** Users gradually adopt new flags for their use cases
3. **Future:** Could add more output formats (YAML, CSV) if needed

## Dependencies

- `serde_json` crate (already in Cargo.toml for other serialization)
- No new external dependencies required

This design ensures minimal disruption while adding powerful new capabilities aligned with taskfile's interface.