# Changelog

Все заметные изменения в проекте `gotr` будут документироваться в этом файле.

Формат основан на [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
и проект использует [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

## [1.0.0] - 2025-12-19 (предыдущий релиз)

- Базовая версия с командами `list`, `get`, `add` и т.д.
- Поддержка TestRail API v2 через HTTP-клиент.
- Глобальные флаги `--url`, `--username`, `--api-key`.
