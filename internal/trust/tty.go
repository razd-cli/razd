package trust

import (
	"os"

	"golang.org/x/term"
)

// IsTerminal returns true if stdin is connected to a terminal (TTY).
// Returns false for pipes, redirected input, CI environments, etc.
func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}