// Package provisioner provides an abstraction layer for package managers.
// It implements the Strategy pattern to support multiple package managers
// (mise, devbox, etc.) through a unified interface.
package provisioner

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/razd-cli/razd/razdfile/ast"
)

// Provisioner defines the interface for package manager integrations.
// Each supported package manager (mise, devbox, etc.) implements this interface.
type Provisioner interface {
	// Name returns the provisioner identifier (mise, devbox, ...)
	Name() string

	// GenerateConfig creates the native config file from parsed dependencies.
	// packages: parsed ensure list from Razdfile
	// extra: pass-through config from dependencies.extra section
	GenerateConfig(packages []ast.ParsedDependency, extra map[string]any) error

	// ReadConfig reads the native config file and returns tool→version mapping.
	// Returns nil map if config file doesn't exist.
	ReadConfig() (map[string]string, error)

	// WriteTools merges the given tool→version map into the native config file,
	// preserving all sections and tools not present in the input.
	WriteTools(tools map[string]string) error

	// AddTools delegates tool installation to the native package manager
	// (e.g. "devbox add", "mise use"). It writes the packages into the native
	// config and installs them, idempotently: calling it again with an already
	// present package is a no-op, not an error. tools maps tool name to version
	// (empty version means "latest"). Success means the tools are present in
	// the native config after the call. It is only called when the native
	// config file already exists.
	AddTools(ctx context.Context, tools map[string]string) error

	// RemoveTools delegates tool removal to the native package manager
	// (e.g. "devbox rm", "mise unuse"). It removes the packages from the native
	// config, idempotently: removing a package that is not present is a no-op,
	// not an error. tools maps tool name to version (empty version means any).
	// Success means the tools are absent from the native config after the call.
	// It is only called when the native config file already exists.
	RemoveTools(ctx context.Context, tools map[string]string) error

	// Install runs the package installation command.
	// This executes the native install command (e.g., "mise install", "devbox install")
	Install(ctx context.Context) error

	// RunCommand wraps a command to run in the provisioner's environment.
	// Returns the wrapped command slice, e.g., ["mise", "exec", "--", ...cmd]
	RunCommand(cmd []string) []string

	// Shell starts an interactive shell with the provisioner's environment.
	Shell(ctx context.Context) error

	// Trust marks the current directory as trusted for the provisioner.
	Trust(ctx context.Context) error

	// Untrust removes trust from the current directory.
	Untrust(ctx context.Context) error

	// IsAvailable checks if the provisioner binary is installed on the system.
	IsAvailable() bool
}

// Config holds common provisioner configuration.
type Config struct {
	// Dir is the working directory for the provisioner
	Dir string
	// Verbose enables verbose output
	Verbose bool
	// Silent disables output
	Silent bool
}

// BaseProvisioner provides common functionality for provisioners.
type BaseProvisioner struct {
	Config Config
}

// checkBinary verifies that a binary is available in PATH.
func checkBinary(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// toolArg formats a tool name+version as a native package argument:
// "name@version" when a version is present, otherwise the bare name (the
// native manager resolves a bare name to "latest"). Never emits a trailing "@".
func toolArg(name, version string) string {
	if version == "" {
		return name
	}
	return name + "@" + version
}

// ErrProvisionerNotFound is returned when a provisioner is not in the registry.
type ErrProvisionerNotFound struct {
	Name      string
	Supported []string
}

func (e *ErrProvisionerNotFound) Error() string {
	return fmt.Sprintf("unknown provisioner: %s, supported: %v", e.Name, e.Supported)
}

// ErrProvisionerNotAvailable is returned when a provisioner binary is not installed.
type ErrProvisionerNotAvailable struct {
	Name string
}

func (e *ErrProvisionerNotAvailable) Error() string {
	return fmt.Sprintf("%s is not installed, please install it first", e.Name)
}
