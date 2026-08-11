# Implementation Plan: `razd shell --print` — print shell activation script

Branch: 1.x
Created: 2026-08-11

## Original Request
я хочу добавить команду razd shell
чтобы активировалось допустим devbox shell (если стоит devbox) или  eval "$(mise activate bash)". Это для Unix систем
для mise и под windows надо аналог eval
вот инфа от агента из другой сессии:

[attachment: mise activate / eval / Windows iex explanation]

## Settings
- Testing: yes
- Logging: verbose
- Docs: yes

## Roadmap Linkage
Milestone: none
Rationale: No ROADMAP.md exists in the project.

## Research Context
Source: none (design from inline exploration; no RESEARCH.md committed)

**Core constraint (verified):** `razd shell` runs as a subprocess, so it CANNOT modify the
parent shell. Shell activation (`eval "$(mise activate bash)"`, `devbox shellenv`) works by
the command PRINTING a script that the user evaluates in the current shell. Therefore the
feature must be a `--print` mode that outputs the activation script, not a change to the
existing subshell behavior.

**Existing behavior:** `razd shell` (no args) launches an interactive subshell via
`prov.Shell()` — `mise shell` or `devbox shell`. This is backward-compatible and stays.

**Provisioner activation commands (verified via --help):**
- mise: `mise activate <shell>` — prints a script for the given shell type. Supported shell
  types: `bash, elvish, fish, nu, xonsh, zsh, pwsh`.
- devbox: `devbox shellenv --format <shell>` — prints shell env commands. `--format` default
  is `bash`; supports `nushell` and (per devbox docs) `pwsh`/`powershell` for Windows.

**Shell detection:** `$SHELL` env var basename (e.g. `/bin/bash` → `bash`, `/bin/zsh` → `zsh`,
`/usr/bin/fish` → `fish`). On Windows PowerShell, `$SHELL` may be unset; fall back to `pwsh`
when `$PSHOME`/`$PROFILE` is set or `pwsh` is in PATH.

**Windows eval analog (from agent):**
- PowerShell: `Invoke-Expression` / `iex` — `iex "$(razd shell --print)"`
- cmd: no direct analog (for /f is fragile); not supported.

**Design decision:** Add `razd shell --print` that prints the activation script for the
current shell. The user evaluates it: `eval "$(razd shell --print)"` (Unix) or
`iex "$(razd shell --print)"` (PowerShell). `razd shell` (no flag) keeps launching a subshell.

Constraints:
- `razd shell` (no args) keeps existing subshell behavior (backward compatible).
- `razd shell --print` prints the activation script to stdout, nothing else.
- Detect the current shell from `$SHELL`; map to the provisioner's supported shell name.
- mise → `mise activate <shell>`; devbox → `devbox shellenv --format <shell>`.
- Windows PowerShell → `pwsh` shell type; cmd → error (no analog).
- No provisioner configured → error with a clear message.

Decisions:
- Add a `--print` flag to the shell command (in `internal/flags/flags.go`).
- New function `printShellActivation(ctx, prov, shell)` in `internal/cli/cmd_shell.go` that
  runs the provisioner's activation command and streams its stdout.
- Shell detection helper `detectShell()` returns the shell name (bash/zsh/fish/pwsh/...).
- Keep `runShell` (subshell) and branch on the `--print` flag at the top.

Open questions:
- none

## Tasks

### Phase 1: Core implementation
- [x] Task 1: Add `--print` flag to the shell command
      In `internal/flags/flags.go`, add `ShellPrint bool` and register it in `Init()`:
      `pflag.BoolVar(&ShellPrint, "print", false, "Print the shell activation script instead of launching a subshell")`.
      LOGGING REQUIREMENTS:
      - None (flag registration).
      Files: `internal/flags/flags.go`.

- [x] Task 2: Implement shell detection and activation script printing
      In `internal/cli/cmd_shell.go`:
      - Add `detectShell() string`: read `$SHELL`, take the basename, and map to a shell name
        the provisioner understands (`bash`, `zsh`, `fish`, `pwsh`). If `$SHELL` is unset or
        empty, fall back to `pwsh` when `pwsh` is in PATH (Windows), else `bash`.
      - Add `printShellActivation(ctx, prov, shell) error`: run the provisioner's activation
        command and stream stdout to the user's stdout. For mise: `mise activate <shell>`.
        For devbox: `devbox shellenv --format <shell>`. Use `exec.CommandContext` with
        `cmd.Dir = prov dir`, `cmd.Stdout = os.Stdout`, `cmd.Stderr = os.Stderr`.
      - In `runShell`, at the top, if `flags.ShellPrint` is set, resolve the provisioner
        (reuse `getProvisioner`), call `printShellActivation`, and return. Keep the existing
        subshell path for the no-flag case.
      LOGGING REQUIREMENTS:
      - DEBUG: log the detected shell and the activation command
        (`ctx.Log.Debugf("Printing activation for shell=%s via %s\n", shell, prov.Name())`).
      - ERROR: return a clear error when no provisioner is configured or the shell type is
        unsupported (e.g. cmd).
      Files: `internal/cli/cmd_shell.go`.

### Phase 2: Tests
- [x] Task 3: Tests for shell detection and activation printing
      In `internal/cli/cmd_shell_test.go` (create if absent):
      - `TestDetectShell`: with `$SHELL=/bin/bash` → `bash`; `/bin/zsh` → `zsh`;
        unset `$SHELL` → `bash` (or `pwsh` if in PATH).
      - `TestPrintShellActivation_Mise`: a fake provisioner (or a stub) that records the
        activation command; assert the command is `mise activate bash` for shell `bash`.
      - `TestPrintShellActivation_Devbox`: assert the command is `devbox shellenv --format bash`.
      Use a fake provisioner implementing the `Provisioner` interface (see
      `internal/cli/cmd_up_test.go` for the existing `fakeProvisioner` pattern) or test the
      command-building helper directly.
      LOGGING REQUIREMENTS:
      - None (assert on observable command construction / output).
      Files: `internal/cli/cmd_shell_test.go`.

### Phase 3: Documentation
- [x] Task 4: Document `razd shell --print` in README
      Update `README.md`:
      - Commands section: add `razd shell --print` — print the shell activation script.
      - Add a short "Shell activation" note: `eval "$(razd shell --print)"` (Unix) or
        `iex "$(razd shell --print)"` (PowerShell) activates the current shell with the
        provisioner's environment; `razd shell` (no flag) launches a subshell.
      LOGGING REQUIREMENTS:
      - None (docs only).
      Files: `README.md`.

## Verification
- `devbox run -- go build ./...` and `devbox run -- go test ./...` pass.
- Smoke:
  - `razd shell --print` in a mise project → prints `mise activate bash` output.
  - `razd shell --print` in a devbox project → prints `devbox shellenv --format bash` output.
  - `eval "$(razd shell --print)"` in a mise project → `mise` function available in the
    current shell (verify `type mise` shows a function).
  - `razd shell` (no flag) still launches a subshell.
- `razd --help` shows the `--print` flag.

## Commit Plan
- **Commit 1** (after tasks 1-2): "feat(shell): add 'razd shell --print' to print activation script"
- **Commit 2** (after tasks 3-4): "test+docs: cover shell activation printing"
