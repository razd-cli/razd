# Implementation Plan: `razd shell` opens a subshell with the provisioner activated

Branch: 1.x
Created: 2026-08-11

## Original Request
Я хочу чтобы отработал razd shell и он внутри себя запустил eval "$(mise activate bash)" или как-то еще прокинул

## Settings
- Testing: yes
- Logging: verbose
- Docs: yes

## Roadmap Linkage
Milestone: none
Rationale: No ROADMAP.md exists in the project.

## Research Context
Source: none (design from inline exploration; no RESEARCH.md committed)

**Current state:** `razd shell` (from `v1.8.0-dev.1`, commit `e455ae6`) prints the
provisioner's activation script (`mise activate <shell>` / `devbox shellenv`) to stdout, and
the user must run `eval "$(razd shell)"` themselves. The user wants `razd shell` to open an
interactive shell where the provisioner is ALREADY activated — without printing the script.

**Core mechanic (verified):** A subprocess cannot modify its parent shell. But it CAN start a
NEW interactive subshell whose startup file runs the activation first. This is the `nix-shell`
/ `devbox shell` pattern. For bash, `bash --rcfile <file>` runs `<file>` (which contains
`eval "$(mise activate bash)"`) before showing the prompt. Verified empirically:
`bash --rcfile <(echo 'export TEST=1') -i` → `echo $TEST` prints `1`.

**Shell startup overrides (verified):**
- bash: `bash --rcfile <file>` — runs `<file>` instead of `~/.bashrc`.
- zsh: `ZDOTDIR=<tmpdir> zsh -i` where `<tmpdir>/.zshrc` contains the activation, then `eval`.
- fish: `fish -C "eval (mise activate fish)"` runs a command before interactive loop.
- PowerShell: `pwsh -NoExit -Command "Invoke-Expression (razd shell --print)"` (or similar).

**Design:** `razd shell` should:
1. Detect the shell (reuse `detectShell()`).
2. Get the provisioner's activation command argv (reuse `activationCommand()`).
3. Start an interactive subshell whose startup runs `eval "$(activation command output)"`.
4. On exit, return to the parent shell (activation does NOT leak to the parent).

The previous `--print`/print-default behavior is superseded. Keep it simple: `razd shell`
opens the activated subshell. (Optionally keep a `--print` flag to print the script instead,
for `eval "$(razd shell --print)"` use.)

Constraints:
- `razd shell` must open an interactive shell with the provisioner activated.
- The activation must NOT print to stdout (no huge script dump).
- Must work for bash, zsh, fish, and PowerShell (pwsh).
- On `exit` from the subshell, return to the parent shell unchanged.
- Keep backward compat: `razd shell` currently prints the script — this changes to open a
  subshell. (Add `--print` flag to preserve the print behavior if needed.)

Decisions:
- Use a temp startup file / env var per shell to inject activation into the subshell.
- bash → `bash --rcfile <tmpfile>` where tmpfile = `eval "$(mise activate bash)"`.
- zsh → `ZDOTDIR=<tmpdir> zsh -i` with `.zshrc` containing activation.
- fish → `fish -C "<activation>"`.
- pwsh → `pwsh -NoExit -Command "iex <activation>"`.
- Clean up the temp file after the shell exits.
- Keep `activationCommand()` for building the provisioner activation argv.

Open questions:
- none

## Tasks

### Phase 1: Core implementation
- [x] Task 1: Add an interactive-shell launcher with activation
      In `internal/cli/cmd_shell.go`, add `openActivatedShell(ctx, prov, dir, shell) error`
      that starts an interactive subshell whose startup file runs the provisioner activation:
      - Build the activation command via `activationCommand(prov.Name(), shell)` and capture
        its output (`cmd.Output()`).
      - Build a temp startup mechanism per shell:
        - bash: write a temp file containing `eval "<activation output>"` and run
          `bash --rcfile <tmpfile> -i`.
        - zsh: create a temp dir with a `.zshrc` containing the activation, set `ZDOTDIR` to
          it, run `zsh -i`.
        - fish: run `fish -C "<activation output>"` (fish uses `source` syntax; escape
          accordingly).
        - pwsh: run `pwsh -NoExit -Command "Invoke-Expression '<activation output>'"`.
      - Wire stdin/stdout/stderr to the terminal.
      - Clean up the temp file/dir after the shell exits.
      LOGGING REQUIREMENTS:
      - DEBUG: log the detected shell, the activation argv, and the launch command
        (`ctx.Log.Debugf("Opening %s shell with activation via %s\n", shell, prov.Name())`).
      - ERROR: return a clear error if the activation command fails or the shell is
        unsupported.
      Files: `internal/cli/cmd_shell.go`.

- [x] Task 2: Make `razd shell` open the activated subshell (and add `--print` opt-out)
      In `internal/cli/cmd_shell.go` `runShell`:
      - Default: resolve the provisioner and call `openActivatedShell(...)`.
      - Add a `--print` flag (in `internal/flags/flags.go`, `ShellPrint bool`) that keeps the
        old behavior: print the activation script to stdout (for `eval "$(razd shell
        --print)"`). When `--print` is set, call `printShellActivation(...)`.
      - Keep `detectShell()` and `activationCommand()`.
      LOGGING REQUIREMENTS:
      - DEBUG: log whether `--print` or subshell mode is used.
      Files: `internal/cli/cmd_shell.go`, `internal/flags/flags.go`.

### Phase 2: Tests
- [x] Task 3: Tests for activation script injection and shell detection
      In `internal/cli/cmd_shell_test.go` (extend):
      - `TestActivationScript_*`: for each shell type, assert the injected startup content
        contains the activation command (`mise activate bash`, `devbox shellenv --format
        bash`, etc.). Factor the per-shell startup-building logic into a testable helper
        `activationStartup(provName, shell) (launchArgs []string, content string, err error)`
        that returns the shell argv and the startup file content.
      - Keep existing `TestDetectShell` and `TestActivationCommand_*`.
      LOGGING REQUIREMENTS:
      - None (assert on observable startup content / argv).
      Files: `internal/cli/cmd_shell_test.go`.

### Phase 3: Documentation
- [x] Task 4: Document `razd shell` subshell behavior in README
      Update `README.md`:
      - Commands section: `razd shell` opens an interactive shell with the provisioner
        activated; `razd shell --print` prints the activation script.
      - Shell activation section: rewrite to describe the subshell (temporary activation,
        `exit` to return) and the `--print` alternative for current-session activation via
        `eval "$(razd shell --print)"`.
      LOGGING REQUIREMENTS:
      - None (docs only).
      Files: `README.md`.

## Verification
- `devbox run -- go build ./...` and `devbox run -- go test ./...` pass.
- Smoke:
  - `razd shell` in a mise project → opens an interactive shell; `type mise` shows a function
    (activated); `exit` returns to the parent shell.
  - `razd shell --print` → still prints the activation script (backward compat).
  - `eval "$(razd shell --print)"` → activates the current shell.
- `razd --help` shows the `--print` flag.

## Commit Plan
- **Commit 1** (after tasks 1-2): "feat(shell): open interactive subshell with provisioner activated"
- **Commit 2** (after tasks 3-4): "test+docs: cover shell subshell activation"
