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
