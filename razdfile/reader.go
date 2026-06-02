package razdfile

import (
	"os"
	"path/filepath"

	"github.com/razd-cli/razd/razdfile/ast"
)

// DebugFunc is a function type for debug logging.
type DebugFunc func(format string, args ...any)

// Reader reads and parses Razdfile configurations.
type Reader struct {
	debug DebugFunc
	dir   string
}

// Option is a functional option for configuring the Reader.
type Option func(*Reader)

// WithDebugFunc sets the debug function for verbose logging.
func WithDebugFunc(fn DebugFunc) Option {
	return func(r *Reader) {
		r.debug = fn
	}
}

// WithDir sets the directory to search for Razdfile.
func WithDir(dir string) Option {
	return func(r *Reader) {
		r.dir = dir
	}
}

// NewReader creates a new Reader with the given options.
func NewReader(opts ...Option) *Reader {
	r := &Reader{
		debug: func(format string, args ...any) {},
		dir:   ".",
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Read reads and parses the Razdfile from the configured directory.
// It searches for default Razdfile names in order of priority.
func (r *Reader) Read() (*ast.Razdfile, error) {
	path, err := r.findRazdfile()
	if err != nil {
		return nil, err
	}

	r.debug("Reading Razdfile from: %s", path)

	node := NewFileNode(path)
	rf, err := node.Read()
	if err != nil {
		return nil, err
	}

	if err := Validate(rf, path); err != nil {
		return nil, err
	}

	return rf, nil
}

// ReadFile reads and parses a Razdfile from a specific file path.
func (r *Reader) ReadFile(path string) (*ast.Razdfile, error) {
	r.debug("Reading Razdfile from: %s", path)

	node := NewFileNode(path)
	rf, err := node.Read()
	if err != nil {
		return nil, err
	}

	if err := Validate(rf, path); err != nil {
		return nil, err
	}

	return rf, nil
}

// ReadNode reads and parses a Razdfile from a Node abstraction.
func (r *Reader) ReadNode(node Node) (*ast.Razdfile, error) {
	r.debug("Reading Razdfile from: %s", node.Location())

	rf, err := node.Read()
	if err != nil {
		return nil, err
	}

	if err := Validate(rf, node.Location()); err != nil {
		return nil, err
	}

	return rf, nil
}

// findRazdfile searches for a Razdfile in the configured directory.
// It uses os.ReadDir to list directory contents first (more reliable on
// Windows NTFS after git clone), then falls back to os.Stat for each
// candidate name.
func (r *Reader) findRazdfile() (string, error) {
	r.debug("Searching for Razdfile in directory: %s", r.dir)

	entries, readErr := os.ReadDir(r.dir)
	if readErr != nil {
		r.debug("Could not read directory %s: %v", r.dir, readErr)
		for _, name := range DefaultRazdfiles {
			path := filepath.Join(r.dir, name)
			if _, err := os.Stat(path); err == nil {
				r.debug("Found Razdfile: %s", path)
				return path, nil
			}
		}
		return "", ErrNotFound
	}

	fileNames := make(map[string]bool, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			fileNames[e.Name()] = true
		}
	}

	for _, name := range DefaultRazdfiles {
		if fileNames[name] {
			path := filepath.Join(r.dir, name)
			r.debug("Found Razdfile: %s", path)
			return path, nil
		}
		r.debug("Not found: %s (checked %d files in directory)", name, len(entries))
	}

	return "", ErrNotFound
}

// Exists checks if a Razdfile exists in the configured directory.
func (r *Reader) Exists() bool {
	_, err := r.findRazdfile()
	return err == nil
}
