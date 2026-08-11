package razdfile

import (
	"os"
	"path/filepath"
	"testing"

	taskast "github.com/go-task/task/v3/taskfile/ast"
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

func TestUpdateEnsureInFile_CreatesEnsureWhenAbsent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Razdfile.yml")

	original := `version: "1"
dependencies:
  using: "mise"
`
	require.NoError(t, os.WriteFile(path, []byte(original), 0644))

	changed, err := UpdateEnsureInFile(path, []string{"node@22"})
	require.NoError(t, err)
	assert.True(t, changed)

	result, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(result), "node@22")
	assert.Contains(t, string(result), "using: \"mise\"")
}

func TestUpdateEnsureInFile_CreatesEnsureWhenAbsent_BareName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Razdfile.yml")

	original := `version: "1"
dependencies:
  using: "mise"
`
	require.NoError(t, os.WriteFile(path, []byte(original), 0644))

	changed, err := UpdateEnsureInFile(path, []string{"node"})
	require.NoError(t, err)
	assert.True(t, changed)

	result, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(result), "- node")
	assert.NotContains(t, string(result), "node@")
}

func TestUpdateTasksInFile_CreatesTasksSection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Razdfile.yml")

	original := `version: "1"
dependencies:
  using: "mise"
`
	require.NoError(t, os.WriteFile(path, []byte(original), 0644))

	task := &taskast.Task{Cmds: []*taskast.Cmd{{Cmd: "echo hi"}}}
	changed, err := UpdateTasksInFile(path, "hello", task)
	require.NoError(t, err)
	assert.True(t, changed)

	result, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(result), "tasks:")
	assert.Contains(t, string(result), "hello:")
	assert.Contains(t, string(result), "cmd: echo hi")
}

func TestUpdateTasksInFile_SingleCommandUsesCmdKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Razdfile.yml")

	original := `version: "1"
tasks:
  existing: echo old
`
	require.NoError(t, os.WriteFile(path, []byte(original), 0644))

	task := &taskast.Task{Cmds: []*taskast.Cmd{{Cmd: "go build ./..."}}}
	changed, err := UpdateTasksInFile(path, "build", task)
	require.NoError(t, err)
	assert.True(t, changed)

	result, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(result), "build:")
	assert.Contains(t, string(result), "cmd: go build ./...")
	// Existing task preserved.
	assert.Contains(t, string(result), "existing: echo old")
}

func TestUpdateTasksInFile_MultipleCommandsUseCmdsKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Razdfile.yml")

	original := `version: "1"
tasks:
  existing: echo old
`
	require.NoError(t, os.WriteFile(path, []byte(original), 0644))

	task := &taskast.Task{
		Cmds: []*taskast.Cmd{{Cmd: "go test ./..."}, {Cmd: "go vet ./..."}},
	}
	changed, err := UpdateTasksInFile(path, "test", task)
	require.NoError(t, err)
	assert.True(t, changed)

	result, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(result), "test:")
	assert.Contains(t, string(result), "cmds:")
	assert.Contains(t, string(result), "- go test ./...")
	assert.Contains(t, string(result), "- go vet ./...")
}

func TestUpdateTasksInFile_MappingForm(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Razdfile.yml")

	original := `version: "1"
tasks:
  existing: echo old
`
	require.NoError(t, os.WriteFile(path, []byte(original), 0644))

	task := &taskast.Task{
		Cmds: []*taskast.Cmd{{Cmd: "go test ./..."}},
		Desc: "Run tests",
		Deps: []*taskast.Dep{{Task: "build"}},
		Dir:  "./src",
	}
	changed, err := UpdateTasksInFile(path, "test", task)
	require.NoError(t, err)
	assert.True(t, changed)

	result, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(result), "test:")
	assert.Contains(t, string(result), "cmd: go test ./...")
	assert.Contains(t, string(result), "desc: Run tests")
	assert.Contains(t, string(result), "deps:")
	assert.Contains(t, string(result), "- build")
	assert.Contains(t, string(result), "dir: ./src")
	// Existing task preserved.
	assert.Contains(t, string(result), "existing: echo old")
}

func TestUpdateTasksInFile_PreservesComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Razdfile.yml")

	original := `# Project config
version: "1"

# Tasks
tasks:
  existing: echo old
`
	require.NoError(t, os.WriteFile(path, []byte(original), 0644))

	task := &taskast.Task{Cmds: []*taskast.Cmd{{Cmd: "echo new"}}}
	_, err := UpdateTasksInFile(path, "newtask", task)
	require.NoError(t, err)

	result, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(result), "# Project config")
	assert.Contains(t, string(result), "# Tasks")
	assert.Contains(t, string(result), "newtask:")
	assert.Contains(t, string(result), "cmd: echo new")
}
