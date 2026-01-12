package razdfile

import (
	"errors"
	"fmt"
)

// ErrNotFound is returned when no Razdfile could be found.
var ErrNotFound = errors.New("no Razdfile found")

// ErrReadFile is returned when a file cannot be read.
type ErrReadFile struct {
	Path string
	Err  error
}

func (e *ErrReadFile) Error() string {
	return fmt.Sprintf("failed to read file %q: %v", e.Path, e.Err)
}

func (e *ErrReadFile) Unwrap() error {
	return e.Err
}

// ErrParseFile is returned when a file cannot be parsed as YAML.
type ErrParseFile struct {
	Path string
	Err  error
}

func (e *ErrParseFile) Error() string {
	return fmt.Sprintf("failed to parse file %q: %v", e.Path, e.Err)
}

func (e *ErrParseFile) Unwrap() error {
	return e.Err
}

// ErrUnsupportedVersion is returned when the Razdfile version is not supported.
type ErrUnsupportedVersion struct {
	Path    string
	Version string
}

func (e *ErrUnsupportedVersion) Error() string {
	return fmt.Sprintf("unsupported version %q in file %q", e.Version, e.Path)
}

// ErrValidation is returned when the Razdfile fails validation.
type ErrValidation struct {
	Path    string
	Message string
}

func (e *ErrValidation) Error() string {
	return fmt.Sprintf("validation error in file %q: %s", e.Path, e.Message)
}
