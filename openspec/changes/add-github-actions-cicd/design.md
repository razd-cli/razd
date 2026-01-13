# Design: GitHub Actions CI/CD

## Overview

CI/CD архитектура вдохновлена go-task/task, но упрощена под нужды razd:
- Бесплатный GoReleaser (не Pro)
- Минимальная конфигурация
- Кросс-платформенная сборка

## Архитектурные решения

### 1. Структура Workflows

```
.github/
└── workflows/
    ├── lint.yml          # Lint на каждый PR/push
    ├── test.yml          # Тесты на всех платформах
    └── release.yml       # Релизы при tag v*
```

**Обоснование**:
- Разделение на отдельные workflows для параллельного выполнения
- Lint и test запускаются независимо
- Release только при тегах

### 2. Matrix Strategy

**Test workflow**:
```yaml
strategy:
  matrix:
    go-version: [1.24.x, 1.25.x]
    platform: [ubuntu-latest, macos-latest, windows-latest]
```

- Тестируем на двух последних версиях Go
- Все три основные платформы

### 3. GoReleaser Configuration

**Выбор: бесплатная версия GoReleaser**

Причины:
- Достаточно для базовых нужд
- Нет необходимости в Pro фичах (nightly, signing)
- Простая интеграция с GitHub Actions

**Целевые платформы**:
- Linux: amd64, arm64
- macOS: amd64, arm64 (Universal Binary)
- Windows: amd64, arm64

**Формат архивов**:
- Linux/macOS: `.tar.gz`
- Windows: `.zip`

### 4. Version Injection

Использование ldflags для внедрения версии при сборке:

```go
// internal/version/version.go
var (
    Version = "dev"       // -X main.Version=v1.0.0
    Commit  = "unknown"   // -X main.Commit=abc123
    Date    = "unknown"   // -X main.Date=2025-01-13
)
```

GoReleaser автоматически подставляет значения через ldflags.

### 5. Альтернативы рассмотренные

| Инструмент | Плюсы | Минусы | Решение |
|------------|-------|--------|---------|
| GoReleaser | Стандарт для Go, простой | — | ✅ Выбран |
| Makefile + scripts | Полный контроль | Много кода | ❌ |
| Ko | Отлично для контейнеров | Не для CLI | ❌ |
| xgo | Кросс-компиляция CGO | Нет CGO в razd | ❌ |

### 6. Naming Convention

Имена артефактов:
```
razd_{{ .Os }}_{{ .Arch }}{{ .Arm }}.{{ .Format }}
```

Примеры:
- `razd_linux_amd64.tar.gz`
- `razd_darwin_arm64.tar.gz`
- `razd_windows_amd64.zip`

### 7. Checksums

GoReleaser автоматически создаёт:
- `checksums.txt` — SHA256 хеши всех артефактов

### 8. Workflow Permissions

Минимальные права:
- `contents: write` — для создания релизов
- `id-token: write` — опционально для OIDC (не используем сейчас)

## Диаграмма CI/CD Pipeline

```
┌─────────────────────────────────────────────────────────────────┐
│                        Push / PR                                 │
└────────────────────────────┬────────────────────────────────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
              ▼              ▼              ▼
        ┌─────────┐    ┌─────────┐    ┌─────────┐
        │  Lint   │    │  Test   │    │  Test   │
        │ ubuntu  │    │ ubuntu  │    │ macos   │  ...
        └────┬────┘    └────┬────┘    └────┬────┘
              │              │              │
              └──────────────┼──────────────┘
                             │
                        (all pass)
                             │
                             ▼
                    ┌─────────────────┐
                    │   ✓ Ready to    │
                    │     merge       │
                    └─────────────────┘


┌─────────────────────────────────────────────────────────────────┐
│                    Push Tag v*                                   │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │   GoReleaser    │
                    │   Build + Pub   │
                    └────────┬────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
              ▼              ▼              ▼
        ┌─────────┐    ┌─────────┐    ┌─────────┐
        │ Linux   │    │ macOS   │    │ Windows │
        │ amd64   │    │ arm64   │    │ amd64   │
        │ arm64   │    │ amd64   │    │ arm64   │
        └────┬────┘    └────┬────┘    └────┬────┘
              │              │              │
              └──────────────┼──────────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │ GitHub Release  │
                    │  + Artifacts    │
                    └─────────────────┘
```

## Security Considerations

1. **Secrets**: Используем только `GITHUB_TOKEN` (автоматически предоставляется)
2. **Permissions**: Минимальные права `contents: write`
3. **Dependencies**: Pinned версии actions (`@v6`, `@v9`)
