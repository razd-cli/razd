# Move Temporary Workflow Files to System Temp Directory

## Why

When executing workflow tasks (dev, build, install, default), razd creates temporary `.razd-workflow-{task_name}.yml` files in the project directory to pass workflow content to taskfile. These files are immediately cleaned up after execution but appear as untracked changes in Git during the brief window they exist, causing confusion and noise in Git status output.

The current implementation creates these files in the project's working directory, which is inappropriate for temporary files that should not be visible to users or version control systems.

## What Changes

- Modify `src/integrations/taskfile.rs` to create temporary workflow files in the system temp directory (`std::env::temp_dir()`) instead of the project directory
- Ensure proper cleanup of temp files after workflow execution
- Maintain the same workflow execution logic, only changing the location of temporary files

## Impact

- **Affected specs**: tool-integration (taskfile integration requirement)
- **Affected code**: `src/integrations/taskfile.rs` - `execute_workflow_task_with_mode()` function
- **Benefits**: 
  - Temporary files never appear in Git status
  - No need to modify `.gitignore` in every project
  - Follows OS best practices for temporary file management
  - Automatic cleanup by OS temp directory mechanisms
- **Risks**: None - temp directory is standard location for ephemeral files
