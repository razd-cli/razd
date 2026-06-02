package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/razd-cli/razd/internal/errors"
	"github.com/razd-cli/razd/internal/output"
	"github.com/razd-cli/razd/provisioner"
	"github.com/razd-cli/razd/razdfile/ast"
)

// runRun implements the "razd run" and "razd <task>" commands.
// It reads a Razdfile, resolves the provisioner, checks trust,
// and executes the specified task(s).
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
		}
	}

	// Determine which task(s) to run
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
			if provResolved {
				cmdArgs = prov.RunCommand([]string{"sh", "-c", cmd.Cmd})
			} else {
				cmdArgs = []string{"sh", "-c", cmd.Cmd}
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

// runDefaultTaskFromRun runs the default task from the Razdfile.
// This is the version used by the "run" command.
func runDefaultTaskFromRun(ctx *Context, rf *ast.Razdfile, prov provisioner.Provisioner, dir string) error {
	defaultTask, ok := rf.Tasks.Get("default")
	if !ok || defaultTask == nil {
		ctx.Log.Warnf("No 'default' task defined in Razdfile\n")
		return nil
	}

	return executeTask(ctx, rf, prov, true, dir, "default")
}