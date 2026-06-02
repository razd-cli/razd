package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/razd-cli/razd/internal/errors"
	"github.com/razd-cli/razd/internal/flags"
	"github.com/razd-cli/razd/internal/git"
	"github.com/razd-cli/razd/internal/output"
	"github.com/razd-cli/razd/provisioner"
	"github.com/razd-cli/razd/razdfile/ast"
)

// runUp executes the up command logic.
func runUp(ctx *Context) error {
	dir, err := resolveUpDir(ctx)
	if err != nil {
		return err
	}

	cloned := len(ctx.Args) > 0 && git.IsURL(ctx.Args[0])

	var rf *ast.Razdfile

	if cloned {
		rf, err = readRazdfileWithRetry(dir, ctx.Log)
	} else {
		rf, err = readRazdfile(dir, ctx.Log)
	}
	if err != nil {
		return err
	}

	if !needsProvisioner(rf) {
		return fmt.Errorf("no provisioner configured in Razdfile — add a dependencies, mise, or devbox section")
	}

	prov, ok := tryGetProvisioner(rf, dir, ctx.Log)
	if !ok {
		provName := provisionerName(rf)
		return fmt.Errorf("provisioner %q is not installed — install it or run 'razd run' to execute tasks without provisioning", provName)
	}

	if err := ensureTrusted(dir, prov, ctx.Log); err != nil {
		return err
	}

	if !flags.NoSync {
		if err := syncConfig(rf, prov, dir, ctx.Log); err != nil {
			ctx.Log.Warnf("Sync failed: %v\n", err)
		}
	}

	if err := generateProvisionerConfig(rf, prov, ctx.Log); err != nil {
		ctx.Log.Warnf("Failed to generate provisioner config: %v\n", err)
	}

	ctx.Log.Infof("Installing tools via %s...\n", prov.Name())
	bgCtx := context.Background()
	if err := prov.Install(bgCtx); err != nil {
		return fmt.Errorf("failed to install tools: %w", err)
	}
	ctx.Log.Successf("Tools installed successfully\n")

	if shouldRunAfterInstall() {
		ctx.Log.Infof("Running default task...\n")
		return runDefaultTask(ctx, rf, prov, dir)
	}

	return nil
}

// resolveUpDir determines the working directory for the up command.
// If a URL argument is provided, it clones the repository and returns
// the path to the cloned directory. Otherwise, falls back to resolveDir.
func resolveUpDir(ctx *Context) (string, error) {
	if len(ctx.Args) == 0 {
		return resolveDir(ctx)
	}

	if len(ctx.Args) > 1 {
		return "", fmt.Errorf("too many arguments for 'up' command")
	}

	arg := ctx.Args[0]

	if !git.IsURL(arg) {
		return "", fmt.Errorf("unexpected argument %q for 'up' command", arg)
	}

	if !git.IsGitAvailable() {
		return "", &errors.GitNotInstalledError{}
	}

	var cloneDir string
	if ctx.Dir != "" {
		cloneDir = ctx.Dir
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
		cloneDir = wd
	}

	ctx.Log.Infof("Cloning %s...\n", arg)

	if err := git.Clone(arg, cloneDir); err != nil {
		return "", &errors.CloneError{URL: arg, Err: err}
	}

	repoName, err := git.ExtractRepoName(arg)
	if err != nil {
		return "", fmt.Errorf("failed to determine repository name from %q: %w", arg, err)
	}

	dir := filepath.Join(cloneDir, repoName)
	ctx.Log.Successf("Cloned to %s\n", dir)

	return dir, nil
}

// shouldRunAfterInstall returns true if the --run flag was set (razd up --run).
func shouldRunAfterInstall() bool {
	return flags.Run
}

// runDefaultTask runs the default task from the Razdfile.
func runDefaultTask(ctx *Context, rf *ast.Razdfile, prov provisioner.Provisioner, dir string) error {
	if !rf.HasTasks() {
		ctx.Log.Warnf("No tasks defined in Razdfile\n")
		return nil
	}

	defaultTask, ok := rf.Tasks.Get("default")
	if !ok || defaultTask == nil {
		ctx.Log.Warnf("No 'default' task defined in Razdfile\n")
		return nil
	}

	ctx.Log.Debugf("Running default task\n")

	for _, cmd := range defaultTask.Cmds {
		if cmd.Cmd != "" {
			wrappedCmd := prov.RunCommand(append(shellCmd(), cmd.Cmd))
			ctx.Log.Debugf("Running: %v\n", wrappedCmd)
			if err := executeCommand(wrappedCmd, dir, ctx.Log); err != nil {
				return err
			}
		} else if cmd.Task != "" {
			ctx.Log.Debugf("Running dependent task: %s\n", cmd.Task)
			if err := runTask(ctx, rf, prov, dir, cmd.Task); err != nil {
				return err
			}
		}
	}

	return nil
}

// runTask runs a specific task from the Razdfile.
func runTask(ctx *Context, rf *ast.Razdfile, prov provisioner.Provisioner, dir string, taskName string) error {
	task, ok := rf.Tasks.Get(taskName)
	if !ok || task == nil {
		return fmt.Errorf("task %q not found", taskName)
	}

	for _, dep := range task.Deps {
		if dep.Task != "" {
			ctx.Log.Debugf("Running dependency: %s\n", dep.Task)
			if err := runTask(ctx, rf, prov, dir, dep.Task); err != nil {
				return err
			}
		}
	}

	for _, cmd := range task.Cmds {
		if cmd.Cmd != "" {
			wrappedCmd := prov.RunCommand(append(shellCmd(), cmd.Cmd))
			ctx.Log.Debugf("Running: %v\n", wrappedCmd)
			if err := executeCommand(wrappedCmd, dir, ctx.Log); err != nil {
				return err
			}
		} else if cmd.Task != "" {
			if err := runTask(ctx, rf, prov, dir, cmd.Task); err != nil {
				return err
			}
		}
	}

	return nil
}

// executeCommand runs a command and streams output.
func executeCommand(cmd []string, dir string, log *output.Logger) error {
	if len(cmd) == 0 {
		return fmt.Errorf("empty command")
	}

	c := exec.Command(cmd[0], cmd[1:]...)
	c.Dir = dir
	c.Stdout = log.Stdout
	c.Stderr = log.Stderr
	c.Stdin = os.Stdin

	return c.Run()
}