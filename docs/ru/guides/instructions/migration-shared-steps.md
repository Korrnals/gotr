# Инструкция: Миграция shared steps

Language: Русский | [English](../../../en/guides/instructions/migration-shared-steps.md)

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

Миграция только **общих тестовых шагов (shared steps)** между проектами.
Команда `gotr sync shared-steps` выполняет:

1. Загрузку всех shared steps из исходного проекта
2. Загрузку cases из указанного набора (если `--src-suite` задан)
3. Фильтрацию: оставляет только shared steps, привязанные к кейсам набора (поле "Used In / case_ids")
4. Дедупликацию: исключает шаги, которые уже есть в целевом проекте (по `title`)
5. Импорт новых shared steps в целевой проект
6. Сохранение mapping старых ID → новых ID

> [!TIP]
> Используйте `--save-mapping`, чтобы сохранить файл соответствий для последующего
> `gotr sync cases --mapping-file` — это гарантирует корректные ссылки в кейсах.

## Предусловия ✅

- [ ] gotr настроен и подключён к TestRail (`gotr self-test`)
- [ ] Известен ID исходного проекта
- [ ] Известен ID целевого проекта
- [ ] (Опционально) Известен ID набора для фильтрации shared steps

## Сценарий 1: Перенос shared steps с фильтрацией по набору 🚀

### Входные данные

| Параметр | Значение | Описание |
| --- | --- | --- |
| Исходный проект | `30` | Проект R189 |
| Исходный набор | `20069` | Набор для фильтрации |
| Целевой проект | `34` | Тестирование E2E сценариев |

### Шаг 1. Разведка

```bash
# Посмотреть все shared steps исходного проекта
gotr get sharedsteps 30

# Экспортировать для детального анализа
gotr export sharedsteps -p 30 --save --format json
```

### Шаг 2. Dry-run

```bash
gotr sync shared-steps \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 34 \
  --dry-run --save-filtered
```

**Что покажет:**

- Сколько shared steps связаны с кейсами набора 20069
- Сколько из них уже есть в целевом проекте (дубликаты)
- Сколько новых шагов будет создано

### Шаг 3. Выполнить миграцию

```bash
gotr sync shared-steps \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 34 \
  --save-mapping --approve
```

### Шаг 4. Проверить результат

```bash
# Проверить shared steps в целевом проекте
gotr get sharedsteps 34

# Сравнить shared steps между проектами
gotr compare sharedsteps --pid1 30 --pid2 34
```

## Сценарий 2: Перенос ВСЕХ shared steps проекта 🚀

Без указания `--src-suite` переносятся все shared steps проекта:

```bash
# Dry-run — посмотреть план
gotr sync shared-steps \
  --src-project 30 \
  --dst-project 34 \
  --dry-run

# Выполнить
gotr sync shared-steps \
  --src-project 30 \
  --dst-project 34 \
  --save-mapping --approve
```

## Синтаксис 🧩

```bash
gotr sync shared-steps \
  --src-project <ID> \
  [--src-suite <ID>] \
  --dst-project <ID> \
  [--compare-field <field>] \
  [--output <path>] \
  [--dry-run] \
  [--save-mapping] \
  [--save-filtered] \
  [--approve] \
  [--quiet]
```

## Флаги ⚙️

| Флаг | Описание | По умолчанию |
| --- | --- | --- |
| `--src-project` | ID исходного проекта | обязательный |
| `--src-suite` | ID набора для фильтрации по "Used In" | — (все шаги) |
| `--dst-project` | ID целевого проекта | обязательный |
| `--compare-field` | Поле для поиска дубликатов | `title` |
| `--output` | Путь для сохранения mapping | — |
| `--dry-run` | Показать план без изменений | `false` |
| `--save-mapping` | Сохранить mapping (старый ID → новый) | `false` |
| `--save-filtered` | Сохранить отфильтрованный список | `false` |
| `--approve` | Пропустить подтверждение | `false` |
| `--quiet` | Подавить служебный вывод | `false` |

## Как работает фильтрация 📐

Алгоритм фильтрации shared steps по набору:

```text
Для каждого shared step в исходном проекте:
  1. Проверить поле case_ids ("Used In")
  2. Если хотя бы один case_id принадлежит кейсам из --src-suite → шаг КАНДИДАТ
  3. Среди кандидатов: сравнить title с шагами целевого проекта
     - Совпадение → добавить в mapping как "existing" (не импортировать)
     - Нет совпадения → ИМПОРТИРОВАТЬ
```

## Ожидаемый результат 🧾

### Артефакты

| Файл | Содержимое |
| --- | --- |
| `mapping.json` | `{ "source_id": 123, "target_id": 456, "status": "created" }` — для новых |
| `mapping.json` | `{ "source_id": 789, "target_id": 101, "status": "existing" }` — для дубликатов |
| `filtered.json` | Список shared steps, прошедших фильтрацию |

## FAQ ❓

- ❓ **Вопрос:** Что если shared step используется в нескольких наборах?
  > ↪️ **Ответ:** фильтрация проверяет пересечение `case_ids` с кейсами указанного набора. Если хотя бы один кейс из набора использует шаг — он попадёт в кандидаты.
  >
  > ---

- ❓ **Вопрос:** Что будет при повторном запуске?
  > ↪️ **Ответ:** шаги, которые уже есть в целевом проекте (по title), будут отмечены как `existing` и не дублированы.
  >
  > ---

- ❓ **Вопрос:** Как использовать mapping-файл дальше?
  > ↪️ **Ответ:** передайте его в `gotr sync cases --mapping-file mapping.json` — см. [Частичная миграция](migration-partial.md).

---

← [Инструкции](index.md)
