package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/razd-cli/razd/internal/flags"
	"github.com/razd-cli/razd/internal/sync"
	"github.com/razd-cli/razd/razdfile"
	"github.com/razd-cli/razd/razdfile/ast"
)

// runRemove implements the "razd rm <tool>" command. It removes dependencies
// from the Razdfile's ensure list and delegates the actual removal to the
// native package manager (devbox rm / mise unuse) when its config exists.
func runRemove(ctx *Context) error {
	if len(ctx.Args) == 0 {
		return fmt.Errorf("usage: razd rm <tool> [tool...]")
	}

	dir, err := resolveDir(ctx)
	if err != nil {
		return err
	}

	reader := razdfile.NewReader(
		razdfile.WithDir(dir),
		razdfile.WithDebugFunc(ctx.Log.Debugf),
	)

	rf, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read Razdfile: %w", err)
	}

	if !rf.HasDependencies() {
		ctx.Log.Debugf("Razdfile has no dependencies section, nothing to remove\n")
		ctx.Log.Infof("No dependencies to remove\n")
		return nil
	}

	// Parse and validate each requested tool.
	toRemove := make(map[string]bool, len(ctx.Args))
	for _, dep := range ctx.Args {
		parsed, err := ast.ParseDependencyString(dep)
		if err != nil {
			return fmt.Errorf("invalid dependency format %q: %w", dep, err)
		}
		toRemove[parsed.Tool] = true
		ctx.Log.Debugf("Removing dependency: %s\n", parsed.Tool)
	}

	// Check which requested tools are actually present (by name) in ensure.
	present := make(map[string]bool)
	for _, d := range rf.Dependencies.Ensure {
		name := d
		if idx := strings.LastIndex(d, "@"); idx > 0 && idx < len(d)-1 {
			name = d[:idx]
		}
		if toRemove[name] {
			present[name] = true
		}
	}

	if len(present) == 0 {
		ctx.Log.Infof("No matching dependencies to remove\n")
		return nil
	}

	// Persist the trimmed ensure list.
	targetPath := filepath.Join(dir, "Razdfile.yml")
	didWrite, err := razdfile.RemoveFromEnsureInFile(targetPath, present)
	if err != nil {
		return fmt.Errorf("failed to update Razdfile: %w", err)
	}

	// Keep the in-memory model in sync with the trimmed ensure list so the
	// subsequent sync does not re-add removed packages from the stale model.
	keptEnsure := make([]string, 0, len(rf.Dependencies.Ensure))
	for _, d := range rf.Dependencies.Ensure {
		name := d
		if idx := strings.LastIndex(d, "@"); idx > 0 && idx < len(d)-1 {
			name = d[:idx]
		}
		if !present[name] {
			keptEnsure = append(keptEnsure, d)
		}
	}
	rf.Dependencies.Ensure = keptEnsure

	if didWrite {
		ctx.Log.Successf("Removed %d dependencies from %s\n", len(present), targetPath)
		for name := range present {
			ctx.Log.Infof("  - %s\n", name)
		}
	} else {
		ctx.Log.Infof("No matching dependencies to remove\n")
	}

	if flags.NoSync {
		return nil
	}

	prov, ok := tryGetProvisioner(rf, dir, ctx.Log)
	if !ok {
		return nil
	}

	// Delegate the removal to the native package manager (devbox rm / mise
	// unuse) when its config exists; it is idempotent for absent packages.
	nativePath := sync.NativeConfig(prov.Name(), dir)
	nativeExists := false
	if nativePath != "" {
		if _, statErr := os.Stat(nativePath); statErr == nil {
			nativeExists = true
		}
	}

	if nativeExists {
		ctx.Log.Debugf("Native %s config exists, delegating removal to %s\n", prov.Name(), prov.Name())
		requested := make(map[string]string, len(present))
		for name := range present {
			requested[name] = ""
		}
		if err := prov.RemoveTools(context.Background(), requested); err != nil {
			ctx.Log.Warnf("Native %s remove failed: %v\n", prov.Name(), err)
		}
		// No reconcile step here: rm already updated both the Razdfile (via
		// RemoveFromEnsureInFile) and the native config (via RemoveTools)
		// directly. A bidirectional sync would re-ask the direction prompt and
		// pull native-only packages back into ensure, defeating the removal.
	}

	return nil
}
