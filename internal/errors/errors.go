// Package errors defines error types and exit codes for razd CLI.
package errors

import "fmt"

// Exit codes
const (
	CodeOK            = 0
	CodeUnknown       = 1
	CodeNoRazdfile    = 100
	CodeInvalidConfig = 101
	CodeTaskNotFound  = 200
	CodeTaskFailed    = 201
	CodeTrustError    = 300
)

// RazdError is the interface for razd-specific errors with exit codes.
type RazdError interface {
	error
	Code() int
}

// TaskNotFoundError is returned when a task is not found.
type TaskNotFoundError struct {
	TaskName string
}

func (e *TaskNotFoundError) Error() string {
	return fmt.Sprintf("task %q not found", e.TaskName)
}

func (e *TaskNotFoundError) Code() int {
	return CodeTaskNotFound
}

// TaskRunError is returned when a task fails during execution.
type TaskRunError struct {
	TaskName     string
	Err          error
	TaskExitCode int
}

func (e *TaskRunError) Error() string {
	return fmt.Sprintf("task %q failed: %v", e.TaskName, e.Err)
}

func (e *TaskRunError) Code() int {
	return CodeTaskFailed
}

// ConfigError is returned for configuration errors.
type ConfigError struct {
	Message string
	Err     error
}

func (e *ConfigError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("config error: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("config error: %s", e.Message)
}

func (e *ConfigError) Code() int {
	return CodeInvalidConfig
}

// NoRazdfileError is returned when no Razdfile is found.
type NoRazdfileError struct {
	Dir string
}

func (e *NoRazdfileError) Error() string {
	return fmt.Sprintf("no Razdfile found in %s", e.Dir)
}

func (e *NoRazdfileError) Code() int {
	return CodeNoRazdfile
}

// TrustError is returned for trust-related errors.
type TrustError struct {
	Path    string
	Message string
}

func (e *TrustError) Error() string {
	return fmt.Sprintf("trust error for %s: %s", e.Path, e.Message)
}

func (e *TrustError) Code() int {
	return CodeTrustError
}
