package sync

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// In CI/test environments stdin is not a TTY, so the interactive prompts must
// fall back to their safe non-blocking defaults: no backup, and Skip on
// conflicts (no data overwritten).

func TestPromptBackup_NonInteractiveReturnsNonInteractive(t *testing.T) {
	// stdin is not a TTY under go test, so this must not block or prompt.
	decision, err := PromptBackup("mise.toml", noopLogger{})
	require.NoError(t, err)
	assert.Equal(t, BackupNonInteractive, decision)
}

func TestPromptConflict_NonInteractiveSkips(t *testing.T) {
	res, err := PromptConflict("node", "22", "23", noopLogger{})
	require.NoError(t, err)
	assert.Equal(t, ResolutionSkip, res)
}

func TestPromptConflict_NonInteractiveIsSafe(t *testing.T) {
	// The default must never overwrite either side.
	res, err := PromptConflict("node", "22", "23", noopLogger{})
	require.NoError(t, err)
	assert.NotEqual(t, ResolutionUseRazdfile, res)
	assert.NotEqual(t, ResolutionUseNative, res)
}
