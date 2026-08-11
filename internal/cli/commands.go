package cli

import (
	"github.com/razd-cli/razd/internal/output"
)

// Context provides context for command execution.
type Context struct {
	Args []string
	Log  *output.Logger
	Dir  string
}

// Command represents a CLI command.
type Command interface {
	// Name returns the command name.
	Name() string
	// Description returns a short description of the command.
	Description() string
	// Run executes the command.
	Run(ctx *Context) error
}

// registerCommands registers all available commands.
func (c *CLI) registerCommands() {
	commands := []Command{
		&VersionCommand{},
		&RunCommand{},
		&UpCommand{},
		&ListCommand{},
		&TrustCommand{},
		&InitCommand{},
		&AddCommand{},
		&ShellCommand{},
		&DevCommand{},
		&BuildCommand{},
	}

	for _, cmd := range commands {
		c.commands[cmd.Name()] = cmd
	}
}

// VersionCommand shows version information.
type VersionCommand struct{}

func (c *VersionCommand) Name() string        { return "version" }
func (c *VersionCommand) Description() string { return "Show version information" }
func (c *VersionCommand) Run(ctx *Context) error {
	// Handled by global --version flag
	return nil
}

// RunCommand executes tasks.
type RunCommand struct{}

func (c *RunCommand) Name() string        { return "run" }
func (c *RunCommand) Description() string { return "Run one or more tasks" }
func (c *RunCommand) Run(ctx *Context) error {
	return runRun(ctx)
}

// UpCommand installs tools and dependencies.
type UpCommand struct{}

func (c *UpCommand) Name() string        { return "up" }
func (c *UpCommand) Description() string { return "Install tools and dependencies via mise/devbox" }
func (c *UpCommand) Run(ctx *Context) error {
	return runUp(ctx)
}

// ListCommand lists available tasks.
type ListCommand struct{}

func (c *ListCommand) Name() string        { return "list" }
func (c *ListCommand) Description() string { return "List available tasks" }
func (c *ListCommand) Run(ctx *Context) error {
	return runList(ctx)
}

// TrustCommand manages project trust.
type TrustCommand struct{}

func (c *TrustCommand) Name() string        { return "trust" }
func (c *TrustCommand) Description() string { return "Trust the current project" }
func (c *TrustCommand) Run(ctx *Context) error {
	return runTrust(ctx)
}

// InitCommand initializes a new Razdfile.
type InitCommand struct{}

func (c *InitCommand) Name() string        { return "init" }
func (c *InitCommand) Description() string { return "Initialize a new Razdfile" }
func (c *InitCommand) Run(ctx *Context) error {
	return runInit(ctx)
}

// AddCommand adds a dependency to Razdfile.
type AddCommand struct{}

func (c *AddCommand) Name() string        { return "add" }
func (c *AddCommand) Description() string { return "Add a dependency to Razdfile" }
func (c *AddCommand) Run(ctx *Context) error {
	return runAdd(ctx)
}

// ShellCommand starts an interactive shell.
type ShellCommand struct{}

func (c *ShellCommand) Name() string        { return "shell" }
func (c *ShellCommand) Description() string { return "Open an interactive shell with the provisioner activated" }
func (c *ShellCommand) Run(ctx *Context) error {
	return runShell(ctx)
}

// DevCommand runs the dev task.
type DevCommand struct{}

func (c *DevCommand) Name() string        { return "dev" }
func (c *DevCommand) Description() string { return "Start development server (runs 'dev' task)" }
func (c *DevCommand) Run(ctx *Context) error {
	ctx.Log.Debugf("Running 'dev' command")
	// Delegate to run command with "dev" task
	runCmd := &RunCommand{}
	ctx.Args = append([]string{"dev"}, ctx.Args...)
	return runCmd.Run(ctx)
}

// BuildCommand runs the build task.
type BuildCommand struct{}

func (c *BuildCommand) Name() string        { return "build" }
func (c *BuildCommand) Description() string { return "Build the project (runs 'build' task)" }
func (c *BuildCommand) Run(ctx *Context) error {
	ctx.Log.Debugf("Running 'build' command")
	// Delegate to run command with "build" task
	runCmd := &RunCommand{}
	ctx.Args = append([]string{"build"}, ctx.Args...)
	return runCmd.Run(ctx)
}
