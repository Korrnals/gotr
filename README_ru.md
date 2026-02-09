```
╔══════════════════════════════════════════════════════════╗
║                                                          ║
║     ██████╗  ██████╗ ████████╗██████╗                    ║
║    ██╔════╝ ██╔═══██╗╚══██╔══╝██╔══██╗                   ║
║    ██║  ███╗██║   ██║   ██║   ██████╔╝                   ║
║    ██║   ██║██║   ██║   ██║   ██╔══██╗                   ║
║    ╚██████╔╝╚██████╔╝   ██║   ██║  ██║                   ║
║     ╚═════╝  ╚═════╝    ╚═╝   ╚═╝  ╚═╝                   ║
║                                                          ║
║           CLI Client for TestRail API v2                 ║
║                                                          ║
╚══════════════════════════════════════════════════════════╝
```

# gotr — CLI-клиент для TestRail API

[English](README.md) | [Русский](README_ru.md)

[![Version](https://img.shields.io/badge/version-2.6.0--dev-blue.svg)](CHANGELOG.md)
[![Go Version](https://img.shields.io/badge/go-1.25.6-blue.svg)](go.mod)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

Профессиональный инструмент командной строки для работы с TestRail API v2. Разработан для QA-инженеров и специалистов по автоматизации тестирования, которым требуется эффективное управление данными, возможности миграции и бесшовная интеграция с CI/CD.

> **Актуальная версия: v2.6.0-dev** — Этап 4 завершён: 106/106 API endpoint'ов реализовано. См. [CHANGELOG](CHANGELOG.md)

## Обзор

`gotr` предоставляет комплексный инструментарий для работы с TestRail:

- **Операции с данными** — Получение и управление тест-кейсами, сьютами, секциями, shared steps, ранами, результатами, майлстоунами, планами и др.
- **Полное покрытие API** — Все 106 endpoint'ов TestRail API v2 реализованы (Этап 4 завершён)
- **Синхронизация проектов** — Миграция сущностей между проектами с интеллектуальным обнаружением дубликатов
- **Интерактивный режим** — Пошаговый выбор проектов и сьютов без необходимости запоминать ID
- **Встроенная обработка** — Фильтрация JSON через встроенный `jq`, отслеживание прогресса и структурированное логирование
- **Гибкая конфигурация** — Поддержка флагов, переменных окружения и конфигурационных файлов

## Быстрый старт

```bash
# Установка (Linux/macOS)
curl -sL https://github.com/Korrnals/gotr/releases/latest/download/gotr-$(uname -s | tr '[:upper:]' '[:lower:]')-amd64 -o gotr
chmod +x gotr && sudo mv gotr /usr/local/bin/

# Инициализация конфигурации
gotr config init

# Проверка установки
gotr self-test
```

## Ключевые возможности

| Возможность | Описание |
|-------------|----------|
| **Полное покрытие API** | 106/106 эндпоинтов TestRail API v2 реализовано |
| **Интерактивный режим** | Визуальный выбор проектов, сьютов и целей миграции |
| **Синхронизация данных** | Миграция кейсов, shared steps, сьютов и секций между проектами |
| **Управление ранами** | Создание ранов, добавление результатов, отслеживание выполнения |
| **Встроенный jq** | Фильтрация и трансформация JSON без внешних зависимостей |
| **Индикаторы прогресса** | Визуальная обратная связь для длительных операций |
| **Автодополнение** | Поддержка bash, zsh и fish |
| **Детальное логирование** | Структурированные JSON-логи для аудита и отладки |

## Примеры использования

### Интерактивный режим

```bash
# Получение кейсов с интерактивным выбором проекта/сьюта
gotr get cases

# Синхронизация с пошаговым мастером
gotr sync full
```

### Получение данных

```bash
# Список всех проектов
gotr get projects

# Получение кейсов из конкретного проекта и сьюта
gotr get cases 30 --suite-id 20069

# Получение кейсов из всех сьютов проекта
gotr get cases 30 --all-suites

# Получение shared steps
gotr get sharedsteps 30
```

### Синхронизация

```bash
# Полная миграция (shared steps + кейсы)
gotr sync full \
  --src-project 30 --src-suite 20069 \
  --dst-project 31 --dst-suite 19859 \
  --approve --save-mapping

# Только shared steps
gotr sync shared-steps \
  --src-project 30 --dst-project 31 \
  --approve --save-mapping

# Кейсы с существующим маппингом
gotr sync cases \
  --src-project 30 --src-suite 20069 \
  --dst-project 31 --dst-suite 19859 \
  --mapping-file mapping.json --approve
```

### Тест-раны и результаты

```bash
# Создание тест-рана
gotr run add 30 --name "Regression Suite" --case-ids "1,2,3,4,5"

# Добавление результата теста
gotr result add 12345 --status-id 1 --comment "Test passed"

# Список результатов тестирования
gotr result list --run-id 100
```

### JSON-фильтрация

```bash
# Извлечение конкретных полей
gotr get projects --jq --jq-filter '.[] | {id: .id, name: .name}'

# Форматированный вывод с jq
gotr get case 12345 --jq
```

## Конфигурация

Приоритет конфигурации (от высшего к низшему):

1. **Флаги командной строки** (`--url`, `--username`, `--api-key`)
2. **Переменные окружения** (`TESTRAIL_BASE_URL`, `TESTRAIL_USERNAME`, `TESTRAIL_API_KEY`)
3. **Конфигурационный файл** (`~/.gotr/config/default.yaml`)

```bash
# Инициализация конфигурации
gotr config init

# Просмотр текущей конфигурации
gotr config view
```

## Документация

- [Руководство по установке](docs/installation.md)
- [Конфигурация](docs/configuration.md)
- [Команды GET](docs/get-commands.md)
- [Команды SYNC](docs/sync-commands.md)
- [Интерактивный режим](docs/interactive-mode.md)

## Структура проекта

```
gotr/
├── cmd/              # CLI команды (get, sync, run, result)
├── docs/             # Документация
├── internal/
│   ├── client/       # Клиент TestRail API
│   ├── service/      # Бизнес-логика (миграция и др.)
│   └── utils/        # Утилиты
├── pkg/              # Публичные пакеты
└── main.go           # Точка входа
```

## Что нового в v2.6.0-dev (Этап 4 завершён)

### Полное покрытие API

Реализованы все 106 endpoint'ов TestRail API v2:

- **Tests API** (3 endpoint'а) — GetTest, GetTests, UpdateTest
- **Milestones API** (5 endpoint'ов) — Полный CRUD для майлстоунов
- **Plans API** (9 endpoint'ов) — Полный CRUD + управление plan entries
- **Attachments API** (5 endpoint'ов) — Загрузка файлов к кейсам/планам/результатам/ранам
- **Configurations API** (7 endpoint'ов) — Управление конфигурациями тестов
- **Users & Reference** (7 endpoint'ов) — Пользователи, приоритеты, статусы, шаблоны
- **Reports API** (3 endpoint'а) — Генерация и получение отчётов
- **Extended APIs** (21 endpoint) — Группы, роли, датасеты, переменные, BDD, лейблы

## Что нового в v2.5.0

### Архитектурные улучшения
- **Унифицированный клиентский интерфейс** — Единый `ClientInterface` во всех пакетах устраняет дублирование кода
- **Расширенное покрытие тестами** — Все sync-тесты теперь используют мокирование на основе интерфейсов (10 новых тестов, 0 пропущенных)
- **Рефакторинг общего пакета** — Устранено дублирование `getClientSafe` в пакетах команд

### Интерактивные возможности
- **Интерактивный выбор** — Визуальные пикеры для проектов и сьютов в `run list` и `result list`
- **Упрощенный workflow** — Снижение трения для частых операций

Подробная история изменений: [CHANGELOG](CHANGELOG.md)

## Установка

Детальные инструкции по установке: [docs/installation.md](docs/installation.md)

## Участие в проекте

Приветствуются issues и pull requests.

## Используемые библиотеки

| Библиотека | Назначение |
|------------|------------|
| [spf13/cobra](https://github.com/spf13/cobra) | CLI-фреймворк |
| [spf13/viper](https://github.com/spf13/viper) | Управление конфигурацией |
| [cheggaaa/pb/v3](https://github.com/cheggaaa/pb) | Прогресс-бары |
| [go.uber.org/zap](https://github.com/uber-go/zap) | Структурированное логирование |
| [stretchr/testify](https://github.com/stretchr/testify) | Тестирование |
| [itchyny/gojq](https://github.com/itchyny/gojq) | Встроенный JSON-процессор |

## Лицензия

MIT License — см. [LICENSE](LICENSE)
