// Package trust manages project trust for razd.
package trust

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Status represents the trust status of a project.
type Status string

const (
	StatusTrusted Status = "trusted"
	StatusIgnored Status = "ignored"
	StatusUnknown Status = "unknown"
)

// Store manages trusted and ignored projects.
type Store struct {
	Trusted []string `json:"trusted"`
	Ignored []string `json:"ignored"`

	path string
	mu   sync.RWMutex
}

// Load loads the trust store from disk.
func Load() (*Store, error) {
	storePath, err := getStorePath()
	if err != nil {
		return nil, err
	}

	store := &Store{
		Trusted: []string{},
		Ignored: []string{},
		path:    storePath,
	}

	data, err := os.ReadFile(storePath)
	if os.IsNotExist(err) {
		return store, nil
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, store); err != nil {
		return nil, err
	}

	return store, nil
}

// Save persists the trust store to disk.
func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure directory exists
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0644)
}

// GetStatus returns the trust status for a path.
func (s *Store) GetStatus(path string) Status {
	s.mu.RLock()
	defer s.mu.RUnlock()

	absPath, _ := filepath.Abs(path)

	for _, p := range s.Trusted {
		if p == absPath {
			return StatusTrusted
		}
	}
	for _, p := range s.Ignored {
		if p == absPath {
			return StatusIgnored
		}
	}
	return StatusUnknown
}

// AddTrusted adds a path to the trusted list.
func (s *Store) AddTrusted(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	absPath, _ := filepath.Abs(path)

	// Remove from ignored if present
	s.Ignored = removeString(s.Ignored, absPath)

	// Add to trusted if not present
	if !containsString(s.Trusted, absPath) {
		s.Trusted = append(s.Trusted, absPath)
	}
}

// AddIgnored adds a path to the ignored list.
func (s *Store) AddIgnored(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	absPath, _ := filepath.Abs(path)

	// Remove from trusted if present
	s.Trusted = removeString(s.Trusted, absPath)

	// Add to ignored if not present
	if !containsString(s.Ignored, absPath) {
		s.Ignored = append(s.Ignored, absPath)
	}
}

// Remove removes a path from both trusted and ignored lists.
func (s *Store) Remove(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	absPath, _ := filepath.Abs(path)
	s.Trusted = removeString(s.Trusted, absPath)
	s.Ignored = removeString(s.Ignored, absPath)
}

// ListTrusted returns all trusted paths.
func (s *Store) ListTrusted() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]string{}, s.Trusted...)
}

// ListIgnored returns all ignored paths.
func (s *Store) ListIgnored() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]string{}, s.Ignored...)
}

// getStorePath returns the path to the trust store file.
// This is a variable for testability — tests can override it.
var getStorePath = defaultStorePath

func defaultStorePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "razd", "trust.json"), nil
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) []string {
	result := make([]string, 0, len(slice))
	for _, v := range slice {
		if v != s {
			result = append(result, v)
		}
	}
	return result
}
