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
