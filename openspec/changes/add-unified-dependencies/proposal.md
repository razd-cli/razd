# Proposal: Add Unified Dependencies Section

## Summary

Добавить новую секцию `dependencies` в Razdfile.yml — унифицированный способ указания зависимостей инструментов, который автоматически транслируется в формат mise или devbox.

## Motivation

1. **Единый формат**: Сейчас пользователь должен знать синтаксис mise (`tools:`) или devbox (`packages:`). Предлагается простой, "пуленепробиваемый" формат `ensure:` со списком строк.

2. **Абстракция от реализации**: Пользователь указывает `node@22`, а razd генерирует правильный манифест в зависимости от `using: mise` или `using: devbox`.

3. **Миграция между инструментами**: Легко переключиться между mise и devbox, изменив только `using:`.

4. **Простота для новичков**: Не нужно изучать особенности каждого менеджера пакетов.

## Proposed Solution

### Новая секция dependencies

```yaml
version: "1"

dependencies:
  using: "mise"    # или "devbox"
  
  ensure:
    - "node@22"
    - "php@8.4"
    - "pnpm@latest"

  # Pass-through для нативных настроек (опционально)
  extra:
    mise:
      env:
        NODE_ENV: development
```

### Трансляция в mise

Из:
```yaml
dependencies:
  using: "mise"
  ensure:
    - "node@22"
    - "php@8.4"
    - "pnpm@latest"
```

В (виртуальный mise.toml или внутреннее представление):
```toml
[tools]
node = "22"
php = "8.4"
pnpm = "latest"
```

### Трансляция в devbox

Из:
```yaml
dependencies:
  using: "devbox"
  ensure:
    - "node@22"
    - "php@8.4"
    - "pnpm@latest"
```

В (виртуальный devbox.json или внутреннее представление):
```json
{
  "packages": [
    "nodejs@22",
    "php@8.4",
    "pnpm@latest"
  ]
}
```

### Формат ensure строк

Единый формат `<tool>@<version>`:

| Razdfile ensure | mise tools | devbox packages |
|-----------------|------------|-----------------|
| `node@22` | `node = "22"` | `nodejs@22` |
| `php@8.4` | `php = "8.4"` | `php@8.4` |
| `pnpm@latest` | `pnpm = "latest"` | `pnpm@latest` |
| `python@3.11` | `python = "3.11"` | `python@3.11` |

**Маппинг имён**: Некоторые пакеты имеют разные имена (node → nodejs в devbox). Razd должен знать о таких случаях.

### Секция extra для pass-through конфигурации

Для расширенных настроек (hooks, scripts, env) используется `extra`:

```yaml
dependencies:
  using: "devbox"
  ensure:
    - "go@1.21"
    - "ripgrep@latest"

  extra:
    devbox:
      # Структура повторяет devbox.json
      shell:
        init_hook: |
          echo "Welcome to Dev Environment!"
          export GOPATH=$PWD/.go
        scripts:
          build: "go build ."
          test: "go test ./..."
      env:
        MY_VAR: "production"

    mise:
      # Структура повторяет mise.toml
      env:
        NODE_ENV: development
      settings:
        experimental: true
```

**Логика генерации**: merge `ensure` пакетов + `extra.<manager>` как есть.

### Совместимость с существующими секциями

Секция `dependencies` — **альтернатива** секциям `mise` и `devbox`. Одновременное использование запрещено:

```yaml
# ❌ Ошибка: нельзя использовать dependencies вместе с mise/devbox
dependencies:
  using: "mise"
  ensure: ["node@22"]

mise:
  tools:
    node: "22"
```

## Impact

### Affected Specs
- `razdfile-parsing` — новый тип AST для dependencies секции

### Affected Code
- `razdfile/ast/` — новые типы
- `razdfile/validate.go` — валидация взаимоисключения
- `razdfile/reader.go` — парсинг новой секции

### Breaking Changes
- **Нет** — это новая опциональная секция

## Scope

### In Scope
- AST тип `DependenciesConfig` с `ensure` и `extra`
- Парсинг `dependencies:` секции
- Валидация формата `ensure` строк
- Валидация взаимоисключения с `mise`/`devbox`
- Pass-through `extra` секции (без валидации внутренней структуры)
- Документация формата

### Out of Scope (Future Work)
- Фактическая генерация mise.toml/devbox.json файлов
- Runtime выполнение `mise install` / `devbox install`
- Валидация структуры внутри `extra.devbox` / `extra.mise`

## Success Criteria

1. `dependencies` секция корректно парсится в AST
2. Валидация отклоняет одновременное использование `dependencies` + `mise`/`devbox`
3. Формат `tool@version` валидируется с понятными ошибками
4. Unit тесты покрывают все сценарии

## Open Questions

1. **Маппинг имён пакетов**: Нужна ли таблица маппинга (node→nodejs) в AST или это runtime concern?
   - **Рекомендация**: Runtime concern

2. **Devbox extensions / Nix flakes**: Как обрабатывать `php84Extensions.xdebug` или `github:...`?
   - **Решение**: Использовать `extra.devbox.packages` для добавления сложных пакетов — они будут merged с `ensure`

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Несовместимые имена пакетов | Таблица маппинга + документация |
| Ограниченные возможности vs нативные секции | Чёткая документация "когда использовать dependencies vs mise/devbox" |
| Версии могут различаться между менеджерами | Документировать известные различия |

## Related Changes

### `add-cli-architecture` (дополняет)

Этот proposal **дополняет** `add-cli-architecture`:

| Область | add-cli-architecture | add-unified-dependencies |
|---------|---------------------|--------------------------|
| Фокус | CLI команды и структура | AST и Provisioner интерфейс |
| `razd install` | Вызывает provisioner | Предоставляет Provisioner |
| Razdfile | Использует `razdfile.Reader` | Расширяет AST полем `Dependencies` |

**Порядок имплементации**:
1. `add-razdfile-package` — базовый парсинг Razdfile (✓ Complete)
2. `add-unified-dependencies` — AST для `dependencies` + Provisioner интерфейс
3. `add-cli-architecture` — CLI, использует Provisioner для `razd install`

**Нет конфликтов**: Proposals работают на разных уровнях абстракции.
