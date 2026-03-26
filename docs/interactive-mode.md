# Интерактивный режим

`gotr` поддерживает интерактивный режим для удобной работы без необходимости запоминать ID проектов и сьютов.

## Как это работает

Если обязательный параметр не указан, утилита автоматически:

1. Получает список доступных сущностей из API
2. Показывает нумерованный список
3. Просит выбрать номер
4. Использует выбранное значение

## Команды с интерактивным режимом

### get cases

```bash
# Полностью интерактивно
gotr get cases
# → Показывает список проектов → выбираем проект
# → Показывает список сьютов → выбираем сьют
# → Получаем кейсы

# Частично интерактивно
gotr get cases 30
# → Проект указан, показывает только список сьютов
```

**Особенности:**

- Если в проекте только один сьют — выбирается автоматически
- Флаг `--all-suites` отменяет выбор (получает из всех сьютов)

### get suites

```bash
gotr get suites
# → Показывает список проектов → выводит сьюты выбранного
```

### get sharedsteps

```bash
gotr get sharedsteps
# → Показывает список проектов → выводит shared steps
```

### sync cases

```bash
gotr sync cases
# → Source проект → Source сьют
# → Destination проект → Destination сьют
# → Подтверждение миграции
```

### sync shared-steps

```bash
gotr sync shared-steps
# → Source проект → (опционально) Source сьют
# → Destination проект
```

### sync sections

```bash
gotr sync sections
# → Source проект → Source сьют
# → Destination проект → Destination сьют
```

### sync full

```bash
gotr sync full
# → Source проект → Source сьют
# → Destination проект → Destination сьют
# → Миграция shared steps + cases
```

## Примеры интерактивной работы

```bash
$ gotr get cases
Доступные проекты:
----------------------------------------------------------------------
  [1] ID: 1 | SAP Hybris
  [2] ID: 2 | SAP CRM
  ...
  [17] ID: 30 | R189
----------------------------------------------------------------------
Выберите номер проекта (1-28): 17

В проекте найдено несколько сьютов:
------------------------------------------------------------
  [1] ID: 8411 | R189 ИТ Наборы и кейсы
  [2] ID: 9709 | R189 ПТ Наборы и кейсы
  ...
  [10] ID: 20069 | Временный набор кейсов
------------------------------------------------------------
Выберите номер сьюта (1-10): 10

[JSON с кейсами]
```

## Преимущества

1. **Не нужно запоминать ID** — выбирайте из списка
2. **Визуальный контроль** — видите названия проектов и сьютов
3. **Гибкость** — можно смешивать: часть параметров через флаги, часть интерактивно
4. **Автоматизация** — все те же команды работают в скриптах с флагами

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
| `add result`, `add result-for-case`, `add attachment` | Нет | Да | N/A |
| `update project` | Да (wizard) | Да | Ошибка при попытке wizard |
| `update suite` | Да (wizard) | Да | Ошибка при попытке wizard |
| `update section` | Да (wizard) | Да | Ошибка при попытке wizard |
| `update case` | Да (wizard) | Да | Ошибка при попытке wizard |
| `update run` | Да (wizard) | Да | Ошибка при попытке wizard |
| `update shared-step` | Да (wizard) | Да | Ошибка при попытке wizard |
| `update labels` | Нет | Да | N/A |
| `get cases` | Да (project/suite select) | Да | Ошибка если нужен выбор |
| `get suites`, `get sharedsteps` | Да (project select) | Да | Ошибка если нужен выбор |
| `run list`, `run get`, `run delete`, `run close`, `run update` | Да (project/run select при отсутствии run-id) | Да | Ошибка если нужен выбор |
| `result list`, `result get`, `result get-case` | Да (project/run/test-case select при отсутствии ID) | Да | Ошибка если нужен выбор |
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
| `bdds` | Да | Нет | Нет | Manual-only |
| `cases` | Да | Да | Частично | Auto есть в `cases list`, `cases get`, `cases delete`, `cases update`, `cases add`, `cases bulk`; часть веток остаётся manual-only |
| `compare` | Да | Нет | Нет | Manual-only |
| `configurations` | Да | Нет | Нет | Manual-only |
| `datasets` | Да | Да | Высокое | Auto есть в `list`, `add`, `get`, `update`, `delete` через выбор project/dataset |
| `get` | Да | Да | Частично | Auto есть в `cases`, `case`, `suites`, `suite`, `sharedsteps`, `sharedstep`, `case-history`, `sharedstep-history`, `project`, `sections list`, `section`; остальные ветки manual-only |
| `groups` | Да | Да | Высокое | Auto есть в `list`, `get`, `add`, `update`, `delete` через выбор project/group |
| `labels` | Да | Нет | Нет | Manual-only |
| `milestones` | Да | Да | Высокое | Auto есть в `list`, `get`, `add`, `update`, `delete` через выбор project/milestone |
| `plans` | Да | Да | Высокое | Auto есть в `list`, `get`, `add`, `update`, `delete`, `close`, `entry add` через выбор project/plan |
| `result` | Да | Да | Частично | Auto есть в `result list`, `result get`, `result get-case`; mutating/fields ветки остаются manual-only |
| `run` | Да | Да | Частично | Auto есть в `run list`, `run get`, `run delete`, `run close`, `run update`; прочие ветки manual-only |
| `sync` | Да | Да | Высокое | Интерактивные цепочки выбора/confirm в основных сценариях синхронизации |
| `test` | Да | Нет | Нет | Manual-only |
| `variables` | Да | Нет | Нет | Manual-only |
| `reports` | Да | Нет | Нет | Manual-only |
| `roles` | Да | Нет | Нет | Manual-only |
| `templates` | Да | Нет | Нет | Manual-only |
| `tests` | Да | Нет | Нет | Manual-only |
| `users` | Да | Нет | Нет | Manual-only |
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
| `cmd-root/add` | Частично/Да/Частично | Да/Да/Да | P0 | Добивка NI в новых prompt-точках и единые ошибки |
| `cmd-root/update` | Частично/Да/Частично | Да/Да/Да | P0 | Добивка NI и выравнивание wizard contract |
| `cmd-root/delete` | Да/Да/Да | Да/Да/Да | P0 | Auto-select endpoint/id + NI guard реализованы |
| `cmd-root/list` | Да/Да/Да | Да/Да/Да | P1 | Auto-select resource + NI guard реализованы |
| `cmd-root/export` | Да/Да/Да | Да/Да/Да | P0 | Auto-select resource/endpoint/id + NI guard реализованы |
| `get` | Частично/Да/Да | Да/Да/Да | P0 | Расширены `project`, `sections list`, `section`, `suite`, `sharedstep`, `case`, `case-history`, `sharedstep-history`; оставшиеся manual-only ветки ещё добиваются |
| `run` | Частично/Да/Да | Да/Да/Да | P0 | Закрыты `list/get/delete/close/update`; remaining manual-only ветки отдельно |
| `result` | Частично/Да/Да | Да/Да/Да | P0 | Закрыты `list/get/get-case`; remaining manual-only ветки отдельно |
| `sync` | Да/Да/Да | Да/Да/Да | P0 | NI закрыт на select/confirm точках (`cases`, `suites`, `sections`, `shared-steps`, `full`) |
| `attachments` | Да/Да/Да | Да/Да/Да | P1 | Закрыты `list case/plan/plan-entry/run/test`, `get`, `delete`, `add case/plan/plan-entry/result/run` |
| `bdds` | Нет/Да/N/A | Частично/Да/Да | P2 | Минимальный выбор case_id в read/mutate ветках |
| `cases` | Частично/Да/Да | Частично/Да/Да | P1 | Закрыты `list/get/delete/update/add/bulk`; next: оставшиеся manual-only ветки |
| `compare` | Нет/Да/N/A | Частично/Да/Да | P2 | Interactive presets для source/destination |
| `configurations` | Нет/Да/N/A | Частично/Да/Да | P2 | Select project/group/config |
| `datasets` | Да/Да/Да | Да/Да/Да | P1 | Закрыты `list/add/get/update/delete` с project/dataset select + NI guard |
| `groups` | Да/Да/Да | Да/Да/Да | P1 | Закрыты `list/get/add/update/delete` с project/group select + NI guard |
| `labels` | Нет/Да/N/A | Частично/Да/Да | P2 | Select project/test/label |
| `milestones` | Да/Да/Да | Да/Да/Да | P1 | Закрыты `list/get/add/update/delete` с project/milestone select + NI guard |
| `plans` | Да/Да/Да | Да/Да/Да | P1 | Закрыты `list/get/add/update/delete/close/entry add` с project/plan select + NI guard |
| `reports` | Нет/Да/N/A | Частично/Да/Да | P2 | Select project/template |
| `roles` | Нет/Да/N/A | Нет/Да/N/A | P2 | Справочник, manual-only допустим |
| `templates` | Нет/Да/N/A | Частично/Да/Да | P2 | Select project |
| `test` | Нет/Да/N/A | Частично/Да/Да | P1 | Select run/test в read ветках |
| `tests` | Нет/Да/N/A | Частично/Да/Да | P1 | Select run/test в list/get/update |
| `users` | Нет/Да/N/A | Частично/Да/Да | P2 | Select user/project для list/get/update |
| `variables` | Нет/Да/N/A | Частично/Да/Да | P1 | Select dataset/variable |

Отдельно по `roles`:

- пакет может оставаться manual-only по `Auto` (это reference directory),
если не требуется UX-унификация с выбором сущностей.

- `NI` для `roles` не обязателен, пока в пакете нет prompts.

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
