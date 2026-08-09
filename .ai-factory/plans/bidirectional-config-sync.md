# Implementation Plan: Двусторонняя синхронизация Razdfile ↔ mise.toml / devbox.json

Branch: 1.x (no-switch, без создания ветки)
Created: 2026-08-09

## Original Request

> есть проблема с синхронизации файлов с mise.toml то есть когда добавляю пакет в razdfile то не синхронизируется с mise.toml то есть очень плохо синхронизируется или перезаписывается

> короче говоря, если я добавляю пакет в Razdfile то добавлять тоже аккуратно в mise.toml. если в mise.toml добавилось, то добавлять и в Razdfile аккуратно. в обоих конфигах могут быть важные данные и важная конфигурация, поэтому нужно делать аккуратную синхронизацию, и делать бекап тоже спрашивать пользователя

> Сделать бекап mise.toml перед перезаписью? y\N (по дефолту Нет - N )

## Settings

- Testing: yes
- Logging: verbose
- Docs: yes

## Roadmap Linkage

Milestone: none
Rationale: Роадмап не существует — пропущено пользователем.

## Background (из explore, проверено в коде)

Текущие дефекты (найдены при анализе `provisioner/*`, `internal/cli/*`, `razdfile/*`):
- `MiseProvisioner.GenerateConfig` (provisioner/mise.go:31-61) **целиком перезаписывает** `mise.toml` из `dependencies.ensure`, теряя `[env]`, `[settings]`, `[hooks]`, `[plugins]` и ручные правки.
- `razd add` (internal/cli/cmd_add.go) пишет только в `Razdfile.yml`, не генерирует/не синхронизирует `mise.toml`.
- `syncConfig` (internal/cli/helpers.go:211-283) — только направление mise→Razdfile, новые инструменты из mise игнорируются.
- `cmd_up.go` вызывает `syncConfig` затем `generateProvisionerConfig` — две однонаправленные операции, потенциально «съедающие» друг друга.
- `ReadConfig` (provisioner/mise.go:68-97) — наивный построчный сканер, ломается на complex tools и sequence-инструментах, не использует готовые `ast.MiseConfig`/`ast.MiseTool`.
- `flags.Backup`/`--backup` уже существует, но это флаг, а не интерактивный запрос y/N.
- В `trust/prompt.go` есть готовый паттерн интерактивного промпта (`huh`) с non-TTY fallback — переиспользовать для backup-промпта.

Цель: настоящий двусторонний sync с merge (без потери данных ни в одном файле), интерактивный запрос бекапа перед перезаписью (по умолчанию N), для mise и devbox.

## Design

### Единый merge-движок

Ввести пакет `internal/sync` с единой моделью «инструмент → версия» + merge-логикой для обоих провижнеров.

```
                    sync.Engine
   ┌─────────────────┴─────────────────┐
   │  reconcile(razdfileTools,         │
   │            nativeTools)           │
   └───────────┬───────────┬───────────┘
               │           │
      mise.toml │           │ devbox.json
   (TOML merge)│           │ (JSON merge)
```

- **Исходные данные (из Razdfile):** `dependencies.ensure` (`tool@version`) + `extra.mise`/`extra.devbox` + топ-уровневые `mise:`/`devbox:` секции (при их наличии).
- **Нативные данные:** полный parse `mise.toml` (через TOML-библиотеку) / `devbox.json` (через json).
- **Merge-правила:**
  1. Инструмент есть в Razdfile, нет в native → добавить в native.
  2. Инструмент есть в native, нет в Razdfile → добавить в Razdfile.
  3. Инструмент есть в обоих, версии совпадают → без изменений.
  4. Инструмент есть в обоих, версии различаются → **спросить пользователя**, какой приоритет (Razdfile / native / пропустить).
  5. Все секции native-файла вне `[tools]`/`packages` **сохраняются как есть** (merge, не перезапись).

### Парсинг/сериализация

- Добавить зависимость `github.com/pelletier/go-toml/v2` для корректного TOML parse/marshal (сейчас TOML-библиотеки в go.mod нет).
- Убрать наивный сканер `MiseProvisioner.ReadConfig` и `writeTomlMap`; заменить на go-toml.
- `ast.MiseConfig`/`ast.MiseTool` (уже умеют complex/sequence) использовать для валидации/моделирования tools из Razdfile.
- Для devbox использовать существующий json-парсинг, но с merge-сохранением неизвестных секций.

### Бекап

- Интерактивный промпт `Make backup of <config> before overwrite? [y/N]` (default N), переиспользуя `huh`-паттерн из `trust/prompt.go` с non-TTY fallback (в CI — без бекапа, продолжить).
- Бекап: копия файла с timestamp-суффиксом рядом (`.mise.toml.bak.<ts>`), паттерн уже есть в `backupNativeConfig` (helpers.go).
- `--backup` флаг сохраняется как принудительный бекап без вопроса.

### Интеграция в команды

- **`razd up`**: заменить последовательность `syncConfig` + `generateProvisionerConfig` на один вызов `sync.Sync` (двусторонний), затем `Install`.
- **`razd add`**: после записи в Razdfile вызвать `sync.Sync` (+ при желании `Install`), чтобы `mise.toml` обновился сразу.
- **`razd init`**: существующая генерация остаётся, но тоже через merge-движок.

### Разрешение конфликтов

Интерактивный промпт (по результатам опроса: «Спрашивать при конфликте»):
```
Version conflict for node: Razdfile has 22, mise.toml has 23.
[1] Use Razdfile (22)   [2] Use mise.toml (23)   [3] Skip (keep both unchanged)
```
Non-TTY fallback: WARN и не менять (безопасный дефолт — ничего не перезаписывать).

## Tasks

### Phase 1: Фундамент — merge-движок и TOML
- [x] Task 1: Добавить `github.com/pelletier/go-toml/v2` в go.mod (`go get`) и убедиться, что проект собирается (`go build ./...`)
- [x] Task 2: Создать `internal/sync/engine.go` — модель `Tool{Name, Version}` + функция `Reconcile(razdTools, nativeTools []Tool, resolve func(conflict) (Resolution, error))`, реализующая merge-правила 1–5

### Phase 2: Парсинг/сериализация mise и devbox
- [x] Task 3: В `provisioner/mise.go` заменить `ReadConfig` (наивный сканер) на parse через go-toml, возвращающий все секции `mise.toml` (не только tools), и добавить метод `WriteTools(native map, tools []Tool) error` для merge-записи, сохраняющей прочие секции
- [x] Task 4: В `provisioner/devbox.go` заменить `ReadConfig`/`GenerateConfig` на json-parse с merge-сохранением неизвестных ключей + метод записи packages по тем же правилам

### Phase 3: Бекап и конфликты
- [x] Task 5: Создать `internal/sync/backup.go` — `BackupFile(path string, log) error` (timestamp-копия) + интерактивный промпт `PromptBackup(config string) (bool, error)` на базе `huh` из `trust/prompt.go`, default N, non-TTY → false (без бекапа)
- [x] Task 6: Создать `internal/sync/conflict.go` — `PromptConflict(tool string, razdVersion, nativeVersion string) (Resolution, error)` c вариантами [1]Razdfile [2]Native [3]Skip, non-TTY → Skip с WARN

### Phase 4: Интеграция в команды
- [x] Task 7: В `internal/cli/helpers.go` заменить `syncConfig` + `generateProvisionerConfig` на единый `sync.Sync(rf, prov, dir, log)` (двусторонний merge, вызывает бекап-промпт при записи в native, разрешает конфликты через промпт)
- [x] Task 8: В `internal/cli/cmd_up.go` вызывать `sync.Sync` вместо текущих двух шагов; сохранить флаг `--no-sync`
- [x] Task 9: В `internal/cli/cmd_add.go` после `UpdateEnsureInFile` вызвать `sync.Sync` для немедленной синхронизации `mise.toml`/`devbox.json`
- [x] Task 10: В `internal/cli/cmd_init.go` перевести генерацию на merge-движок

### Phase 5: Тесты
- [x] Task 11: `internal/sync/engine_test.go` — все merge-правила 1–5, порядок, сохранение секций
- [x] Task 12: `internal/sync/backup_test.go` + `conflict_test.go` — timestamp-бекап, non-TTY fallback, разрешение конфликтов
- [x] Task 13: `provisioner/mise_test.go` + `devbox_test.go` — merge-запись сохраняет `[env]`/`[settings]`/ручные секции, complex tools, идемпотентность
- [x] Task 14: `internal/cli/cmd_add_test.go` + `cmd_up_test.go` — add синхронизирует, up делает двусторонний sync, `--no-sync` отключает

### Phase 6: Документация
- [x] Task 15: Обновить `README.md` — описать двустороннюю синхронизацию, промпт бекапа, разрешение конфликтов, `--no-sync`/`--backup`

## Commit Plan

- **Commit 1** (tasks 1-4): `feat(sync): add TOML dep and bidirectional merge engine with parsers`
- **Commit 2** (tasks 5-6): `feat(sync): backup prompt and conflict resolution`
- **Commit 3** (tasks 7-10): `feat(cli): wire bidirectional sync into up/add/init`
- **Commit 4** (tasks 11-14): `test(sync): merge, backup, conflict, cli integration`
- **Commit 5** (task 15): `docs: document config sync behavior`

## Edge Cases

- **Конфликт версий** (node@22 vs node@23): спрашивать пользователя; non-TTY → Skip (ничего не менять).
- **Complex tools** (`node = { version = "22", os = [...] }`): сохранять как есть при merge, не упрощать до скаляра.
- **Sequence tools** (`node = ["20", "22"]`): primary version для sync-логики, остальное сохранять.
- **Секции вне tools** (`[env]`, `[settings]`, `[tasks]`): всегда сохранять.
- **Нет native-файла**: считать пустым, создать при первом sync.
- **Нет dependencies в Razdfile**: sync не выполняется (early return).
- **Non-TTY (CI)**: бекап и конфликт-промпты не блокируют — безопасные дефолты (без бекапа, Skip).
- **`--backup` флаг**: принудительный бекап без промпта.
