# Implementation Plan: Smarter `razd add` — bare `task` package + conflict-aware sync

Branch: 1.x
Created: 2026-08-11

## Original Request
можно ли сделать исключение что если не добавляется -- то пусть ставит task пакет

и когда ставится то надо более умно спрашивать sync

то есть если нет конфликта версий и ставится через razd пакет то логично что оно из razdfile в mise будет записываться
если перезаписывается запись допустим стоит в mise.toml node@24 а я устанавливаю razd add node@26 то спросит что сделать типо взять из razd версию или нет
и надо первым пунктом предлагать копировать из razd

## Settings
- Testing: yes
- Logging: verbose
- Docs: yes

## Roadmap Linkage
Milestone: none
Rationale: No ROADMAP.md exists in the project.

## Research Context
Source: none (design from inline exploration; no RESEARCH.md committed)

Two problems reproduced empirically:

**Problem A — `razd add task` is always intercepted as the task-creation keyword.**
`razd add task` (bare) errors with "usage: razd add task <name> -- <cmd>" instead of
adding the `task` package. `razd add task@latest` works (goes to dependency path) but the
bare form is blocked. Discriminator: `Args[0]=="task"` currently always routes to
`runAddTask`. Fix: only route to `runAddTask` when there is a task name after `task`
(`len(args) > 1`); a bare `razd add task` (no name) should fall through to the dependency
path and add the `task` package.

**Problem B — the sync direction prompt fires even when the direction is obvious.**
Reproduced:
- `razd add python` (new package, mise.toml exists, no conflict) → shows "How to
  synchronize config?" direction prompt, even though the only change is Razdfile→native.
- `razd add node@26` (conflict: mise has node@24) → shows BOTH the "Version conflict"
  prompt AND then the "How to synchronize config?" direction prompt. The user already
  resolved the conflict; the second prompt is redundant.

Desired behavior:
- New package added via razd, no version conflict → write Razdfile→native silently
  (no direction prompt).
- Version conflict → ONE prompt ("Version conflict"), with "Use Razdfile" as the FIRST
  option (recommended). After the user picks, apply silently — no second direction prompt.
- Native config file missing → already silent (existing fix).
- Reverse direction (native→Razdfile) is only relevant when a tool appears in native
  without being added via razd (e.g. manual `mise use node`); that path already flows
  through `ToRazdfile` and is applied without a direction prompt when there's no conflict.

Key facts verified in code:
- `internal/cli/cli.go` dispatches `cmdArgs[0]` to a registered command; `add` → `runAdd`.
- `internal/cli/cmd_add.go` `runAdd` branches on `ctx.Args[0] == "task"` → `runAddTask`.
- `internal/sync/sync.go` `Sync`: after `Reconcile`, if the native file exists and
  `confirmSync` is true, it always calls `PromptConfirmSync` (direction prompt). The
  missing-file case already skips it.
- `internal/sync/engine.go` `Reconcile`: rule 4 (version conflict) invokes the resolver
  (`PromptConflict`); rule 1 (only in Razdfile) → `ToNative`; rule 2 (only in native) →
  `ToRazdfile`.
- `internal/sync/conflict.go` `PromptConflict`: options are [Use Razdfile, Use native,
  Skip] — "Use Razdfile" is already first, but the direction prompt still fires after.
- `internal/sync/confirm.go` `PromptConfirmSync`: the direction prompt.

Constraints:
- `razd add <pkg>` (no `task` keyword) keeps existing dependency behavior.
- `razd add task` (bare) → add the `task` package to dependencies.ensure.
- `razd add task <name> -- <cmd>` → create a task (unchanged).
- No direction prompt when the direction is unambiguous (new package via razd, or after a
  resolved conflict).
- Version conflict prompt keeps "Use Razdfile" as the first/recommended option.
- Preserve the existing missing-native-file silent path.

Decisions:
- In `runAdd`, route to `runAddTask` only when `len(ctx.Args) > 1` (a task name follows);
  a bare `razd add task` falls through to the dependency path.
- In `Sync`, skip `PromptConfirmSync` when there are no `ToRazdfile` changes (i.e. the only
  direction is Razdfile→native, which is the natural result of `razd add`). The direction
  prompt is only meaningful when both directions have changes (genuine ambiguity).
- Keep `PromptConflict` as the single prompt for version conflicts; after it resolves, the
  direction is determined and applied without a second prompt.

Open questions:
- none

## Tasks

### Phase 1: Core implementation
- [x] Task 1: Allow bare `razd add task` to add the `task` package
      In `internal/cli/cmd_add.go` `runAdd`, change the branch condition from
      `ctx.Args[0] == "task"` to `ctx.Args[0] == "task" && len(ctx.Args) > 1`. A bare
      `razd add task` (no name) then falls through to the dependency path and adds the
      `task` package to `dependencies.ensure` (and syncs `task=latest` to native via the
      existing dependency sync). `razd add task <name> -- <cmd>` still creates a task.
      LOGGING REQUIREMENTS:
      - DEBUG: log the branch decision (`ctx.Log.Debugf("add task: %d args, routing to %s\n",
        len(ctx.Args), ...)`).
      Files: `internal/cli/cmd_add.go`.

- [x] Task 2: Skip the direction prompt when only Razdfile→native changes exist
      In `internal/sync/sync.go` `Sync`, after computing `changes`, only call
      `PromptConfirmSync` when `len(changes.ToRazdfile) > 0` (i.e. there are native→Razdfile
      changes that make the direction ambiguous). When `ToRazdfile` is empty, the only
      direction is Razdfile→native — apply it silently. Keep the existing missing-native-file
      early path. This makes `razd add python` (new package) write to native without prompting,
      and `razd add node@26` (conflict) apply the conflict resolution without a second prompt.
      LOGGING REQUIREMENTS:
      - DEBUG: log when the direction prompt is skipped because only Razdfile→native changes
        exist (`log.Debugf("[SYNC] only Razdfile->native changes, skipping direction prompt\n")`).
      Files: `internal/sync/sync.go`.

- [x] Task 3: Ensure "Use Razdfile" is the first/recommended conflict option
      In `internal/sync/conflict.go` `PromptConflict`, confirm the option order is
      [Use Razdfile, Use native, Skip] and that "Use Razdfile" is the default/highlighted
      choice. If huh supports a default selection, set it to the Razdfile option (index 0).
      This is mostly a verification task — the order is already correct, but make the
      Razdfile option the explicit default so the user can press Enter to accept it.
      LOGGING REQUIREMENTS:
      - None (prompt ordering; existing logs already cover the resolution).
      Files: `internal/sync/conflict.go`.

### Phase 2: Tests
- [x] Task 4: Tests for bare `razd add task` and sync direction
      In `internal/cli/cmd_add_test.go`, add `TestRunAdd_BareTaskAddsPackage`: `Args =
      ["task"]` → the `task` package is added to `dependencies.ensure` (not an error, no
      task created). In `internal/sync/sync_test.go`, add:
      - `TestSync_NewPackageNoDirectionPrompt`: Razdfile has `python`, mise.toml exists with
        an unrelated tool, confirmSync=true → `python` written to native, no direction prompt
        (assert the native file contains python).
      - `TestSync_ConflictAppliesWithoutSecondPrompt`: Razdfile has `node@26`, mise.toml has
        `node@24`, confirmSync=true, non-interactive → conflict resolves to skip (safe
        default) and no direction prompt fires (assert neither side changed).
      LOGGING REQUIREMENTS:
      - None (assert on observable file content).
      Files: `internal/cli/cmd_add_test.go`, `internal/sync/sync_test.go`.

### Phase 3: Documentation
- [x] Task 5: Document bare `razd add task` and sync behavior in README
      Update `README.md`:
      - Commands section: note that `razd add task` (bare) adds the `task` package, while
        `razd add task <name> -- <cmd>` creates a task.
      - Config Synchronization section: note that a package added via `razd add` is written
        to the native config without a direction prompt when there is no version conflict;
        version conflicts show a single prompt with "Use Razdfile" as the recommended option.
      LOGGING REQUIREMENTS:
      - None (docs only).
      Files: `README.md`.

## Verification
- `devbox run -- go build ./...` and `devbox run -- go test ./...` pass.
- Smoke:
  - `razd add task` (bare) → `task` added to `dependencies.ensure`, no usage error.
  - `razd add task hello -- echo 'hi'` → task created (unchanged).
  - `razd add python` (new package, mise.toml exists) → written to native, NO direction
    prompt.
  - `razd add node@26` (mise has node@24) → single "Version conflict" prompt with
    "Use Razdfile" first; after choosing, applied without a second prompt.
- `razd list` still works.

## Commit Plan
- **Commit 1** (after tasks 1-3): "feat(add): bare 'razd add task' adds package; smarter sync direction"
- **Commit 2** (after tasks 4-5): "test+docs: cover bare task add and conflict-aware sync"
