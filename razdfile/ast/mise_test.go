package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"
)

func TestMiseTool_UnmarshalYAML_SimpleString(t *testing.T) {
	yamlContent := `"22"`

	var tool MiseTool
	err := yaml.Unmarshal([]byte(yamlContent), &tool)

	require.NoError(t, err)
	assert.Equal(t, "22", tool.Version)
}

func TestMiseTool_UnmarshalYAML_ComplexObject(t *testing.T) {
	yamlContent := `
version: "22"
os:
  - linux
  - darwin
postinstall: "npm install -g pnpm"
`

	var tool MiseTool
	err := yaml.Unmarshal([]byte(yamlContent), &tool)

	require.NoError(t, err)
	assert.Equal(t, "22", tool.Version)
	assert.Equal(t, []string{"linux", "darwin"}, tool.OS)
	assert.Equal(t, "npm install -g pnpm", tool.Postinstall)
}

func TestMiseTool_UnmarshalYAML_MultipleVersions(t *testing.T) {
	yamlContent := `["20", "22"]`

	var tool MiseTool
	err := yaml.Unmarshal([]byte(yamlContent), &tool)

	require.NoError(t, err)
	assert.Equal(t, "20", tool.Version) // first version is primary
}

func TestMiseConfig_UnmarshalYAML(t *testing.T) {
	yamlContent := `
tools:
  node: "22"
  python: "3.11"
env:
  NODE_ENV: development
settings:
  auto_install: true
  experimental: false
`

	var config MiseConfig
	err := yaml.Unmarshal([]byte(yamlContent), &config)

	require.NoError(t, err)
	require.NotNil(t, config.Tools["node"])
	assert.Equal(t, "22", config.Tools["node"].Version)
	require.NotNil(t, config.Tools["python"])
	assert.Equal(t, "3.11", config.Tools["python"].Version)
	assert.Equal(t, "development", config.Env["NODE_ENV"])
	require.NotNil(t, config.Settings)
	require.NotNil(t, config.Settings.AutoInstall)
	assert.True(t, *config.Settings.AutoInstall)
	require.NotNil(t, config.Settings.Experimental)
	assert.False(t, *config.Settings.Experimental)
}

func TestMiseConfig_HasTools(t *testing.T) {
	tests := []struct {
		name     string
		config   *MiseConfig
		expected bool
	}{
		{
			name:     "nil config",
			config:   nil,
			expected: false,
		},
		{
			name:     "empty tools",
			config:   &MiseConfig{},
			expected: false,
		},
		{
			name: "with tools",
			config: &MiseConfig{
				Tools: map[string]*MiseTool{
					"node": {Version: "22"},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.config.HasTools())
		})
	}
}

func TestMiseSettings_UnmarshalYAML(t *testing.T) {
	yamlContent := `
auto_install: true
experimental: false
jobs: 4
`

	var settings MiseSettings
	err := yaml.Unmarshal([]byte(yamlContent), &settings)

	require.NoError(t, err)
	require.NotNil(t, settings.AutoInstall)
	assert.True(t, *settings.AutoInstall)
	require.NotNil(t, settings.Experimental)
	assert.False(t, *settings.Experimental)
	require.NotNil(t, settings.Jobs)
	assert.Equal(t, 4, *settings.Jobs)
}
