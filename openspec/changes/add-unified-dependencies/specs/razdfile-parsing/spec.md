# Razdfile Parsing — Dependencies Delta

Расширение парсинга Razdfile для поддержки секции `dependencies`.

## ADDED Requirements

### Requirement: Parse dependencies section

Система MUST парсить секцию `dependencies` в типизированную структуру.

#### Scenario: Parse dependencies with mise

**Given** файл `Razdfile.yml` с содержимым:
```yaml
version: "1"

dependencies:
  using: "mise"
  ensure:
    - "node@22"
    - "php@8.4"
    - "pnpm@latest"

tasks:
  install:
    cmds:
      - npm install
```

**When** вызывается `razdfile.Reader.Read(ctx, node)`

**Then** возвращается `*ast.Razdfile` где:
- `Dependencies.Using == "mise"`
- `Dependencies.Ensure == ["node@22", "php@8.4", "pnpm@latest"]`
- `Mise == nil`
- `Devbox == nil`

#### Scenario: Parse dependencies with devbox

**Given** файл `Razdfile.yml` с содержимым:
```yaml
version: "1"

dependencies:
  using: "devbox"
  ensure:
    - "node@22"
    - "python@3.11"

tasks:
  install:
    cmds:
      - npm install
```

**When** вызывается `razdfile.Reader.Read(ctx, node)`

**Then** возвращается `*ast.Razdfile` где:
- `Dependencies.Using == "devbox"`
- `Dependencies.Ensure == ["node@22", "python@3.11"]`

#### Scenario: Parse dependencies with empty ensure

**Given** файл `Razdfile.yml` с содержимым:
```yaml
version: "1"

dependencies:
  using: "mise"
  ensure: []

tasks:
  hello:
    cmds:
      - echo "hello"
```

**When** вызывается `razdfile.Reader.Read(ctx, node)`

**Then** возвращается `*ast.Razdfile` где:
- `Dependencies.Using == "mise"`
- `Dependencies.Ensure` — пустой слайс
- Парсинг успешен (пустой ensure допустим)

---

### Requirement: Validate dependencies using field

Система MUST валидировать поле `using` в секции `dependencies`.

#### Scenario: Missing using field

**Given** файл `Razdfile.yml` с содержимым:
```yaml
version: "1"

dependencies:
  ensure:
    - "node@22"
```

**When** вызывается `razdfile.Reader.Read(ctx, node)`

**Then** возвращается ошибка валидации:
- Сообщение содержит "using is required"

#### Scenario: Invalid using value

**Given** файл `Razdfile.yml` с содержимым:
```yaml
version: "1"

dependencies:
  using: "npm"
  ensure:
    - "node@22"
```

**When** вызывается `razdfile.Reader.Read(ctx, node)`

**Then** возвращается ошибка валидации:
- Сообщение содержит "using must be 'mise' or 'devbox'"

#### Scenario: Valid using mise

**Given** файл с `dependencies.using: "mise"`

**When** Reader парсит файл

**Then** парсинг успешен

#### Scenario: Valid using devbox

**Given** файл с `dependencies.using: "devbox"`

**When** Reader парсит файл

**Then** парсинг успешен

---

### Requirement: Validate dependencies ensure format

Система MUST валидировать формат строк в `ensure`.

#### Scenario: Valid ensure format

**Given** файл с:
```yaml
dependencies:
  using: "mise"
  ensure:
    - "node@22"
    - "php@8.4"
    - "pnpm@latest"
    - "python@3.11.5"
```

**When** Reader парсит файл

**Then** парсинг успешен

#### Scenario: Invalid ensure format — missing version

**Given** файл с:
```yaml
dependencies:
  using: "mise"
  ensure:
    - "node"
```

**When** Reader парсит файл

**Then** возвращается ошибка валидации:
- Сообщение содержит "invalid dependency format 'node'"
- Сообщение содержит "expected 'tool@version'"

#### Scenario: Invalid ensure format — empty string

**Given** файл с:
```yaml
dependencies:
  using: "mise"
  ensure:
    - ""
```

**When** Reader парсит файл

**Then** возвращается ошибка валидации:
- Сообщение содержит "invalid dependency format"

#### Scenario: Invalid ensure format — special characters

**Given** файл с:
```yaml
dependencies:
  using: "mise"
  ensure:
    - "node@22!"
```

**When** Reader парсит файл

**Then** возвращается ошибка валидации

---

### Requirement: Mutual exclusion with mise/devbox sections

Система MUST запрещать одновременное использование `dependencies` с `mise` или `devbox`.

#### Scenario: Dependencies with mise section

**Given** файл `Razdfile.yml` с содержимым:
```yaml
version: "1"

dependencies:
  using: "mise"
  ensure:
    - "node@22"

mise:
  tools:
    node: "22"
```

**When** вызывается `razdfile.Reader.Read(ctx, node)`

**Then** возвращается ошибка валидации:
- Сообщение содержит "cannot use 'dependencies' together with 'mise'"

#### Scenario: Dependencies with devbox section

**Given** файл `Razdfile.yml` с содержимым:
```yaml
version: "1"

dependencies:
  using: "devbox"
  ensure:
    - "node@22"

devbox:
  packages:
    - nodejs@22
```

**When** вызывается `razdfile.Reader.Read(ctx, node)`

**Then** возвращается ошибка валидации:
- Сообщение содержит "cannot use 'dependencies' together with 'devbox'"

#### Scenario: Dependencies alone is valid

**Given** файл только с `dependencies` секцией (без `mise` и `devbox`)

**When** Reader парсит файл

**Then** парсинг успешен

---

### Requirement: Parse ensure into structured format

Система MUST предоставлять метод для парсинга `ensure` строк в структурированный формат.

#### Scenario: Parse ensure strings

**Given** `DependenciesConfig` с:
```go
ensure := []string{"node@22", "php@8.4", "pnpm@latest"}
```

**When** вызывается `ParseEnsure()`

**Then** возвращается `[]ParsedDependency`:
- `[0]: {Tool: "node", Version: "22", Raw: "node@22"}`
- `[1]: {Tool: "php", Version: "8.4", Raw: "php@8.4"}`
- `[2]: {Tool: "pnpm", Version: "latest", Raw: "pnpm@latest"}`

---

### Requirement: HasDependencies helper

Система MUST предоставлять метод `HasDependencies()` для проверки наличия секции.

#### Scenario: Razdfile with dependencies

**Given** `Razdfile` с заполненной секцией `Dependencies`

**When** вызывается `rf.HasDependencies()`

**Then** возвращается `true`

#### Scenario: Razdfile without dependencies

**Given** `Razdfile` где `Dependencies == nil`

**When** вызывается `rf.HasDependencies()`

**Then** возвращается `false`

---

### Requirement: Parse extra section as pass-through

Система MUST парсить секцию `extra` как pass-through конфигурацию без валидации внутренней структуры.

#### Scenario: Parse extra.devbox with shell configuration

**Given** файл `Razdfile.yml` с содержимым:
```yaml
version: "1"

dependencies:
  using: "devbox"
  ensure:
    - "go@1.21"
  extra:
    devbox:
      shell:
        init_hook: |
          echo "Welcome!"
          export GOPATH=$PWD/.go
        scripts:
          build: "go build ."
      env:
        MY_VAR: "production"
```

**When** вызывается `razdfile.Reader.Read(ctx, node)`

**Then** возвращается `*ast.Razdfile` где:
- `Dependencies.Extra.Devbox["shell"]` содержит map с `init_hook` и `scripts`
- `Dependencies.Extra.Devbox["env"]` содержит map с `MY_VAR`
- Структура сохраняется как есть для последующего merge

#### Scenario: Parse extra.mise with env and settings

**Given** файл `Razdfile.yml` с содержимым:
```yaml
version: "1"

dependencies:
  using: "mise"
  ensure:
    - "node@22"
  extra:
    mise:
      env:
        NODE_ENV: development
      settings:
        experimental: true
```

**When** вызывается `razdfile.Reader.Read(ctx, node)`

**Then** возвращается `*ast.Razdfile` где:
- `Dependencies.Extra.Mise["env"]` содержит map с `NODE_ENV`
- `Dependencies.Extra.Mise["settings"]` содержит map с `experimental`

#### Scenario: Extra section is optional

**Given** файл без `extra` секции:
```yaml
version: "1"

dependencies:
  using: "mise"
  ensure:
    - "node@22"
```

**When** вызывается `razdfile.Reader.Read(ctx, node)`

**Then** возвращается `*ast.Razdfile` где:
- `Dependencies.Extra == nil`
- Парсинг успешен

#### Scenario: Extra with additional packages for devbox

**Given** файл с дополнительными пакетами в extra:
```yaml
version: "1"

dependencies:
  using: "devbox"
  ensure:
    - "node@22"
  extra:
    devbox:
      packages:
        - "php84Extensions.xdebug@latest"
        - "github:numtide/llm-agents.nix#openspec"
```

**When** вызывается `razdfile.Reader.Read(ctx, node)`

**Then** возвращается `*ast.Razdfile` где:
- `Dependencies.Extra.Devbox["packages"]` содержит список дополнительных пакетов
- Эти пакеты будут merged с `ensure` при генерации

## MODIFIED Requirements

### Requirement: Validate Razdfile content

Система MUST проверять что Razdfile содержит хотя бы одну секцию.

**Previous**: Razdfile должен содержать tasks, includes, mise, или devbox.

**Modified**: Razdfile должен содержать tasks, includes, mise, devbox, **или dependencies**.

#### Scenario: Razdfile with only dependencies

**Given** файл:
```yaml
version: "1"
dependencies:
  using: "mise"
  ensure:
    - "node@22"
```

**When** Reader парсит файл

**Then** парсинг успешен (dependencies section is sufficient)
