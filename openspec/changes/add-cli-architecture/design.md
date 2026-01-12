# Design: CLI Architecture

## Overview

Архитектура CLI вдохновлена go-task/task, но адаптирована под специфику razd:
- Единый бинарник с субкомандами
- Интеграция с task executor как библиотекой
- Поддержка mise/devbox для управления инструментами

## Архитектурные решения

### 1. Структура пакетов

```
razd/
├── cmd/
│   └── razd/
│       └── main.go              # Entry point, минимальная логика
├── internal/
│   ├── cli/
│   │   ├── cli.go               # CLI orchestrator
│   │   ├── commands.go          # Command registry
│   │   ├── up.go                # razd up (clone + setup, или local setup)
│   │   ├── run.go               # razd run <task>
│   │   ├── install.go           # razd install (mise/devbox)
│   │   ├── setup.go             # razd setup (project dependencies)
│   │   ├── dev.go               # razd dev
│   │   ├── build.go             # razd build
│   │   ├── list.go              # razd list (--json, --all)
│   │   └── trust.go             # razd trust (--untrust, --show, --all, --ignore)
│   ├── flags/
│   │   ├── flags.go             # Global flag definitions
│   │   └── validate.go          # Flag validation
│   ├── output/
│   │   ├── logger.go            # Logging with colors
│   │   └── format.go            # Output formatting
│   ├── trust/
│   │   ├── store.go             # Trust store (JSON file in ~/.config/razd/)
│   │   └── check.go             # Trust verification before execution
│   └── version/
│       └── version.go           # Version info
├── razdfile/                    # Existing package (unchanged)
└── main.go                      # Deprecated, redirects to cmd/razd
```

**Обоснование**: 
- `cmd/razd/` — стандартная Go структура для бинарников
- `internal/` — приватные пакеты, недоступные извне
- Разделение на `cli/`, `flags/`, `output/` как в go-task

### 2. Подход к парсингу флагов

**Выбор: spf13/pflag (как в go-task)**

```go
// internal/flags/flags.go
package flags

import (
    "github.com/spf13/pflag"
)

var (
    // Global flags
    Version   bool
    Help      bool
    Verbose   bool
    Silent    bool
    Dir       string
    Color     bool
    Yes       bool
    List      bool      // Global --list flag
    NoSync    bool      // Skip Razdfile <-> mise.toml sync
    
    // File path flags
    TaskFile  string    // -t, --taskfile
    RazdFile  string    // --razdfile (priority over --taskfile)
)

func init() {
    pflag.BoolVar(&Version, "version", false, "Show razd version")
    pflag.BoolVarP(&Help, "help", "h", false, "Show help")
    pflag.BoolVarP(&Verbose, "verbose", "v", false, "Enable verbose output")
    pflag.BoolVarP(&Silent, "silent", "s", false, "Disable output")
    pflag.StringVarP(&Dir, "dir", "d", "", "Working directory")
    pflag.BoolVar(&Color, "color", true, "Enable colored output")
    pflag.BoolVarP(&Yes, "yes", "y", false, "Assume yes to all prompts")
    pflag.BoolVar(&List, "list", false, "List all available tasks")
    pflag.BoolVar(&NoSync, "no-sync", false, "Skip Razdfile <-> mise.toml sync")
    pflag.StringVarP(&TaskFile, "taskfile", "t", "", "Path to taskfile/razdfile")
    pflag.StringVar(&RazdFile, "razdfile", "", "Path to razdfile (priority over --taskfile)")
}
```

**Альтернативы рассмотренные**:
- `cobra` — более тяжеловесный, избыточен для razd
- `urfave/cli` — менее гибкий для сложных флагов
- `kingpin` — устаревший

**Обоснование pflag**:
- Используется в go-task, проверенное решение
- POSIX-совместимый синтаксис (-v, --verbose)
- Легко интегрируется с go-task executor

### 3. Dispatch Pattern для команд

**Подход: Function map (как в go-task)**

```go
// internal/cli/commands.go
package cli

type Command struct {
    Name        string
    Aliases     []string
    Description string
    Run         func(args []string) error
    Flags       func(*pflag.FlagSet)
}

var commands = map[string]*Command{
    "up":      {Name: "up", Run: runUp, Description: "Clone and set up project, or set up local project"},
    "run":     {Name: "run", Run: runRun, Description: "Execute a custom task"},
    "install": {Name: "install", Run: runInstall, Description: "Install development tools via mise/devbox"},
    "setup":   {Name: "setup", Run: runSetup, Description: "Install project dependencies"},
    "dev":     {Name: "dev", Run: runDev, Description: "Start development workflow"},
    "build":   {Name: "build", Run: runBuild, Description: "Build project"},
    "list":    {Name: "list", Aliases: []string{"ls"}, Run: runList, Description: "List available tasks"},
    "trust":   {Name: "trust", Run: runTrust, Description: "Manage project trust status"},
}
```

**Альтернатива**: go-task использует прямой dispatch в `run()`. 
Мы добавляем registry для:
- Автогенерации help
- Shell completion
- Расширяемости

### 4. CLI Orchestrator

```go
// internal/cli/cli.go
package cli

import (
    "github.com/razd-cli/razd/internal/flags"
    "github.com/razd-cli/razd/internal/output"
    "github.com/spf13/pflag"
)

type CLI struct {
    Logger *output.Logger
    Dir    string
}

func New() *CLI {
    return &CLI{
        Logger: output.NewLogger(os.Stdout, os.Stderr),
    }
}

func (c *CLI) Run() error {
    pflag.Parse()
    
    if err := flags.Validate(); err != nil {
        return err
    }
    
    c.Logger.SetVerbose(flags.Verbose)
    c.Logger.SetSilent(flags.Silent)
    c.Logger.SetColor(flags.Color)
    
    if flags.Version {
        return c.showVersion()
    }
    
    if flags.Help {
        return c.showHelp()
    }
    
    args := pflag.Args()
    if len(args) == 0 {
        // Default: run "default" task or "up"
        return c.runDefault()
    }
    
    cmd := args[0]
    if command, ok := commands[cmd]; ok {
        return command.Run(args[1:])
    }
    
    // Assume it's a task name
    return c.runTask(cmd, args[1:])
}
```

### 5. Error Handling и Exit Codes

**Паттерн из go-task**:

```go
// internal/errors/errors.go
package errors

const (
    CodeOk            = 0
    CodeUnknown       = 1
    CodeNoRazdfile    = 100
    CodeInvalidConfig = 101
    CodeTaskNotFound  = 200
    CodeTaskFailed    = 201
)

type RazdError interface {
    error
    Code() int
}

type TaskRunError struct {
    TaskName     string
    Err          error
    TaskExitCode int
}

func (e *TaskRunError) Code() int {
    return CodeTaskFailed
}
```

**В main.go**:
```go
func main() {
    if err := cli.New().Run(); err != nil {
        logger := output.NewLogger(os.Stdout, os.Stderr)
        
        if razdErr, ok := err.(errors.RazdError); ok {
            logger.Errf("%v\n", err)
            os.Exit(razdErr.Code())
        }
        
        logger.Errf("%v\n", err)
        os.Exit(errors.CodeUnknown)
    }
}
```

### 6. Logger/Output

```go
// internal/output/logger.go
package output

import (
    "fmt"
    "io"
    
    "github.com/fatih/color"
)

type Logger struct {
    Stdout  io.Writer
    Stderr  io.Writer
    Verbose bool
    Silent  bool
    Color   bool
}

func NewLogger(stdout, stderr io.Writer) *Logger {
    return &Logger{
        Stdout: stdout,
        Stderr: stderr,
        Color:  true,
    }
}

func (l *Logger) Infof(format string, args ...any) {
    if l.Silent {
        return
    }
    fmt.Fprintf(l.Stdout, format, args...)
}

func (l *Logger) Debugf(format string, args ...any) {
    if !l.Verbose || l.Silent {
        return
    }
    gray := color.New(color.FgHiBlack)
    gray.Fprintf(l.Stdout, format, args...)
}

func (l *Logger) Errf(format string, args ...any) {
    red := color.New(color.FgRed)
    red.Fprintf(l.Stderr, format, args...)
}

func (l *Logger) Successf(format string, args ...any) {
    green := color.New(color.FgGreen)
    green.Fprintf(l.Stdout, format, args...)
}
```

### 7. Интеграция с go-task Executor

```go
// internal/cli/run.go
func runRun(args []string) error {
    reader := razdfile.NewReader(
        razdfile.WithDir(flags.Dir),
    )
    
    rf, err := reader.Read()
    if err != nil {
        return err
    }
    
    // Use go-task executor
    e := task.NewExecutor(
        task.WithDir(flags.Dir),
        task.WithEntrypoint(rf.Location),
        task.WithStdin(os.Stdin),
        task.WithStdout(os.Stdout),
        task.WithStderr(os.Stderr),
        task.WithColor(flags.Color),
        task.WithSilent(flags.Silent),
    )
    
    if err := e.Setup(); err != nil {
        return err
    }
    
    taskName := "default"
    if len(args) > 0 {
        taskName = args[0]
    }
    
    return e.Run(context.Background(), &task.Call{Task: taskName})
}
```

### 8. Shell Completion

```go
// internal/cli/completion.go
func generateCompletion(shell string) (string, error) {
    switch shell {
    case "bash":
        return bashCompletion, nil
    case "zsh":
        return zshCompletion, nil
    case "fish":
        return fishCompletion, nil
    default:
        return "", fmt.Errorf("unsupported shell: %s", shell)
    }
}

const bashCompletion = `
_razd() {
    local cur prev opts commands
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    commands="up run install dev build list init"
    opts="--help --version --verbose --silent --dir --yes"
    
    if [[ ${cur} == -* ]]; then
        COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
        return 0
    fi
    
    if [[ ${COMP_CWORD} == 1 ]]; then
        COMPREPLY=( $(compgen -W "${commands}" -- ${cur}) )
        return 0
    fi
}
complete -F _razd razd
`
```

## Trade-offs

| Решение | Плюсы | Минусы |
|---------|-------|--------|
| pflag вместо cobra | Легковесность, совместимость с go-task | Меньше out-of-box фич |
| Command registry | Расширяемость, автогенерация help | Чуть больше boilerplate |
| internal/ packages | Инкапсуляция, невозможность импорта извне | Нужно дублировать типы для публичного API |

### 9. Trust System

Аналогично Rust версии, реализуем проверку доверия перед выполнением:

```go
// internal/trust/store.go
package trust

import (
    "encoding/json"
    "os"
    "path/filepath"
)

type TrustStatus string

const (
    StatusTrusted TrustStatus = "trusted"
    StatusIgnored TrustStatus = "ignored"
    StatusUnknown TrustStatus = "unknown"
)

type TrustStore struct {
    Trusted []string `json:"trusted"`
    Ignored []string `json:"ignored"`
}

func (s *TrustStore) GetStatus(path string) TrustStatus {
    for _, p := range s.Trusted {
        if p == path { return StatusTrusted }
    }
    for _, p := range s.Ignored {
        if p == path { return StatusIgnored }
    }
    return StatusUnknown
}

func Load() (*TrustStore, error) {
    configDir, _ := os.UserConfigDir()
    storePath := filepath.Join(configDir, "razd", "trust.json")
    // ... load or create
}
```

```go
// internal/trust/check.go
func EnsureTrusted(path string, autoYes bool) error {
    store, _ := Load()
    status := store.GetStatus(path)
    
    switch status {
    case StatusTrusted:
        return nil
    case StatusIgnored:
        return errors.New("project is ignored")
    case StatusUnknown:
        if autoYes {
            store.AddTrusted(path)
            return nil
        }
        // Prompt user
        if promptTrust(path) {
            store.AddTrusted(path)
            return nil
        }
        return errors.New("project not trusted")
    }
    return nil
}
```

## Future Considerations

1. **Plugins** — возможность расширять razd плагинами
2. **Remote Razdfiles** — поддержка `razd up https://...`
3. **Config file** — `~/.config/razd/config.toml` для глобальных настроек
4. **Watch mode** — `razd dev --watch`
5. **Mise sync** — автоматическая синхронизация Razdfile.yml ↔ mise.toml
