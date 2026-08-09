package sync

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/razd-cli/razd/provisioner"
	"github.com/razd-cli/razd/razdfile"
	"github.com/razd-cli/razd/razdfile/ast"
)

// NativeConfig returns the path to the native config file for a provisioner.
func NativeConfig(provName, dir string) string {
	switch provName {
	case "mise":
		return filepath.Join(dir, "mise.toml")
	case "devbox":
		return filepath.Join(dir, "devbox.json")
	default:
		return ""
	}
}

// Sync performs a bidirectional, non-destructive merge between the Razdfile's
// dependencies.ensure list and the native provisioner config file.
//
// Steps:
//  1. Read tools from both sides (Razdfile ensure + native ReadConfig).
//  2. Reconcile them (engine), asking the user about version conflicts.
//  3. Apply changes:
//     a. New tools from native  -> appended to Razdfile dependencies.ensure.
//     b. New tools from Razdfile -> merged into the native config, with an
//        interactive backup prompt before overwriting (unless --backup is set).
//
// forceBackup true makes the backup unconditional (skips the prompt).
func Sync(rf *ast.Razdfile, prov provisioner.Provisioner, dir string, log Logger, forceBackup bool) error {
	if !rf.HasDependencies() {
		log.Debugf("[SYNC] no dependencies section, skipping sync\n")
		return nil
	}

	// Parse Razdfile-side tools.
	razdTools, err := ensureTools(rf)
	if err != nil {
		return err
	}

	// Parse native-side tools.
	nativeTools, err := prov.ReadConfig()
	if err != nil {
		log.Debugf("[SYNC] could not read native config: %v\n", err)
		return nil
	}
	if len(nativeTools) == 0 && len(razdTools) == 0 {
		log.Debugf("[SYNC] both configs empty, nothing to sync\n")
		return nil
	}

	// Reconcile.
	changes, err := Reconcile(razdTools, toolsFromMap(nativeTools), func(tool string, razdVer, nativeVer string) (Resolution, error) {
		return PromptConflict(tool, razdVer, nativeVer, log)
	})
	if err != nil {
		return fmt.Errorf("reconcile failed: %w", err)
	}

	if !changes.HasChanges() {
		log.Debugf("[SYNC] configs already in sync\n")
		return nil
	}

	// Apply native -> Razdfile.
	if len(changes.ToRazdfile) > 0 {
		if err := applyToRazdfile(rf, changes.ToRazdfile, dir, log); err != nil {
			return err
		}
	}

	// Apply Razdfile -> native (with backup prompt).
	if len(changes.ToNative) > 0 {
		nativePath := NativeConfig(prov.Name(), dir)
		if nativePath != "" {
			shouldBackup := forceBackup
			if !shouldBackup {
				decision, pErr := PromptBackup(nativePath, log)
				if pErr != nil {
					log.Warnf("[SYNC] backup prompt failed: %v\n", pErr)
				} else if decision == BackupYes {
					shouldBackup = true
				}
			}
			if shouldBackup {
				if _, bErr := BackupFile(nativePath, log); bErr != nil {
					log.Warnf("[SYNC] backup failed: %v\n", bErr)
				}
			}
		}

		toNative := toolsToMap(changes.ToNative)
		if err := prov.WriteTools(toNative); err != nil {
			return fmt.Errorf("failed to write %s config: %w", prov.Name(), err)
		}
		log.Successf("Synchronized %s config with Razdfile\n", prov.Name())
	}

	return nil
}

// ensureTools extracts the tool list from the Razdfile dependencies.ensure.
func ensureTools(rf *ast.Razdfile) ([]Tool, error) {
	if !rf.HasDependencies() {
		return nil, nil
	}
	parsed, err := rf.Dependencies.ParseEnsure()
	if err != nil {
		return nil, fmt.Errorf("failed to parse dependencies: %w", err)
	}
	out := make([]Tool, 0, len(parsed))
	for _, p := range parsed {
		out = append(out, Tool{Name: p.Tool, Version: p.Version})
	}
	return out, nil
}

// applyToRazdfile appends new tools to dependencies.ensure and persists via the
// AST-preserving writer.
func applyToRazdfile(rf *ast.Razdfile, tools []Tool, dir string, log Logger) error {
	existing := make(map[string]bool, len(rf.Dependencies.Ensure))
	for _, dep := range rf.Dependencies.Ensure {
		existing[dep] = true
	}

	for _, t := range tools {
		entry := t.Name + "@" + t.Version
		if existing[entry] {
			continue
		}
		rf.Dependencies.Ensure = append(rf.Dependencies.Ensure, entry)
		existing[entry] = true
		log.Infof("[SYNC] Added %s to Razdfile dependencies\n", entry)
	}

	targetPath := filepath.Join(dir, "Razdfile.yml")
	if _, statErr := os.Stat(targetPath); statErr != nil && os.IsNotExist(statErr) {
		for _, name := range razdfile.DefaultRazdfiles {
			candidate := filepath.Join(dir, name)
			if _, sErr := os.Stat(candidate); sErr == nil {
				targetPath = candidate
				break
			}
		}
	}

	didWrite, err := razdfile.UpdateEnsureInFile(targetPath, rf.Dependencies.Ensure)
	if err != nil {
		return fmt.Errorf("failed to update Razdfile: %w", err)
	}
	if didWrite {
		log.Successf("Razdfile synchronized with %s config\n", "native")
	}
	return nil
}

// toolsFromMap converts a name->version map to a Tool slice.
func toolsFromMap(m map[string]string) []Tool {
	out := make([]Tool, 0, len(m))
	for name, version := range m {
		out = append(out, Tool{Name: name, Version: version})
	}
	return out
}

// toolsToMap converts a Tool slice to a name->version map.
func toolsToMap(tools []Tool) map[string]string {
	out := make(map[string]string, len(tools))
	for _, t := range tools {
		out[t.Name] = t.Version
	}
	return out
}
