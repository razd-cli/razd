# Delete Temporary Workflow Files Immediately After Process Spawn

## Why

Currently, when executing workflow tasks (dev, build, install, default), razd creates temporary files in the system temp directory and deletes them only after the entire task execution completes. This means the temporary file exists on disk for the full duration of the workflow, which could be minutes or even hours for long-running tasks like development servers.

The comment in the code suggests the intended behavior was different:
```rust
// Create temporary taskfile in project directory for task to load.
// File is deleted immediately after task process starts (loads the file into memory).
```

Since `task` (taskfile.dev) loads the entire Taskfile configuration into memory at startup, the temporary file is only needed until the process successfully spawns and reads the file. Keeping the file for the entire execution duration is unnecessary and wastes disk space, especially for long-running tasks or when multiple workflow tasks run simultaneously.

## What Changes

- Modify `execute_workflow_task_with_mode()` in `src/integrations/taskfile.rs` to delete the temporary file immediately after spawning the task process, rather than after the process completes
- Refactor `execute_command()` and `execute_command_interactive()` in `src/integrations/process.rs` to separate process spawning from process waiting:
  - Add new functions `spawn_command()` and `spawn_command_interactive()` that return a process handle
  - Modify existing functions to use the new spawn functions internally
- Update workflow execution logic to: create temp file → spawn process → delete temp file → wait for process completion

## Impact

- **Affected specs**: tool-integration (taskfile workflow execution)
- **Affected code**: 
  - `src/integrations/taskfile.rs` - `execute_workflow_task_with_mode()` function
  - `src/integrations/process.rs` - command execution functions
- **Benefits**:
  - Immediate cleanup of temporary files (seconds vs minutes/hours)
  - Reduced disk space usage during long-running tasks
  - Multiple simultaneous workflows won't accumulate temporary files
  - Implementation matches documented intent
- **Risks**:
  - If `task` doesn't load the file immediately at startup, early deletion could cause failures
  - Race condition if process spawn is slow and file is deleted before task reads it
  - Need to add small delay (e.g., 100ms) after spawn to ensure file is loaded
- **Mitigation**: Add configurable delay after process spawn before file deletion to ensure reliable file loading across different system loads
