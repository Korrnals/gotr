# Инструкция: Пошаговая миграция через интерактивный режим

Language: Русский | [English](../../../en/guides/instructions/migration-interactive-walkthrough.md)

## Навигация

- [Документация](../../index.md)
  - [Гайды](../index.md)
    - [Установка](../installation.md)
    - [Конфигурация](../configuration.md)
    - [Интерактивный режим](../interactive-mode.md)
    - [Прогресс](../progress.md)
    - [Каталог команд](../commands/index.md)
    - [Инструкции](index.md)
      - **Пошаговая интерактивная миграция**
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

Эта инструкция описывает **все вариации миграций** через интерактивный режим gotr.
Каждый сценарий — пошаговый walkthrough: что вводит пользователь, что показывает утилита,
какие промпты появляются и в каком порядке.

В конце приведён **реальный пример задачи** — перенос shared steps и кейсов из проекта R189
в проект «Тестирование E2E сценариев».

> [!TIP]
> Все миграционные команды поддерживают гибридный режим: часть параметров можно задать
> флагами, а остальные — выбрать интерактивно. Подробности — в разделе
> [Интерактивный режим](../interactive-mode.md).

## Предусловия ✅

- [ ] gotr настроен и подключён к TestRail (`gotr self-test`)
- [ ] Есть права на чтение исходного проекта и запись в целевой
- [ ] Целевой проект существует (или будет создан через `gotr add project`)

---

## Как работает перенос shared steps и связь с кейсами

При переносе shared steps в другой проект TestRail присваивает им **новые ID**.
Если просто скопировать кейсы «как есть» — ссылки на shared steps внутри шагов
кейсов будут указывать на старые (несуществующие в целевом проекте) ID.

gotr решает эту проблему через **автоматический маппинг ID**.

### Алгоритм (две фазы)

```text
┌─────────────────────────────────────────────────────────────────┐
│ Фаза 1: MigrateSharedSteps                                     │
│                                                                 │
│  1. Загрузить shared steps из SOURCE и DESTINATION              │
│  2. Сравнить по title (или другому полю)                        │
│     ├─ Совпадение найдено → status: "existing"                  │
│     │   mapping: old_id → existing_target_id                    │
│     └─ Совпадения нет → создать в DESTINATION                   │
│         mapping: old_id → new_created_id (status: "created")    │
│  3. Результат: таблица маппинга в памяти                        │
│                                                                 │
│  Пример маппинга:                                               │
│  ┌──────────┬──────────┬──────────┐                             │
│  │ Source ID │ Target ID│ Status   │                             │
│  ├──────────┼──────────┼──────────┤                             │
│  │ 1001     │ 2001     │ created  │                             │
│  │ 1002     │ 2002     │ created  │                             │
│  │ 1003     │ 1850     │ existing │                             │
│  └──────────┴──────────┴──────────┘                             │
├─────────────────────────────────────────────────────────────────┤
│ Фаза 2: MigrateCases                                           │
│                                                                 │
│  Для каждого кейса:                                             │
│    Для каждого шага (CustomStepsSeparated):                     │
│      Если SharedStepID ≠ 0:                                     │
│        → Найти new_id в маппинге по old SharedStepID            │
│        → Подставить new_id                                      │
│    Создать кейс в DESTINATION с обновлёнными ID                 │
│                                                                 │
│  До:  Step { SharedStepID: 1001 }  ← старый ID                 │
│  После: Step { SharedStepID: 2001 }  ← новый ID в целевом      │
└─────────────────────────────────────────────────────────────────┘
```

### Обнаружение дубликатов

Перед импортом gotr проверяет, есть ли в целевом проекте shared step
с таким же **title** (поле сравнения настраивается). Если найден —
шаг не создаётся повторно, а его ID добавляется в маппинг как `"existing"`.
Это позволяет корректно связать кейсы даже с уже существующими shared steps.

### Фильтрация по набору (suite)

Если указан source suite — переносятся только те shared steps, которые
**фактически используются** кейсами этого набора. Неиспользуемые — пропускаются.
Это экономит время и не засоряет целевой проект.

### Сохранение маппинга

При двухшаговой миграции (вариация 2) маппинг сохраняется в файл:

```json
{
  "src_project_id": 30,
  "dst_project_id": 34,
  "created_at": "2026-04-18T14:30:45Z",
  "pairs": [
    {"source_id": 1001, "target_id": 2001, "status": "created"},
    {"source_id": 1003, "target_id": 1850, "status": "existing"}
  ]
}
```

Затем передаётся в `sync cases --mapping-file ...` для подстановки ID.

> [!WARNING]
> Если mapping-файл не передан и маппинг не построен (например, cases мигрируются
> отдельно без предшествующего sync shared-steps) — ссылки на shared steps
> в кейсах **останутся со старыми ID**, что создаст «мёртвые» ссылки.

### При `sync full` — всё автоматически

В режиме `sync full` обе фазы выполняются последовательно в одном процессе.
Маппинг строится в памяти на фазе 1 и сразу используется на фазе 2 —
**никаких дополнительных действий от пользователя не требуется**.

---

## Вариация 1: Полная миграция одной командой (`sync full`)

**Когда использовать:** нужно перенести и shared steps, и test cases за один проход.

### Шаг 1. Запуск

```bash
gotr sync full
```

### Шаг 2. Интерактивные промпты

```text
? Select SOURCE project:
  ← Back
  ✕ Exit
  ──────────────────────
  ID: 1  | SAP Hybris
  ID: 2  | SAP CRM
  ...
  ID: 30 | R189
→ Выбираем: R189
```

```text
? Select SOURCE suite:
  ← Back
  ✕ Exit
  ──────────────────────
  ID: 8411  | R189 ИТ Наборы и кейсы
  ID: 9709  | R189 ПТ Наборы и кейсы
  ...
  ID: 20069 | Временный набор кейсов
→ Выбираем: Временный набор кейсов
```

```text
? Select DESTINATION project:
  ← Back                          ← вернёт к выбору source suite
  ✕ Exit
  ──────────────────────
  ID: 34 | Тестирование E2E сценариев
  ...
→ Выбираем: Тестирование E2E сценариев
```

```text
? Select DESTINATION suite:
  ← Back
  ✕ Exit
  ──────────────────────
  ID: 19859 | Сценарии R189 (перенос)
  ...
→ Выбираем: Сценарии R189 (перенос)
```

### Шаг 3. Snapshot и подтверждение

```text
? 📦 Create snapshot before migration? (recommended) [Y/n] Y
? 🏷  Snapshot label (optional, press Enter to skip): R189 → E2E full migration

  ┌────────────────────────────────────────────┐
  │ Migration Plan                             │
  │ Source:      R189 / Временный набор кейсов │
  │ Destination: E2E / Сценарии R189 (перенос) │
  │ Shared steps: 12 new, 3 existing           │
  │ Cases: 47 to migrate                       │
  └────────────────────────────────────────────┘

? Continue? [y/N] y
```

### Шаг 4. Результат и post-action меню

```text
✓ Shared steps: 12 created, 3 mapped as existing
✓ Cases: 47 created in destination suite

? What next?
  ✕ Exit
  ↻ Rollback this migration
  📊 Compare: verify current state
  🔄 Sync: migrate data
  📦 Snap: manage snapshots
→ Выбираем: 📊 Compare — чтобы верифицировать результат
```

### Гибридный вариант (часть флагов задана)

```bash
# Source задан, destination выбираем интерактивно
gotr sync full --src-project 30 --src-suite 20069

# Всё задано, только подтверждение интерактивно
gotr sync full --src-project 30 --src-suite 20069 --dst-project 34 --dst-suite 19859

# Полностью автоматически (без промптов)
gotr sync full --src-project 30 --src-suite 20069 --dst-project 34 --dst-suite 19859 --approve --save-mapping
```

---

## Вариация 2: Двухшаговая миграция (shared steps → cases)

**Когда использовать:** нужен контроль над каждым этапом. Например, после переноса
shared steps хочется проверить mapping перед переносом кейсов.

### Шаг A: Перенос shared steps

```bash
gotr sync shared-steps
```

```text
? Select SOURCE project (copy shared steps from):
→ ID: 30 | R189

? Specify source suite? [y/N] y
  (без указания suite — переносятся ВСЕ shared steps проекта)

? Select SOURCE suite:
→ ID: 20069 | Временный набор кейсов

? Select DESTINATION project (copy shared steps to):
→ ID: 34 | Тестирование E2E сценариев

  ┌──────────────────────────────────────────┐
  │ Shared Steps Summary                     │
  │ Total in source project: 45              │
  │ Used by cases in suite 20069: 15         │
  │ Already exist in destination: 3          │
  │ To import: 12                            │
  └──────────────────────────────────────────┘

? 📦 Create snapshot before migration? (recommended) [Y/n] Y
? Continue? [y/N] y

✓ Shared steps: 12 created, 3 mapped as existing

? Save mapping? [y/N] y
  → Saved: shared_steps_mapping_2026-04-18_14-30-00.json

? Save filtered shared steps list? [y/N] y
  → Saved: shared_steps_filtered_2026-04-18_14-30-00.json
```

### Шаг B: Перенос cases с mapping

```bash
gotr sync cases --mapping-file shared_steps_mapping_2026-04-18_14-30-00.json
```

```text
? Select SOURCE project (copy from):
→ ID: 30 | R189

? Select SOURCE suite:
→ ID: 20069 | Временный набор кейсов

? Select DESTINATION project (copy to):
→ ID: 34 | Тестирование E2E сценариев

? Select DESTINATION suite:
→ ID: 19859 | Сценарии R189 (перенос)

  ┌────────────────────────────────────────────┐
  │ Cases Migration Plan                       │
  │ Cases to migrate: 47                       │
  │ With shared_step_id replacement: 15        │
  │ Mapping file: shared_steps_mapping_....json│
  └────────────────────────────────────────────┘

? 📦 Create snapshot before migration? (recommended) [Y/n] Y
? Continue? [y/N] y

✓ Cases: 47 created (15 with remapped shared_step_id)
```

---

## Вариация 3: Перенос ВСЕХ shared steps проекта (без фильтрации)

**Когда использовать:** нужно перенести все общие шаги, не ограничиваясь конкретным набором.

```bash
gotr sync shared-steps
```

```text
? Select SOURCE project (copy shared steps from):
→ ID: 30 | R189

? Specify source suite? [y/N] n
  (без suite → все shared steps проекта)

? Select DESTINATION project (copy shared steps to):
→ ID: 34 | Тестирование E2E сценариев

  ┌──────────────────────────────────────┐
  │ Shared Steps Summary                 │
  │ Total in source project: 45          │
  │ Already exist in destination: 8      │
  │ To import: 37                        │
  └──────────────────────────────────────┘

? Continue? [y/N] y

✓ Shared steps: 37 created, 8 mapped as existing
```

---

## Вариация 4: Миграция структуры (suites → sections)

**Когда использовать:** перед переносом кейсов нужно создать наборы и секции
в целевом проекте.

### Шаг A: Перенос наборов (suites)

```bash
gotr sync suites
```

```text
? Select SOURCE project:
→ ID: 30 | R189

? Select DESTINATION project:
→ ID: 34 | Тестирование E2E сценариев

  Suites to migrate: 3 (2 new, 1 existing)

? 📦 Create snapshot before migration? (recommended) [Y/n] Y
? Continue? [y/N] y

✓ Suites: 2 created, 1 mapped as existing

? Save mapping? [y/N] y
  → Saved: suites_mapping_2026-04-18_14-35-00.json
```

### Шаг B: Перенос секций (sections)

```bash
gotr sync sections
```

```text
? Select SOURCE project:
→ ID: 30 | R189

? Select SOURCE suite:
→ ID: 20069 | Временный набор кейсов

? Select DESTINATION project:
→ ID: 34 | Тестирование E2E сценариев

? Select DESTINATION suite:
→ ID: 19859 | Сценарии R189 (перенос)

  Sections to migrate: 8 (6 new, 2 existing)

? 📦 Create snapshot before migration? (recommended) [Y/n] Y
? Continue? [y/N] y

✓ Sections: 6 created (parent-child hierarchy preserved)
```

---

## Вариация 5: Полный пайплайн (структура → shared steps → cases)

**Когда использовать:** полная миграция с максимальным контролем на каждом этапе.
Порядок: suites → sections → shared steps → cases.

```bash
# 1. Перенести наборы
gotr sync suites --src-project 30 --dst-project 34 --save-mapping --approve

# 2. Перенести секции
gotr sync sections --src-project 30 --src-suite 20069 --dst-project 34 --dst-suite 19859 --save-mapping --approve

# 3. Перенести shared steps (с фильтрацией по набору)
gotr sync shared-steps --src-project 30 --src-suite 20069 --dst-project 34 --save-mapping --approve

# 4. Перенести cases (с mapping от шага 3)
gotr sync cases --src-project 30 --src-suite 20069 --dst-project 34 --dst-suite 19859 --mapping-file shared_steps_mapping_*.json

# 5. Верифицировать результат
gotr compare all --pid1 30 --pid2 34 --save
```

---

## Вариация 6: Разведка → сравнение → миграция (кросс-навигация)

**Когда использовать:** сначала нужно оценить расхождения, потом мигрировать.
Кросс-навигация позволяет перейти от compare к sync без выхода из утилиты.

```bash
gotr compare all
```

```text
? Select first project (pid1):
→ ID: 30 | R189

? Select second project (pid2):
→ ID: 34 | Тестирование E2E сценариев

  ╔══════════════════════════════════════════╗
  ║ Compare All Results                      ║
  ╠══════════════════════════════════════════╣
  ║ Cases:       147 vs 100 (+47 missing)    ║
  ║ Shared steps: 45 vs 33 (+12 missing)    ║
  ║ Suites:       10 vs 10 (match)          ║
  ║ Sections:     24 vs 18 (+6 missing)     ║
  ╚══════════════════════════════════════════╝

? Compare all complete. What next?
  ✕ Exit
  🔍 Drill-down: view resource details
  💾 Save results to file
  📊 Compare: verify current state
  🔄 Sync: migrate data
  📦 Snap: manage snapshots
→ Выбираем: 🔄 Sync: migrate data
```

```text
? What do you want to migrate?
  Full migration (cases + shared steps)    ★
  Suites
  Sections
  Shared steps
→ Выбираем: Full migration

# Автоматически подставляются: --src-project 30 --dst-project 34
# Осталось выбрать только suite:

? Select SOURCE suite:
→ ID: 20069 | Временный набор кейсов

? Select DESTINATION suite:
→ ID: 19859 | Сценарии R189 (перенос)

? 📦 Create snapshot before migration? (recommended) [Y/n] Y
? Continue? [y/N] y

✓ Migration complete
```

---

## Вариация 7: Dry-run перед миграцией

**Когда использовать:** всегда рекомендуется перед реальной миграцией.

```bash
gotr sync full --dry-run
```

Промпты те же самые (выбор проектов и suite), но:

- Никаких данных не создаётся
- Показывается полный план миграции
- Нет промпта `Continue?`
- Нет post-action меню с rollback

```text
? Select SOURCE project:
→ ID: 30 | R189

? Select SOURCE suite:
→ ID: 20069 | Временный набор кейсов

? Select DESTINATION project:
→ ID: 34 | Тестирование E2E сценариев

? Select DESTINATION suite:
→ ID: 19859 | Сценарии R189 (перенос)

  [DRY-RUN] Migration Plan:
  Shared steps: 12 would be created, 3 existing
  Cases: 47 would be migrated
  No changes were made.
```

---

## Практический пример: задача миграции R189 → E2E

Реальная задача с пятью шагами. Демонстрирует полный workflow через интерактивный режим.

### Исходные данные

| Элемент | ID | Описание |
| --- | --- | --- |
| Исходный проект | 30 | R189 |
| Исходный набор | S20069 | Временный набор кейсов с общими шагами для переноса |
| Целевой проект | 34 | Тестирование E2E сценариев |
| Целевой набор | S19859 | Сценарии R189 (перенос из одноименного проекта) |

### Задача

1. Разобраться с общими тестовыми шагами проекта R189 — экспорт/анализ
2. Отфильтровать shared steps по привязке к кейсам набора S20069
3. Импортировать отфильтрованные shared steps в целевой проект
4. Экспортировать кейсы набора S20069
5. Импортировать кейсы в целевой набор S19859

### Шаг 1. Разведка — экспорт и анализ shared steps

```bash
gotr get sharedsteps
```

```text
? Select project:
→ ID: 30 | R189

[Таблица со всеми shared steps проекта R189]
```

Для сохранения в файл:

```bash
gotr export sharedsteps -p 30 --save --format json
# → Saved: ~/.gotr/exports/export/sharedsteps_30_2026-04-18_15-00-00.json
```

### Шаг 2. Экспорт кейсов набора для анализа

```bash
gotr export cases -p 30 -s 20069 --save --format json
# → Saved: ~/.gotr/exports/export/cases_30_20069_2026-04-18_15-01-00.json
```

Или интерактивно:

```bash
gotr export
```

```text
? Select export resource:
→ cases

? Select export endpoint:
→ get_cases

? Enter main ID:
→ 30

# Добавляем --suite-id 20069 --save для сохранения
```

### Шаг 3. Перенос shared steps с фильтрацией

```bash
gotr sync shared-steps
```

```text
? Select SOURCE project (copy shared steps from):
→ ID: 30 | R189

? Specify source suite? [y/N] y

? Select SOURCE suite:
→ ID: 20069 | Временный набор кейсов

? Select DESTINATION project (copy shared steps to):
→ ID: 34 | Тестирование E2E сценариев

  Shared Steps Summary:
  Total in R189: 45
  Used by cases in suite 20069: 15
  Already in E2E: 3 (by title match)
  To import: 12

? 📦 Create snapshot before migration? (recommended) [Y/n] Y
? 🏷  Snapshot label: shared-steps R189→E2E

? Continue? [y/N] y

✓ 12 shared steps created, 3 mapped as existing

? Save mapping? [y/N] y
  → shared_steps_mapping_2026-04-18_15-05-00.json

? Save filtered shared steps list? [y/N] y
  → shared_steps_filtered_2026-04-18_15-05-00.json
```

> [!IMPORTANT]
> Mapping-файл понадобится на следующем шаге для замены `shared_step_id` в кейсах.

### Шаг 4. Перенос кейсов с подстановкой shared_step_id

```bash
gotr sync cases --mapping-file shared_steps_mapping_2026-04-18_15-05-00.json
```

```text
? Select SOURCE project (copy from):
→ ID: 30 | R189

? Select SOURCE suite:
→ ID: 20069 | Временный набор кейсов

? Select DESTINATION project (copy to):
→ ID: 34 | Тестирование E2E сценариев

? Select DESTINATION suite:
→ ID: 19859 | Сценарии R189 (перенос)

  Cases Migration Plan:
  Cases to migrate: 47
  With shared_step_id replacement: 15 (using mapping file)
  Without shared steps: 32

? 📦 Create snapshot before migration? (recommended) [Y/n] Y
? 🏷  Snapshot label: cases R189→E2E

? Continue? [y/N] y

✓ 47 cases created (15 with remapped shared_step_id)
```

### Шаг 5. Верификация результата

```bash
gotr compare all --pid1 30 --pid2 34
```

```text
  Compare All Results:
  Cases:        147 vs 147 (match for suite 20069 → 19859)
  Shared steps:  45 vs 45  (match)
```

Или через post-action меню:

```text
? What next?
→ 📊 Compare: verify current state

# Проекты подставляются автоматически (30 и 34)
```

### Альтернатива: всё за одну команду

Если не нужен пошаговый контроль — та же задача решается одной командой:

```bash
gotr sync full
# → Выбираем R189 → S20069 → E2E → S19859
# → Подтверждаем → Готово
```

Или полностью неинтерактивно:

```bash
gotr sync full \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 34 \
  --dst-suite 19859 \
  --approve --save-mapping
```

---

## Карта вариаций миграции

| Вариация | Команда(ы) | Shared steps | Cases | Sections | Суть |
| --- | --- | --- | --- | --- | --- |
| Полная за 1 проход | `sync full` | ✓ авто | ✓ авто | — | Самый простой путь |
| Двухшаговая | `sync shared-steps` → `sync cases` | ✓ ручной | ✓ с mapping | — | Контроль каждого этапа |
| Только shared steps | `sync shared-steps` | ✓ | — | — | Подготовка mapping |
| Только cases | `sync cases` | — | ✓ (с/без mapping) | — | Кейсы без shared steps |
| Полный пайплайн | `suites` → `sections` → `shared-steps` → `cases` | ✓ | ✓ | ✓ | Максимальный контроль |
| Разведка → миграция | `compare all` → кросс-навигация → `sync` | зависит | зависит | зависит | Сначала анализ |

## Советы

- **Всегда начинайте с `--dry-run`** — увидите план без внесения изменений
- **Сохраняйте mapping** (`--save-mapping`) — пригодится для повторных миграций
- **Используйте snapshots** — позволяют откатить миграцию через post-action меню
- **Кросс-навигация** — после compare можно сразу перейти в sync, проекты подставятся
- **Гибридный режим** — задайте известные ID флагами, остальное выберите интерактивно
- **`--non-interactive`** — для CI/CD и скриптов; все параметры должны быть заданы флагами

---

← [Инструкции](index.md) · [Полная миграция](migration-full.md) · [Интерактивный режим](../interactive-mode.md)
