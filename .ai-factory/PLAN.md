# Plan: Interactive Trust Prompt with Arrow-Key Selection

**Branch:** 1.x (current)
**Created:** 2026-06-02
**Settings:** Testing: yes | Logging: verbose | Docs: no

## Problem

When a user runs `razd up` or `razd run` on an untrusted project, razd prints a message asking them to run `razd trust` separately and exits with error code 300. No interactive prompt is shown — the user must manually re-run a different command. The UX should let the user trust a project inline with arrow-key selection (No default, Y switches to Yes).

## Architecture Decision

Use `charmbracelet/huh` for the interactive prompt. It provides:
- Arrow-key selectable options with visual highlighting
- Y/N keybindings (press Y to jump to "Yes")
-TTY detection (auto-fallback when stdin is not a terminal)
- Consistent look with `fatih/color` (already in go.mod)

Alternative considered: raw `fmt.Scan` + `bufio.Scanner` — rejected because it cannot do arrow-key selection.

## Tasks

- [x] T1: Add `charmbracelet/huh` dependency
  - Run `go get github.com/charmbracelet/huh`
  - Run `go mod tidy`
  - File: `go.mod`, `go.sum`

- [x] T2: Create `internal/trust/prompt.go` — interactive trust prompt
  - New file: `internal/trust/prompt.go`
  - Implement `PromptTrust(path string, stdin io.Reader) (bool, error)`:
    - Detect if stdin is a terminal (`os.Stdin.Stat()` or `term.IsTerminal()`)
    - If not a terminal (CI, pipe): return `false, nil` with a warning
    - If `--yes` flag is set: return `true, nil` (auto-trust, no prompt)
    - Otherwise: show `huh.NewSelect` with options "No" (default) and "Yes", press Y to select Yes
    - On "Yes": return `true, nil`
    - On "No": return `false, nil`
  - Use `huh.NewSelect[string]` with `huh.WithTheme(huh.ThemeCatppuccin())` or a custom theme matching razd's output style
  - Display project path in the prompt title
  - Logging: `[FIX] PromptTrust: path=<path>, result=<trusted|declined|non-interactive>`

- [x] T3: Wire prompt into `EnsureTrusted()` in `internal/trust/check.go`
  - Modify `EnsureTrusted(path string, log *output.Logger, autoTrust bool) (bool, error)`:
    - Keep existing behavior for `StatusTrusted` and `StatusIgnored`
    - For `StatusUnknown` without `autoTrust`:
      - Call `PromptTrust(path, os.Stdin)` instead of just logging and returning `false`
      - If user chose Yes: call `Trust()` inline and return `true, nil`
      - If user chose No or non-interactive: return `false, nil`
  - Add `os` import if not present
  - Logging: `[FIX] EnsureTrusted: status=unknown, prompting user`

- [x] T4: Simplify `ensureTrusted()` in `internal/cli/helpers.go`
  - Current code has complex branching for `!trusted && !flags.Yes` and `!trusted && flags.Yes`
  - After T3, `EnsureTrusted()` handles both interactive and auto-trust cases directly
  - Simplify `ensureTrusted()`:
    ```go
    func ensureTrusted(dir string, prov provisioner.Provisioner, log *output.Logger) error {
        trusted, err := trust.EnsureTrusted(dir, log, flags.Yes)
        if err != nil {
            return err
        }
        if !trusted {
            return &errors.TrustError{Path: dir, Message: "project not trusted"}
        }
        return nil
    }
    ```
  - Remove the separate `trust.Trust()` call that was in `ensureTrusted` — it's now handled inside `PromptTrust`/`EnsureTrusted`

- [x] T5: Add terminal detection utility
  - New file: `internal/trust/tty.go`
  - Implement `IsTerminal() bool` using `golang.org/x/term.IsTerminal(int(os.Stdin.Fd()))`
  - `golang.org/x/term` is already an indirect dependency in go.mod
  - This cleanly separates TTY detection from prompt logic

- [x] T6: Write tests for trust prompt
  - New file: `internal/trust/prompt_test.go`
  - Test cases:
    - `TestPromptTrust_NonTerminal`: mock non-TTY stdin, verify it returns `false, nil` without prompting
    - `TestPromptTrust_AutoTrust`: with `autoTrust=true`, verify it returns `true, nil` without prompting
    - `TestPromptTrust_TTY_SelectYes`: simulate user selecting "Yes"
    - `TestPromptTrust_TTY_SelectNo`: simulate user selecting "No"
  - New file: `internal/trust/tty_test.go`
    - `TestIsTerminal`: basic coverage

- [x] T7: Build, test, bump version, tag, and push
  - `go build ./...`
  - `go test ./...`
  - Bump tag to v1.1.0 (minor bump — new feature)
  - Push and verify GitHub Actions

## Commit Plan

- Commit 1 (T1+T2+T5): `feat(trust): add interactive prompt with arrow-key selection`
- Commit 2 (T3+T4): `feat(trust): wire interactive prompt into ensureTrusted flow`
- Commit 3 (T6): `test(trust): add tests for prompt and TTY detection`
- Commit 4 (T7): version bump + tag

## Edge Cases

- **Non-interactive shell (CI/pipe):** Must not hang. `IsTerminal()` returns false, `PromptTrust` returns false, and trust error propagates normally. User must use `--yes` or `razd trust`.
- **WSL2:** `golang.org/x/term` works on WSL2. `huh` uses the same TTY detection.
- **Windows Terminal:** `huh` supports Windows via virtual terminal sequences.
- **Signal interruption:** `huh` handles SIGINT gracefully (returns ErrUserAborted).