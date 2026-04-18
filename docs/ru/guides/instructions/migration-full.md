# Инструкция: Полная миграция

Language: Русский | [English](../../../en/guides/instructions/migration-full.md)

## Навигация

- [Документация](../../index.md)
  - [Гайды](../index.md)
    - [Установка](../installation.md)
    - [Конфигурация](../configuration.md)
    - [Интерактивный режим](../interactive-mode.md)
    - [Прогресс](../progress.md)
    - [Каталог команд](../commands/index.md)
    - [Инструкции](index.md)
      - [Пошаговая интерактивная миграция](migration-interactive-walkthrough.md)
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

Полная миграция переносит **shared steps + test cases** из одного проекта/набора в другой за один проход.
Команда `gotr sync full` автоматически:

1. Загружает shared steps из исходного проекта
2. Применяет фильтрацию shared steps с учетом source suite
3. Исключает дубликаты по `title` с целевым проектом
4. Импортирует новые shared steps и сохраняет mapping (старый ID → новый ID)
5. Загружает cases из исходного набора
6. Заменяет `shared_step_id` в кейсах по mapping
7. Импортирует cases в целевой набор

> [!TIP]
> Всегда начинайте с `--dry-run`, чтобы увидеть план миграции без внесения изменений.

## Важно: как назначаются ID shared steps

- ID shared step в source проекте не переносится «как есть» в target.
- При создании shared step вызывается API `add_shared_step/<project_id>`, а новый ID назначает TestRail.
- Даже если в target «свободен» такой же числовой ID, утилита не может принудительно занять именно его.
- Связь кейсов с общими шагами сохраняется через mapping `source_shared_step_id -> target_shared_step_id`.
- В `sync full` mapping строится на шаге переноса shared steps и автоматически применяется при переносе кейсов.
- Если shared step уже существует в target (дубликат по полю сравнения, обычно `title`), в mapping пишется статус `existing`, и кейсы перенаправляются на уже существующий target ID.

## Предусловия ✅

- [ ] gotr настроен и подключён к TestRail (`gotr self-test`)
- [ ] Известен ID исходного проекта и набора (suite)
- [ ] Известен ID целевого проекта и набора (suite)
- [ ] Целевой набор (suite) уже создан в целевом проекте
- [ ] Есть права на чтение исходного проекта и запись в целевой

## Пример: миграция между проектами 🚀

### Входные данные

| Параметр | Значение | Описание |
| --- | --- | --- |
| Исходный проект | `30` | Проект R189 |
| Исходный набор | `20069` | Набор с кейсами для переноса |
| Целевой проект | `34` | Тестирование E2E сценариев |
| Целевой набор | `19859` | Сценарии R189 (перенос) |

### Шаг 1. Разведка — проверить исходные данные

```bash
# Проверить подключение
gotr self-test

# Посмотреть shared steps исходного проекта
gotr get sharedsteps 30

# Посмотреть кейсы исходного набора
gotr export cases -p 30 -s 20069 --save --format json
```

### Шаг 2. Dry-run — увидеть план миграции

```bash
gotr sync full \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 34 \
  --dst-suite 19859 \
  --dry-run --save-filtered
```

**Что проверить:**

- Количество shared steps, которые будут перенесены
- Количество cases для миграции
- Какие shared steps отмечены как дубликаты (уже есть в целевом проекте)

### Шаг 3. Выполнить миграцию

```bash
gotr sync full \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 34 \
  --dst-suite 19859 \
  --save-mapping --approve
```

### Шаг 4. Проверить результат

```bash
# Проверить shared steps в целевом проекте
gotr get sharedsteps 34

# Проверить кейсы в целевом наборе
gotr export cases -p 34 -s 19859 --save --format json

# Сравнить проекты для верификации
gotr compare all --pid1 30 --pid2 34 --save
```

## Синтаксис 🧩

```bash
gotr sync full \
  --src-project <ID> \
  --src-suite <ID> \
  --dst-project <ID> \
  --dst-suite <ID> \
  [--compare-field <field>] \
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
| `--src-suite` | ID исходного набора | обязательный |
| `--dst-project` | ID целевого проекта | обязательный |
| `--dst-suite` | ID целевого набора | обязательный |
| `--compare-field` | Поле для поиска дубликатов | `title` |
| `--dry-run` | Показать план без изменений | `false` |
| `--save-mapping` | Сохранить mapping в файл | `false` |
| `--save-filtered` | Сохранить отфильтрованный список | `false` |
| `--approve` | Пропустить подтверждение | `false` |
| `--quiet` | Подавить служебный вывод | `false` |

## Ожидаемый результат 🧾

### Успешная миграция

- Shared steps из исходного проекта появились в целевом проекте
- Test cases созданы в целевом наборе с корректными `shared_step_id`
- Mapping-файл сохранён (если использовался `--save-mapping`)
- Команда завершилась с кодом `0`

### Артефакты

| Файл | Когда создаётся | Содержимое |
| --- | --- | --- |
| `mapping.json` | при `--save-mapping` | Соответствие старых ID shared steps → новых |
| `filtered.json` | при `--save-filtered` | Список кандидатов после фильтрации |

## Rollback миграции

`sync full` поддерживает rollback через snapshot.

### Быстрый откат сразу после миграции

В post-action меню выберите:

- `↻ Rollback this migration`

Утилита удалит созданные в target сущности в безопасном порядке зависимостей:

1. cases
2. shared steps

### Откат позже по snapshot ID

```bash
# Найти snapshot
gotr snap list

# Посмотреть детали
gotr snap info <snapshot_id>

# Превью отката без изменений
gotr snap rollback <snapshot_id> --dry-run

# Выполнить откат
gotr snap rollback <snapshot_id>
```

### Частичный rollback

Можно откатить только часть созданных target-объектов:

```bash
gotr snap rollback <snapshot_id> --entity-ids 12345,12346
```

### Важные границы rollback

- Rollback удаляет только объекты, созданные в рамках этой миграции.
- Существовавшие ранее объекты в target (дубликаты `existing`) не удаляются.
- Если часть сущностей уже удалена вручную, rollback продолжит работу и отметит частичный результат как resumable.
- Повторный запуск того же rollback повторно обработает только неуспешные/необработанные элементы.

## FAQ ❓

- ❓ **Вопрос:** Что если shared steps уже есть в целевом проекте?
  > ↪️ **Ответ:** gotr автоматически определяет дубликаты по полю `title` (или другое через `--compare-field`). Существующие шаги не дублируются, а добавляются в mapping как `existing`.
  >
  > ---

- ❓ **Вопрос:** Можно ли перенести только shared steps без cases?
  > ↪️ **Ответ:** да, используйте `gotr sync shared-steps` — см. [Миграция shared steps](migration-shared-steps.md).
  >
  > ---

- ❓ **Вопрос:** Что если целевой набор (suite) не существует?
  > ↪️ **Ответ:** создайте его заранее через `gotr add suite` или используйте `gotr sync suites` для миграции целого набора.
  >
  > ---

- ❓ **Вопрос:** Как откатить миграцию?
  > ↪️ **Ответ:** через snapshots: `gotr snap rollback <snapshot_id>` или пункт `↻ Rollback this migration` сразу после `sync full`.

---

← [Инструкции](index.md)
