# Fix Workflow Working Directory in CI/CD

## Why

After moving temporary workflow files to system temp directory (v0.4.5), workflows fail in CI/CD environments because `task` executes commands in the temp directory context instead of the project directory. This causes commands like `npm install`, `npm run build`, etc. to fail because they cannot find project files (`package.json`, source code, etc.).

**Root cause:** When using `task --taskfile /tmp/razd-workflow-build.yml build`, `task` may default to using the taskfile's directory as the working directory, not the project directory where files like `package.json` exist.

**Evidence from CI/CD:**
- ❌ `package.json` not found
- ❌ `download-vscode.js` not found  
- ❌ All project files are inaccessible

## What Changes

- Add `--dir` parameter to `task` command invocation to explicitly set working directory to the project directory
- Ensure all workflow task executions use the correct working directory regardless of where the taskfile is located
- Maintain backward compatibility with existing workflow execution

## Impact

- **Affected specs**: tool-integration (taskfile integration requirement)
- **Affected code**: `src/integrations/taskfile.rs` - `execute_workflow_task_with_mode()` function
- **Benefits**: 
  - Workflows work correctly in CI/CD environments
  - Project files are accessible during task execution
  - Consistent behavior across local and CI/CD environments
  - No breaking changes to user workflows
- **Risks**: None - `--dir` is a standard taskfile parameter that should have been used from the start
