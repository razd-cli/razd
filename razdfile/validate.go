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

	// Check that there is some content
	if !rf.HasContent() {
		return &ErrValidation{
			Path:    location,
			Message: "Razdfile must have either tasks or mise configuration",
		}
	}

	return nil
}
