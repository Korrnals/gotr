# Stage 9.0 — Стандарты проекта

> Разработка и фиксация общих стандартов для проекта gotr.
> Этот этап определяет правила, по которым будет проводиться аудит в Stage 10.0.

---

## Цель

Создать полный набор стандартов, покрывающих:
- Архитектуру и слои проекта
- Правила кодирования и именования
- Пакетную связность и зависимости
- Соответствие TestRail API v2
- Шаблоны для нового кода
- CI/Lint конфигурацию

---

## Подзадачи

### 9.1 — Обновить `docs/architecture.md`

**Что сделать:**
- Актуализировать схему слоёв (добавить `internal/flags/`, `internal/output/`)
- Обновить список endpoints (106 → проверить актуальность)
- Добавить раздел «Что запрещено» (нарушения слоёв)
- Обновить пример команды на новый стандарт (с `flags.*`, `ui.*`, `output.*`)

**Артефакт:** Обновлённый `docs/architecture.md`

---

### 9.2 — `docs/standards.md` ✅ СОЗДАН

Файл уже создан со следующими разделами:
1. Общие принципы (SRP, явность, ошибки-значения, интерфейсы, DRY, YAGNI)
2. Архитектура слоёв и правила зависимостей
3. Структура пакетов (cmd/, client/, service/, ui/, flags/, output/)
4. Правила кодирования (именование, ошибки, context, конкурентность, тесты, язык)
5. Паттерны проекта (Constructor Injection, Generic Factory, Functional Options, Builder, ISP)
6. Контрольный список (checklist) для нового кода

---

### 9.3 — Анализ пакетной связности

**Что сделать:**
```bash
# Проверить циклические зависимости
go vet ./...

# Построить граф импортов
go list -f '{{.ImportPath}}: {{join .Imports ", "}}' ./internal/... ./cmd/... ./pkg/...

# Найти нарушения: internal/ → cmd/
grep -rn 'github.com/Korrnals/gotr/cmd' internal/ --include='*.go' | grep -v '_test.go'

# Найти нарушения: pkg/ → internal/
grep -rn 'github.com/Korrnals/gotr/internal' pkg/ --include='*.go' | grep -v '_test.go'
```

**Проверить:**
- [ ] Нет `internal/` → `cmd/` импортов
- [ ] Нет `pkg/` → `internal/` импортов
- [ ] `internal/ui/` не импортирует `internal/client/`
- [ ] `internal/service/` не импортирует `cmd/`
- [ ] Нет циклических зависимостей

**Артефакт:** Отчёт с найденными нарушениями + план их устранения

---

### 9.4 — API Contract

**Что сделать:**
- Открыть [TestRail API v2 Reference](https://support.testrail.com/hc/en-us/categories/7077196481428)
- Сверить КАЖДЫЙ метод в `internal/client/interfaces.go` с документацией:
  - Правильный HTTP-метод (GET/POST)
  - Правильный URL-шаблон (`/api/v2/get_case/{case_id}`)
  - Правильные параметры запроса
  - Правильная модель ответа
- Составить таблицу: метод | URL | статус (✅ соответствует / ⚠️ расхождение / ❌ отсутствует)

**Артефакт:** `docs/api-contract.md` — таблица соответствия

---

### 9.5 — Шаблоны для нового кода

**Что сделать:**
Создать `docs/templates/` с файлами:
- `new-resource-cmd.go.tmpl` — шаблон новой CRUD-команды
- `new-client-method.go.tmpl` — шаблон нового API-метода
- `new-test.go.tmpl` — шаблон тестового файла
- `new-compare-cmd.go.tmpl` — шаблон compare-подкоманды

**Артефакт:** `docs/templates/`

---

### 9.6 — CI/Lint конфигурация

**Что сделать:**
Создать `.golangci.yml` с правилами:

```yaml
linters:
  enable:
    - errcheck          # Проверка обработки ошибок
    - govet             # go vet
    - staticcheck       # Расширенный анализ
    - unused            # Неиспользуемый код
    - gosimple          # Упрощение кода
    - ineffassign       # Неиспользуемые присваивания
    - misspell          # Опечатки
    - gocyclo           # Цикломатическая сложность
    - gocritic          # Расширенные проверки
    - revive            # Линтер стиля
    - nolintlint        # Правильные nolint-комментарии

linters-settings:
  gocyclo:
    min-complexity: 15
  gocritic:
    enabled-tags:
      - diagnostic
      - style
      - performance

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gocyclo
        - gocritic
```

**Артефакт:** `.golangci.yml`

---

## Критерии готовности Stage 9.0

- [ ] `docs/standards.md` — полный и актуальный
- [ ] `docs/architecture.md` — обновлён до текущей реальности
- [ ] Анализ пакетной связности — проведён, нарушения зафиксированы
- [ ] API Contract — сверен с документацией TestRail
- [ ] Шаблоны созданы в `docs/templates/`
- [ ] `.golangci.yml` создан и проходит без critical-ошибок
