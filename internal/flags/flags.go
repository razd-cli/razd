// Package flags defines global flags for razd CLI.
package flags

import (
	"errors"

	"github.com/spf13/pflag"
)

// Global flags
var (
	// Version shows razd version
	Version bool
	// Help shows help
	Help bool
	// Verbose enables verbose output
	Verbose bool
	// Silent disables output
	Silent bool
	// Dir is the working directory
	Dir string
	// Color enables colored output
	Color bool
	// Yes assumes yes to all prompts
	Yes bool
	// List shows all available tasks
	List bool
	// NoSync skips Razdfile <-> mise.toml sync
	NoSync bool

	// File path flags
	// TaskFile is the path to taskfile/razdfile
	TaskFile string
	// RazdFile is the path to razdfile (priority over taskfile)
	RazdFile string

	// Command-specific flags
	// Run flag for `razd up -r`
	Run bool
	// Force flag for overwriting
	Force bool
	// Using specifies the provisioner for `razd init`
	Using string
	// Migrate enables automatic migration
	Migrate bool

	// Trust command flags
	Untrust bool
	Show    bool
	All     bool
	Ignore  bool

	// NoInstall skips automatic dependency installation before running tasks
	NoInstall bool

	// Output format
	JSON bool
)

// Init registers all flags with pflag.
func Init() {
	// Global flags
	pflag.BoolVar(&Version, "version", false, "Show razd version")
	pflag.BoolVarP(&Version, "V", "V", false, "Show razd version")
	pflag.BoolVarP(&Help, "help", "h", false, "Show help")
	pflag.BoolVarP(&Verbose, "verbose", "v", false, "Enable verbose output")
	pflag.BoolVarP(&Silent, "silent", "s", false, "Disable output")
	pflag.StringVarP(&Dir, "dir", "d", "", "Working directory")
	pflag.BoolVar(&Color, "color", true, "Enable colored output")
	pflag.BoolVarP(&Yes, "yes", "y", false, "Assume yes to all prompts")
	pflag.BoolVar(&List, "list", false, "List all available tasks")
	pflag.BoolVar(&NoSync, "no-sync", false, "Skip Razdfile <-> mise.toml sync")

	// File path flags
	pflag.StringVarP(&TaskFile, "taskfile", "t", "", "Path to taskfile/razdfile")
	pflag.StringVar(&RazdFile, "razdfile", "", "Path to razdfile (priority over --taskfile)")

	// Command-specific flags
	pflag.BoolVarP(&Run, "run", "r", false, "Run default task after setup")
	pflag.BoolVarP(&Force, "force", "f", false, "Force overwrite existing files")
	pflag.StringVar(&Using, "using", "", "Package manager to use (mise or devbox)")
	pflag.BoolVar(&Migrate, "migrate", false, "Migrate from existing mise.toml/devbox.json")

	// Trust flags
	pflag.BoolVar(&Untrust, "untrust", false, "Remove trust from project")
	pflag.BoolVar(&Show, "show", false, "Show trust status")
	pflag.BoolVar(&All, "all", false, "Show all trusted projects")
	pflag.BoolVar(&Ignore, "ignore", false, "Ignore project (don't ask again)")

	// Output format
	pflag.BoolVar(&JSON, "json", false, "Output in JSON format")

	// Install control
	pflag.BoolVar(&NoInstall, "no-install", false, "Skip automatic dependency installation before running tasks")
}

// Validate checks for mutually exclusive flags.
func Validate() error {
	if Verbose && Silent {
		return errors.New("--verbose and --silent are mutually exclusive")
	}
	return nil
}

// Parse parses command line flags.
func Parse() {
	pflag.Parse()
}

// Args returns non-flag arguments.
func Args() []string {
	return pflag.Args()
}
