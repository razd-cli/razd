package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/razd-cli/razd/internal/flags"
	"github.com/razd-cli/razd/razdfile/ast"
)

// runList implements the "razd list" command.
// It reads a Razdfile and lists all available tasks.
func runList(ctx *Context) error {
	dir, err := resolveDir(ctx)
	if err != nil {
		return err
	}

	rf, err := readRazdfile(dir, ctx.Log)
	if err != nil {
		return err
	}

	if !rf.HasTasks() {
		ctx.Log.Infof("No tasks defined in Razdfile\n")
		return nil
	}

	if flags.JSON {
		return listJSON(rf, ctx)
	}

	return listText(rf, ctx)
}

// listText prints tasks in a human-readable table format.
func listText(rf *ast.Razdfile, ctx *Context) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TASK\tDESC")

	for name := range rf.Tasks.All(nil) {
		task, ok := rf.Tasks.Get(name)
		if !ok || task == nil {
			fmt.Fprintf(w, "%s\t\n", name)
			continue
		}

		// Skip internal tasks (prefixed with _) unless --all is set
		if len(name) > 0 && name[0] == '_' && !flags.All {
			continue
		}

		desc := ""
		if task.Desc != "" {
			desc = task.Desc
		}
		fmt.Fprintf(w, "%s\t%s\n", name, desc)
	}

	w.Flush()

	if rf.HasDependencies() {
		fmt.Fprintf(os.Stdout, "\nProvisioner: %s\n", rf.Dependencies.Using)
		if rf.Dependencies.HasEnsure() {
			fmt.Fprintf(os.Stdout, "Dependencies:\n")
			deps, _ := rf.Dependencies.ParseEnsure()
			for _, dep := range deps {
				fmt.Fprintf(os.Stdout, "  - %s@%s\n", dep.Tool, dep.Version)
			}
		}
	} else if rf.HasMise() {
		fmt.Fprintf(os.Stdout, "\nProvisioner: mise\n")
	} else if rf.HasDevbox() {
		fmt.Fprintf(os.Stdout, "\nProvisioner: devbox\n")
	}

	return nil
}

// listJSON prints tasks in JSON format.
func listJSON(rf *ast.Razdfile, ctx *Context) error {
	type taskEntry struct {
		Name        string   `json:"name"`
		Description string   `json:"description,omitempty"`
		Deps        []string `json:"deps,omitempty"`
		Commands    []string `json:"cmds,omitempty"`
	}

	var tasks []taskEntry
	for name := range rf.Tasks.All(nil) {
		task, ok := rf.Tasks.Get(name)
		if !ok || task == nil {
			continue
		}

		if len(name) > 0 && name[0] == '_' && !flags.All {
			continue
		}

		entry := taskEntry{
			Name: name,
		}

		if task.Desc != "" {
			entry.Description = task.Desc
		}

		for _, dep := range task.Deps {
			entry.Deps = append(entry.Deps, dep.Task)
		}

		for _, cmd := range task.Cmds {
			if cmd.Cmd != "" {
				entry.Commands = append(entry.Commands, cmd.Cmd)
			}
		}

		tasks = append(tasks, entry)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(tasks)
}