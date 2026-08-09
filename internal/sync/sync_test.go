package sync

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/razd-cli/razd/provisioner"
	"github.com/razd-cli/razd/razdfile/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeRazdfile(t *testing.T, dir string, ensure []string) *ast.Razdfile {
	t.Helper()
	rf := &ast.Razdfile{
		Version: "1",
		Dependencies: &ast.DependenciesConfig{
			Using:  "mise",
			Ensure: ensure,
		},
	}
	// Persist to disk so applyToRazdfile's UpdateEnsureInFile can write back.
	data := "version: \"1\"\ndependencies:\n  using: mise\n  ensure:\n"
	for _, e := range ensure {
		data += "    - \"" + e + "\"\n"
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Razdfile.yml"), []byte(data), 0644))
	return rf
}

func TestSync_NativeOnlyToolAddedToRazdfile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "mise.toml"),
		[]byte("[tools]\nnode = \"22\"\n"), 0644))

	// Razdfile has go but not node; node exists only in mise.toml.
	rf := writeRazdfile(t, dir, []string{"go@1.21"})
	prov := provisioner.NewMiseProvisioner(provisioner.Config{Dir: dir})

	err := Sync(rf, prov, dir, noopLogger{}, false)
	require.NoError(t, err)

	// node pulled into Razdfile.
	content, err := os.ReadFile(filepath.Join(dir, "Razdfile.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "node@22")
	// mise.toml now has both go (pushed from razd) and node (pre-existing).
	mise, err := os.ReadFile(filepath.Join(dir, "mise.toml"))
	require.NoError(t, err)
	assert.Contains(t, string(mise), "node = '22'")
}

func TestSync_RazdfileOnlyToolAddedToNative(t *testing.T) {
	dir := t.TempDir()
	rf := writeRazdfile(t, dir, []string{"node@22"})
	prov := provisioner.NewMiseProvisioner(provisioner.Config{Dir: dir})

	err := Sync(rf, prov, dir, noopLogger{}, false)
	require.NoError(t, err)

	mise, err := os.ReadFile(filepath.Join(dir, "mise.toml"))
	require.NoError(t, err)
	assert.Contains(t, string(mise), "node = '22'")
}

func TestSync_PreservesNativeSections(t *testing.T) {
	dir := t.TempDir()
	// Pre-existing mise.toml with an [env] section that must survive.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "mise.toml"),
		[]byte("[env]\nNODE_ENV = \"production\"\n\n[tools]\npython = \"3.11\"\n"), 0644))

	rf := writeRazdfile(t, dir, []string{"node@22"})
	prov := provisioner.NewMiseProvisioner(provisioner.Config{Dir: dir})

	err := Sync(rf, prov, dir, noopLogger{}, false)
	require.NoError(t, err)

	mise, err := os.ReadFile(filepath.Join(dir, "mise.toml"))
	require.NoError(t, err)
	// [env] preserved.
	assert.Contains(t, string(mise), "NODE_ENV")
	assert.Contains(t, string(mise), "production")
	// Both node (from razd) and python (pre-existing) present.
	assert.Contains(t, string(mise), "node = '22'")
	assert.Contains(t, string(mise), "python = '3.11'")
}

func TestSync_NoDependenciesSkips(t *testing.T) {
	dir := t.TempDir()
	rf := &ast.Razdfile{Version: "1"} // no dependencies
	prov := provisioner.NewMiseProvisioner(provisioner.Config{Dir: dir})

	err := Sync(rf, prov, dir, noopLogger{}, false)
	require.NoError(t, err)
	// No mise.toml should be created.
	_, statErr := os.Stat(filepath.Join(dir, "mise.toml"))
	assert.True(t, os.IsNotExist(statErr))
}

func TestSync_ConflictNonInteractiveSkips(t *testing.T) {
	dir := t.TempDir()
	// Native has node@23, Razdfile has node@22 -> conflict.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "mise.toml"),
		[]byte("[tools]\nnode = \"23\"\n"), 0644))
	rf := writeRazdfile(t, dir, []string{"node@22"})
	prov := provisioner.NewMiseProvisioner(provisioner.Config{Dir: dir})

	err := Sync(rf, prov, dir, noopLogger{}, false)
	require.NoError(t, err)

	// Non-TTY -> Skip: neither side changes.
	mise, err := os.ReadFile(filepath.Join(dir, "mise.toml"))
	require.NoError(t, err)
	assert.Contains(t, string(mise), "node = \"23\"")
	razd, err := os.ReadFile(filepath.Join(dir, "Razdfile.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(razd), "node@22")
}
