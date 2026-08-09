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
	// go-toml emits single quotes and sorted keys.
	assert.Equal(t, "[tools]\nnode = '22'\npnpm = 'latest'\n", string(content))
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
	assert.Contains(t, string(content), "[tools]\nnode = '22'")
	assert.Contains(t, string(content), "[env]")
	assert.Contains(t, string(content), "NODE_ENV = 'development'")
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

func TestMiseProvisioner_ReadConfig(t *testing.T) {
	dir := t.TempDir()
	p := NewMiseProvisioner(Config{Dir: dir})

	err := os.WriteFile(filepath.Join(dir, "mise.toml"), []byte("[tools]\nnode = \"22\"\npnpm = \"latest\"\n"), 0644)
	require.NoError(t, err)

	tools, err := p.ReadConfig()
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"node": "22", "pnpm": "latest"}, tools)
}

func TestMiseProvisioner_ReadConfig_NotExists(t *testing.T) {
	dir := t.TempDir()
	p := NewMiseProvisioner(Config{Dir: dir})

	tools, err := p.ReadConfig()
	require.NoError(t, err)
	assert.Nil(t, tools)
}

func TestMiseProvisioner_ReadConfig_IgnoresOtherSections(t *testing.T) {
	dir := t.TempDir()
	p := NewMiseProvisioner(Config{Dir: dir})

	content := "[env]\nNODE_ENV = \"development\"\n\n[tools]\nnode = \"22\"\npython = \"3.11\"\n\n[settings]\nexperimental = true\n"
	err := os.WriteFile(filepath.Join(dir, "mise.toml"), []byte(content), 0644)
	require.NoError(t, err)

	tools, err := p.ReadConfig()
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"node": "22", "python": "3.11"}, tools)
}

func TestDevboxProvisioner_ReadConfig(t *testing.T) {
	dir := t.TempDir()
	p := NewDevboxProvisioner(Config{Dir: dir})

	content := `{"packages": ["nodejs@22", "python@3.11"]}`
	err := os.WriteFile(filepath.Join(dir, "devbox.json"), []byte(content), 0644)
	require.NoError(t, err)

	tools, err := p.ReadConfig()
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"nodejs": "22", "python": "3.11"}, tools)
}

func TestDevboxProvisioner_ReadConfig_NotExists(t *testing.T) {
	dir := t.TempDir()
	p := NewDevboxProvisioner(Config{Dir: dir})

	tools, err := p.ReadConfig()
	require.NoError(t, err)
	assert.Nil(t, tools)
}

func TestMiseProvisioner_WriteTools_PreservesOtherSections(t *testing.T) {
	dir := t.TempDir()
	p := NewMiseProvisioner(Config{Dir: dir})

	// Pre-existing mise.toml with env/settings that must survive the merge.
	existing := "[env]\nNODE_ENV = \"production\"\n\n[tools]\npython = \"3.11\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "mise.toml"), []byte(existing), 0644))

	err := p.WriteTools(map[string]string{"node": "22"})
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, "mise.toml"))
	require.NoError(t, err)
	// Other sections preserved.
	assert.Contains(t, string(content), "NODE_ENV")
	assert.Contains(t, string(content), "production")
	// Existing tool preserved, new tool added.
	assert.Contains(t, string(content), "python = '3.11'")
	assert.Contains(t, string(content), "node = '22'")
}

func TestMiseProvisioner_WriteTools_ComplexToolsPreserved(t *testing.T) {
	dir := t.TempDir()
	p := NewMiseProvisioner(Config{Dir: dir})

	// Complex tool object must survive the round-trip.
	existing := "[tools]\nnode = { version = \"22\", os = [\"linux\"] }\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "mise.toml"), []byte(existing), 0644))

	err := p.WriteTools(map[string]string{"go": "1.21"})
	require.NoError(t, err)

	// Complex node preserved.
	tools, err := p.ReadConfig()
	require.NoError(t, err)
	assert.Equal(t, "22", tools["node"])
	assert.Equal(t, "1.21", tools["go"])
}

func TestDevboxProvisioner_WriteTools_PreservesUnknownKeys(t *testing.T) {
	dir := t.TempDir()
	p := NewDevboxProvisioner(Config{Dir: dir})

	existing := `{"packages": ["nodejs@22"], "env": {"NODE_ENV": "production"}, "shell": {"init_hook": "echo hi"}}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "devbox.json"), []byte(existing), 0644))

	err := p.WriteTools(map[string]string{"go": "1.21"})
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, "devbox.json"))
	require.NoError(t, err)
	// Unknown keys preserved.
	assert.Contains(t, string(content), "NODE_ENV")
	assert.Contains(t, string(content), "init_hook")
	// Existing package preserved, new added.
	assert.Contains(t, string(content), "nodejs@22")
	assert.Contains(t, string(content), "go@1.21")
}

func TestDevboxProvisioner_ReadConfig_VersionlessPackages(t *testing.T) {
	dir := t.TempDir()
	p := NewDevboxProvisioner(Config{Dir: dir})

	// Versionless devbox packages (php84Extensions.*, php84Packages.composer)
	// must be kept, with an empty version, so they participate in sync.
	content := `{"packages": ["php84Extensions.xdebug", "php84Packages.composer", "php@8.4.15", "nodejs@22"]}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "devbox.json"), []byte(content), 0644))

	tools, err := p.ReadConfig()
	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"php84Extensions.xdebug": "",
		"php84Packages.composer": "",
		"php":                   "8.4.15",
		"nodejs":                "22",
	}, tools)
}

func TestDevboxProvisioner_WriteTools_VersionlessNoTrailingAt(t *testing.T) {
	dir := t.TempDir()
	p := NewDevboxProvisioner(Config{Dir: dir})

	// WriteTools with an empty version must produce a bare name, never "x@" .
	err := p.WriteTools(map[string]string{"php84Extensions.xdebug": ""})
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(dir, "devbox.json"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "php84Extensions.xdebug")
	assert.NotContains(t, string(content), "php84Extensions.xdebug@")
}

