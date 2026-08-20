package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	apperrors "github.com/razd-cli/razd/internal/errors"
	"github.com/razd-cli/razd/internal/git"
	"github.com/razd-cli/razd/internal/output"
	"github.com/razd-cli/razd/razdfile/ast"
)

func TestResolveUpDir_NoArgs(t *testing.T) {
	ctx := &Context{
		Args: []string{},
		Log:  output.NewLogger(&testWriter{}, &testWriter{}),
	}

	dir, err := resolveUpDir(ctx)
	if err != nil {
		t.Fatalf("resolveUpDir with no args should not error, got: %v", err)
	}
	if dir == "" {
		t.Error("resolveUpDir with no args should return a directory")
	}
}

func TestResolveUpDir_TooManyArgs(t *testing.T) {
	ctx := &Context{
		Args: []string{"https://github.com/test/repo", "extra"},
		Log:  output.NewLogger(&testWriter{}, &testWriter{}),
	}

	_, err := resolveUpDir(ctx)
	if err == nil {
		t.Fatal("resolveUpDir with too many args should return error")
	}
	if err.Error() != "too many arguments for 'up' command" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestResolveUpDir_NonURLArg(t *testing.T) {
	ctx := &Context{
		Args: []string{"dev"},
		Log:  output.NewLogger(&testWriter{}, &testWriter{}),
	}

	_, err := resolveUpDir(ctx)
	if err == nil {
		t.Fatal("resolveUpDir with non-URL arg should return error")
	}
	expected := `unexpected argument "dev" for 'up' command`
	if err.Error() != expected {
		t.Errorf("unexpected error message: got %q, want %q", err.Error(), expected)
	}
}

func TestResolveUpDir_GitNotInstalled(t *testing.T) {
	origGitAvailable := git.IsGitAvailableFunc
	defer func() { git.IsGitAvailableFunc = origGitAvailable }()

	git.IsGitAvailableFunc = func() bool { return false }

	ctx := &Context{
		Args: []string{"https://github.com/test/repo"},
		Log:  output.NewLogger(&testWriter{}, &testWriter{}),
	}

	_, err := resolveUpDir(ctx)
	if err == nil {
		t.Fatal("resolveUpDir should return error when git is not installed")
	}

	var gitErr *apperrors.GitNotInstalledError
	if !errors.As(err, &gitErr) {
		t.Errorf("expected GitNotInstalledError, got %T: %v", err, err)
	}
}

func TestResolveUpDir_CloneFailed(t *testing.T) {
	origGitAvailable := git.IsGitAvailableFunc
	origClone := git.CloneFunc
	defer func() {
		git.IsGitAvailableFunc = origGitAvailable
		git.CloneFunc = origClone
	}()

	git.IsGitAvailableFunc = func() bool { return true }
	git.CloneFunc = func(url string, dir string) error {
		return errors.New("fatal: repository not found")
	}

	ctx := &Context{
		Args: []string{"https://github.com/test/nonexistent"},
		Log:  output.NewLogger(&testWriter{}, &testWriter{}),
	}

	_, err := resolveUpDir(ctx)
	if err == nil {
		t.Fatal("resolveUpDir should return error when clone fails")
	}

	var cloneErr *apperrors.CloneError
	if !errors.As(err, &cloneErr) {
		t.Errorf("expected CloneError, got %T: %v", err, err)
	}
}

func TestResolveUpDir_CloneSuccess(t *testing.T) {
	origGitAvailable := git.IsGitAvailableFunc
	origClone := git.CloneFunc
	defer func() {
		git.IsGitAvailableFunc = origGitAvailable
		git.CloneFunc = origClone
	}()

	git.IsGitAvailableFunc = func() bool { return true }
	git.CloneFunc = func(url string, dir string) error {
		return nil
	}

	ctx := &Context{
		Args: []string{"https://github.com/test/my-repo"},
		Log:  output.NewLogger(&testwriterSimple{}, &testwriterSimple{}),
	}

	dir, err := resolveUpDir(ctx)
	if err != nil {
		t.Fatalf("resolveUpDir should not error on successful clone, got: %v", err)
	}

	expectedSuffix := "my-repo"
	if dir == "" {
		t.Error("resolveUpDir should return a directory path")
	}
	if len(dir) < len(expectedSuffix) || dir[len(dir)-len(expectedSuffix):] != expectedSuffix {
		t.Errorf("resolveUpDir returned dir = %q, want path ending with %q", dir, expectedSuffix)
	}
}

func TestResolveUpDir_CloneSuccessWithDir(t *testing.T) {
	origGitAvailable := git.IsGitAvailableFunc
	origClone := git.CloneFunc
	defer func() {
		git.IsGitAvailableFunc = origGitAvailable
		git.CloneFunc = origClone
	}()

	git.IsGitAvailableFunc = func() bool { return true }
	git.CloneFunc = func(url string, dir string) error {
		return nil
	}

	ctx := &Context{
		Args: []string{"https://github.com/test/my-repo"},
		Dir:  "/tmp/testdir",
		Log:  output.NewLogger(&testwriterSimple{}, &testwriterSimple{}),
	}

	dir, err := resolveUpDir(ctx)
	if err != nil {
		t.Fatalf("resolveUpDir should not error, got: %v", err)
	}

	expected := filepath.Join("/tmp/testdir", "my-repo")
	if dir != expected {
		t.Errorf("resolveUpDir = %q, want %q", dir, expected)
	}
}

type testWriter struct{}

func (t *testWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

type testwriterSimple struct{}

func (t *testwriterSimple) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// fakeProvisioner is a minimal provisioner.Provisioner stub used to exercise
// the up flow without a real mise/devbox binary on PATH.
type fakeProvisioner struct {
	name       string
	installCtx *context.Context
	trusted    bool
	addedTools map[string]string
}

func (f *fakeProvisioner) Name() string                       { return f.name }
func (f *fakeProvisioner) GenerateConfig(_ []ast.ParsedDependency, _ map[string]any) error {
	return nil
}
func (f *fakeProvisioner) ReadConfig() (map[string]string, error) { return nil, nil }
func (f *fakeProvisioner) WriteTools(_ map[string]string) error   { return nil }
func (f *fakeProvisioner) AddTools(_ context.Context, tools map[string]string) error {
	f.addedTools = tools
	return nil
}
func (f *fakeProvisioner) Install(ctx context.Context) error {
	f.installCtx = &ctx
	return nil
}
func (f *fakeProvisioner) RunCommand(cmd []string) []string {
	// Passthrough so tests can actually execute commands without a real mise.
	return cmd
}
func (f *fakeProvisioner) Shell(_ context.Context) error   { return nil }
func (f *fakeProvisioner) Trust(_ context.Context) error   { f.trusted = true; return nil }
func (f *fakeProvisioner) Untrust(_ context.Context) error { return nil }
func (f *fakeProvisioner) IsAvailable() bool               { return true }

// TestRunDefaultTask_ExecutesDefaultTask verifies the core change: runUp now
// unconditionally invokes runDefaultTask, which executes the default task's
// commands after installation. A marker file proves the command ran.
func TestRunDefaultTask_ExecutesDefaultTask(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "marker.txt")
	content := "version: \"1\"\ntasks:\n  default:\n    cmds:\n      - touch " + marker + "\n"
	if err := os.WriteFile(filepath.Join(dir, "Razdfile.yml"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write Razdfile: %v", err)
	}

	rf, err := readRazdfile(dir, output.NewLogger(&testWriter{}, &testWriter{}))
	if err != nil {
		t.Fatalf("readRazdfile error: %v", err)
	}

	prov := &fakeProvisioner{name: "mise"}
	ctx := &Context{
		Dir: dir,
		Log: output.NewLogger(&testWriter{}, &testWriter{}),
	}

	if err := runDefaultTask(ctx, rf, prov, dir); err != nil {
		t.Fatalf("runDefaultTask error: %v", err)
	}

	if _, err := os.Stat(marker); err != nil {
		t.Errorf("default task command did not run; marker %s missing: %v", marker, err)
	}
}

// TestRunDefaultTask_NoDefaultTaskWarns verifies the fallback: when the
// Razdfile has tasks but no 'default' task, runDefaultTask returns nil without
// error, so up completes after installing ensure packages.
func TestRunDefaultTask_NoDefaultTaskWarns(t *testing.T) {
	dir := t.TempDir()
	content := "version: \"1\"\ntasks:\n  build:\n    cmds:\n      - echo \"build\"\n"
	if err := os.WriteFile(filepath.Join(dir, "Razdfile.yml"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write Razdfile: %v", err)
	}

	rf, err := readRazdfile(dir, output.NewLogger(&testWriter{}, &testWriter{}))
	if err != nil {
		t.Fatalf("readRazdfile error: %v", err)
	}

	prov := &fakeProvisioner{name: "mise"}
	ctx := &Context{
		Dir: dir,
		Log: output.NewLogger(&testWriter{}, &testWriter{}),
	}

	if err := runDefaultTask(ctx, rf, prov, dir); err != nil {
		t.Fatalf("runDefaultTask with no default task should not error, got: %v", err)
	}
}

// TestRunDefaultTask_NoTasksWarns verifies the fallback for a Razdfile with
// dependencies but no tasks at all: runDefaultTask warns and returns nil,
// so up completes after installing ensure packages.
func TestRunDefaultTask_NoTasksWarns(t *testing.T) {
	dir := t.TempDir()
	content := "version: \"1\"\ndependencies:\n  using: mise\n  ensure:\n    - \"node@22\"\n"
	if err := os.WriteFile(filepath.Join(dir, "Razdfile.yml"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write Razdfile: %v", err)
	}

	rf, err := readRazdfile(dir, output.NewLogger(&testWriter{}, &testWriter{}))
	if err != nil {
		t.Fatalf("readRazdfile error: %v", err)
	}

	prov := &fakeProvisioner{name: "mise"}
	ctx := &Context{
		Dir: dir,
		Log: output.NewLogger(&testWriter{}, &testWriter{}),
	}

	if err := runDefaultTask(ctx, rf, prov, dir); err != nil {
		t.Fatalf("runDefaultTask with no tasks should not error, got: %v", err)
	}
}