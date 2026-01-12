package ast

import (
	"fmt"

	"go.yaml.in/yaml/v4"
)

// MiseConfig represents the mise configuration section in Razdfile
type MiseConfig struct {
	Tools      map[string]*MiseTool `yaml:"tools,omitempty"`
	Env        map[string]any       `yaml:"env,omitempty"`
	Settings   *MiseSettings        `yaml:"settings,omitempty"`
	Hooks      map[string]any       `yaml:"hooks,omitempty"`
	Plugins    map[string]string    `yaml:"plugins,omitempty"`
	MinVersion string               `yaml:"min_version,omitempty"`
}

// MiseTool represents a tool version specification
// Supports both simple string format ("22") and complex object format
type MiseTool struct {
	Version     string         `yaml:"version,omitempty"`
	OS          []string       `yaml:"os,omitempty"`
	InstallEnv  map[string]any `yaml:"install_env,omitempty"`
	Postinstall string         `yaml:"postinstall,omitempty"`
}

// UnmarshalYAML implements custom unmarshaling for MiseTool
// to support both simple string and complex object formats
func (t *MiseTool) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		// Simple version string: node: "22"
		t.Version = node.Value
		return nil

	case yaml.MappingNode:
		// Complex object format
		type toolAlias MiseTool
		return node.Decode((*toolAlias)(t))

	case yaml.SequenceNode:
		// Multiple versions: node: ["20", "22"]
		var versions []string
		if err := node.Decode(&versions); err != nil {
			return err
		}
		if len(versions) > 0 {
			t.Version = versions[0] // primary version
		}
		return nil
	}

	return fmt.Errorf("invalid tool format at line %d, column %d", node.Line, node.Column)
}

// MiseSettings represents mise settings configuration
type MiseSettings struct {
	AutoInstall  *bool `yaml:"auto_install,omitempty"`
	Experimental *bool `yaml:"experimental,omitempty"`
	Quiet        *bool `yaml:"quiet,omitempty"`
	Verbose      *bool `yaml:"verbose,omitempty"`
	Jobs         *int  `yaml:"jobs,omitempty"`
}

// MiseEnvSpecial represents special environment directives
type MiseEnvSpecial struct {
	File   any      `yaml:"file,omitempty"`   // string or []string
	Path   any      `yaml:"path,omitempty"`   // string or []string
	Source any      `yaml:"source,omitempty"` // string or []string
	Python *MiseEnvPython `yaml:"python,omitempty"`
}

// MiseEnvPython represents Python-specific environment settings
type MiseEnvPython struct {
	Venv any `yaml:"venv,omitempty"` // string or object with path, create, python
}

// HasTools returns true if mise config has any tools defined
func (m *MiseConfig) HasTools() bool {
	return m != nil && len(m.Tools) > 0
}

// HasEnv returns true if mise config has any environment variables
func (m *MiseConfig) HasEnv() bool {
	return m != nil && len(m.Env) > 0
}
