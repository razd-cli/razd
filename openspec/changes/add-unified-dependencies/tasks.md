# Tasks: Add Unified Dependencies

## 1. AST Types

- [ ] Создать `razdfile/ast/dependencies.go` с типом `DependenciesConfig`
- [ ] Создать тип `DependenciesExtra` для pass-through конфигурации
- [ ] Создать тип `ParsedDependency` для распарсенных зависимостей
- [ ] Добавить поле `Dependencies` в `ast.Razdfile`
- [ ] Добавить методы `HasDependencies()`, `ParseEnsure()`
- [ ] Unit тесты для `DependenciesConfig` и `DependenciesExtra`

## 2. Validation

- [ ] Добавить правило mutual exclusion (dependencies vs mise/devbox)
- [ ] Добавить валидацию обязательного поля `using`
- [ ] Добавить валидацию допустимых значений `using` (mise, devbox)
- [ ] Добавить валидацию формата `ensure` строк (tool@version)
- [ ] НЕ валидировать структуру `extra` — pass-through
- [ ] Unit тесты для валидации

## 3. Parsing

- [ ] Добавить парсинг `dependencies` секции в Reader
- [ ] Добавить парсинг `extra` как `map[string]any`
- [ ] Integration тесты с примерами Razdfile
- [ ] Обновить примеры в `examples/` (опционально)

## 4. Documentation

- [ ] Обновить JSON schema если есть
- [ ] Добавить пример использования в README или examples/

---

## Dependencies

- Task 1 (AST) → Task 2 (Validation) → Task 3 (Parsing)
- Task 4 (Documentation) можно делать параллельно после Task 1

## Parallelizable

- Tasks 1.1-1.5 можно делать последовательно в одном файле
- Task 4 параллельно с остальными после определения AST
