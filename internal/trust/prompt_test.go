package trust

import (
	"io"
	"testing"

	"github.com/razd-cli/razd/internal/output"
)

func TestIsTerminal(t *testing.T) {
	result := IsTerminal()
	t.Logf("IsTerminal() = %v (environment-dependent)", result)
}

func TestPromptTrustResult_Constants(t *testing.T) {
	if PromptTrusted != 0 {
		t.Errorf("PromptTrusted = %d, want 0", PromptTrusted)
	}
	if PromptDeclined != 1 {
		t.Errorf("PromptDeclined = %d, want 1", PromptDeclined)
	}
	if PromptNonInteractive != 2 {
		t.Errorf("PromptNonInteractive = %d, want 2", PromptNonInteractive)
	}
}

func TestPromptTrust_NonTerminal(t *testing.T) {
	log := output.NewLogger(io.Discard, io.Discard)
	result, err := PromptTrust("/tmp/test-project", log)
	if err != nil {
		t.Fatalf("PromptTrust() error = %v", err)
	}
	if result != PromptNonInteractive {
		t.Errorf("PromptTrust() in non-TTY = %v, want PromptNonInteractive", result)
	}
}

func TestEnsureTrusted_StatusTrusted(t *testing.T) {
	log := output.NewLogger(io.Discard, io.Discard)

	store := &Store{
		Trusted: []string{},
		Ignored: []string{},
		path:    t.TempDir() + "/trust.json",
	}
	if err := store.Save(); err != nil {
		t.Fatal(err)
	}

	origGetStorePath := getStorePath
	getStorePath = func() (string, error) { return store.path, nil }
	defer func() { getStorePath = origGetStorePath }()

	testPath := t.TempDir()
	store.AddTrusted(testPath)
	if err := store.Save(); err != nil {
		t.Fatal(err)
	}

	trusted, err := EnsureTrusted(testPath, nil, log, false)
	if err != nil {
		t.Fatalf("EnsureTrusted() error = %v", err)
	}
	if !trusted {
		t.Error("EnsureTrusted() = false for trusted project, want true")
	}
}

func TestEnsureTrusted_AutoTrust(t *testing.T) {
	log := output.NewLogger(io.Discard, io.Discard)

	store := &Store{
		Trusted: []string{},
		Ignored: []string{},
		path:    t.TempDir() + "/trust.json",
	}
	if err := store.Save(); err != nil {
		t.Fatal(err)
	}

	origGetStorePath := getStorePath
	getStorePath = func() (string, error) { return store.path, nil }
	defer func() { getStorePath = origGetStorePath }()

	testPath := t.TempDir()

	trusted, err := EnsureTrusted(testPath, nil, log, true)
	if err != nil {
		t.Fatalf("EnsureTrusted() error = %v", err)
	}
	if !trusted {
		t.Error("EnsureTrusted() with autoTrust=true = false, want true")
	}
}