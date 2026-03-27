# Stage 13 - CLI Contract Matrix

Дата: 2026-03-27
Ветка: stage-13.0-final-refactoring
Статус: Audit completed

## Методология

- Источники: статический анализ кода (rg), ручные проверки.
- Критерии: покрытие quiet флага, non-interactive handling, наличие interactive helpers.
- Статусы: `ok` (полный контракт), `partial` (частичный), `risk` (отсутствует или ненормативный), `n/a`.

## Global flags (root)

- `--quiet, -q` (persistent): объявлен в cmd/commands.go:84, применяется через `ui.SetMessageQuiet(quiet)` в PersistentPreRunE.
- `--non-interactive` (persistent): объявлен в cmd/commands.go:87, применяется через `interactive.WithPrompter` в PersistentPreRunE.

## Command Group Matrix

| Command Group | Quiet status | Non-interactive status | Notes |
|---|---|---|---|
| compare | ok (10 refs) | partial (1 ref) | quiet применяется корректно; non-interactive — только в всплывающем диалоге save |
| get | ok (3 refs) | ok (18 refs) | корректное использование |
| sync | partial (1 ref) | partial (5 refs) | есть локальная helper-функция `isQuiet` вместо uses глобального seam |
| result | partial (1 ref) | ok (10 refs) | quiet проверяется только в 1 месте |
| run | риск: LOCAL flag | ok (10 refs) | cmd/run/run.go определяет `--quiet, -q` локально, создавая hidden override |
| test | partial (2 refs) | ok (8 refs) | частичная проверка quiet |
| groups | partial (1 ref) | partial (2 refs) | тонкое покрытие |
| attachments | n/a (0) | ok (9 refs) | quiet не проверяется в коде |
| bdds | n/a (0) | ok (8 refs) | quiet не проверяется в коде |
| cases | n/a (0) | partial (5 refs) | quiet не проверяется в коде |
| configurations | n/a (0) | ok (24 refs) | quiet не проверяется в коде |
| datasets | n/a (0) | partial (2 refs) | quiet не проверяется в коде |
| labels | n/a (0) | ok (17 refs) | quiet не проверяется в коде |
| milestones | n/a (0) | ok (7 refs) | quiet не проверяется в коде |
| plans | n/a (0) | ok (14 refs) | quiet не проверяется в коде |
| reports | n/a (0) | ok (11 refs) | quiet не проверяется в коде |
| roles | n/a (0) | partial (4 refs) | quiet не проверяется в коде |
| templates | n/a (0) | partial (4 refs) | quiet не проверяется в коде |
| tests | n/a (0) | ok (11 refs) | quiet не проверяется в коде |
| users | n/a (0) | partial (5 refs) | quiet не проверяется в коде |
| variables | n/a (0) | partial (7 refs) | quiet не проверяется в коде |
| list | n/a (0) | n/a (0) | нет interactive endpoint |
| add (root) | ok (11 refs) | ok (использует PrompterFromContext) | |
| update (root) | ok (8 refs) | ok (использует PrompterFromContext) | |
| delete (root) | n/a (0) | ok (использует PrompterFromContext) | |
| export (root) | partial (1 ref) | n/a | |

## Ключевые находки

### F1. LOCAL quiet override в cmd/run/run.go (HIGH)

- Файл: cmd/run/run.go:81
- Описание: команды run-семейства переопределяют `--quiet, -q` на уровне subCmd.Flags() вместо использования глобального PersistentFlag.
- Риск: переопределение shadou-ирует глобальный quiet, возможно несинхронизированное поведение.
- Remediation: удалить локальные BoolP-декларации quiet в cmd/run/run.go, cmd/test/list.go, cmd/test/get.go, cmd/result/result.go; оставить только global PersistentFlags.

Подтвержденные файлы:

- cmd/run/run.go:81
- cmd/test/list.go (найдено в поиске)
- cmd/test/get.go (найдено в поиске)
- cmd/result/result.go (найдено в поиске)

### F2. Команды с quiet=n/a (MEDIUM)

- 15 command groups не проверяют quiet напрямую, но это допустимо: service-output подавляется через `ui.SetMessageQuiet` глобально.
- Риск: команды, использующие stdout напрямую (вне ui.* helpers), не будут подавляться.
- Remediation direction: провести точечный аудит всех прямых fmt.Fprintf/os.Stdout в этих пакетах.

### F3. Локальная функция isQuiet в cmd/sync (LOW)

- Файл: cmd/sync/sync_helpers.go:15-21
- Описание: дублирует логику получения quiet-флага, хотя глобальный seam уже есть через ui.SetMessageQuiet.
- Риск: дублирование может привести к несогласованности, если PersistentPreRunE-логика изменится.
- Remediation direction: инкапсулировать получение quiet в один central path, убрать локальные wrapper-functions.

### F4. NonInteractivePrompter check паттерн (LOW – стандартизация)

- Найдены вызовы вида `_, ok := interactive.PrompterFromContext(ctx).(*interactive.NonInteractivePrompter)` в ~15 файлах.
- Это fragile type assertion вместо context-aware API.
- Remediation direction: добавить helper `interactive.IsNonInteractive(ctx)` для стандартизации.

## Матрица interactive contract

| Command Group | Имеет interactive_helpers.go | PrompterFromContext usage | NonInteractive check |
|---|---|---|---|
| compare | нет | да (all.go:602) | через ErrNonInteractive |
| get | нет | да (широко) | через type assertion |
| sync | нет | да (sync_*.go) | через ErrNonInteractive |
| run | да | да | через type assertion |
| result | нет | да | через type assertion |
| attachments/bdds/cases/configurations/datasets | да | да | type assertion + ErrNonInteractive mix |
| labels/milestones/plans/reports/roles | да | да | type assertion + ErrNonInteractive mix |
| templates/tests/test/users/variables | да | да | type assertion + ErrNonInteractive mix |

## Рекомендованные remediation

- R1 (HIGH): Убрать локальные quiet-flag декларации из cmd/run, cmd/test, cmd/result (4 файла).
- R2 (MEDIUM): Добавить `interactive.IsNonInteractive(ctx)` helper и мигрировать type assertion pattern.
- R3 (LOW): Убрать `isQuiet()` wrapper из cmd/sync/sync_helpers.go.
- R4 (MEDIUM): Провести аудит прямых fmt.Fprintf/os.Stdout вызовов в quiet=n/a command groups.
