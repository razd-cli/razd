# Change: Add CLI Architecture

## Why

Razd нуждается в хорошо структурированной CLI архитектуре, которая позволит:
- Легко добавлять новые команды
- Единообразно обрабатывать флаги и опции
- Обеспечить качественный user experience (help, completion, error messages)
- Поддерживать как интерактивный, так и автоматизированный (CI) режимы

Текущий `main.go` содержит hardcoded пути и не имеет структуры команд.

## What Changes

- **Новая структура пакетов** для CLI (`cmd/razd/`, `internal/cli/`, `internal/flags/`)
- **Команды** с субкомандами: `up`, `run`, `install`, `setup`, `dev`, `build`, `list`, `trust`
- **Система флагов** на основе `spf13/pflag` (как в go-task)
- **Trust система** — проверка доверия перед выполнением задач
- **Logger/Output** с поддержкой цветов и уровней вывода
- **Exit codes** для разных типов ошибок
- **Shell completion** для bash/zsh/fish

### Команды

| Команда | Описание |
|---------|----------|
| `razd` (без аргументов) | Запустить проект (run default task) |
| `razd up [url]` | Настроить проект (install tools) или clone + настроить |
| `razd run <task>` | Выполнить произвольную задачу |
| `razd init` | Создать новый Razdfile.yml |
| `razd add <tool@version>` | Добавить зависимости в dependencies.ensure |
| `razd shell` | Запустить интерактивную оболочку с окружением |
| `razd dev` | Запустить dev workflow |
| `razd build` | Собрать проект |
| `razd list` | Показать список задач |
| `razd trust` | Управление доверием к проектам |
| `razd <task>` | Выполнить задачу (если не совпадает с командой) |

## Impact

- Affected specs: новая спека `cli-interface`
- Affected code: 
  - `main.go` → `cmd/razd/main.go`
  - Новые пакеты: `internal/cli/`, `internal/flags/`, `internal/output/`, `internal/trust/`
  - Использует существующий `razdfile/` пакет

### Глобальные флаги (из Rust версии)

| Флаг | Описание |
|------|----------|
| `-t, --taskfile <FILE>` | Путь к taskfile/razdfile |
| `--razdfile <FILE>` | Путь к razdfile (приоритет) |
| `--no-sync` | Пропустить синхронизацию Razdfile ↔ mise.toml |
| `-y, --yes` | Автоматически отвечать "да" |
| `--list` | Показать список задач |
| `-v, --verbose` | Подробный вывод |
| `-s, --silent` | Тихий режим |

## Design Decisions

См. `design.md` для детального описания архитектурных решений.
