# Инструкция: Частичная миграция (cases с mapping)

Language: Русский | [English](../../../en/guides/instructions/migration-partial.md)

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

Частичная миграция — перенос **только test cases** между наборами двух проектов.
Используется когда shared steps **уже перенесены** отдельно и у вас есть mapping-файл с соответствием старых и новых ID.

Команда `gotr sync cases` автоматически:

1. Загружает cases из исходного набора
2. Заменяет `shared_step_id` в каждом кейсе по mapping-файлу
3. Импортирует cases в целевой набор

> [!TIP]
> Этот сценарий — второй шаг после `gotr sync shared-steps --save-mapping`.
> Для переноса всего сразу используйте [Полную миграцию](migration-full.md).

## Предусловия ✅

- [ ] gotr настроен и подключён к TestRail (`gotr self-test`)
- [ ] Shared steps уже перенесены (или не используются в кейсах)
- [ ] Есть mapping-файл от предыдущего шага `sync shared-steps`
- [ ] Целевой набор (suite) уже создан в целевом проекте

## Пример: перенос cases после миграции shared steps 🚀

### Входные данные

| Параметр | Значение | Описание |
| --- | --- | --- |
| Исходный проект | `30` | Проект R189 |
| Исходный набор | `20069` | Набор с кейсами |
| Целевой проект | `34` | Тестирование E2E сценариев |
| Целевой набор | `19859` | Сценарии R189 (перенос) |
| Mapping-файл | `mapping.json` | Результат предыдущего `sync shared-steps` |

### Шаг 1. Убедиться что mapping-файл на месте

```bash
# Проверить содержимое mapping-файла
cat mapping.json
```

Mapping-файл содержит пары `старый_id → новый_id` для shared steps.

### Шаг 2. Dry-run — проверить план

```bash
gotr sync cases \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 34 \
  --dst-suite 19859 \
  --mapping-file mapping.json \
  --dry-run
```

**Что проверить:**

- Количество кейсов для переноса
- Корректность замены `shared_step_id`

### Шаг 3. Выполнить миграцию

```bash
gotr sync cases \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 34 \
  --dst-suite 19859 \
  --mapping-file mapping.json
```

### Шаг 4. Проверить результат

```bash
# Проверить кейсы в целевом наборе
gotr export cases -p 34 -s 19859 --save --format json

# Сравнить наборы
gotr compare cases --pid1 30 --pid2 34 --save
```

## Синтаксис 🧩

```bash
gotr sync cases \
  --src-project <ID> \
  --src-suite <ID> \
  --dst-project <ID> \
  --dst-suite <ID> \
  [--mapping-file <path>] \
  [--compare-field <field>] \
  [--output <path>] \
  [--dry-run] \
  [--quiet]
```

## Флаги ⚙️

| Флаг | Описание | По умолчанию |
| --- | --- | --- |
| `--src-project` | ID исходного проекта | обязательный |
| `--src-suite` | ID исходного набора | обязательный |
| `--dst-project` | ID целевого проекта | обязательный |
| `--dst-suite` | ID целевого набора | обязательный |
| `--mapping-file` | Путь к файлу mapping shared steps | — |
| `--compare-field` | Поле для поиска дубликатов | `title` |
| `--output` | Путь для JSON-файла с результатами | — |
| `--dry-run` | Показать план без изменений | `false` |
| `--quiet` | Подавить служебный вывод | `false` |

## Пошаговый сценарий: два шага вместо sync full 🧩

Если по каким-то причинам `sync full` не подходит, выполняйте два шага раздельно:

```bash
# Шаг A: перенести shared steps и сохранить mapping
gotr sync shared-steps \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 34 \
  --save-mapping --approve

# Шаг B: перенести cases с подстановкой ID
gotr sync cases \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 34 \
  --dst-suite 19859 \
  --mapping-file mapping.json
```

## FAQ ❓

- ❓ **Вопрос:** Что если mapping-файл не указан, а кейсы ссылаются на shared steps?
  > ↪️ **Ответ:** кейсы будут перенесены с оригинальными `shared_step_id`. Если такие ID не существуют в целевом проекте, ссылки будут нерабочими.
  >
  > ---

- ❓ **Вопрос:** Можно ли перенести кейсы без shared steps вообще?
  > ↪️ **Ответ:** да, если кейсы не используют shared steps — просто не указывайте `--mapping-file`.
  >
  > ---

- ❓ **Вопрос:** Что если часть shared steps уже была в целевом проекте?
  > ↪️ **Ответ:** mapping-файл от `sync shared-steps` содержит записи `existing` для дубликатов — замена выполнится корректно.

---

← [Инструкции](index.md)
