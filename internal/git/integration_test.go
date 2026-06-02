//go:build integration

package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCloneIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	if !IsGitAvailable() {
		t.Skip("git is not installed")
	}

	tmpDir, err := os.MkdirTemp("", "razd-clone-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	repoURL := "https://github.com/razd-cli/razd-nodejs-example.git"

	t.Logf("Cloning %s into %s", repoURL, tmpDir)
	if err := Clone(repoURL, tmpDir); err != nil {
		t.Fatalf("Clone failed: %v", err)
	}

	repoName, err := ExtractRepoName(repoURL)
	if err != nil {
		t.Fatalf("ExtractRepoName failed: %v", err)
	}
	if repoName != "razd-nodejs-example" {
		t.Errorf("ExtractRepoName = %q, want %q", repoName, "razd-nodejs-example")
	}

	repoPath := filepath.Join(tmpDir, repoName)
	razdfile := filepath.Join(repoPath, "Razdfile.yml")

	info, err := os.Stat(razdfile)
	if err != nil {
		entries, readErr := os.ReadDir(repoPath)
		if readErr != nil {
			t.Fatalf("os.Stat failed for Razdfile.yml: %v, and os.ReadDir also failed: %v", err, readErr)
		}
		t.Fatalf("os.Stat failed for Razdfile.yml: %v\nDirectory contents:", err)
		for _, e := range entries {
			t.Logf("  %s (isDir=%v)", e.Name(), e.IsDir())
		}
	}
	t.Logf("Found Razdfile.yml: isDir=%v, size=%d", info.IsDir(), info.Size())
}

func TestExtractRepoNameIntegration(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://github.com/razd-cli/razd-nodejs-example", "razd-nodejs-example"},
		{"https://github.com/razd-cli/razd-nodejs-example.git", "razd-nodejs-example"},
		{"git@github.com:razd-cli/razd-nodejs-example.git", "razd-nodejs-example"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got, err := ExtractRepoName(tt.url)
			if err != nil {
				t.Fatalf("ExtractRepoName(%q) error: %v", tt.url, err)
			}
			if got != tt.want {
				t.Errorf("ExtractRepoName(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestGitCloneAndVerifyOnDisk(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	if !IsGitAvailable() {
		t.Skip("git is not installed")
	}

	tmpDir, err := os.MkdirTemp("", "razd-clone-verify-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	repoURL := "https://github.com/razd-cli/razd-nodejs-example.git"

	if err := Clone(repoURL, tmpDir); err != nil {
		t.Fatalf("Clone failed: %v", err)
	}

	repoName, _ := ExtractRepoName(repoURL)
	repoPath := filepath.Join(tmpDir, repoName)

	razdfileNames := []string{"Razdfile.yml", "Razdfile.yaml", "razdfile.yml", "razdfile.yaml"}
	found := false
	for _, name := range razdfileNames {
		path := filepath.Join(repoPath, name)
		if _, err := os.Stat(path); err == nil {
			t.Logf("Found: %s", path)
			found = true

			data, readErr := os.ReadFile(path)
			if readErr != nil {
				t.Errorf("os.ReadFile(%s) failed: %v", path, readErr)
			} else {
				t.Logf("Razdfile content length: %d bytes", len(data))
			}
			break
		} else {
			t.Logf("os.Stat(%s) failed: %v", path, err)
		}
	}

	if !found {
		entries, _ := os.ReadDir(repoPath)
		t.Errorf("No Razdfile found in cloned repo. Directory contents:")
		for _, e := range entries {
			t.Errorf("  %s (isDir=%v)", e.Name(), e.IsDir())
		}
	}
}

func TestGitLsRemoteAvailable(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	if !IsGitAvailable() {
		t.Skip("git is not installed")
	}

	cmd := exec.Command("git", "ls-remote", "--heads", "https://github.com/razd-cli/razd-nodejs-example.git")
	if err := cmd.Run(); err != nil {
		t.Skip("cannot reach remote repository (network error or repo unavailable)")
	}
}