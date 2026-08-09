package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/huh"
	"golang.org/x/term"
)

// Logger is the minimal logging surface sync needs. The concrete
// internal/output.Logger satisfies it.
type Logger interface {
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Successf(format string, args ...any)
}

// isTerminal reports whether stdin is a TTY (interactive terminal).
func isTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// BackupFile writes a timestamped copy of srcPath next to it and returns the
// backup path. It is a no-op (returns "") if the source file does not exist.
func BackupFile(srcPath string, log Logger) (string, error) {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Debugf("[SYNC] backup: %s does not exist, skipping\n", srcPath)
			return "", nil
		}
		return "", fmt.Errorf("failed to read %s for backup: %w", srcPath, err)
	}

	ts := time.Now().Format("20060102-150405")
	dir := filepath.Dir(srcPath)
	base := filepath.Base(srcPath)

	// Ensure a unique backup path even for multiple backups within the same
	// second: append a counter suffix on collision.
	backupPath := filepath.Join(dir, fmt.Sprintf("%s.bak.%s", base, ts))
	for i := 1; ; i++ {
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			break
		}
		backupPath = filepath.Join(dir, fmt.Sprintf("%s.bak.%s.%d", base, ts, i))
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write backup %s: %w", backupPath, err)
	}

	log.Infof("[SYNC] Backed up %s to %s\n", srcPath, backupPath)
	return backupPath, nil
}

// BackupDecision is the outcome of the interactive backup prompt.
type BackupDecision int

const (
	// BackupYes means the user chose to create a backup.
	BackupYes BackupDecision = iota
	// BackupNo means the user declined a backup.
	BackupNo
	// BackupNonInteractive means stdin is not a TTY; caller should pick a safe default.
	BackupNonInteractive
)

// PromptBackup asks the user whether to back up a config file before it is
// overwritten. The default is No. In non-interactive mode it returns
// BackupNonInteractive without blocking.
func PromptBackup(configPath string, log Logger) (BackupDecision, error) {
	if !isTerminal() {
		log.Debugf("[SYNC] backup prompt skipped (non-interactive), path=%s\n", configPath)
		return BackupNonInteractive, nil
	}

	var confirm bool
	prompt := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Make backup of " + filepath.Base(configPath) + " before overwrite?").
				Description(configPath).
				Affirmative("Yes").
				Negative("No").
				Value(&confirm),
		),
	).WithTheme(huh.ThemeCatppuccin())

	if err := prompt.Run(); err != nil {
		if err == huh.ErrUserAborted {
			log.Debugf("[SYNC] backup prompt aborted, path=%s\n", configPath)
			return BackupNo, nil
		}
		return BackupNo, fmt.Errorf("backup prompt failed: %w", err)
	}

	if confirm {
		log.Debugf("[SYNC] user chose to back up %s\n", configPath)
		return BackupYes, nil
	}
	log.Debugf("[SYNC] user declined backup for %s\n", configPath)
	return BackupNo, nil
}
