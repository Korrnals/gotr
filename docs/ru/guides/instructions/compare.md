# Инструкция: Сравнение проектов (compare)

Language: Русский | [English](../../../en/guides/instructions/compare.md)

## Навигация

- [Документация](../../index.md)
  - [Гайды](../index.md)
    - [Установка](../installation.md)
    - [Конфигурация](../configuration.md)
    - [Интерактивный режим](../interactive-mode.md)
    - [Прогресс](../progress.md)
    - [Каталог команд](../commands/index.md)
    - [Инструкции](index.md)
      - [Полная миграция](migration-full.md)
      - [Частичная миграция](migration-partial.md)
      - [Миграция shared steps](migration-shared-steps.md)
      - [Миграция ресурсов](migration-resources.md)
      - [Получение данных](crud-get.md)
      - [Экспорт данных](crud-export.md)
      - [Создание объектов](crud-add.md)
      - [Обновление объектов](crud-update.md)
      - [Удаление объектов](crud-delete.md)
      - [Сравнение проектов](compare.md)
  - [Архитектура](../../architecture/index.md)
  - [Эксплуатация](../../operations/index.md)
  - [Отчёты](../../reports/index.md)
- [Главная](../../../../README_ru.md)

## Обзор 🎯

Команда `gotr compare` — сравнение ресурсов между двумя проектами TestRail.
Применяется для **аудита перед миграцией**, **верификации после миграции** и **мониторинга расхождений**.

> [!TIP]
> `compare` — безопасная операция чтения. Поддерживает параллельную загрузку данных
> и сохранение результатов в JSON/CSV/HTML.

## Доступные ресурсы

| Подкоманда | Описание |
| --- | --- |
| `all` | Сравнить ВСЕ поддерживаемые ресурсы |
| `cases` | Test cases |
| `suites` | Test suites |
| `sections` | Sections |
| `sharedsteps` | Shared steps |
| `runs` | Test runs |
| `plans` | Test plans |
| `milestones` | Milestones |
| `configurations` | Конфигурации |
| `datasets` | Датасеты |
| `groups` | Группы |
| `labels` | Метки |
| `templates` | Шаблоны |

## Примеры 🚀

### Полное сравнение двух проектов

```bash
# Сравнить все ресурсы
gotr compare all --pid1 30 --pid2 34 --save

# Сохранить в конкретный файл
gotr compare all --pid1 30 --pid2 34 --save-to comparison-report.json
```

### Сравнение конкретных ресурсов

```bash
# Shared steps
gotr compare sharedsteps --pid1 30 --pid2 34

# Cases
gotr compare cases --pid1 30 --pid2 34

# Suites
gotr compare suites --pid1 30 --pid2 34

# Sections
gotr compare sections --pid1 30 --pid2 34
```

### Сохранение в разных форматах

```bash
# JSON (для скриптов и анализа)
gotr compare all --pid1 30 --pid2 34 --format json --save

# HTML (для отчётов)
gotr compare all --pid1 30 --pid2 34 --format html --save

# CSV (для таблиц)
gotr compare cases --pid1 30 --pid2 34 --format csv --save
```

### Тюнинг производительности

```bash
# Увеличить параллелизм для больших проектов
gotr compare all --pid1 30 --pid2 34 \
  --parallel-suites 15 \
  --parallel-pages 10 \
  --timeout 60m
```

## Синтаксис 🧩

```bash
gotr compare <resource> --pid1 <ID> --pid2 <ID> [flags]
```

## Флаги ⚙️

| Флаг | Описание | По умолчанию |
| --- | --- | --- |
| `-1, --pid1` | ID первого проекта | обязательный |
| `-2, --pid2` | ID второго проекта | обязательный |
| `--save` | Сохранить в `~/.gotr/exports/` | `false` |
| `--save-to` | Сохранить в конкретный файл | — |
| `-f, --format` | Формат: json, csv, md, html | `table` |
| `--parallel-suites` | Параллелизм по suites | `10` |
| `--parallel-pages` | Параллелизм по страницам | `6` |
| `--rate-limit` | Rate limit API (-1=auto, 0=unlimited) | `-1` |
| `--timeout` | Таймаут операции | `30m` |
| `-q, --quiet` | Подавить служебный вывод | `false` |

## Сценарий: аудит перед миграцией 🧩

```bash
# 1. Сравнить все ресурсы для полной картины
gotr compare all --pid1 30 --pid2 34 --save-to pre-migration-audit.json

# 2. Проверить shared steps отдельно
gotr compare sharedsteps --pid1 30 --pid2 34

# 3. Проверить cases
gotr compare cases --pid1 30 --pid2 34

# 4. По результатам — решить какой тип миграции нужен
# → Полная миграция: migration-full.md
# → Только shared steps: migration-shared-steps.md
```

## Сценарий: верификация после миграции 🧩

```bash
# 1. Сравнить shared steps — все ли перенеслись
gotr compare sharedsteps --pid1 30 --pid2 34

# 2. Сравнить cases — все ли кейсы на месте
gotr compare cases --pid1 30 --pid2 34

# 3. Полный отчёт для документирования
gotr compare all --pid1 30 --pid2 34 \
  --format html --save-to post-migration-report.html
```

## FAQ ❓

- ❓ **Вопрос:** Что показывает сравнение?
  > ↪️ **Ответ:** количество объектов в каждом проекте, совпадения (по title/name), уникальные для каждого проекта элементы, общие и различающиеся.
  >
  > ---

- ❓ **Вопрос:** Как ускорить сравнение для больших проектов?
  > ↪️ **Ответ:** увеличьте `--parallel-suites` и `--parallel-pages`. Для проектов с 10000+ кейсами рекомендуется `--timeout 60m`.
  >
  > ---

- ❓ **Вопрос:** Можно ли сравнить один конкретный ресурс?
  > ↪️ **Ответ:** да, используйте конкретную подкоманду: `gotr compare cases`, `gotr compare sharedsteps` и т.д.
  >
  > ---

- ❓ **Вопрос:** Что делать с `retry-failed-pages`?
  > ↪️ **Ответ:** если при сравнении часть страниц не загрузилась (таймаут, rate limit), используйте `gotr compare retry-failed-pages` для повторной загрузки.

---

← [Инструкции](index.md)
