package sync

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

// ApplyDecision is the direction of sync chosen in the confirm prompt.
type ApplyDecision int

const (
	// ApplyFromNative means sync native config -> Razdfile (mirror native
	// packages into dependencies.ensure).
	ApplyFromNative ApplyDecision = iota
	// ApplyFromRazdfile means sync Razdfile -> native config (write
	// dependencies.ensure into mise.toml / devbox.json).
	ApplyFromRazdfile
	// ApplySkip means leave both files unchanged.
	ApplySkip
	// ApplyNonInteractive means stdin is not a TTY; the caller decides the default.
	ApplyNonInteractive
)

// PromptConfirmSync asks the user which direction to sync pending changes
// between the Razdfile and the native config. It is only shown when there are
// actual changes to apply. In non-interactive mode it returns
// ApplyNonInteractive without blocking, so callers can apply changes as before
// (CI-safe default).
func PromptConfirmSync(provName string, addedToRazdfile, addedToNative int, log Logger) (ApplyDecision, error) {
	if !isTerminal() {
		log.Debugf("[SYNC] confirm prompt skipped (non-interactive)\n")
		return ApplyNonInteractive, nil
	}

	desc := fmt.Sprintf("Razdfile <-> %s: %d package(s) from native, %d package(s) from Razdfile",
		provName, addedToRazdfile, addedToNative)

	var choice int
	options := []string{
		fmt.Sprintf("Sync from %s config (native -> Razdfile)", provName),
		"Sync from Razdfile (Razdfile -> native config)",
		"Skip (keep both unchanged)",
	}

	prompt := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("How to synchronize config?").
				Description(desc).
				Options(
					huh.NewOption(options[0], int(ApplyFromNative)),
					huh.NewOption(options[1], int(ApplyFromRazdfile)),
					huh.NewOption(options[2], int(ApplySkip)),
				).
				Value(&choice),
		),
	).WithTheme(huh.ThemeCatppuccin())

	if err := prompt.Run(); err != nil {
		if err == huh.ErrUserAborted {
			log.Debugf("[SYNC] confirm prompt aborted\n")
			return ApplySkip, nil
		}
		return ApplySkip, fmt.Errorf("confirm prompt failed: %w", err)
	}

	switch ApplyDecision(choice) {
	case ApplyFromNative:
		log.Infof("[SYNC] syncing from %s config (native -> Razdfile)\n", provName)
		return ApplyFromNative, nil
	case ApplyFromRazdfile:
		log.Infof("[SYNC] syncing from Razdfile (Razdfile -> %s config)\n", provName)
		return ApplyFromRazdfile, nil
	default:
		log.Infof("[SYNC] skipping config sync\n")
		return ApplySkip, nil
	}
}
