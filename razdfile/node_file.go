package razdfile

import (
	"os"

	"github.com/razd-cli/razd/razdfile/ast"
	"go.yaml.in/yaml/v4"
)

// DefaultRazdfiles lists the default file names to search for in order of priority.
var DefaultRazdfiles = []string{
	"Razdfile.yml",
	"Razdfile.yaml",
	"razdfile.yml",
	"razdfile.yaml",
}

// FileNode represents a Razdfile read from the filesystem.
type FileNode struct {
	path string
}

// NewFileNode creates a new FileNode for the given file path.
func NewFileNode(path string) *FileNode {
	return &FileNode{path: path}
}

// Read reads and parses the Razdfile from the filesystem.
func (n *FileNode) Read() (*ast.Razdfile, error) {
	data, err := os.ReadFile(n.path)
	if err != nil {
		return nil, &ErrReadFile{Path: n.path, Err: err}
	}

	var rf ast.Razdfile
	if err := yaml.Unmarshal(data, &rf); err != nil {
		return nil, &ErrParseFile{Path: n.path, Err: err}
	}

	return &rf, nil
}

// Location returns the file path of this node.
func (n *FileNode) Location() string {
	return n.path
}

// Ensure FileNode implements Node interface.
var _ Node = (*FileNode)(nil)
