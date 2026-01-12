package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"
)

func TestRazdfile_UnmarshalYAML_MinimalWithTasks(t *testing.T) {
	yamlContent := `
version: "1"

tasks:
  hello:
    cmds:
      - echo "Hello"
`

	var rf Razdfile
	err := yaml.Unmarshal([]byte(yamlContent), &rf)

	require.NoError(t, err)
	assert.Equal(t, "1", rf.Version)
	assert.True(t, rf.HasTasks())
	assert.False(t, rf.HasMise())
}

func TestRazdfile_UnmarshalYAML_WithMise(t *testing.T) {
	yamlContent := `
version: "1"

mise:
  tools:
    node: "22"
    python: "3.11"

tasks:
  install:
    cmds:
      - npm install
`

	var rf Razdfile
	err := yaml.Unmarshal([]byte(yamlContent), &rf)

	require.NoError(t, err)
	assert.Equal(t, "1", rf.Version)
	assert.True(t, rf.HasTasks())
	assert.True(t, rf.HasMise())
	require.NotNil(t, rf.Mise.Tools["node"])
	assert.Equal(t, "22", rf.Mise.Tools["node"].Version)
	require.NotNil(t, rf.Mise.Tools["python"])
	assert.Equal(t, "3.11", rf.Mise.Tools["python"].Version)
}

func TestRazdfile_UnmarshalYAML_OnlyMise(t *testing.T) {
	yamlContent := `
version: "1"

mise:
  tools:
    node: "22"
`

	var rf Razdfile
	err := yaml.Unmarshal([]byte(yamlContent), &rf)

	require.NoError(t, err)
	assert.True(t, rf.HasMise())
	assert.True(t, rf.HasContent())
}

func TestRazdfile_HasContent(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected bool
	}{
		{
			name: "with tasks",
			yaml: `
version: "1"
tasks:
  hello:
    cmds:
      - echo "hello"
`,
			expected: true,
		},
		{
			name: "with mise only",
			yaml: `
version: "1"
mise:
  tools:
    node: "22"
`,
			expected: true,
		},
		{
			name: "empty",
			yaml: `
version: "1"
`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rf Razdfile
			err := yaml.Unmarshal([]byte(tt.yaml), &rf)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, rf.HasContent())
		})
	}
}

func TestRazdfile_GetTask(t *testing.T) {
	yamlContent := `
version: "1"

tasks:
  hello:
    cmds:
      - echo "Hello"
  world:
    cmds:
      - echo "World"
`

	var rf Razdfile
	err := yaml.Unmarshal([]byte(yamlContent), &rf)
	require.NoError(t, err)

	// Existing task
	task := rf.GetTask("hello")
	assert.NotNil(t, task)

	// Non-existing task
	task = rf.GetTask("nonexistent")
	assert.Nil(t, task)
}

func TestVersion_IsSupported(t *testing.T) {
	assert.True(t, IsVersionSupported("1"))
	assert.False(t, IsVersionSupported("2"))
	assert.False(t, IsVersionSupported("0"))
	assert.False(t, IsVersionSupported(""))
}
