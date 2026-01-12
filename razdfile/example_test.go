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
	
	// Check mise config
	assert.True(t, rf.HasMise())
	require.NotNil(t, rf.Mise.Tools["node"])
	assert.Equal(t, "22", rf.Mise.Tools["node"].Version)
	require.NotNil(t, rf.Mise.Tools["python"])
	assert.Equal(t, "3.11", rf.Mise.Tools["python"].Version)
	
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
