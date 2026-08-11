package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/razd-cli/razd/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeAddRazdfile(t *testing.T, dir string) {
	t.Helper()
	content := "version: \"1\"\ndependencies:\n  using: mise\n  ensure:\n    - \"go@1.21\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Razdfile.yml"), []byte(content), 0644))
}

func TestRunAdd_WritesEnsureList(t *testing.T) {
	dir := t.TempDir()
	writeAddRazdfile(t, dir)

	ctx := &Context{
		Args: []string{"node@22"},
		Log:  output.NewLogger(os.Stderr, os.Stderr),
		Dir:  dir,
	}

	err := runAdd(ctx)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, "Razdfile.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "node@22")
	assert.Contains(t, string(data), "go@1.21")
}

func TestRunAdd_NoArgsErrors(t *testing.T) {
	ctx := &Context{
		Args: []string{},
		Log:  output.NewLogger(os.Stderr, os.Stderr),
	}
	err := runAdd(ctx)
	assert.Error(t, err)
}

func TestRunAdd_InvalidFormatErrors(t *testing.T) {
	dir := t.TempDir()
	writeAddRazdfile(t, dir)

	ctx := &Context{
		Args: []string{"@22"},
		Log:  output.NewLogger(os.Stderr, os.Stderr),
		Dir:  dir,
	}
	err := runAdd(ctx)
	assert.Error(t, err)
}

func TestRunAdd_NoDependenciesSectionErrors(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Razdfile.yml"),
		[]byte("version: \"1\"\ntasks:\n  default:\n    cmds: [echo hi]\n"), 0644))

	ctx := &Context{
		Args: []string{"node@22"},
		Log:  output.NewLogger(os.Stderr, os.Stderr),
		Dir:  dir,
	}
	err := runAdd(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dependencies")
}

func TestRunAdd_DuplicateSkips(t *testing.T) {
	dir := t.TempDir()
	writeAddRazdfile(t, dir)

	ctx := &Context{
		Args: []string{"go@1.21"}, // already present
		Log:  output.NewLogger(os.Stderr, os.Stderr),
		Dir:  dir,
	}
	err := runAdd(ctx)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, "Razdfile.yml"))
	require.NoError(t, err)
	// No duplicate go@1.21 entries.
	assert.Equal(t, 1, countOccurrences(string(data), "go@1.21"))
}

func countOccurrences(s, sub string) int {
	n := 0
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			n++
		}
	}
	return n
}

func writeFreshInitRazdfile(t *testing.T, dir string) {
	t.Helper()
	// Mirrors `razd init`: dependencies.using present, no ensure key.
	content := "version: \"1\"\ndependencies:\n  using: mise\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Razdfile.yml"), []byte(content), 0644))
}

func TestRunAdd_CreatesEnsureOnFreshInit(t *testing.T) {
	dir := t.TempDir()
	writeFreshInitRazdfile(t, dir)

	ctx := &Context{
		Args: []string{"node@22"},
		Log:  output.NewLogger(os.Stderr, os.Stderr),
		Dir:  dir,
	}

	err := runAdd(ctx)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, "Razdfile.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "node@22")
}

func TestRunAdd_BareNameCreatesEnsure(t *testing.T) {
	dir := t.TempDir()
	writeFreshInitRazdfile(t, dir)

	ctx := &Context{
		Args: []string{"node"},
		Log:  output.NewLogger(os.Stderr, os.Stderr),
		Dir:  dir,
	}

	err := runAdd(ctx)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, "Razdfile.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "- node")
	assert.NotContains(t, string(data), "node@")
}
