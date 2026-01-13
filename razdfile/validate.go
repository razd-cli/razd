package razdfile

import "github.com/razd-cli/razd/razdfile/ast"

// Validate validates a Razdfile configuration.
// It checks version compatibility and required fields.
func Validate(rf *ast.Razdfile, location string) error {
	// Check version
	if rf.Version == "" {
		return &ErrValidation{
			Path:    location,
			Message: "version field is required",
		}
	}

	if !ast.IsVersionSupported(rf.Version) {
		return &ErrUnsupportedVersion{
			Path:    location,
			Version: rf.Version,
		}
	}

	// Validate dependencies section if present
	if err := validateDependencies(rf, location); err != nil {
		return err
	}

	// Check that there is some content
	if !rf.HasContent() {
		return &ErrValidation{
			Path:    location,
			Message: "Razdfile must have either tasks, mise, devbox, or dependencies configuration",
		}
	}

	return nil
}

// validateDependencies validates the dependencies section.
func validateDependencies(rf *ast.Razdfile, location string) error {
	if rf.Dependencies == nil {
		return nil
	}

	// Rule 1: Mutual exclusion - dependencies cannot be used with mise or devbox
	if rf.HasMise() {
		return &ErrValidation{
			Path:    location,
			Message: "cannot use 'dependencies' together with 'mise' section",
		}
	}
	if rf.HasDevbox() {
		return &ErrValidation{
			Path:    location,
			Message: "cannot use 'dependencies' together with 'devbox' section",
		}
	}

	// Rule 2: Using is required
	if rf.Dependencies.Using == "" {
		return &ErrValidation{
			Path:    location,
			Message: "dependencies.using is required (must be 'mise' or 'devbox')",
		}
	}

	// Rule 3: Valid using value
	if !ast.ValidUsing[rf.Dependencies.Using] {
		return &ErrValidation{
			Path:    location,
			Message: "dependencies.using must be 'mise' or 'devbox'",
		}
	}

	// Rule 4: Validate ensure format
	_, err := rf.Dependencies.ParseEnsure()
	if err != nil {
		return &ErrValidation{
			Path:    location,
			Message: err.Error(),
		}
	}

	return nil
}
