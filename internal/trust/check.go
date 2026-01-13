package trust

import (
	"context"
	"fmt"

	"github.com/razd-cli/razd/internal/output"
	"github.com/razd-cli/razd/provisioner"
)

// CheckResult represents the result of a trust check.
type CheckResult struct {
	Status      Status
	Path        string
	NeedsPrompt bool
}

// Check checks the trust status of a project path.
func Check(path string) (*CheckResult, error) {
	store, err := Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load trust store: %w", err)
	}

	status := store.GetStatus(path)
	return &CheckResult{
		Status:      status,
		Path:        path,
		NeedsPrompt: status == StatusUnknown,
	}, nil
}

// EnsureTrusted ensures a project is trusted, prompting if necessary.
// If autoTrust is true, the project will be automatically trusted.
// Returns true if trusted (or auto-trusted), false if not trusted.
func EnsureTrusted(path string, log *output.Logger, autoTrust bool) (bool, error) {
	result, err := Check(path)
	if err != nil {
		return false, err
	}

	switch result.Status {
	case StatusTrusted:
		return true, nil
	case StatusIgnored:
		log.Warnf("Project is ignored: %s\n", path)
		return false, nil
	case StatusUnknown:
		if autoTrust {
			// Auto-trust when --yes flag is passed
			return true, nil
		}
		// Prompt user (or just inform them)
		log.Warnf("Project not trusted: %s\n", path)
		log.Infof("Run 'razd trust' to trust this project\n")
		return false, nil
	default:
		return false, fmt.Errorf("unknown trust status: %s", result.Status)
	}
}

// Trust adds a project to the trusted list and syncs with provisioner if needed.
func Trust(path string, prov provisioner.Provisioner, log *output.Logger) error {
	store, err := Load()
	if err != nil {
		return fmt.Errorf("failed to load trust store: %w", err)
	}

	store.AddTrusted(path)
	if err := store.Save(); err != nil {
		return fmt.Errorf("failed to save trust store: %w", err)
	}

	log.Successf("Trusted project: %s", path)

	// Sync with provisioner (e.g., mise trust)
	if prov != nil {
		ctx := context.Background()
		if err := prov.Trust(ctx); err != nil {
			log.Warnf("Failed to sync trust with %s: %v", prov.Name(), err)
		} else {
			log.Debugf("Synced trust with %s", prov.Name())
		}
	}

	return nil
}

// Untrust removes a project from the trusted list.
func Untrust(path string, prov provisioner.Provisioner, log *output.Logger) error {
	store, err := Load()
	if err != nil {
		return fmt.Errorf("failed to load trust store: %w", err)
	}

	store.Remove(path)
	if err := store.Save(); err != nil {
		return fmt.Errorf("failed to save trust store: %w", err)
	}

	log.Successf("Untrusted project: %s", path)

	// Sync with provisioner
	if prov != nil {
		ctx := context.Background()
		if err := prov.Untrust(ctx); err != nil {
			log.Warnf("Failed to sync untrust with %s: %v", prov.Name(), err)
		} else {
			log.Debugf("Synced untrust with %s", prov.Name())
		}
	}

	return nil
}

// Ignore adds a project to the ignored list.
func Ignore(path string, log *output.Logger) error {
	store, err := Load()
	if err != nil {
		return fmt.Errorf("failed to load trust store: %w", err)
	}

	store.AddIgnored(path)
	if err := store.Save(); err != nil {
		return fmt.Errorf("failed to save trust store: %w", err)
	}

	log.Successf("Ignored project: %s", path)
	return nil
}
