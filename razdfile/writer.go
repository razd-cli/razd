package razdfile

import (
	"bytes"
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

// UpdateEnsureInFile updates the dependencies.ensure list in a Razdfile
// while preserving comments, key ordering, quoting style, and all other content.
// It reads the raw YAML as a Node tree, modifies only the ensure sequence,
// and writes back. Returns true if changes were made.
func UpdateEnsureInFile(filePath string, newEnsure []string) (bool, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to read Razdfile: %w", err)
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return false, fmt.Errorf("failed to parse Razdfile: %w", err)
	}

	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return false, fmt.Errorf("unexpected Razdfile structure: expected mapping node")
	}

	changed := false

	for i := 0; i < len(root.Content); i += 2 {
		key := root.Content[i]
		if key.Value != "dependencies" {
			continue
		}

		depNode := root.Content[i+1]
		if depNode.Kind != yaml.MappingNode {
			continue
		}

		ensureNode := (*yaml.Node)(nil)
		for j := 0; j < len(depNode.Content); j += 2 {
			depKey := depNode.Content[j]
			if depKey.Value != "ensure" {
				continue
			}
			ensureNode = depNode.Content[j+1]
			break
		}

		if ensureNode == nil {
			// The dependencies mapping exists but has no "ensure" key.
			// Append a new empty sequence so the ensure list is written.
			ensureNode = &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
			depNode.Content = append(depNode.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: "ensure", Tag: "!!str"},
				ensureNode,
			)
		}

		if ensureNode.Kind != yaml.SequenceNode {
			continue
		}

		currentValues := ensureListValues(ensureNode)
		if !ensureListsEqual(currentValues, newEnsure) {
			setEnsureList(ensureNode, newEnsure, currentValues)
			changed = true
		}
		break
	}

	if !changed {
		return false, nil
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(&doc); err != nil {
		return false, fmt.Errorf("failed to encode Razdfile: %w", err)
	}
	enc.Close()

	return true, os.WriteFile(filePath, buf.Bytes(), 0644)
}

// ensureListValues extracts string values from a sequence node.
func ensureListValues(seqNode *yaml.Node) []string {
	vals := make([]string, 0, len(seqNode.Content))
	for _, item := range seqNode.Content {
		if item.Kind == yaml.ScalarNode {
			vals = append(vals, item.Value)
		}
	}
	return vals
}

// ensureListsEqual compares two string slices.
func ensureListsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// setEnsureList replaces the sequence node's content with new values,
// preserving the quoting style of existing items where possible.
func setEnsureList(seqNode *yaml.Node, newEnsure []string, oldValues []string) {
	oldStyles := make(map[string]yaml.Style)
	for _, item := range seqNode.Content {
		if item.Kind == yaml.ScalarNode {
			oldStyles[item.Value] = item.Style
		}
	}

	newContent := make([]*yaml.Node, 0, len(newEnsure))
	for _, val := range newEnsure {
		style := yaml.Style(0)
		if s, ok := oldStyles[val]; ok {
			style = s
		} else if len(oldStyles) > 0 {
			for _, s := range oldStyles {
				style = s
				break
			}
		}
		newContent = append(newContent, &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: val,
			Style: style,
		})
	}
	seqNode.Content = newContent
}