package ast

import (
	"fmt"

	"go.yaml.in/yaml/v4"
)

// UnmarshalYAML implements custom unmarshaling for DevboxPackages
// to support both array and object formats
func (p *DevboxPackages) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.SequenceNode:
		// Array format: ["python@3.10", "poetry@latest"]
		var list []string
		if err := node.Decode(&list); err != nil {
			return err
		}
		p.List = list
		return nil

	case yaml.MappingNode:
		// Object format: {"python": {"version": "3.10"}}
		p.Map = make(map[string]*DevboxPackageConfig)
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			key := keyNode.Value

			switch valueNode.Kind {
			case yaml.ScalarNode:
				// Simple version string
				p.Map[key] = &DevboxPackageConfig{Version: valueNode.Value}
			case yaml.MappingNode:
				// Full config object
				var config DevboxPackageConfig
				if err := valueNode.Decode(&config); err != nil {
					return err
				}
				p.Map[key] = &config
			default:
				return fmt.Errorf("invalid package format at line %d", valueNode.Line)
			}
		}
		return nil
	}

	return fmt.Errorf("invalid packages format at line %d", node.Line)
}

// UnmarshalYAML implements custom unmarshaling for DevboxInitHook
// to support both string and array formats
func (h *DevboxInitHook) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		// Single command string
		h.Commands = []string{node.Value}
		return nil

	case yaml.SequenceNode:
		// Array of commands
		var commands []string
		if err := node.Decode(&commands); err != nil {
			return err
		}
		h.Commands = commands
		return nil
	}

	return fmt.Errorf("invalid init_hook format at line %d", node.Line)
}

// UnmarshalYAML implements custom unmarshaling for DevboxScript
// to support both string and array formats
func (s *DevboxScript) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		// Single command string
		s.Commands = []string{node.Value}
		return nil

	case yaml.SequenceNode:
		// Array of commands
		var commands []string
		if err := node.Decode(&commands); err != nil {
			return err
		}
		s.Commands = commands
		return nil
	}

	return fmt.Errorf("invalid script format at line %d", node.Line)
}
