package cli

import (
	"fmt"
	"os"

	"github.com/razd-cli/razd/internal/errors"
	"github.com/razd-cli/razd/internal/flags"
	"github.com/razd-cli/razd/internal/output"
	"github.com/razd-cli/razd/internal/trust"
	"github.com/razd-cli/razd/provisioner"
	"github.com/razd-cli/razd/razdfile"
	"github.com/razd-cli/razd/razdfile/ast"
)

// resolveDir returns the working directory from the context or falls back to
// os.Getwd(). Returns an error if the directory cannot be determined.
func resolveDir(ctx *Context) (string, error) {
	if ctx.Dir != "" {
		ctx.Log.Debugf("Using provided directory: %s\n", ctx.Dir)
		return ctx.Dir, nil
	}
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}
	ctx.Log.Debugf("Using current directory: %s\n", dir)
	return dir, nil
}

// readRazdfile finds and parses the Razdfile in the given directory.
func readRazdfile(dir string, log *output.Logger) (*ast.Razdfile, error) {
	log.Debugf("Searching for Razdfile in: %s\n", dir)

	reader := razdfile.NewReader(
		razdfile.WithDir(dir),
		razdfile.WithDebugFunc(log.Debugf),
	)

	rf, err := reader.Read()
	if err != nil {
		entries, readErr := os.ReadDir(dir)
		if readErr == nil && len(entries) > 0 {
			log.Debugf("Directory %s contains %d entries:\n", dir, len(entries))
			for _, e := range entries {
				log.Debugf("  - %s (isDir=%v)\n", e.Name(), e.IsDir())
			}
		} else if readErr != nil {
			log.Debugf("Could not read directory %s: %v\n", dir, readErr)
		} else {
			log.Debugf("Directory %s is empty\n", dir)
		}
		return nil, &errors.NoRazdfileError{Dir: dir}
	}

	log.Debugf("Found Razdfile: %s\n", rf.Location)
	return rf, nil
}

// needsProvisioner returns true if the Razdfile declares a provisioner
// (dependencies, mise, or devbox section).
func needsProvisioner(rf *ast.Razdfile) bool {
	return rf.HasDependencies() || rf.HasMise() || rf.HasDevbox()
}

// provisionerName returns the provisioner name declared in the Razdfile,
// or an empty string if none is configured.
func provisionerName(rf *ast.Razdfile) string {
	switch {
	case rf.HasDependencies():
		return rf.Dependencies.Using
	case rf.HasMise():
		return "mise"
	case rf.HasDevbox():
		return "devbox"
	default:
		return ""
	}
}

// tryGetProvisioner attempts to resolve and return a provisioner from the
// Razdfile. Returns (provisioner, true) on success, or (nil, false) if the
// provisioner is not configured or not available on the system.
// This is the lazy approach: missing provisioner binaries are not fatal.
func tryGetProvisioner(rf *ast.Razdfile, dir string, log *output.Logger) (provisioner.Provisioner, bool) {
	provName := provisionerName(rf)
	if provName == "" {
		log.Debugf("No provisioner configured in Razdfile\n")
		return nil, false
	}

	log.Debugf("[FIX] Resolving provisioner: %s\n", provName)

	provConfig := provisioner.Config{
		Dir:     dir,
		Verbose: flags.Verbose,
		Silent:  flags.Silent,
	}

	prov, err := provisioner.Get(provName, provConfig)
	if err != nil {
		log.Debugf("[FIX] Provisioner %q not registered: %v\n", provName, err)
		return nil, false
	}

	if !prov.IsAvailable() {
		log.Warnf("[FIX] Provisioner %q is not installed, skipping provisioner setup\n", provName)
		log.Infof("Install %s to enable environment provisioning, or remove the %s section from Razdfile\n", provName, provName)
		return nil, false
	}

	log.Debugf("[FIX] Provisioner %q is available\n", provName)
	return prov, true
}

// getProvisioner determines the provisioner from the Razdfile and returns it.
// Returns an error if the Razdfile has no provisioner configured,
// or if the provisioner binary is not installed.
// This is the strict variant used by commands that require a provisioner (up, shell).
func getProvisioner(rf *ast.Razdfile, dir string, log *output.Logger) (provisioner.Provisioner, error) {
	provName := provisionerName(rf)
	if provName == "" {
		return nil, fmt.Errorf("no dependencies or provisioner configured in Razdfile")
	}

	provConfig := provisioner.Config{
		Dir:     dir,
		Verbose: flags.Verbose,
		Silent:  flags.Silent,
	}

	prov, err := provisioner.Get(provName, provConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get provisioner %q: %w", provName, err)
	}

	if !prov.IsAvailable() {
		return nil, &provisioner.ErrProvisionerNotAvailable{Name: provName}
	}

	log.Debugf("Provisioner %q is available\n", provName)
	return prov, nil
}

// ensureTrusted checks the trust status of the project directory and prompts
// the user to trust it interactively if needed. Returns nil if the project is trusted.
// When --yes is passed, auto-trusts. When stdin is not a TTY and --yes is not set,
// returns a TrustError requiring manual trust.
func ensureTrusted(dir string, prov provisioner.Provisioner, log *output.Logger) error {
	trusted, err := trust.EnsureTrusted(dir, log, flags.Yes)
	if err != nil {
		return err
	}

	if !trusted {
		if flags.Yes {
			log.Debugf("[FIX] Auto-trusting project with --yes flag\n")
			if err := trust.Trust(dir, prov, log); err != nil {
				log.Warnf("Failed to trust project: %v\n", err)
			}
			return nil
		}
		return &errors.TrustError{
			Path:    dir,
			Message: "project not trusted",
		}
	}

	return nil
}