package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/razd-cli/razd/internal/output"
	"github.com/razd-cli/razd/razdfile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRunRemove_RemovesFromEnsure verifies razd rm removes the package from
// Razdfile.ensure. When no provisioner binary is available the delegation step
// is skipped, but the ensure list is still trimmed.
func TestRunRemove_RemovesFromEnsure(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Razdfile.yml"),
		[]byte("version: \"1\"\ndependencies:\n  using: devbox\n  ensure:\n    - \"nodejs@22\"\n    - \"uv\"\n    - \"bun\"\n"), 0644))

	ctx := &Context{
		Args: []string{"uv"},
		Log:  output.NewLogger(os.Stderr, os.Stderr),
		Dir:  dir,
	}

	err := runRemove(ctx)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, "Razdfile.yml"))
	require.NoError(t, err)
	content := string(data)
	assert.NotContains(t, content, "uv")
	assert.Contains(t, content, "nodejs@22")
	assert.Contains(t, content, "bun")
}

// TestRunRemove_NoMatchReportsNone verifies razd rm is a no-op when the
// requested tool is not present.
func TestRunRemove_NoMatchReportsNone(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Razdfile.yml"),
		[]byte("version: \"1\"\ndependencies:\n  using: devbox\n  ensure:\n    - \"nodejs@22\"\n"), 0644))

	ctx := &Context{
		Args: []string{"uv"},
		Log:  output.NewLogger(os.Stderr, os.Stderr),
		Dir:  dir,
	}

	err := runRemove(ctx)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, "Razdfile.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "nodejs@22")
}

// TestRunRemove_NoArgsErrors verifies razd rm requires an argument.
func TestRunRemove_NoArgsErrors(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Razdfile.yml"),
		[]byte("version: \"1\"\ndependencies:\n  using: devbox\n  ensure:\n    - \"nodejs@22\"\n"), 0644))

	ctx := &Context{
		Args: []string{},
		Log:  output.NewLogger(os.Stderr, os.Stderr),
		Dir:  dir,
	}

	err := runRemove(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "usage")
}

// TestRunRemove_NoDependenciesReportsNone verifies razd rm is a no-op when the
// Razdfile has no dependencies section.
func TestRunRemove_NoDependenciesReportsNone(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Razdfile.yml"),
		[]byte("version: \"1\"\ntasks:\n  default:\n    cmd: echo hi\n"), 0644))

	ctx := &Context{
		Args: []string{"uv"},
		Log:  output.NewLogger(os.Stderr, os.Stderr),
		Dir:  dir,
	}

	err := runRemove(ctx)
	require.NoError(t, err)
}

// TestRemoveFromEnsureInFile_RemovesByName verifies the writer removes both a
// bare entry and a versioned entry for the same tool name.
func TestRemoveFromEnsureInFile_RemovesByName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Razdfile.yml")
	require.NoError(t, os.WriteFile(path,
		[]byte("version: \"1\"\ndependencies:\n  using: devbox\n  ensure:\n    - \"nodejs@22\"\n    - \"uv\"\n    - \"uv@latest\"\n    - \"bun\"\n"), 0644))

	changed, err := razdfile.RemoveFromEnsureInFile(path, map[string]bool{"uv": true})
	require.NoError(t, err)
	assert.True(t, changed)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	content := string(data)
	assert.NotContains(t, content, "uv")
	assert.Contains(t, content, "nodejs@22")
	assert.Contains(t, content, "bun")
}

// TestRemoveFromEnsureInFile_NoMatchReturnsFalse verifies the writer reports no
// change when the requested tool is absent.
func TestRemoveFromEnsureInFile_NoMatchReturnsFalse(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Razdfile.yml")
	require.NoError(t, os.WriteFile(path,
		[]byte("version: \"1\"\ndependencies:\n  using: devbox\n  ensure:\n    - \"nodejs@22\"\n"), 0644))

	changed, err := razdfile.RemoveFromEnsureInFile(path, map[string]bool{"uv": true})
	require.NoError(t, err)
	assert.False(t, changed)
}
