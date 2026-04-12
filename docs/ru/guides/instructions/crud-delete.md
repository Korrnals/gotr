# Инструкция: Удаление объектов (delete)

Language: Русский | [English](../../../en/guides/instructions/crud-delete.md)

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

Команда `gotr delete` удаляет объекты из TestRail.
Поддерживает мягкое (`--soft`) и жёсткое удаление.

> [!CAUTION]
> `delete` **необратимо удаляет данные** из TestRail.
> Всегда используйте `--dry-run` перед удалением.
> Рекомендуется предварительно экспортировать данные через `gotr export`.

## Примеры 🚀

### Удаление с проверкой

```bash
# 1. Сначала экспорт для бэкапа
gotr export cases -p 30 -s 20069 --save --format json

# 2. Dry-run — проверить что будет удалено
gotr delete case 12345 --dry-run

# 3. Выполнить удаление
gotr delete case 12345
```

### Удаление различных ресурсов

```bash
# Удалить test case
gotr delete case 12345

# Удалить suite
gotr delete suite 20069

# Удалить section
gotr delete section 500

# Удалить test run
gotr delete run 789

# Удалить plan
gotr delete plan 100

# Удалить проект (мягкое удаление)
gotr delete project 30 --soft

# Удалить shared step
gotr delete shared-step 456

# Удалить milestone
gotr delete milestone 50
```

> [!NOTE]
> Диспетчер `gotr delete` не поддерживает endpoint'ы `milestone` и `plan` напрямую.
> Используйте выделенные подкоманды:
> - `gotr milestones delete <id>` — удалить milestone
> - `gotr plans delete <id>` — удалить plan

### Мягкое удаление

```bash
# Мягкое удаление — объект помечается как удалённый, но данные остаются
gotr delete project 30 --soft

# Жёсткое удаление — данные удаляются безвозвратно
gotr delete project 30
```

## Синтаксис 🧩

```bash
gotr delete <endpoint> <id> [flags]
```

## Флаги ⚙️

| Флаг | Описание | По умолчанию |
| --- | --- | --- |
| `--dry-run` | Показать что будет удалено без выполнения | `false` |
| `--soft` | Мягкое удаление (где поддерживается) | `false` |

## Безопасный пайплайн удаления 🧩

```bash
# 1. Экспортировать данные для бэкапа
gotr export <resource> -p <project_id> --save --format json

# 2. Проверить объект перед удалением
gotr get <resource> <id>

# 3. Dry-run
gotr delete <endpoint> <id> --dry-run

# 4. Удалить
gotr delete <endpoint> <id>

# 5. Убедиться что удалено
gotr get <resource> <id>  # Ожидаем ошибку 404
```

## FAQ ❓

- ❓ **Вопрос:** Можно ли восстановить удалённый объект?
  > ↪️ **Ответ:** при `--soft` — объект можно восстановить через UI TestRail. При жёстком удалении — данные утеряны навсегда. Всегда делайте экспорт перед удалением.
  >
  > ---

- ❓ **Вопрос:** Что если удалить suite, в котором есть кейсы?
  > ↪️ **Ответ:** TestRail API удалит suite вместе со всеми кейсами и секциями внутри. Убедитесь что это намеренное действие.
  >
  > ---

- ❓ **Вопрос:** Можно ли удалить несколько объектов за раз?
  > ↪️ **Ответ:** `gotr delete` работает с одним объектом за вызов. Для массовых операций используйте shell-скрипт или `gotr cases bulk`.

---

← [Инструкции](index.md)
