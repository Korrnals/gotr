# Changelog

Все заметные изменения в проекте `gotr` будут документироваться в этом файле.

Формат основан на [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
и проект использует [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased] - 2026-01-24

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
- Явные и информативные подсказки в `Short` и `Long` для всех подкома# Changelog

- Проверка обязательных параметров в `RunE` с понятными сообщениями об ошибках (например, про suite_id для cases).
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

- Поддержка конфигурационного файла `~/.gotr/config.yaml` с автоматическим чтением (Viper).
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
