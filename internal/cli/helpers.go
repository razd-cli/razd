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
		return nil, &errors.NoRazdfileError{Dir: dir}
	}

	log.Debugf("Found Razdfile: %s\n", rf.Location)
	return rf, nil
}

// getProvisioner determines the provisioner from the Razdfile and returns it.
// Returns an error if the Razdfile has no provisioner configured,
// or if the provisioner binary is not installed.
func getProvisioner(rf *ast.Razdfile, dir string, log *output.Logger) (provisioner.Provisioner, error) {
	var provName string
	switch {
	case rf.HasDependencies():
		provName = rf.Dependencies.Using
		log.Debugf("Using provisioner from dependencies: %s\n", provName)
	case rf.HasMise():
		provName = "mise"
		log.Debugf("Using mise provisioner from razdfile.mise section\n")
	case rf.HasDevbox():
		provName = "devbox"
		log.Debugf("Using devbox provisioner from razdfile.devbox section\n")
	default:
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
// the user to trust it if needed. Returns nil if the project is trusted.
func ensureTrusted(dir string, prov provisioner.Provisioner, log *output.Logger) error {
	trusted, err := trust.EnsureTrusted(dir, log, flags.Yes)
	if err != nil {
		return err
	}

	if !trusted && !flags.Yes {
		log.Infof("Run 'razd trust' to trust this project, or use --yes flag\n")
		return &errors.TrustError{
			Path:    dir,
			Message: "project not trusted",
		}
	}

	if !trusted && flags.Yes {
		log.Debugf("Auto-trusting project with --yes flag\n")
		if err := trust.Trust(dir, prov, log); err != nil {
			log.Warnf("Failed to trust project: %v\n", err)
		}
	}

	return nil
}