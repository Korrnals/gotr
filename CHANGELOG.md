# Changelog

Все заметные изменения в проекте `gotr` будут документироваться в этом файле.

Формат основан на [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
и проект использует [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added

#### Stage 5: CLI Test Coverage (986 tests, 10 packages at 100%)

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
  - `gotr compare all` - compare all resources at once
- **BREAKING CHANGE: `--save` flag replaces `--output` across ALL commands:**
  - `--save` is now a boolean flag (no value required)
  - Saves to `~/.gotr/exports/{resource}/{resource}_YYYY-MM-DD_HH-MM-SS.{format}`
  - Supports JSON, YAML, and CSV formats via `--format` flag (where applicable)
  - Affected commands: `get`, `export`, `users list`, `labels list`, `reports list-cross-project`, 
    `test get/list`, `tests list`, `groups add/update`, and all `compare` subcommands
- **Field-based comparison** for cases: `--field title`, `--field priority_id`, etc.
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
  - Нормализация тега: поддержка `VERSION=v2.6.0` и `VERSION=2.6.0`

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

- **Системная документация** `.systems/ARCHITECTURE.md` (660 строк):
  - Полное описание 4 слоёв архитектуры
  - Таблицы разделения ответственности
  - Полный перечень компонентов (22 команды, 3 сервиса, 40+ API методов)
  - Примеры рефакторинга
- **Пользовательская документация** `docs/architecture.md` (243 строки):
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

- Создана директория `.systems/` для файлов разработки
- Директория `.systems/` добавлена в `.gitignore`
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
  - `installation.md` — установка
  - `configuration.md` — конфигурация
  - `get-commands.md` — команды получения данных
  - `sync-commands.md` — команды синхронизации
  - `interactive-mode.md` — интерактивный режим
  - `other-commands.md` — другие команды

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
