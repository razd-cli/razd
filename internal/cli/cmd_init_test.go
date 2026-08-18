package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/huh"
	"github.com/razd-cli/razd/internal/flags"
	"github.com/razd-cli/razd/internal/output"
	"github.com/razd-cli/razd/razdfile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildInitRazdfile_None(t *testing.T) {
	rf := buildInitRazdfile("none")

	// No dependencies section for "none".
	assert.Nil(t, rf.Dependencies)
	// HasContent must pass: needs a task to be valid.
	require.True(t, rf.HasTasks(), "none-init Razdfile must have a task to pass HasContent")
	// Validation must accept the generated file.
	require.NoError(t, razdfile.Validate(rf, "Razdfile.yml"))
}

func TestBuildInitRazdfile_NoneDefaultTask(t *testing.T) {
	rf := buildInitRazdfile("none")

	task := rf.GetTask("default")
	require.NotNil(t, task, "expected a 'default' task for none-init")
	assert.Len(t, task.Cmds, 1, "default task should have exactly one command")
}

func TestBuildInitRazdfile_Mise(t *testing.T) {
	rf := buildInitRazdfile("mise")

	require.NotNil(t, rf.Dependencies)
	assert.Equal(t, "mise", rf.Dependencies.Using)
	require.NoError(t, razdfile.Validate(rf, "Razdfile.yml"))
}

func TestBuildInitRazdfile_Devbox(t *testing.T) {
	rf := buildInitRazdfile("devbox")

	require.NotNil(t, rf.Dependencies)
	assert.Equal(t, "devbox", rf.Dependencies.Using)
	require.NoError(t, razdfile.Validate(rf, "Razdfile.yml"))
}

func TestDetectProvider_FallbackNone(t *testing.T) {
	dir := t.TempDir()
	assert.Equal(t, "none", detectProvider(dir))
}

func TestDetectProvider_MiseConfig(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "mise.toml"), []byte("[tools]"), 0644))
	assert.Equal(t, "mise", detectProvider(dir))

	// .tool-versions also maps to mise.
	dir2 := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir2, ".tool-versions"), []byte("node 22\n"), 0644))
	assert.Equal(t, "mise", detectProvider(dir2))
}

func TestDetectProvider_DevboxConfig(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "devbox.json"), []byte("{}"), 0644))
	assert.Equal(t, "devbox", detectProvider(dir))
}

func TestInitProviderOptions_UnixIncludesDevbox(t *testing.T) {
	opts := initProviderOptions("linux")

	values := optionValues(opts)
	assert.Equal(t, []string{"mise", "devbox", "none"}, values)
}

func TestInitProviderOptions_WindowsHidesDevbox(t *testing.T) {
	opts := initProviderOptions("windows")

	values := optionValues(opts)
	assert.Equal(t, []string{"mise", "none"}, values)
	assert.NotContains(t, values, "devbox", "devbox must not be offered on Windows")
}

func TestResolveInitProvider_UsingFlag(t *testing.T) {
	defer resetFlagsUsing()
	flags.Using = "none"

	using, err := resolveInitProvider(t.TempDir(), testLogger())
	require.NoError(t, err)
	assert.Equal(t, "none", using)
}

func TestResolveInitProvider_UsingFlagValid(t *testing.T) {
	defer resetFlagsUsing()
	flags.Using = "mise"

	using, err := resolveInitProvider(t.TempDir(), testLogger())
	require.NoError(t, err)
	assert.Equal(t, "mise", using)
}

func TestResolveInitProvider_UsingFlagInvalid(t *testing.T) {
	defer resetFlagsUsing()
	flags.Using = "npm"

	_, err := resolveInitProvider(t.TempDir(), testLogger())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid provider")
}

func TestResolveInitProvider_YesDefaultsNone(t *testing.T) {
	defer resetFlagsUsing()
	flags.Yes = true

	using, err := resolveInitProvider(t.TempDir(), testLogger())
	require.NoError(t, err)
	assert.Equal(t, "none", using)
}

func TestResolveInitProvider_NonInteractiveDetect(t *testing.T) {
	defer resetFlagsUsing()
	flags.Yes = false

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "mise.toml"), []byte("[tools]"), 0644))

	// Non-TTY (as in tests) → detectProvider finds mise.
	using, err := resolveInitProvider(dir, testLogger())
	require.NoError(t, err)
	assert.Equal(t, "mise", using)
}

func TestResolveInitProvider_NonInteractiveFallbackNone(t *testing.T) {
	defer resetFlagsUsing()
	flags.Yes = false

	dir := t.TempDir()
	using, err := resolveInitProvider(dir, testLogger())
	require.NoError(t, err)
	assert.Equal(t, "none", using)
}

// helpers

func optionValues[T comparable](opts []huh.Option[T]) []T {
	vals := make([]T, 0, len(opts))
	for _, o := range opts {
		vals = append(vals, o.Value)
	}
	return vals
}

func testLogger() *output.Logger {
	return output.NewLogger(os.Stderr, os.Stderr)
}

func resetFlagsUsing() {
	flags.Using = ""
	flags.Yes = false
}
