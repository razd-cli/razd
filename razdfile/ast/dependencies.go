package ast

import (
	"fmt"
	"regexp"
	"strings"
)

// DependenciesConfig represents the unified dependencies section in Razdfile.
// It provides a simple, tool-agnostic way to specify dependencies that
// can be translated to mise or devbox format.
type DependenciesConfig struct {
	// Using specifies the package manager to use.
	// Valid values: "mise", "devbox"
	Using string `yaml:"using"`

	// Ensure is a list of tools to install.
	// Format: "tool@version" (e.g., "node@22", "php@8.4", "pnpm@latest")
	Ensure []string `yaml:"ensure,omitempty"`

	// Extra contains pass-through configuration for native managers.
	// Structure mirrors devbox.json / mise.toml - not validated, just passed through.
	Extra *DependenciesExtra `yaml:"extra,omitempty"`
}

// DependenciesExtra holds native configurations for package managers.
// These are passed through without validation and deep-merged with generated config.
type DependenciesExtra struct {
	// Devbox configuration - structure mirrors devbox.json
	Devbox map[string]any `yaml:"devbox,omitempty"`

	// Mise configuration - structure mirrors mise.toml
	Mise map[string]any `yaml:"mise,omitempty"`
}

// ParsedDependency represents a parsed ensure string.
type ParsedDependency struct {
	Tool    string // canonical tool name (e.g., "node")
	Version string // version string (e.g., "22", "latest")
	Raw     string // original string from ensure (e.g., "node@22")
}

// ValidUsing contains the allowed values for the Using field.
var ValidUsing = map[string]bool{
	"mise":   true,
	"devbox": true,
}

// ensureRegex validates the format of ensure strings.
// Format: tool@version where:
// - tool: lowercase letters, numbers, underscores, hyphens, dots, colons, slashes
//   (starts with letter). Colons/slashes support registry/plugin names such as
//   "vfox:dealenx/vfox-plugin-lux" or "cargo:ripgrep".
// - version: alphanumeric, dots, underscores, hyphens
// A bare name without a version is also valid (devbox/mise default to "latest").
var ensureRegex = regexp.MustCompile(`^[a-z][a-z0-9._:/_-]*@[a-zA-Z0-9._-]+$|^[a-z][a-z0-9._:/_-]*$`)

// ParseEnsure parses all ensure strings into structured ParsedDependency objects.
// Returns an error if any string has an invalid format.
func (d *DependenciesConfig) ParseEnsure() ([]ParsedDependency, error) {
	if d == nil {
		return nil, nil
	}

	result := make([]ParsedDependency, 0, len(d.Ensure))
	for _, raw := range d.Ensure {
		parsed, err := ParseDependencyString(raw)
		if err != nil {
			return nil, err
		}
		result = append(result, parsed)
	}
	return result, nil
}

// ParseDependencyString parses a single "tool@version" string.
func ParseDependencyString(s string) (ParsedDependency, error) {
	if s == "" {
		return ParsedDependency{}, &InvalidDependencyFormatError{Value: s}
	}

	if !ensureRegex.MatchString(s) {
		return ParsedDependency{}, &InvalidDependencyFormatError{Value: s}
	}

	// A bare name (no "@") is valid and has an empty version.
	if !strings.Contains(s, "@") {
		return ParsedDependency{
			Tool:    s,
			Version: "",
			Raw:     s,
		}, nil
	}

	parts := strings.SplitN(s, "@", 2)
	return ParsedDependency{
		Tool:    parts[0],
		Version: parts[1],
		Raw:     s,
	}, nil
}

// HasEnsure returns true if there are any ensure dependencies.
func (d *DependenciesConfig) HasEnsure() bool {
	return d != nil && len(d.Ensure) > 0
}

// HasExtra returns true if extra configuration is present.
func (d *DependenciesConfig) HasExtra() bool {
	return d != nil && d.Extra != nil
}

// GetProviderExtra returns the extra config for the current provider (using).
// Returns nil if no extra config exists for that provider.
func (d *DependenciesConfig) GetProviderExtra() map[string]any {
	if d == nil || d.Extra == nil {
		return nil
	}
	switch d.Using {
	case "mise":
		return d.Extra.Mise
	case "devbox":
		return d.Extra.Devbox
	default:
		return nil
	}
}

// InvalidDependencyFormatError is returned when an ensure string has invalid format.
type InvalidDependencyFormatError struct {
	Value string
}

func (e *InvalidDependencyFormatError) Error() string {
	if e.Value == "" {
		return "invalid dependency format: empty string, expected 'tool@version'"
	}
	return fmt.Sprintf("invalid dependency format '%s', expected 'tool@version' (e.g., 'node@22')", e.Value)
}
