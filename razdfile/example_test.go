package razdfile

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReader_ReadExample(t *testing.T) {
	// Test reading the example Razdfile
	exampleDir := filepath.Join("..", "examples", "nodejs-project")
	
	reader := NewReader(WithDir(exampleDir))
	rf, err := reader.Read()
	
	require.NoError(t, err)
	assert.Equal(t, "1", rf.Version)
	
	// Check dependencies config (unified format)
	assert.True(t, rf.HasDependencies())
	assert.Equal(t, "mise", rf.Dependencies.Using)
	assert.Len(t, rf.Dependencies.Ensure, 2)
	assert.Contains(t, rf.Dependencies.Ensure, "node@22")
	assert.Contains(t, rf.Dependencies.Ensure, "python@3.11")
	
	// Check tasks
	assert.True(t, rf.HasTasks())
	
	// Check specific tasks exist
	defaultTask := rf.GetTask("default")
	assert.NotNil(t, defaultTask)
	
	installTask := rf.GetTask("install")
	assert.NotNil(t, installTask)
	
	devTask := rf.GetTask("dev")
	assert.NotNil(t, devTask)
}
