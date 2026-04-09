# Инструкция: Экспорт данных (export)

Language: Русский | [English](../../../en/guides/instructions/crud-export.md)

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

Команда `gotr export` — универсальный экспорт данных из TestRail в файл.
Поддерживает **30+ типов ресурсов** и 5 форматов вывода.

> [!TIP]
> `export` — безопасная операция чтения. Подходит для резервного копирования,
> разведки перед миграцией и подготовки данных для анализа.

## Доступные ресурсы

```text
all, cases, casefields, casetypes, configurations, projects, priorities,
runs, tests, suites, sections, statuses, milestones, plans, results,
resultfields, reports, attachments, users, roles, templates, groups,
sharedsteps, variables, labels, datasets, bdds
```

## Примеры 🚀

### Экспорт проектов и структуры

```bash
# Все проекты
gotr export projects --save

# Все suites проекта
gotr export suites -p 30 --save

# Все sections набора
gotr export sections -p 30 -s 20069 --save
```

### Экспорт кейсов

```bash
# Все кейсы набора в JSON
gotr export cases -p 30 -s 20069 --save --format json

# Кейсы конкретной секции
gotr export cases -p 30 -s 20069 --section-id 12345 --save

# Кейсы в CSV для анализа в таблицах
gotr export cases -p 30 -s 20069 --format csv --save
```

### Экспорт shared steps

```bash
# Все shared steps проекта
gotr export sharedsteps -p 30 --save --format json

# В табличном виде для быстрого просмотра
gotr export sharedsteps -p 30
```

### Экспорт test runs и результатов

```bash
# Все runs проекта
gotr export runs -p 30 --save

# Runs для конкретного milestone
gotr export runs -p 30 --milestone-id 5 --save

# Все результаты
gotr export results -p 30 --save
```

### Массовый экспорт

```bash
# Экспорт ВСЕХ поддерживаемых ресурсов проекта
gotr export all -p 30 --save
```

## Форматы вывода 🧩

| Формат | Флаг | Применение |
| --- | --- | --- |
| Таблица | `--format table` | Просмотр в терминале (по умолчанию) |
| JSON | `--format json` | Анализ, скрипты, хранение |
| CSV | `--format csv` | Импорт в таблицы (Excel, Sheets) |
| Markdown | `--format md` | Документация |
| HTML | `--format html` | Отчёты для браузера |

## Сохранение

Файлы сохраняются в `~/.gotr/exports/export/`:

```bash
# Автоматическое имя файла
gotr export cases -p 30 -s 20069 --save

# Тихий режим (только файл, без вывода в терминал)
gotr export cases -p 30 -s 20069 --save --quiet
```

## Синтаксис 🧩

```bash
gotr export <resource> <endpoint> [id] [flags]
```

## Флаги ⚙️

| Флаг | Описание | По умолчанию |
| --- | --- | --- |
| `-p, --project-id` | ID проекта | — |
| `-s, --suite-id` | ID набора (для cases) | — |
| `--section-id` | ID секции (для cases) | — |
| `--milestone-id` | ID milestone (для runs) | — |
| `--save` | Сохранить в файл | `false` |
| `-f, --format` | Формат: table, json, csv, md, html | `table` |
| `-q, --quiet` | Подавить служебный вывод | `false` |

## Практический сценарий: подготовка к миграции 🧩

```bash
# 1. Экспортировать shared steps для анализа
gotr export sharedsteps -p 30 --save --format json

# 2. Экспортировать кейсы для ревью
gotr export cases -p 30 -s 20069 --save --format json

# 3. Проанализировать данные локально
cat ~/.gotr/exports/export/*.json | jq '.[] | .title'

# 4. Если всё ок — перейти к миграции
# См. Полная миграция или Миграция shared steps
```

## FAQ ❓

- ❓ **Вопрос:** Куда сохраняются файлы?
  > ↪️ **Ответ:** в `~/.gotr/exports/export/`. Имя файла формируется автоматически из ресурса и timestamp.
  >
  > ---

- ❓ **Вопрос:** Как экспортировать данные нескольких проектов?
  > ↪️ **Ответ:** выполняйте `export` для каждого проекта с `-p <id>`. Или используйте `gotr export all` без `-p` для глобальных ресурсов (users, roles, templates).
  >
  > ---

- ❓ **Вопрос:** Можно ли использовать экспортированный JSON для импорта?
  > ↪️ **Ответ:** для импорта используйте `gotr sync`, а не raw JSON. Экспорт — для анализа и архивирования.

---

← [Инструкции](index.md)
