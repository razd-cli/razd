# Tasks for Add razd.yml Configuration System

## Core Implementation

- [x] **Create built-in default workflows module**
  - Create `src/defaults.rs` with embedded default workflows
  - Define const DEFAULT_WORKFLOWS with Taskfile v3 format
  - Include standard up/install/dev/build workflows

- [x] **Implement Razdfile.yml parsing**
  - Add serde_yaml dependency to Cargo.toml
  - Create `src/config/razdfile.rs` for parsing Razdfile.yml
  - Add fallback chain: Razdfile.yml → built-in defaults

- [x] **Update command delegation logic**
  - Modify `src/commands/up.rs` to use workflow system
  - Modify `src/commands/install.rs` to use workflow system
  - Add `src/commands/dev.rs` for development workflow
  - Add `src/commands/build.rs` for build workflow
  - Ensure `src/commands/task.rs` delegates directly to `task <anything>`

- [x] **Enhance razd init command**
  - Update `src/commands/init.rs` to support optional config generation
  - Add `--config` flag to create Razdfile.yml
  - Add `--full` flag to create all files (Razdfile.yml, Taskfile.yml, mise.toml)
  - Make default `razd init` work without creating files

## Workflow Execution

- [x] **Implement workflow executor**
  - ~~Create `src/workflow/executor.rs` for running workflows~~ (implemented in config module)
  - Handle both Razdfile.yml tasks and built-in defaults
  - Add clear feedback about which workflow source is being used

- [x] **Add taskfile integration for workflows**
  - Update `src/integrations/taskfile.rs` to support `--taskfile` parameter
  - Add function to execute tasks from custom taskfile
  - Maintain existing direct task delegation for `razd task`

## Configuration Management

- [x] **Create configuration module structure**
  - Create `src/config/mod.rs` as config module entry point
  - Add `src/config/defaults.rs` for default workflow definitions
  - Add `src/config/detection.rs` for project type detection

- [x] **Implement project type detection**
  - Detect common project types (Node.js, Python, Rust, Go, etc.)
  - Generate appropriate default workflows based on project type
  - Create templates for common project configurations

## Testing

- [x] **Add unit tests for configuration parsing**
  - Test Razdfile.yml parsing with various formats
  - Test fallback chain functionality
  - Test built-in defaults execution

- [ ] **Add integration tests for workflow execution**
  - Test `razd up/install/dev/build` with Razdfile.yml present
  - Test `razd up/install/dev/build` with built-in defaults
  - Test `razd task` direct delegation

- [ ] **Add tests for razd init enhancements**
  - Test `razd init` without file creation
  - Test `razd init --config` with Razdfile.yml generation
  - Test `razd init --full` with all file generation

## Documentation

- [x] **Update CLI help text**
  - Update help for `razd init` with new flags
  - Add help for new `razd dev` and `razd build` commands
  - Document workflow vs task delegation concepts

- [ ] **Update README with configuration examples**
  - Add examples of built-in workflows
  - Show Razdfile.yml customization examples
  - Document fallback chain behavior