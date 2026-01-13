// Package output provides logging and output formatting for razd CLI.
package output

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
)

// Logger handles all output for razd CLI.
type Logger struct {
	Stdout  io.Writer
	Stderr  io.Writer
	Verbose bool
	Silent  bool
	Color   bool
}

// NewLogger creates a new Logger with stdout/stderr.
func NewLogger(stdout, stderr io.Writer) *Logger {
	// Check for NO_COLOR environment variable
	noColor := os.Getenv("NO_COLOR") != ""
	forceColor := os.Getenv("FORCE_COLOR") != ""

	useColor := !noColor || forceColor

	return &Logger{
		Stdout: stdout,
		Stderr: stderr,
		Color:  useColor,
	}
}

// Default returns a logger with default stdout/stderr.
func Default() *Logger {
	return NewLogger(os.Stdout, os.Stderr)
}

// SetVerbose enables or disables verbose output.
func (l *Logger) SetVerbose(v bool) {
	l.Verbose = v
}

// SetSilent enables or disables silent mode.
func (l *Logger) SetSilent(s bool) {
	l.Silent = s
}

// SetColor enables or disables colored output.
func (l *Logger) SetColor(c bool) {
	l.Color = c
	color.NoColor = !c
}

// Infof prints an info message.
func (l *Logger) Infof(format string, args ...any) {
	if l.Silent {
		return
	}
	fmt.Fprintf(l.Stdout, format, args...)
}

// Debugf prints a debug message (only in verbose mode).
func (l *Logger) Debugf(format string, args ...any) {
	if !l.Verbose || l.Silent {
		return
	}
	if l.Color {
		gray := color.New(color.FgHiBlack)
		gray.Fprintf(l.Stdout, format, args...)
	} else {
		fmt.Fprintf(l.Stdout, format, args...)
	}
}

// Warnf prints a warning message.
func (l *Logger) Warnf(format string, args ...any) {
	if l.Silent {
		return
	}
	if l.Color {
		yellow := color.New(color.FgYellow)
		yellow.Fprintf(l.Stderr, "warning: ")
	} else {
		fmt.Fprint(l.Stderr, "warning: ")
	}
	fmt.Fprintf(l.Stderr, format, args...)
}

// Errf prints an error message.
func (l *Logger) Errf(format string, args ...any) {
	if l.Color {
		red := color.New(color.FgRed)
		red.Fprintf(l.Stderr, format, args...)
	} else {
		fmt.Fprintf(l.Stderr, format, args...)
	}
}

// Successf prints a success message.
func (l *Logger) Successf(format string, args ...any) {
	if l.Silent {
		return
	}
	if l.Color {
		green := color.New(color.FgGreen)
		green.Fprintf(l.Stdout, format, args...)
	} else {
		fmt.Fprintf(l.Stdout, format, args...)
	}
}

// TaskOutput prints task execution output with prefix.
func (l *Logger) TaskOutput(taskName string, output string) {
	if l.Silent {
		return
	}
	if l.Color {
		cyan := color.New(color.FgCyan)
		cyan.Fprintf(l.Stdout, "task: [%s] ", taskName)
	} else {
		fmt.Fprintf(l.Stdout, "task: [%s] ", taskName)
	}
	fmt.Fprint(l.Stdout, output)
}
