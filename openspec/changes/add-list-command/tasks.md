# Implementation Tasks: Add List Command

## Task Breakdown

### 1. Create `list` Command Module
- [x] Create `src/commands/list.rs`
- [x] Implement `execute_list()` function
- [x] Parse `Razdfile.yml` and extract tasks
- [x] Format task list output

### 2. Update CLI Argument Parser
- [x] Add `list` subcommand to main CLI
- [x] Add `--list` flag to `run` subcommand
- [x] Add `--list` global flag to main CLI
- [x] Wire up all three variants to call `execute_list()`

### 3. Implement Task Formatting
- [x] Calculate proper column alignment
- [x] Handle tasks without descriptions
- [x] Sort tasks alphabetically
- [x] Add optional filtering for internal tasks

### 4. Error Handling
- [x] Handle missing `Razdfile.yml` gracefully
- [x] Handle empty task list
- [x] Handle YAML parsing errors

### 5. Testing
- [x] Unit tests for task extraction
- [x] Unit tests for formatting
- [x] Integration test for `razd list`
- [x] Integration test for `razd run --list`
- [x] Integration test for `razd --list`

### 6. Documentation
- [x] Update README.md with `list` command
- [x] Add examples to documentation
- [x] Update CHANGELOG.md

## Implementation Notes

### Task Extraction

```rust
pub async fn execute_list() -> Result<()> {
    let razdfile = RazdfileConfig::load()?;
    
    // Extract tasks
    let mut tasks: Vec<(String, String)> = razdfile.tasks
        .iter()
        .filter(|(_, config)| !config.internal.unwrap_or(false))
        .map(|(name, config)| {
            (name.clone(), config.desc.clone().unwrap_or_default())
        })
        .collect();
    
    // Sort alphabetically
    tasks.sort_by(|a, b| a.0.cmp(&b.0));
    
    // Format output
    println!("task: Available tasks for this project:");
    for (name, desc) in tasks {
        println!("* {:<30} {}", format!("{}:", name), desc);
    }
    
    Ok(())
}
```

### CLI Integration

```rust
// In main.rs
match cli.command {
    Commands::List => {
        commands::list::execute_list().await?;
    }
    Commands::Run { task, list } => {
        if list {
            commands::list::execute_list().await?;
        } else {
            commands::run::execute_run(&task).await?;
        }
    }
    // ... other commands
}

// Global flag handling
if cli.list {
    commands::list::execute_list().await?;
    return Ok(());
}
```

## Testing Strategy

1. **Unit Tests**: Test task extraction and formatting logic
2. **Integration Tests**: Test actual CLI behavior
3. **Edge Cases**: Empty lists, missing files, malformed YAML

## Estimated Effort

- Implementation: 2-3 hours
- Testing: 1-2 hours
- Documentation: 30 minutes
- **Total**: 4-6 hours
