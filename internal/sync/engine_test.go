package sync

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// alwaysUseRazdfile resolves every conflict to the Razdfile version.
func alwaysUseRazdfile(tool, razdVer, nativeVer string) (Resolution, error) {
	return ResolutionUseRazdfile, nil
}

// alwaysUseNative resolves every conflict to the native version.
func alwaysUseNative(tool, razdVer, nativeVer string) (Resolution, error) {
	return ResolutionUseNative, nil
}

func TestReconcile_ToolOnlyInRazdfile(t *testing.T) {
	razd := []Tool{{Name: "node", Version: "22"}}
	changes, err := Reconcile(razd, nil, alwaysUseRazdfile)
	require.NoError(t, err)
	assert.Equal(t, []Tool{{Name: "node", Version: "22"}}, changes.ToNative)
	assert.Empty(t, changes.ToRazdfile)
	assert.True(t, changes.HasChanges())
}

func TestReconcile_ToolOnlyInNative(t *testing.T) {
	native := []Tool{{Name: "python", Version: "3.11"}}
	changes, err := Reconcile(nil, native, alwaysUseRazdfile)
	require.NoError(t, err)
	assert.Equal(t, []Tool{{Name: "python", Version: "3.11"}}, changes.ToRazdfile)
	assert.Empty(t, changes.ToNative)
}

func TestReconcile_SameVersionNoChange(t *testing.T) {
	razd := []Tool{{Name: "node", Version: "22"}}
	native := []Tool{{Name: "node", Version: "22"}}
	changes, err := Reconcile(razd, native, alwaysUseRazdfile)
	require.NoError(t, err)
	assert.False(t, changes.HasChanges())
	assert.Empty(t, changes.ToNative)
	assert.Empty(t, changes.ToRazdfile)
}

func TestReconcile_Conflict_UseRazdfile(t *testing.T) {
	razd := []Tool{{Name: "node", Version: "22"}}
	native := []Tool{{Name: "node", Version: "23"}}
	changes, err := Reconcile(razd, native, alwaysUseRazdfile)
	require.NoError(t, err)
	// Razdfile wins: push razd version to native.
	assert.Equal(t, []Tool{{Name: "node", Version: "22"}}, changes.ToNative)
	assert.Empty(t, changes.ToRazdfile)
}

func TestReconcile_Conflict_UseNative(t *testing.T) {
	razd := []Tool{{Name: "node", Version: "22"}}
	native := []Tool{{Name: "node", Version: "23"}}
	changes, err := Reconcile(razd, native, alwaysUseNative)
	require.NoError(t, err)
	// Native wins: pull native version into razd.
	assert.Empty(t, changes.ToNative)
	assert.Equal(t, []Tool{{Name: "node", Version: "23"}}, changes.ToRazdfile)
}

func TestReconcile_Conflict_Skip(t *testing.T) {
	razd := []Tool{{Name: "node", Version: "22"}}
	native := []Tool{{Name: "node", Version: "23"}}
	changes, err := Reconcile(razd, native, func(tool, a, b string) (Resolution, error) {
		return ResolutionSkip, nil
	})
	require.NoError(t, err)
	assert.False(t, changes.HasChanges())
	assert.Empty(t, changes.ToNative)
	assert.Empty(t, changes.ToRazdfile)
}

func TestReconcile_BothDirections(t *testing.T) {
	razd := []Tool{
		{Name: "node", Version: "22"},
		{Name: "go", Version: "1.21"},
	}
	native := []Tool{
		{Name: "node", Version: "22"},
		{Name: "python", Version: "3.11"},
	}
	changes, err := Reconcile(razd, native, alwaysUseRazdfile)
	require.NoError(t, err)
	// go only in razd -> native; python only in native -> razd.
	assert.Equal(t, []Tool{{Name: "go", Version: "1.21"}}, changes.ToNative)
	assert.Equal(t, []Tool{{Name: "python", Version: "3.11"}}, changes.ToRazdfile)
}

func TestReconcile_ResolverError(t *testing.T) {
	razd := []Tool{{Name: "node", Version: "22"}}
	native := []Tool{{Name: "node", Version: "23"}}
	_, err := Reconcile(razd, native, func(tool, a, b string) (Resolution, error) {
		return ResolutionSkip, assert.AnError
	})
	require.Error(t, err)
}

func TestReconcile_EmptyBoth(t *testing.T) {
	changes, err := Reconcile(nil, nil, alwaysUseRazdfile)
	require.NoError(t, err)
	assert.False(t, changes.HasChanges())
}
