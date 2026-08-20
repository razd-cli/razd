package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/razd-cli/razd/internal/flags"
	"github.com/razd-cli/razd/provisioner"
)

// runShell implements the "razd shell" command.
// By default it opens an interactive subshell with the provisioner activated.
// With --print it prints the activation script instead (for eval "$(razd shell
// --print)").
func runShell(ctx *Context) error {
	dir, err := resolveDir(ctx)
	if err != nil {
		return err
	}

	rf, err := readRazdfile(dir, ctx.Log)
	if err != nil {
		return err
	}

	prov, err := getProvisioner(rf, dir, ctx.Log)
	if err != nil {
		if !needsProvisioner(rf) {
			return fmt.Errorf("no provisioner configured — nothing to activate")
		}
		provName := provisionerName(rf)
		return fmt.Errorf("provisioner %q is not installed — install it to use 'razd shell', or use 'razd run' to execute tasks without provisioning", provName)
	}

	shell := detectShell()

	// --print: print the activation script to stdout (backward compat).
	if flags.ShellPrint {
		ctx.Log.Debugf("Printing activation for shell=%s via %s\n", shell, prov.Name())
		return printShellActivation(ctx, prov, dir, shell)
	}

	// devbox: delegate to `devbox shell` so the environment is fully isolated
	// (nix-built binaries + the "(devbox)" prompt prefix), matching how the
	// user expects a project shell to behave. mise has no isolation and keeps
	// the activation-script path.
	if prov.Name() == "devbox" {
		// Sync the native config with the Razdfile first so tools added/edited
		// in Razdfile are available inside the shell (honor --no-sync).
		if !flags.NoSync {
			if err := syncRazdfile(rf, prov, dir, ctx.Log); err != nil {
				ctx.Log.Warnf("Failed to sync %s config: %v\n", prov.Name(), err)
			}
		}
		ctx.Log.Debugf("Opening devbox shell via 'devbox shell'\n")
		return prov.Shell(context.Background())
	}

	ctx.Log.Debugf("Opening %s shell with activation via %s\n", shell, prov.Name())
	return openActivatedShell(ctx, prov, dir, shell)
}

// detectShell returns the name of the current shell (bash, zsh, fish, pwsh).
// It reads $SHELL and takes the basename. If $SHELL is unset, it falls back to
// pwsh on Windows (PowerShell), else bash. The GOOS parameter (runtime.GOOS in
// production) makes the fallback deterministic and testable across platforms.
func detectShell() string {
	return detectShellFor(runtime.GOOS)
}

func detectShellFor(goos string) string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		if goos == "windows" {
			return "pwsh"
		}
		return "bash"
	}
	base := filepath.Base(shell)
	base = strings.TrimSuffix(base, ".exe")
	switch base {
	case "bash", "zsh", "fish", "pwsh", "powershell":
		if base == "powershell" {
			return "pwsh"
		}
		return base
	default:
		return "bash"
	}
}

// openActivatedShell starts an interactive subshell whose startup file runs the
// provisioner activation first, so the provisioner is active inside the shell.
// On exit, the parent shell is unchanged.
func openActivatedShell(ctx *Context, prov provisioner.Provisioner, dir, shell string) error {
	// Capture the activation script output.
	args, err := activationCommand(prov.Name(), shell)
	if err != nil {
		return err
	}
	actCmd := exec.CommandContext(context.Background(), args[0], args[1:]...)
	actCmd.Dir = dir
	var actStderr bytes.Buffer
	actCmd.Stderr = &actStderr
	activation, err := actCmd.Output()
	if err != nil {
		// Surface the provisioner's own diagnostics (e.g. devbox failing to
		// install a package) so the user can see why activation failed,
		// instead of a bare "exit status 1".
		if stderrMsg := actStderr.String(); stderrMsg != "" {
			ctx.Log.Errf("%s activation error:\n%s", prov.Name(), strings.TrimRight(stderrMsg, "\n"))
		}
		return fmt.Errorf("failed to get activation script from %s: %w", prov.Name(), err)
	}

	launch, env, cleanup, err := activationStartup(shell, string(activation))
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}

	ctx.Log.Debugf("Launching %s with args %v\n", shell, launch)
	cmd := exec.CommandContext(context.Background(), launch[0], launch[1:]...)
	cmd.Dir = dir
	if env != nil {
		cmd.Env = env
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// activationStartup returns the argv to launch an interactive shell whose
// startup runs the given activation script, an optional env override (for
// shells like zsh that use ZDOTDIR), and a cleanup func for any temp
// files/dirs created. The activation script is the output of the provisioner's
// activation command (e.g. `mise activate bash`).
func activationStartup(shell, activation string) ([]string, []string, func(), error) {
	switch shell {
	case "bash":
		tmp, err := os.CreateTemp("", "razd-shell-*.sh")
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to create temp rcfile: %w", err)
		}
		// Write the activation verbatim; bash --rcfile sources it directly.
		// Wrapping it in `eval "..."` would re-parse the whole script in the
		// outer context, expanding $1/$@ to empty (breaking mise's
		// `[[ $1 != "x" ]]` guards) and splitting PATH entries containing
		// spaces into separate export arguments.
		if _, err := tmp.WriteString(activation); err != nil {
			tmp.Close()
			os.Remove(tmp.Name())
			return nil, nil, nil, fmt.Errorf("failed to write temp rcfile: %w", err)
		}
		// Prepend a "(Razd)" marker to the prompt so the user can see the
		// project environment is active. Opt out with RAZD_NO_PROMPT (mirrors
		// devbox's DEVBOX_NO_PROMPT).
		if _, err := tmp.WriteString(razdBashPrompt()); err != nil {
			tmp.Close()
			os.Remove(tmp.Name())
			return nil, nil, nil, fmt.Errorf("failed to write razd prompt: %w", err)
		}
		tmp.Close()
		return []string{"bash", "--rcfile", tmp.Name(), "-i"}, nil, func() { os.Remove(tmp.Name()) }, nil

	case "zsh":
		dir, err := os.MkdirTemp("", "razd-shell-*")
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to create temp zsh dir: %w", err)
		}
		rc := filepath.Join(dir, ".zshrc")
		// Write the activation verbatim; ZDOTDIR makes zsh source it as its
		// startup file. Wrapping it in `eval "..."` re-parses the whole script
		// in the outer context, breaking positional params and paths with
		// spaces (see bash case).
		if err := os.WriteFile(rc, []byte(activation), 0644); err != nil {
			os.RemoveAll(dir)
			return nil, nil, nil, fmt.Errorf("failed to write temp .zshrc: %w", err)
		}
		// Append the "(Razd)" prompt marker; opt out with RAZD_NO_PROMPT.
		f, err := os.OpenFile(rc, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			os.RemoveAll(dir)
			return nil, nil, nil, fmt.Errorf("failed to append razd prompt: %w", err)
		}
		if _, err := f.WriteString(razdZshPrompt()); err != nil {
			f.Close()
			os.RemoveAll(dir)
			return nil, nil, nil, fmt.Errorf("failed to append razd prompt: %w", err)
		}
		f.Close()
		env := append(os.Environ(), "ZDOTDIR="+dir)
		return []string{"zsh", "-i"}, env, func() { os.RemoveAll(dir) }, nil

	case "fish":
		// fish uses `source` syntax; the activation output is already fish-compatible.
		// Mirror devbox's fish prompt: copy the original fish_prompt then wrap it.
		prompt := "\nif not set -q RAZD_NO_PROMPT\n    functions -c fish_prompt __razd_orig_fish_prompt\n    function fish_prompt\n        echo -n '(Razd) '\n        __razd_orig_fish_prompt\n    end\nend\n"
		return []string{"fish", "-C", "eval " + activation + prompt}, nil, nil, nil

	case "pwsh":
		prompt := "if ($env:RAZD_NO_PROMPT -ne '1') { function global:prompt { '(Razd) ' + (Get-Location) + '> ' } }"
		return []string{"pwsh", "-NoExit", "-Command", "Invoke-Expression '" + strings.ReplaceAll(activation, "'", "''") + "'; " + prompt}, nil, nil, nil

	default:
		return nil, nil, nil, fmt.Errorf("unsupported shell %q for activation", shell)
	}
}

// razdBashPrompt returns a snippet that prepends "(Razd)" to the PS1 prompt,
// unless the user opted out with RAZD_NO_PROMPT. Mirrors devbox's
// DEVBOX_NO_PROMPT behavior.
func razdBashPrompt() string {
	return "\nif [ -z \"$RAZD_NO_PROMPT\" ]; then\n  export PS1=\"(Razd) $PS1\"\nfi\n"
}

// razdZshPrompt returns a snippet that prepends "(Razd)" to the zsh prompt at
// startup (the temp .zshrc replaces the user's own), unless RAZD_NO_PROMPT is
// set. One-time assignment is enough because ZDOTDIR bypasses the user config.
func razdZshPrompt() string {
	return "\nif [ -z \"$RAZD_NO_PROMPT\" ]; then\n  export PS1=\"(Razd) $PS1\"\nfi\n"
}

// printShellActivation runs the provisioner's activation command for the given
// shell and streams its stdout to the user's stdout. The user evaluates the
// output (eval "$(razd shell --print)" / iex "$(razd shell --print)") to
// activate the current shell.
func printShellActivation(ctx *Context, prov provisioner.Provisioner, dir, shell string) error {
	args, err := activationCommand(prov.Name(), shell)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(context.Background(), args[0], args[1:]...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// activationCommand returns the argv for the provisioner's shell activation
// command for the given shell type.
func activationCommand(provName, shell string) ([]string, error) {
	switch provName {
	case "mise":
		return []string{"mise", "activate", shell}, nil
	case "devbox":
		return []string{"devbox", "shellenv", "--format", shell}, nil
	default:
		return nil, fmt.Errorf("unsupported provisioner %q for shell activation", provName)
	}
}
