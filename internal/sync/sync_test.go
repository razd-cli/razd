package sync

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/razd-cli/razd/provisioner"
	"github.com/razd-cli/razd/razdfile/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeRazdfile(t *testing.T, dir string, ensure []string) *ast.Razdfile {
	t.Helper()
	rf := &ast.Razdfile{
		Version: "1",
		Dependencies: &ast.DependenciesConfig{
			Using:  "mise",
			Ensure: ensure,
		},
	}
	// Persist to disk so applyToRazdfile's UpdateEnsureInFile can write back.
	data := "version: \"1\"\ndependencies:\n  using: mise\n  ensure:\n"
	for _, e := range ensure {
		data += "    - \"" + e + "\"\n"
	}
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Razdfile.yml"), []byte(data), 0644))
	return rf
}

func TestSync_NativeOnlyToolAddedToRazdfile(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "mise.toml"),
		[]byte("[tools]\nnode = \"22\"\n"), 0644))

	// Razdfile has go but not node; node exists only in mise.toml.
	rf := writeRazdfile(t, dir, []string{"go@1.21"})
	prov := provisioner.NewMiseProvisioner(provisioner.Config{Dir: dir})

	err := Sync(rf, prov, dir, noopLogger{}, false, false)
	require.NoError(t, err)

	// node pulled into Razdfile.
	content, err := os.ReadFile(filepath.Join(dir, "Razdfile.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "node@22")
	// mise.toml now has both go (pushed from razd) and node (pre-existing).
	mise, err := os.ReadFile(filepath.Join(dir, "mise.toml"))
	require.NoError(t, err)
	assert.Contains(t, string(mise), "node = '22'")
}

func TestSync_RazdfileOnlyToolAddedToNative(t *testing.T) {
	dir := t.TempDir()
	rf := writeRazdfile(t, dir, []string{"node@22"})
	prov := provisioner.NewMiseProvisioner(provisioner.Config{Dir: dir})

	err := Sync(rf, prov, dir, noopLogger{}, false, false)
	require.NoError(t, err)

	mise, err := os.ReadFile(filepath.Join(dir, "mise.toml"))
	require.NoError(t, err)
	assert.Contains(t, string(mise), "node = '22'")
}

func TestSync_PreservesNativeSections(t *testing.T) {
	dir := t.TempDir()
	// Pre-existing mise.toml with an [env] section that must survive.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "mise.toml"),
		[]byte("[env]\nNODE_ENV = \"production\"\n\n[tools]\npython = \"3.11\"\n"), 0644))

	rf := writeRazdfile(t, dir, []string{"node@22"})
	prov := provisioner.NewMiseProvisioner(provisioner.Config{Dir: dir})

	err := Sync(rf, prov, dir, noopLogger{}, false, false)
	require.NoError(t, err)

	mise, err := os.ReadFile(filepath.Join(dir, "mise.toml"))
	require.NoError(t, err)
	// [env] preserved.
	assert.Contains(t, string(mise), "NODE_ENV")
	assert.Contains(t, string(mise), "production")
	// Both node (from razd) and python (pre-existing) present.
	assert.Contains(t, string(mise), "node = '22'")
	assert.Contains(t, string(mise), "python = '3.11'")
}

func TestSync_NoDependenciesSkips(t *testing.T) {
	dir := t.TempDir()
	rf := &ast.Razdfile{Version: "1"} // no dependencies
	prov := provisioner.NewMiseProvisioner(provisioner.Config{Dir: dir})

	err := Sync(rf, prov, dir, noopLogger{}, false, false)
	require.NoError(t, err)
	// No mise.toml should be created.
	_, statErr := os.Stat(filepath.Join(dir, "mise.toml"))
	assert.True(t, os.IsNotExist(statErr))
}

func TestApplyToRazdfile_ReplacesNotDuplicates(t *testing.T) {
	dir := t.TempDir()
	// Existing bun@latest entry; apply bun@1 (from native, "use native").
	rf := writeRazdfile(t, dir, []string{"bun@latest"})

	err := applyToRazdfile(rf, []Tool{{Name: "bun", Version: "1"}}, dir, noopLogger{})
	require.NoError(t, err)

	razd, err := os.ReadFile(filepath.Join(dir, "Razdfile.yml"))
	require.NoError(t, err)
	// The bun entry must have been replaced, not duplicated.
	assert.Contains(t, string(razd), "bun@1")
	assert.NotContains(t, string(razd), "bun@latest")
}

func TestApplyToRazdfile_VersionlessWritesBareEntry(t *testing.T) {
	dir := t.TempDir()
	// A versionless native devbox tool must be written as a bare name,
	// never with a trailing "@".
	rf := writeRazdfile(t, dir, []string{"nodejs@22"})

	err := applyToRazdfile(rf, []Tool{{Name: "php84Extensions.xdebug", Version: ""}}, dir, noopLogger{})
	require.NoError(t, err)

	razd, err := os.ReadFile(filepath.Join(dir, "Razdfile.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(razd), "php84Extensions.xdebug")
	assert.NotContains(t, string(razd), "php84Extensions.xdebug@")
	assert.Contains(t, string(razd), "nodejs@22")
}

func TestSync_DevboxVersionlessAddedToRazdfile(t *testing.T) {
	dir := t.TempDir()
	// Native devbox.json has a versionless php84 package and a versioned one.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "devbox.json"),
		[]byte(`{"packages": ["php84Extensions.xdebug", "php@8.4.15"]}`), 0644))
	// Razdfile has only the versioned tool.
	rf := writeRazdfile(t, dir, []string{"php@8.4.15"})
	prov := provisioner.NewDevboxProvisioner(provisioner.Config{Dir: dir})

	err := Sync(rf, prov, dir, noopLogger{}, false, false)
	require.NoError(t, err)

	// The versionless package must be mirrored into Razdfile.ensure as a bare entry.
	razd, err := os.ReadFile(filepath.Join(dir, "Razdfile.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(razd), "php84Extensions.xdebug")
	assert.NotContains(t, string(razd), "php84Extensions.xdebug@")
}


func TestSync_ConflictNonInteractiveSkips(t *testing.T) {
	dir := t.TempDir()
	// Native has node@23, Razdfile has node@22 -> conflict.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "mise.toml"),
		[]byte("[tools]\nnode = \"23\"\n"), 0644))
	rf := writeRazdfile(t, dir, []string{"node@22"})
	prov := provisioner.NewMiseProvisioner(provisioner.Config{Dir: dir})

	err := Sync(rf, prov, dir, noopLogger{}, false, false)
	require.NoError(t, err)

	// Non-TTY -> Skip: neither side changes.
	mise, err := os.ReadFile(filepath.Join(dir, "mise.toml"))
	require.NoError(t, err)
	assert.Contains(t, string(mise), "node = \"23\"")
	razd, err := os.ReadFile(filepath.Join(dir, "Razdfile.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(razd), "node@22")
}

func TestSync_ConfirmNonInteractiveApplies(t *testing.T) {
	dir := t.TempDir()
	// Native has node@22, Razdfile has go@1.21 (no node). With confirmation
	// enabled but a non-interactive stdin (CI), the confirm prompt returns
	// ApplyNonInteractive and the change is still applied (no data loss).
	require.NoError(t, os.WriteFile(filepath.Join(dir, "mise.toml"),
		[]byte("[tools]\nnode = \"22\"\ngo = \"1.21\"\n"), 0644))
	rf := writeRazdfile(t, dir, []string{"go@1.21"})
	prov := provisioner.NewMiseProvisioner(provisioner.Config{Dir: dir})

	err := Sync(rf, prov, dir, noopLogger{}, false, true)
	require.NoError(t, err)

	// node@22 was mirrored into the Razdfile despite the confirm flag, because
	// stdin is not a TTY under go test.
	razd, err := os.ReadFile(filepath.Join(dir, "Razdfile.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(razd), "node@22")
}

func TestSync_NoChangesNoConfirm(t *testing.T) {
	dir := t.TempDir()
	// Both sides identical -> no changes, confirm prompt must not be reached.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "mise.toml"),
		[]byte("[tools]\nnode = \"22\"\n"), 0644))
	rf := writeRazdfile(t, dir, []string{"node@22"})
	prov := provisioner.NewMiseProvisioner(provisioner.Config{Dir: dir})

	err := Sync(rf, prov, dir, noopLogger{}, false, true)
	require.NoError(t, err)

	// Neither side changed.
	mise, err := os.ReadFile(filepath.Join(dir, "mise.toml"))
	require.NoError(t, err)
	assert.Contains(t, string(mise), "node = \"22\"")
	razd, err := os.ReadFile(filepath.Join(dir, "Razdfile.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(razd), "node@22")
}

func TestSync_NoNativeFileSkipsPromptAndWritesNative(t *testing.T) {
	dir := t.TempDir()
	// No mise.toml exists. With confirmSync=true, the direction prompt must be
	// skipped because there is nothing to reconcile against: Razdfile is the
	// only source, so the native file is written from it directly.
	rf := writeRazdfile(t, dir, []string{"node@22"})
	prov := provisioner.NewMiseProvisioner(provisioner.Config{Dir: dir})

	err := Sync(rf, prov, dir, noopLogger{}, false, true)
	require.NoError(t, err)

	mise, err := os.ReadFile(filepath.Join(dir, "mise.toml"))
	require.NoError(t, err)
	assert.Contains(t, string(mise), "node = '22'")
}

func TestSync_ExistingNativeFileStillPrompts(t *testing.T) {
	dir := t.TempDir()
	// A native config exists with a different tool. The native -> Razdfile
	// direction remains possible, so the sync still reconciles (existing
	// behavior unchanged, non-interactive stdin applies changes as before).
	require.NoError(t, os.WriteFile(filepath.Join(dir, "mise.toml"),
		[]byte("[tools]\npython = \"3.11\"\n"), 0644))
	rf := writeRazdfile(t, dir, []string{"node@22"})
	prov := provisioner.NewMiseProvisioner(provisioner.Config{Dir: dir})

	err := Sync(rf, prov, dir, noopLogger{}, false, true)
	require.NoError(t, err)

	// node pushed from Razdfile to native.
	mise, err := os.ReadFile(filepath.Join(dir, "mise.toml"))
	require.NoError(t, err)
	assert.Contains(t, string(mise), "node = '22'")
	// python (pre-existing native) mirrored into Razdfile.ensure.
	razd, err := os.ReadFile(filepath.Join(dir, "Razdfile.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(razd), "python@3.11")
}

func TestSync_NewPackageNoDirectionPrompt(t *testing.T) {
	dir := t.TempDir()
	// mise.toml has go@1.21 which is already in Razdfile; Razdfile adds python.
	// With confirmSync=true, the direction prompt must be skipped because the
	// only change is Razdfile->native (no native->Razdfile changes).
	require.NoError(t, os.WriteFile(filepath.Join(dir, "mise.toml"),
		[]byte("[tools]\ngo = \"1.21\"\n"), 0644))
	rf := writeRazdfile(t, dir, []string{"go@1.21", "python"})
	prov := provisioner.NewMiseProvisioner(provisioner.Config{Dir: dir})

	err := Sync(rf, prov, dir, noopLogger{}, false, true)
	require.NoError(t, err)

	// python written to native without a direction prompt.
	mise, err := os.ReadFile(filepath.Join(dir, "mise.toml"))
	require.NoError(t, err)
	assert.Contains(t, string(mise), "python")
	// go (already in both) unchanged.
	assert.Contains(t, string(mise), "go = '1.21'")
	// No native->Razdfile changes were applied.
	razd, err := os.ReadFile(filepath.Join(dir, "Razdfile.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(razd), "python")
	assert.NotContains(t, string(razd), "node")
}

func TestSync_ConflictAppliesWithoutSecondPrompt(t *testing.T) {
	dir := t.TempDir()
	// Native has node@24, Razdfile has node@26 -> version conflict. With
	// confirmSync=true and non-interactive stdin, the conflict resolves to
	// skip (safe default) and no direction prompt fires.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "mise.toml"),
		[]byte("[tools]\nnode = \"24\"\n"), 0644))
	rf := writeRazdfile(t, dir, []string{"node@26"})
	prov := provisioner.NewMiseProvisioner(provisioner.Config{Dir: dir})

	err := Sync(rf, prov, dir, noopLogger{}, false, true)
	require.NoError(t, err)

	// Conflict skipped (non-interactive): neither side changes.
	mise, err := os.ReadFile(filepath.Join(dir, "mise.toml"))
	require.NoError(t, err)
	assert.Contains(t, string(mise), "node = \"24\"")
	razd, err := os.ReadFile(filepath.Join(dir, "Razdfile.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(razd), "node@26")
}

