package cli

import (
	"errors"
	"testing"

	apperrors "github.com/razd-cli/razd/internal/errors"
	"github.com/razd-cli/razd/internal/git"
	"github.com/razd-cli/razd/internal/output"
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

	if dir != "/tmp/testdir/my-repo" {
		t.Errorf("resolveUpDir = %q, want /tmp/testdir/my-repo", dir)
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