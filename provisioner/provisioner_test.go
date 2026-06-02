package provisioner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/razd-cli/razd/razdfile/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry_Get(t *testing.T) {
	tests := []struct {
		name       string
		want       string
		wantErr    bool
		errMessage string
	}{
		{
			name:    "mise",
			want:    "mise",
			wantErr: false,
		},
		{
			name:    "devbox",
			want:    "devbox",
			wantErr: false,
		},
		{
			name:       "unknown",
			wantErr:    true,
			errMessage: "unknown provisioner: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := Get(tt.name, Config{})
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMessage)
				assert.Nil(t, p)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, p)
				assert.Equal(t, tt.want, p.Name())
			}
		})
	}
}

func TestRegistry_Supported(t *testing.T) {
	supported := Supported()
	assert.Contains(t, supported, "mise")
	assert.Contains(t, supported, "devbox")
	assert.Len(t, supported, 2)
}

func TestRegistry_IsSupported(t *testing.T) {
	assert.True(t, IsSupported("mise"))
	assert.True(t, IsSupported("devbox"))
	assert.False(t, IsSupported("unknown"))
	assert.False(t, IsSupported(""))
}

func TestMiseProvisioner_Name(t *testing.T) {
	p := NewMiseProvisioner(Config{})
	assert.Equal(t, "mise", p.Name())
}

func TestDevboxProvisioner_Name(t *testing.T) {
	p := NewDevboxProvisioner(Config{})
	assert.Equal(t, "devbox", p.Name())
}

func TestMiseProvisioner_RunCommand(t *testing.T) {
	p := NewMiseProvisioner(Config{})
	cmd := p.RunCommand([]string{"npm", "install"})
	assert.Equal(t, []string{"mise", "exec", "--", "npm", "install"}, cmd)
}

func TestDevboxProvisioner_RunCommand(t *testing.T) {
	p := NewDevboxProvisioner(Config{})
	cmd := p.RunCommand([]string{"npm", "install"})
	assert.Equal(t, []string{"devbox", "run", "--", "npm", "install"}, cmd)
}

func TestMiseProvisioner_GenerateConfig(t *testing.T) {
	dir := t.TempDir()
	p := NewMiseProvisioner(Config{Dir: dir})

	packages := []ast.ParsedDependency{
		{Tool: "node", Version: "22", Raw: "node@22"},
		{Tool: "pnpm", Version: "latest", Raw: "pnpm@latest"},
	}

	err := p.GenerateConfig(packages, nil)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, "mise.toml"))
	require.NoError(t, err)
	assert.Equal(t, "[tools]\nnode = \"22\"\npnpm = \"latest\"\n", string(content))
}

func TestMiseProvisioner_GenerateConfig_WithExtra(t *testing.T) {
	dir := t.TempDir()
	p := NewMiseProvisioner(Config{Dir: dir})

	packages := []ast.ParsedDependency{
		{Tool: "node", Version: "22", Raw: "node@22"},
	}
	extra := map[string]any{
		"env": map[string]any{
			"NODE_ENV": "development",
		},
	}

	err := p.GenerateConfig(packages, extra)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, "mise.toml"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "[tools]\nnode = \"22\"")
	assert.Contains(t, string(content), "env:")
	assert.Contains(t, string(content), "NODE_ENV = \"development\"")
}

func TestMiseProvisioner_GenerateConfig_Idempotent(t *testing.T) {
	dir := t.TempDir()
	p := NewMiseProvisioner(Config{Dir: dir})

	packages := []ast.ParsedDependency{
		{Tool: "node", Version: "22", Raw: "node@22"},
	}

	err := p.GenerateConfig(packages, nil)
	require.NoError(t, err)

	info1, err := os.Stat(filepath.Join(dir, "mise.toml"))
	require.NoError(t, err)

	err = p.GenerateConfig(packages, nil)
	require.NoError(t, err)

	info2, err := os.Stat(filepath.Join(dir, "mise.toml"))
	require.NoError(t, err)
	assert.Equal(t, info1.ModTime(), info2.ModTime())
}

func TestDevboxProvisioner_GenerateConfig(t *testing.T) {
	dir := t.TempDir()
	p := NewDevboxProvisioner(Config{Dir: dir})

	packages := []ast.ParsedDependency{
		{Tool: "nodejs", Version: "22", Raw: "nodejs@22"},
	}

	err := p.GenerateConfig(packages, nil)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, "devbox.json"))
	require.NoError(t, err)
	assert.Contains(t, string(content), `"nodejs@22"`)
	assert.Contains(t, string(content), `"packages"`)
}

func TestDevboxProvisioner_GenerateConfig_WithExtra(t *testing.T) {
	dir := t.TempDir()
	p := NewDevboxProvisioner(Config{Dir: dir})

	packages := []ast.ParsedDependency{
		{Tool: "nodejs", Version: "22", Raw: "nodejs@22"},
	}
	extra := map[string]any{
		"shell": map[string]any{
			"init_hook": []any{"echo 'Welcome!'"},
		},
	}

	err := p.GenerateConfig(packages, extra)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, "devbox.json"))
	require.NoError(t, err)
	assert.Contains(t, string(content), `"nodejs@22"`)
	assert.Contains(t, string(content), `"init_hook"`)
}
