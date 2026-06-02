package razdfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateEnsureInFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Razdfile.yml")

	original := `version: "1"

dependencies:
  using: "mise"
  ensure:
    - "node@22"
    - "pnpm@10"

tasks:
  default:
    desc: "Install deps and run"
    cmds:
      - task: install
  install:
    desc: "Install dependencies"
    cmds:
      - pnpm install
`
	require.NoError(t, os.WriteFile(path, []byte(original), 0644))

	changed, err := UpdateEnsureInFile(path, []string{"node@24", "pnpm@10"})
	require.NoError(t, err)
	assert.True(t, changed)

	result, err := os.ReadFile(path)
	require.NoError(t, err)

	resultStr := string(result)
	assert.Contains(t, resultStr, `node@24`)
	assert.Contains(t, resultStr, `pnpm@10`)
	assert.Contains(t, resultStr, "tasks:")
	assert.Contains(t, resultStr, "pnpm install")
	assert.Contains(t, resultStr, `version: "1"`)
	assert.Contains(t, resultStr, `using: "mise"`)
}

func TestUpdateEnsureInFile_NoChange(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Razdfile.yml")

	original := `version: "1"
dependencies:
  using: "mise"
  ensure:
    - "node@22"
`
	require.NoError(t, os.WriteFile(path, []byte(original), 0644))

	changed, err := UpdateEnsureInFile(path, []string{"node@22"})
	require.NoError(t, err)
	assert.False(t, changed)
}

func TestUpdateEnsureInFile_PreservesComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Razdfile.yml")

	original := `# Project config
version: "1"

# Dependencies
dependencies:
  using: "mise"
  ensure:
    - "node@22"  # Node.js runtime

# Tasks
tasks:
  default:
    cmds:
      - echo hello
`
	require.NoError(t, os.WriteFile(path, []byte(original), 0644))

	_, err := UpdateEnsureInFile(path, []string{"node@24"})
	require.NoError(t, err)

	result, err := os.ReadFile(path)
	require.NoError(t, err)

	resultStr := string(result)
	assert.Contains(t, resultStr, "node@24")
	assert.Contains(t, resultStr, "# Project config")
	assert.Contains(t, resultStr, "echo hello")
}

func TestUpdateEnsureInFile_AddNewItem(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Razdfile.yml")

	original := `version: "1"
dependencies:
  using: "mise"
  ensure:
    - "node@22"
tasks:
  default:
    cmds:
      - echo hello
`
	require.NoError(t, os.WriteFile(path, []byte(original), 0644))

	changed, err := UpdateEnsureInFile(path, []string{"node@22", "python@3.12"})
	require.NoError(t, err)
	assert.True(t, changed)

	result, err := os.ReadFile(path)
	require.NoError(t, err)

	assert.Contains(t, string(result), "python@3.12")
	assert.Contains(t, string(result), "echo hello")
}