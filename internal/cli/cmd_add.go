package cli

import (
	"fmt"
	"path/filepath"

	"github.com/razd-cli/razd/razdfile"
	"github.com/razd-cli/razd/razdfile/ast"
)

// runAdd implements the "razd add <tool@version>" command.
// It adds dependencies to an existing Razdfile.yml.
func runAdd(ctx *Context) error {
	if len(ctx.Args) == 0 {
		return fmt.Errorf("usage: razd add <tool@version> [tool@version...]")
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
		return fmt.Errorf("Razdfile does not have a 'dependencies' section. Use 'razd init' to create one")
	}

	// Parse and validate each dependency
	var added []string
	for _, dep := range ctx.Args {
		parsed, err := ast.ParseDependencyString(dep)
		if err != nil {
			return fmt.Errorf("invalid dependency format %q: %w", dep, err)
		}

		// Check for duplicates
		if containsDep(rf.Dependencies.Ensure, parsed.Raw) {
			ctx.Log.Debugf("Dependency %q already exists, skipping\n", parsed.Raw)
			continue
		}

		rf.Dependencies.Ensure = append(rf.Dependencies.Ensure, parsed.Raw)
		added = append(added, parsed.Raw)
		ctx.Log.Debugf("Added dependency: %s\n", parsed.Raw)
	}

	if len(added) == 0 {
		ctx.Log.Infof("No new dependencies to add\n")
		return nil
	}

	targetPath := filepath.Join(dir, "Razdfile.yml")
	_, err = razdfile.UpdateEnsureInFile(targetPath, rf.Dependencies.Ensure)
	if err != nil {
		return fmt.Errorf("failed to update Razdfile: %w", err)
	}

	ctx.Log.Successf("Added %d dependencies to %s\n", len(added), targetPath)
	for _, dep := range added {
		ctx.Log.Infof("  + %s\n", dep)
	}

	return nil
}

// containsDep checks if a dependency already exists in the ensure list.
func containsDep(ensure []string, dep string) bool {
	for _, d := range ensure {
		if d == dep {
			return true
		}
	}
	return false
}