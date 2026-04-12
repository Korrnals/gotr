# Changelog

Все заметные изменения в проекте `gotr` будут документироваться в этом файле.

Формат основан на [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
и проект использует [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

---

## [3.0.1] - 2026-04-12

### Fixed

- Removed dead `--soft` flag from `gotr delete` (was declared but never used).
- Removed misleading milestone/plan/entry references from `gotr add` and `gotr update` help text.
- Fixed `--save-filtered` flag: wired into `sync full` command to actually save filtered shared steps list after migration.
- Fixed self-test table alignment for dynamic content widths.
- Fixed spinner first-frame delay in `ui.RunWithStatus` (immediate render before ticker).

### Added

- Progress spinners for all API-calling commands (~48 commands across cases, attachments, configurations, milestones, plans, groups, labels, run, result, tests, reports, export).
- Spinner wrapper for `gotr self-test` execution.

### Changed

- Documentation: added missing compare flags (`--include-refs`, `--include-custom-statuses`, `--include-custom-steps`, `--include-updated-by`, `--include-details`) to EN/RU guides.
- Documentation: corrected CRUD instruction redirects for milestones/plans.
- Documentation: fixed artifact filenames in migration-shared-steps guide.

---

## [3.0.0] - 2026-04-09

### Added

#### Stage 13.5: Quality Hardening & Full Audit

- **`api_paths.go`** — +14 endpoints added to the endpoint registry, complete coverage of TestRail API v2.
- **`attachments list --for-project`** — new subcommand wrapping `GetAttachmentsForProject()`.
- **`bdds add`** — stdin reading support: `cat scenario.feature | gotr bdds add 12345`.
- **`sync shared-steps --save-filtered`** — automatic/interactive saving of filtered shared steps list via `ExportSharedSteps()`.
- **Generic CRUD executor** (`internal/crud`) — eliminates boilerplate for simple CRUD commands.
- **Compare resource registry** (`cmd/compare`) — declarative resource registration replacing manual wiring.

### Changed

- `compare all`: stage-by-stage progress tracker in terminal (`done/active/pending`) for all resources.
- `compare all`: shared suites prefetch for `cases/suites/sections` to avoid repeated `get_suites` calls.
- `compare all`: resource failures are now marked as `PARTIAL` (instead of misleading `INTERRUPTED`).
- `compare all`: unsupported TestRail endpoints (`404 Unknown method`) are shown as `UNSUPPORTED` with a dedicated `Unsupported endpoints` summary block.
- `compare all` JSON/YAML meta now distinguishes real errors from unsupported endpoints:
  - `error_summary_count` / `error_resources` for real failures
  - `unsupported_summary_count` / `unsupported_resources` for server-unsupported methods
- Legacy `internal/progress` package removed; progress/status flow is unified via `internal/ui` runtime.
- All Russian text in Go source files translated to English (i18n pass: 1738+ lines across 2 passes).
- `panic(err)` in `main.go` and `cmd/commands.go` replaced with `fmt.Fprintf(os.Stderr)` + `os.Exit(1)`.
- `ClientInterface` unified across all service packages (B-2..B-4 audit remediation).

### Fixed

- `internal/client` paginator: fixed potential infinite loop for flat-array API responses with page size at or above 250.
- `compare sections`: stabilized loading path via client pagination behavior and added regression coverage in paginator tests.
- All `io.ReadAll(resp.Body)` calls wrapped with `io.LimitReader` (10 MiB cap) to prevent unbounded memory allocation.
- File descriptor leak in `migration/types.go` — `logFile` now properly closed in `Migration.Close()`.
- `os.Remove` error paths in `embedded/jq_embed.go` now checked and logged.
- All `json.Marshal` errors across the codebase handled (45+ fixes in 17 files).
- Safe type assertions with comma-ok pattern throughout; `os.Getwd` errors properly handled.
- Context propagation ensured across all API calls (F-2..F-7 audit findings).

### Security

- Bounded parallelism enforced in all concurrent operations.
- All HTTP response body reads are size-limited.
- `ReadResponse` documentation clarifies `resp.Body` ownership contract.

### Quality

- **golangci-lint**: 290 findings → **0** (errcheck, staticcheck, gocritic, gocyclo, misspell, unused, nolintlint, ineffassign).
- **Test suite**: 43/43 packages pass with race detector, 0 data races.
- **0 TODO/FIXME/HACK** markers in production code.
- **Audit verdict**: UNCONDITIONAL PASS (7 audit rounds completed).

## [3.0.0] - 2026-03-12

### Added

#### Stage 6.8: Concurrency Unification & Compare Subcommands

- **`internal/concurrency/`** — новый unified concurrency-пакет (переименован из `internal/parallel/`)
  - Три уровня стратегий:
    - `FetchParallel[T]` — лёгкая: один API-вызов на проект, параллельная загрузка P1+P2
    - `FetchParallelBySuite[T]` — средняя: per-suite запросы (для `sections`)
    - `FetchParallelPaginated` — тяжёлая: `ParallelController` с пагинацией (для `cases`)
  - Generic API через Go generics `[T any]`

- **`pkg/reporter/`** — универсальный reporter вынесен в публичный пакет (из `internal/ui/reporter/`)
  - Builder pattern: `Section` / `Stat` / `StatIf` / `StatFmt` / `Print`
  - go-pretty/v6 для выровненного boxed-вывода

- **Generic `newSimpleCompareCmd`** — одна generic-фабрика вместо 9 идентичных файлов (`cmd/compare/simple.go`)
  - Устранено ~1200 строк копипасты
  - Все простые подкоманды используют `FetchParallel[T]` для параллельной загрузки проектов

- **`compare sections`** — параллельная загрузка секций по сьютам через `FetchParallelBySuite[T]`

- **`compare all`** — единообразный вывод через `pkg/reporter`, partial results при недоступных API

### Changed

- `internal/parallel/` → `internal/concurrency/` (переименование пакета и всех импортов)
- `internal/ui/reporter/` → `pkg/reporter/` (вынесен в публичный пакет)
- Все 13 compare-подкоманд используют `pkg/reporter` для вывода статистики
- `OnSuiteComplete` → `OnItemComplete` в интерфейсе `ProgressReporter`
- Дефолтные значения: `parallel-suites=10`, `parallel-pages=6` (стабильные для TestRail Server)

### Fixed

- `compare all` более не использует `fmt.Println` с emoji и box-drawing символами
- Устранено некорректное выравнивание статистики в терминалах без поддержки emoji

### Performance

- Простые compare-подкоманды (runs, plans, milestones и др.): загрузка P1 и P2 **параллельно**
- `compare sections`: параллельная загрузка по сьютам вместо последовательной

---

#### Stage 6.9: Generic Paginator & Pagination Audit

### Added

- **`internal/client/paginator.go`** — универсальный generic-пагинатор `fetchAllPages[T]`
  - Обрабатывает оба формата TestRail API без ветвлений в бизнес-логике:
    - **Paginated wrapper** (TestRail 6.7+): `{"offset":0,"limit":250,"size":N,"<key>":[...]}`
    - **Flat array** (старые TestRail Server): `[item1, item2, ...]`
  - Автоматическое определение формата по первому байту ответа
  - Стандартный размер страницы: 250 элементов (TestRail default)
  - Выход по условию: `len(page) < limit` (последняя страница)

- **Миграция 9 критичных list-методов** на `fetchAllPages[T]`:
  - `GetRuns(projectID)` — runs теперь не обрезаются на 250
  - `GetPlans(projectID)` — планы теперь не обрезаются на 250
  - `GetSections(projectID, suiteID)` — секции (критично для `compare sections`)
  - `GetSharedSteps(projectID)` — shared steps
  - `GetMilestones(projectID)` — milestones
  - `GetResults(runID)` — результаты рана
  - `GetResultsForRun(runID)` — расширенный вариант
  - `GetTests(runID)` — тесты рана
  - `GetSuites(projectID)` — сьюты проекта

### Changed

- Все 9 мигрированных методов: тело метода упрощено с ~30 строк ручного цикла до 1 вызова `fetchAllPages`
- Удалено ~145 строк дублированного pagination boilerplate из `internal/client/`

### Tests

- `internal/client/paginator_test.go` — 11 новых unit-тестов:
  - Оба формата ответа (paginated wrapper и flat array)
  - Многостраничная загрузка (multi-page accumulation)
  - Граничные случаи: пустой ответ, последняя неполная страница
  - Тест на ошибку сервера (HTTP 500)
  - Table-driven tests для `decodeListResponse`

### Verified

- `compare all --pid1 30 --pid2 34`: 20 509 кейсов (87 стр.) + 116 009 кейсов (475 стр.) — пагинация подтверждена на реальных данных
- `compare runs`, `compare plans`, `compare milestones`, `compare sections`, `compare sharedsteps`: все работают корректно
- `go test ./...` — все тесты зелёные

---

#### Stage 7.0: Context Propagation

### Added

- **`context.Context`** во все ~100 методов `ClientInterface`
  - `signal.NotifyContext` → корректное завершение по Ctrl+C
  - Контекст пробрасывается CLI → Service → Client → HTTP

### Changed

- Все API-методы принимают `ctx context.Context` первым аргументом
- `cmd.ExecuteContext()` вместо `cmd.Execute()`
- `MockClient` обновлён под новые сигнатуры

---

#### Stage 8.0: UI/Output Refactoring

### Added

- **`internal/ui/`** — универсальные хелперы:
  - `ui.Table(headers, rows)` — обёртка над go-pretty вместо tabwriter
  - `ui.JSON(v)` — форматированный JSON-вывод
  - `ui.Success()`, `ui.Warn()`, `ui.Error()`, `ui.Info()` — цветные сообщения
  - `ui.Print()`, `ui.Printf()`, `ui.Println()` — обёртки стандартного вывода
- **`--format` PersistentFlag** — глобальный флаг формата вывода на root-уровне
- Массовая миграция: `tabwriter` → `ui.Table`, `json.MarshalIndent` → `ui.JSON`, `fmt.Print*` → `ui.*` (49 файлов)

### Changed

- `internal/flags/`: `*Var` → `GetFlag`, `ValidateRequiredID`
- `os.Exit` → `panic` в `GetClient*` (тестируемость)
- Все error messages переведены на английский

---

## [2.7.0] - 2026-02-20

### Added

#### Stage 6: Performance Optimization & UX Enhancement (In Progress)

- **Universal Progress Monitoring**: Channel-based progress system
  - `internal/progress.Monitor` — decoupled from UI
  - Real-time updates via buffered channels
  - Thread-safe, non-blocking implementation
  - Works with any long-running operation
  - See [docs/guides/progress.md](docs/guides/progress.md) for details

- **Multi-Progress-Bars (mpb)**: Visual feedback for long-running operations
  - `github.com/vbauerster/mpb/v8` integration (multi-progress-bar library)
  - Multiple simultaneous progress bars on separate lines
  - Real-time updates for parallel project loading
  - ETA, speed, and percentage decorators
  
- **Parallel API Requests**: 60-80% performance improvement
  - Worker pool pattern for concurrent requests
  - Rate limiting (180 req/min — TestRail maximum)
  - Parallel fetching for cases, suites, shared steps
  - Integrated with progress monitoring
  - **Page-level progress**: GetCasesWithProgress updates after each 250 cases page
  
- **Compare Cases Command**: Full comparison with parallel loading
  - Two-phase progress: spinner → progress bars
  - Parallel loading of both projects simultaneously  
  - Project-level statistics (suites count, cases count, duration)
  - Analysis phase with timing
  - Debug mode support via `--debug` flag
  
- **Response Caching**: Disk-based cache with TTL
  - Cache location: `~/.gotr/cache/`
  - TTL: Projects 1h, Cases 15min, Suites 30min
  - `--no-cache` flag to bypass
  
- **Retry Logic**: Exponential backoff for resilience
  - Automatic retry on transient failures
  - Circuit breaker pattern
  - `--timeout` flag (default: 5min)
  
- **Batch Operations**: Optimized for large projects
  - Batch fetching (250 items per request)
  - Streaming output for large datasets
  - Memory optimization (<500MB peak)

### Changed

- **Progress Bar Library**: Migrated from `progressbar/v3` to `mpb/v8`
  - Better support for multiple simultaneous progress bars
  - Improved UX with parallel operations
  - New API: methods called directly on bar objects (`bar.Add()`, `bar.Finish()`)

#### Stage 5: CLI Test Coverage (986 tests, 10 packages at 100%) - COMPLETE

- **Comprehensive test suite** for all CLI commands:
  - `cmd/run` — 38 tests (95.2% coverage)
  - `cmd/result` — 46+ tests (90.6% coverage)
  - `cmd/get` — 71 tests (89.7% coverage)
  - `cmd/attachments` — 18 tests (100% coverage)
  - `cmd/labels` — 60+ tests (100% coverage)
  - `cmd/groups` — 26 tests (100% coverage)
  - `cmd/cases` — 87 tests (99.2% coverage)
  - `cmd/test` + `cmd/tests` — 35 tests (90.5%+ coverage)
  - `cmd/templates`, `cmd/reports`, `cmd/sync`, `cmd/milestones`, `cmd/plans`
  - `cmd/users`, `cmd/variables`, `cmd/configurations`, `cmd/datasets`
  - `cmd/bdds`, `cmd/roles` — all at 87-100% coverage
- **Test infrastructure:**
  - Shared `testhelper` package for mock client injection
  - `serviceWrapper` pattern for interface-based testing
  - Constructor pattern for all commands (`newXxxCmd`)
- **Total:** 110 test files, 986 test functions, 18 packages ≥ 90% coverage

#### Stage 5.2: Project Comparison Command + Unified --save Flag

- **New `cmd/compare/` package** with subcommand structure:
  - `gotr compare cases` - compare test cases with field-based diff
  - `gotr compare suites` - compare test suites
  - `gotr compare sections` - compare sections
  - `gotr compare sharedsteps` - compare shared steps
  - `gotr compare runs` - compare test runs
  - `gotr compare plans` - compare test plans
  - `gotr compare milestones` - compare milestones
  - `gotr compare datasets` - compare datasets
  - `gotr compare groups` - compare groups
  - `gotr compare labels` - compare labels
  - `gotr compare templates` - compare templates
  - `gotr compare configurations` - compare configurations
  - `gotr compare all` - compare all resources at once with formatted table output
- **Enhanced `--save` and new `--save-to` flags:**
  - `--save` - saves table output as text file to `~/.gotr/exports/{resource}/`
  - `--save-to <path>` - saves to specified path with format from `--format` or auto-detected from extension
  - Auto-detection: `.json` → JSON, `.yaml`/`.yml` → YAML, `.csv` → CSV, `.txt` → table
  - Supports JSON, YAML, CSV, and table (text) formats
  - Affects all `compare` subcommands
- **BREAKING CHANGE: `--save` flag replaces `--output` across ALL commands:**
  - `--save` is now a boolean flag (no value required)
  - Saves to `~/.gotr/exports/{resource}/{resource}_YYYY-MM-DD_HH-MM-SS.{format}`
  - Supports JSON, YAML, and CSV formats via `--format` flag (where applicable)
  - Affected commands: `get`, `export`, `users list`, `labels list`, `reports list-cross-project`,
    `test get/list`, `tests list`, `groups add/update`, and all `compare` subcommands
- **Field-based comparison** for cases: `--field title`, `--field priority_id`, etc.
- **Formatted table output** for `compare all`:
  - Unicode box-drawing characters for clean presentation
  - Status indicators: ✓ (perfect match), ⚠ (has differences), ✗ (error loading)
  - Compact summary showing counts per resource type
  - Error section for failed resource comparisons
- **Package structure:**
  - `types.go` - shared types (CompareResult, ItemInfo, CommonItemInfo)
  - `register.go` - command registration with root command
  - Individual files per resource (cases.go, suites.go, etc.)

#### Save Package (cmd/common/flags/save)

- **New package** for standardized output saving across all commands:
  - `SaveWithOptions()` - unified save function supporting JSON, YAML, CSV formats
  - `GenerateFilename()` - generates timestamped filenames: `{resource}_YYYY-MM-DD_HH-MM-SS.{ext}`
  - `GetExportsDir()` - returns `~/.gotr/exports/` directory path
  - Automatic directory creation with 0755 permissions
  - CSV export with dynamic header detection from struct tags
  - Over 40 comprehensive tests (100% coverage)

#### Build System Improvements

- **Автоматическая синхронизация версии в Makefile:**
  - Команда `make build` теперь извлекает версию из `cmd/root.go` (единый источник правды)
  - Для релизных версий (без `-dev`) автоматически создаётся/проверяется git tag
  - Приоритет версии: 1) `make build VERSION=x`, 2) версия из кода, 3) git tag
  - Нормализация тега: поддержка `VERSION=v2.7.0` и `VERSION=2.7.0`

#### Stage 4: Complete API Coverage (106/106 endpoints)

- **Attachments API** — 5 endpoints:
  - `AddAttachmentToCase`, `AddAttachmentToPlan`, `AddAttachmentToPlanEntry`
  - `AddAttachmentToResult`, `AddAttachmentToRun`
  - Поддержка multipart/form-data для загрузки файлов
- **Configurations API** — 7 endpoints:
  - `GetConfigs`, `AddConfigGroup`, `AddConfig`
  - `UpdateConfigGroup`, `UpdateConfig`, `DeleteConfigGroup`, `DeleteConfig`
- **Users API** — 4 endpoints:
  - `GetUsers`, `GetUser`, `GetUserByEmail`
- **Reference Data APIs** — 3 endpoints:
  - `GetPriorities`, `GetStatuses`, `GetTemplates`
- **Reports API** — 3 endpoints:
  - `GetReports`, `RunReport`, `RunCrossProjectReport`
- **Extended APIs** — 21 endpoints:
  - Groups: `GetGroups`, `GetGroup`, `AddGroup`, `UpdateGroup`, `DeleteGroup`
  - Roles: `GetRoles`, `GetRole`
  - ResultFields: `GetResultFields`
  - Datasets: `GetDatasets`, `GetDataset`, `AddDataset`, `UpdateDataset`, `DeleteDataset`
  - Variables: `GetVariables`, `AddVariable`, `UpdateVariable`, `DeleteVariable`
  - BDDs: `GetBDD`, `AddBDD`
  - Labels: `UpdateTestLabels`, `UpdateTestsLabels`

**Всего реализовано:** 44 новых endpoint'а  
**Общее покрытие:** 106/106 endpoint'ов TestRail API (100%)

### Added

#### Dry-run режим

- **Флаг** `--dry-run` — единый флаг для всех команд, изменяющих состояние:
  - `add` — project, suite, section, case, run, result, shared-step
  - `update` — project, suite, section, case, run, shared-step
  - `delete` — project, suite, section, case, run, shared-step
  - `run create/update/close/delete`
  - `result add/add-case/add-bulk`
- **Пакет** `cmd/common/dryrun/` — централизованное форматирование вывода dry-run

#### Интерактивный wizard mode

- **Флаг** `--interactive/-i` — интерактивный режим для команд:
  - `add` — project, suite, case, run
  - `update` — project, suite, case
- **Пакет** `cmd/common/wizard/` — библиотека интерактивных prompt'ов на survey/v2
- Паттерн: ввод → предпросмотр → подтверждение/отмена

### Changed

- **Флаг** `-i` теперь используется для `--interactive` (вместо `--insecure`)
- **Флаг** `--insecure` — только длинная форма (без shorthand)

---

## [2.5.0] - 2026-02-05

### Added

#### Интерактивный режим

- **Команда** `gotr run list` — интерактивный выбор проекта при отсутствии аргументов
- **Команда** `gotr result list` — интерактивный выбор проекта → test run
- **Пакет** `internal/interactive/` — единый механизм интерактивного выбора

#### Client Interface + Mock (Архитектурное улучшение)

- **Пакет** `internal/client/interfaces.go` — полный композитный интерфейс:
  - `ProjectsAPI` — 5 методов
  - `CasesAPI` — 14 методов
  - `SuitesAPI` — 5 методов
  - `SectionsAPI` — 5 методов
  - `SharedStepsAPI` — 6 методов
  - `RunsAPI` — 6 методов
  - `ResultsAPI` — 7 методов
- **Пакет** `internal/client/mock.go` — полный `MockClient` (43 метода)
- Проверка компиляции: `var _ ClientInterface = (*HTTPClient)(nil)`

#### Общие утилиты (Рефакторинг)

- **Пакет** `cmd/common/client.go` — `ClientAccessor` для единого доступа к HTTP клиенту
- **Пакет** `cmd/common/flags.go` — общие функции парсинга флагов
- Рефакторинг `cmd/result/`, `cmd/run/`, `cmd/sync/` — использование `common.ClientAccessor`
- Удалено дублирование `getClientSafe` из 3 пакетов

### Fixed

#### Унификация интерфейсов миграции

- **Удалён дублирующий пакет** `internal/migration` (оставлен `internal/service/migration`)
- **Унифицирован интерфейс** — `internal/service/migration` теперь использует `client.ClientInterface`
- **Обновлён `MockClient`** — дефолтные возвращаемые значения предотвращают nil pointer dereference
- **Рефакторинг sync тестов** — все 10 тестов переписаны с использованием `client.MockClient`
- Убраны пропуски тестов (`t.Skip`) — все тесты проходят

### Changed

- **README.md** — реструктурировано описание, acknowledgements перенесены в конец
- Версия обновлена до `2.5.0`

---

## [2.4.0] - 2026-02-04

### Added

#### Results API (Полная реализация)

- **Новый client** `internal/client/results.go` с методами:
  - `AddResult` — добавление результата для теста
  - `AddResultForCase` — добавление результата для кейса в run
  - `AddResults` — массовое добавление результатов (bulk)
  - `AddResultsForCases` — массовое добавление для кейсов (bulk)
  - `GetResults` — получение результатов для теста
  - `GetResultsForRun` — получение всех результатов run
  - `GetResultsForCase` — получение результатов для кейса в run

#### Runs API (Полная реализация)

- **Новый client** `internal/client/runs.go` с методами:
  - `GetRun` — получение информации о run
  - `GetRuns` — список runs проекта
  - `AddRun` — создание нового run
  - `UpdateRun` — обновление существующего run
  - `CloseRun` — закрытие run
  - `DeleteRun` — удаление run

#### CLI команды для Results

- **Новый пакет** `cmd/result/` с командами:
  - `gotr result get <test-id>` — получить результаты
  - `gotr result get-case <run-id> <case-id>` — получить результаты для кейса
  - `gotr result add <test-id>` — добавить результат
  - `gotr result add-case <run-id>` — добавить результат для кейса
  - `gotr result add-bulk <run-id>` — массовое добавление из JSON-файла

#### CLI команды для Runs

- **Новый пакет** `cmd/run/` с командами:
  - `gotr run get <run-id>` — получить информацию о run
  - `gotr run list <project-id>` — список runs проекта
  - `gotr run create <project-id>` — создать run
  - `gotr run update <run-id>` — обновить run
  - `gotr run close <run-id>` — закрыть run
  - `gotr run delete <run-id>` — удалить run

#### Service Layer (Архитектурное улучшение)

- **Новый пакет** `internal/service/`:
  - `RunService` — бизнес-логика для runs с валидацией
  - `ResultService` — бизнес-логика для results с валидацией
  - `internal/service/migration/` — перенесён из `internal/migration/`
- **Валидация** в сервисах:
  - Проверка ID > 0
  - Проверка обязательных полей (name, suite_id, status_id)
  - Валидация bulk-запросов (непустые массивы)
- **Утилиты** в `internal/utils/helpers.go`:
  - `ParseID` — парсинг ID
  - `OutputResult` — вывод результата (JSON + сохранение в файл)
  - `PrintSuccess` — вывод сообщений
  - `SaveToFile` — сохранение данных в JSON-файл

#### Архитектурная документация

- **Системная документация** `.github/copilot/instructions/`:
  - Полное описание 4 слоёв архитектуры
  - Таблицы разделения ответственности
  - Полный перечень компонентов (22 команды, 3 сервиса, 40+ API методов)
  - Примеры рефакторинга
- **Пользовательская документация** `docs/architecture/overview.md` (243 строки):
  - Упрощённое описание архитектуры
  - Примеры потоков данных
  - Полный список команд

#### Тесты

- **Тесты для Service Layer**:
  - `internal/service/run_test.go` — 6 тестов для валидации RunService
  - `internal/service/result_test.go` — 9 тестов для валидации ResultService

---

## [2.3.0] - 2026-02-03

### Added

#### Модели для Results и Runs API

- **Новые модели данных** в `internal/models/data/`:
  - `results.go` — модели `Result`, `AddResultRequest`, `AddResultsRequest`, `AddResultsForCasesRequest`
  - `runs.go` — модели `Run`, `AddRunRequest`, `UpdateRunRequest`, `CloseRunRequest`
  - `tests.go` — модели `Test`, `UpdateTestRequest`
  - `statuses.go` — модель `Status` с константами статусов
- Подготовка к реализации Results и Runs API

#### Исправления по результатам аудита

- **Обновлены request-структуры** в `cases.go`:
  - `AddCaseRequest` — добавлены поля `custom_steps` и `custom_expected` (текстовый формат)
  - `UpdateCaseRequest` — добавлены поля `type_id`, `suite_id`, `section_id`, `template_id` для перемещения кейсов
- **Исправлена модель `Section`** — добавлены `omitempty` к необязательным полям
- **Удалён дубликат метода** `AddCaseRequest` из `internal/client/cases.go`
- **Исправлен метод `GetSections`** — `suite_id` теперь передаётся как query-параметр

#### Системные изменения

- Системные файлы разработки вынесены в служебные инструкции `.github/copilot/instructions/`
- Внедрено осознанное версионирование (Semantic Versioning)

---

## [2.2.3] - 2026-02-03

### Added

#### Интерактивный режим

- **Интерактивный выбор** для всех команд `get` и `sync`:
  - `gotr get cases` — интерактивный выбор проекта и сьюта
  - `gotr get suites` — интерактивный выбор проекта
  - `gotr get sharedsteps` — интерактивный выбор проекта
  - `gotr sync cases` — интерактивный выбор source/destination проектов и сьютов
  - `gotr sync shared-steps` — интерактивный выбор проектов
  - `gotr sync sections` — интерактивный выбор проектов и сьютов
  - `gotr sync full` — интерактивный выбор для полной миграции
- **Автоматический выбор**: если в проекте один сьют — используется автоматически
- **Флаг `--all-suites`** для `gotr get cases` — получение кейсов из всех сьютов проекта

#### Реструктуризация кода

- Новая структура пакета `cmd/`:
  - `cmd/get/` — отдельный пакет для GET-команд
  - `cmd/sync/` — отдельный пакет для SYNC-команд
  - `cmd/commands.go` — централизованная регистрация всех команд
  - `cmd/interactive.go` — общие функции интерактивного выбора
- Dependency injection для избежания циклических зависимостей между пакетами

#### Документация

- Создана директория `docs/` с подробной документацией:
  - `guides/installation.md` — установка
  - `guides/configuration.md` — конфигурация
  - `guides/commands/get.md` — команды получения данных
  - `guides/commands/sync.md` — команды синхронизации
  - `guides/interactive-mode.md` — интерактивный режим
  - `guides/commands/other.md` — другие команды

### Changed

- Улучшена работа с `suite-id` в `gotr get cases`:
  - Убрано жёсткое требование флага `--suite-id`
  - Интерактивный выбор при отсутствии флага
  - Понятное сообщение об ошибке от API при отсутствии suite_id для multiple suites
- Обновлены `Long` описания всех sync-команд с описанием интерактивного режима
- Регистрация флагов перенесена из `init()` функций в `cmd/sync/sync.go`

### Fixed

- Исправлено дублирование флагов в `sync` командах
- Убраны неиспользуемые переменные в тестах

---

## [2.1.0] - 2026-01-24

### Added

- `gotr sync suites` — новая команда синхронизации suites: Fetch → Filter → Import.
- `gotr sync sections` — новая команда синхронизации sections.
- Общий хелпер `addSyncFlags()` для унификации флагов команд `sync/*`.
- Unit-тесты для `sync suites` и `sync sections`.

### Changed

- Команды `sync/*` переведены на единый поток миграции (internal/migration) и теперь используют централизованную логику Fetch → Filter → Import.
- Улучшены `Long` описания команд и добавлены русские комментарии-«Шаги» в коде команд для удобства русскоязычных пользователей.

### Testing

- В тестах используется отдельная папка логов: `.testrail/logs/test_runs`.
- Введён тестовый seam `sync_helpers.go` (переменная `newMigration`) для инъекции мок-миграций в тестах.

---

## [2.0.0] - 2026-01-15

### Breaking Changes

- Полная переработка команды `get`: переход на подкоманды вместо универсального подхода.
  - Теперь `gotr get <resource>` с подкомандами: `cases`, `case`, `projects`, `project`, `sharedsteps`, `sharedstep`, `sharedstep-history`, `suites`, `suite`.
  - Убраны старые универсальные вызовы (например, `gotr get get_cases 30`).
- Все ID теперь строго типизированы как `int64` в методах клиента и структурах (было string в некоторых местах).
- `get_cases` теперь требует `suite_id` (обязательно для проектов в режиме multiple suites).
- Изменена структура ответов для некоторых эндпоинтов (например, `GetProjectsResponse`, `GetSharedStepsResponse` стали срезами вместо объектов с полем).

### Added

- Новые подкоманды в группе `get`:
  - `gotr get case <case-id>` — получить один кейс по ID кейса.
  - `gotr get case-history <case-id>` — получить историю изменений кейса.
  - `gotr get sharedstep <step-id>` — получить один shared step по ID шага.
  - `gotr get sharedstep-history <step-id>` — получить историю изменений shared step.
  - `gotr get suites` — получить список тест-сюит проекта.
  - `gotr get suite <suite-id>` — получить одну тест-сюиту по ID.
- Поддержка **позиционных аргументов** для ID проекта в `cases`, `sharedsteps`, `suites`.
- Явные и информативные подсказки в `Short` и `Long` для всех подкоманд.
- Проверка обязательных параметров в `RunE` с понятными сообщениями об ошибках.
- Методы клиента для suites: `GetSuites`, `GetSuite`, `AddSuite`, `UpdateSuite`, `DeleteSuite`.

### Changed

- Улучшена обработка ошибок в клиенте: проверка StatusCode перед декодированием, информативные сообщения.
- Все ответы на список (projects, cases, shared steps, suites) возвращают срез напрямую (массив), а не объект с полем.
- Убраны лишние обёртки в структурах ответов (GetProjectResponse → Project, GetCaseResponse → Case и т.д.).
- Подсказки в `help` теперь максимально понятные: указывают, какой ID нужен и где его взять.

### Fixed

- Исправлено декодирование массивов из API (projects, shared steps, cases).
- Исправлена проблема с `MarkFlagRequired` — теперь позиционные аргументы работают без конфликта с обязательными флагами.
- Исправлено поле `is_deleted` в Case (теперь int, так как API возвращает 0/1).

---

## [2.0.0] - 2025-12-21

### Breaking Changes

- Изменён префикс переменных окружения с `GOTR_` на `TESTRAIL_` для лучшей совместимости с экосистемой TestRail (например, `TESTRAIL_BASE_URL`, `TESTRAIL_USERNAME`, `TESTRAIL_API_KEY`).
- Убраны старые ключи в конфиге и Viper (`testrail_base_url`, `testrail_username` и т.д.) — теперь используются `base_url`, `username`, `api_key`.

### Added

- Поддержка конфигурационного файла `~/.gotr/config/default.yaml` с автоматическим чтением (Viper).
- Новые подкоманды в группе `config`:
  - `gotr config init` — создание дефолтного конфига с комментариями.
  - `gotr config path` — показ пути к конфигу.
  - `gotr config view` — вывод содержимого конфига.
  - `gotr config edit` — открытие конфига в редакторе по умолчанию (`$EDITOR`).
- Автодополнение для bash (через `gotr completion bash`).
- Отключение обязательных проверок для служебных команд (`config`, `completion`).
- Условный вывод сообщений (без "Using config file" для чистоты stdout).
- Поддержка `insecure` в конфиге (для пропуска TLS-проверки).

### Changed

- Унифицированы ключи Viper: `base_url`, `username`, `api_key` (без `testrail_`).
- Улучшена обработка env-переменных с префиксом `TESTRAIL_`.

### Fixed

- Убрано дублирование сообщений "Using config file".
- Исправлено автодополнение (без мусора из файлов и вывода).

### Removed

- Старые env-переменные с префиксом `GOTR_`.

---

## [1.0.0] - 2025-12-19 (предыдущий релиз)

- Базовая версия с командами `list`, `get`, `add` и т.д.
- Поддержка TestRail API v2 через HTTP-клиент.
- Глобальные флаги `--url`, `--username`, `--api-key`.

