# Design: Unified Dependencies

## Overview

Этот документ описывает архитектурные решения для секции `dependencies` в Razdfile.

## Core Concepts

### 1. Dependency Specification Format

Формат `ensure` строк:

```
<tool>@<version>
```

Где:
- `<tool>` — имя инструмента (node, php, python, pnpm, etc.)
- `@` — разделитель
- `<version>` — версия (конкретная: "22", "8.4", "3.11" или "latest")

### 2. Tool Name Normalization

Razd использует **канонические имена** инструментов. При трансляции в devbox/mise выполняется маппинг.

```
Canonical → mise      → devbox
node      → node      → nodejs
php       → php       → php
python    → python    → python
pnpm      → pnpm      → pnpm
```

**Решение**: Маппинг хранится в runtime коде (не в AST). AST хранит канонические имена.

### 3. Package Manager Selection

Поле `using` определяет целевой менеджер:

```yaml
dependencies:
  using: "mise"   # или "devbox"
```

**Допустимые значения**: `mise`, `devbox`

**Default**: Нет значения по умолчанию — поле обязательно.

## AST Structure

### DependenciesConfig

```go
// ast/dependencies.go
type DependenciesConfig struct {
    // Using specifies the package manager to use
    // Valid values: "mise", "devbox"
    Using string `yaml:"using"`
    
    // Ensure is a list of tools to install
    // Format: "tool@version" (e.g., "node@22", "php@8.4")
    Ensure []string `yaml:"ensure,omitempty"`
    
    // Extra contains pass-through configuration for native managers
    // Structure mirrors devbox.json / mise.toml
    Extra *DependenciesExtra `yaml:"extra,omitempty"`
}

// DependenciesExtra holds native configurations for package managers
type DependenciesExtra struct {
    // Devbox configuration - structure mirrors devbox.json
    // Will be deep-merged with generated config
    Devbox map[string]any `yaml:"devbox,omitempty"`
    
    // Mise configuration - structure mirrors mise.toml
    // Will be deep-merged with generated config
    Mise map[string]any `yaml:"mise,omitempty"`
}
```

**Примечание**: `Extra.Devbox` и `Extra.Mise` — это `map[string]any`, не типизированные структуры. Это позволяет:
1. Передавать любую валидную конфигурацию devbox/mise без изменений в AST
2. Пользователь использует документацию devbox/mise напрямую
3. Генератор делает простой deep merge

### Parsed Dependency

Для удобства работы с распарсенными зависимостями:

```go
// ast/dependencies.go
type ParsedDependency struct {
    Tool    string // canonical tool name
    Version string // version string
    Raw     string // original string from ensure
}
```

### Integration with Razdfile

```go
// ast/razdfile.go
type Razdfile struct {
    // ... existing fields ...
    
    // Dependencies section (mutually exclusive with Mise/Devbox)
    Dependencies *DependenciesConfig `yaml:"dependencies,omitempty"`
}
```

## Validation Rules

### Rule 1: Mutual Exclusion

`dependencies` нельзя использовать одновременно с `mise` или `devbox`:

```go
if rf.HasDependencies() && (rf.HasMise() || rf.HasDevbox()) {
    return &MutualExclusionError{
        Field1: "dependencies",
        Field2: "mise or devbox",
    }
}
```

### Rule 2: Using Required

Если `dependencies` присутствует, поле `using` обязательно:

```go
if rf.Dependencies != nil && rf.Dependencies.Using == "" {
    return &ValidationError{
        Field:   "dependencies.using",
        Message: "using is required when dependencies is specified",
    }
}
```

### Rule 3: Valid Using Value

Значение `using` должно быть "mise" или "devbox":

```go
validUsing := map[string]bool{"mise": true, "devbox": true}
if !validUsing[rf.Dependencies.Using] {
    return &ValidationError{
        Field:   "dependencies.using",
        Message: "using must be 'mise' or 'devbox'",
    }
}
```

### Rule 4: Ensure Format

Каждая строка в `ensure` должна соответствовать формату `tool@version`:

```go
var ensureRegex = regexp.MustCompile(`^[a-z][a-z0-9_-]*@[a-zA-Z0-9._-]+$`)

for _, dep := range rf.Dependencies.Ensure {
    if !ensureRegex.MatchString(dep) {
        return &InvalidDependencyFormatError{
            Value: dep,
        }
    }
}
```

## Trade-offs

### Simplicity vs Power

**Выбор**: Простота + Pass-through

- `dependencies.ensure` — простой список `tool@version` для базовых случаев
- `dependencies.extra` — pass-through для продвинутых настроек (hooks, scripts, env)
- Пользователь использует документацию devbox/mise напрямую для `extra`

### Canonical Names vs Native Names

**Выбор**: Canonical names в Razdfile

Пользователь пишет `node`, а не `nodejs`. Razd транслирует в нужный формат. Это упрощает миграцию между менеджерами.

### Validation at Parse vs Runtime

**Выбор**: Базовая валидация при парсинге

- Формат `ensure` строки проверяется при парсинге
- Структура `extra` **не валидируется** — передаётся как есть
- Существование пакета проверяется при runtime (вызов mise/devbox)

## Generation Logic (Runtime)

Алгоритм генерации конфига (для будущей реализации):

### Devbox

```go
func GenerateDevboxJSON(deps *DependenciesConfig) map[string]any {
    result := map[string]any{
        "packages": translatePackages(deps.Ensure, "devbox"),
    }
    
    if deps.Extra != nil && deps.Extra.Devbox != nil {
        result = deepMerge(result, deps.Extra.Devbox)
    }
    
    return result
}
```

### Mise

```go
func GenerateMiseConfig(deps *DependenciesConfig) map[string]any {
    result := map[string]any{
        "tools": translateTools(deps.Ensure),
    }
    
    if deps.Extra != nil && deps.Extra.Mise != nil {
        result = deepMerge(result, deps.Extra.Mise)
    }
    
    return result
}
```

**Merge strategy**: `extra` переопределяет сгенерированные значения. Если в `extra.devbox.packages` есть дополнительные пакеты, они добавляются к `ensure`.

## Runtime Architecture (Future Implementation)

Для расширяемости используется паттерн **Strategy + Provisioner Registry**.

### Provisioner Interface

Каждый менеджер пакетов реализует общий интерфейс:

```go
// provisioner/provisioner.go
type Provisioner interface {
    // Name returns the provisioner identifier (mise, devbox, ...)
    Name() string
    
    // GenerateConfig creates the native config file
    // packages: parsed ensure list, extra: pass-through config
    GenerateConfig(packages []ParsedDependency, extra map[string]any) error
    
    // Install runs the package installation command
    Install(ctx context.Context) error
    
    // RunCommand wraps a command to run in the provisioner's environment
    // Returns: ["mise", "exec", "--", ...cmd] or ["devbox", "run", "--", ...cmd]
    RunCommand(cmd []string) []string
    
    // IsAvailable checks if the provisioner binary is installed
    IsAvailable() bool
}
```

### Provisioner Implementations

```go
// provisioner/mise.go
type MiseProvisioner struct{}

func (m *MiseProvisioner) Name() string { return "mise" }

func (m *MiseProvisioner) GenerateConfig(packages []ParsedDependency, extra map[string]any) error {
    config := map[string]any{
        "tools": m.packagesToTools(packages),
    }
    // Deep merge extra
    if extra != nil {
        config = deepMerge(config, extra)
    }
    return writeTOML("mise.toml", config)
}

func (m *MiseProvisioner) Install(ctx context.Context) error {
    return exec.CommandContext(ctx, "mise", "install").Run()
}

func (m *MiseProvisioner) RunCommand(cmd []string) []string {
    return append([]string{"mise", "exec", "--"}, cmd...)
}
```

```go
// provisioner/devbox.go
type DevboxProvisioner struct{}

func (d *DevboxProvisioner) Name() string { return "devbox" }

func (d *DevboxProvisioner) GenerateConfig(packages []ParsedDependency, extra map[string]any) error {
    config := map[string]any{
        "packages": d.packagesToList(packages),
    }
    // Deep merge extra (shell, env, additional packages)
    if extra != nil {
        config = deepMerge(config, extra)
    }
    return writeJSON("devbox.json", config)
}

func (d *DevboxProvisioner) Install(ctx context.Context) error {
    // devbox is lazy, but explicit install is supported
    return exec.CommandContext(ctx, "devbox", "install").Run()
}

func (d *DevboxProvisioner) RunCommand(cmd []string) []string {
    return append([]string{"devbox", "run", "--"}, cmd...)
}
```

### Provisioner Registry

```go
// provisioner/registry.go
var registry = map[string]Provisioner{
    "mise":   &MiseProvisioner{},
    "devbox": &DevboxProvisioner{},
    // Easy to extend:
    // "asdf":  &AsdfProvisioner{},
    // "pixi":  &PixiProvisioner{},
}

func Get(name string) (Provisioner, error) {
    p, ok := registry[name]
    if !ok {
        return nil, fmt.Errorf("unknown provisioner: %s, supported: %v", 
            name, maps.Keys(registry))
    }
    return p, nil
}

func Supported() []string {
    return maps.Keys(registry)
}
```

### Main Loop (Polymorphic)

```go
// cmd/up.go
func runUp(ctx context.Context, rf *ast.Razdfile) error {
    deps := rf.Dependencies
    
    // 1. Get provisioner by name
    provisioner, err := provisioner.Get(deps.Using)
    if err != nil {
        return err
    }
    
    // 2. Check availability
    if !provisioner.IsAvailable() {
        return fmt.Errorf("%s is not installed", provisioner.Name())
    }
    
    // 3. Get provisioner-specific extra config
    var extra map[string]any
    if deps.Extra != nil {
        switch deps.Using {
        case "mise":
            extra = deps.Extra.Mise
        case "devbox":
            extra = deps.Extra.Devbox
        }
    }
    
    // 4. Generate config (polymorphic)
    packages := deps.ParseEnsure()
    if err := provisioner.GenerateConfig(packages, extra); err != nil {
        return err
    }
    
    // 5. Install (polymorphic)
    return provisioner.Install(ctx)
}
```

### Adding a New Provisioner

Чтобы добавить поддержку нового менеджера (например, `pixi`):

1. Создать `provisioner/pixi.go` с реализацией `Provisioner`
2. Добавить в registry: `"pixi": &PixiProvisioner{}`
3. Добавить `Pixi map[string]any` в `DependenciesExtra`

```yaml
# User's Razdfile.yml
dependencies:
  using: "pixi"
  ensure:
    - "python@3.11"
    - "numpy@latest"
  extra:
    pixi:
      channels: ["conda-forge"]
      system-requirements:
        linux: "5.4"
```

**Open-Closed Principle**: Система открыта для расширения (новые провайдеры), закрыта для модификации (ядро не меняется).

### Note on Scope

Эта архитектура описывает **будущую реализацию**. Текущий proposal фокусируется на AST и парсинге. Runtime реализация — отдельный proposal.

