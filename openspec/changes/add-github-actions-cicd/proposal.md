# Change: Add GitHub Actions CI/CD

## Why

Razd нуждается в автоматизированной CI/CD инфраструктуре для:
- Автоматического тестирования на каждый PR и push
- Lint проверок кода (golangci-lint)
- Кросс-платформенной сборки (Windows, Linux, macOS)
- Автоматических релизов при создании тегов
- Публикации бинарников на GitHub Releases

Сейчас нет никакой автоматизации сборки и релизов.

## What Changes

- **GitHub Actions workflows** для lint, test, build, release
- **GoReleaser** конфигурация для кросс-платформенной сборки
- **Multi-platform testing** на ubuntu, macos, windows
- **Automated releases** при создании тегов `v*`

### Новые файлы

| Файл | Описание |
|------|----------|
| `.github/workflows/lint.yml` | Проверка кода golangci-lint |
| `.github/workflows/test.yml` | Кросс-платформенные тесты |
| `.github/workflows/release.yml` | Сборка и публикация релизов |
| `.goreleaser.yml` | Конфигурация GoReleaser |

### Workflow детали

1. **Lint** (на каждый PR и push в main):
   - Проверка golangci-lint
   - Go версии: 1.24.x, 1.25.x

2. **Test** (на каждый PR и push в main):
   - Matrix: ubuntu-latest, macos-latest, windows-latest
   - Go версии: 1.24.x, 1.25.x
   - Build + Test

3. **Release** (при push тега `v*`):
   - GoReleaser собирает бинарники
   - Платформы: Linux (amd64, arm64), macOS (amd64, arm64), Windows (amd64, arm64)
   - Публикация на GitHub Releases

## Impact

- Affected specs: новая спека `ci-cd`
- Affected code: 
  - Новые файлы `.github/workflows/*.yml`
  - Новый файл `.goreleaser.yml`
  - Обновление `internal/version/version.go` для ldflags
