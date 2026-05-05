# Интерактивный режим

Language: Русский | [English](../../en/guides/interactive-mode.md)

## Навигация

- [Документация](../index.md)
  - [Гайды](index.md)
    - [Установка](installation.md)
    - [Конфигурация](configuration.md)
    - [Интерактивный режим](interactive-mode.md)
    - [Прогресс](progress.md)
    - [Каталог команд](commands/index.md)
    - [Инструкции](instructions/index.md)
  - [Архитектура](../architecture/index.md)
  - [Эксплуатация](../operations/index.md)
  - [Отчёты](../reports/index.md)
- [Главная](../../../README_ru.md)

## Как это работает

Если обязательный параметр не указан, утилита автоматически:

1. Получает список доступных сущностей из API
2. Показывает интерактивное меню выбора с навигацией
3. Просит выбрать элемент из списка
4. Использует выбранное значение и переходит к следующему шагу

Три режима работы:

- **Auto-interactive** (по умолчанию) — промпты появляются только для неуказанных параметров
- **Manual** — все параметры через флаги, промптов нет
- **Non-interactive** (`--non-interactive`) — промпты запрещены; ошибка если требуется ввод (для CI/CD)

## Навигация в интерактивных меню

Все интерактивные списки содержат элементы навигации:

```text
? Select project:
  ← Back                    ← вернуться к предыдущему шагу
  ✕ Exit                    ← выйти из интерактива
  ──────────────────────
  ID: 1  | SAP Hybris
  ID: 2  | SAP CRM
  ID: 30 | R189
  ...
  ← Back                    ← дублируется внизу списка (если >5 элементов)
```

- **← Back** — возврат к предыдущему шагу выбора (если доступен)
- **✕ Exit** — завершение работы без ошибки
- Элементы отображаются с выравниванием колонок (ID + название)
- Если в проекте **один suite** — он выбирается автоматически (без промпта)

## Команды с интерактивным режимом

### get — чтение данных

```bash
# Полностью интерактивно
gotr get cases
# → Select project: → Select suite: → [JSON с кейсами]

# Частично интерактивно (проект указан)
gotr get cases 30
# → Select suite: → [JSON с кейсами]

# Полностью ручной
gotr get cases 30 --suite-id 20069
```

Интерактивный режим доступен для:

| Команда | Промпт 1 | Промпт 2 | Промпт 3 |
| --- | --- | --- | --- |
| `get cases` | Select project | Select suite | — |
| `get case` | Select project | Select suite | Select case |
| `get suites` | Select project | — | — |
| `get suite` | Select project | Select suite | — |
| `get sharedsteps` | Select project | — | — |
| `get sharedstep` | Select project | Select shared step | — |
| `get case-history` | Select project | Select suite → Select case | — |
| `get sharedstep-history` | Select project | Select shared step | — |
| `get project` | Select project | — | — |
| `get sections list` | Select project | — | — |

Особенности `get cases`:

- Если в проекте **один suite** — выбирается автоматически
- Флаг `--all-suites` загружает кейсы из всех наборов (без выбора)

### export — экспорт данных

```bash
# Полностью интерактивно
gotr export
# → Select export resource: → Select export endpoint: → [результат]

# Частично
gotr export cases get_cases 30 --suite-id 20069 --save --format json
```

Промпты export (каждый следующий — только если не указан):

1. `Select export resource:` — выбор типа ресурса (cases, suites, sharedsteps...)
2. `Select export endpoint:` — выбор API-эндпоинта
3. `Enter main ID:` — ввод ID (если эндпоинт содержит `{id}`)

### compare — сравнение проектов

```bash
# Полностью интерактивно
gotr compare cases
# → Select first project (pid1): → Select second project (pid2):
# → [результат сравнения]
# → Comparison complete. What next?

# Ручной
gotr compare all --pid1 30 --pid2 34 --save
```

Промпты compare:

1. `Select first project (pid1):` — выбор первого проекта
2. `Select second project (pid2):` — выбор второго (← Back возвращает к шагу 1)
3. `Save compare result to file?` — предложение сохранить (если выбран интерактивно)

### sync — миграция данных

Все sync-подкоманды поддерживают полный интерактивный режим.

#### sync full

```bash
gotr sync full
# → Select SOURCE project:
# → Select SOURCE suite:
# → Select DESTINATION project:
# → Select DESTINATION suite:
# → 📦 Create snapshot before migration? (recommended) [Y/n]
# → 🏷  Snapshot label (optional, press Enter to skip):
# → [сводка миграции]
# → Continue? [y/N]
# → [выполнение: shared steps → cases]
# → [post-action меню]
```

#### sync cases

```bash
gotr sync cases
# → Select SOURCE project (copy from):
# → Select SOURCE suite:
# → Select DESTINATION project (copy to):
# → Select DESTINATION suite:
# → 📦 Create snapshot before migration? (recommended) [Y/n]
# → Continue? [y/N]
# → [выполнение миграции кейсов]
# → [post-action меню]
```

#### sync shared-steps

```bash
gotr sync shared-steps
# → Select SOURCE project (copy shared steps from):
# → Specify source suite? [y/N]
#   (если да) → Select SOURCE suite:
# → Select DESTINATION project (copy shared steps to):
# → [фильтрация и сводка]
# → 📦 Create snapshot before migration? (recommended) [Y/n]
# → Continue? [y/N]
# → [импорт shared steps]
# → Save mapping? [y/N]
# → Save filtered shared steps list? [y/N]
# → [post-action меню]
```

#### sync sections

```bash
gotr sync sections
# → Select SOURCE project:
# → Select SOURCE suite:
# → Select DESTINATION project:
# → Select DESTINATION suite:
# → 📦 Create snapshot before migration? (recommended) [Y/n]
# → Continue? [y/N]
# → [перенос секций]
# → Save mapping? [y/N]
# → [post-action меню]
```

#### sync suites

```bash
gotr sync suites
# → Select SOURCE project:
# → Select DESTINATION project:
# → 📦 Create snapshot before migration? (recommended) [Y/n]
# → Continue? [y/N]
# → [перенос наборов]
# → Save mapping? [y/N]
# → [post-action меню]
```

### attachments cleanup — массовая очистка вложений

`gotr attachments cleanup` поддерживает интерактивные запросы, если
команда вызвана с недостаточным набором флагов на TTY. Флаги, явно
указанные в командной строке, мастер никогда не перезаписывает.

```bash
gotr attachments cleanup
# → Область: Все проекты / Конкретные проекты
#   (конкретные) → ID проектов (через запятую):
# → Пресет типов сущностей:
#     all (case,run,plan,plan_entry,result,test)  ← по умолчанию, печатает ⚠️
#     case · run · plan,plan_entry · result,test · custom (через запятую)
# → Старше (например, 90d, 3M, 1y):
# → Concurrency:
# → 📦 Создать снапшот перед удалением? (рекомендуется) [Y/n]
#   (да) → Срок хранения снапшота (дней):
# → Сначала dry-run? [Y/n]
# → Финальное подтверждение: Y/N
```

Переопределения:

- `--non-interactive` — отключает запросы; команда завершится с ошибкой,
  если требуется ввод (для CI).
- `--force` — пропускает финальное подтверждение, но не отключает
  мастер для отсутствующих значений.
- `--no-snapshot` — отказ от страховочного снапшота; откат после этого
  невозможен.

Полный список флагов и сценарий отката см.
[`gotr attachments cleanup`](commands/attachments.md).

## Post-action меню и кросс-навигация

После завершения sync- и compare-операций появляется интерактивное меню действий.

### После sync

```text
? What next?
  ✕ Exit
  ↻ Rollback this migration          ← (только если был создан snapshot)
  📊 Compare: verify current state    ← кросс-навигация → gotr compare all
  🔄 Sync: migrate data              ← кросс-навигация → gotr sync full
  📦 Snap: manage snapshots          ← кросс-навигация → gotr snap list
```

- **Rollback** — откатывает миграцию через ранее созданный snapshot
- **Кросс-навигация** — прямой переход к связанным командам без выхода

### После compare

```text
? Comparison complete. What next?
  ✕ Exit
  📋 View detailed results
  💾 Save results to file
  → Sync: migrate differences        ← (если есть расхождения)
  📊 Compare: verify current state
  🔄 Sync: migrate data
  📦 Snap: manage snapshots
```

- **Sync: migrate differences** — запускает sync с передачей project ID из compare
- **Save results to file** — выбор формата (json/yaml/csv/table) и пути сохранения

### Наследование параметров (WorkSession)

Параметры `src-project`, `dst-project`, `src-suite`, `dst-suite` передаются через сессию:

```text
compare → sync:  project ID из compare подставляются в sync автоматически
sync → compare:  project ID из sync подставляются в compare автоматически
```

Это означает что при кросс-навигации **не нужно повторно выбирать проекты** — они
наследуются из предыдущей команды.

## Snapshot-подтверждение

Перед mutating-операциями (sync) gotr предлагает создать snapshot:

```text
📦 Create snapshot before migration? (recommended) [Y/n]
```

Логика приоритетов:

1. Флаг `--snapshot` — принудительно включает/выключает
2. Параметр конфигурации `snap.enabled` — если задан
3. Интерактивный промпт — если ни (1) ни (2) не заданы (по умолчанию: **да**)

При создании snapshot предлагается ввести метку:

```text
🏷  Snapshot label (optional, press Enter to skip):
```

## Примеры интерактивной работы

### Пример 1: get cases

```text
$ gotr get cases

? Select project:
  ← Back
  ✕ Exit
  ──────────────────────
  ID: 1  | SAP Hybris
  ID: 2  | SAP CRM
  ...
  ID: 30 | R189
  ← Back

→ Выбираем: ID: 30 | R189

? Select suite:
  ← Back
  ✕ Exit
  ──────────────────────
  ID: 8411  | R189 ИТ Наборы и кейсы
  ID: 9709  | R189 ПТ Наборы и кейсы
  ...
  ID: 20069 | Временный набор кейсов
  ← Back

→ Выбираем: ID: 20069 | Временный набор кейсов

[JSON с кейсами]
```

### Пример 2: sync full (полная миграция)

```text
$ gotr sync full

? Select SOURCE project:
→ ID: 30 | R189

? Select SOURCE suite:
→ ID: 20069 | Временный набор кейсов

? Select DESTINATION project:
→ ID: 34 | Тестирование E2E сценариев

? Select DESTINATION suite:
→ ID: 19859 | Сценарии R189 (перенос)

? 📦 Create snapshot before migration? (recommended) Yes
? 🏷  Snapshot label (optional): R189 → E2E migration

  ┌─────────────────────────────────────┐
  │ Migration summary                   │
  │ Shared steps: 12 new, 3 existing    │
  │ Cases: 47 to migrate                │
  └─────────────────────────────────────┘

? Continue? Yes

✓ Shared steps migrated (12 created, 3 mapped)
✓ Cases migrated (47 created)

? What next?
→ ✕ Exit
```

### Пример 3: compare → sync (кросс-навигация)

```text
$ gotr compare cases --pid1 30 --pid2 34

  Cases comparison:
  Project 30: 147 cases
  Project 34: 100 cases
  Differences: 47 missing in project 34

? Comparison complete. What next?
→ → Sync: migrate differences

? What do you want to migrate?
→ Full migration (cases + shared steps)

# Проекты наследуются из compare → запускается sync full
# с --src-project 30 --dst-project 34
```

## Частичный интерактив (гибридный режим)

Можно указать **часть** параметров флагами, а остальные выбрать интерактивно:

```bash
# Только source задан — destination выбираем интерактивно
gotr sync full --src-project 30 --src-suite 20069

# Только проекты заданы — suite выбираем интерактивно
gotr sync cases --src-project 30 --dst-project 34

# Mapping-файл задан, остальное — интерактивно
gotr sync cases --mapping-file mapping.json
```

## Преимущества

1. **Не нужно запоминать ID** — выбирайте из списка
2. **Визуальный контроль** — видите названия проектов и сьютов
3. **Гибкость** — можно смешивать: часть параметров через флаги, часть интерактивно
4. **Автоматизация** — все те же команды работают в скриптах с флагами
5. **Кросс-навигация** — переход между compare/sync/snap без выхода
6. **Наследование параметров** — проекты передаются между командами через сессию
7. **Snapshot safety** — автоматическое предложение создать точку отката перед миграцией

## Дорожная карта Stage 12 (Interactive System Unification)

Ниже зафиксирован объединённый план по интерактивному режиму с переходом к единой модели
`auto-interactive + --non-interactive`.

### Stage 12.0 — Foundation (выполнено)

- Введён единый контракт `Prompter`.

- Добавлены реализации: terminal prompter (survey/v2), non-interactive
prompter и mock prompter для тестов.

- Добавлена контекстная инъекция prompter в root runtime.

- Добавлен глобальный флаг `--non-interactive`.

### Stage 12.1 — Миграция команд на единый prompter (выполнено)

- Синхронизация (`sync`) переведена на `SelectProject/SelectSuiteForProject`
и `p.Confirm`.

- `get`, `run`, `result` переведены на `PrompterFromContext`.

- Тесты мигрированы с `os.Stdin` и ручных мока-обёрток на `MockPrompter`.

### Stage 12.2 — Унификация UX: auto-interactive (выполнено)

- Для `add`/`update` включён auto-wizard, если пользователь не задал
ручные input-флаги.

- Явный флаг `--interactive` сохранён для обратной совместимости.

- Режим `--non-interactive` остаётся главным переключателем для CI/CD и
automation.

### Stage 12.3 — Полный тест-аудит и добивка покрытия (новая стадия)

Цель: закрыть пробелы по тестам для всей кодовой базы (в первую очередь CLI слой и
интерактивные/безопасные сценарии), чтобы Stage 12 считался завершённым по DoD качества.

#### 12.3.1 Инвентаризация покрытия

- Построить матрицу покрытий по пакетам `cmd/*`, `internal/interactive` и
критичному runtime (`internal/service`, `internal/output`, `internal/flags`).

- Собрать список файлов без тестов, ветвей без негативных тестов и команд
без сценариев `--non-interactive`.

- Зафиксировать baseline метрики (`go test -cover ./...` + фокусный
`-coverprofile`).

#### 12.3.2 Приоритезация тест-долга

- P0 (обязательно): mutating-команды с dry-run gate и non-interactive gate,
а также интерактивные цепочки выбора (`project -> suite -> run`) и ошибки
выбора.

- P1: автопереключение auto-interactive vs ручные флаги, регрессии по
error wrapping и сообщениям.

- P2: edge cases, пустые ответы API, частичные данные.

#### 12.3.3 Реализация недостающих тестов

- Добавить table-driven тесты для повторяемых сценариев.

- Использовать `MockPrompter` как стандарт для всех интерактивных веток.

- Добавить тесты на `ErrNonInteractive` в точках, где требуется ввод,
отсутствие mutating API-вызовов в dry-run и корректный fallback на ручные
флаги без prompts.

#### 12.3.4 Верификация и критерии готовности

- Полный прогон `go test ./...` зелёный.

- Фокусные наборы (`cmd/*`, `internal/interactive`) проходят с покрытием
без регрессий.

- Все найденные P0/P1 пробелы закрыты тестами и отражены в changelog stage.

### Stage 12.4 — Cleanup и удаление legacy compatibility wrappers (запланировано)

- Удалить compatibility-обёртки в `internal/interactive/*`, которые больше
не используются.

- Обновить стандарты и примеры кода на новый API (`PrompterFromContext`,
auto-interactive).

### Stage 12.5 — Documentation и release readiness (запланировано)

- Обновить docs по интерактивному режиму и non-interactive работе.

- Подготовить release notes с картой изменений Stage 12.

- Провести финальный smoke-check CLI сценариев (manual + CI).

### Stage 12.6 — Smoke-check status (выполнено)

Финальные smoke-проверки выполнены на свежесобранном бинарнике `gotr-test` с
реальной конфигурацией и безопасным ограничением: только read-only команды или
`--dry-run` для mutating-операций.

Выполненные проверки:

- `go test ./... -count=1` — **PASS** (полный зелёный прогон).
- `gotr templates list` (интерактивный выбор проекта) — **PASS**,
  получен валидный JSON-ответ шаблонов.
- `gotr run create --name <name> --dry-run` — интерактивный выбор проекта и
  suite достигается, прежняя блокировка `required flag(s) "suite-id" not set`
  устранена.
- `gotr result add --status-id 1 --dry-run` — интерактивный выбор проекта и run
  достигается.

Примечание по live-runner:

- Для многошаговых TTY-списков при подаче ввода через `printf | ...` возможен
  `EOF` на поздних шагах выбора. Это ограничение non-TTY пайпинга, а не
  функциональная регрессия Stage 12: ключевые интерактивные ветки успешно
  активируются и проходят целевые точки выбора.

## Матрица поведения команд

Ниже отражено фактическое поведение в двух срезах:

- Top-level команды и ключевые подкоманды (операционный срез).

- Полная карта по всем подпакетам `cmd/**` (архитектурный срез слоя CLI).

Обозначения:

- **Auto**: автоматический интерактивный режим при отсутствии обязательных входных значений.
- **Manual**: ручной режим через явные флаги/аргументы без prompts.
- **NI**: `--non-interactive`, который запрещает prompts и завершает команду ошибкой при необходимости ввода.

| Команда/подкоманда | Auto | Manual | NI |
| --- | --- | --- | --- |
| `add project` | Да (wizard) | Да | Ошибка при попытке wizard |
| `add suite` | Да (wizard + auto select project) | Да | Ошибка при попытке wizard |
| `add section` | Да (wizard + auto select project) | Да | Ошибка при попытке wizard |
| `add case` | Да (wizard + auto select section) | Да | Ошибка при попытке wizard |
| `add run` | Да (wizard + auto select project) | Да | Ошибка при попытке wizard |
| `add shared-step` | Да (wizard + auto select project) | Да | Ошибка при попытке wizard |
| `add result` | Да (project/run/test select при отсутствии `test-id`) | Да | Ошибка если нужен выбор |
| `add result-for-case` | Частично (project/run select при отсутствии `run-id`; `case-id` остаётся manual) | Да | Ошибка если нужен выбор |
| `add attachment` | Частично (container IDs optional у части subcommands, `file_path` manual) | Да | Ошибка если нужен выбор |
| `update project` | Да (wizard) | Да | Ошибка при попытке wizard |
| `update suite` | Да (wizard) | Да | Ошибка при попытке wizard |
| `update section` | Да (wizard) | Да | Ошибка при попытке wizard |
| `update case` | Да (wizard) | Да | Ошибка при попытке wizard |
| `update run` | Да (wizard) | Да | Ошибка при попытке wizard |
| `update shared-step` | Да (wizard) | Да | Ошибка при попытке wizard |
| `update labels` | Частично (`labels update test` с выбором test; bulk/manual ветки частично manual-only) | Да | Ошибка если нужен выбор |
| `get cases` | Да (project/suite select) | Да | Ошибка если нужен выбор |
| `get suites`, `get sharedsteps` | Да (project select) | Да | Ошибка если нужен выбор |
| `run list`, `run get`, `run delete`, `run close`, `run update`, `run create` | Да (project/run select при отсутствии ID; `run create` также выбирает suite) | Да | Ошибка если нужен выбор |
| `plans list`, `plans get`, `plans add`, `plans update`, `plans delete`, `plans close`, `plans entry add/update/delete` | Да (project/plan/entry select при отсутствии ID) | Да | Ошибка если нужен выбор |
| `result list`, `result get`, `result get-case`, `result add` | Да (project/run/test-case select при отсутствии ID) | Да | Ошибка если нужен выбор |
| `users get`, `users update`, `users get-by-email` | Да (select из users list при отсутствии ID/email) | Да | Ошибка если нужен выбор |
| `roles get` | Да (select из roles list при отсутствии ID) | Да | Ошибка если нужен выбор |
| `reports list`, `reports run`, `reports run-cross-project`, `templates list`, `bdds add/get` | Да | Да | Ошибка если нужен выбор |
| `sync *` | Да (project/suite/select + confirm) | Да | Ошибка если нужен выбор/confirm |
| `delete` | Да (endpoint/id select) | Да | Ошибка если нужен выбор |
| `list` | Да (resource select) | Да | Ошибка если нужен выбор |
| `export` | Да (resource/endpoint/id input) | Да | Ошибка если нужен выбор |

## Полная карта `cmd/**` (все подпакеты)

Источник истины по регистрации root-команд: `cmd/commands.go`.

Обозначения:

- **Registered**: подпакет подключён в `rootCmd` через `*.Register(...)`.

- **Interactive**: наличие интерактивной логики в production-коде пакета.

- **Coverage**: уровень покрытия интерактивом внутри пакета.

| Пакет `cmd/**` | Registered | Interactive | Coverage | Комментарий |
| --- | --- | --- | --- | --- |
| `attachments` | Да | Да | Высокое | Auto есть в `attachments list case/plan/plan-entry/run/test`, `attachments get`, `attachments delete`, `attachments add case/plan/plan-entry/result/run` |
| `bdds` | Да | Да | Высокое | Auto есть в `add`, `get` через выбор case |
| `cases` | Да | Да | Частично | Auto есть в `cases list`, `cases get`, `cases delete`, `cases update`, `cases add`, `cases bulk`; часть веток остаётся manual-only |
| `compare` | Да | Нет | Нет | Manual-only |
| `configurations` | Да | Да | Высокое | Auto есть в `list`, `add-group`, `add-config`, `update-group`, `update-config`, `delete-group`, `delete-config` |
| `datasets` | Да | Да | Высокое | Auto есть в `list`, `add`, `get`, `update`, `delete` через выбор project/dataset |
| `get` | Да | Да | Частично | Auto есть в `cases`, `case`, `suites`, `suite`, `sharedsteps`, `sharedstep`, `case-history`, `sharedstep-history`, `project`, `sections list`, `section`; остальные ветки manual-only |
| `groups` | Да | Да | Высокое | Auto есть в `list`, `get`, `add`, `update`, `delete` через выбор project/group |
| `labels` | Да | Да | Частично | Auto есть в `get`, `list`, `update test`, `update-label`; часть bulk/manual веток требует ручные списки/флаги |
| `milestones` | Да | Да | Высокое | Auto есть в `list`, `get`, `add`, `update`, `delete` через выбор project/milestone |
| `plans` | Да | Да | Высокое | Auto есть в `list`, `get`, `add`, `update`, `delete`, `close`, `entry add`, `entry update`, `entry delete` через выбор project/plan/entry |
| `result` | Да | Да | Частично | Auto есть в `list`, `get`, `get-case`, `add`, `add-case`; `add-bulk`/`fields` остаются manual-oriented |
| `run` | Да | Да | Высокое | Auto есть в `list`, `get`, `delete`, `close`, `update`, `create` через выбор project/run/suite |
| `sync` | Да | Да | Высокое | Интерактивные цепочки выбора/confirm в основных сценариях синхронизации |
| `test` | Да | Нет | Нет | Manual-only |
| `variables` | Да | Нет | Нет | Manual-only |
| `reports` | Да | Да | Высокое | Auto есть в `list`, `run`, `run-cross-project` |
| `roles` | Да | Да | Частичное | Auto есть в `get`; `list` read-only/manual |
| `templates` | Да | Да | Частичное | Auto есть в `list` через выбор project |
| `tests` | Да | Да | Высокое | Auto есть в `list`, `get`, `update` через выбор run/test |
| `users` | Да | Да | Частичное | Auto есть в `get`, `update`, `get-by-email`; `list`/`add` manual/read-only |
| `list` | Нет | Нет | Нет | Служебная директория; standalone Register отсутствует |
| `internal` | Нет | Нет | Нет | Вспомогательные тестовые утилиты, не CLI-команды |

### Root-команды в `cmd/*.go`

| Команда | Interactive | Coverage | Комментарий |
| --- | --- | --- | --- |
| `add` | Да | Высокое | Wizard + auto-interactive + `--non-interactive` gate |
| `update` | Да | Высокое | Wizard + auto-interactive + `--non-interactive` gate |
| `delete` | Да | Высокое | Auto-select endpoint/id + `--non-interactive` guard |
| `list` | Да | Высокое | Auto-select resource + `--non-interactive` guard |
| `export` | Да | Частичное | Auto-select inputs; требуется добивка e2e сценариев |
| `config`, `completion`, `selftest` | Нет | N/A | Служебные команды без интерактивного выбора сущностей |

## Целевая матрица Auto/Manual/NI по `cmd/**`

Матрица ниже фиксирует не только текущее состояние, но и целевое поведение для
унификации проекта целиком.

Обозначения:

- **Current**: фактическое состояние на текущем этапе.

- **Target**: состояние после завершения унификации.

- **Priority**: порядок реализации (`P0` -> `P1` -> `P2`).

| Пакет | Current (Auto/Manual/NI) | Target (Auto/Manual/NI) | Priority | Scope |
| --- | --- | --- | --- | --- |
| `cmd-root/add` | Да/Да/Да | Да/Да/Да | P0 | Dry-run и NI gate есть; attachment/result paths различаются по глубине auto-select |
| `cmd-root/update` | Да/Да/Да | Да/Да/Да | P0 | Wizard/NI выровнены для root update и package-команд |
| `cmd-root/delete` | Да/Да/Да | Да/Да/Да | P0 | Auto-select endpoint/id + NI guard реализованы |
| `cmd-root/list` | Да/Да/Да | Да/Да/Да | P1 | Auto-select resource + NI guard реализованы |
| `cmd-root/export` | Да/Да/Да | Да/Да/Да | P0 | Auto-select resource/endpoint/id + NI guard реализованы |
| `get` | Да/Да/Да | Да/Да/Да | P0 | Основные read-only ветки покрыты auto-select; остатки без prompts являются осознанно manual/read-only |
| `run` | Да/Да/Да | Да/Да/Да | P0 | Закрыты `list/get/delete/close/update/create`; NI guard есть в интерактивных ветках |
| `result` | Частично/Да/Да | Да/Да/Да | P0 | Закрыты `list/get/get-case/add/add-case`; manual-only остаётся `add-bulk` (required file input) |
| `sync` | Да/Да/Да | Да/Да/Да | P0 | NI закрыт на select/confirm точках (`cases`, `suites`, `sections`, `shared-steps`, `full`) |
| `attachments` | Да/Да/Да | Да/Да/Да | P1 | Закрыты `list case/plan/plan-entry/run/test`, `get`, `delete`, `add case/plan/plan-entry/result/run` |
| `bdds` | Да/Да/Да | Частично/Да/Да | P2 | `add/get` поддерживают auto-select case_id + NI guard |
| `cases` | Частично/Да/Да | Частично/Да/Да | P1 | Закрыты `list/get/delete/update/add/bulk`; next: оставшиеся manual-only ветки |
| `compare` | Нет/Да/N/A | Частично/Да/Да | P2 | Interactive presets для source/destination |
| `configurations` | Да/Да/Да | Частично/Да/Да | P2 | Закрыты `list/add-group/add-config/update-group/update-config/delete-group/delete-config` |
| `datasets` | Да/Да/Да | Да/Да/Да | P1 | Закрыты `list/add/get/update/delete` с project/dataset select + NI guard |
| `groups` | Да/Да/Да | Да/Да/Да | P1 | Закрыты `list/get/add/update/delete` с project/group select + NI guard |
| `labels` | Частично/Да/Да | Частично/Да/Да | P2 | Auto есть в `get/list/update test/update-label`; `update tests` остаётся manual-only по входу |
| `milestones` | Да/Да/Да | Да/Да/Да | P1 | Закрыты `list/get/add/update/delete` с project/milestone select + NI guard |
| `plans` | Да/Да/Да | Да/Да/Да | P1 | Закрыты `list/get/add/update/delete/close/entry add/update/delete` с project/plan/entry select + NI guard |
| `reports` | Да/Да/Да | Частично/Да/Да | P2 | `list/run/run-cross-project` поддерживают auto-select + NI guard |
| `roles` | Частично/Да/Да | Нет/Да/N/A | P2 | `get` интерактивный; `list` read-only/manual |
| `templates` | Частично/Да/Да | Частично/Да/Да | P2 | `list` с auto-select project + NI guard |
| `test` | Нет/Да/N/A | Частично/Да/Да | P1 | Select run/test в read ветках |
| `tests` | Да/Да/Да | Частично/Да/Да | P1 | `list/get/update` с auto-select run/test + NI guard |
| `users` | Частично/Да/Да | Частично/Да/Да | P2 | `get/update/get-by-email` интерактивны; `list/add` остаются manual/read-only |
| `variables` | Да/Да/Да | Частично/Да/Да | P1 | `list/add/update/delete/get` с dataset/variable select + NI guard |

Отдельно по `roles`:

- пакет частично интерактивный (`roles get`), но может оставаться преимущественно
reference/manual-only без полной UX-унификации всех веток.

- `NI` для `roles get` уже обязателен и реализован (есть prompt-точка).

## Матрица dry-run (имитация без мутаций)

Для mutating-команд dry-run считается корректным только если:

- команда выполняется без mutating API-вызова;

- пользователю показывается операция-имитация (method + endpoint + body);

- есть тест, подтверждающий отсутствие реальной мутации.

| Область | Current | Target | Priority | Примечание |
| --- | --- | --- | --- | --- |
| `cmd-root/add` | Есть | Есть | P0 | Покрыт dry-run маршрутизацией и тестами |
| `cmd-root/update` | Есть | Есть | P0 | Покрыт dry-run маршрутизацией и тестами |
| `cmd-root/delete` | Есть | Есть | P0 | Добавлена интерактивная ветка; добавлены safety-тесты |
| `sync/*` | Есть | Есть | P0 | Есть dry-run флаги и тесты no-mutating |
| `run/*` mutating | Есть | Есть | P1 | В целом покрыто, требуется точечная ревизия форматов dry-run вывода |
| `result/add*` | Есть | Есть | P1 | Есть dry-run флаги и unit-тесты |
| `cases/*` mutating | Есть | Есть | P1 | Есть dry-run в add/update/delete/bulk |
| `groups/configurations/datasets/milestones/plans/variables` mutating | Есть | Есть | P1 | Есть dry-run флаги и тесты |
| `users add/update` | Есть | Есть | P0 | Добавлен dry-run + no-mutation тесты |
| `labels update-label` | Есть | Есть | P0 | Добавлен dry-run + no-mutation тест |
| `attachments delete` | Есть | Есть | P1 | Добавлен dry-run + no-mutation тест |
| `reports run*` | Есть | Есть | P2 | Добавлен dry-run + no-mutation тесты |

Наблюдение:

- В части read-only команд dry-run уже присутствует, но не обязателен по смыслу.

- Приоритет исправлений dry-run задаётся только для реальных mutating-веток.

Примечание:

- Явный флаг `--interactive` пока сохранён для обратной совместимости.
- Рекомендуемая модель использования: auto-интерактив по умолчанию и `--non-interactive` для CI/CD.

---

← [Гайды](index.md) · [Документация](../index.md)
