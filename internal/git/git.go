package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	IsGitAvailableFunc = func() bool {
		_, err := exec.LookPath("git")
		return err == nil
	}
	CloneFunc = func(url string, dir string) error {
		cmd := exec.Command("git", "clone", url)
		cmd.Dir = dir
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.Stdin = nil

		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("git clone failed: %s", string(output))
		}

		return nil
	}
)

func IsGitAvailable() bool {
	return IsGitAvailableFunc()
}

func Clone(url string, dir string) error {
	return CloneFunc(url, dir)
}

func IsURL(arg string) bool {
	if arg == "" {
		return false
	}

	if strings.HasPrefix(arg, "https://") ||
		strings.HasPrefix(arg, "http://") ||
		strings.HasPrefix(arg, "git@") ||
		strings.HasPrefix(arg, "ssh://") ||
		strings.HasPrefix(arg, "git://") {
		return true
	}

	if strings.HasSuffix(arg, ".git") {
		return true
	}

	return false
}

func ExtractRepoName(rawURL string) (string, error) {
	if rawURL == "" {
		return "", fmt.Errorf("empty URL")
	}

	s := rawURL

	if strings.HasPrefix(s, "git@") {
		s = strings.TrimPrefix(s, "git@")
		colonIdx := strings.Index(s, ":")
		if colonIdx >= 0 {
			s = s[colonIdx+1:]
		}
	} else {
		schemeEnd := strings.Index(s, "://")
		if schemeEnd >= 0 {
			s = s[schemeEnd+3:]
		}
	}

	s = strings.TrimSuffix(s, ".git")

	s = filepath.Base(s)

	if s == "" || s == "." || s == "/" {
		return "", fmt.Errorf("could not extract repository name from %q", rawURL)
	}

	return s, nil
}