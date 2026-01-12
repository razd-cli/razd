package ast

import (
	taskast "github.com/go-task/task/v3/taskfile/ast"
)

// Razdfile represents the root structure of a Razdfile.yml
type Razdfile struct {
	// Metadata (set by Reader, not from YAML)
	Location string `yaml:"-"`

	// Required field
	Version string `yaml:"version"`

	// Mise configuration (razd-specific)
	Mise *MiseConfig `yaml:"mise,omitempty"`

	// Taskfile-compatible fields (reuse go-task types)
	Output   taskast.Output   `yaml:"output,omitempty"`
	Method   string           `yaml:"method,omitempty"`
	Includes *taskast.Includes `yaml:"includes,omitempty"`
	Vars     *taskast.Vars    `yaml:"vars,omitempty"`
	Env      *taskast.Vars    `yaml:"env,omitempty"`
	Tasks    *taskast.Tasks   `yaml:"tasks,omitempty"`
	Silent   bool             `yaml:"silent,omitempty"`
	Dotenv   []string         `yaml:"dotenv,omitempty"`
	Run      string           `yaml:"run,omitempty"`
	Interval string           `yaml:"interval,omitempty"`
	Set      []string         `yaml:"set,omitempty"`
	Shopt    []string         `yaml:"shopt,omitempty"`
}

// HasTasks returns true if the Razdfile has any tasks defined
func (r *Razdfile) HasTasks() bool {
	return r.Tasks != nil && r.Tasks.Len() > 0
}

// HasIncludes returns true if the Razdfile has any includes defined
func (r *Razdfile) HasIncludes() bool {
	return r.Includes != nil && r.Includes.Len() > 0
}

// HasMise returns true if the Razdfile has mise configuration
func (r *Razdfile) HasMise() bool {
	return r.Mise != nil
}

// HasContent returns true if the Razdfile has any meaningful content
// (tasks, includes, or mise configuration)
func (r *Razdfile) HasContent() bool {
	return r.HasTasks() || r.HasIncludes() || r.HasMise()
}

// GetTask returns a task by name, or nil if not found
func (r *Razdfile) GetTask(name string) *taskast.Task {
	if r.Tasks == nil {
		return nil
	}
	task, ok := r.Tasks.Get(name)
	if !ok {
		return nil
	}
	return task
}
