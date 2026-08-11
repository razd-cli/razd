package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/razd-cli/razd/internal/flags"
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

func TestRunAddTask_CreatesTask(t *testing.T) {
	dir := t.TempDir()
	writeFreshInitRazdfile(t, dir)

	ctx := &Context{
		Args: []string{"task", "hello", "echo", "hi"},
		Log:  output.NewLogger(os.Stderr, os.Stderr),
		Dir:  dir,
	}

	err := runAdd(ctx)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, "Razdfile.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "tasks:")
	assert.Contains(t, string(data), "hello:")
	assert.Contains(t, string(data), "cmd: echo hi")
}

func TestRunAddTask_WithFlags(t *testing.T) {
	dir := t.TempDir()
	writeFreshInitRazdfile(t, dir)

	// Set task flags directly (pflag is not exercised in unit tests).
	flags.TaskDesc = "Run tests"
	flags.TaskDeps = []string{"build"}
	flags.TaskDir = "./src"
	flags.TaskSilent = true
	flags.TaskInteractive = true
	t.Cleanup(func() {
		flags.TaskDesc = ""
		flags.TaskDeps = nil
		flags.TaskDir = ""
		flags.TaskSilent = false
		flags.TaskInteractive = false
	})

	ctx := &Context{
		Args: []string{"task", "test", "go", "test"},
		Log:  output.NewLogger(os.Stderr, os.Stderr),
		Dir:  dir,
	}

	err := runAdd(ctx)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, "Razdfile.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "test:")
	assert.Contains(t, string(data), "cmd: go test")
	assert.Contains(t, string(data), "desc: Run tests")
	assert.Contains(t, string(data), "deps:")
	assert.Contains(t, string(data), "- build")
	assert.Contains(t, string(data), "dir: ./src")
	assert.Contains(t, string(data), "silent: true")
	assert.Contains(t, string(data), "interactive: true")
}

func TestRunAddTask_InvalidName(t *testing.T) {
	dir := t.TempDir()
	writeFreshInitRazdfile(t, dir)

	ctx := &Context{
		Args: []string{"task", "bad@name", "echo", "hi"},
		Log:  output.NewLogger(os.Stderr, os.Stderr),
		Dir:  dir,
	}

	err := runAdd(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid task name")
}

func TestRunAddTask_MissingCommand(t *testing.T) {
	dir := t.TempDir()
	writeFreshInitRazdfile(t, dir)

	ctx := &Context{
		Args: []string{"task", "hello"},
		Log:  output.NewLogger(os.Stderr, os.Stderr),
		Dir:  dir,
	}

	err := runAdd(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires at least one command")
}

func TestRunAdd_StillAddsDependency(t *testing.T) {
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
	assert.NotContains(t, string(data), "tasks:")
}

func TestRunAdd_BareTaskAddsPackage(t *testing.T) {
	dir := t.TempDir()
	writeFreshInitRazdfile(t, dir)

	// A bare "razd add task" (no name) must add the `task` package, not error.
	ctx := &Context{
		Args: []string{"task"},
		Log:  output.NewLogger(os.Stderr, os.Stderr),
		Dir:  dir,
	}

	err := runAdd(ctx)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, "Razdfile.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "task")
	assert.NotContains(t, string(data), "tasks:")
}

func TestRunAddTask_OverwriteNonInteractiveSkips(t *testing.T) {
	dir := t.TempDir()
	// Task already exists.
	content := "version: \"1\"\ntasks:\n  hello:\n    cmd: echo old\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Razdfile.yml"), []byte(content), 0644))

	ctx := &Context{
		Args: []string{"task", "hello", "echo", "new"},
		Log:  output.NewLogger(os.Stderr, os.Stderr),
		Dir:  dir,
	}

	// Non-interactive stdin -> overwrite prompt returns false, task unchanged.
	err := runAdd(ctx)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, "Razdfile.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "echo old")
	assert.NotContains(t, string(data), "echo new")
}

func TestRunAddTask_OverwriteWithYes(t *testing.T) {
	dir := t.TempDir()
	content := "version: \"1\"\ntasks:\n  hello:\n    cmd: echo old\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Razdfile.yml"), []byte(content), 0644))

	// --yes forces overwrite without prompting.
	flags.Yes = true
	t.Cleanup(func() { flags.Yes = false })

	ctx := &Context{
		Args: []string{"task", "hello", "echo", "new"},
		Log:  output.NewLogger(os.Stderr, os.Stderr),
		Dir:  dir,
	}

	err := runAdd(ctx)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, "Razdfile.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(data), "echo new")
	assert.NotContains(t, string(data), "echo old")
}

func TestRunAddTask_AddsTaskToolToMise(t *testing.T) {
	dir := t.TempDir()
	writeFreshInitRazdfile(t, dir)

	ctx := &Context{
		Args: []string{"task", "hello", "echo", "hi"},
		Log:  output.NewLogger(os.Stderr, os.Stderr),
		Dir:  dir,
	}

	err := runAdd(ctx)
	require.NoError(t, err)

	// The 'task' tool must be added to mise.toml with 'latest' by default.
	mise, err := os.ReadFile(filepath.Join(dir, "mise.toml"))
	require.NoError(t, err)
	assert.Contains(t, string(mise), "task")
	assert.Contains(t, string(mise), "latest")
}

func TestRunAddTask_PreservesPinnedTaskVersion(t *testing.T) {
	dir := t.TempDir()
	writeFreshInitRazdfile(t, dir)
	// task already pinned in mise.toml.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "mise.toml"),
		[]byte("[tools]\ntask = \"1.9.0\"\n"), 0644))

	ctx := &Context{
		Args: []string{"task", "hello", "echo", "hi"},
		Log:  output.NewLogger(os.Stderr, os.Stderr),
		Dir:  dir,
	}

	err := runAdd(ctx)
	require.NoError(t, err)

	mise, err := os.ReadFile(filepath.Join(dir, "mise.toml"))
	require.NoError(t, err)
	assert.Contains(t, string(mise), "1.9.0")
	assert.NotContains(t, string(mise), "latest")
}
