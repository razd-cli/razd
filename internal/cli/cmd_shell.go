package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/razd-cli/razd/internal/errors"
	"github.com/razd-cli/razd/internal/flags"
	"github.com/razd-cli/razd/provisioner"
)

// runShell implements the "razd shell" command.
// It starts an interactive shell with the provisioned environment, or prints
// the shell activation script when --print is set.
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
			if flags.ShellPrint {
				return fmt.Errorf("no provisioner configured — nothing to activate")
			}
			ctx.Log.Infof("No provisioner configured, starting system shell\n")
			return runFallbackShell(ctx)
		}
		provName := provisionerName(rf)
		return fmt.Errorf("provisioner %q is not installed — install it to use 'razd shell', or use 'razd run' to execute tasks without provisioning", provName)
	}

	// --print: print the shell activation script for the current shell.
	if flags.ShellPrint {
		shell := detectShell()
		ctx.Log.Debugf("Printing activation for shell=%s via %s\n", shell, prov.Name())
		return printShellActivation(ctx, prov, dir, shell)
	}

	if err := ensureTrusted(dir, prov, ctx.Log); err != nil {
		return err
	}

	// Synchronize the Razdfile with the native config so tools added/edited in
	// either file are available in the provisioned shell, honoring --no-sync.
	if !flags.NoSync {
		if err := syncRazdfile(rf, prov, dir, ctx.Log); err != nil {
			ctx.Log.Warnf("Sync failed: %v\n", err)
		}
	}

	ctx.Log.Debugf("Starting shell with provisioner: %s\n", prov.Name())
	ctx.Log.Infof("Starting %s shell...\n", prov.Name())

	shellCtx := context.Background()
	if err := prov.Shell(shellCtx); err != nil {
		// If the shell command is not found, try fallback
		if isCommandNotFound(err) {
			ctx.Log.Warnf("%s shell not available, falling back to system shell\n", prov.Name())
			return runFallbackShell(ctx)
		}
		return &errors.TaskRunError{
			TaskName: "shell",
			Err:      err,
		}
	}

	return nil
}

// detectShell returns the name of the current shell (bash, zsh, fish, pwsh).
// It reads $SHELL and takes the basename. If $SHELL is unset, it falls back to
// pwsh when available (Windows PowerShell), else bash.
func detectShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		if _, err := exec.LookPath("pwsh"); err == nil {
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

// isCommandNotFound checks if the error is due to a command not being found.
func isCommandNotFound(err error) bool {
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode() == 127
	}
	return false
}

// runFallbackShell starts the user's default shell.
func runFallbackShell(ctx *Context) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	cmd := exec.Command(shell)
	cmd.Dir = ctx.Dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}