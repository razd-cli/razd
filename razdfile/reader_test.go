package razdfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReader_Read_WithTasks(t *testing.T) {
	// Create a temporary directory
	dir := t.TempDir()
	
	// Create a test Razdfile
	content := []byte(`
version: "1"

tasks:
  hello:
    cmds:
      - echo "Hello"
`)
	err := os.WriteFile(filepath.Join(dir, "Razdfile.yml"), content, 0644)
	require.NoError(t, err)
	
	// Create a reader and read the file
	reader := NewReader(WithDir(dir))
	rf, err := reader.Read()
	
	require.NoError(t, err)
	assert.Equal(t, "1", rf.Version)
	assert.True(t, rf.HasTasks())
}

func TestReader_Read_WithMise(t *testing.T) {
	// Create a temporary directory
	dir := t.TempDir()
	
	// Create a test Razdfile with mise config
	content := []byte(`
version: "1"

mise:
  tools:
    node: "22"
    python: "3.11"

tasks:
  install:
    cmds:
      - npm install
`)
	err := os.WriteFile(filepath.Join(dir, "Razdfile.yml"), content, 0644)
	require.NoError(t, err)
	
	// Create a reader and read the file
	reader := NewReader(WithDir(dir))
	rf, err := reader.Read()
	
	require.NoError(t, err)
	assert.True(t, rf.HasTasks())
	assert.True(t, rf.HasMise())
	require.NotNil(t, rf.Mise.Tools["node"])
	assert.Equal(t, "22", rf.Mise.Tools["node"].Version)
}

func TestReader_Read_WithDevbox(t *testing.T) {
	// Create a temporary directory
	dir := t.TempDir()
	
	// Create a test Razdfile with devbox config
	content := []byte(`
version: "1"

devbox:
  packages:
    - python@3.11
    - ripgrep@latest
  shell:
    init_hook: echo "Hello devbox!"
    scripts:
      test: python --version

tasks:
  hello:
    cmds:
      - echo "Hello"
`)
	err := os.WriteFile(filepath.Join(dir, "Razdfile.yml"), content, 0644)
	require.NoError(t, err)
	
	// Create a reader and read the file
	reader := NewReader(WithDir(dir))
	rf, err := reader.Read()
	
	require.NoError(t, err)
	assert.True(t, rf.HasTasks())
	assert.True(t, rf.HasDevbox())
	require.True(t, rf.Devbox.HasPackages())
	assert.Len(t, rf.Devbox.Packages.List, 2)
	assert.Equal(t, "python@3.11", rf.Devbox.Packages.List[0])
	assert.True(t, rf.Devbox.HasShell())
	assert.Equal(t, []string{"echo \"Hello devbox!\""}, rf.Devbox.Shell.InitHook.Commands)
}

func TestReader_Read_NotFound(t *testing.T) {
	// Create an empty temporary directory
	dir := t.TempDir()
	
	// Create a reader and try to read
	reader := NewReader(WithDir(dir))
	_, err := reader.Read()
	
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestReader_Read_UnsupportedVersion(t *testing.T) {
	// Create a temporary directory
	dir := t.TempDir()
	
	// Create a Razdfile with unsupported version
	content := []byte(`
version: "99"

tasks:
  hello:
    cmds:
      - echo "Hello"
`)
	err := os.WriteFile(filepath.Join(dir, "Razdfile.yml"), content, 0644)
	require.NoError(t, err)
	
	// Create a reader and read the file
	reader := NewReader(WithDir(dir))
	_, err = reader.Read()
	
	assert.Error(t, err)
	var unsupportedErr *ErrUnsupportedVersion
	assert.ErrorAs(t, err, &unsupportedErr)
}

func TestReader_Exists(t *testing.T) {
	// Create a temporary directory
	dir := t.TempDir()
	
	reader := NewReader(WithDir(dir))
	assert.False(t, reader.Exists())
	
	// Create a Razdfile
	content := []byte(`version: "1"
tasks:
  hello:
    cmds:
      - echo "Hello"
`)
	err := os.WriteFile(filepath.Join(dir, "Razdfile.yml"), content, 0644)
	require.NoError(t, err)
	
	assert.True(t, reader.Exists())
}

func TestDefaultRazdfiles(t *testing.T) {
	// Check default Razdfile names
	assert.Contains(t, DefaultRazdfiles, "Razdfile.yml")
	assert.Contains(t, DefaultRazdfiles, "Razdfile.yaml")
	assert.Contains(t, DefaultRazdfiles, "razdfile.yml")
	assert.Contains(t, DefaultRazdfiles, "razdfile.yaml")
}

func TestReader_WithDebugFunc(t *testing.T) {
	debugCalled := false
	debugFn := func(format string, args ...any) {
		debugCalled = true
	}
	
	dir := t.TempDir()
	content := []byte(`version: "1"
tasks:
  hello:
    cmds:
      - echo "Hello"
`)
	err := os.WriteFile(filepath.Join(dir, "Razdfile.yml"), content, 0644)
	require.NoError(t, err)
	
	reader := NewReader(WithDir(dir), WithDebugFunc(debugFn))
	_, err = reader.Read()
	
	require.NoError(t, err)
	assert.True(t, debugCalled)
}

func TestReader_Read_WithDependencies(t *testing.T) {
	// Create a temporary directory
	dir := t.TempDir()
	
	// Create a test Razdfile with dependencies
	content := []byte(`
version: "1"

dependencies:
  using: "mise"
  ensure:
    - "node@22"
    - "php@8.4"
    - "pnpm@latest"

tasks:
  install:
    cmds:
      - npm install
`)
	err := os.WriteFile(filepath.Join(dir, "Razdfile.yml"), content, 0644)
	require.NoError(t, err)
	
	// Create a reader and read the file
	reader := NewReader(WithDir(dir))
	rf, err := reader.Read()
	
	require.NoError(t, err)
	assert.True(t, rf.HasDependencies())
	assert.Equal(t, "mise", rf.Dependencies.Using)
	assert.Len(t, rf.Dependencies.Ensure, 3)
	assert.Equal(t, "node@22", rf.Dependencies.Ensure[0])
	
	// Test ParseEnsure
	parsed, err := rf.Dependencies.ParseEnsure()
	require.NoError(t, err)
	assert.Len(t, parsed, 3)
	assert.Equal(t, "node", parsed[0].Tool)
	assert.Equal(t, "22", parsed[0].Version)
}

func TestReader_Read_WithDependenciesDevbox(t *testing.T) {
	dir := t.TempDir()
	
	content := []byte(`
version: "1"

dependencies:
  using: "devbox"
  ensure:
    - "go@1.21"
    - "ripgrep@latest"
  extra:
    devbox:
      shell:
        init_hook: |
          echo "Welcome!"
          export GOPATH=$PWD/.go
        scripts:
          build: "go build ."
      env:
        MY_VAR: "production"

tasks:
  build:
    cmds:
      - go build .
`)
	err := os.WriteFile(filepath.Join(dir, "Razdfile.yml"), content, 0644)
	require.NoError(t, err)
	
	reader := NewReader(WithDir(dir))
	rf, err := reader.Read()
	
	require.NoError(t, err)
	assert.True(t, rf.HasDependencies())
	assert.Equal(t, "devbox", rf.Dependencies.Using)
	assert.Len(t, rf.Dependencies.Ensure, 2)
	
	// Check extra section
	require.NotNil(t, rf.Dependencies.Extra)
	require.NotNil(t, rf.Dependencies.Extra.Devbox)
	
	// Check shell config is passed through
	shell, ok := rf.Dependencies.Extra.Devbox["shell"].(map[string]any)
	require.True(t, ok, "shell should be a map")
	_, hasInitHook := shell["init_hook"]
	assert.True(t, hasInitHook, "should have init_hook")
	
	// Check env is passed through
	env, ok := rf.Dependencies.Extra.Devbox["env"].(map[string]any)
	require.True(t, ok, "env should be a map")
	assert.Equal(t, "production", env["MY_VAR"])
	
	// Test GetProviderExtra
	extra := rf.Dependencies.GetProviderExtra()
	require.NotNil(t, extra)
	_, hasShell := extra["shell"]
	assert.True(t, hasShell)
}

func TestReader_Read_DependenciesOnlyIsValid(t *testing.T) {
	dir := t.TempDir()
	
	// Razdfile with only dependencies (no tasks)
	content := []byte(`
version: "1"

dependencies:
  using: "mise"
  ensure:
    - "node@22"
`)
	err := os.WriteFile(filepath.Join(dir, "Razdfile.yml"), content, 0644)
	require.NoError(t, err)
	
	reader := NewReader(WithDir(dir))
	rf, err := reader.Read()
	
	require.NoError(t, err)
	assert.True(t, rf.HasDependencies())
	assert.True(t, rf.HasContent())
	assert.False(t, rf.HasTasks())
}
