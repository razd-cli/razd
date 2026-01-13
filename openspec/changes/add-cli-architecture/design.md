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
│   │   ├── up.go                # razd up (install tools via mise/devbox)
│   │   ├── run.go               # razd run <task>
│   │   ├── init.go              # razd init (create Razdfile.yml)
│   │   ├── add.go               # razd add <tool@version>
│   │   ├── shell.go             # razd shell (interactive shell)
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
    "up":      {Name: "up", Run: runUp, Description: "Set up project (install tools)"},
    "run":     {Name: "run", Run: runRun, Description: "Execute a custom task"},
    "init":    {Name: "init", Run: runInit, Description: "Create new Razdfile.yml"},
    "add":     {Name: "add", Run: runAdd, Description: "Add dependencies to Razdfile"},
    "shell":   {Name: "shell", Aliases: []string{"sh"}, Run: runShell, Description: "Start interactive shell with environment"},
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

### 3.1 Command Details

#### `razd up` — настройка проекта

Устанавливает dev tools через provisioner. Работает как `npx` / `pnpm dlx` для клонирования.

**Сценарий 1: Локальный проект**
```bash
cd my-project
razd up           # настроить текущий проект
```

Выполняет:
1. Читает Razdfile.yml
2. Проверяет trust (спрашивает при первом запуске)
3. Устанавливает dev tools (`mise install` / `devbox install`)

**Сценарий 2: Клонирование + настройка**
```bash
razd up https://github.com/user/repo
razd up git@github.com:user/repo.git
razd up gh:user/repo                    # короткий синтаксис GitHub
```

Выполняет:
1. `git clone <url>` в текущую директорию
2. `cd <repo-name>`
3. То же что Сценарий 1

```go
// internal/cli/up.go
func runUp(args []string) error {
    if len(args) > 0 {
        // Clone mode
        url := args[0]
        repoDir, err := gitClone(url)
        if err != nil {
            return err
        }
        if err := os.Chdir(repoDir); err != nil {
            return err
        }
    }
    
    // Local setup mode
    rf, err := razdfile.NewReader().Read()
    if err != nil {
        return err
    }
    
    if err := trust.EnsureTrusted(rf.Location, flags.Yes); err != nil {
        return err
    }
    
    // Install dev tools via provisioner
    return installTools(rf)
}

func installTools(rf *ast.Razdfile) error {
    if rf.Dependencies == nil {
        return nil
    }
    
    switch rf.Dependencies.Using {
    case "mise":
        return exec.Command("mise", "install").Run()
    case "devbox":
        return exec.Command("devbox", "install").Run()
    }
    return nil
}
```

**Флаги:**
- `--yes` / `-y` — автоматически доверять проекту

#### `razd` (без аргументов) — запуск проекта

Запускает задачу `default` из Razdfile. Предполагает что проект уже настроен.

```bash
razd              # → run default task
razd dev          # → run dev task  
razd build        # → run build task
razd mytask       # → run mytask
```

#### `razd init`

Создаёт новый Razdfile.yml в текущей директории.

```go
// internal/cli/init.go
func runInit(args []string) error {
    if razdfile.Exists(".") && !flags.Force {
        return errors.New("Razdfile.yml already exists, use --force to overwrite")
    }
    
    config := InitConfig{
        Using: flags.Using,  // --using mise|devbox
    }
    
    if config.Using == "" {
        // Interactive mode: спрашиваем пользователя
        config.Using = promptUsing()
    }
    
    // Сканируем проект и предлагаем миграцию
    if existingMise := detectMiseToml("."); existingMise != nil {
        if promptMigrate("mise.toml") {
            config.Ensure = parseMiseTools(existingMise)
        }
    }
    
    return writeRazdfile(config)
}
```

**Флаги:**
- `--using mise|devbox` — выбор provisioner
- `--force` — перезаписать существующий файл
- `--migrate` — автоматически мигрировать из mise.toml/devbox.json

#### `razd add`

Добавляет зависимости в `dependencies.ensure` секцию Razdfile.yml.

```go
// internal/cli/add.go
func runAdd(args []string) error {
    if len(args) == 0 {
        return errors.New("usage: razd add <tool@version> [tool@version...]")
    }
    
    rf, err := razdfile.NewReader().Read()
    if err != nil {
        return err
    }
    
    for _, dep := range args {
        parsed, err := ast.ParseDependency(dep)
        if err != nil {
            return fmt.Errorf("invalid dependency format: %s", dep)
        }
        rf.Dependencies.Ensure = appendUnique(rf.Dependencies.Ensure, dep)
    }
    
    return rf.Save()
}
```

**Использование:**
```bash
razd add node@22                    # Добавить node v22
razd add python@3.12 go@1.22       # Добавить несколько
razd add node@latest               # Последняя версия
```

#### `razd shell`

Запускает интерактивную оболочку с настроенным окружением.

```go
// internal/cli/shell.go
func runShell(args []string) error {
    rf, err := razdfile.NewReader().Read()
    if err != nil {
        return err
    }
    
    if err := trust.EnsureTrusted(rf.Location, flags.Yes); err != nil {
        return err
    }
    
    provisioner := rf.Dependencies.Using // "mise" или "devbox"
    
    switch provisioner {
    case "mise":
        return exec.Command("mise", "shell").Run()
    case "devbox":
        return exec.Command("devbox", "shell").Run()
    default:
        // Fallback: запускаем $SHELL с настроенным PATH
        return runDefaultShell(rf)
    }
}
```

**Примечание**: Алиас `razd sh` для краткости.

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
        // No args: run default task (assumes project is set up)
        return c.runTask("default", nil)
    }
    
    cmd := args[0]
    if command, ok := commands[cmd]; ok {
        return command.Run(args[1:])
    }
    
    // Assume it's a task name (like pnpm behavior)
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
    commands="up run init add shell sh dev build list ls trust"
    opts="--help --version --verbose --silent --dir --yes --force"
    
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
