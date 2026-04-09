# Инструкция: Получение данных (get)

Language: Русский | [English](../../../en/guides/instructions/crud-get.md)

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

Команда `gotr get` — основной инструмент для **чтения данных** из TestRail.
Поддерживает получение списков и отдельных объектов для всех типов ресурсов.

> [!TIP]
> Все команды `get` безопасны — они только читают данные и не изменяют ничего в TestRail.

## Доступные ресурсы

| Ресурс | Синтаксис | Описание |
| --- | --- | --- |
| `projects` | `gotr get projects` | Все проекты |
| `project` | `gotr get project <id>` | Один проект |
| `suites` | `gotr get suites <project_id>` | Наборы проекта |
| `suite` | `gotr get suite <id>` | Один набор |
| `cases` | `gotr get cases <project_id>` | Кейсы проекта |
| `case` | `gotr get case <id>` | Один кейс |
| `case-fields` | `gotr get case-fields` | Поля кейсов |
| `case-types` | `gotr get case-types` | Типы кейсов |
| `case-history` | `gotr get case-history <id>` | История кейса |
| `sharedsteps` | `gotr get sharedsteps <project_id>` | Shared steps |
| `sharedstep` | `gotr get sharedstep <id>` | Один shared step |
| `sharedstep-history` | `gotr get sharedstep-history <id>` | История shared step |
| `runs` | `gotr get runs <project_id>` | Test runs проекта |
| `run` | `gotr get run <id>` | Один run |
| `tests` | `gotr get tests <run_id>` | Тесты в run |
| `test` | `gotr get test <id>` | Один тест |
| `results` | `gotr get results <test_id>` | Результаты теста |
| `plans` | `gotr get plans <project_id>` | Test plans |
| `plan` | `gotr get plan <id>` | Один plan |
| `milestones` | `gotr get milestones <project_id>` | Вехи проекта |
| `users` | `gotr get users` | Все пользователи |
| `user` | `gotr get user <id>` | Один пользователь |

## Примеры 🚀

### Проекты и структура

```bash
# Все проекты
gotr get projects

# Конкретный проект
gotr get project 30

# Наборы (suites) проекта
gotr get suites 30

# Конкретный набор
gotr get suite 20069
```

### Кейсы

```bash
# Все кейсы проекта (требует --suite-id для multi-suite проектов)
gotr get cases 30 --suite-id 20069

# Конкретный кейс
gotr get case 12345

# История изменений кейса
gotr get case-history 12345
```

### Shared steps

```bash
# Все shared steps проекта
gotr get sharedsteps 30

# Конкретный shared step
gotr get sharedstep 456

# История shared step
gotr get sharedstep-history 456
```

### Test runs и результаты

```bash
# Все runs проекта
gotr get runs 30

# Тесты в конкретном run
gotr get tests 789

# Результаты конкретного теста
gotr get results 101
```

## Формат вывода 🧩

```bash
# Табличный (по умолчанию)
gotr get projects

# JSON
gotr get projects --format json

# CSV
gotr get projects --format csv

# Markdown
gotr get projects --format md

# HTML
gotr get projects --format html
```

## Интерактивный режим 🧩

Если запустить `get` без ID — активируется интерактивный выбор:

```bash
# Интерактивный выбор проекта, затем suites
gotr get suites

# Интерактивный выбор кейса
gotr get case
```

## Сохранение результата

```bash
# Сохранить в ~/.gotr/exports/
gotr get sharedsteps 30 --save

# JSON с сохранением
gotr get cases 30 --suite-id 20069 --format json --save
```

## FAQ ❓

- ❓ **Вопрос:** Чем `get` отличается от `export`?
  > ↪️ **Ответ:** `get` получает данные через конкретные GET-эндпоинты с типизированным выводом. `export` — более универсальный, работает с любым ресурсом и ориентирован на сохранение в файл. Для быстрого просмотра используйте `get`, для сохранения архивов — `export`.
  >
  > ---

- ❓ **Вопрос:** Как получить кейсы определённой секции?
  > ↪️ **Ответ:** используйте `gotr export cases -p <project_id> -s <suite_id> --section-id <id>`.

---

← [Инструкции](index.md)
