package trust

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/razd-cli/razd/internal/output"
)

// PromptTrustResult represents the result of a trust prompt interaction.
type PromptTrustResult int

const (
	// PromptTrusted means the user selected "Yes" in the prompt.
	PromptTrusted PromptTrustResult = iota
	// PromptDeclined means the user selected "No" in the prompt.
	PromptDeclined
	// PromptNonInteractive means stdin is not a TTY (CI, pipe, etc.).
	PromptNonInteractive
)

// PromptTrust shows an interactive trust confirmation prompt.
// If stdin is not a TTY, it returns PromptNonInteractive without blocking.
// The user can select "No" (default) or "Yes" using arrow keys or Y/N keys.
func PromptTrust(path string, log *output.Logger) (PromptTrustResult, error) {
	if !IsTerminal() {
		log.Debugf("[FIX] PromptTrust: non-interactive mode, path=%s\n", path)
		return PromptNonInteractive, nil
	}

	var confirm bool
	prompt := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Trust this project?").
				Description(path).
				Affirmative("Yes").
				Negative("No").
				Value(&confirm),
		),
	).WithTheme(huh.ThemeCatppuccin())

	if err := prompt.Run(); err != nil {
		if err == huh.ErrUserAborted {
			log.Debugf("[FIX] PromptTrust: user aborted, path=%s\n", path)
			return PromptDeclined, nil
		}
		return PromptDeclined, fmt.Errorf("trust prompt failed: %w", err)
	}

	if confirm {
		log.Debugf("[FIX] PromptTrust: trusted, path=%s\n", path)
		return PromptTrusted, nil
	}

	log.Debugf("[FIX] PromptTrust: declined, path=%s\n", path)
	return PromptDeclined, nil
}