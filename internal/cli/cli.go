// Package cli implements the razd command-line interface.
package cli

import (
	"fmt"
	"os"

	"github.com/razd-cli/razd/internal/errors"
	"github.com/razd-cli/razd/internal/flags"
	"github.com/razd-cli/razd/internal/output"
	"github.com/razd-cli/razd/internal/version"
)

// CLI represents the command-line interface.
type CLI struct {
	log      *output.Logger
	commands map[string]Command
}

// New creates a new CLI instance.
func New() *CLI {
	cli := &CLI{
		commands: make(map[string]Command),
	}
	cli.registerCommands()
	return cli
}

// Run executes the CLI with the given arguments.
func (c *CLI) Run(args []string) int {
	// Initialize flags
	flags.Init()

	// Parse flags
	flags.Parse()

	// Validate flags
	if err := flags.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return errors.CodeUnknown
	}

	// Initialize logger
	c.log = output.NewLogger(os.Stdout, os.Stderr)
	c.log.Verbose = flags.Verbose
	c.log.Silent = flags.Silent
	c.log.SetColor(flags.Color)

	// Handle --version flag
	if flags.Version {
		fmt.Println(version.Info())
		return errors.CodeOK
	}

	// Handle --help flag
	if flags.Help {
		c.printUsage()
		return errors.CodeOK
	}

	// Get command from remaining args
	cmdArgs := flags.Args()

	// Handle --list flag (list tasks)
	if flags.List {
		return c.runCommand("list", cmdArgs)
	}

	// No command specified - run default task
	if len(cmdArgs) == 0 {
		return c.runCommand("run", []string{})
	}

	// Check if first arg is a known command
	cmdName := cmdArgs[0]
	if _, exists := c.commands[cmdName]; exists {
		return c.runCommand(cmdName, cmdArgs[1:])
	}

	// Otherwise, treat as task name(s)
	return c.runCommand("run", cmdArgs)
}

// runCommand executes a registered command.
func (c *CLI) runCommand(name string, args []string) int {
	cmd, exists := c.commands[name]
	if !exists {
		c.log.Errf("Unknown command: %s", name)
		c.printUsage()
		return errors.CodeUnknown
	}

	ctx := &Context{
		Args: args,
		Log:  c.log,
		Dir:  flags.Dir,
	}

	if err := cmd.Run(ctx); err != nil {
		return handleError(err, c.log)
	}

	return errors.CodeOK
}

// handleError handles an error and returns appropriate exit code.
func handleError(err error, log *output.Logger) int {
	switch e := err.(type) {
	case *errors.TaskNotFoundError:
		log.Errf("Task not found: %s", e.TaskName)
		return errors.CodeTaskNotFound
	case *errors.TaskRunError:
		log.Errf("Task failed: %s - %v", e.TaskName, e.Err)
		return errors.CodeTaskFailed
	case *errors.ConfigError:
		log.Errf("Configuration error: %s", e.Message)
		return errors.CodeInvalidConfig
	case *errors.NoRazdfileError:
		log.Errf("No Razdfile found in %s", e.Dir)
		return errors.CodeNoRazdfile
	case *errors.TrustError:
		log.Errf("Trust error: %s", e.Message)
		return errors.CodeTrustError
	case *errors.GitNotInstalledError:
		log.Errf("git is not installed. Please install git to clone repositories.")
		return errors.CodeGitNotInstalled
	case *errors.CloneError:
		log.Errf("Clone error: %v", e)
		return errors.CodeCloneFailed
	default:
		log.Errf("Error: %v", err)
		return errors.CodeUnknown
	}
}

// printUsage prints the CLI usage information.
func (c *CLI) printUsage() {
	fmt.Println(`razd - Polyglot task runner with automatic environment provisioning

Usage:
  razd [flags] [task...]           Run task(s) (default task if none specified)
  razd <command> [flags] [args]    Run a command

Commands:
  up          Install tools and run the default task
  trust       Trust the current project
  list        List available tasks
  init        Initialize a new Razdfile
  add         Add a dependency or task (razd add task <name> -- <cmd>)
  shell       Open an interactive shell with the provisioner activated
  dev         Start development server (runs 'dev' task)
  build       Build the project (runs 'build' task)

Global Flags:
  -v, --verbose     Show detailed output
  -s, --silent      Suppress all output except errors
  -d, --dir         Change working directory
  -c, --color       Color output: auto, always, never (default: auto)
  -y, --yes         Answer yes to all prompts
  -l, --list        List available tasks
      --version     Show version information
  -h, --help        Show this help message

Task Flags:
  -t, --taskfile    Path to Taskfile.yml
      --razdfile    Path to Razdfile.yml
  -r, --run         Run the default task after setup (razd up)

Examples:
  razd                  Run the default task
  razd up               Install tools and run the default task
  razd up -r            Install tools and run the default task (same as above)
  razd build test       Run 'build' and 'test' tasks
  razd add node         Add a dependency to Razdfile
  razd add task hello -- echo 'hi'   Create a task
  razd trust            Trust the current project
  razd list             List all available tasks
  razd shell            Open interactive shell (provisioner activated)

Documentation: https://github.com/razd-cli/razd`)
}
