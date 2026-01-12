# Razdfile Parsing Capability

Парсинг и валидация Razdfile.yml — объединённого формата конфигурации.

## ADDED Requirements

### Requirement: Parse Razdfile AST

Система MUST парсить Razdfile.yml в типизированную AST структуру.

#### Scenario: Parse minimal Razdfile with tasks

**Given** файл `Razdfile.yml` с содержимым:
```yaml
version: "1"

tasks:
  hello:
    cmds:
      - echo "Hello"
```

**When** вызывается `razdfile.Reader.Read(ctx, node)`

**Then** возвращается `*ast.Razdfile` где:
- `Version == "1"`
- `Tasks` содержит task "hello"
- `Mise == nil`

#### Scenario: Parse Razdfile with mise configuration

**Given** файл `Razdfile.yml` с содержимым:
```yaml
version: "1"

mise:
  tools:
    node: "22"
    python: "3.11"

tasks:
  install:
    cmds:
      - npm install
```

**When** вызывается `razdfile.Reader.Read(ctx, node)`

**Then** возвращается `*ast.Razdfile` где:
- `Mise.Tools["node"].Version == "22"`
- `Mise.Tools["python"].Version == "3.11"`
- `Tasks` содержит task "install"

#### Scenario: Parse complex mise tool configuration

**Given** файл `Razdfile.yml` с содержимым:
```yaml
version: "1"

mise:
  tools:
    node:
      version: "22"
      os: [linux, darwin]
      postinstall: "npm install -g pnpm"
```

**When** вызывается `razdfile.Reader.Read(ctx, node)`

**Then** `Mise.Tools["node"]` содержит:
- `Version == "22"`
- `OS == ["linux", "darwin"]`
- `Postinstall == "npm install -g pnpm"`

---

### Requirement: Validate Razdfile version

Система MUST проверять что версия Razdfile поддерживается.

#### Scenario: Valid version "1"

**Given** файл с `version: "1"`

**When** Reader парсит файл

**Then** парсинг успешен

#### Scenario: Invalid version "2"

**Given** файл с `version: "2"`

**When** Reader парсит файл

**Then** возвращается `RazdfileVersionError` с сообщением:
- "unsupported Razdfile version '2', expected '1'"

#### Scenario: Missing version

**Given** файл без поля `version`

**When** Reader парсит файл

**Then** возвращается ошибка "version is required"

---

### Requirement: Validate Razdfile content

Система MUST проверять что Razdfile содержит хотя бы одну секцию.

#### Scenario: Empty Razdfile

**Given** файл с только `version: "1"` без tasks/includes/mise

**When** Reader парсит файл

**Then** возвращается `EmptyRazdfileError`

#### Scenario: Razdfile with only mise

**Given** файл:
```yaml
version: "1"
mise:
  tools:
    node: "22"
```

**When** Reader парсит файл

**Then** парсинг успешен (mise section is sufficient)

---

### Requirement: Report parsing errors with location

Система MUST сообщать об ошибках с указанием строки и колонки.

#### Scenario: YAML syntax error

**Given** файл с невалидным YAML:
```yaml
version: "1"
tasks:
  hello:
    cmds:
      - echo "test
```

**When** Reader парсит файл

**Then** возвращается `RazdfileDecodeError` с:
- `Line` указывающий на проблемную строку
- `Column` указывающий на проблемный символ
- Понятное сообщение об ошибке

#### Scenario: Invalid task structure

**Given** файл:
```yaml
version: "1"
tasks:
  hello: 123  # invalid, should be string, array, or object
```

**When** Reader парсит файл

**Then** возвращается `RazdfileDecodeError` с line/column информацией

---

### Requirement: Locate Razdfile automatically

Система MUST автоматически находить Razdfile в директории.

#### Scenario: Find Razdfile.yml in current directory

**Given** директория содержит `Razdfile.yml`

**When** вызывается `NewFileNode("", dir)`

**Then** node указывает на `Razdfile.yml`

#### Scenario: Find Razdfile.yaml (alternate extension)

**Given** директория содержит только `Razdfile.yaml` (не .yml)

**When** вызывается `NewFileNode("", dir)`

**Then** node указывает на `Razdfile.yaml`

#### Scenario: Razdfile not found

**Given** директория без Razdfile

**When** вызывается `NewFileNode("", dir)`

**Then** возвращается `RazdfileNotFoundError`

---

### Requirement: Support mise env section

Система MUST парсить mise.env секцию.

#### Scenario: Parse simple env variables

**Given** файл:
```yaml
version: "1"
mise:
  env:
    NODE_ENV: development
    DEBUG: "true"
```

**When** Reader парсит файл

**Then** `Mise.Env` содержит:
- `NODE_ENV == "development"`
- `DEBUG == "true"`

#### Scenario: Parse env with special directives

**Given** файл:
```yaml
version: "1"
mise:
  env:
    _:
      file: .env.local
      path:
        - ./node_modules/.bin
```

**When** Reader парсит файл

**Then** `Mise.Env["_"]` содержит special directives

---

### Requirement: Support mise settings

Система MUST парсить mise.settings секцию.

#### Scenario: Parse settings

**Given** файл:
```yaml
version: "1"
mise:
  settings:
    auto_install: true
    experimental: false
    jobs: 4
```

**When** Reader парсит файл

**Then** `Mise.Settings` содержит:
- `AutoInstall == true`
- `Experimental == false`
- `Jobs == 4`
