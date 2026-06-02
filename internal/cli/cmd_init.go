package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v4"
	"github.com/razd-cli/razd/internal/flags"
	"github.com/razd-cli/razd/razdfile"
	"github.com/razd-cli/razd/razdfile/ast"
)

// runInit implements the "razd init" command.
// It creates a new Razdfile.yml in the current directory.
func runInit(ctx *Context) error {
	dir, err := resolveDir(ctx)
	if err != nil {
		return err
	}

	// Check if Razdfile already exists
	reader := razdfile.NewReader(razdfile.WithDir(dir))
	if reader.Exists() && !flags.Force {
		return fmt.Errorf("Razdfile already exists in %s. Use --force to overwrite", dir)
	}

	// Determine provisioner
	using := flags.Using
	if using == "" {
		using = detectProvider(dir)
		ctx.Log.Debugf("Auto-detected provider: %s\n", using)
	}

	if using != "mise" && using != "devbox" {
		return fmt.Errorf("invalid provider %q: must be 'mise' or 'devbox'", using)
	}

	// Build Razdfile content
	rf := buildInitRazdfile(using)

	data, err := yaml.Marshal(rf)
	if err != nil {
		return fmt.Errorf("failed to marshal Razdfile: %w", err)
	}

	targetPath := filepath.Join(dir, "Razdfile.yml")
	if err := os.WriteFile(targetPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write Razdfile: %w", err)
	}

	ctx.Log.Successf("Created %s\n", targetPath)
	return nil
}

// detectProvider detects the appropriate provider based on existing config files.
func detectProvider(dir string) string {
	miseToml := filepath.Join(dir, "mise.toml")
	devboxJSON := filepath.Join(dir, "devbox.json")

	if _, err := os.Stat(miseToml); err == nil {
		return "mise"
	}
	if _, err := os.Stat(devboxJSON); err == nil {
		return "devbox"
	}

	return "mise"
}

// buildInitRazdfile creates a default Razdfile structure.
func buildInitRazdfile(using string) *ast.Razdfile {
	rf := &ast.Razdfile{
		Version: "1",
		Dependencies: &ast.DependenciesConfig{
			Using:  using,
			Ensure: []string{},
		},
	}

	// Add a minimal default task
	return rf
}