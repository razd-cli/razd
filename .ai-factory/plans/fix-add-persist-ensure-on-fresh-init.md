# Implementation Plan: Fix `razd add` not persisting `ensure` on a fresh init

Branch: 1.x
Created: 2026-08-11

## Original Request
[FEATURE] razd add package_name (без указывания версии )

```
PS C:\Users\dealenx\dev\demo\fnox-example> razd add node
Error: invalid dependency format "node": invalid dependency format 'node', expected 'tool@version' (e.g., 'node@22')
```

## Settings
- Testing: yes
- Logging: verbose
- Docs: yes

## Roadmap Linkage
Milestone: none
Rationale: No ROADMAP.md exists in the project.

## Research Context
Source: none (findings from inline exploration; no RESEARCH.md committed)

Root cause: bare-name support for `razd add <tool>` is already implemented on `1.x`
(`ensureRegex` accepts a bare name with an empty version in `razdfile/ast/dependencies.go`,
verified end-to-end: `razd add node` succeeds when the Razdfile already has an `ensure` key).
However, `razd add` silently fails to persist when the Razdfile was freshly created by
`razd init`, which writes `dependencies.using` but no `ensure` key:

- `razdfile/writer.go` `UpdateEnsureInFile` only rewrites an **existing** `ensure` sequence
  node; if the key is absent it returns `false` without writing anything.
- `internal/cli/cmd_add.go` `runAdd` ignores that `false` return and always reports
  "Added N dependencies" + runs the mise/devbox sync, so the tool lands only in the native
  config (`mise.toml` / `devbox.json`) and never in `dependencies.ensure`.

Existing tests only exercise Razdfiles that already contain `ensure`
(`internal/cli/cmd_add_test.go` `writeAddRazdfile`, `razdfile/writer_test.go` fixtures), which
is why this slipped through.

Goal: teach the writer to insert an `ensure` sequence under `dependencies` when missing, and
have `runAdd` fail loudly if a write was expected but did not happen.

Constraints:
- Preserve YAML formatting/comments/key ordering (the writer already does this).
- Bare-name (`node`) and versioned (`node@22`) adds must both persist.
- Keep the existing rejection of empty strings and clearly invalid formats.
- Only create `ensure` under an existing `dependencies` mapping; do not create `dependencies`
  itself (both callers — `runAdd` and sync `applyToRazdfile` — guard on `HasDependencies()`).

Decisions:
- Fix at the source in `UpdateEnsureInFile` (benefits both `add` and sync callers).
- `runAdd` treats a `false` writer return as an error, since at that point `added` is
  non-empty and the ensure list must have changed.

Open questions:
- none

## Tasks

### Phase 1: Core fix
- [x] Task 1: Insert `ensure` sequence in the writer when absent
      In `razdfile/writer.go` `UpdateEnsureInFile`, track whether the `dependencies`
      mapping was found and whether an `ensure` key exists under it. When `dependencies`
      exists but `ensure` is missing, append a new scalar `ensure` key + an empty
      `yaml.SequenceNode` (tag `!!seq`) to `depNode.Content`, then populate it via the
      existing `setEnsureList` helper and set `changed = true`. Keep the current behavior
      when `dependencies` itself is absent (return `false`, no write).
      LOGGING REQUIREMENTS:
      - None (writer is a pure function with no logger); observable output is the returned
        bool and the written file.
      Files: `razdfile/writer.go`.

- [x] Task 2: Fail loudly when `add` expected a write but none happened
      In `internal/cli/cmd_add.go` `runAdd`, capture the bool returned by
      `razdfile.UpdateEnsureInFile(...)`. If it is `false` while `added` is non-empty, return
      `fmt.Errorf("failed to update Razdfile: ensure list was not written")` instead of
      proceeding to the success log + native sync. Preserve the existing success path.
      LOGGING REQUIREMENTS:
      - DEBUG: log when the writer reports no change for a non-empty add
        (`ctx.Log.Debugf("ensure list unchanged after add, expected a write\n")`) before
        returning the error.
      Files: `internal/cli/cmd_add.go`.

### Phase 2: Tests
- [x] Task 3: Writer test for missing `ensure` key
      In `razdfile/writer_test.go`, add `TestUpdateEnsureInFile_CreatesEnsureWhenAbsent`: a
      Razdfile with `version: "1"` and `dependencies.using: mise` (no `ensure`) is written
      via `UpdateEnsureInFile(path, []string{"node@22"})`; assert it returns `true` and the
      result file contains `node@22` under an `ensure` sequence. Add a parallel case for a
      bare name (`[]string{"node"}`) asserting the bare entry persists with no trailing `@`.
      LOGGING REQUIREMENTS:
      - None (assert on observable file content).
      Files: `razdfile/writer_test.go`.

- [x] Task 4: `add` command test for fresh-init persistence
      In `internal/cli/cmd_add_test.go`, add `TestRunAdd_CreatesEnsureOnFreshInit`: write a
      Razdfile with `dependencies.using: mise` and no `ensure`, run `runAdd` with
      `[]string{"node@22"}`, assert no error and the file contains `node@22`. Add
      `TestRunAdd_BareNameCreatesEnsure` for `[]string{"node"}` asserting bare `node` is
      written. Use a nil provisioner setup (matching existing tests' non-interactive path) so
      the native sync is skipped.
      LOGGING REQUIREMENTS:
      - None (assert on observable file content).
      Files: `internal/cli/cmd_add_test.go`.

### Phase 3: Documentation
- [x] Task 5: Note versionless + fresh-init behavior in README
      Update `README.md` Config Synchronization / `add` usage section to state that
      `razd add <tool>` accepts a bare package name (version defaults to "latest") and that
      `add` creates the `dependencies.ensure` list if the Razdfile lacks it (e.g. right after
      `razd init`).
      LOGGING REQUIREMENTS:
      - None (docs only).
      Files: `README.md`.

## Verification
- `devbox run -- go build ./...` and `devbox run -- go test ./...` pass.
- Smoke: `razd init` in a temp dir → `razd add node` → `Razdfile.yml` contains
  `- node` under `dependencies.ensure` (and `razd add node@22` writes `- node@22`).
- Confirmed regression reproduction before the fix: `add` reported success but left
  `Razdfile.yml` unchanged; after the fix the ensure list must be present.

### Phase 4: Sync prompt refinement
- [x] Task 6: Skip the sync direction/backup prompts when the native config file is missing
      In `internal/sync/sync.go` `Sync`, detect whether the native config file exists via
      `os.Stat(NativeConfig(prov.Name(), dir))`. When it does NOT exist (`os.IsNotExist`),
      there is nothing to reconcile against — `ReadConfig()` returns empty, so only
      Razdfile->native (`ToNative`) changes are possible. In that case skip BOTH the
      `PromptConfirmSync` direction prompt and the `PromptBackup` backup prompt, and apply
      the `ToNative` changes directly (writing the native file from the Razdfile). Keep the
      prompts whenever the native file exists (genuine version conflicts remain possible).
      LOGGING REQUIREMENTS:
      - DEBUG: log when the native config is absent and prompts are skipped
        (`log.Debugf("[SYNC] %s config missing, skipping sync prompts\n", prov.Name())`).
      Files: `internal/sync/sync.go`.

- [x] Task 7: Tests for the missing-native-config sync path
      In `internal/sync/sync_test.go`, add two cases using the existing `writeRazdfile`
      helper:
      - `TestSync_NoNativeFileSkipsPromptAndWritesNative`: a Razdfile with `ensure: [node]`
        and NO `mise.toml`; run `Sync(rf, prov, dir, noopLogger{}, false, true)` (confirmSync
        = true). Assert no error and that `mise.toml` is created containing `node`.
      - `TestSync_ExistingNativeFileStillPrompts`: a Razdfile with `ensure: [node]` and an
        existing `mise.toml` containing a different tool; assert the sync still reconciles
        (native file present → existing behavior unchanged).
      LOGGING REQUIREMENTS:
      - None (assert on observable file content).
      Files: `internal/sync/sync_test.go`.

## Commit Plan
- **Commit 1** (after tasks 1-5): "fix(add): persist dependencies.ensure when the Razdfile lacks an ensure key"
- **Commit 2** (after tasks 6-7): "fix(sync): skip direction/backup prompt when native config is missing"
