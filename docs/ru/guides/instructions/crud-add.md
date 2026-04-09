# Инструкция: Создание объектов (add)

Language: Русский | [English](../../../en/guides/instructions/crud-add.md)

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

Команда `gotr add` создаёт новые объекты в TestRail через POST API.
Поддерживает интерактивный режим (wizard), dry-run и создание из JSON-файла.

> [!WARNING]
> `add` **изменяет данные** в TestRail. Всегда используйте `--dry-run` для проверки перед созданием.

## Примеры 🚀

### Создание проекта

```bash
# Через флаги
gotr add project --name "Новый проект" --description "Описание"

# Dry-run — проверить что будет создано
gotr add project --name "Новый проект" --dry-run

# Интерактивный wizard
gotr add project -i
```

### Создание suite

```bash
# Создать набор в проекте
gotr add suite 30 --name "Регрессия" --description "Регрессионные кейсы"

# Dry-run
gotr add suite 30 --name "Регрессия" --dry-run
```

### Создание test case

```bash
# С минимальными полями
gotr add case --title "Проверка авторизации" \
  --suite-id 20069 --section-id 500

# С дополнительными полями
gotr add case --title "Проверка авторизации" \
  --suite-id 20069 --section-id 500 \
  --type-id 1 --priority-id 3 --template-id 1

# Из JSON-файла
gotr add case --json-file case-data.json

# Интерактивный wizard
gotr add case -i
```

### Создание test run

```bash
# Создать run со всеми кейсами
gotr add run 30 --name "Smoke test" --suite-id 20069

# Создать run с конкретными кейсами
gotr add run 30 --name "Smoke test" \
  --suite-id 20069 \
  --include-all=false \
  --case-ids "101,102,103"

# Назначить на пользователя
gotr add run 30 --name "Smoke test" \
  --suite-id 20069 \
  --assignedto-id 5
```

### Создание shared step

```bash
# Создать shared step
gotr add shared-step 30 --title "Войти в систему"

# Из JSON-файла с шагами
gotr add shared-step 30 --json-file shared-step-data.json
```

### Добавление результата теста

```bash
# Добавить результат
gotr add result --test-id 12345 \
  --status-id 1 --comment "Passed" --elapsed "30s"
```

## Режимы создания 🧩

### Флаги (inline)

```bash
gotr add <endpoint> [id] --name "Название" --description "Описание"
```

### JSON-файл

```bash
gotr add <endpoint> [id] --json-file data.json
```

### Интерактивный wizard

```bash
gotr add <endpoint> -i
```

### Dry-run (проверка)

```bash
gotr add <endpoint> [id] --name "Название" --dry-run
```

## Основные флаги ⚙️

| Флаг | Описание |
| --- | --- |
| `--dry-run` | Показать что будет создано без отправки |
| `-i, --interactive` | Интерактивный wizard |
| `--json-file` | Путь к JSON-файлу с данными |
| `--save` | Сохранить результат в файл |
| `-n, --name` | Название ресурса |
| `--title` | Заголовок (для case) |
| `--description` | Описание |
| `--suite-id` | ID набора |
| `--section-id` | ID секции |
| `--type-id` | ID типа (для case) |
| `--priority-id` | ID приоритета (для case) |
| `--template-id` | ID шаблона (для case) |
| `--milestone-id` | ID milestone |
| `--assignedto-id` | ID пользователя |
| `--case-ids` | ID кейсов через запятую (для run) |
| `--include-all` | Включить все кейсы (для run) |

## Проверка результата

```bash
# После создания — проверить через get
gotr get project <new_id>
gotr get case <new_id>
gotr get suite <new_id>
```

## FAQ ❓

- ❓ **Вопрос:** Как создать объект из заранее подготовленного JSON?
  > ↪️ **Ответ:** `gotr add <endpoint> --json-file data.json`. Формат JSON соответствует API TestRail.
  >
  > ---

- ❓ **Вопрос:** Что если в dry-run всё ок, а при создании — ошибка?
  > ↪️ **Ответ:** dry-run проверяет формат данных локально. Ошибки API (дубликаты, недостаточные права, отсутствующие зависимости) возникают только при реальном запросе.
  >
  > ---

- ❓ **Вопрос:** Можно ли создать несколько объектов за раз?
  > ↪️ **Ответ:** для массового создания кейсов используйте `gotr cases bulk`. Для остальных ресурсов — последовательные вызовы `add`.

---

← [Инструкции](index.md)
