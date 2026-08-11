# Implementation Plan: `razd add task <name> [flags] -- <cmd>` — create tasks via CLI

Branch: 1.x
Created: 2026-08-11

## Original Request
Хорошо, давай

 razd add task <name> [flags] -- <cmd>

и сделай еще версию 1.7.0-dev.0

## Settings
- Testing: yes
- Logging: verbose
- Docs: yes

## Roadmap Linkage
Milestone: none
Rationale: No ROADMAP.md exists in the project.

## Research Context
Source: none (design from inline discussion; no RESEARCH.md committed)

Design decision (from conversation): add a CLI command to create tasks in the Razdfile,
mirroring `mise tasks add <TASK> [FLAGS] -- RUN...`. The `--` separator is the established
convention (mise, docker, git) and is consumed by pflag as the flag terminator, so everything
after `--` is the task command verbatim.

Syntax:
```
razd add node                    → package in dependencies.ensure (existing behavior)
razd add task hello -- echo 'hi' → create task "hello" with command "echo 'hi'"
razd add task hello --desc "..." --dep lint -- echo 'go test'   → task with flags
```

Key facts verified in code:
- `internal/flags/flags.go` uses `spf13/pflag`; `pflag.Parse()` consumes `--` as the flag
  terminator and returns everything after it via `pflag.Args()`. Verified empirically:
  `razd add task hello -- echo 'hi'` → `Args = [add, task, hello, echo, hi]`.
- `internal/cli/cli.go` dispatches `cmdArgs[0]` to a registered command; `add` → `runAdd`.
- `razdfile/ast/razdfile.go` reuses `go-task/task/v3/taskfile/ast` types: `Tasks` (ordered map),
  `Task` (has `Cmds []*Cmd`, `Desc`, `Deps`, `Dir`, `Silent`, `Interactive`, etc.), `Cmd`
  (has `Cmd string`). Task YAML supports scalar (`build: go build ./...`), sequence, and mapping
  forms.
- `razdfile/writer.go` `UpdateEnsureInFile` is the AST-preserving writer pattern (reads raw YAML
  as a Node tree, modifies only the target, writes back preserving comments/ordering). A similar
  `UpdateTasksInFile` is needed for tasks.
- `internal/cli/cmd_add.go` `runAdd` currently only handles dependencies; it must branch on
  `Args[0] == "task"`.

Constraints:
- `razd add <pkg>` (no `task` keyword) must keep existing dependency behavior unchanged.
- `razd add task <name> -- <cmd>` writes into the `tasks:` section of Razdfile.yml.
- Preserve YAML formatting/comments/key ordering (AST-preserving writer).
- If `tasks:` section is absent, create it (like the `ensure` fix).
- Task name must be valid (no `@`, no whitespace, no `:` namespace separator for a simple task).
- `--` is required to separate task flags from the command; without it, the command would be
  parsed as razd flags.
- Flags: `--desc`, `--dep` (repeatable), `--dir`, `--silent`, `--interactive` (mirror mise).

Decisions:
- Branch on `Args[0] == "task"` inside `runAdd`; keep dependency path untouched.
- Register task flags (`--desc`, `--dep`, `--dir`, `--silent`, `--interactive`) in
  `internal/flags/flags.go` so pflag parses them before `--`.
- New writer function `UpdateTasksInFile` in `razdfile/writer.go` following the
  `UpdateEnsureInFile` pattern.
- Write scalar form (`name: cmd`) for a single command with no flags; mapping form
  (`name: {cmds: [...], desc: ...}`) for multiple commands or any flags.

Open questions:
- none

## Tasks

### Phase 1: Core implementation
- [x] Task 1: Register task flags in the flags package
      In `internal/flags/flags.go`, add exported vars and register them in `Init()`:
      `TaskDesc string` (`--desc`), `TaskDeps []string` (`--dep`, repeatable, `StringSliceVar`),
      `TaskDir string` (`--dir`), `TaskSilent bool` (`--silent`), `TaskInteractive bool`
      (`--interactive`). These are parsed by pflag before the `--` terminator.
      LOGGING REQUIREMENTS:
      - None (flag registration; no runtime logging).
      Files: `internal/flags/flags.go`.

- [x] Task 2: Add AST-preserving task writer
      In `razdfile/writer.go`, add `UpdateTasksInFile(filePath string, taskName string, task
      *taskast.Task) (bool, error)` following the `UpdateEnsureInFile` pattern: read raw YAML as
      a Node tree, locate the `tasks:` mapping (create it under root if absent), and set/update
      the task key. Serialize the task to YAML nodes: scalar form (`name: <cmd>`) when the task
      has exactly one command and no flags; mapping form (`name: {cmds: [...], desc: ...}`)
      otherwise. Preserve existing tasks, comments, and key ordering. Return true if changed.
      LOGGING REQUIREMENTS:
      - None (pure function; observable output is the returned bool and written file).
      Files: `razdfile/writer.go`.

- [x] Task 3: Branch `runAdd` on the `task` keyword
      In `internal/cli/cmd_add.go`, at the top of `runAdd`, detect `ctx.Args[0] == "task"` and
      delegate to a new `runAddTask(ctx, args[1:])`. Keep the existing dependency path unchanged
      for all other inputs. `runAddTask` parses `<name> [flags] -- <cmd...>`: name is the first
      positional arg, the command is everything after it (pflag already stripped `--`). Build a
      `taskast.Task` with `Cmds`, `Desc`, `Deps`, `Dir`, `Silent`, `Interactive` from the flags,
      call `UpdateTasksInFile`, and report success. Validate: name non-empty, no `@`, no
      whitespace, no `:`; at least one command required.
      LOGGING REQUIREMENTS:
      - DEBUG: log parsed task name, command, and flags
        (`ctx.Log.Debugf("Adding task %q with command %q\n", name, cmd)`).
      - INFO: log each flag applied (`--desc`, `--dep`, `--dir`, `--silent`, `--interactive`).
      - ERROR: return descriptive errors for invalid name / missing command.
      Files: `internal/cli/cmd_add.go`.

### Phase 2: Tests
- [x] Task 4: Writer tests for task insertion
      In `razdfile/writer_test.go`, add tests for `UpdateTasksInFile`:
      - `TestUpdateTasksInFile_CreatesTasksSection`: Razdfile without `tasks:` → task written,
        section created.
      - `TestUpdateTasksInFile_ScalarForm`: single command, no flags → `name: cmd` scalar.
      - `TestUpdateTasksInFile_MappingForm`: multiple commands or flags → `name: {cmds: [...],
        desc: ...}` mapping.
      - `TestUpdateTasksInFile_PreservesExisting`: existing tasks/comments preserved.
      LOGGING REQUIREMENTS:
      - None (assert on observable file content).
      Files: `razdfile/writer_test.go`.

- [x] Task 5: `add task` command tests
      In `internal/cli/cmd_add_test.go`, add:
      - `TestRunAddTask_CreatesTask`: `Args = ["task", "hello", "echo", "hi"]` → task written.
      - `TestRunAddTask_WithFlags`: `Args = ["task", "hello", "echo", "go test"]` with
        `flags.TaskDesc`/`TaskDeps` set → mapping form with desc/deps.
      - `TestRunAddTask_InvalidName`: name with `@`/whitespace → error.
      - `TestRunAddTask_MissingCommand`: no command after name → error.
      - `TestRunAdd_StillAddsDependency`: `Args = ["node@22"]` → dependency path unchanged.
      LOGGING REQUIREMENTS:
      - None (assert on observable file content).
      Files: `internal/cli/cmd_add_test.go`.

### Phase 3: Documentation
- [x] Task 6: Document `razd add task` in README and fix `add` help text
      Update `README.md` Commands section: add `razd add task <name> -- <cmd>` with flags
      (`--desc`, `--dep`, `--dir`, `--silent`, `--interactive`). In `internal/cli/cli.go`
      `printUsage`, fix the misleading `add  Add a new task` line to
      `add  Add a dependency or task (razd add task <name> -- <cmd>)`.
      LOGGING REQUIREMENTS:
      - None (docs only).
      Files: `README.md`, `internal/cli/cli.go`.

## Verification
- `devbox run -- go build ./...` and `devbox run -- go test ./...` pass.
- Smoke: `razd init` in a temp dir → `razd add task hello -- echo 'hi'` → Razdfile.yml contains
  `hello: echo 'hi'` under `tasks:`. `razd add task build --desc "Build" -- go build ./...` →
  mapping form. `razd add node` still adds a dependency.
- `razd list` shows the new task.

## Commit Plan
- **Commit 1** (after tasks 1-3): "feat(add): support 'razd add task <name> -- <cmd>'"
- **Commit 2** (after tasks 4-5): "test(add): cover task creation via CLI"
- **Commit 3** (after task 6): "docs: document razd add task and fix add help text"
