package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/razd-cli/razd/internal/errors"
	"github.com/razd-cli/razd/internal/flags"
	"github.com/razd-cli/razd/internal/output"
	"github.com/razd-cli/razd/provisioner"
	"github.com/razd-cli/razd/razdfile/ast"
)

// shellCmd returns the appropriate shell command prefix for the current OS.
// On Windows it returns ["cmd", "/c"], on Unix ["sh", "-c"].
func shellCmd() []string {
	if runtime.GOOS == "windows" {
		return []string{"cmd", "/c"}
	}
	return []string{"sh", "-c"}
}

func runRun(ctx *Context) error {
	dir, err := resolveDir(ctx)
	if err != nil {
		return err
	}

	rf, err := readRazdfile(dir, ctx.Log)
	if err != nil {
		return err
	}

	if !rf.HasTasks() {
		return &errors.ConfigError{
			Message: "no tasks defined in Razdfile",
		}
	}

	var prov provisioner.Provisioner
	provResolved := false

	if rf.HasDependencies() || rf.HasMise() || rf.HasDevbox() {
		p, err := getProvisioner(rf, dir, ctx.Log)
		if err != nil {
			ctx.Log.Debugf("Provisioner resolution failed: %v, continuing without provisioner\n", err)
		} else {
			prov = p
			provResolved = true

			if err := ensureTrusted(dir, prov, ctx.Log); err != nil {
				return err
			}

			if !flags.NoInstall {
				ctx.Log.Infof("Installing dependencies via %s...\n", prov.Name())
				installCtx := context.Background()
				if err := prov.Install(installCtx); err != nil {
					ctx.Log.Warnf("Dependency installation failed: %v\n", err)
					ctx.Log.Infof("Continuing anyway — tasks may fail if dependencies are missing\n")
				} else {
					ctx.Log.Successf("Dependencies installed\n")
				}
			}
		}
	}

	taskNames := ctx.Args
	if len(taskNames) == 0 {
		taskNames = []string{"default"}
	}

	for _, taskName := range taskNames {
		ctx.Log.Debugf("Running task: %s\n", taskName)
		if err := executeTask(ctx, rf, prov, provResolved, dir, taskName); err != nil {
			return err
		}
	}

	return nil
}

// executeTask runs a single task by name.
func executeTask(ctx *Context, rf *ast.Razdfile, prov provisioner.Provisioner, provResolved bool, dir string, taskName string) error {
	task, ok := rf.Tasks.Get(taskName)
	if !ok || task == nil {
		return &errors.TaskNotFoundError{TaskName: taskName}
	}

	// Run dependencies first
	for _, dep := range task.Deps {
		if dep.Task != "" {
			ctx.Log.Debugf("Running dependency: %s\n", dep.Task)
			if err := executeTask(ctx, rf, prov, provResolved, dir, dep.Task); err != nil {
				return err
			}
		}
	}

	// Execute commands in the task
	for _, cmd := range task.Cmds {
		if cmd.Cmd != "" {
			var cmdArgs []string
			shellPrefix := shellCmd()
			if provResolved {
				cmdArgs = prov.RunCommand(append(shellPrefix, cmd.Cmd))
			} else {
				cmdArgs = append(shellPrefix, cmd.Cmd)
			}
			ctx.Log.Debugf("Running: %v\n", cmdArgs)
			if err := runCommand(cmdArgs, dir, ctx.Log); err != nil {
				return &errors.TaskRunError{
					TaskName: taskName,
					Err:      err,
				}
			}
		} else if cmd.Task != "" {
			if err := executeTask(ctx, rf, prov, provResolved, dir, cmd.Task); err != nil {
				return err
			}
		}
	}

	return nil
}

// runCommand executes a command and streams output.
func runCommand(cmd []string, dir string, log *output.Logger) error {
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

