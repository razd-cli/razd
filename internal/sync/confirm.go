package sync

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

// ApplyDecision is the outcome of the "confirm sync" prompt.
type ApplyDecision int

const (
	// ApplyYes means the user chose to apply the sync changes.
	ApplyYes ApplyDecision = iota
	// ApplyNo means the user declined; nothing is written.
	ApplyNo
	// ApplyNonInteractive means stdin is not a TTY; the caller decides the default.
	ApplyNonInteractive
)

// PromptConfirmSync asks the user whether to apply pending sync changes between
// the Razdfile and the native config. It is only shown when there are actual
// changes to apply. In non-interactive mode it returns ApplyNonInteractive
// without blocking, so callers can apply changes as before (CI-safe default).
func PromptConfirmSync(provName string, addedToRazdfile, addedToNative int, log Logger) (ApplyDecision, error) {
	if !isTerminal() {
		log.Debugf("[SYNC] confirm prompt skipped (non-interactive)\n")
		return ApplyNonInteractive, nil
	}

	var apply bool
	desc := fmt.Sprintf("Razdfile <-> %s: %d package(s) to Razdfile, %d package(s) to %s config",
		provName, addedToRazdfile, addedToNative, provName)

	prompt := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Apply config sync changes?").
				Description(desc).
				Affirmative("Apply").
				Negative("Skip").
				Value(&apply),
		),
	).WithTheme(huh.ThemeCatppuccin())

	if err := prompt.Run(); err != nil {
		if err == huh.ErrUserAborted {
			log.Debugf("[SYNC] confirm prompt aborted\n")
			return ApplyNo, nil
		}
		return ApplyNo, fmt.Errorf("confirm prompt failed: %w", err)
	}

	if apply {
		log.Infof("[SYNC] applying config sync changes\n")
		return ApplyYes, nil
	}

	log.Infof("[SYNC] skipping config sync changes\n")
	return ApplyNo, nil
}
