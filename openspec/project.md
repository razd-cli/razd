# Project Context

## Purpose

Razd — современный инструмент для настройки проектов, который помогает разработчикам быстро запустить рабочее окружение одной командой.

Razd оркестрирует [Task](https://taskfile.dev/) (автоматизация задач) и опционально интегрируется с менеджерами окружений ([mise](https://github.com/jdx/mise) или [devbox](https://www.jetify.com/devbox)) для one-command инициализации проекта.

### Ключевые возможности

- **Zero Dependency**: Один бинарник `razd` — всё работает из коробки
- **Клонирование + настройка**: `razd up https://github.com/user/repo` — клонирует и настраивает проект
- **Автоопределение**: Автоматически находит `Razdfile.yml`, `mise.toml`, `Taskfile.yml`
- **Automation-friendly**: Флаг `--yes` для CI/CD пайплайнов

### Основные команды

```bash
razd up              # Настроить проект (или клонировать + настроить)
razd install         # Установить инструменты (через mise или devbox)
razd dev             # Запустить dev-сервер
razd build           # Собрать проект
razd run <task>      # Выполнить произвольную задачу
```

## Tech Stack

- **Go** — основной язык
- **go-task/task/v3** — Task runner интегрирован как библиотека (не внешняя зависимость)
- **devbox** — локальное dev окружение для разработки razd

### Опциональные интеграции (выбор пользователя)

- **mise** — управление версиями инструментов (Node.js, Python, Go и др.)
- **devbox** — Nix-based reproducible dev environments

## Project Conventions

### Code Style

- Go стандартные соглашения (`gofmt`, `go vet`)
- Пакеты организованы по функциональности (cmd/, internal/, pkg/)
- Ошибки возвращаются как `error`, обёрнутые через `fmt.Errorf("context: %w", err)`

### Architecture Patterns

- **Native Task Integration**: Task runner встроен как Go библиотека через `github.com/go-task/task/v3`
- **CLI-first**: Основной интерфейс — командная строка
- **Single Binary**: Всё компилируется в один исполняемый файл

### Testing Strategy

- Unit тесты для внутренних пакетов
- Integration тесты для CLI команд
- Примеры проектов в `examples/` для E2E проверки

### Git Workflow

- Основная ветка: `main`
- Feature ветки: `feature/<name>` или `v1-go-rewrite` (текущая миграция)
- Conventional commits рекомендуются

## Domain Context

### Razdfile.yml

Единый конфигурационный файл, объединяющий:

```yaml
# Опционально: управление инструментами через mise
mise:
  tools:
    node: "22"
    python: "3.11"

# Или через devbox (альтернатива mise)
# devbox:
#   packages:
#     - nodejs@22
#     - python@3.11

# Задачи (Taskfile синтаксис)
tasks:
  default:
    cmds:
      - task: install
      - task: dev
  
  install:
    cmds:
      - npm install
  
  dev:
    cmds:
      - npm run dev
```

### Рабочий процесс

1. Пользователь запускает `razd up`
2. Razd читает `Razdfile.yml` (или `mise.toml` / `devbox.json` + `Taskfile.yml`)
3. Устанавливает инструменты через выбранный менеджер (mise или devbox)
4. Выполняет задачи настройки (default task)
5. Проект готов к работе

## Important Constraints

- **Обратная совместимость**: Поддержка существующих `mise.toml` и `Taskfile.yml`
- **go-task API**: Пакет API ещё экспериментальный, могут быть breaking changes
- **Binary size**: ~40-50MB из-за встроенного Task runner

## External Dependencies

- **mise** (опционально): Управление версиями инструментов
- **devbox** (опционально): Nix-based reproducible environments
- **git**: Для клонирования репозиториев
- **github.com/go-task/task/v3**: Встроен как Go библиотека
