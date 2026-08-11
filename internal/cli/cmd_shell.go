package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/razd-cli/razd/provisioner"
)

// runShell implements the "razd shell" command.
// It prints the shell activation script for the current shell so the user can
// evaluate it to activate the provisioner's environment in the current
// session: eval "$(razd shell)" (Unix) or iex "$(razd shell)" (PowerShell).
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
	ctx.Log.Debugf("Printing activation for shell=%s via %s\n", shell, prov.Name())
	return printShellActivation(ctx, prov, dir, shell)
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
// output (eval "$(razd shell)" / iex "$(razd shell)") to activate the current
// shell.
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
