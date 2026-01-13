package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/razd-cli/razd/internal/flags"
	"github.com/razd-cli/razd/internal/output"
	"github.com/razd-cli/razd/internal/trust"
	"github.com/razd-cli/razd/provisioner"
	"github.com/razd-cli/razd/razdfile"
	"github.com/razd-cli/razd/razdfile/ast"
)

// runUp executes the up command logic.
func runUp(ctx *Context) error {
	// Determine working directory
	dir := ctx.Dir
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	ctx.Log.Debugf("Working directory: %s\n", dir)

	// Find and parse Razdfile
	reader := razdfile.NewReader(
		razdfile.WithDir(dir),
		razdfile.WithDebugFunc(ctx.Log.Debugf),
	)

	rf, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read Razdfile: %w", err)
	}

	ctx.Log.Debugf("Found Razdfile\n")

	// Determine provisioner
	var provName string
	if rf.HasDependencies() {
		provName = rf.Dependencies.Using
	} else if rf.HasMise() {
		provName = "mise"
	} else if rf.HasDevbox() {
		provName = "devbox"
	} else {
		ctx.Log.Warnf("No dependencies or provisioner configured in Razdfile\n")
		return nil
	}

	ctx.Log.Debugf("Using provisioner: %s\n", provName)

	// Get provisioner
	provConfig := provisioner.Config{
		Dir:     dir,
		Verbose: flags.Verbose,
		Silent:  flags.Silent,
	}

	prov, err := provisioner.Get(provName, provConfig)
	if err != nil {
		return fmt.Errorf("failed to get provisioner %q: %w", provName, err)
	}

	// Check if provisioner is available
	if !prov.IsAvailable() {
		return fmt.Errorf("provisioner %q is not installed on this system", provName)
	}

	// Check trust
	trusted, err := trust.EnsureTrusted(dir, ctx.Log, flags.Yes)
	if err != nil {
		return err
	}
	if !trusted && !flags.Yes {
		ctx.Log.Infof("Run 'razd trust' to trust this project, or use --yes flag\n")
		return nil
	}

	// If --yes flag is set and not trusted, auto-trust
	if !trusted && flags.Yes {
		if err := trust.Trust(dir, prov, ctx.Log); err != nil {
			ctx.Log.Warnf("Failed to trust project: %v\n", err)
		}
	}

	// Install tools
	ctx.Log.Infof("Installing tools via %s...\n", provName)
	bgCtx := context.Background()
	if err := prov.Install(bgCtx); err != nil {
		return fmt.Errorf("failed to install tools: %w", err)
	}
	ctx.Log.Successf("Tools installed successfully\n")

	// If --run flag is set, run default task
	if flags.Run {
		ctx.Log.Infof("Running default task...\n")
		return runDefaultTask(ctx, rf, prov, dir)
	}

	return nil
}

// runDefaultTask runs the default task from the Razdfile.
func runDefaultTask(ctx *Context, rf *ast.Razdfile, prov provisioner.Provisioner, dir string) error {
	// Check if tasks are defined
	if !rf.HasTasks() {
		ctx.Log.Warnf("No tasks defined in Razdfile\n")
		return nil
	}

	// Check for default task
	defaultTask, ok := rf.Tasks.Get("default")
	if !ok || defaultTask == nil {
		ctx.Log.Warnf("No 'default' task defined in Razdfile\n")
		return nil
	}

	ctx.Log.Debugf("Running default task\n")

	// Execute commands from the task directly
	for _, cmd := range defaultTask.Cmds {
		if cmd.Cmd != "" {
			// It's a direct command
			wrappedCmd := prov.RunCommand([]string{"sh", "-c", cmd.Cmd})
			ctx.Log.Debugf("Running: %v\n", wrappedCmd)
			if err := executeCommand(wrappedCmd, dir, ctx.Log); err != nil {
				return err
			}
		} else if cmd.Task != "" {
			// It's a task reference - run that task
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

	// Run dependencies first
	for _, dep := range task.Deps {
		if dep.Task != "" {
			ctx.Log.Debugf("Running dependency: %s\n", dep.Task)
			if err := runTask(ctx, rf, prov, dir, dep.Task); err != nil {
				return err
			}
		}
	}

	// Execute commands
	for _, cmd := range task.Cmds {
		if cmd.Cmd != "" {
			wrappedCmd := prov.RunCommand([]string{"sh", "-c", cmd.Cmd})
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
