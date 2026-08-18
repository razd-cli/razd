# Implementation Plan: `razd init` — интерактивный выбор провижионера (mise / devbox / none) с Windows-гейтом для devbox

Branch: 1.x
Created: 2026-08-18

## Original Request
Когда создается проект через razd то спрашивать пользователя о том какой проект создавать: через mise или devbox. но если это windows система. то есть отобразить select. еще можно оставить чтобы не нужно было указывать mise или devbox, так как пользователь может быть без провижионеров, то есть просто None а если это windows то не давать доступ к выбору пункта Devbox, так как devbox доступен только на Unix системах

## Settings
- Testing: yes
- Logging: verbose
- Docs: yes

## Roadmap Linkage
Milestone: none
Rationale: ROADMAP.md отсутствует в проекте.

## Research Context
Source: none (прямой анализ кода; RESEARCH.md в проекте нет)

**Текущее состояние `razd init`** (`internal/cli/cmd_init.go`):
- `using := flags.Using`; если пусто → `detectProvider(dir)` (ищет `mise.toml`, `.mise.toml`, `.mise.local.toml`, `.tool-versions` → `mise`; `devbox.json` → `devbox`; иначе **fallback `mise`**).
- Валидация `using != "mise" && using != "devbox"` → ошибка.
- `buildInitRazdfile(using)` **всегда** пишет секцию `dependencies.using`, `ensure: []`. Minimal default-таск закомментирован и не создаётся.

**Валидация Razdfile** (`razdfile/validate.go`):
- Rule 2/3: если секция `dependencies` присутствует, `using` обязателен и ∈ `ValidUsing{mise,devbox}` (`ast/dependencies.go`).
- `HasContent()` требует наличия `tasks | includes | mise | devbox | dependencies`. Файл только с `version: "1"` **невалиден** → для варианта `none` нужно добавить минимальный default-таск, иначе Razdfile не пройдёт валидацию при следующем чтении.

**Промпты в проекте** (`internal/sync/confirm.go`, `internal/trust/prompt.go`):
- `charmbracelet/huh v1.0.0`, `huh.NewSelect[int]()` / `huh.NewConfirm()`, `.WithTheme(huh.ThemeCatppuccin())`.
- TTY-гейт: `sync.isTerminal()` / `trust.IsTerminal()` (`term.IsTerminal(int(os.Stdin.Fd()))`).
- `--yes` (`flags.Yes`) автоподтверждает промпты без блокировки.

**Платформа:** `runtime.GOOS == "windows"`; devbox — только Unix. goreleaser собирает windows.

**Поведение без провижионера:**
- `razd up` (cmd_up.go) требует провижионер → ошибка "no provisioner configured". Это ожидаемо для варианта none.
- `razd run` (cmd_run.go) работает без провижионера (`provResolved=false` → задачи выполняются напрямую). Вариант none корректен для run.

**Версия:** сейчас тег `v1.8.0-dev.2`. Релиз `v1.8.0-dev.3` = bump + tag + push (release.yml триггерится на push тега `v*`, goreleaser, `GORELEASER_CURRENT_TAG` уже закреплён за `github.ref_name`).

## Commit Plan
- **Commit 1** (задачи 1-3): `feat: add interactive provisioner selection to init`
- **Commit 2** (задача 4): `feat: support --using none and --yes default in init`
- **Commit 3** (задача 5): `feat: hide devbox option on windows in init`
- **Commit 4** (задачи 6-7): `test: cover init provider selection` / `docs: document init provider options`
- **Commit 5** (задача 8): `chore: release v1.8.0-dev.3`

## Tasks

### Phase 1: Интерактивный выбор провижионера
- [x] Task 1: Добавить TTY-гейт и функцию `promptInitProvider()` в `internal/cli/cmd_init.go`, которая через `huh.NewSelect[string]` показывает выбор: `mise`, `devbox`, `none` (default — `none`). В неинтерактивном режиме возвращает пустую строку без блокировки.
  - LOGGING: DEBUG при входе/выходе, DEBUG при non-TTY (skip), INFO при выборе значения.
  - Files: `internal/cli/cmd_init.go`.
- [x] Task 2: Переписать логику определения `using` в `runInit`: если `flags.Using` задан — использовать его; иначе в интерактивном TTY вызвать `promptInitProvider()`; иначе `detectProvider(dir)`. `detectProvider` меняет финальный fallback с `"mise"` на `"none"`.
  - LOGGING: DEBUG источник (флаг/промпт/автоопределение), INFO итоговый выбор.
  - Files: `internal/cli/cmd_init.go`.
- [x] Task 3: Расширить `buildInitRazdfile(using)`: для `using == "none"` НЕ создавать секцию `dependencies`, а добавлять минимальный default-таск (`tasks.default`) чтобы Razdfile прошёл `HasContent()`; для `mise`/`devbox` сохранить текущее поведение.
  - LOGGING: DEBUG при ветке none vs mise/devbox.
  - Files: `internal/cli/cmd_init.go`.

### Phase 2: Флаг `--using none` и `--yes`
- [x] Task 4: Поддержать `--using none` (и `--using mise|devbox`) — валидация допускает `none` наравне с `mise`/`devbox`. При `--yes` в интерактивном режиме промпт пропускается и используется default (`none`).
  - LOGGING: DEBUG при `--yes` skip, DEBUG при `--using none`.
  - Files: `internal/cli/cmd_init.go`.

### Phase 3: Windows-гейт для devbox
- [x] Task 5: В `promptInitProvider()` исключить пункт `devbox` при `runtime.GOOS == "windows"` (доступны только `mise` и `none`). Остальная логика без изменений.
  - LOGGING: DEBUG с указанием платформы при фильтрации опций.
  - Files: `internal/cli/cmd_init.go`.

### Phase 4: Тесты
- [x] Task 6: Создать `internal/cli/cmd_init_test.go`: `buildInitRazdfile("none")` не содержит `dependencies` и содержит default-таск; `buildInitRazdfile("mise"|"devbox")` пишет `dependencies.using`; `detectProvider` с fallback `none`; `promptInitProvider` в non-TTY возвращает пусто; на Windows вариант devbox отсутствует (через инъекцию GOOS-предиката, если функция принимает его как параметр).
  - LOGGING: нет (тесты).
  - Files: `internal/cli/cmd_init_test.go`.
- [x] Task 7: README: задокументировать `razd init` (интерактивный select), `--using none`, поведение на Windows (без devbox).
  - LOGGING: нет (docs).
  - Files: `README.md`.

### Phase 5: Релиз
- [ ] Task 8: Выпустить версию `v1.8.0-dev.3`: убедиться что `internal/version` собирается корректно, создать тег `v1.8.0-dev.3`, push (триггер release.yml), подтвердить что goreleaser-релиз создан с `name_template: razd v1.8.0-dev.3`.
  - LOGGING: нет.
  - Files: git tags, .github/workflows/release.yml (без изменений).

## Acceptance
- `razd init` в TTY показывает select: mise / devbox / none, default none.
- На Windows select показывает только mise / none (devbox скрыт).
- `razd init --using none` создаёт Razdfile без секции `dependencies` и с default-таском; файл проходит `razdfile.Validate`.
- `razd init --using mise|devbox` сохраняет текущее поведение.
- Non-TTY (`CI`) без `--using` → `detectProvider`: найденная конфигурация wins, иначе `none`.
- `--yes` пропускает промпт с default none.
- Smoke: `razd init` в temp dir → select `none` → `razd list` показывает default-таск, `razd run default` выполняется.
- Версия `v1.8.0-dev.3` отрелижена.
