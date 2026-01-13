# Tasks: CLI Architecture Implementation

## Phase 1: Foundation

- [ ] **1.1 Create package structure**
  - Create `cmd/razd/main.go` with minimal entry point
  - Create `internal/flags/flags.go` with global flag definitions
  - Create `internal/output/logger.go` with colored logging
  - Create `internal/version/version.go` with version info

- [ ] **1.2 Implement flag parsing**
  - Add spf13/pflag dependency
  - Define global flags: `--version`, `--help`, `-v/--verbose`, `-s/--silent`, `-d/--dir`, `--color`, `-y/--yes`
  - Add file path flags: `-t/--taskfile`, `--razdfile`
  - Add control flags: `--list`, `--no-sync`
  - Implement `flags.Validate()` for mutual exclusion checks

- [ ] **1.3 Implement logger**
  - Create Logger struct with Stdout/Stderr writers
  - Add methods: Infof, Debugf, Warnf, Errf, Successf
  - Add fatih/color dependency for colored output
  - Support `NO_COLOR` and `FORCE_COLOR` environment variables

## Phase 2: Command Infrastructure

- [ ] **2.1 Create CLI orchestrator**
  - Create `internal/cli/cli.go` with CLI struct
  - Implement `Run()` method with flag parsing and dispatch
  - Handle `--version` and `--help` early exits
  - Handle global `--list` flag

- [ ] **2.2 Create command registry**
  - Create `internal/cli/commands.go` with Command struct
  - Define command map with name → handler mapping
  - Support command aliases (e.g., `list` → `ls`)

- [ ] **2.3 Implement error handling**
  - Create `internal/errors/errors.go` with exit codes
  - Define RazdError interface with Code() method
  - Create specific error types: TaskRunError, ConfigError, TrustError

## Phase 3: Trust System

- [ ] **3.1 Implement trust store**
  - Create `internal/trust/store.go` with TrustStore struct
  - Store trust data in `~/.config/razd/trust.json`
  - Implement `GetStatus()`, `AddTrusted()`, `AddIgnored()`, `Remove()`

- [ ] **3.2 Implement trust verification**
  - Create `internal/trust/check.go` with EnsureTrusted()
  - Prompt user for trust decision on first run
  - Respect `--yes` flag for auto-approve

- [ ] **3.3 Implement trust command**
  - Create `internal/cli/trust.go`
  - Support flags: `--untrust`, `--show`, `--all`, `--ignore`
  - Auto-run `mise trust` when trusting project with mise provisioner
  - Auto-run `mise trust --untrust` when untrusting

## Phase 4: Core Commands

- [ ] **4.1 Implement `razd run <task>`**
  - Parse task name from args
  - Check trust before execution
  - Use razdfile.Reader to find Razdfile
  - Create go-task executor and run task
  - Pass CLI_ARGS after `--` to task

- [ ] **4.2 Implement `razd up`**
  - Support `razd up` — setup current project (install tools)
  - Support `razd up <url>` (git clone + setup)
  - Support short GitHub syntax: `razd up gh:user/repo`
  - Steps: trust check → mise install / devbox install
  - Check trust before execution
  - Handle `--yes` / `-y` for non-interactive mode
  - Handle `--run` / `-r` to run default task after install

- [ ] **4.3 Implement `razd list`**
  - Read Razdfile and list all tasks
  - Show task descriptions
  - Support `--json` output format
  - Support `--all` to show internal tasks
  - Include task location (file, line, column) in JSON

## Phase 5: Project Commands

- [ ] **5.1 Implement `razd init`**
  - Create new Razdfile.yml in current directory
  - Support `--using mise|devbox` flag
  - Interactive mode: prompt for provisioner if not specified
  - Detect existing mise.toml/devbox.json and offer migration
  - Support `--force` to overwrite existing file
  - Support `--migrate` for automatic migration

- [ ] **5.2 Implement `razd add`**
  - Parse `tool@version` arguments
  - Read existing Razdfile.yml
  - Append to `dependencies.ensure` list (avoid duplicates)
  - Save updated Razdfile.yml preserving formatting
  - Validate dependency format before adding

- [ ] **5.3 Implement `razd shell`**
  - Check trust before execution
  - Detect provisioner from Razdfile
  - If mise: run `mise shell`
  - If devbox: run `devbox shell`
  - Fallback: run $SHELL with configured PATH

- [ ] **5.4 Implement `razd dev`**
  - Run "dev" task if exists
  - Check trust before execution
  - Otherwise, show helpful error

- [ ] **5.5 Implement `razd build`**
  - Run "build" task if exists
  - Check trust before execution
  - Otherwise, show helpful error

## Phase 6: Polish

- [ ] **6.1 Implement shell completion**
  - Add `--completion <shell>` flag
  - Generate bash completion script
  - Generate zsh completion script
  - Generate fish completion script

- [ ] **6.2 Improve help output**
  - Create formatted help with examples
  - Show available commands
  - Show common flags
  - Support `razd help <command>`

- [ ] **6.3 Add CI annotations**
  - Detect GitHub Actions, GitLab CI, etc.
  - Emit error annotations for failed tasks
  - Support `--exit-code` for pass-through exit codes

## Phase 7: Testing & Documentation

- [ ] **7.1 Unit tests**
  - Test flag parsing and validation
  - Test logger output formatting
  - Test command registry
  - Test error code mapping
  - Test trust store operations

- [ ] **7.2 Integration tests**
  - Test `razd run` with example project
  - Test `razd list` output (text and JSON)
  - Test `razd init` creates valid file
  - Test `razd add` modifies Razdfile correctly
  - Test `razd up -r` installs and runs default
  - Test `razd trust` commands
  - Test error handling and exit codes

- [ ] **7.3 Update README**
  - Document CLI usage
  - Add command reference
  - Include examples
  - Document trust system

## Dependencies

- Phase 2 depends on Phase 1
- Phase 3 depends on Phase 2
- Phase 4 depends on Phase 3 (trust checks)
- Phase 5 depends on Phase 4
- Phase 6-7 can run in parallel after Phase 5
