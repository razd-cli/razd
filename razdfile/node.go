// Package razdfile provides types and utilities for parsing Razdfile.yml configuration files.
package razdfile

import "github.com/razd-cli/razd/razdfile/ast"

// Node is an abstraction over different sources of Razdfile configuration.
// It allows the Reader to work uniformly with files, remote sources, or in-memory data.
type Node interface {
	// Read parses the source and returns the Razdfile AST.
	Read() (*ast.Razdfile, error)

	// Location returns a human-readable location of this node (e.g., file path).
	Location() string
}
