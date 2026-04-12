# Инструкция: Обновление объектов (update)

Language: Русский | [English](../../../en/guides/instructions/crud-update.md)

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

Команда `gotr update` изменяет существующие объекты в TestRail.
Поддерживает обновление по ID, интерактивный режим и обновление из JSON-файла.

> [!WARNING]
> `update` **изменяет данные** в TestRail. Используйте `--dry-run` для проверки.

## Примеры 🚀

### Обновление проекта

```bash
# Изменить название
gotr update project 30 --name "R189 (обновлено)"

# Изменить описание
gotr update project 30 --description "Новое описание проекта"

# Dry-run
gotr update project 30 --name "R189 (обновлено)" --dry-run
```

### Обновление test case

```bash
# Изменить заголовок
gotr update case 12345 --title "Обновлённый заголовок"

# Изменить приоритет и тип
gotr update case 12345 --priority-id 4 --type-id 2

# Из JSON-файла
gotr update case 12345 --json-file case-update.json

# Интерактивный режим
gotr update case 12345 -i
```

### Обновление suite

```bash
gotr update suite 20069 --name "Обновлённый набор" --description "Новое описание"
```

### Обновление shared step

```bash
# Изменить заголовок
gotr update shared-step 456 --title "Обновлённый шаг"

# Из JSON с полным набором данных
gotr update shared-step 456 --json-file step-update.json
```

### Обновление test run

```bash
# Изменить название и описание
gotr update run 789 --name "Регрессия v2" --description "Обновлённый run"

# Переназначить
gotr update run 789 --assignedto-id 10
```

### Обновление milestone

```bash
gotr update milestone 50 --name "Релиз 3.1" --description "Новый milestone"
```

> [!NOTE]
> Диспетчер `gotr update` не поддерживает endpoint'ы `milestone` и `plan` напрямую.
> Используйте выделенные подкоманды:
> - `gotr milestones update <id> --name "..."` — обновить milestone
> - `gotr plans update <id> --name "..."` — обновить plan

## Режимы обновления 🧩

### Флаги (inline)

```bash
gotr update <endpoint> <id> --title "Новое значение"
```

### JSON-файл

```bash
gotr update <endpoint> <id> --json-file data.json
```

### Интерактивный wizard

```bash
gotr update <endpoint> <id> -i
```

### Dry-run (проверка)

```bash
gotr update <endpoint> <id> --title "Новое значение" --dry-run
```

## Основные флаги ⚙️

| Флаг | Описание |
| --- | --- |
| `--dry-run` | Показать что будет изменено без отправки |
| `-i, --interactive` | Интерактивный wizard |
| `--json-file` | Путь к JSON-файлу с данными |
| `--title` | Новый заголовок |
| `-n, --name` | Новое название |
| `--description` | Новое описание |
| `--priority-id` | Новый приоритет |
| `--type-id` | Новый тип |
| `--labels` | Новые метки |

## Проверка результата

```bash
# После обновления — проверить через get
gotr get case 12345
gotr get project 30
gotr get suite 20069
```

## Типичный пайплайн: get → проверить → update → get

```bash
# 1. Получить текущее состояние
gotr get case 12345 --format json

# 2. Проверить что будет изменено
gotr update case 12345 --title "Новый заголовок" --dry-run

# 3. Выполнить обновление
gotr update case 12345 --title "Новый заголовок"

# 4. Проверить результат
gotr get case 12345
```

## FAQ ❓

- ❓ **Вопрос:** Какие поля можно обновить?
  > ↪️ **Ответ:** зависит от эндпоинта. Используйте `gotr update <endpoint> --help` для полного списка флагов конкретного ресурса.
  >
  > ---

- ❓ **Вопрос:** Можно ли обновить несколько объектов сразу?
  > ↪️ **Ответ:** для массового обновления кейсов используйте `gotr cases bulk`. Для остальных — последовательные вызовы.
  >
  > ---

- ❓ **Вопрос:** Что если указать несуществующий ID?
  > ↪️ **Ответ:** API TestRail вернёт ошибку 400/404, gotr покажет сообщение об ошибке.

---

← [Инструкции](index.md)
