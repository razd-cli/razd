package sync

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

// PromptConflict asks the user how to resolve a version conflict between the
// Razdfile and the native config.
//
// Options:
//  1. Use Razdfile version
//  2. Use native config version
//  3. Skip (leave both unchanged)
//
// In non-interactive mode it returns ResolutionSkip (the safe default: no data
// is overwritten).
func PromptConflict(tool string, razdVersion, nativeVersion string, log Logger) (Resolution, error) {
	if !isTerminal() {
		log.Warnf("[SYNC] version conflict for %s (Razdfile=%s, native=%s) skipped (non-interactive)\n",
			tool, razdVersion, nativeVersion)
		return ResolutionSkip, nil
	}

	var choice int
	options := []string{
		fmt.Sprintf("Use Razdfile (%s)", razdVersion),
		fmt.Sprintf("Use %s config (%s)", "native", nativeVersion),
		"Skip (keep both unchanged)",
	}

	prompt := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[int]().
				Title(fmt.Sprintf("Version conflict for %s", tool)).
				Description(fmt.Sprintf("Razdfile has %s, native config has %s", razdVersion, nativeVersion)).
				Options(
					huh.NewOption(options[0], 0),
					huh.NewOption(options[1], 1),
					huh.NewOption(options[2], 2),
				).
				Value(&choice),
		),
	).WithTheme(huh.ThemeCatppuccin())

	if err := prompt.Run(); err != nil {
		if err == huh.ErrUserAborted {
			log.Debugf("[SYNC] conflict prompt aborted for %s\n", tool)
			return ResolutionSkip, nil
		}
		return ResolutionSkip, fmt.Errorf("conflict prompt failed: %w", err)
	}

	switch choice {
	case 0:
		log.Infof("[SYNC] %s: using Razdfile version %s\n", tool, razdVersion)
		return ResolutionUseRazdfile, nil
	case 1:
		log.Infof("[SYNC] %s: using native config version %s\n", tool, nativeVersion)
		return ResolutionUseNative, nil
	default:
		log.Infof("[SYNC] %s: skipping version change\n", tool)
		return ResolutionSkip, nil
	}
}
