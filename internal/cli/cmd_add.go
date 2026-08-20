package cli

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	taskast "github.com/go-task/task/v3/taskfile/ast"
	"github.com/razd-cli/razd/internal/flags"
	"github.com/razd-cli/razd/internal/output"
	"github.com/razd-cli/razd/internal/trust"
	"github.com/razd-cli/razd/razdfile"
	"github.com/razd-cli/razd/razdfile/ast"
)

// runAdd implements the "razd add <tool@version>" command.
// It adds dependencies to an existing Razdfile.yml, or creates a task when
// invoked as "razd add task <name> -- <cmd>".
func runAdd(ctx *Context) error {
	if len(ctx.Args) == 0 {
		return fmt.Errorf("usage: razd add <tool@version> [tool@version...] | razd add task <name> -- <cmd>")
	}

	// "razd add task <name> -- <cmd>" creates a task. A bare "razd add task"
	// (no name) falls through to the dependency path and adds the `task` package.
	if ctx.Args[0] == "task" && len(ctx.Args) > 1 {
		ctx.Log.Debugf("add task: %d args, routing to task creation\n", len(ctx.Args))
		return runAddTask(ctx, ctx.Args[1:])
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

	// The Razdfile has no dependencies section (e.g. `razd init` with "none").
	// Adding a package requires a provisioner, so ask for one interactively.
	// This is the only place the provisioner is clarified — init is friction-free.
	createdDeps := false
	if !rf.HasDependencies() {
		using, err := resolveAddProvisioner(ctx.Log)
		if err != nil {
			return err
		}
		if using == "none" {
			return fmt.Errorf("cannot add packages without a provisioner; choose 'mise' or 'devbox'")
		}
		rf.Dependencies = &ast.DependenciesConfig{
			Using:  using,
			Ensure: []string{},
		}
		createdDeps = true
		ctx.Log.Debugf("Created dependencies section with using=%s\n", using)
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
	var didWrite bool
	if createdDeps {
		didWrite, err = razdfile.CreateDependenciesInFile(targetPath, rf.Dependencies.Using, rf.Dependencies.Ensure)
	} else {
		didWrite, err = razdfile.UpdateEnsureInFile(targetPath, rf.Dependencies.Ensure)
	}
	if err != nil {
		return fmt.Errorf("failed to update Razdfile: %w", err)
	}
	if !didWrite {
		ctx.Log.Debugf("ensure list unchanged after add, expected a write\n")
		return fmt.Errorf("failed to update Razdfile: ensure list was not written")
	}

	ctx.Log.Successf("Added %d dependencies to %s\n", len(added), targetPath)
	for _, dep := range added {
		ctx.Log.Infof("  + %s\n", dep)
	}

	// Synchronize the native config (mise.toml / devbox.json) with the updated
	// Razdfile, preserving any sections not managed by razd.
	if !flags.NoSync {
		if prov, ok := tryGetProvisioner(rf, dir, ctx.Log); ok {
			if err := syncRazdfile(rf, prov, dir, ctx.Log); err != nil {
				ctx.Log.Warnf("Failed to sync %s config: %v\n", prov.Name(), err)
			}
		}
	}

	return nil
}

// resolveAddProvisioner asks the user which provisioner to use when the
// Razdfile has no dependencies section. It never auto-selects: the user must
// explicitly choose. In non-interactive mode (--yes, pipes, CI) it returns an
// error so the caller can fail without blocking.
func resolveAddProvisioner(log *output.Logger) (string, error) {
	if !trust.IsTerminal() {
		log.Debugf("No dependencies section and non-interactive stdin, cannot prompt for provisioner\n")
		return "", fmt.Errorf("Razdfile does not have a 'dependencies' section. Use 'razd init' to create one")
	}

	using, err := promptInitProvider(log, runtime.GOOS)
	if err != nil {
		return "", err
	}
	if using == "" {
		log.Debugf("Provisioner prompt aborted\n")
		return "", fmt.Errorf("provisioner selection aborted")
	}
	log.Debugf("Provisioner selected for add: %s\n", using)
	return using, nil
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

// runAddTask implements "razd add task <name> -- <cmd>". It creates or updates
// a task in the Razdfile's tasks: section. pflag has already consumed the "--"
// terminator, so args are [name, cmd...].
func runAddTask(ctx *Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: razd add task <name> -- <cmd>")
	}

	name := args[0]
	if err := validateTaskName(name); err != nil {
		return err
	}

	cmds := args[1:]
	if len(cmds) == 0 {
		return fmt.Errorf("task %q requires at least one command after '--'", name)
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

	// If the task already exists, ask before overwriting it.
	if rf.GetTask(name) != nil {
		ok, err := confirmOverwrite(name, ctx.Log)
		if err != nil {
			return err
		}
		if !ok {
			ctx.Log.Infof("Task %q not overwritten\n", name)
			return nil
		}
	}

	task := &taskast.Task{
		Desc:        flags.TaskDesc,
		Dir:         flags.TaskDir,
		Silent:      flags.TaskSilent,
		Interactive: flags.TaskInteractive,
	}
	// Join the command args into a single command string so that
	// `razd add task hello -- echo 'hi'` writes `cmd: echo 'hi'`.
	task.Cmds = append(task.Cmds, &taskast.Cmd{Cmd: strings.Join(cmds, " ")})
	for _, d := range flags.TaskDeps {
		task.Deps = append(task.Deps, &taskast.Dep{Task: d})
	}

	ctx.Log.Debugf("Adding task %q with command %q\n", name, task.Cmds[0].Cmd)
	if task.Desc != "" {
		ctx.Log.Infof("  desc: %s\n", task.Desc)
	}
	for _, d := range flags.TaskDeps {
		ctx.Log.Infof("  dep: %s\n", d)
	}
	if task.Dir != "" {
		ctx.Log.Infof("  dir: %s\n", task.Dir)
	}
	if task.Silent {
		ctx.Log.Infof("  silent: true\n")
	}
	if task.Interactive {
		ctx.Log.Infof("  interactive: true\n")
	}

	targetPath := filepath.Join(dir, "Razdfile.yml")
	didWrite, err := razdfile.UpdateTasksInFile(targetPath, name, task)
	if err != nil {
		return fmt.Errorf("failed to update Razdfile: %w", err)
	}
	if !didWrite {
		return fmt.Errorf("failed to update Razdfile: task was not written")
	}

	ctx.Log.Successf("Added task %q to %s\n", name, targetPath)

	// If a provisioner is configured, ensure the `task` tool is available in
	// the native config (e.g. `mise use task`). Without an explicit version it
	// defaults to "latest". An existing pinned version is left untouched.
	if !flags.NoSync {
		if prov, ok := tryGetProvisioner(rf, dir, ctx.Log); ok {
			native, err := prov.ReadConfig()
			if err != nil {
				ctx.Log.Warnf("Failed to read %s config: %v\n", prov.Name(), err)
			} else if _, exists := native["task"]; !exists {
				if err := prov.WriteTools(map[string]string{"task": "latest"}); err != nil {
					ctx.Log.Warnf("Failed to add 'task' to %s config: %v\n", prov.Name(), err)
				} else {
					ctx.Log.Infof("Added 'task' tool to %s config\n", prov.Name())
				}
			} else {
				ctx.Log.Debugf("'task' tool already present in %s config\n", prov.Name())
			}
		}
	}

	return nil
}

// validateTaskName checks that a task name is a simple, valid identifier.
func validateTaskName(name string) error {
	if name == "" {
		return fmt.Errorf("task name cannot be empty")
	}
	if strings.ContainsAny(name, "@ \t\n:") {
		return fmt.Errorf("invalid task name %q: must not contain '@', whitespace, or ':'", name)
	}
	return nil
}