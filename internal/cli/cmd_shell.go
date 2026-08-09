package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/razd-cli/razd/internal/errors"
	"github.com/razd-cli/razd/internal/flags"
)

// runShell implements the "razd shell" command.
// It starts an interactive shell with the provisioned environment.
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
			ctx.Log.Infof("No provisioner configured, starting system shell\n")
			return runFallbackShell(ctx)
		}
		provName := provisionerName(rf)
		return fmt.Errorf("provisioner %q is not installed — install it to use 'razd shell', or use 'razd run' to execute tasks without provisioning", provName)
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