# gotr — CLI-клиент для TestRail API

```text
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

[English](README.md) | [Русский](README_ru.md)

[![Version](https://img.shields.io/badge/version-3.0.0-blue.svg)](CHANGELOG.md)
[![Go Version](https://img.shields.io/badge/go-1.25.0-blue.svg)](go.mod)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

Профессиональный инструмент командной строки для работы с TestRail API v2. Разработан для QA-инженеров и специалистов по автоматизации тестирования, которым требуется эффективное управление данными, возможности миграции и бесшовная интеграция с CI/CD.

> **Актуальная версия: v3.0.0** — Stage 13.5 завершён: quality hardening, 7 раундов аудита, 0 lint-findings, полное тестовое покрытие. См. [CHANGELOG](CHANGELOG.md)

## Обзор

`gotr` предоставляет комплексный инструментарий для работы с TestRail:

- **Операции с данными** — Получение и управление тест-кейсами, сьютами, секциями, shared steps, ранами, результатами, майлстоунами, планами и др.
- **Полное покрытие API** — Все 121 endpoint'ов TestRail API v2 реализованы (Этап 4 завершён)
- **Синхронизация проектов** — Миграция сущностей между проектами с интеллектуальным обнаружением дубликатов
- **Интерактивный режим** — Пошаговый выбор проектов и сьютов без необходимости запоминать ID
- **Встроенная обработка** — Фильтрация JSON через встроенный `jq`, отслеживание прогресса и структурированное логирование
- **Прогресс в реальном времени** — Визуальные прогресс-бары с обновлением через каналы для параллельных операций
- **Гибкая конфигурация** — Поддержка флагов, переменных окружения и конфигурационных файлов

## Навигация

- [Документация](docs/index.md)
  - [Гайды](docs/ru/guides/index.md)
    - [Установка](docs/ru/guides/installation.md)
    - [Конфигурация](docs/ru/guides/configuration.md)
    - [Интерактивный режим](docs/ru/guides/interactive-mode.md)
    - [Прогресс](docs/ru/guides/progress.md)
    - [Каталог команд](docs/ru/guides/commands/index.md)
      - [Группы команд](docs/ru/guides/commands/index.md#группы-команд-и-подгруппы)
  - [Архитектура](docs/ru/architecture/index.md)
  - [Эксплуатация](docs/ru/operations/index.md)
  - [Отчёты](docs/ru/reports/index.md)
- [Главная](README_ru.md)

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
| **Полное покрытие API** | 121/121 эндпоинтов TestRail API v2 реализовано |
| **Интерактивный режим** | Визуальный выбор проектов, сьютов и целей миграции |
| **Синхронизация данных** | Миграция кейсов, shared steps, сьютов и секций между проектами |
| **Управление ранами** | Создание ранов, добавление результатов, отслеживание выполнения |
| **Встроенный jq** | Фильтрация и трансформация JSON без внешних зависимостей |
| **Прогресс в реальном времени** | Прогресс-бары с живым обновлением через каналы для параллельных операций |
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

### Сравнение проектов

Сравнение ресурсов между двумя проектами для выявления различий и совпадений:

```bash
# Сравнить все ресурсы между проектами
gotr compare all --pid1 30 --pid2 34

# Сравнить конкретные типы ресурсов
gotr compare cases --pid1 30 --pid2 34
gotr compare suites --pid1 30 --pid2 34
gotr compare sharedsteps --pid1 30 --pid2 34

# Сохранить результаты сравнения
gotr compare all --pid1 30 --pid2 34 --save
gotr compare cases --pid1 30 --pid2 34 --save-to results.json --format json

# Автоопределение формата по расширению файла
gotr compare all --pid1 30 --pid2 34 --save-to comparison.yaml
```

**Поддерживаемые ресурсы:** `cases`, `suites`, `sections`, `sharedsteps`, `runs`, `plans`, `milestones`, `datasets`, `groups`, `labels`, `templates`, `configurations`, `all`

#### Тюнинг производительности

```bash
# Server (без rate-limit, максимальная скорость)
gotr compare cases --pid1 30 --pid2 34 --rate-limit 0

# Cloud Enterprise (повышенный лимит)
gotr compare cases --pid1 30 --pid2 34 --rate-limit 300

# Больше параллелизма
gotr compare cases --pid1 30 --pid2 34 --parallel-suites 10 --parallel-pages 6
```

Автоматическое определение деплоймента: gotr определяет `cloud/server` по URL и подбирает rate-limit автоматически. Настраивается в конфиге (`compare.deployment`, `compare.cloud_tier`).

#### Точечный дозабор failed pages

```bash
# Если часть страниц не загрузилась — дозабрать только их
gotr compare retry-failed-pages --from ~/.gotr/exports/compare/failed_pages_2026-03-03_10-15-00.json
```

По умолчанию compare cases автоматически пытается дозабрать проблемные страницы.

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

## Отладка

Для диагностики и получения детальной информации о выполнении используйте флаг `--debug` (или `-d`):

```bash
# Показать debug-вывод для любой команды
gotr compare cases --pid1 30 --pid2 34 --debug
gotr sync cases --src-project 30 --dst-project 31 --debug
gotr get cases --project-id 30 --debug

# Debug-вывод включает:
# - Детали API-запросов
# - Информацию о прогрессе
# - Тайминг каждой фазы операции
# - Детали обработки сьютов/кейсов
```

> **Примечание:** Флаг `--debug` скрыт из автодополнения, но доступен во всех командах.

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

## Структура проекта

```text
gotr/
├── cmd/                          # CLI команды (29 подкоманд)
│   ├── internal/testhelper/     #   Общие тест-утилиты
│   ├── get/                     #   GET-команды (cases, suites, projects)
│   ├── run/                     #   Управление тест-ранами
│   ├── result/                  #   Управление результатами
│   ├── compare/                 #   Кросс-проектное сравнение
│   ├── sync/                    #   Миграция данных
│   └── ...                      #   Прочие ресурсные подкоманды
├── docs/                         # Документация (EN + RU)
│   ├── en/                      #   Английская документация
│   └── ru/                      #   Русская документация
├── embedded/                     # Встроенные бинарники (jq)
├── internal/
│   ├── client/                  #   Клиент TestRail API
│   │   ├── interfaces.go       #     ClientInterface (130+ методов, 16 API)
│   │   ├── mock.go             #     MockClient для тестирования
│   │   └── *.go                #     Реализации API
│   ├── concurrency/            #   Доменная параллельная оркестрация
│   │   ├── controller.go       #     ParallelController — стриминг suite/page
│   │   └── simple.go           #     FetchParallel[T], FetchParallelBySuite[T]
│   ├── concurrent/             #   Низкоуровневые примитивы параллелизма
│   │   ├── pool.go             #     WorkerPool
│   │   ├── limiter.go          #     AdaptiveRateLimiter (180 req/min)
│   │   └── retry.go            #     Экспоненциальный backoff retry
│   ├── interactive/            #   Интерактивные промпты (survey)
│   ├── service/                #   Бизнес-логика
│   │   ├── run.go              #     RunService
│   │   ├── result.go           #     ResultService
│   │   └── migration/          #     Движок миграции данных
│   ├── models/                 #   Модели данных
│   │   ├── data/              #     API DTO
│   │   └── config/            #     Модель конфигурации
│   ├── output/                 #   Форматирование вывода (JSON/YAML/table)
│   ├── ui/                     #   Терминальный UI (прогресс, превью)
│   ├── flags/                  #   Парсинг общих флагов
│   ├── log/                    #   Структурированное логирование (zap)
│   └── paths/                  #   Утилиты путей
├── pkg/                          # Публичные пакеты
│   ├── testrailapi/            #   Определения API endpoint'ов (135 endpoints)
│   └── reporter/               #   Унифицированный репортер статистики
└── main.go                       # Точка входа
```

См. [docs/ru/architecture/overview.md](docs/ru/architecture/overview.md) для полной структуры.

## Что нового в v3.0.0

- **135 endpoint'ов TestRail API** определены, 98% реализованы в клиенте
- **29 CLI-команд** для всех основных ресурсов TestRail
- **Потоковая параллельная пагинация** с адаптивным rate limiting (180 req/min)
- **100% покрытие тестами** в 35/42 пакетах, минимум 97.4% во всех
- **Ноль замечаний golangci-lint** с порогом gocyclo ≤15
- **Полная документация EN/RU** — 125 doc-страниц

См. [CHANGELOG](CHANGELOG.md) для полной истории изменений.

## Установка

Детальные инструкции по установке: [docs/ru/guides/installation.md](docs/ru/guides/installation.md)

## Участие в проекте

Приветствуются issues и pull requests.

## Используемые библиотеки

| Библиотека | Назначение |
|------------|------------|
| [spf13/cobra](https://github.com/spf13/cobra) | CLI-фреймворк |
| [spf13/viper](https://github.com/spf13/viper) | Управление конфигурацией |
| [go.uber.org/zap](https://github.com/uber-go/zap) | Структурированное логирование |
| [stretchr/testify](https://github.com/stretchr/testify) | Тестирование |
| [AlecAivazis/survey/v2](https://github.com/AlecAivazis/survey) | Интерактивные промпты |
| [jedib0t/go-pretty/v6](https://github.com/jedib0t/go-pretty) | Табличный вывод |
| [fatih/color](https://github.com/fatih/color) | Цветной вывод в терминале |
| [golang.org/x/sync](https://pkg.go.dev/golang.org/x/sync) | Утилиты параллелизма |
| [golang.org/x/time](https://pkg.go.dev/golang.org/x/time) | Rate limiting |

### Встроенные инструменты

| Инструмент | Назначение |
|------------|------------|
| [jq](https://github.com/jqlang/jq) | Легковесный JSON-процессор, встроен как статический бинарник для поддержки `--jq` / `--jq-filter` |

## Лицензия

MIT License — см. [LICENSE](LICENSE)
