package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-task/task/v3"
	"github.com/razd-cli/razd/razdfile"
)

type NativeRunner struct {
	WorkDir    string
	Entrypoint string
}

func (r *NativeRunner) Run(ctx context.Context, taskName string) error {
	e := task.NewExecutor(
		task.WithDir(r.WorkDir),
		task.WithEntrypoint(r.Entrypoint),
		task.WithStdin(os.Stdin),
		task.WithStdout(os.Stdout),
		task.WithStderr(os.Stderr),
		task.WithColor(true),
	)

	if err := e.Setup(); err != nil {
		return fmt.Errorf("failed to parse Taskfile: %w", err)
	}

	return e.Run(ctx, &task.Call{Task: taskName})
}

func main() {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting cwd: %v\n", err)
		os.Exit(1)
	}

	// Try to find Razdfile.yml first using razdfile.Reader
	taskDir := cwd + "/examples/nodejs-project"
	reader := razdfile.NewReader(
		razdfile.WithDir(taskDir),
		razdfile.WithDebugFunc(func(format string, args ...any) {
			fmt.Printf("[DEBUG] "+format+"\n", args...)
		}),
	)

	rf, err := reader.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading Razdfile: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Razdfile version: %s\n", rf.Version)
	
	if rf.HasDevbox() && rf.Devbox.HasPackages() {
		fmt.Println("Devbox packages detected:")
		for _, pkg := range rf.Devbox.Packages.List {
			fmt.Printf("  - %s\n", pkg)
		}
		for name, cfg := range rf.Devbox.Packages.Map {
			if cfg != nil {
				fmt.Printf("  - %s: %s\n", name, cfg.Version)
			}
		}
	}

	if rf.HasMise() && rf.Mise.HasTools() {
		fmt.Println("Mise tools detected:")
		for name, tool := range rf.Mise.Tools {
			fmt.Printf("  - %s: %s\n", name, tool.Version)
		}
	}

	if rf.HasTasks() {
		fmt.Println("Tasks detected:")
		for name := range rf.Tasks.All(nil) {
			fmt.Printf("  - %s\n", name)
		}
	}

	// Use Razdfile.yml as entrypoint for go-task
	entrypoint := taskDir + "/Razdfile.yml"
	
	runner := &NativeRunner{
		WorkDir:    taskDir,
		Entrypoint: entrypoint,
	}

	ctx := context.Background()
	if err := runner.Run(ctx, "default"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
