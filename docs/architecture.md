# Архитектура gotr

> Общее описание архитектуры CLI-утилиты gotr для пользователей  
> **Важно:** Этот файл актуализируется при добавлении новых команд или изменении структуры проекта. Последнее обновление: 2026-03-12 (v3.0.0-dev) — Stage 9.0: Standards.

## Что такое gotr

`gotr` — это CLI-клиент для TestRail API v2, построенный по многослойной архитектуре с чётким разделением ответственности между слоями.

## Общая схема

```
┌─────────────────────────────────────────────────────────────┐
│  CLI Layer (cmd/*)                                          │
│  • Парсинг аргументов и флагов (flags.*)                    │
│  • Интерактивный выбор (internal/interactive)               │
│  • Вывод данных (ui.*, output.*)                            │
│  • Вызов сервисов и клиента                                 │
└──────────────┬──────────────────────┬───────────────────────┘
               │                      │
    ┌──────────▼──────────┐  ┌────────▼──────────┐
    │  UI Layer           │  │  Flags Layer       │
    │  (internal/ui/)     │  │  (internal/flags/) │
    │  • Table, JSON      │  │  • ValidateID      │
    │  • Info, Success,   │  │  • GetFlag[T]      │
    │    Warning, Error   │  │  • ParseID         │
    │  • Live display     │  └───────────────────┘
    └─────────────────────┘
               │
┌──────────────▼──────────────────────────────────────────────┐
│  Service Layer (internal/service/*)                         │
│  • Бизнес-логика                                            │
│  • Валидация данных                                         │
│  • Миграция данных (migration)                              │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│  Concurrency Layer (internal/concurrency/*)                 │
│  • ParallelController — pipeline pagination по сьютам      │
│  • FetchParallel[T] — лёгкая стратегия (по проектам)       │
│  • FetchParallelBySuite[T] — средняя стратегия (по сьютам) │
│  • ResultAggregator — сбор результатов из горутин          │
│  • PriorityQueue — приоритезация больших сьютов            │
│  • SuiteFetcher — интерфейс для реализаций загрузки        │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│  Concurrent Layer (internal/concurrent/*)                   │
│  • WorkerPool — параллельная обработка запросов            │
│  • RateLimiter — контроль 180 запросов/минуту              │
│  • Retry — повторные попытки с экспоненциальной задержкой  │
│  • CircuitBreaker — защита от каскадных ошибок             │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│  Client Layer (internal/client/*)                           │
│  • HTTPClient — реальный клиент                             │
│  • ClientInterface — абстракция для тестов (106 endpoints)  │
│  • MockClient — реализация для тестирования                 │
│  • fetchAllPages[T] — generic-пагинатор (Stage 6.9)        │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│  Output Layer (internal/output/*)                           │
│  • OutputResult — вывод + сохранение в файл                │
│  • DryRunPrinter — вывод для dry-run режима                │
│  • --save / --save-to управление                            │
└─────────────────────────────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│  TestRail API v2                                            │
└─────────────────────────────────────────────────────────────┘
```

### Правила зависимостей

| Слой | Может зависеть от | НЕ может зависеть от |
| ---- | ----------------- | -------------------- |
| `cmd/*` | `internal/service`, `internal/client`, `internal/ui`, `internal/flags`, `internal/interactive`, `internal/output`, `pkg/*` | — |
| `internal/service` | `internal/client`, `internal/concurrency`, `internal/concurrent`, `internal/models` | `cmd/*`, `internal/ui` |
| `internal/client` | `internal/concurrent`, `internal/models/data` | `cmd/*`, `internal/service` |
| `internal/ui` | stdlib, `go-pretty/v6` | `internal/client`, `internal/service` |
| `internal/concurrency` | stdlib | `internal/client`, `cmd/*` |
| `pkg/*` | stdlib, `go-pretty/v6` | `internal/*`, `cmd/*` |

**Запрещено:**
- `service/` → `cmd/` (обращение вверх)
- `client/` → `ui/` (клиент не знает о UI)
- `pkg/` → `internal/` (публичный не импортирует приватный)

## Слои подробно

### 1. CLI Layer (`cmd/`)

**Ответственность:** Принимает команды от пользователя, парсит аргументы, вызывает сервисы.

**Структура:**
```
cmd/
├── root.go              # Корневая команда, Execute(ctx)
├── commands.go          # Регистрация всех подкоманд (init())
├── add.go               # gotr add <resource>
├── update.go            # gotr update <resource>
├── delete.go            # gotr delete <resource>
├── list.go              # gotr list <resource>
├── export.go            # gotr export
├── config.go            # gotr config {init|path|view|edit}
├── resources.go         # gotr resources (API endpoints)
├── selftest.go          # gotr selftest
├── completion.go        # gotr completion {bash|zsh|fish}
├── <resource>/          # Подкоманды для ресурса
│   ├── <resource>.go   #   Register() + getClient()
│   ├── add.go          #   newAddCmd(clientFn)
│   ├── get.go          #   newGetCmd(clientFn)
│   ├── list.go         #   newListCmd(clientFn)
│   ├── update.go       #   newUpdateCmd(clientFn)
│   ├── delete.go       #   newDeleteCmd(clientFn)
│   └── *_test.go       #   Table-driven тесты
├── compare/             # gotr compare <resource> --pid1 X --pid2 Y
├── sync/                # gotr sync {cases|sections|suites|...}
└── internal/            # Общие тестовые хелперы (testhelper)
```

**Паттерн команды (Stage 7.0+):**
```go
func newXxxCmd(clientFn func(*cobra.Command) client.ClientInterface) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "xxx",
        Short: "Brief description",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            cli := clientFn(cmd)
            ctx := cmd.Context()

            id, err := flags.ValidateRequiredID(args, 0, "resource_id")
            if err != nil { return err }

            result, err := cli.GetXxx(ctx, id)
            if err != nil { return err }

            return output.OutputResult(cmd, result, "xxx")
        },
    }
    output.AddFlag(cmd)
    return cmd
}
```

**Пример:**
```bash
gotr run get 12345 --jq
# cmd/run/get.go → client.GetRun(ctx, 12345) → вывод через output.OutputResult
```

### 2. Service Layer (`internal/service/`)

**Ответственность:** Бизнес-логика, валидация, оркестрация операций.

**Компоненты:**
- `RunService` — работа с test runs
- `ResultService` — работа с результатами тестов
- `migration/` — миграция данных между проектами
  - `types.go` — контекст миграции
  - `fetch.go` — получение данных
  - `filter.go` — фильтрация дубликатов
  - `import.go` — импорт сущностей
  - `export.go` — экспорт данных
  - `mapping.go` — управление mapping ID

**Валидация:**
```go
// Проверки перед созданием run:
// - projectID > 0
// - name не пустое
// - suite_id > 0 (если указан)
```

### 3. Client Layer (`internal/client/`)

**Ответственность:** HTTP-запросы к TestRail API.

**Структура:**
```
internal/client/
├── client.go           # HTTPClient — конструктор, DoRequest (http.NewRequestWithContext)
├── interfaces.go       # ClientInterface + 14 API групп (106 endpoints)
├── mock.go             # MockClient для тестирования
├── paginator.go        # Generic fetchAllPages[T] — автопагинация list-методов (Stage 6.9)
├── request.go          # sendRequest(), debug вывод
├── accessor.go         # ClientAccessor — lazy init
├── concurrent.go       # Thread-safe обёртки
├── projects.go         # ProjectsAPI (5 endpoints)
├── cases.go            # CasesAPI (14 endpoints)
├── suites.go           # SuitesAPI (5 endpoints)
├── sections.go         # SectionsAPI (5 endpoints)
├── sharedsteps.go      # SharedStepsAPI (6 endpoints)
├── runs.go             # RunsAPI (6 endpoints)
├── results.go          # ResultsAPI (7 endpoints)
├── tests.go            # TestsAPI (3 endpoints)
├── milestones.go       # MilestonesAPI (5 endpoints)
├── plans.go            # PlansAPI (9 endpoints)
├── attachments.go      # AttachmentsAPI (5 endpoints)
├── configs.go          # ConfigurationsAPI (7 endpoints)
├── users.go            # UsersAPI (4 endpoints)
├── reports.go          # ReportsAPI (3 endpoints)
└── extended.go         # ExtendedAPI (21 endpoint)
```

**Ключевые факты (Stage 7.0+):**
- Каждый метод принимает `ctx context.Context` первым аргументом
- List-методы используют `fetchAllPages[T]` для автопагинации
- Отмена через context (Ctrl+C → signal.NotifyContext)
- 14 интерфейсов по ISP, 106 endpoints, 100% покрытие TestRail API v2

### 4. Concurrent Layer (`internal/concurrent/`)

**Ответственность:** Параллельная обработка API запросов с контролем нагрузки.

**Зачем нужен:** TestRail API имеет лимит 180 запросов/минуту. Последовательная обработка множества сьютов/кейсов занимает минуты. Concurrent Layer позволяет:
- Выполнять запросы параллельно (до 5 одновременно)
- Автоматически регулировать скорость (150 req/min)
- Автоматически повторять при ошибках
- Защищать от перегрузки API

**Пример ускорения:**
```
Загрузка кейсов из 10 сьютов:
- Последовательно: 10 запросов × 1 сек = 10 секунд
- Параллельно: 10 запросов / 5 workers = ~2 секунды (5x ускорение)
```

**Компоненты:**
- **WorkerPool** — управление пулом горутин (3-5 воркеров)
- **RateLimiter** — token bucket (150 запросов/минуту)
- **Retry** — повтор с экспоненциальной задержкой (1с, 2с, 4с...)
- **CircuitBreaker** — блокировка при множестве ошибок

Подробнее: [docs/concurrent.md](./concurrent.md)

### 5. Concurrency Layer (`internal/concurrency/`)

**Ответственность:** Унифицированные стратегии параллелизации для всех compare-подкоманд.

**Компоненты:**
- **FetchParallel[T]** — лёгкая стратегия: загрузка ресурса из N проектов параллельно (generic, Go 1.24+)
- **FetchParallelBySuite[T]** — средняя стратегия: загрузка per-suite ресурсов параллельно (sections и др.)
- **ParallelController** — тяжёлая стратегия: pipeline pagination по сьютам (cases)
- **ResultAggregator** — потокобезопасная агрегация результатов из горутин
- **PriorityQueue** — heap-based очередь, большие сьюты обрабатываются первыми
- **SuiteFetcher** — интерфейс (`FetchPageCtx`) для подстановки реальных и mock-реализаций
- **ProgressReporter** — универсальный интерфейс прогресса (`OnItemComplete`, `OnBatchReceived`, `OnError`)
- **PaginatedProgressReporter** — расширение для стратегий с пагинацией (`OnPageFetched`)

**FetchOption-паттерн:** `WithReporter()`, `WithContinueOnError()`, `WithMaxConcurrency(n)`

**Конфигурация:** `--parallel-suites`, `--parallel-pages`, `--page-retries`, `--rate-limit`, `--timeout`

Подробнее: [docs/recursive-parallelization-plan.md](./recursive-parallelization-plan.md)

### 6. UI Layer (`internal/ui/` + `pkg/reporter/`)

**Ответственность:** Унифицированный вывод для всех команд (Stage 8.0).

**Компоненты:**

- **internal/ui/display.go** — ANSI live display с динамическими задачами (для `compare cases`)
  - `New()`, `SetHeader()`, `AddTask()`, `Finish()` — lifecycle
  - Реализует `ProgressReporter`, `PaginatedProgressReporter` из `concurrency/`

- **internal/ui/table.go** — статический вывод данных
  - `NewTable(cmd)` — go-pretty таблица с учётом `--format` (table/json/csv/md/html)
  - `Table(cmd, t)` — рендеринг таблицы
  - `JSON(cmd, data)` — JSON-вывод с учётом `--quiet`
  - `IsJSON(cmd)`, `IsQuiet(cmd)` — проверки формата

- **internal/ui/helpers.go** — стилизованные сообщения (Stage 8.0)
  - `Info(w, msg)` — ℹ️ информация
  - `Success(w, msg)` — ✅ успех
  - `Warning(w, msg)` — ⚠️ предупреждение
  - `Error(w, msg)` — ❌ ошибка
  - `Phase(w, msg)` — 🔄 фаза
  - `Stat(w, icon, label, val)` — статистика
  - `Section(w, msg)` — заголовок секции
  - `Preview(w, title, fields)` — окно предпросмотра

- **pkg/reporter/** — builder-pattern для структурированных отчётов (ANSI + go-pretty)

**Правила:**
- Весь пользовательский вывод — через `ui.*` (кроме интерактивных промптов и debug)
- Emoji-префиксы только в `ui.*` — никаких хардкоженных emoji в cmd/
- Первый аргумент — `io.Writer` (обычно `os.Stdout`)

### 7. Flags Layer (`internal/flags/`)

**Ответственность:** Типобезопасная валидация аргументов и флагов CLI.

**Функции:**
```go
flags.ValidateRequiredID(args, index, name)   // Парсинг ID из аргументов
flags.GetFlagInt64(cmd, name)                 // int64 флаг
flags.GetFlagString(cmd, name)                // string флаг
flags.GetFlagBool(cmd, name)                  // bool флаг
flags.ParseID(s)                              // строка → int64
```

### 8. Output Layer (`internal/output/`)

**Ответственность:** Сохранение результатов в файлы, dry-run, форматирование.

**Функции:**
```go
output.AddFlag(cmd)                       // Регистрация --save, --save-to
output.OutputResult(cmd, data, resource)  // Вывод + сохранение
output.Output(cmd, data, dir, format)     // Сохранение в ~/.gotr/exports/
output.NewDryRunPrinter(cmd)              // Вывод для dry-run
```

**Использование:**
- `gotr compare cases` — live display с прогрессом в реальном времени
- Все 13 compare-подкоманд — reporter для итогового вывода
- `gotr compare all` — go-pretty table + reporter для сводной таблицы
- `ui.Infof(os.Stderr, ...)` — стилизованные сообщения

### 7. Progress Runtime (`internal/ui/runtime.go`)

**Ответственность:** Единый runtime прогресса и статусов для CLI-команд.

**Использование:**
- `ui.RunWithStatus(...)` — простые статусные операции
- `ui.NewOperation(...)` + `AddTask(...)` — многофазные и потоковые операции

**Особенности:**
- Учитывает `--quiet`
- Поддерживает фазы и task-level прогресс
- Используется в compare/sync/get и других командах

### 8. Interactive Layer (`internal/interactive/`)

**Ответственность:** Интерактивный выбор проектов, сьютов, ранов.

**Использование:**
- `gotr run list` — выбор проекта → список ранов
- `gotr result list` — выбор проекта → выбор рана → результаты
- `gotr get cases` — выбор проекта → выбор сьюта

### 9. Models (`internal/models/data/`)

**Ответственность:** DTO (Data Transfer Objects) для API.

**Основные структуры:**
- `Project`, `Suite`, `Section`, `Case`
- `Run`, `Test`, `Result`
- `SharedStep`, `Milestone`, `Plan`
- `Attachment`, `Config`, `User`
- `Report`, `Group`, `Role`, `Dataset`
- `Status`, `Priority` — константы

### 10. Utilities (`internal/utils/`)

**Ответственность:** Вспомогательные функции.

**Компоненты:**
- `helpers.go` — парсинг ID, вывод результатов, сохранение в файл
- `log.go` — директории для логов

## Поток данных

### Пример 1: Создание test run

```
Пользователь
    ↓
gotr run create 30 --suite-id 100 --name "Smoke"
    ↓
CLI Layer (cmd/run/create.go)
    ↓
RunService.Create(projectID=30, req={suite_id:100, name:"Smoke"})
    ↓
Валидация: projectID>0? suite_id>0? name не пустое?
    ↓
HTTPClient.AddRun(30, req)
    ↓
POST /index.php?/api/v2/add_run/30
    ↓
TestRail API
```

### Пример 2: Миграция данных (sync full)

```
Пользователь
    ↓
gotr sync full --src-project 30 --dst-project 31
    ↓
CLI Layer (cmd/sync/sync_full.go)
    ↓
migration.NewMigration(client, 30, 0, 31, 0, "title", logDir)
    ↓
Migration.FetchSharedStepsData()  // Получение данных
    ↓
Migration.FilterSharedSteps()     // Фильтрация дубликатов
    ↓
Migration.ImportSharedSteps()     // Импорт
    ↓
Аналогично для cases
    ↓
TestRail API (src) → Migration → TestRail API (dst)
```

## Полная структура проекта

```
gotr/
├── cmd/                          # CLI команды (Cobra)
│   ├── root.go                  #   Execute(ctx), initConfig()
│   ├── commands.go              #   init() — регистрация всех подкоманд
│   ├── add.go                   #   gotr add <endpoint>
│   ├── update.go                #   gotr update <endpoint>
│   ├── delete.go                #   gotr delete <endpoint>
│   ├── list.go                  #   gotr list <resource>
│   ├── export.go                #   gotr export <resource>
│   ├── config.go                #   gotr config {init|path|view|edit}
│   ├── resources.go             #   gotr resources
│   ├── selftest.go              #   gotr selftest
│   ├── completion.go            #   gotr completion {bash|zsh|fish}
│   ├── attachments/             #   gotr attachments {add|get|list|delete}
│   ├── bdds/                    #   gotr bdds {add|get}
│   ├── cases/                   #   gotr cases {add|get|list|update|delete|bulk}
│   ├── compare/                 #   gotr compare {cases|suites|sections|...}
│   ├── configurations/          #   gotr configurations {add|list|update|delete}
│   ├── datasets/                #   gotr datasets
│   ├── get/                     #   gotr get {project|suite|case|...}
│   ├── groups/                  #   gotr groups
│   ├── labels/                  #   gotr labels
│   ├── milestones/              #   gotr milestones
│   ├── plans/                   #   gotr plans
│   ├── reports/                 #   gotr reports
│   ├── result/                  #   gotr result {get|list|add}
│   ├── roles/                   #   gotr roles
│   ├── run/                     #   gotr run {get|list|create|update|close|delete}
│   ├── sync/                    #   gotr sync {full|cases|shared-steps|suites|sections}
│   ├── templates/               #   gotr templates
│   ├── test/                    #   gotr test
│   ├── tests/                   #   gotr tests
│   ├── users/                   #   gotr users
│   ├── variables/               #   gotr variables
│   └── internal/                #   Тестовые хелперы (testhelper)
├── docs/                         # Документация (русский)
│   ├── architecture.md          #   Этот файл
│   ├── standards.md             #   Стандарты кодирования
│   ├── concurrent.md            #   Параллельная обработка
│   ├── configuration.md         #   Настройка
│   ├── get-commands.md          #   GET команды
│   ├── sync-commands.md         #   SYNC команды
│   ├── installation.md          #   Установка
│   ├── interactive-mode.md      #   Интерактивный режим
│   ├── progress.md              #   Прогресс-бары
│   └── other-commands.md        #   Прочие команды
├── embedded/                     # Встроенные утилиты
│   └── jq_embed.go             #   Встроенный jq
├── internal/                     # Внутренний код
│   ├── client/                  #   HTTP клиент TestRail API
│   │   ├── client.go           #     HTTPClient (DoRequest + http.NewRequestWithContext)
│   │   ├── interfaces.go       #     ClientInterface (14 интерфейсов, 106 endpoints)
│   │   ├── mock.go             #     MockClient для тестов
│   │   ├── paginator.go        #     fetchAllPages[T] — generic-пагинатор
│   │   ├── accessor.go         #     ClientAccessor (lazy init)
│   │   └── <domain>.go         #     Endpoints по доменам
│   ├── concurrent/             #   Параллельная обработка (низкоуровневая)
│   │   ├── pool.go            #     WorkerPool, ParallelMap
│   │   ├── limiter.go         #     RateLimiter (150 req/min)
│   │   ├── retry.go           #     Retry с backoff
│   │   └── circuit.go         #     CircuitBreaker
│   ├── concurrency/            #   Стратегии параллелизации (высокоуровневая)
│   │   ├── types.go           #     ProgressReporter, FetchOption
│   │   ├── fetch_parallel.go  #     FetchParallel[T]
│   │   ├── fetch_by_suite.go  #     FetchParallelBySuite[T]
│   │   ├── controller.go      #     ParallelController (pipeline)
│   │   ├── priority_queue.go  #     PriorityQueue (heap)
│   │   └── aggregator.go      #     ResultAggregator
│   ├── ui/                     #   Унифицированный вывод (Stage 8.0)
│   │   ├── display.go         #     ANSI live display + Task reporter
│   │   ├── table.go           #     Table(), JSON(), NewTable(), --format
│   │   └── helpers.go         #     Info, Success, Warning, Error, Phase, Preview
│   ├── flags/                  #   Валидация флагов и аргументов (Stage 8.0)
│   │   └── helpers.go         #     ValidateRequiredID, GetFlag*, ParseID
│   ├── output/                 #   Сохранение результатов
│   │   ├── save.go            #     OutputResult, AddFlag, SaveToFile
│   │   ├── dryrun.go          #     DryRunPrinter
│   │   ├── filename.go        #     GenerateTimestamp, BuildFilename
│   │   └── paths.go           #     GetExportsDir, EnsureDir
│   ├── interactive/            #   Интерактивный выбор
│   │   ├── interactive.go     #     SelectProject, SelectSuite, SelectRun
│   │   └── wizard.go          #     InteractiveWizard
│   ├── paths/                  #   Управление путями
│   │   └── paths.go           #     BaseDir, ConfigFile, EnsureAllDirs
│   ├── log/                    #   Структурное логирование (zap)
│   │   └── logger.go          #     Init, L(), WithField, WithFields
│   ├── selftest/               #   Самодиагностика
│   │   ├── types.go           #     CheckResult, Report
│   │   └── checks.go          #     ConfigChecker, BaseDirChecker, ...
│   ├── service/                #   Бизнес-логика
│   │   ├── run.go              #     RunService
│   │   ├── result.go           #     ResultService
│   │   └── migration/          #     Миграция данных
│   │       ├── types.go       #       Migration struct
│   │       ├── fetch.go       #       Получение данных
│   │       ├── filter.go      #       Фильтрация
│   │       ├── import.go      #       Импорт
│   │       ├── export.go      #       Экспорт
│   │       ├── mapping.go     #       Mapping ID
│   │       └── log.go         #       Логирование
│   ├── models/                 #   Модели данных
│   │   └── data/              #     DTO для TestRail API
│   └── utils/                  #   Утилиты (legacy, сокращается)
│       ├── helpers.go         #     ParseID, SaveToFile, LoadMapping
│       └── log.go             #     LogDir
├── pkg/                          # Публичные пакеты
│   ├── reporter/               #   Builder-pattern репортер (Section/Stat/Print)
│   │   └── reporter.go
│   └── testrailapi/            #   Описания API endpoints
│       └── api_paths.go
├── main.go                       # Точка входа (signal.NotifyContext + ExecuteContext)
├── go.mod                        # Go модули
├── Makefile                     # Сборка
├── CHANGELOG.md                 # История изменений
└── README.md                    # Основная документация
```

## Доступные команды

### Получение данных (`gotr get`)
| Команда | Описание |
|---------|----------|
| `gotr get projects` | Все проекты |
| `gotr get project <id>` | Конкретный проект |
| `gotr get suites [project-id]` | Сьюты проекта |
| `gotr get suite <id>` | Конкретный сьют |
| `gotr get cases [project-id]` | Кейсы (интерактивный выбор сьюта) |
| `gotr get case <id>` | Конкретный кейс |
| `gotr get sharedsteps <project-id>` | Shared steps |
| `gotr get sections <project-id>` | Секции |

### Управление test runs (`gotr run`)
| Команда | Описание |
|---------|----------|
| `gotr run get <id>` | Получить информацию о run |
| `gotr run list [project-id]` | Список runs (интерактивный выбор) |
| `gotr run create <project-id>` | Создать run |
| `gotr run update <id>` | Обновить run |
| `gotr run close <id>` | Закрыть run |
| `gotr run delete <id>` | Удалить run |

### Управление результатами (`gotr result`)
| Команда | Описание |
|---------|----------|
| `gotr result get <test-id>` | Получить результаты test |
| `gotr result get-case <run-id> <case-id>` | Получить результаты case |
| `gotr result list [--run-id <id>]` | Список результатов (интерактивно) |
| `gotr result add <test-id>` | Добавить результат |
| `gotr result add-case <run-id>` | Добавить результат для case |
| `gotr result add-bulk <run-id>` | Массовое добавление из файла |

### Сравнение проектов (`gotr compare`)
| Команда | Описание |
|---------|----------|
| `gotr compare cases --pid1 X --pid2 Y` | Сравнение кейсов (параллельно) |
| `gotr compare suites --pid1 X --pid2 Y` | Сравнение сьютов |
| `gotr compare sections --pid1 X --pid2 Y` | Сравнение секций |
| `gotr compare runs --pid1 X --pid2 Y` | Сравнение ранов |
| `gotr compare all --pid1 X --pid2 Y` | Полное сравнение (13 ресурсов) |

### Миграция данных (`gotr sync`)
| Команда | Описание |
|---------|----------|
| `gotr sync full` | Полная миграция (shared-steps + cases) |
| `gotr sync cases` | Миграция кейсов |
| `gotr sync shared-steps` | Миграция shared steps |
| `gotr sync suites` | Миграция suites |
| `gotr sync sections` | Миграция sections |

### Прочие команды
| Команда | Описание |
|---------|----------|
| `gotr add <endpoint>` | POST запросы к API |
| `gotr update <endpoint>` | POST/PATCH запросы |
| `gotr delete <endpoint>` | DELETE запросы |
| `gotr list <resource>` | Список API endpoints |
| `gotr export <resource>` | Экспорт данных (JSON) |
| `gotr config init` | Инициализация конфигурации |
| `gotr selftest` | Самодиагностика |
| `gotr completion {bash\|zsh\|fish}` | Автодополнение |

### Глобальные флаги

| Флаг | Короткий | Тип | Описание |
|------|----------|-----|----------|
| `--url` | — | string | URL TestRail инстанса |
| `--username` | `-u` | string | Email пользователя |
| `--api-key` | `-k` | string | API ключ |
| `--config` | `-c` | bool | Создать дефолтный конфиг |
| `--format` | `-f` | string | Формат: `table`, `json`, `csv`, `md`, `html` |
| `--quiet` | `-q` | bool | Тихий режим (CI/CD) |
| `--debug` | `-d` | bool | Отладочный вывод (скрытый) |
| `--insecure` | — | bool | Пропуск проверки TLS |
| `--save` | — | bool | Сохранить результат в файл |
| `--save-to` | — | string | Сохранить в конкретный путь |

## Почему такая архитектура

### Преимущества

1. **Чёткое разделение** — каждый слой знает только про свой уровень
2. **Тестируемость** — можно тестировать сервисы с MockClient без реальных HTTP запросов
3. **Расширяемость** — добавление новой команды не требует изменения client
4. **Переиспользование** — один сервис используется в разных командах
5. **Интерактивность** — единый механизм выбора в `internal/interactive/`

### Добавление retry

Если TestRail возвращает "rate limit", retry добавляется только в Service Layer:

```go
func (s *RunService) Get(id int64) (*data.Run, error) {
    var run *data.Run
    err := retry.Do(3, func() error {
        var err error
        run, err = s.client.GetRun(id)
        return err
    })
    return run, err
}
```

CLI команды не требуют изменений!

## Для разработчиков

Где вносить изменения:

| Задача | Локация |
|--------|---------|
| Новая команда | `cmd/<group>/*.go` |
| Новая валидация | `internal/service/*.go` |
| Новый API метод | `internal/client/*.go` + `interfaces.go` |
| Новая структура данных | `internal/models/data/*.go` |
| Интерактивный выбор | `internal/interactive/wizard.go` |
| Параллельная обработка (generic) | `internal/concurrent/*.go` |
| Стратегии конкурентности | `internal/concurrency/*.go` |
| Generic compare factory | `cmd/compare/simple.go` |
| Унифицированный вывод (compare) | `pkg/reporter/*.go` |
| Progress runtime | `internal/ui/runtime.go` |

Подробная техническая документация: `.github/copilot/instructions/`

### Параллельная обработка в командах

Для ускорения операций с множеством сьютов/кейсов используйте `internal/concurrent`:

```go
// Пример: параллельная загрузка кейсов
import "github.com/Korrnals/gotr/internal/concurrent"

func fetchAllCases(client ClientInterface, projectID int64) ([]Case, error) {
    suites, _ := client.GetSuites(projectID)
    
    // Параллельно загружаем кейсы из всех сьютов
    results, err := concurrent.ParallelMap(suites, 5,
        func(suite Suite, index int) ([]Case, error) {
            return client.GetCases(projectID, suite.ID, 0)
        })
    
    // Собираем результаты
    var allCases []Case
    for _, r := range results {
        if r.Error == nil {
            allCases = append(allCases, r.Data...)
        }
    }
    return allCases, err
}
```

Rate limiting (150 req/min) и retry применяются автоматически.
