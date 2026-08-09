# Implementation Plan: Sync versionless devbox packages into Razdfile

Branch: 1.x
Created: 2026-08-09

## Original Request
Пользователь: при `razd up` на проекте `razd-php-laravel-example` (devbox) безверсионные пакеты `php84Packages.composer`, `php84Extensions.*` не синхронизируются из `devbox.json` в `Razdfile.yml`, хотя с mise это работает. Требуется, чтобы они зеркалились в `dependencies.ensure` и обратно.

## Settings
- Testing: yes
- Logging: verbose
- Docs: yes

## Roadmap Linkage
Milestone: none
Rationale: Skipped by user — no ROADMAP.md exists.

## Research Context
Source: none (investigated inline; no RESEARCH.md committed)

Root cause: `DevboxProvisioner.ReadConfig()` (provisioner/devbox.go) keeps only packages containing `@` (the `idx > 0` filter), silently dropping versionless devbox packages such as `php84Extensions.xdebug` and `php84Packages.composer`. `mise.ReadConfig()` has no such filter (it iterates the TOML `[tools]` table), which is why mise sync works but devbox does not.

Confirmed valid devbox format: jetify.com/docs/devbox/configuration states package version defaults to `latest`; a plain name without `@` is valid (list form `["php84Extensions.xdebug"]`, or map form `{"pkg": "latest"}`).

Versionless packages break in 4 places, not just ReadConfig:
1. `provisioner/devbox.go` ReadConfig — filter drops them.
2. `provisioner/devbox.go` writeConfig — `entry := tool + "@" + version` yields a trailing `@` (`xdebug@`) when version is empty.
3. `internal/sync/sync.go` applyToRazdfile — `entry := t.Name + "@" + t.Version` yields `xdebug@`; also `splitDep` only indexes entries that have `@`.
4. `razdfile/ast/dependencies.go` ensureRegex + ParseDependencyString — regex `^[a-z][a-z0-9._:/_-]*@[a-zA-Z0-9._-]+$` requires a version after `@`; a bare `php84Extensions.xdebug` is rejected.

Scope decision: only devbox versionless packages (not mise `@latest`).

## Tasks

### Phase 1: Core fix
- [x] Task 1: Include versionless packages in devbox ReadConfig
      In `provisioner/devbox.go` ReadConfig, change the loop so a package entry is kept regardless of whether it has `@`:
      - `name@version` → `tools[name] = version`.
      - bare `name` (no `@`) → `tools[name] = ""` (empty version).
      Use `strings.LastIndex` and only split when the index is > 0; do NOT drop entries that lack `@`.
      LOGGING REQUIREMENTS:
      - Keep existing error handling; no new logs required here (the caller logs). Optionally Debugf when a versionless package is parsed.
      Files: `provisioner/devbox.go`.

- [x] Task 2: Write versionless devbox packages without a trailing `@`
      In `provisioner/devbox.go` writeConfig, build the package entry conditionally:
      - version non-empty → `tool + "@" + version`
      - version empty → just `tool`
      Ensure the dedupe (`seen`) map and sort still work for both forms.
      LOGGING REQUIREMENTS:
      - Preserve the existing `[FIX] Writing devbox.json` verbose log.
      Files: `provisioner/devbox.go`.

- [x] Task 3: Accept and render versionless tools in the sync engine
      In `internal/sync/sync.go` applyToRazdfile, build the ensure entry conditionally so an empty version writes just the bare tool name (no trailing `@`), and index entries by name even when they have no `@`:
      - Adjust `splitDep` usage: an entry without `@` must still register in `indexByName` (name only, empty version).
      - `entry := t.Name` when version is empty, else `t.Name + "@" + t.Version`.
      Also confirm `Reconcile` treats an empty version sensibly (a versionless tool present on both sides with empty versions → same, no conflict; present on one side → added to the other).
      LOGGING REQUIREMENTS:
      - Keep existing `[SYNC] Added/Updated` Info logs; they must print the correct bare entry for versionless tools.
      Files: `internal/sync/sync.go`.

- [x] Task 4: Allow versionless entries in Razdfile dependencies.ensure
      In `razdfile/ast/dependencies.go`, widen `ensureRegex` (or the parse path) so a bare lowercase tool name without `@version` is accepted:
      - Accept `^[a-z][a-z0-9._:/_-]*$` (name only) as well as the existing `name@version`.
      - `ParseDependencyString` must return `Version: ""` for a bare name (Tool = full name, no `@`).
      Keep the existing rejection for empty strings and clearly invalid formats.
      LOGGING REQUIREMENTS:
      - None (parse path; error type/messages preserved).
      Files: `razdfile/ast/dependencies.go`.

### Phase 2: Tests
- [x] Task 5: Tests for versionless devbox sync
      Add tests covering:
      - `provisioner` devbox ReadConfig returns versionless packages with empty version and versioned ones with their version.
      - `provisioner` devbox writeConfig writes bare names (no trailing `@`) for empty versions and preserves existing entries.
      - `sync` applyToRazdfile writes bare entries (no `@`) for versionless tools and dedupes by name.
      - `razdfile/ast` ParseDependencyString accepts a bare name (empty version) and rejects still-invalid forms (e.g. empty string).
      - End-to-end: a Razdfile with versionless devbox ensure + a devbox.json containing `php84Extensions.xdebug` reconciles without producing `xdebug@`.
      Follow existing test style in `provisioner/provisioner_test.go`, `internal/sync/sync_test.go`, `razdfile/ast/dependencies_test.go`.
      LOGGING REQUIREMENTS:
      - Tests assert on observable output (config content), not log text.
      Files: `provisioner/provisioner_test.go`, `internal/sync/sync_test.go`, `razdfile/ast/dependencies_test.go`.

### Phase 3: Documentation
- [x] Task 6: Update docs for versionless devbox packages
      Update `README.md` (Config Synchronization section) to state that versionless devbox packages (e.g. `php84Extensions.*`) are synced into Razdfile `dependencies.ensure` as bare entries. If the example repo comment is in the docs, refresh the note that said these live only in devbox.json.
      LOGGING REQUIREMENTS:
      - None (docs only).
      Files: `README.md`.

## Verification
- `go build ./...` and `go test ./...` pass.
- Smoke: on `razd-php-laravel-example`, `razd up` (or a sync path) adds `php84Extensions.*` / `php84Packages.composer` as bare entries into Razdfile `dependencies.ensure`, with no trailing `@`.
- `devbox.json` packages are preserved (no data loss).

## Commit Plan
6 tasks — single commit at the end:
`fix(sync): sync versionless devbox packages into Razdfile`
