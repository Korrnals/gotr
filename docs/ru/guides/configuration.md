# Конфигурация

Language: Русский | [English](../../en/guides/configuration.md)

## Навигация

- [Документация](../index.md)
  - [Гайды](index.md)
    - [Установка](installation.md)
    - [Конфигурация](configuration.md)
    - [Интерактивный режим](interactive-mode.md)
    - [Прогресс](progress.md)
    - [Каталог команд](commands/index.md)
  - [Архитектура](../architecture/index.md)
  - [Эксплуатация](../operations/index.md)
  - [Отчёты](../reports/index.md)
- [Главная](../../../README_ru.md)

## Способы конфигурации

Приоритет (от высшего к низшему):

1. **Флаги командной строки** (`--url`, `--username`, `--api-key`)
2. **Переменные окружения** (`TESTRAIL_BASE_URL`, `TESTRAIL_USERNAME`, `TESTRAIL_API_KEY`)
3. **Конфигурационный файл** (`~/.gotr/config/default.yaml`)

## Переменные окружения

```bash
export TESTRAIL_BASE_URL="https://testrail.example.com"
export TESTRAIL_USERNAME="user@example.com"
export TESTRAIL_API_KEY="your_api_key"
```

## Конфигурационный файл

### Создание конфига

```bash
# Создать дефолтный конфиг
gotr config init

# Показать путь к конфигу
gotr config path

# Просмотреть содержимое
gotr config view

# Редактировать
gotr config edit
```

Примечания безопасности:
- gotr config init создает файл с правами 0600 (чтение/запись только для владельца).
- gotr config view маскирует чувствительные ключи api_key, password, token, authorization как "***".

### Структура config.yaml

```yaml
base_url: "https://testrail.example.com"
username: "user@example.com"
api_key: "your_api_key"

# Опциональные параметры
insecure: false      # пропустить проверку TLS
jq_format: false     # включить jq-форматирование по умолчанию
debug: false         # отладочный вывод

# Настройки compare-команд (performance tuning)
compare:
  # Авто-определение окружения: auto | cloud | server
  deployment: "auto"
  
  # Для cloud: professional | enterprise
  cloud_tier: "professional"
  
  # Лимит запросов/мин: -1 = авто по профилю, 0 = без лимита, >0 = фиксированный
  rate_limit: -1
  
  cases:
    parallel_suites: 10    # Одновременно загружаемых сьютов
    parallel_pages: 6      # Одновременно загружаемых страниц внутри сьюта
    page_retries: 5        # Retry для каждой страницы в основном этапе
    timeout: "30m"         # Таймаут полной операции
    auto_retry_failed_pages: true  # Авто-дозабор проблемных страниц
    
    retry:
      attempts: 5          # Попытки на страницу при точечном ретрае
      workers: 12          # Параллельных воркеров для дозабора
      delay: "200ms"       # Пауза между попытками
```

### Расположение файла

Поиск конфига в порядке приоритета:

1. `~/.gotr/config/default.yaml`
2. `./config.yaml` (текущая директория)

## Флаги глобальные

| Флаг | Описание | Переменная окружения |
|------|----------|---------------------|
| `--url` | URL TestRail | `TESTRAIL_BASE_URL` |
| `-u, --username` | Email пользователя | `TESTRAIL_USERNAME` |
| `-k, --api-key` | API ключ | `TESTRAIL_API_KEY` |
| `-i, --insecure` | Пропустить проверку TLS | - |
| `-d, --debug` | Отладочный вывод | `TESTRAIL_DEBUG` |

## Флаги compare

Эти флаги доступны для всех подкоманд `compare` (`cases`, `all`, `retry-failed-pages` и др.):

| Флаг | Описание | Конфиг-ключ |
|------|----------|-------------|
| `--rate-limit` | Лимит запросов/мин (-1=авто, 0=без лимита) | `compare.rate_limit` |
| `--parallel-suites` | Параллельных сьютов (default 10) | `compare.cases.parallel_suites` |
| `--parallel-pages` | Параллельных страниц (default 6) | `compare.cases.parallel_pages` |
| `--page-retries` | Retry на страницу (default 5) | `compare.cases.page_retries` |
| `--timeout` | Таймаут операции (default 30m) | `compare.cases.timeout` |
| `--retry-attempts` | Попытки авто-ретрая (default 3) | `compare.cases.retry.attempts` |
| `--retry-workers` | Воркеров авто-ретрая (default 12) | `compare.cases.retry.workers` |
| `--retry-delay` | Пауза авто-ретрая (default 200ms) | `compare.cases.retry.delay` |

**Приоритет:** CLI-флаг > конфиг YAML > default.

## Примеры использования

```bash
# Через флаги
gotr get projects --url https://testrail.example.com --username user@example.com --api-key xxx

# Через env
gotr get projects

# Через конфиг
gotr get projects
```

---

← [Гайды](index.md) · [Документация](../index.md)
