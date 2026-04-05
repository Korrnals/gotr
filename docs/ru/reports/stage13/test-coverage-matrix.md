# Stage 13 - Матрица тестового покрытия (к цели 100%)

Language: Русский | [English](../../../en/reports/stage13/test-coverage-matrix.md)

## Навигация

- [Документация](../../index.md)
  - [Гайды](../../guides/index.md)
    - [Установка](../../guides/installation.md)
    - [Конфигурация](../../guides/configuration.md)
    - [Интерактивный режим](../../guides/interactive-mode.md)
    - [Прогресс](../../guides/progress.md)
    - [Каталог команд](../../guides/commands/index.md)
      - [Общие](../../guides/commands/index.md#общие)
      - [CRUD операции](../../guides/commands/index.md#crud-операции)
      - [Основные ресурсы](../../guides/commands/index.md#основные-ресурсы)
      - [Специальные ресурсы](../../guides/commands/index.md#специальные-ресурсы)
  - [Архитектура](../../architecture/index.md)
  - [Эксплуатация](../../operations/index.md)
  - [Отчёты](../index.md)
    - [Stage 13](index.md)
    - [История](../history/index.md)
      - [Final Audit](final-coverage-audit-2026-04-05.md)
      - [Release Summary](release-summary.md)
      - [Audit Report](audit-report.md)
      - [Quality Metrics](quality-metrics.md)
      - [API Compliance](api-compliance-matrix.md)
      - [CLI Contract](cli-contract-matrix.md)
      - [Architecture Conformance](architecture-conformance.md)
      - [Reliability Audit](reliability-audit.md)
      - [Coverage Matrix](test-coverage-matrix.md)
      - [Checklist](coverage-checklist.md)
      - [Layer 2 Wave](layer2-coverage-wave.md)
      - [TODO](todo.md)
- [Главная](../../../../README_ru.md)

## Назначение документа

- Этот файл является единым рабочим источником истины по Coverage 100%.
- Перед каждым новым пакетом тестов сверяемся с чек-листом ниже.
- После каждого значимого шага обновляем:
- Текущий total.
- Закрытые пункты checklist.
- Последний baseline snapshot.

## Строгий протокол замера (обязательно)

- Команда замера total:
- `go test -vet=off -coverpkg=./... -coverprofile=/tmp/stage13_full.cover ./... >/tmp/stage13_full.log 2>&1`
- Команда извлечения total:
- `go tool cover -func=/tmp/stage13_full.cover | tail -n 5`
- Контрольные проверки:
- Проверять, что `/tmp/stage13_full.cover` пересоздан в текущем прогоне.
- Не брать total из старого/частично записанного profile.
- При флейке teardown в `cmd` повторить замер до стабильного `PASS` и только потом фиксировать total.

## Почему мог появиться дрейф 87% vs 85%

- В истории этапа были прогоны с нестабильным завершением `-coverpkg` (SIGQUIT/timeout в long teardown).
- Если профиль частично записан или взят от другого запуска, total может отличаться.
- Актуальным считается только значение, полученное по протоколу выше из последнего успешного полного прогона.

## Последний подтвержденный baseline

- Total statements coverage: **86.3%**.
- Команда: `go test -vet=off -coverpkg=./... -coverprofile=/tmp/stage13_full.cover ./...`.
- Артефакт: `/tmp/stage13_full.cover`.

## Повторный аудит покрытия (реальная картина)

- Дата: 2026-03-28.
- Профиль для глобального KPI:
- `go test -vet=off -coverpkg=./... -coverprofile=/tmp/stage13_audit.cover.out ./...`
- Итог global total (`-coverpkg`): **86.3%**.
- Профиль для файловой классификации (аудит покрыто/частично/не покрыто):
- `go test -vet=off -coverprofile=/tmp/stage13_pkg.cover.out ./...`
- Итог package-local total: **87.0%**.
- Файловая классификация (из package-local профиля):
- fully covered files: **74**
- partially covered files: **155**
- zero-covered files: **0**
- all tracked files: **229**
- Не покрыто (0%): нет.
- Lowest partial hotspots (приоритет на добор):
- `cmd/compare/register.go`
- `cmd/export.go`
- `cmd/root.go`
- `cmd/internal/testhelper/testhelper.go`
- `internal/client/mock.go`
- `internal/output/*`
- `internal/ui/*`
- Важно: для per-file buckets нельзя использовать агрегированные записи из `-coverpkg` профиля; они искажают файловую картину из-за повторной инструментализации в разных package test binaries.

## История контрольных baseline (коротко)

| Дата | Total | Комментарий |
| --- | --- | --- |
| 2026-03-27 | 67.4% | старт COV-workstream (COV-1) |
| 2026-03-28 | 77.4% | partial-18 snapshot |
| 2026-03-28 | 85.1% | актуальный контрольный baseline |
| 2026-03-28 | 86.9% | контрольный full-pass после расширения `internal/client/users_test.go` |
| 2026-03-28 | 87.0% | контрольный full-pass после блоков `internal/client/results` + `internal/client/runs` |
| 2026-03-28 | 87.1% | контрольный full-pass после расширения `internal/client/sections_test.go` |
| 2026-03-28 | 86.3% | повторная full-pass ресинхронизация после cmd utility wave (подтверждено повторным запуском) |

## Приоритетная матрица зон

| Приоритет | Зона | Статус | Комментарий |
| --- | --- | --- | --- |
| P0 | internal/client + internal/service | in progress | Основной резерв до 100% |
| P1 | internal/concurrency + internal/concurrent | in progress | Критично для reliability |
| P1 | cmd/resources + cmd utility helpers | in progress | Нужен полный branch-coverage добор |
| P2 | internal/log, internal/paths, internal/models/config | mostly done | Проверять остаточные edge-ветки |
| P3 | thin cmd wrappers | in progress | Массовый table-driven добор |

## COV чек-лист (рабочий)

- [x] COV-1: baseline + матрица.
- [x] COV-2: закрытие 0%-файлов инфраструктуры.
- [ ] COV-3: internal/client + internal/service >= 95%.
- [ ] COV-4: internal/concurrency + internal/concurrent >= 95%.
- [ ] COV-5: cmd thin wrappers + utility functions.
- [ ] COV-6: финальный targeted micro-pass до total 100.0%.

## Детальный TODO на текущую волну

- [ ] Закрыть `cmd` utility checklist.
- [ ] Довести `cmd/resources.go` функции до 100%.
- [x] Закрыть 0%-файлы: `embedded/jq_embed.go`, `pkg/testrailapi/api_paths.go`.
- [ ] Довести `internal/concurrency` до >=95% package coverage.
- [ ] Довести `internal/concurrent` до >=95% package coverage.
- [ ] Продолжить добор `internal/client` low-coverage функций (users/results/runs/sections/sharedsteps).
- [ ] Контрольный full-pass по всему проекту после каждого завершенного подпакета.
- [ ] Обновление этого файла после каждого шага (total + checklist).

## Последний выполненный шаг

- Добавлены error-branch тесты в `internal/client/users_test.go`:
- HTTP non-OK ветки для `GetUsers`, `GetUsersByProject`, `GetUser`, `GetUserByEmail`, `AddUser`, `UpdateUser`, `GetStatuses`.
- Decode-error ветки для `GetUsers`, `GetPriorities`, `GetTemplates`.
- Локальный итог по `internal/client/users.go` после шага: функции в диапазоне 71.4%-81.8%.
- Глобальный итог после контрольного full-pass: **86.9%**.

- Добавлены error-branch тесты в `internal/client/results_test.go`:
- Decode-error ветки для `AddResult`, `AddResultForCase`, `AddResults`, `AddResultsForCases`, `GetResultsForCase`.
- HTTP non-OK/request-error ветки для `AddResultsForCases`, `GetResults`, `GetResultsForRun`, `GetResultsForCase`.
- Локальный итог по `internal/client/results.go` после шага: 80.0%/80.0%/80.0%/80.0%/100.0%/100.0%/83.3% по функциям.
- Глобальный итог после контрольного full-pass: **86.9%** (без изменения total на этом микрошаге).

- Добавлены error/success тесты в `internal/client/runs_test.go`:
- Покрыты ветки `GetRuns` (success + request error), decode-error для `GetRun`/`AddRun`/`UpdateRun`/`CloseRun`.
- Локальный итог по `internal/client/runs.go` после шага: 83.3%/100.0%/80.0%/80.0%/83.3%/77.8% по функциям.
- Глобальный итог после контрольного full-pass: **87.0%**.

- Добавлены error-branch тесты в `internal/client/sharedsteps_test.go`:
- Покрыты request-error для `GetSharedSteps`, decode-error для `GetSharedStep`/`AddSharedStep`/`UpdateSharedStep`, non-OK для `GetSharedStepHistory`, а также `keep_in_cases=1` в `DeleteSharedStep`.
- Локальный итог по `internal/client/sharedsteps.go` после шага: 100.0%/83.3%/83.3%/80.0%/80.0%/80.0% по функциям.
- Глобальный итог после контрольного full-pass: **87.0%** (без изменения total на этом микрошаге).

- Добавлены utility-тесты в `cmd/result/add_test.go` и `cmd/root_test.go`:
- `buildAddResultRequest` доведен до 100.0% (all-fields + required-status validation).
- `cmd/root.go`: `GetClient` доведен до 100.0%, добавлены panic/success ветки для client accessors и ветка `initConfig` с invalid YAML.
- `cmd/resources.go` дополнительно проверен default-mode branch (с параметрами), но остаются хвосты `extractGetEndpointName` 91.7% и `getResourceEndpoints` 98.4%.
- Глобальный итог после контрольного full-pass: **87.0%** (без изменения total на этом микрошаге).

- Добавлены error-branch тесты в `internal/client/sections_test.go`:
- Decode-error для `GetSection`/`AddSection`/`UpdateSection`.
- Request-error ветка для `DeleteSection`.
- Частичный-успех + ошибка для `GetSectionsParallelCtx` (retain partial data + error).
- Локальный итог по `internal/client/sections.go`: 100.0%/87.2%/83.3%/80.0%/80.0%/77.8%.
- Глобальный итог после контрольного full-pass: **87.1%**.

- Выполнен refactor + тесты для `cmd/resources.go`:
- Удалена недостижимая JSON-error ветка в `getResourceEndpoints`.
- Добавлена валидация bare endpoint `get_` в `extractGetEndpointName` + тест-кейс.
- По package-профилю `./cmd` файл `cmd/resources.go` покрыт на **100%** по функциям.
- После повторного full-pass (coverpkg) подтвержден текущий total: **86.3%**.

- Выполнен повторный coverage-аудит с dual-profile протоколом:
- `-coverpkg` профиль зафиксирован как global KPI: **86.3%**.
- package-local профиль использован для реальной файловой классификации: total **85.8%**, buckets `70/157/2` (full/partial/zero).
- Зафиксированы 2 zero-covered файла и обновлен приоритет на их закрытие в текущем TODO.

- Закрыта zero-coverage волна (targeted tests):
- Добавлены тесты `embedded/jq_embed_test.go` (success + invalid filter для `RunEmbeddedJQ`).
- Добавлены тесты `pkg/testrailapi/api_paths_test.go` (инициализация API, агрегирование путей, ресурс `Groups`).
- Новый package-local snapshot: total **86.4%**, buckets `71/158/0` (full/partial/zero).
- Global KPI (`-coverpkg`) после шага подтвержден: **86.3%**.

- Выполнена Wave A/B для cmd checklist:
- Добавлены тесты `cmd/compare/register_test.go` (регистрация compare-команды, persistent flags и subcommands).
- Расширены тесты `cmd/export_test.go` (resource/endpoint/id ветки, interactive select/input ошибки).
- Расширены тесты `cmd/root_test.go` (fallback/panic path для `GetClientInterface`, `initConfig` not-found ветка).
- Расширены тесты `cmd/internal/testhelper/testhelper_test.go` (полное покрытие helper-функций и edge cases).
- Новый package-local snapshot: total **86.8%**, buckets `73/156/0` (full/partial/zero).

- Выполнен targeted step для `internal/output/dryrun.go`:
- Расширены тесты `internal/output/dryrun_test.go` (ветки `PrintOperation`, `PrintSimple`, marshal-error path).
- Файл `internal/output/dryrun.go` доведен до **100.0%** по функциям.
- Новый package-local snapshot: total **87.0%**, buckets `74/155/0` (full/partial/zero).

- Выполнена параллельная subagent-волна (output/ui + cmd tail + extended):
- Расширены `internal/output/save_test.go`, `internal/ui/display_test.go`, `internal/client/extended_test.go`.
- Дополнены `cmd/export_test.go`, `cmd/root_test.go`; в `cmd/root.go` добавлен test hook для home-dir edge path.
- Новый package-local snapshot: total **88.2%**, buckets `74/155/0` (full/partial/zero).
- Global KPI (`-coverpkg`) после полного прогона: **86.3%**.
- Ключевые улучшения файлов:
- `internal/output/save.go`: **56.12% -> 83.16%**.
- `internal/ui/display.go`: **56.74% -> 98.58%**.
- `internal/client/extended.go`: **68.62% -> 75.86%**.

## Coverage Gate (финальный)

- `go test ./...` == PASS.
- `CGO_ENABLED=1 go test -race ./...` == PASS.
- `go test -vet=off -coverpkg=./... -coverprofile=/tmp/stage13_full.cover ./...` == PASS.
- `go tool cover -func=/tmp/stage13_full.cover` == `total: 100.0%`.

## Рабочие правила

- Не сканировать проект заново «с нуля» перед каждым микрошагом.
- Работать по этому checklist сверху вниз.
- В конце каждого шага:
- Обновить этот файл.
- Обновить quality metrics.
- Сделать контрольный прогон.

---

← [Stage 13](index.md) · [Отчёты](../index.md) · [Документация](../../index.md)
