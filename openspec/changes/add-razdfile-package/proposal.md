# Proposal: Add Razdfile Package

## Summary

Создать пакет `razdfile` для парсинга и валидации `Razdfile.yml` — объединённого формата Taskfile + mise конфигурации. Архитектура будет вдохновлена пакетом `taskfile` из go-task.

## Motivation

1. **Объединённый формат**: Razdfile.yml объединяет Taskfile tasks и mise tool management в одном файле
2. **Валидация**: Нужна валидация структуры до выполнения задач
3. **Расширяемость**: Пакет должен поддерживать добавление devbox интеграции в будущем
4. **Соответствие go-task**: Архитектура схожа с `taskfile/` пакетом для консистентности

## Proposed Solution

### Package Structure

```
razdfile/
├── ast/                    # Abstract Syntax Tree types
│   ├── razdfile.go         # Main Razdfile struct
│   ├── mise.go             # Mise config types (tools, env, settings)
│   ├── task.go             # Task types (re-export or wrap taskfile/ast)
│   └── version.go          # Version handling
├── reader.go               # Reader with functional options pattern
├── node.go                 # Node interface (file, remote in future)
├── node_file.go            # FileNode implementation
├── validate.go             # Validation logic
└── errors.go               # Razdfile-specific errors
```

### Key Types

```go
// ast/razdfile.go
type Razdfile struct {
    Location string
    Version  string              // "1"
    Mise     *MiseConfig         // mise section
    Tasks    *taskfile_ast.Tasks // reuse from go-task
    Vars     *taskfile_ast.Vars
    Env      *taskfile_ast.Vars
    // ... other Taskfile fields
}

// ast/mise.go
type MiseConfig struct {
    Tools    map[string]MiseTool
    Env      map[string]any
    Settings *MiseSettings
    Hooks    map[string]any
    Plugins  map[string]string
}

type MiseTool struct {
    Version    string
    OS         []string
    InstallEnv map[string]any
    Postinstall string
}
```

### Reader Pattern

Следуем functional options pattern как в go-task:

```go
reader := razdfile.NewReader(
    razdfile.WithDebugFunc(log.Println),
)
rf, err := reader.Read(ctx, node)
```

## Scope

### In Scope
- AST types для Razdfile.yml структуры
- Парсинг YAML в AST
- Базовая валидация (version, required fields)
- FileNode для чтения локальных файлов
- Интеграция с существующим go-task AST для tasks секции

### Out of Scope
- Remote nodes (HTTP, Git) — будущая работа
- Devbox интеграция в AST — отдельный proposal
- CLI команда `razd validate` — будет использовать этот пакет

## Success Criteria

1. `razdfile.Reader` успешно парсит все примеры в `examples/`
2. Валидация выдаёт понятные ошибки с line/column информацией
3. Mise секция корректно парсится в структуры
4. Tasks секция совместима с go-task execution

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| go-task AST API нестабилен | Минимальная зависимость, wrap types если нужно |
| Дублирование кода с taskfile | Реиспользовать taskfile/ast types где возможно |
| Сложная mise schema | Начать с базовых полей, расширять итеративно |

## Alternatives Considered

1. **Форк taskfile пакета**: Слишком много кода, сложно поддерживать
2. **Просто добавить mise в main.go**: Не масштабируется, нет валидации
3. **Использовать JSON Schema для валидации**: Runtime validation без compile-time types

## References

- [go-task/taskfile](https://github.com/go-task/task/tree/main/taskfile) — reference architecture
- [schemas/razdfile.json](../../schemas/razdfile.json) — JSON Schema для Razdfile
