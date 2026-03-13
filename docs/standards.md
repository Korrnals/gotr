# Стандарты кодирования gotr

> Полные стандарты разработки CLI-утилиты gotr.
> Обновлено: 2026-03-12 (Stage 9.0, rev.3).

---

## Содержание

1. [Общие принципы](#1-общие-принципы)
2. [Архитектура слоёв](#2-архитектура-слоёв)
3. [Структура пакетов](#3-структура-пакетов)
4. [Правила кодирования](#4-правила-кодирования)
5. [Паттерны проекта](#5-паттерны-проекта)
6. [Конфигурация и окружение](#6-конфигурация-и-окружение)
7. [Конкурентность и устойчивость](#7-конкурентность-и-устойчивость)
8. [Логирование](#8-логирование)
9. [Тестирование](#9-тестирование)
10. [Сборка и релиз](#10-сборка-и-релиз)
11. [Контрольный список](#11-контрольный-список-checklist)

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
cmd/ → log.*     (логирование)
```

### 2.1. Правила зависимостей

| Слой | Может зависеть от | НЕ может зависеть от |
| ---- | ----------------- | -------------------- |
| `cmd/*` | `internal/*`, `pkg/*` | — |
| `internal/service` | `internal/client`, `internal/concurrency`, `internal/concurrent`, `internal/models` | `cmd/*`, `internal/ui` |
| `internal/client` | `internal/concurrent`, `internal/models/data` | `cmd/*`, `internal/service` |
| `internal/ui` | stdlib, `go-pretty/v6` | `internal/client`, `internal/service` |
| `internal/concurrency` | stdlib | `internal/client`, `cmd/*` |
| `internal/concurrent` | `golang.org/x/time/rate`, `golang.org/x/sync/errgroup` | `cmd/*`, `internal/service` |
| `internal/log` | `go.uber.org/zap`, `internal/paths` | `cmd/*`, `internal/client` |
| `internal/paths` | stdlib | всё остальное |
| `pkg/*` | stdlib, `go-pretty/v6` | `internal/*`, `cmd/*` |

**Запрещено:**
- `service/` → `cmd/` (обращение вверх по слоям)
- `client/` → `ui/` (клиент не знает о UI)
- `pkg/` → `internal/` (публичный API не импортирует приватный)
- Циклические зависимости между пакетами

---

## 3. Структура пакетов

### 3.1. `cmd/` — CLI-команды

- Конструктор: `newXxxCmd(clientFn func(*cobra.Command) client.ClientInterface)`
- Регистрация: `Register(rootCmd, clientFn)` — вызывается из `commands.go`
- Используем `RunE` (не `Run`) — ошибки возвращаются Cobra
- Контекст: `ctx := cmd.Context()` — пробрасывается во все вызовы

**Регистрация команд:** Все подпакеты вызывают `Register(rootCmd, clientFn)` в `cmd/commands.go`:

```go
func init() {
    initConfig()
    initGlobalFlags()
    // ...
    cases.Register(rootCmd, GetClientInterface)
    compare.Register(rootCmd, GetClientInterface)
    sync.Register(rootCmd, GetClient)
}
```

### 3.2. `internal/client/` — HTTP-клиент

- Каждый метод: `func (c *HTTPClient) GetXxx(ctx context.Context, ...)`
- List-методы: `fetchAllPages[T]` для автопагинации
- 14 интерфейсов по ISP, 106 endpoints, 100% покрытие TestRail API v2
- Новый endpoint: добавить в `interfaces.go` + `mock.go` + доменный файл

**HTTP-транспорт (`authTransport`):**

`authTransport` — кастомный `http.RoundTripper`, оборачивающий стандартный транспорт.
Автоматически добавляет ко всем запросам:
- `Authorization: Basic ...` — через `req.SetBasicAuth(username, apiKey)`
- `Content-Type: application/json` — если не задан вызывающим кодом
- `User-Agent: Mozilla/5.0 (compatible; gotr/2.7; ...)` — **обязателен**, некоторые инсталляции TestRail возвращают 403/401 без браузерного заголовка

**Тюнинг транспорта:**

| Параметр | Значение | Причина |
| -------- | -------- | ------- |
| `MaxConnsPerHost` | `0` (unlimited) | Конкурентность управляется `concurrent.WorkerPool`, не транспортом |
| `MaxIdleConns` | `200` | Пул для повторного использования соединений |
| `MaxIdleConnsPerHost` | `200` | Совпадает с `MaxIdleConns` (один хост TestRail) |
| `IdleConnTimeout` | `90s` | Стандартное значение Go |
| `TLSHandshakeTimeout` | `10s` | Из `defaultOptions` |
| `Timeout` (Client) | `30s` | Из `defaultOptions`, переопределяется через `WithTimeout()` |

> **Важно:** `MaxConnsPerHost=0` (unlimited) — осознанное решение. При `MaxConnsPerHost=50` и 160 параллельных запросах (2 проекта × 8 сьютов × 10 страниц) 110 запросов встают в очередь внутри Go-транспорта. `http.Client.Timeout` включает время ожидания в очереди → каскадные таймауты → экспоненциальный backoff → 3× замедление.

**Пагинация (`paginator.go`):**

```go
const paginationLimit = 250  // Стандартный размер страницы TestRail API
```

- `fetchAllPages[T](ctx, client, endpoint, baseQuery, itemsField)` — загрузка всех страниц
- `decodeListResponse[T](body, itemsField)` — dual-format детекция:
  - `{` → Paginated wrapper (TestRail 6.7+): `{"offset":0, "limit":250, "<itemsField>":[...]}`
  - `[` → Flat array (старые версии TestRail Server): `[item1, item2, ...]`
- `itemsField` — имя JSON-ключа: `"runs"`, `"plans"`, `"sections"`, `"milestones"` и т.д.
- Цикл: `offset += paginationLimit` пока `pageLen >= paginationLimit`

**Functional Options:**
```go
client.NewClient(baseURL, username, apiKey, debug,
    client.WithSkipTlsVerify(true),   // --insecure
    client.WithTimeout(60*time.Second),
)
```

### 3.3. `internal/service/` — Бизнес-логика

Сервисный слой инкапсулирует бизнес-правила и валидацию:

- `RunService` — создание/обновление/закрытие test runs (валидация параметров)
- `ResultService` — работа с результатами тестов
- `TestService` — операции с тестами
- `migration/` — подсистема миграции данных между проектами TestRail:
  - `types.go` — контекст миграции (`Migration` struct)
  - `fetch.go` — загрузка source/target данных
  - `filter.go` — фильтрация дубликатов
  - `import.go` — импорт сущностей
  - `export.go` — экспорт данных и маппингов
  - `mapping.go` — управление mapping ID (source→target)
  - `mapping_loader.go` — загрузка маппингов из файлов
  - `log.go` — логирование операций миграции
  - `migrate.go` — оркестрация полной миграции

### 3.4. `internal/ui/` — Вывод

- `ui.NewTable(cmd)` — создание go-pretty таблицы с учётом `--format`
- `ui.Table(cmd, t)` — рендеринг таблицы (table/json/csv/md/html)
- `ui.JSON(cmd, data)` — JSON-вывод с учётом `--quiet`
- `ui.IsJSON(cmd)`, `ui.IsQuiet(cmd)` — проверки формата
- Стилизованные сообщения (emoji-префиксы только здесь!):
  - `ui.Info(w, msg)` — ℹ️ информация
  - `ui.Success(w, msg)` — ✅ успех
  - `ui.Warning(w, msg)` — ⚠️ предупреждение
  - `ui.Error(w, msg)` — ❌ ошибка
  - `ui.Phase(w, msg)` — 🔄 фаза
  - `ui.Stat(w, icon, label, val)` — статистика
  - `ui.Section(w, msg)` — заголовок секции
  - `ui.Preview(w, title, fields)` — окно предпросмотра
- `display.go` — ANSI live display с динамическими задачами (для `compare cases`):
  - `ui.New(opts...)` — создание Display + фоновый refresh loop (~5 Hz)
  - `d.SetHeader(text)` — заголовок над задачами
  - `d.AddTask(name, total) *Task` — трекинг задачи; `*Task` реализует `parallel.ProgressReporter`
  - `t.OnCasesReceived(n)` — обновление счётчика кейсов
  - `t.OnPageFetched()` — страница загружена
  - `t.OnSuiteComplete()` — сьют завершён
  - `d.Finish()` — остановка refresh loop, финальная отрисовка
  - Опции: `WithWriter(w)`, `WithQuiet(true)` — отключает вывод
  - Рендеринг: ANSI escape codes для in-place перезаписи строк

**Правила:**
- Весь пользовательский вывод — через `ui.*` (кроме интерактивных промптов и debug)
- Emoji-префиксы только в `ui.*` — никаких хардкоженных emoji в `cmd/`
- Первый аргумент — `io.Writer` (обычно `os.Stdout`)

### 3.5. `internal/flags/` — Валидация

```go
flags.ValidateRequiredID(args, index, name)   // Парсинг ID из аргументов
flags.GetFlagInt64(cmd, name)                 // int64 флаг
flags.GetFlagString(cmd, name)                // string флаг
flags.GetFlagBool(cmd, name)                  // bool флаг
flags.ParseID(s)                              // строка → int64
```

### 3.6. `internal/output/` — Сохранение

```go
output.AddFlag(cmd)                       // Регистрация --save, --save-to
output.OutputResult(cmd, data, resource)  // Вывод + сохранение
output.Output(cmd, data, dir, format)     // Сохранение в ~/.gotr/exports/
```

**DryRunPrinter** — вывод для `--dry-run` режима (в `os.Stderr`):

```go
printer := output.NewDryRunPrinter("sync cases")

// Полная операция с HTTP-деталями:
printer.PrintOperation("Create case", "POST", "/api/v2/add_case/1", requestBody)

// Простая операция без body:
printer.PrintSimple("Delete case", "Would delete case #123")

// Пакетная операция (показывает до 10 элементов):
printer.PrintBatch("Sync shared steps", []string{"Step 1", "Step 2", ...})
```

Вывод: ASCII-рамки с метаданными (Command, Operation, HTTP Method, Endpoint, Request Body).

### 3.7. `internal/log/` — Структурированное логирование

Централизованное логирование через `go.uber.org/zap`:

```go
log.InitDefault()     // В main.go — инициализация с дефолтным конфигом
defer log.Sync()      // Сброс буферов при выходе
log.L()               // Глобальный логгер (fallback → zap.NewNop())
log.Debug("msg")      // Отладочное сообщение
```

**Конфигурация** (`log.Config`):
- `Level` — уровень (`debug`, `info`, `warn`, `error`)
- `JSONFormat` — JSON-формат (для machine-parseable логов)
- `LogDir` — директория логов (по умолчанию `~/.gotr/logs/`)
- `Development` — режим разработки (stack traces, line numbers)

**Когда что использовать:**
- `log.L()` — внутренние события, аудит, диагностика (записывается в файл)
- `ui.Info()` — пользовательский вывод в терминал
- `utils.DebugPrint()` — отладка по флагу `--debug` (в stderr)

### 3.8. `internal/paths/` — Управление путями

Централизованные пути — все директории gotr в одном месте:

```
~/.gotr/                    # BaseDir()
├── config/                 # ConfigDirPath() — конфигурация
│   └── default.yaml        # Основной конфиг
├── logs/                   # LogsDirPath() — логи zap
├── selftest/               # SelftestDirPath() — отчёты selftest
├── cache/                  # CacheDirPath() — кэш API
├── exports/                # ExportsDirPath() — экспорт данных (--save)
└── temp/                   # TempDirPath() — временные файлы (jq)
```

**Правило:** Все пути через `paths.*` — не конструировать вручную.

### 3.9. `internal/progress/` — Прогресс-бары

Прогресс-бары на `github.com/vbauerster/mpb/v8`:

- `NewManager()` — создание менеджера прогресс-баров
- `NewBar(total, label)` — отдельный прогресс-бар
- `WithOutput(w)`, `WithQuiet()` — опции
- `WithMonitorCtx(ctx)` — мониторинг через context (автоотмена)
- Автоматически отключается в `--quiet` режиме

**Использование:** sync-команды, get-команды. В compare-командах заменён на `internal/ui/display`.

### 3.10. `internal/interactive/` — Интерактивный выбор

Зависимость: `github.com/AlecAivazis/survey/v2` — интерактивные промпты в терминале.

**Selector-функции** (API загрузка + промпт выбора):
```go
interactive.SelectProjectInteractively(ctx, client)  // Выбор проекта
interactive.SelectSuiteInteractively(ctx, client, projectID)  // Выбор сьюта
interactive.SelectRunInteractively(ctx, client, projectID)    // Выбор рана
```

**Wizard-функции** (`wizard.go`) — формы создания/обновления ресурсов:
```go
interactive.AskProject(isUpdate)  // → *ProjectAnswers
interactive.AskSuite(isUpdate)    // → *SuiteAnswers
interactive.AskCase(isUpdate)     // → *CaseAnswers
interactive.AskRun(isUpdate)      // → *RunAnswers
```

Типы промптов survey/v2: `survey.Input`, `survey.Multiline`, `survey.Confirm`, `survey.Select`.

**Использование:** Когда пользователь не указал ID — автоматический промпт.

### 3.11. `internal/selftest/` — Самодиагностика

Пакет для `gotr selftest` — проверка окружения:

**Интерфейс:**
```go
type Checker interface {
    Name() string
    Category() string
    Check() CheckResult
}
```

**Результаты:** `PASS` (✓), `FAIL` (✗), `WARN` (⚠), `SKIP` (⊘) — с ANSI-цветами.

**Встроенные проверки (6 checkers):**

| Checker | Категория | Что проверяет |
| ------- | --------- | ------------ |
| `ConfigChecker` | Configuration | Конфиг `~/.gotr/config/default.yaml` существует и валиден |
| `BaseDirChecker` | Configuration | Все 6 поддиректорий `~/.gotr/` (config, logs, selftest, cache, exports, temp); автосоздание через `os.MkdirAll` |
| `BinaryInfoChecker` | System | Версия, commit, дата сборки (всегда PASS) |
| `GoEnvChecker` | System | Go version, OS/arch, количество CPU |
| `AllTestsChecker` | Tests | Запуск `go test ./... -v`; подсчёт passed/failed/skipped; сохранение отчёта в `~/.gotr/selftest/` с симлинком `latest.log` |
| `CoverageChecker` | Coverage | `go test -coverprofile`; парсинг процента покрытия; WARN если < 50% |

- Самовосстановление: `CanFix: true` + `FixCommand: "gotr config init"` (ConfigChecker)
- Отчёт: `Report` struct с timestamp, version, platform, все CheckResult, общий Health

### 3.12. `internal/models/` — Модели данных

**`models/config/`** — Конфигурация приложения:
```go
type ConfigData struct {
    BaseURL  string `yaml:"base_url"`
    Username string `yaml:"username"`
    APIKey   string `yaml:"api_key"`
    Insecure bool   `yaml:"insecure"`
    JqFormat bool   `yaml:"jq_format"`
    Debug    bool   `yaml:"debug"`
}
```

Builder: `config.Default()` → `.WithDefaults()` → `.Create()`

**`models/data/`** — DTO для TestRail API (20+ файлов):
- Cases, Projects, Runs, Results, Suites, Sections
- SharedSteps, Milestones, Plans, Attachments
- Configs, Users, Reports, Roles, Templates, Tests
- Groups, Labels, Datasets, BDDs
- Status, Priority — константы

### 3.13. `internal/utils/` — Утилиты (legacy)

> **⚠️ God-package.** Расформирование запланировано в Stage 10.0.

- `DebugPrint(format, args...)` — вывод при `--debug`
- `OpenEditor(path)` — открытие файла в `$EDITOR` (fallback: vi/notepad)
- `GetFieldValue(obj, field)` — reflection-based извлечение поля (case-insensitive)
- `LoadMapping(path)` — загрузка JSON/YAML маппингов `map[int64]int64`

### 3.14. `pkg/reporter/` — Builder для отчётов

Builder-паттерн для структурированных отчётов (go-pretty):

```go
reporter.New("cases").
    Section("General statistics").
    Stat("⏱️", "Execution time", elapsed).
    Stat("📦", "Total processed", total).
    Section("Comparison results").
    Stat("✅", "Only in P1", 145).
    StatIf("⚠️", "Errors", errs, errs > 0).  // Условный вывод
    Print()

// Или быстрый вариант:
reporter.CompareStats("suites", pid1, pid2, onlyIn1, onlyIn2, common, elapsed).Print()
```

**Emoji → ANSI:** Emoji передаются как hint, reporter маппит их в ANSI-colored width-1 символы (решает проблему alignment в разных терминалах).

### 3.15. `pkg/testrailapi/` — API-справочник

Структурированное представление всех TestRail API v2 endpoints:

```go
api := testrailapi.New()
allPaths := api.Paths()           // Все endpoints (для gotr resources)
casePaths := api.Cases.Paths()    // Endpoints ресурса Cases
```

Типы: `APIPath{Method, URI, Description, Params}`, 26 ресурсов.

### 3.16. `embedded/` — Встроенный jq

Встроенные бинарники jq через `//go:embed`:

- `jq-linux-amd64`, `jq-macos-amd64`, `jq-windows-i386.exe`
- `RunEmbeddedJQ(rawBody, filter)` — извлечение → temp файл → выполнение → очистка
- Авто-определение платформы через `runtime.GOOS`
- Вызывается для `--jq` флага (фильтрация JSON-ответов)

### 3.17. `cmd/compare/` — Подсистема сравнения

Архитектура 13 compare-подкоманд:

**Типы данных:**
- `CompareResult` — результат сравнения (OnlyInFirst, OnlyInSecond, Common)
- `ItemInfo` — ID + Name
- `CommonItemInfo` — Name + ID1 + ID2 + IDsMatch

**Профили конфигурации** (`config_profile.go`):
- Авто-определение deployment: Cloud vs Server по URL
- Cloud rate limits: professional (180 req/min), enterprise (300 req/min)
- Server: без ограничений
- `resolveCompareCasesRuntimeConfig()` — расчёт параметров

**Экспорт** — multi-format: JSON, YAML, CSV, Table (авто по расширению файла).

**Generic Factory** — `newSimpleCompareCmd[T]()` для DRY 13 подкоманд.

### 3.18. `cmd/sync/` — Подсистема синхронизации

Пакет для синхронизации данных между проектами:

**Структура:**
- `sync.go` — регистрация и оркестрация
- `sync_full.go` — полная миграция (SharedSteps → Suites → Sections → Cases)
- `sync_cases.go`, `sync_suites.go`, `sync_sections.go`, `sync_shared_steps.go` — per-resource
- `sync_flags.go` — валидация флагов `--src-project`, `--dst-project`
- `sync_helpers.go` — общие хелперы
- `interactive.go` — интерактивный выбор source/destination
- `sync_test_helper.go`, `sync_test_skip.go` — тестовая инфраструктура

**Порядок sync full:** SharedSteps → Suites → Sections → Cases (зависимости!).

### 3.19. `cmd/internal/testhelper` — Тестовые утилиты

Общий пакет для тестирования CLI-команд (доступен только внутри `cmd/`):

```go
testhelper.HTTPClientKey                          // Ключ контекста для mock
testhelper.SetupTestCmd(t, mock)                  // Команда + mock в контексте
testhelper.SetupTestCmdWithBuffer(t, mock)        // + буфер вывода
testhelper.GetClientForTests(cmd)                 // Извлечение mock из контекста
```

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
| Константы | CamelCase | `DefaultBaseURL`, `ResultPass` |
| Context key | Типизированный тип | `type ctxKey string` (не `string` напрямую) |

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

**Иерархия обработки:**
1. Библиотечный код — возвращает `error` с контекстом (`%w`)
2. Сервисный слой — валидирует входные данные, оборачивает ошибки
3. CLI слой (`RunE`) — возвращает ошибку Cobra (Cobra печатает в stderr)
4. `GetClient*` — единственное место с `panic` (невозможно продолжить без клиента)

### 4.3. Context

- Все I/O функции принимают `ctx context.Context` первым аргументом
- `cmd.Context()` — источник контекста в CLI-слое
- Ctrl+C работает через `signal.NotifyContext` (Stage 7.0)
- `http.NewRequestWithContext(ctx, ...)` — в HTTP-транспорте
- Context key: всегда типизированный, не `string` напрямую

### 4.4. Язык

| Контекст | Язык |
| -------- | ---- |
| Исходный код (переменные, функции) | Английский |
| Пользовательский вывод (UI) | Английский |
| Ошибки (`fmt.Errorf`) | Английский |
| Комментарии в коде | Английский |
| Документация (`docs/`, `README*.md`) | Русский |

### 4.5. Глобальные флаги

| Флаг | Короткий | Тип | Описание |
| ---- | -------- | --- | -------- |
| `--url` | — | string | Базовый URL TestRail |
| `--username` | `-u` | string | Email пользователя |
| `--api-key` | `-k` | string | API ключ |
| `--insecure` | — | bool | Пропустить проверку TLS |
| `--config` | `-c` | bool | Создание дефолтного конфига |
| `--debug` | `-d` | bool | Отладочный вывод (скрытый) |
| `--quiet` | `-q` | bool | Тихий режим (CI/CD) |
| `--format` | `-f` | string | Формат: `table`, `json`, `csv`, `md`, `html` |

### 4.6. Форматы вывода

| Формат | Описание | Пример |
| ------ | -------- | ------ |
| `table` | ASCII-таблица (go-pretty), по умолчанию | `gotr cases list 30` |
| `json` | JSON output | `gotr cases list 30 -f json` |
| `csv` | CSV (для Excel/скриптов) | `gotr cases list 30 -f csv` |
| `md` | Markdown таблица | `gotr cases list 30 -f md` |
| `html` | HTML таблица | `gotr cases list 30 -f html` |

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
// WorkerPool
pool := concurrent.NewWorkerPool(
    concurrent.WithMaxWorkers(5),
    concurrent.WithRateLimit(150),
    concurrent.WithProgressMonitor(monitor),
)

// FetchParallel
result, err := concurrency.FetchParallel(ctx, fetchFn,
    concurrency.WithReporter(reporter),
    concurrency.WithMaxConcurrency(5),
    concurrency.WithContinueOnError(),
)
```

### 5.4. Builder Pattern

```go
// Reporter
reporter.New("cases").
    Section("Projects").
    Stat("📁", "Project 1", p1Name).
    Print()

// Config
config.Default().WithDefaults().Create()
```

### 5.5. PersistentPreRunE Override

Некоторые команды (например, `config`) не требуют клиента TestRail. Для них отключается `PersistentPreRunE`:

```go
configCmd := &cobra.Command{
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        return nil // Пропускаем создание клиента
    },
}
```

### 5.6. Entry Point Pattern

```go
func main() {
    log.InitDefault()          // 1. Инициализация логгера
    defer log.Sync()           // 2. Сброс буферов при выходе
    ctx, stop := signal.NotifyContext(context.Background(),
        os.Interrupt, syscall.SIGTERM)
    defer stop()               // 3. Ctrl+C → отмена контекста
    cmd.Execute(ctx)           // 4. Запуск CLI
}
```

---

## 6. Конфигурация и окружение

### 6.1. Приоритет конфигурации

1. **CLI flags** — наивысший приоритет (`--url`, `--username`, `--api-key`)
2. **Environment variables** — `TESTRAIL_BASE_URL`, `TESTRAIL_USERNAME`, `TESTRAIL_API_KEY`
3. **Config file** — `~/.gotr/config/default.yaml`

### 6.2. Viper-интеграция

- `viper.AutomaticEnv()` — автоматическое чтение env-переменных
- `viper.BindPFlag(key, flag)` — привязка CLI-флагов к Viper-ключам
- `TESTRAIL_*` — префикс для env-переменных
- `TESTRAIL_PASSWORD` имеет приоритет над `TESTRAIL_API_KEY` (обратная совместимость — для пользователей, мигрирующих с password auth)

### 6.3. Config-команды

| Команда | Описание |
| ------- | -------- |
| `gotr config init` | Создать `~/.gotr/config/default.yaml` с placeholder'ами |
| `gotr config path` | Показать путь к конфигу |
| `gotr config view` | Вывести текущий конфиг |
| `gotr config edit` | Открыть конфиг в `$EDITOR` |

### 6.4. Структура `~/.gotr/`

```
~/.gotr/
├── config/default.yaml     # Конфигурация (base_url, username, api_key, ...)
├── logs/                   # Логи zap (ежедневная ротация)
├── selftest/               # Отчёты gotr selftest
├── cache/                  # Кэш API-ответов
├── exports/                # Экспорт данных (--save, --save-to)
└── temp/                   # Временные файлы (jq бинарник и т.д.)
```

---

## 7. Конкурентность и устойчивость

Два уровня абстракции + профили деплоймента.

### 7.0. Режимы деплоймента и rate limit

gotr автоматически определяет режим деплоймента TestRail и подбирает rate limit.

**Авто-определение** (`config_profile.go`):
```go
func detectDeploymentByURL(baseURL string) string {
    if strings.Contains(url, ".testrail.io") { return "cloud" }
    return "server"
}
```

**Профили rate limit:**

| Deployment | Cloud Tier | Rate Limit | Источник |
| ---------- | ---------- | ---------- | -------- |
| **Cloud** | Professional | 180 req/min | Лимит TestRail API (hardcoded) |
| **Cloud** | Enterprise | 300 req/min | Viper `compare.cloud_rate_limit` (default 300) |
| **Server** | — | 0 (unlimited) | Viper `compare.server_rate_limit` (default 0) |

**Почему дефолт WorkerPool = 150 req/min:**

TestRail Cloud Professional разрешает 180 req/min. WorkerPool по умолчанию использует 150 — это консервативный предел с запасом ~17%, чтобы:
- Избежать 429 (Too Many Requests) при burst-нагрузке
- Оставить запас для burst (15% от rate = 22 токена)
- Работать стабильно без тюнинга на большинстве инсталляций

**Расчёт burst:** `burst = requestsPerMinute * 15 / 100` (минимум 10).
При 150 req/min: burst = 22. При 300 req/min: burst = 45.

**RateLimiter fallback:** Если `requestsPerMinute <= 0`, `NewRateLimiter` использует 180 req/min (потолок Professional).

**Viper-ключи конфигурации (compare):**

| Ключ | Default | Описание |
| ---- | ------- | -------- |
| `compare.deployment` | `auto` | `auto`, `cloud`, `server` |
| `compare.cloud_tier` | `professional` | `professional`, `enterprise` |
| `compare.rate_limit` | `-1` (авто) | Явный лимит (перекрывает профиль) |
| `compare.cloud_rate_limit` | `300` | Лимит для cloud (когда rate_limit=-1) |
| `compare.server_rate_limit` | `0` | Лимит для server (0 = без лимита) |
| `compare.cases.parallel_suites` | `10` | Параллельных сьютов |
| `compare.cases.parallel_pages` | `6` | Параллельных страниц на сьют |
| `compare.cases.page_retries` | `5` | Ретраев на страницу |
| `compare.cases.timeout` | `30m` | Общий таймаут compare cases |
| `compare.cases.auto_retry_failed_pages` | `true` | Авто-ретрай упавших страниц |

**Приоритет rate limit:** CLI flag `--rate-limit` > Viper `compare.rate_limit` > профиль (по deployment + tier).

### 7.1. `internal/concurrent/` — Низкоуровневые примитивы

**WorkerPool** — пул горутин на `errgroup`:
```go
pool := concurrent.NewWorkerPool(
    concurrent.WithMaxWorkers(5),       // 5 параллельных воркеров (default)
    concurrent.WithRateLimit(150),      // 150 req/min (default)
    concurrent.WithProgressMonitor(m),  // опциональный Increment() callback
)
pool.Submit(func() error { return doWork() })
err := pool.Wait()
```

**RateLimiter** — token bucket (`golang.org/x/time/rate`):
- Преобразование req/min → rate/sec: `rate.Limit(float64(rpm) / 60.0)`
- Burst: 15% от rate (минимум 10)
- Дефолт: 180 req/min (fallback в NewRateLimiter при rpm ≤ 0)
- WorkerPool дефолт: 150 req/min (консервативный предел)
- API: `Wait()`, `WaitWithTimeout(timeout)`, `Allow()`, `Tokens()`, `Reserve()`

**Retry** — экспоненциальная задержка:
```go
config := &concurrent.RetryConfig{
    MaxRetries:   5,           // Количество попыток
    InitialDelay: 1 * time.Second,
    MaxDelay:     30 * time.Second,
    Multiplier:   2.0,         // 1s → 2s → 4s → 8s → 16s
}
err := concurrent.Retry(config, func() error { return apiCall() })
// Или с context:
err := concurrent.RetryWithContext(ctx, config, fn)
```

**CircuitBreaker** — защита от каскадных ошибок:
- Состояния: Closed → Open (при N ошибках) → HalfOpen (после timeout) → Closed
- `NewCircuitBreaker(maxFailures, timeout)`
- `Execute(fn)` — выполнение с защитой

### 7.2. `internal/concurrency/` — Высокоуровневые стратегии

Для compare-подкоманд — три стратегии параллелизации:

| Стратегия | Паттерн | Пример |
| --------- | ------- | ------ |
| `FetchParallel[T]` | Лёгкая: N проектов параллельно | compare suites/milestones |
| `FetchParallelBySuite[T]` | Средняя: per-suite параллельно | compare sections |
| `ParallelController` | Тяжёлая: pipeline pagination | compare cases |

**FetchOption-паттерн:** `WithReporter()`, `WithContinueOnError()`, `WithMaxConcurrency(n)`

**Интерфейсы прогресса:**
- `ProgressReporter` — `OnItemComplete`, `OnBatchReceived`, `OnError`
- `PaginatedProgressReporter` — расширение: `OnPageFetched`

**Конфигурационные флаги (compare cases):**
- `--parallel-suites` — параллельность по сьютам
- `--parallel-pages` — параллельность по страницам
- `--page-retries` — количество повторов для страниц
- `--rate-limit` — лимит запросов
- `--timeout` — таймаут

### 7.3. Правила конкурентности

- Горутины **всегда** получают `context.Context`
- `errgroup.WithContext()` — для управления группой горутин
- Мьютексы: `sync.Mutex` для shared state, `sync.Once` для инициализации
- Каналы: предпочитать каналы мьютексам для коммуникации
- Race detection: `go test -race ./...` — обязательно в CI

---

## 8. Логирование

### 8.1. Три уровня вывода

| Уровень | Инструмент | Куда | Когда |
| ------- | ---------- | ---- | ----- |
| **Пользовательский** | `ui.Info()`, `ui.Success()`, etc. | stdout | Всегда (кроме `--quiet`) |
| **Отладочный** | `utils.DebugPrint()` | stderr | Только при `--debug` |
| **Структурированный** | `log.L().Info()`, `log.L().Error()` | `~/.gotr/logs/` | Всегда (в файл) |

### 8.2. Когда что использовать

```go
// Пользователь должен увидеть в терминале:
ui.Info(os.Stdout, "Loading cases...")
ui.Success(os.Stdout, "Done!")

// Debug-вывод (только с --debug):
utils.DebugPrint("{syncCases} Processing suite %d", suiteID)

// В файл для диагностики:
log.L().Info("API call completed",
    zap.Int64("project_id", pid),
    zap.Duration("elapsed", elapsed),
)
log.L().Error("API request failed",
    zap.Error(err),
    zap.String("endpoint", url),
)
```

---

## 9. Тестирование

### 9.1. Общие требования

| Требование | Описание |
| ---------- | -------- |
| Покрытие | ≥ 85% для каждого пакета (цель 90%+) |
| Паттерн | Table-driven tests |
| Моки | `client.MockClient` |
| DI | `newXxxCmd(clientFn)` для инъекции |
| Naming | `TestXxx_Success`, `TestXxx_Error`, `TestXxx_EdgeCase` |
| Без сети | Unit-тесты без внешних вызовов |
| Race | `go test -race ./...` — 0 data races |

### 9.2. Тестовый паттерн CMD

Стандартный паттерн для тестирования CLI-команд:

```go
func TestGetCase_Success(t *testing.T) {
    mock := &client.MockClient{
        GetCaseFunc: func(ctx context.Context, caseID int64) (*data.Case, error) {
            return &data.Case{ID: caseID, Title: "Test"}, nil
        },
    }
    cmd := testhelper.SetupTestCmd(t, mock)

    getCmd := newGetCmd(testhelper.GetClientForTests)
    getCmd.SetArgs([]string{"123"})
    getCmd.SetContext(cmd.Context())

    err := getCmd.Execute()
    assert.NoError(t, err)
}
```

### 9.3. Тестовый паттерн Service

```go
func TestRunService_Create_ValidatesProjectID(t *testing.T) {
    tests := []struct {
        name      string
        projectID int64
        wantErr   bool
    }{
        {"valid", 30, false},
        {"zero", 0, true},
        {"negative", -1, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            svc := NewRunService(mockClient)
            _, err := svc.Create(ctx, tt.projectID, req)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

---

## 10. Сборка и релиз

### 10.1. Makefile targets

| Target | Описание |
| ------ | -------- |
| `make build` | Сборка с версией из `cmd/root.go` |
| `make build VERSION=v3.0.0` | Сборка с явной версией |
| `make test` | Запуск тестов (`go test ./... -v`) |
| `make install` | Установка в `/usr/local/bin` (sudo) |
| `make release` | Cross-compilation: Linux + macOS + Windows |
| `make release-compressed` | То же + UPX сжатие |
| `make tag VERSION=v3.0.0` | Создание git tag + push |
| `make clean` | Удаление бинарника |

### 10.2. Версионирование

Версия встраивается через `-ldflags` при сборке:

```go
var (
    Version = "2.7.0"   // Значение по умолчанию (для go run)
    Commit  = "unknown"
    Date    = "unknown"
)
```

`cmd/root.go` → `rootCmd.Version = Version` → `gotr --version`

### 10.3. Кросс-компиляция

Поддерживаемые платформы:
- `linux/amd64`
- `darwin/amd64`
- `windows/amd64`

UPX сжатие опционально (если установлен). Для macOS — `--force-macos`.

### 10.4. Линтинг

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
golangci-lint run --timeout 5m
```

Конфигурация: `.golangci.yml` (11 линтеров). Baseline: 208 замечаний (Stage 9.0).

### 10.5. Зависимости (go.mod)

| Зависимость | Версия | Назначение |
| ----------- | ------ | ---------- |
| `github.com/spf13/cobra` | v1.10.2 | CLI-фреймворк (команды, флаги, completion) |
| `github.com/spf13/viper` | v1.21.0 | Конфигурация (yaml, env, pflags) |
| `go.uber.org/zap` | v1.27.1 | Структурированное логирование |
| `github.com/jedib0t/go-pretty/v6` | v6.7.8 | Таблицы, рендеринг (table/json/csv/md/html) |
| `github.com/vbauerster/mpb/v8` | v8.12.0 | Прогресс-бары (sync, get команды) |
| `github.com/AlecAivazis/survey/v2` | v2.3.7 | Интерактивные промпты (wizard) |
| `golang.org/x/time` | v0.14.0 | Token bucket rate limiter |
| `golang.org/x/sync` | v0.19.0 | errgroup (WorkerPool) |
| `github.com/fatih/color` | v1.18.0 | ANSI-цвета (selftest) |
| `gopkg.in/yaml.v3` | v3.0.1 | YAML parsing |
| `github.com/stretchr/testify` | v1.11.1 | Assertions (require, assert) |

---

## 11. Контрольный список (Checklist)

### Новая команда

- [ ] Конструктор: `newXxxCmd(clientFn)`
- [ ] `ctx := cmd.Context()` передаётся во все I/O вызовы
- [ ] Ошибки обёрнуты: `fmt.Errorf("... : %w", err)`
- [ ] Вывод — через `ui.*` (не `fmt.Printf`)
- [ ] Валидация — через `flags.*`
- [ ] Сохранение — через `output.*`
- [ ] Регистрация в `commands.go` через `Register(rootCmd, clientFn)`
- [ ] Тест: ≥ success + error + edge case (через `testhelper`)
- [ ] Нет дублирования (≥ 3 раз → хелпер)

### Новый API-метод

- [ ] Сигнатура: `func (c *HTTPClient) GetXxx(ctx context.Context, ...)`
- [ ] Добавлен в `interfaces.go` (соответствующий интерфейс)
- [ ] Добавлен в `mock.go` (`MockClient`)
- [ ] List-методы: `fetchAllPages[T]`
- [ ] Тест с `MockClient`

### Общая проверка

- [ ] Нет кириллицы в коде (только `docs/` и `README`)
- [ ] `go vet ./...` — 0 предупреждений
- [ ] `go build ./...` — 0 ошибок
- [ ] `go test ./...` — 0 FAIL
- [ ] `go test -race ./...` — 0 data races
- [ ] Пути через `paths.*` (не хардкоженные)
- [ ] Логирование: `ui.*` для терминала, `log.L()` для файла
