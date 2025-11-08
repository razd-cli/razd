# Design: Enhanced List JSON Output

## Current Implementation

The `list.rs` command currently serializes minimal task information:

```rust
#[derive(Serialize)]
struct TaskInfo {
    name: String,
    desc: String,
    internal: bool,
}

#[derive(Serialize)]
struct TaskListOutput {
    tasks: Vec<TaskInfo>,
}
```

**Issues:**
- No source location information
- Missing taskfile-compatible fields
- No root-level metadata

## Proposed Architecture

### Enhanced Data Structures

```rust
#[derive(Serialize)]
struct TaskInfo {
    name: String,
    task: String,              // NEW: duplicate of name (taskfile convention)
    desc: String,
    summary: String,           // NEW: extended description (empty for now)
    aliases: Vec<String>,      // NEW: alternative names (empty for now)
    #[serde(skip_serializing_if = "Option::is_none")]
    up_to_date: Option<bool>,  // NEW: runtime status (None for now)
    location: TaskLocation,    // NEW: source position
    #[serde(skip_serializing_if = "is_false")]
    internal: bool,            // EXISTING: razd-specific field
}

#[derive(Serialize)]
struct TaskLocation {
    taskfile: String,   // Absolute path to Razdfile.yml
    line: usize,        // Line number in file (1-indexed)
    column: usize,      // Column number (1-indexed)
}

#[derive(Serialize)]
struct TaskListOutput {
    tasks: Vec<TaskInfo>,
    location: String,   // NEW: path to Razdfile.yml
}

// Helper to skip serializing false booleans
fn is_false(b: &bool) -> bool {
    !b
}
```

## Implementation Approach

### Option 1: YAML Position Tracking (Recommended)

Use `serde_yaml::Value` with manual position tracking:

```rust
pub async fn execute(list_all: bool, json: bool) -> Result<()> {
    let razdfile_path = find_razdfile_path()?;
    
    if json {
        // Parse YAML with position tracking
        let yaml_content = fs::read_to_string(&razdfile_path)?;
        let yaml_value: serde_yaml::Value = serde_yaml::from_str(&yaml_content)?;
        
        // Extract tasks with positions
        let tasks = extract_tasks_with_positions(&yaml_value, &razdfile_path, list_all)?;
        
        let output = TaskListOutput {
            tasks,
            location: razdfile_path.to_string_lossy().to_string(),
        };
        
        println!("{}", serde_json::to_string_pretty(&output)?);
    } else {
        // Existing text output (unchanged)
        let razdfile = RazdfileConfig::load()?;
        // ... existing logic
    }
}

fn extract_tasks_with_positions(
    yaml: &serde_yaml::Value,
    razdfile_path: &Path,
    list_all: bool,
) -> Result<Vec<TaskInfo>> {
    // Navigate to tasks: section
    let tasks_map = yaml["tasks"].as_mapping()
        .ok_or_else(|| RazdError::config("No tasks found"))?;
    
    let mut result = Vec::new();
    
    for (task_name, task_config) in tasks_map {
        let name = task_name.as_str().unwrap_or("").to_string();
        let desc = task_config["desc"].as_str().unwrap_or("").to_string();
        let internal = task_config["internal"].as_bool().unwrap_or(false);
        
        // Filter internal tasks
        if !list_all && internal {
            continue;
        }
        
        // Extract position from YAML (approximate - serde_yaml doesn't preserve exact positions)
        // Alternative: parse with yaml-rust or scan file manually
        let line = estimate_task_line(&name, razdfile_path)?;
        
        result.push(TaskInfo {
            name: name.clone(),
            task: name,
            desc,
            summary: String::new(),
            aliases: Vec::new(),
            up_to_date: None,
            location: TaskLocation {
                taskfile: razdfile_path.to_string_lossy().to_string(),
                line,
                column: 3, // Tasks are typically indented 2 spaces
            },
            internal,
        });
    }
    
    Ok(result)
}

fn estimate_task_line(task_name: &str, razdfile_path: &Path) -> Result<usize> {
    let content = fs::read_to_string(razdfile_path)?;
    
    // Find line containing "  task_name:"
    for (idx, line) in content.lines().enumerate() {
        let trimmed = line.trim_start();
        if trimmed.starts_with(&format!("{}:", task_name)) {
            return Ok(idx + 1); // 1-indexed
        }
    }
    
    Ok(1) // Fallback
}
```

**Pros:**
- Leverages existing YAML parsing
- Simple line-number estimation via string search
- No new dependencies

**Cons:**
- Approximate positions (not exact column tracking)
- Duplicate parsing in JSON mode

### Option 2: yaml-rust with Full Position Tracking

Use `yaml-rust` crate that preserves source positions:

```rust
// In Cargo.toml
[dependencies]
yaml-rust = "0.4"

// In list.rs
use yaml_rust::{YamlLoader, scanner::Marker};

fn extract_exact_positions(yaml_content: &str) -> HashMap<String, (usize, usize)> {
    // yaml-rust provides Marker with line/column info
    // More complex but provides exact positions
}
```

**Pros:**
- Exact line and column numbers
- Professional-grade position tracking

**Cons:**
- Additional dependency
- More complex parsing logic
- Duplicate YAML parsing library

### Option 3: Post-parse File Scanning (Chosen)

Parse with serde_yaml, then scan file for exact positions:

**Pros:**
- No new dependencies
- Exact line numbers
- Reasonable column estimates

**Cons:**
- String search overhead
- Potentially fragile with complex YAML

## Decision: Option 1 (Estimated Positions)

Start with Option 1 for simplicity. Exact positions are nice-to-have but not critical for initial implementation. Can upgrade to Option 2 if users need precise positions.

## Edge Cases

### Multiple tasks with same name
Not possible - YAML keys are unique

### Tasks in included files
Out of scope - only main Razdfile.yml supported

### Comments in YAML
Line numbers may shift - acceptable tradeoff

### Very large Razdfiles
String scanning is O(n) - acceptable for typical files (<1000 lines)

## Testing Strategy

```rust
#[test]
fn test_json_output_with_location() {
    let json = run_razd(&["list", "--json"]);
    let parsed: TaskListOutput = serde_json::from_str(&json).unwrap();
    
    assert!(parsed.location.ends_with("Razdfile.yml"));
    assert_eq!(parsed.tasks[0].name, parsed.tasks[0].task);
    assert_eq!(parsed.tasks[0].summary, "");
    assert_eq!(parsed.tasks[0].aliases.len(), 0);
    assert!(parsed.tasks[0].location.line > 0);
}
```

## Migration Path

**Backward Compatibility:**
- Existing consumers see new fields (additive change)
- `internal` field retained for razd-specific workflows
- Text output unchanged

**Future Extensions:**
- Add `summary` to Razdfile schema → populate field
- Add alias support → populate `aliases` array
- Implement status tracking → populate `up_to_date`
