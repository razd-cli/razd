# Implementation Plan: Installer — стабильная версия по умолчанию и поддержка `-dev.0` пре-релизов

Branch: 1.x (no-switch, без создания ветки)
Created: 2026-08-09

## Original Request

> добавить еще задачу проверить /root/razd-workflow/installer что там ставится стабильная версия и можно указать параметры для -dev.0 подобные пре-релизы

## Settings

- Testing: yes
- Logging: verbose
- Docs: yes

## Roadmap Linkage

Milestone: none
Rationale: Роадмап не существует — пропущено пользователем.

## Background (проверено перед планированием)

Живые данные GitHub API (2026-08-09):
- `/releases/latest` → `v1.5.0` (stable), `prerelease=false`
- `v1.6.0-dev.0` → `prerelease=true`

Логика в `install.sh` / `install.ps1` уже поддерживает запрошенное:
- По умолчанию (`latest`) качают `v1.5.0` — stable, не трогая prerelease.
- `--pre-release` / `-PreRelease` → `get_latest_prerelease_version()`.
- `--version 1.6.0-dev.0` / `-Version` / `RAZD_VERSION` → точная версия.
- `is_prerelease()`/`Test-Prerelease` предупреждает о пре-релизе.

Пробел: `test-installer.yml` не покрывает эти ветки. Нужно добавить CI-проверки и зафиксировать поведение.

## Tasks

### Phase 1: Верификация поведения скриптов
- [x] Task 1: Проверить `get_latest_version()` / `get_latest_prerelease_version()` / `resolve_version()` / `is_prerelease()` в `install.sh` — подтвердить, что default=stable, `--pre-release`=последний prerelease, `--version X.Y.Z-dev.N`=точная версия
- [x] Task 2: Проверить то же в `install.ps1` (`Get-LatestVersion`, `Get-LatestPrereleaseVersion`, `Resolve-Version`, `Test-Prerelease`)

### Phase 2: Добавление CI-тестов
- [x] Task 3: В `test-installer.yml` добавить шаг: default-установка НЕ ставит prerelease (убедиться, что установился `v1.5.0`, а не `v1.6.0-dev.0`)
- [x] Task 4: В `test-installer.yml` добавить шаг: `./install.sh --pre-release` ставит последний prerelease
- [x] Task 5: В `test-installer.yml` добавить шаг: `RAZD_VERSION=1.6.0-dev.0 ./install.sh` (и `--version 1.6.0-dev.0`) ставит точную dev-версию
- [x] Task 6: В `test-installer.yml` добавить Windows-шаг: `-PreRelease` и `-Version 1.6.0-dev.0` в `install.ps1`

### Phase 3: Документация
- [x] Task 7: Обновить `README.md` — явно указать, что default ставит stable, и расширить примеры `-dev.0` пре-релизов

### Phase 4: Проверка
- [x] Task 8: `bash -n install.sh` и PowerShell syntax-check `install.ps1` (без реальной установки)
- [x] Task 9: Локальный прогон `./install.sh --list` и `./install.sh --pre-release` на linux (с `RAZD_INSTALL_DIR` в temp) — убедиться, что версии резолвятся корректно

## Commit Plan

- **Commit 1** (tasks 1-2): `chore: verify installer version resolution logic`
- **Commit 2** (tasks 3-6): `ci(installer): cover stable-default and -dev pre-release installs`
- **Commit 3** (task 7): `docs(installer): document pre-release install options`
