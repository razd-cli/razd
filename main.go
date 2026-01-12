package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-task/task/v3"
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

	taskDir := cwd + "/examples/simple-task-file"

	runner := &NativeRunner{
		WorkDir:    taskDir,
		Entrypoint: taskDir + "/Taskfile.yml",
	}

	ctx := context.Background()
	if err := runner.Run(ctx, "default"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
