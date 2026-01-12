package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"
)

func TestDevboxPackages_UnmarshalYAML_Array(t *testing.T) {
	yamlContent := `
- python@3.10
- poetry@latest
- nodejs@22
`

	var packages DevboxPackages
	err := yaml.Unmarshal([]byte(yamlContent), &packages)

	require.NoError(t, err)
	assert.Len(t, packages.List, 3)
	assert.Equal(t, "python@3.10", packages.List[0])
	assert.Equal(t, "poetry@latest", packages.List[1])
	assert.Equal(t, "nodejs@22", packages.List[2])
}

func TestDevboxPackages_UnmarshalYAML_Object(t *testing.T) {
	yamlContent := `
python:
  version: "3.10"
  platforms:
    - x86_64-linux
    - aarch64-darwin
poetry: "latest"
`

	var packages DevboxPackages
	err := yaml.Unmarshal([]byte(yamlContent), &packages)

	require.NoError(t, err)
	require.NotNil(t, packages.Map["python"])
	assert.Equal(t, "3.10", packages.Map["python"].Version)
	assert.Equal(t, []string{"x86_64-linux", "aarch64-darwin"}, packages.Map["python"].Platforms)
	require.NotNil(t, packages.Map["poetry"])
	assert.Equal(t, "latest", packages.Map["poetry"].Version)
}

func TestDevboxInitHook_UnmarshalYAML_String(t *testing.T) {
	yamlContent := `"poetry install"`

	var hook DevboxInitHook
	err := yaml.Unmarshal([]byte(yamlContent), &hook)

	require.NoError(t, err)
	assert.Len(t, hook.Commands, 1)
	assert.Equal(t, "poetry install", hook.Commands[0])
}

func TestDevboxInitHook_UnmarshalYAML_Array(t *testing.T) {
	yamlContent := `
- poetry install
- npm install
- echo "Done"
`

	var hook DevboxInitHook
	err := yaml.Unmarshal([]byte(yamlContent), &hook)

	require.NoError(t, err)
	assert.Len(t, hook.Commands, 3)
	assert.Equal(t, "poetry install", hook.Commands[0])
	assert.Equal(t, "npm install", hook.Commands[1])
}

func TestDevboxScript_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected []string
	}{
		{
			name:     "single command",
			yaml:     `"npm run build"`,
			expected: []string{"npm run build"},
		},
		{
			name: "multiple commands",
			yaml: `
- npm run lint
- npm run build
- npm run test
`,
			expected: []string{"npm run lint", "npm run build", "npm run test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var script DevboxScript
			err := yaml.Unmarshal([]byte(tt.yaml), &script)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, script.Commands)
		})
	}
}

func TestDevboxConfig_Full(t *testing.T) {
	yamlContent := `
name: my-project
description: My awesome project
packages:
  - python@3.10
  - poetry@latest
env:
  DEBUG: "true"
  PORT: "8080"
shell:
  init_hook:
    - poetry install
    - npm install
  scripts:
    start: poetry run python -m main
    test:
      - poetry run pytest
      - npm run test
include:
  - github:my-org/mongodb
  - github:my-org/redis
`

	var config DevboxConfig
	err := yaml.Unmarshal([]byte(yamlContent), &config)

	require.NoError(t, err)
	assert.Equal(t, "my-project", config.Name)
	assert.Equal(t, "My awesome project", config.Description)
	assert.Len(t, config.Packages.List, 2)
	assert.Equal(t, "true", config.Env["DEBUG"])
	assert.Equal(t, "8080", config.Env["PORT"])
	require.NotNil(t, config.Shell)
	assert.Len(t, config.Shell.InitHook.Commands, 2)
	assert.Len(t, config.Shell.Scripts, 2)
	assert.Equal(t, []string{"poetry run python -m main"}, config.Shell.Scripts["start"].Commands)
	assert.Equal(t, []string{"poetry run pytest", "npm run test"}, config.Shell.Scripts["test"].Commands)
	assert.Len(t, config.Include, 2)
}

func TestDevboxConfig_HasMethods(t *testing.T) {
	t.Run("nil config", func(t *testing.T) {
		var config *DevboxConfig
		assert.False(t, config.HasPackages())
		assert.False(t, config.HasShell())
		assert.False(t, config.HasInclude())
	})

	t.Run("empty config", func(t *testing.T) {
		config := &DevboxConfig{}
		assert.False(t, config.HasPackages())
		assert.False(t, config.HasShell())
		assert.False(t, config.HasInclude())
	})

	t.Run("with packages list", func(t *testing.T) {
		config := &DevboxConfig{
			Packages: DevboxPackages{List: []string{"python@3.10"}},
		}
		assert.True(t, config.HasPackages())
	})

	t.Run("with packages map", func(t *testing.T) {
		config := &DevboxConfig{
			Packages: DevboxPackages{Map: map[string]*DevboxPackageConfig{"python": {Version: "3.10"}}},
		}
		assert.True(t, config.HasPackages())
	})

	t.Run("with shell", func(t *testing.T) {
		config := &DevboxConfig{
			Shell: &DevboxShell{},
		}
		assert.True(t, config.HasShell())
	})

	t.Run("with include", func(t *testing.T) {
		config := &DevboxConfig{
			Include: []string{"github:my-org/plugin"},
		}
		assert.True(t, config.HasInclude())
	})
}
