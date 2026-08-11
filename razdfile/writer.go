package razdfile

import (
	"bytes"
	"fmt"
	"os"

	taskast "github.com/go-task/task/v3/taskfile/ast"
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

// UpdateTasksInFile adds or updates a task in the tasks: section of a Razdfile
// while preserving comments, key ordering, quoting style, and all other content.
// It reads the raw YAML as a Node tree, modifies only the target task, and
// writes back. Returns true if changes were made.
//
// The task is serialized in the compact scalar form (`name: <cmd>`) when it has
// exactly one command and no extra fields; otherwise the mapping form
// (`name: {cmds: [...], desc: ...}`) is used.
func UpdateTasksInFile(filePath string, taskName string, task *taskast.Task) (bool, error) {
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

	// Locate the tasks: mapping, creating it under root if absent.
	tasksNode := (*yaml.Node)(nil)
	for i := 0; i < len(root.Content); i += 2 {
		key := root.Content[i]
		if key.Value != "tasks" {
			continue
		}
		tasksNode = root.Content[i+1]
		break
	}
	if tasksNode == nil {
		tasksNode = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		root.Content = append(root.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "tasks", Tag: "!!str"},
			tasksNode,
		)
	}
	if tasksNode.Kind != yaml.MappingNode {
		return false, fmt.Errorf("unexpected tasks structure: expected mapping node")
	}

	// Replace or append the task key.
	taskValue := taskToYAMLNode(task)
	replaced := false
	for i := 0; i < len(tasksNode.Content); i += 2 {
		key := tasksNode.Content[i]
		if key.Value != taskName {
			continue
		}
		tasksNode.Content[i+1] = taskValue
		replaced = true
		break
	}
	if !replaced {
		tasksNode.Content = append(tasksNode.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: taskName, Tag: "!!str"},
			taskValue,
		)
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

// taskToYAMLNode serializes a task into a YAML node. A single command is
// written as `cmd: <command>`; multiple commands use `cmds: [...]`. Extra
// fields (desc, deps, dir, silent, interactive) are always written in the
// mapping form.
func taskToYAMLNode(task *taskast.Task) *yaml.Node {
	// Mapping form.
	mapping := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}

	// Single command -> cmd: <command>; multiple -> cmds: [...]
	if len(task.Cmds) == 1 && task.Cmds[0] != nil && task.Cmds[0].Cmd != "" {
		mapping.Content = append(mapping.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "cmd", Tag: "!!str"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: task.Cmds[0].Cmd, Tag: "!!str"},
		)
	} else if len(task.Cmds) > 0 {
		cmds := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		for _, c := range task.Cmds {
			if c != nil && c.Cmd != "" {
				cmds.Content = append(cmds.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: c.Cmd, Tag: "!!str"})
			}
		}
		if len(cmds.Content) > 0 {
			mapping.Content = append(mapping.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: "cmds", Tag: "!!str"},
				cmds,
			)
		}
	}

	if task.Desc != "" {
		mapping.Content = append(mapping.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "desc", Tag: "!!str"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: task.Desc, Tag: "!!str"},
		)
	}
	if len(task.Deps) > 0 {
		deps := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		for _, d := range task.Deps {
			if d != nil && d.Task != "" {
				deps.Content = append(deps.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: d.Task, Tag: "!!str"})
			}
		}
		if len(deps.Content) > 0 {
			mapping.Content = append(mapping.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: "deps", Tag: "!!str"},
				deps,
			)
		}
	}
	if task.Dir != "" {
		mapping.Content = append(mapping.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "dir", Tag: "!!str"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: task.Dir, Tag: "!!str"},
		)
	}
	if task.Silent {
		mapping.Content = append(mapping.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "silent", Tag: "!!bool"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: "true", Tag: "!!bool"},
		)
	}
	if task.Interactive {
		mapping.Content = append(mapping.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "interactive", Tag: "!!bool"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: "true", Tag: "!!bool"},
		)
	}

	return mapping
}