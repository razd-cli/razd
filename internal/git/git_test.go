package git

import (
	"errors"
	"testing"
)

func TestIsURL(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want bool
	}{
		{"https URL", "https://github.com/razd-cli/razd-nodejs-example", true},
		{"https URL with .git", "https://github.com/razd-cli/razd-nodejs-example.git", true},
		{"SSH URL", "git@github.com:razd-cli/razd-nodejs-example.git", true},
		{"ssh protocol URL", "ssh://git@github.com/razd-cli/razd-nodejs-example.git", true},
		{"git protocol URL", "git://github.com/razd-cli/razd-nodejs-example.git", true},
		{"http URL", "http://github.com/razd-cli/razd-nodejs-example", true},
		{"task name dev", "dev", false},
		{"task name build", "build", false},
		{"local path", "./local/path", false},
		{"empty string", "", false},
		{"dot path", ".", false},
		{"double dot", "..", false},
		{"just .git suffix", "something.git", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsURL(tt.arg); got != tt.want {
				t.Errorf("IsURL(%q) = %v, want %v", tt.arg, got, tt.want)
			}
		})
	}
}

func TestExtractRepoName(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{"https without .git", "https://github.com/razd-cli/razd-nodejs-example", "razd-nodejs-example", false},
		{"https with .git", "https://github.com/razd-cli/razd-nodejs-example.git", "razd-nodejs-example", false},
		{"SSH URL", "git@github.com:razd-cli/razd-nodejs-example.git", "razd-nodejs-example", false},
		{"ssh protocol", "ssh://git@github.com/razd-cli/razd-nodejs-example.git", "razd-nodejs-example", false},
		{"git protocol", "git://github.com/razd-cli/razd-nodejs-example.git", "razd-nodejs-example", false},
		{"http URL", "http://github.com/razd-cli/razd-nodejs-example", "razd-nodejs-example", false},
		{"empty string", "", "", true},
		{"nested path", "https://github.com/org/repo-name", "repo-name", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractRepoName(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractRepoName(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExtractRepoName(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestIsGitAvailable(t *testing.T) {
	original := IsGitAvailableFunc
	defer func() { IsGitAvailableFunc = original }()

	IsGitAvailableFunc = func() bool { return true }
	if !IsGitAvailable() {
		t.Error("IsGitAvailable() = false, want true (mocked)")
	}

	IsGitAvailableFunc = func() bool { return false }
	if IsGitAvailable() {
		t.Error("IsGitAvailable() = true, want false (mocked)")
	}
}

func TestClone_Mock(t *testing.T) {
	original := CloneFunc
	defer func() { CloneFunc = original }()

	CloneFunc = func(url string, dir string) error {
		return nil
	}
	if err := Clone("https://github.com/test/repo.git", "/tmp"); err != nil {
		t.Errorf("Clone() mock unexpectedly returned error: %v", err)
	}

	CloneFunc = func(url string, dir string) error {
		return errors.New("git clone failed")
	}
	if err := Clone("https://github.com/test/repo.git", "/tmp"); err == nil {
		t.Error("Clone() mock expected error, got nil")
	}
}