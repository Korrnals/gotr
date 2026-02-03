# Конфигурация

## Способы конфигурации

Приоритет (от высшего к низшему):

1. **Флаги командной строки** (`--url`, `--username`, `--api-key`)
2. **Переменные окружения** (`TESTRAIL_BASE_URL`, `TESTRAIL_USERNAME`, `TESTRAIL_API_KEY`)
3. **Конфигурационный файл** (`~/.gotr/config.yaml`)

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

### Структура config.yaml

```yaml
base_url: "https://testrail.example.com"
username: "user@example.com"
api_key: "your_api_key"

# Опциональные параметры
insecure: false      # пропустить проверку TLS
jq_format: false     # включить jq-форматирование по умолчанию
debug: false         # отладочный вывод
```

### Расположение файла

Поиск конфига в порядке приоритета:

1. `~/.gotr/config.yaml`
2. `./config.yaml` (текущая директория)

## Флаги глобальные

| Флаг | Описание | Переменная окружения |
|------|----------|---------------------|
| `--url` | URL TestRail | `TESTRAIL_BASE_URL` |
| `-u, --username` | Email пользователя | `TESTRAIL_USERNAME` |
| `-k, --api-key` | API ключ | `TESTRAIL_API_KEY` |
| `-i, --insecure` | Пропустить проверку TLS | - |
| `-d, --debug` | Отладочный вывод | `TESTRAIL_DEBUG` |

## Примеры использования

```bash
# Через флаги
gotr get projects --url https://testrail.example.com --username user@example.com --api-key xxx

# Через env
gotr get projects

# Через конфиг
gotr get projects
```
