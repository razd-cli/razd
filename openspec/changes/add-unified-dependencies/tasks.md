# Tasks: Add Unified Dependencies

## 1. AST Types

- [x] Создать `razdfile/ast/dependencies.go` с типом `DependenciesConfig`
- [x] Создать тип `DependenciesExtra` для pass-through конфигурации
- [x] Создать тип `ParsedDependency` для распарсенных зависимостей
- [x] Добавить поле `Dependencies` в `ast.Razdfile`
- [x] Добавить методы `HasDependencies()`, `ParseEnsure()`
- [x] Unit тесты для `DependenciesConfig` и `DependenciesExtra`

## 2. Validation

- [x] Добавить правило mutual exclusion (dependencies vs mise/devbox)
- [x] Добавить валидацию обязательного поля `using`
- [x] Добавить валидацию допустимых значений `using` (mise, devbox)
- [x] Добавить валидацию формата `ensure` строк (tool@version)
- [x] НЕ валидировать структуру `extra` — pass-through
- [x] Unit тесты для валидации

## 3. Parsing

- [x] Добавить парсинг `dependencies` секции в Reader
- [x] Добавить парсинг `extra` как `map[string]any`
- [x] Integration тесты с примерами Razdfile
- [x] Обновить примеры в `examples/`

## 4. Provisioner Architecture

- [x] Создать `provisioner/provisioner.go` с интерфейсом Provisioner
- [x] Создать `provisioner/mise.go` с MiseProvisioner
- [x] Создать `provisioner/devbox.go` с DevboxProvisioner
- [x] Создать `provisioner/registry.go` с Registry
- [x] Добавить методы: Name(), GenerateConfig(), Install(), RunCommand(), Shell(), Trust(), Untrust()
- [x] Unit тесты для Provisioner Registry и implementations

## 5. Documentation

- [ ] Обновить JSON schema если есть
- [x] Добавить пример использования в examples/

---

## Dependencies

- Task 1 (AST) → Task 2 (Validation) → Task 3 (Parsing)
- Task 4 (Documentation) можно делать параллельно после Task 1

## Parallelizable

- Tasks 1.1-1.5 можно делать последовательно в одном файле
- Task 4 параллельно с остальными после определения AST
