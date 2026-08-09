package sync

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// noopLogger satisfies Logger with no-op output.
type noopLogger struct{}

func (noopLogger) Debugf(string, ...any)   {}
func (noopLogger) Infof(string, ...any)    {}
func (noopLogger) Warnf(string, ...any)    {}
func (noopLogger) Successf(string, ...any) {}

func TestBackupFile_CreatesTimestampedCopy(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "mise.toml")
	content := "[tools]\nnode = \"22\"\n"
	require.NoError(t, os.WriteFile(src, []byte(content), 0644))

	backupPath, err := BackupFile(src, noopLogger{})
	require.NoError(t, err)
	assert.NotEmpty(t, backupPath)
	assert.FileExists(t, backupPath)

	data, err := os.ReadFile(backupPath)
	require.NoError(t, err)
	assert.Equal(t, content, string(data))

	// Original untouched.
	orig, err := os.ReadFile(src)
	require.NoError(t, err)
	assert.Equal(t, content, string(orig))
}

func TestBackupFile_MissingSourceNoop(t *testing.T) {
	dir := t.TempDir()
	backupPath, err := BackupFile(filepath.Join(dir, "nope.toml"), noopLogger{})
	require.NoError(t, err)
	assert.Empty(t, backupPath)
}

func TestBackupFile_UniqueTimestamps(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "devbox.json")
	require.NoError(t, os.WriteFile(src, []byte(`{"packages":[]}`), 0644))

	p1, err := BackupFile(src, noopLogger{})
	require.NoError(t, err)
	p2, err := BackupFile(src, noopLogger{})
	require.NoError(t, err)
	assert.NotEqual(t, p1, p2)
	assert.FileExists(t, p1)
	assert.FileExists(t, p2)
}
