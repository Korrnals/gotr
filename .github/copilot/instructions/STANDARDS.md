# Стандарты проекта gotr

> Стандарты кодирования, архитектуры и организации кода для CLI-утилиты gotr.
> Обновлено: 2026-03-12 (v3.0.0-dev) — Stage 8.0 complete.

---

## 1. Общие принципы

| Принцип | Описание |
| ------- | -------- |
| **Единая ответственность** | Каждый пакет / файл / функция отвечают ровно за одну задачу. |
| **Явность > Магия** | Нет глобальных state, скрытых init(), implicit-зависимостей. Зависимости передаются через аргументы или опции. |
| **Ошибки — значения** | Никаких `os.Exit` / `log.Fatal` в библиотечном коде. Паника допустима только при инициализации клиента (`GetClient*`). |
| **Интерфейсы — маленькие** | Интерфейс определяется в пакете-потребителе, а не в пакете-поставщике. |
| **Тестируемость** | Любой слой может быть подменён моком через интерфейс. Никаких внешних вызовов в unit-тестах. |
| **DRY (Don't Repeat Yourself)** | Дублирование логики — баг. Generic-фабрики (`newSimpleCompareCmd`), хелперы (`ui.*`, `flags.*`) используются вместо копипасты. |
| **YAGNI (You Ain't Gonna Need It)** | Не добавляем абстракции «про запас». Рефакторим, когда появляется третий потребитель. |

---

## 2. Архитектура слоёв

```
CLI (cmd/*)  →  Service (internal/service/)  →  Client (internal/client/)  →  TestRail API
      ↕                     ↕                          ↕
   UI Layer            Concurrency              Rate Limiter
(internal/ui/)    (internal/concurrency/)    (internal/concurrent/)
```

### 2.1. Правила зависимостей

| Слой | Может зависеть от | НЕ может зависеть от |
| ---- | ----------------- | -------------------- |
| `cmd/*` | `internal/service`, `internal/client`, `internal/ui`, `internal/flags`, `internal/interactive`, `internal/output`, `pkg/*` | — |
| `internal/service` | `internal/client`, `internal/concurrency`, `internal/concurrent`, `internal/models` | `cmd/*`, `internal/ui` |
| `internal/client` | `internal/concurrent` (rate limiter, retry), `internal/models/data` | `cmd/*`, `internal/service` |
| `internal/ui` | стандартная библиотека, `go-pretty/v6` | `internal/client`, `internal/service` |
| `internal/concurrency` | стандартная библиотека | `internal/client`, `cmd/*` |
| `pkg/*` | стандартная библиотека, `go-pretty/v6` | `internal/*`, `cmd/*` |

### 2.2.  Направление вызовов

```
cmd/ → service/ → client/ → HTTP
cmd/ → ui.*  (вывод)
cmd/ → flags.*  (валидация входных данных)
cmd/ → interactive.*  (интерактивный ввод)
cmd/ → output.*  (сохранение в файл)
```

Запрещено:
- `service/` → `cmd/` (обращение вверх)
- `client/` → `ui/` (клиент не должен знать о UI)
- `pkg/` → `internal/` (публичный пакет не импортирует приватный)

---

## 3. Структура пакетов

### 3.1. `cmd/` — CLI-команды

```
cmd/
├── root.go           # Корневая команда, initConfig()
├── commands.go       # Регистрация всех подкоманд
├── add.go            # gotr add <resource>
├── update.go         # gotr update <resource>
├── delete.go         # gotr delete <resource>
├── list.go           # gotr list <resource>
├── export.go         # gotr export
├── config.go         # gotr config {init|path|view|edit}
├── resources.go      # gotr resources (API endpoints)
├── selftest.go       # gotr selftest
├── completion.go     # gotr completion {bash|zsh|fish}
├── <resource>/       # Подкоманды для ресурса
│   ├── <resource>.go # Регистрация + getClient()
│   ├── add.go        # Создание
│   ├── get.go        # Получение одного
│   ├── list.go       # Получение списка
│   ├── update.go     # Обновление
│   ├── delete.go     # Удаление
│   └── *_test.go     # Тесты
├── compare/          # gotr compare <resource> --pid1 X --pid2 Y
├── sync/             # gotr sync {cases|sections|suites|...}
└── internal/         # Общие тестовые хелперы (testhelper)
```

**Паттерн команды:**

```go
func newXxxCmd(clientFn func(*cobra.Command) client.ClientInterface) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "xxx",
        Short: "Краткое описание",
        Args:  cobra.ExactArgs(N),
        RunE: func(cmd *cobra.Command, args []string) error {
            cli := clientFn(cmd)
            ctx := cmd.Context()

            // 1. Валидация через flags.*
            id, err := flags.ValidateRequiredID(args, 0, "resource_id")
            if err != nil { return err }

            // 2. Бизнес-логика через client/service
            result, err := cli.GetXxx(ctx, id)
            if err != nil { return err }

            // 3. Вывод через output/ui
            return output.OutputResult(cmd, result, "xxx")
        },
    }
    output.AddFlag(cmd)
    return cmd
}
```

**Правила:**
- Конструктор `newXxxCmd(clientFn)` — для тестируемости (mock client injection).
- `RunE` (не `Run`) — ошибки возвращаются, а не печатаются.
- `cmd.Context()` — пробрасывается во все нижние слои.
- Флаги регистрируются в конструкторе, а не в `init()`.

### 3.2. `internal/client/` — HTTP-клиент TestRail

```
internal/client/
├── client.go         # HTTPClient — конструктор, базовые методы
├── interfaces.go     # ClientInterface (106 endpoints, 14 интерфейсов)
├── mock.go           # MockClient для тестов
├── paginator.go      # fetchAllPages[T] — generic-пагинатор
├── request.go        # sendRequest(), debug вывод
├── accessor.go       # ClientAccessor — lazy init
├── concurrent.go     # Thread-safe обёртки
├── <domain>.go       # Endpoints по доменам (cases, runs, plans...)
└── *_test.go         # Тесты
```

**Правила:**
- Каждый метод принимает `ctx context.Context` первым аргументом.
- List-методы используют `fetchAllPages[T]` для автопагинации.
- Новые endpoints добавляются в `interfaces.go` + `mock.go` + доменный файл.
- `MockClient` реализует `ClientInterface` полностью — для тестов.

### 3.3. `internal/service/` — Бизнес-логика

```
internal/service/
├── run.go            # RunService — управление test runs
├── result.go         # ResultService — результаты тестов
├── test.go           # TestService — тестовые данные
├── migration/        # Миграция между проектами
│   ├── types.go      # Migration struct, конструкторы
│   ├── fetch.go      # Загрузка данных
│   ├── filter.go     # Фильтрация дубликатов
│   ├── import.go     # Импорт сущностей
│   ├── export.go     # Экспорт данных
│   ├── mapping.go    # Управление маппингом ID
│   ├── log.go        # Логирование операций
│   └── migrate.go    # Оркестрация миграции
└── *_test.go
```

**Правила:**
- Сервис получает `client.ClientInterface` через конструктор.
- Сервис НЕ знает о `cobra.Command`, флагах, stdin/stdout.
- Валидация бизнес-правил — здесь (не в cmd/).
- Операции с эффектами (создание/удаление) — через сервис.

### 3.4. `internal/ui/` — Унифицированный вывод

**Статические хелперы (display.go):**
```go
ui.Info(w, msg)              // ℹ️  msg
ui.Infof(w, fmt, args...)    // ℹ️  formatted msg
ui.Success(w, msg)           // ✅ msg
ui.Successf(w, fmt, args...) // ✅ formatted msg
ui.Warning(w, msg)           // ⚠️  msg
ui.Warningf(w, fmt, args...) // ⚠️  formatted msg
ui.Error(w, msg)             // ❌ msg
ui.Phase(w, msg)             // 🔄 msg
ui.Stat(w, icon, label, val) //    📊 label: val
ui.Section(w, msg)           //    📊 msg (header)
ui.Cancelled(w)              //    ❌ Cancelled
ui.Preview(w, title, fields) //    Bordered preview box
```

**Таблицы (table.go):**
```go
ui.NewTable(cmd)        // go-pretty table с учётом --format
ui.JSON(cmd, data)      // JSON-вывод с учётом --quiet
ui.IsJSON(cmd) bool     // Проверка формата
ui.IsQuiet(cmd) bool    // Проверка quiet-режима
```

**Правила:**
- Весь пользовательский вывод — через `ui.*` (кроме интерактивных промптов и отладки).
- Первый аргумент — `io.Writer` (обычно `os.Stdout`).
- Emoji-префиксы только в `ui.*` — никаких хардкоженных emoji в cmd/.

### 3.5. `internal/flags/` — Валидация входных данных

```go
flags.ValidateRequiredID(args, index, name)   // Парсинг ID из аргументов
flags.GetFlag[T](cmd, name)                   // Типобезопасное получение флага
flags.GetOptionalFlag[T](cmd, name)           // Опциональный флаг
```

**Правила:**
- Валидация аргументов/флагов выполняется через `flags.*`, а не вручную.
- Ошибки валидации возвращаются (не паника, не os.Exit).

### 3.6. `internal/output/` — Сохранение результатов

```go
output.OutputResult(cmd, data, resource)  // Вывод + сохранение в файл
output.Output(cmd, data, dir, format)     // Сохранение в ~/.gotr/exports/
output.AddFlag(cmd)                       // Регистрация --save, --save-to
output.DryRunPrinter(cmd)                 // Вывод для dry-run
```

---

## 4. Правила кодирования

### 4.1. Именование

| Элемент | Стиль | Пример |
| ------- | ----- | ------ |
| Пакеты | lowercase, одно слово | `client`, `flags`, `ui` |
| Экспортируемые функции | CamelCase | `GetCases`, `ValidateRequiredID` |
| Приватные функции | camelCase | `fetchAllPages`, `parseResponse` |
| Константы | CamelCase или UPPER_SNAKE | `DefaultPageSize`, `MaxRetries` |
| Файлы | snake_case | `sync_cases.go`, `add_config.go` |
| Тестовые файлы | `<file>_test.go` | `add_test.go`, `cases_test.go` |
| Тестовые хелперы | `test_helper.go` | `cmd/cases/test_helper.go` |

### 4.2. Обработка ошибок

```go
// ✅ Правильно — оборачиваем контекстом
if err != nil {
    return fmt.Errorf("failed to get cases for project %d: %w", projectID, err)
}

// ✅ Правильно — sentinel errors для предсказуемых ситуаций
var ErrNotFound = errors.New("resource not found")

// ❌ Неправильно — голый return
if err != nil { return err }

// ❌ Неправильно — os.Exit / log.Fatal
if err != nil { os.Exit(1) }

// ❌ Неправильно — fmt.Printf ошибки
if err != nil { fmt.Printf("error: %v\n", err) }
```

### 4.3. Context

- Все функции, делающие I/O (HTTP, файлы, ожидание), принимают `ctx context.Context` первым аргументом.
- `cmd.Context()` — источник контекста в CLI-слое.
- Отмена по Ctrl+C — через `signal.NotifyContext` (Stage 7.0).

### 4.4. Конкурентность

- Worker pool: `internal/concurrent/pool.go` — ограниченный пул горутин.
- Rate limiter: `internal/concurrent/limiter.go` — token bucket (150 req/min к TestRail).
- Retry: `internal/concurrent/retry.go` — экспоненциальная задержка.
- Стратегии параллелизации: `internal/concurrency/` — generic-стратегии `FetchParallel[T]`, `FetchParallelBySuite[T]`, `ParallelController`.

### 4.5. Тестирование

| Требование | Описание |
| ---------- | -------- |
| **Покрытие** | ≥ 85% для каждого пакета (цель — 90%+). |
| **Паттерн** | Table-driven tests (`[]struct{ name string; ... }`). |
| **Моки** | `client.MockClient` — единственный способ мокать HTTP. |
| **Инъекция** | Конструктор `newXxxCmd(clientFn)` — для DI клиента в тестах. |
| **Хелперы** | `cmd/internal/testhelper/` — общие утилиты для тестов команд. |
| **Naming** | `TestXxx_Success`, `TestXxx_Error`, `TestXxx_InvalidArgs`. |
| **Assertions** | Стандартная библиотека или `testify/assert`. |
| **Нет внешних вызовов** | Unit-тесты никогда не ходят в сеть. |

### 4.6. Язык

- **Исходный код**: комментарии, переменные, функции — **английский**.
- **Пользовательский вывод (UI)**: **английский** (мигрировано в Stage 8.0).
- **Документация** (`docs/`, `README.md`): **русский** (основной язык проекта).
- **Ошибки** (`fmt.Errorf`): **английский** (для grep-able логов).

---

## 5. Паттерны проекта

### 5.1. Constructor Injection

```go
// cmd/cases/cases.go
func Register(rootCmd *cobra.Command, clientFn GetClientFunc) {
    casesCmd := newCasesCmd(clientFn)
    rootCmd.AddCommand(casesCmd)
}

// Тест
func TestAddCase(t *testing.T) {
    mock := &client.MockClient{...}
    cmd := newAddCaseCmd(func(_ *cobra.Command) client.ClientInterface { return mock })
    cmd.SetArgs([]string{"123", "--title", "Test"})
    err := cmd.Execute()
    assert.NoError(t, err)
}
```

### 5.2. Generic Factory (DRY)

```go
// cmd/compare/simple.go — одна фабрика для 9 подкоманд
func newSimpleCompareCmd[T any](cfg simpleCompareCfg[T]) *cobra.Command { ... }
```

### 5.3. Functional Options

```go
// internal/concurrency/
result, err := concurrency.FetchParallel(ctx, fetchFn,
    concurrency.WithReporter(reporter),
    concurrency.WithMaxConcurrency(5),
    concurrency.WithContinueOnError(true),
)
```

### 5.4. Builder Pattern (Reporter)

```go
reporter.New("Compare Results").
    Section("Projects").
    Stat("📁", "Project 1", p1Name).
    Stat("📁", "Project 2", p2Name).
    Section("Statistics").
    Stat("📊", "Common", len(common)).
    Print()
```

### 5.5. Interface Segregation

```go
// internal/client/interfaces.go — 14 маленьких интерфейсов
type CasesAPI interface { ... }
type RunsAPI interface { ... }
type ClientInterface interface {
    CasesAPI
    RunsAPI
    // ...
}
```

---

## 6. Контрольный список (Checklist) для нового кода

- [ ] Функция/метод принимает `ctx context.Context` (если делает I/O)
- [ ] Ошибки обёрнуты контекстом: `fmt.Errorf("... : %w", err)`
- [ ] Вывод пользователю — через `ui.*` (не `fmt.Printf`)
- [ ] Валидация — через `flags.*` (не ручной парсинг)
- [ ] Сохранение — через `output.*` (не ручной `os.WriteFile`)
- [ ] Конструктор команды: `newXxxCmd(clientFn)` (инъекция клиента)
- [ ] Тест написан: ≥ success + error + edge case
- [ ] Нет дублирования: если логика повторяется ≥ 3 раз — вынести в хелпер
- [ ] Нет кириллицы в коде (кроме `docs/` и `README*.md`)
- [ ] `goimports -w` применён
- [ ] `go vet ./...` без предупреждений
- [ ] `go build ./...` без ошибок
