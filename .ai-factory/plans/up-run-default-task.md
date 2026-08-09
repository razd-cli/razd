# Implementation Plan: `razd up` runs default task after install

Branch: 1.x
Created: 2026-08-09

## Original Request
Пользователь: "если указана установку, то пусть будет, сам пользователь решает, главное чтобы запустилось razd default" (после `razd up <url>` должен запускаться default task проекта; fallback — если default task нет, просто установить пакеты из ensure через провижионер).

## Settings
- Testing: yes
- Logging: verbose
- Docs: yes

## Roadmap Linkage
Milestone: none
Rationale: Skipped by user — no ROADMAP.md exists in this repo.

## Context (from exploration)
- `razd up <url>` (cmd_up.go) currently: clone → read Razdfile → ensureTrusted → syncRazdfile → prov.Install() → **only if `flags.Run`** runDefaultTask(). Default task is NOT run without `--run`.
- The `runDefaultTask()` helper (cmd_up.go:128) is fully implemented and correct: it executes the default task, recursing into `task:` dependencies (install/dev) via `runTask()` and `executeCommand()`.
- Fallback requirement (no default task) is already satisfied by `runDefaultTask`: it warns `No 'default' task defined` and returns nil. `prov.Install()` already ran before it, so ensure packages are installed regardless.
- `--run`/`-r` flag is registered globally in flags.go:82 as `Run`. After the change it becomes an optional no-op (behavior default), kept for compatibility.
- Docs bug: cli.go:166 usage shows `-r, --razdfile`, but `-r` is actually bound to `--run` (flags.go:82); `--razdfile` has no short form. Fix while updating docs.
- `cmd_up_test.go` only tests `resolveUpDir` — the default-task-after-install contract is untested.

## Tasks

### Phase 1: Core change
- [x] Task 1: Make `razd up` always run the default task after install
      In `internal/cli/cmd_up.go`, replace the `if shouldRunAfterInstall() { ... }` gate with an unconditional call to `runDefaultTask`. Keep `shouldRunAfterInstall()` and `flags.Run` as-is (no-op compatibility — the `--run` flag remains accepted but is no longer required).
      - After `prov.Install()` succeeds, log `Running default task...` (Info) and `return runDefaultTask(ctx, rf, prov, dir)` unconditionally.
      - Remove the now-dead branch; update the `shouldRunAfterInstall` comment to note it is retained for backward compatibility with the `--run`/`-r` flag (currently unused by the flow).
      LOGGING REQUIREMENTS:
      - Log "Running default task..." at INFO before invoking runDefaultTask.
      - Keep existing Debugf logs inside runDefaultTask (already present).
      - Do not add new flags or change the CLI surface.
      Files: `internal/cli/cmd_up.go`.

- [x] Task 2: Add regression test that `up` runs the default task after install
      In `internal/cli/cmd_up_test.go`, add a test exercising `runUp` end-to-end with a stub provisioner and a Razdfile that has a `default` task. Assert that the default task's command is executed after installation (i.e. `prov.Install` was called and the default task's commands ran). Use the existing testwriter/logger pattern and stub `git.IsGitAvailableFunc`/`git.CloneFunc`/`git.ExtractRepoName` where needed.
      Also add a test that when the Razdfile has NO `default` task, `runUp` still installs (prov.Install called) and completes without error.
      LOGGING REQUIREMENTS:
      - Tests assert on observable behavior (install + default task run / no-default fallback), not on log text.
      Files: `internal/cli/cmd_up_test.go`.

### Phase 2: Documentation
- [x] Task 3: Update docs and help text
      Update `README.md` and `internal/cli/cli.go` usage/Examples so `razd up` is documented as "install tools and run default task", with `razd up --run` noted as an optional compatible flag (behavior now default).
      Fix the `-r, --razdfile` mislabel in cli.go:166 (the `-r` short flag is `--run`; `--razdfile` has no short form).
      LOGGING REQUIREMENTS:
      - None (docs only).
      Files: `README.md`, `internal/cli/cli.go`.

## Verification
- `go build ./...` succeeds.
- `go test ./...` — all packages pass, including new cmd_up_test cases.
- Manual smoke: `razd up <url>` on a repo with a default task runs install then the default task; on a repo without a default task installs ensure packages and exits cleanly.

## Commit Plan
Less than 5 tasks — single commit at the end:
`fix(up): run default task after install by default`
