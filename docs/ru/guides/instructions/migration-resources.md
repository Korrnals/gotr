# Инструкция: Миграция ресурсов (suites, sections)

Language: Русский | [English](../../../en/guides/instructions/migration-resources.md)

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

Миграция структурных ресурсов: **наборов (suites)** и **секций (sections)** между проектами.
Используется для подготовки целевого проекта перед переносом кейсов.

> [!TIP]
> Порядок миграции ресурсов: suites → sections → shared steps → cases.
> Для полного переноса сразу используйте [Полную миграцию](migration-full.md).

## Предусловия ✅

- [ ] gotr настроен и подключён к TestRail (`gotr self-test`)
- [ ] Известны ID исходного и целевого проекта

---

## Сценарий 1: Миграция suites 🚀

Перенос наборов (test suites) из одного проекта в другой.

### Разведка

```bash
# Посмотреть suites исходного проекта
gotr get suites 30

# Посмотреть suites целевого проекта
gotr get suites 34
```

### Dry-run

```bash
gotr sync suites \
  --src-project 30 \
  --dst-project 34 \
  --dry-run
```

### Выполнить

```bash
gotr sync suites \
  --src-project 30 \
  --dst-project 34 \
  --save-mapping --approve
```

### Проверить

```bash
gotr get suites 34
gotr compare suites --pid1 30 --pid2 34
```

---

## Сценарий 2: Миграция sections 🚀

Перенос секций (sections) между наборами двух проектов.

### Разведка

```bash
# Посмотреть секции исходного набора
gotr export sections -p 30 -s 20069 --save --format json

# Посмотреть секции целевого набора
gotr export sections -p 34 -s 19859 --save --format json
```

### Dry-run

```bash
gotr sync sections \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 34 \
  --dst-suite 19859 \
  --dry-run
```

### Выполнить

```bash
gotr sync sections \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 34 \
  --dst-suite 19859 \
  --save-mapping --approve
```

### Проверить

```bash
gotr compare sections --pid1 30 --pid2 34
```

---

## Синтаксис 🧩

### sync suites

```bash
gotr sync suites \
  --src-project <ID> \
  --dst-project <ID> \
  [--compare-field <field>] \
  [--dry-run] \
  [--save-mapping] \
  [--approve] \
  [--quiet]
```

### sync sections

```bash
gotr sync sections \
  --src-project <ID> \
  --src-suite <ID> \
  --dst-project <ID> \
  --dst-suite <ID> \
  [--compare-field <field>] \
  [--dry-run] \
  [--save-mapping] \
  [--approve] \
  [--quiet]
```

## Флаги ⚙️

| Флаг | Описание | По умолчанию |
| --- | --- | --- |
| `--src-project` | ID исходного проекта | обязательный |
| `--src-suite` | ID исходного набора (для sections) | обязательный для sections |
| `--dst-project` | ID целевого проекта | обязательный |
| `--dst-suite` | ID целевого набора (для sections) | обязательный для sections |
| `--compare-field` | Поле для поиска дубликатов | `title` |
| `--dry-run` | Показать план без изменений | `false` |
| `--save-mapping` | Сохранить mapping в файл | `false` |
| `--approve` | Пропустить подтверждение | `false` |
| `--quiet` | Подавить служебный вывод | `false` |

## Полный пайплайн миграции структуры 🧩

Если нужно перенести всю структуру проекта:

```bash
# 1. Перенести наборы
gotr sync suites \
  --src-project 30 --dst-project 34 \
  --save-mapping --approve

# 2. Перенести секции для каждого набора
gotr sync sections \
  --src-project 30 --src-suite 20069 \
  --dst-project 34 --dst-suite 19859 \
  --save-mapping --approve

# 3. Далее — shared steps и cases
# См. Полная миграция или Миграция shared steps
```

## FAQ ❓

- ❓ **Вопрос:** Что если suite с таким именем уже существует?
  > ↪️ **Ответ:** gotr определяет дубликаты по `title` и не создаёт повторные наборы.
  >
  > ---

- ❓ **Вопрос:** Переносится ли иерархия вложенных секций?
  > ↪️ **Ответ:** да, `sync sections` сохраняет parent-child связи между секциями.
  >
  > ---

- ❓ **Вопрос:** Нужно ли переносить секции отдельно, если я использую `sync full`?
  > ↪️ **Ответ:** `sync full` обрабатывает shared steps и cases, но не sections. Если нужна структура секций — переносите её отдельно перед `sync full`.

---

← [Инструкции](index.md)
