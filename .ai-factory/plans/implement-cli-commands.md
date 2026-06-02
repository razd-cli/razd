# Plan: Implement Remaining CLI Commands

**Branch:** `feature/implement-cli-commands`
**Created:** 2026-06-01
**Mode:** Full

## Settings

- **Testing:** Yes — unit and integration tests for each command
- **Logging:** Verbose — detailed DEBUG logs for development
- **Docs:** No — warn-only, no mandatory checkpoint

## Research Context

From `/aif-explore` session:

- All 5 examples parse correctly via `razdfile.Reader`
- `main.go` is a hardcoded prototype — must be replaced with `cmd/razd/main.go` as entry point
- CLI infrastructure (flags, logger, errors, trust, provisioner) is fully implemented
- Only `razd up` command is fully implemented; all others are stubs
- `provisioner.GenerateConfig()` is stub for both mise and devbox
- `internal/cli/cmd_up.go` has reusable patterns: `readRazdfile`, `getProvisioner`, `runTask`, `executeCommand`

## Tasks

### Phase 1: Foundation — Switch Entry Point & Extract Helpers

- [x] **1.1** Switch build target from `main.go` to `cmd/razd/main.go`
  - Update `go.mod` or build scripts so `go build ./cmd/razd` produces the `razd` binary
  - Keep `main.go` as-is temporarily (can be removed later) but ensure `cmd/razd/main.go` is the canonical entry point
  - Verify `go build ./cmd/razd` compiles and `razd --help` works
  - Files: `cmd/razd/main.go`, `go.mod` (if needed)

- [x] **1.2** Extract helper functions from `cmd_up.go` into shared file
  - Create `internal/cli/helpers.go` with:
    - `resolveDir(ctx *Context) (string, error)` — get working dir from ctx or os.Getwd()
    - `readRazdfile(dir string, log *output.Logger) (*ast.Razdfile, error)` — find + parse + validate
    - `getProvisioner(rf *ast.Razdfile, dir string, log *output.Logger) (provisioner.Provisioner, error)` — resolve provisioner from Razdfile
    - `ensureTrusted(dir string, prov provisioner.Provisioner, log *output.Logger) error` — trust check + auto-trust
  - Refactor `cmd_up.go` to use these helpers
  - Add DEBUG logging at each step
  - Files: `internal/cli/helpers.go`, `internal/cli/cmd_up.go`

- [ ] **1.3** Add tests for helpers
  - Test `resolveDir` with empty and non-empty dir
  - Test `readRazdfile` with valid and invalid dirs
  - Test `getProvisioner` with dependencies/mise/devbox Razdfiles
  - Files: `internal/cli/helpers_test.go`

### Phase 2: Core Commands — `run`, `list`, `trust`

- [x] **2.1** Implement `razd run <task>` command
  - Create `internal/cli/cmd_run.go`
  - Reuse task execution logic from `cmd_up.go` (move `runTask`, `executeCommand` to helpers or keep in `cmd_run.go`)
  - Support `razd run` (no args = default task) and `razd run <task>` and `razd run <task1> <task2>`
  - Check trust before execution
  - Use provisioner from Razdfile for wrapping commands
  - Handle task not found with `errors.TaskNotFoundError`
  - Fallback: if no Razdfile found, try go-task directly with `--taskfile`/`--razdfile` flags
  - Add DEBUG logging at each step
  - Files: `internal/cli/cmd_run.go`, `internal/cli/commands.go`

- [x] **2.2** Implement `razd list` command
  - Create `internal/cli/cmd_list.go`
  - Read Razdfile and list all task names with descriptions
  - Support `--json` flag for machine-readable output (task name, description, deps, cmds)
  - Support `--all` flag to show internal tasks (e.g., tasks starting with `_`)
  - Show provisioner info if available (mise/devbox + tools)
  - Add DEBUG logging
  - Files: `internal/cli/cmd_list.go`, `internal/cli/commands.go`, `internal/flags/flags.go`

- [x] **2.3** Implement `razd trust` command
  - Create `internal/cli/cmd_trust.go`
  - Support flags: `--untrust`, `--show`, `--all`, `--ignore`
  - Default: trust current project + auto-run `mise trust` if using mise
  - `--untrust`: remove trust + `mise trust --untrust`
  - `--show`: display trust status for current project
  - `--all`: list all trusted projects
  - `--ignore`: add to ignore list (never prompt)
  - Use `internal/trust` package (already fully implemented)
  - Add DEBUG logging
  - Files: `internal/cli/cmd_trust.go`, `internal/cli/commands.go`

- [ ] **2.4** Add tests for Phase 2 commands
  - Test `cmd_run.go`: task execution, task not found, trust check, default task
  - Test `cmd_list.go`: text output, JSON output, --all flag
  - Test `cmd_trust.go`: trust, untrust, show, all, ignore
  - Use temp directories with sample Razdfiles
  - Files: `internal/cli/cmd_run_test.go`, `internal/cli/cmd_list_test.go`, `internal/cli/cmd_trust_test.go`

### Phase 3: Project Commands — `init`, `add`, `shell`

- [x] **3.1** Implement `razd init` command
  - Create `internal/cli/cmd_init.go`
  - Create Razdfile.yml in current directory with minimal template
  - Support `--using <mise|devbox>` flag (required or interactive prompt)
  - Support `--force` to overwrite existing file
  - Support `--migrate` flag: detect existing mise.toml/devbox.json and offer to migrate
  - Detect existing config files: if mise.toml exists, default to mise; if devbox.json exists, default to devbox
  - Generated template should include version, dependencies section, and default tasks
  - Add DEBUG logging
  - Files: `internal/cli/cmd_init.go`, `internal/cli/commands.go`, `internal/flags/flags.go`

- [x] **3.2** Implement `razd add <tool@version>` command
  - Create `internal/cli/cmd_add.go`
  - Parse `tool@version` arguments using `ast.ParseDependencyString()`
  - Read existing Razdfile, append to `dependencies.ensure` (avoid duplicates)
  - Write updated Razdfile preserving YAML structure (use YAML marshalling)
  - Validate that Razdfile has `dependencies` section before adding
  - Support adding multiple deps at once: `razd add node@22 go@1.22`
  - Add DEBUG logging
  - Files: `internal/cli/cmd_add.go`, `internal/cli/commands.go`

- [x] **3.3** Implement `razd shell` command
  - Create `internal/cli/cmd_shell.go`
  - Read Razdfile, detect provisioner
  - If mise: run `mise shell`
  - If devbox: run `devbox shell`
  - Check trust before execution
  - Fallback: if no provisioner, warn and exit with error
  - Add DEBUG logging
  - Files: `internal/cli/cmd_shell.go`, `internal/cli/commands.go`

- [ ] **3.4** Add tests for Phase 3 commands
  - Test `cmd_init.go`: create file, --using flag, --force, migrate detection
  - Test `cmd_add.go`: add single dep, add multiple deps, duplicate prevention, validation errors
  - Test `cmd_shell.go`: mise shell, devbox shell, no provisioner error
  - Files: `internal/cli/cmd_init_test.go`, `internal/cli/cmd_add_test.go`, `internal/cli/cmd_shell_test.go`

### Phase 4: Wire `dev` and `build` commands + Root `main.go`

- [x] **4.1** Finalize `razd dev` and `razd build` commands
  - These already delegate to `RunCommand` in `commands.go`
  - Once `RunCommand` is implemented (Task 2.1), `dev` and `build` will work automatically
  - Verify they work: `razd dev` should run the "dev" task, `razd build` should run the "build" task
  - Add DEBUG logging for delegation
  - Files: `internal/cli/commands.go`

- [x] **4.2** Update root `main.go` — redirect to CLI entry point
  - Option A: Replace root `main.go` with a thin wrapper that calls `cli.New().Run(os.Args[1:])`
  - Option B: Remove root `main.go` entirely and make `cmd/razd/main.go` the sole entry point (update `go.mod` or build scripts)
  - Recommended: Option B — clean removal, `cmd/razd/main.go` is already correct
  - Verify build works: `devbox run go build ./cmd/razd`
  - Files: `main.go` (remove or simplify), `go.mod` (if needed)

- [x] **4.3** Verify all example projects work end-to-end
  - `razd run default` in each example should parse Razdfile and execute tasks
  - `razd list` should show tasks for each example
  - `razd up --yes` should install tools (when provisioner is available)
  - Test with `devbox run go build ./cmd/razd && ./razd --help`
  - Files: no code changes, manual verification

### Phase 5: Integration Tests & Polish

- [ ] **5.1** Add integration tests for all commands
  - Create `internal/cli/integration_test.go`
  - Test each command with real Razdfile examples from `examples/`
  - Test error cases: no Razdfile, invalid Razdfile, missing provisioner
  - Test flag combinations: `--verbose`, `--silent`, `--dir`, `--yes`
  - Files: `internal/cli/integration_test.go`

- [ ] **5.2** Fix root `main.go` output to display Dependencies section
  - Currently `main.go` (the prototype) only prints `Devbox` and `Mise` sections
  - Add `HasDependencies()` output to the prototype (or ensure it's removed per Task 4.2)
  - This is a minor fix, primarily for exploration testing
  - Files: `main.go` (if kept)

## Commit Plan

- **Commit 1** (after Phase 1): `feat(cli): switch entry point to cmd/razd and extract shared helpers`
- **Commit 2** (after Phase 2): `feat(cli): implement run, list, and trust commands`
- **Commit 3** (after Phase 3): `feat(cli): implement init, add, and shell commands`
- **Commit 4** (after Phase 4): `feat(cli): finalize dev/build commands and remove prototype main.go`
- **Commit 5** (after Phase 5): `test(cli): add integration tests for all commands`

## Dependencies

```
Phase 1 → Phase 2 → Phase 3 → Phase 4 → Phase 5
Task 1.2 → Task 2.1 (helpers needed for run)
Task 2.1 → Task 4.1 (dev/build depend on run)
Task 4.2 → Task 4.3 (clean entry point needed for E2E testing)
```

## Open Questions

1. **Task execution approach**: `cmd_up.go` uses raw `exec.Command` with provisioner wrapping. Should `razd run` use go-task as a library (like current `main.go` prototype) or the simpler direct-exec approach? The go-task approach gives better task dependency resolution but adds complexity.

2. **Razdfile write-back**: `razd init` and `razd add` need to write/modify YAML. Should we use a YAML round-trip library to preserve comments/formatting, or is re-marshalling acceptable?