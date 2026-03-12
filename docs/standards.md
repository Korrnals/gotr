# Стандарты кодирования gotr

> Стандарты разработки CLI-утилиты gotr.
> Обновлено: 2026-03-12 (Stage 9.0).

---

## 1. Общие принципы

| Принцип | Описание |
| ------- | -------- |
| **Единая ответственность** | Каждый пакет/файл/функция — одна задача |
| **Явность > Магия** | Нет глобального state, нет скрытых `init()`, DI через аргументы |
| **Ошибки — значения** | Нет `os.Exit`/`log.Fatal` в библиотечном коде. Паника только в `GetClient*` |
| **Интерфейсы — маленькие** | Interface Segregation: интерфейс в пакете-потребителе |
| **Тестируемость** | Моки через интерфейс. Нет внешних вызовов в unit-тестах |
| **DRY** | Дублирование — баг. Generic-фабрики, хелперы `ui.*`, `flags.*` |
| **YAGNI** | Нет абстракций «про запас». Рефакторим при третьем потребителе |

---

## 2. Архитектура слоёв

```
cmd/ → service/ → client/ → HTTP (TestRail API)
cmd/ → ui.*      (вывод)
cmd/ → flags.*   (валидация)
cmd/ → output.*  (сохранение)
```

### 2.1. Правила зависимостей

| Слой | Может зависеть от | НЕ может зависеть от |
| ---- | ----------------- | -------------------- |
| `cmd/*` | `internal/*`, `pkg/*` | — |
| `internal/service` | `internal/client`, `internal/concurrency`, `internal/models` | `cmd/*`, `internal/ui` |
| `internal/client` | `internal/concurrent`, `internal/models/data` | `cmd/*`, `internal/service` |
| `internal/ui` | stdlib, `go-pretty/v6` | `internal/client`, `internal/service` |
| `pkg/*` | stdlib, `go-pretty/v6` | `internal/*`, `cmd/*` |

**Запрещено:**
- `service/` → `cmd/` (обращение вверх по слоям)
- `client/` → `ui/` (клиент не знает о UI)
- `pkg/` → `internal/` (публичный API не импортирует приватный)

---

## 3. Структура пакетов

### 3.1. `cmd/` — CLI-команды

- Конструктор: `newXxxCmd(clientFn func(*cobra.Command) client.ClientInterface)`
- Регистрация: `Register(rootCmd, clientFn)` — вызывается из `commands.go`
- Используем `RunE` (не `Run`) — ошибки возвращаются Cobra
- Контекст: `ctx := cmd.Context()` — пробрасывается во все вызовы

### 3.2. `internal/client/` — HTTP-клиент

- Каждый метод: `func (c *HTTPClient) GetXxx(ctx context.Context, ...)` 
- List-методы: `fetchAllPages[T]` для автопагинации
- Новый endpoint: добавить в `interfaces.go` + `mock.go` + доменный файл

### 3.3. `internal/ui/` — Вывод

- `ui.Table(cmd, t)` — таблица с учётом `--format`
- `ui.JSON(cmd, data)` — JSON-вывод
- `ui.Info(w, msg)`, `ui.Success(w, msg)` и т.д. — стилизованные сообщения
- Emoji-префиксы только в `ui.*`

### 3.4. `internal/flags/` — Валидация

- `flags.ValidateRequiredID(args, index, name)` — парсинг ID
- `flags.GetFlagInt64(cmd, name)` — типобезопасный флаг

### 3.5. `internal/output/` — Сохранение

- `output.AddFlag(cmd)` — регистрация `--save`, `--save-to`
- `output.OutputResult(cmd, data, resource)` — вывод + сохранение

---

## 4. Правила кодирования

### 4.1. Именование

| Элемент | Стиль | Пример |
| ------- | ----- | ------ |
| Пакеты | lowercase, одно слово | `client`, `flags`, `ui` |
| Экспортируемые | CamelCase | `GetCases`, `ValidateRequiredID` |
| Приватные | camelCase | `fetchAllPages`, `parseResponse` |
| Файлы | snake_case | `sync_cases.go`, `add_config.go` |
| Тесты | `<file>_test.go` | `add_test.go` |

### 4.2. Обработка ошибок

```go
// ✅ Правильно — оборачиваем контекстом
if err != nil {
    return fmt.Errorf("failed to get cases for project %d: %w", projectID, err)
}

// ❌ Голый return
if err != nil { return err }

// ❌ os.Exit / log.Fatal
if err != nil { os.Exit(1) }
```

### 4.3. Context

- Все I/O функции принимают `ctx context.Context` первым аргументом
- `cmd.Context()` — источник контекста в CLI-слое
- Ctrl+C работает через `signal.NotifyContext` (Stage 7.0)

### 4.4. Тестирование

| Требование | Описание |
| ---------- | -------- |
| Покрытие | ≥ 85% для каждого пакета (цель 90%+) |
| Паттерн | Table-driven tests |
| Моки | `client.MockClient` |
| DI | `newXxxCmd(clientFn)` для инъекции |
| Naming | `TestXxx_Success`, `TestXxx_Error` |
| Без сети | Unit-тесты без внешних вызовов |

### 4.5. Язык

| Контекст | Язык |
| -------- | ---- |
| Исходный код (переменные, функции) | Английский |
| Пользовательский вывод (UI) | Английский |
| Ошибки (`fmt.Errorf`) | Английский |
| Документация (`docs/`, `README*.md`) | Русский |
| Комментарии в коде | Английский |

---

## 5. Паттерны проекта

### 5.1. Constructor Injection

```go
func Register(rootCmd *cobra.Command, clientFn GetClientFunc) {
    casesCmd := newCasesCmd(clientFn)
    rootCmd.AddCommand(casesCmd)
}
```

### 5.2. Generic Factory (DRY)

```go
func newSimpleCompareCmd[T any](cfg simpleCompareCfg[T]) *cobra.Command { ... }
```

### 5.3. Functional Options

```go
result, err := concurrency.FetchParallel(ctx, fetchFn,
    concurrency.WithReporter(reporter),
    concurrency.WithMaxConcurrency(5),
)
```

### 5.4. Builder Pattern (Reporter)

```go
reporter.New("Compare Results").
    Section("Projects").
    Stat("📁", "Project 1", p1Name).
    Print()
```

---

## 6. Контрольный список (Checklist)

- [ ] `ctx context.Context` передаётся во все I/O вызовы
- [ ] Ошибки обёрнуты: `fmt.Errorf("... : %w", err)`
- [ ] Вывод — через `ui.*` (не `fmt.Printf`)
- [ ] Валидация — через `flags.*`
- [ ] Сохранение — через `output.*`
- [ ] Конструктор: `newXxxCmd(clientFn)`
- [ ] Тест: ≥ success + error + edge case
- [ ] Нет дублирования (≥ 3 раз → хелпер)
- [ ] Нет кириллицы в коде
- [ ] `go vet ./...` — 0 предупреждений
- [ ] `go build ./...` — 0 ошибок
