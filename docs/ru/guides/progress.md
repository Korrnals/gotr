# Система прогресса в gotr

Language: Русский | [English](../../en/guides/progress.md)

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

## Кратко

В проекте используется **единый runtime прогресса** из `internal/ui`.

- Основной API: `ui.RunWithStatus(...)`, `ui.NewOperation(...)`, `TaskHandle`.
- Runtime работает с `context.Context` (Ctrl+C/timeout отменяют операции корректно).
- Для параллельных загрузок используется контракт `concurrency.ProgressReporter` / `concurrency.PaginatedProgressReporter`.
- Команды не должны напрямую реализовывать рендер прогресса на уровне низкоуровневых баров.

Legacy-пакет `internal/progress` и старые mpb-сценарии больше не являются текущим runtime-контрактом.

## Архитектурная роль

Поток данных:

1. `cmd/*` создаёт operation/task через `internal/ui`.
2. Клиент/конкурентный слой получает reporter (`TaskHandle`) через конфиг.
3. Concurrency-слой отправляет события прогресса (`OnItemComplete`, `OnBatchReceived`, `OnPageFetched`, `OnError`).
4. UI-runtime показывает прогресс и итоговые статусы.

## Основные сущности

### 1) `ui.RunWithStatus`

Высокоуровневый helper для операций с фазами/статусом.

```go
_, err := ui.RunWithStatus(ctx, ui.StatusConfig{
    Title: "Loading data",
}, func(ctx context.Context) (struct{}, error) {
    // long-running work
    return struct{}{}, nil
})
```

### 2) `ui.Operation`

Подходит, когда нужно несколько задач в рамках одной операции.

```go
op := ui.NewOperation(ui.StatusConfig{Title: "Compare cases"})
defer op.Finish()

task := op.AddTask("Project 30", 12)
// task реализует ProgressReporter/PaginatedProgressReporter
```

### 3) `TaskHandle` как reporter

`TaskHandle` передаётся в `concurrency.ControllerConfig.Reporter` или `FetchParallel*`-опции.

```go
cfg := &concurrency.ControllerConfig{
    MaxConcurrentSuites: 10,
    MaxConcurrentPages:  6,
    Timeout:             30 * time.Minute,
    Reporter:            task,
}
```

## Как это используется в compare

### `compare cases`

- Использует heavy runtime (`GetCasesParallelCtx`) + `TaskHandle` reporter.
- Поддерживает timeout/cancel/retry/failed pages flow.

### `compare sections` (Stage 11)

- Переведён на adapter-path (`GetSectionsParallelCtx`) с тем же runtime-конфигом heavy compare.
- Командный слой больше не должен реализовывать собственные suite/page loop.

## Конфигурация heavy compare runtime

Единый профиль для heavy-команд (`cases`, `sections`):

- `--parallel-suites`
- `--parallel-pages`
- `--page-retries`
- `--rate-limit`
- `--timeout`

Источники значений:

1. Флаги команды (приоритет выше)
2. Viper-конфиг (`compare.*`, `compare.cases.*`)
3. Default значения

## Правила для нового кода

1. Для долгих операций использовать только `internal/ui` runtime.
2. В `cmd/*` не писать кастомные рендереры прогресса и не дублировать concurrency loops.
3. Прогресс должен строиться через reporter-контракты concurrency.
4. Всегда прокидывать `ctx` до client/concurrency слоя.
5. Ошибки отмены (`context.Canceled`, `context.DeadlineExceeded`) обрабатывать как штатный сценарий.

## Проверка после изменений

Минимальный чек:

1. `go test ./cmd/compare/...`
2. `go test ./internal/client/...`
3. `go test ./...`
4. Smoke:
   - `gotr compare cases --pid1 <id> --pid2 <id>`
   - `gotr compare sections --pid1 <id> --pid2 <id>`
   - Ctrl+C во время загрузки не печатает шум и завершает операцию корректно

## Связанные документы

- `docs/architecture/overview.md`
- `docs/guides/interactive-mode.md` (дорожная карта Stage 12, включая Stage 12.3 по тестовому покрытию)
- `.github/instructions/STAGE_11.0_DESIGN.md`
- `.github/instructions/PLAN.md`

## Текущий фокус (Stage 12)

1. Stage 12.0-12.2 выполнены.
2. Активная стадия: **12.3 — полный тест-аудит и добивка покрытия**.
3. Следом: Stage 12.4 (cleanup wrappers) и Stage 12.5 (docs + release readiness).

---

← [Гайды](index.md) · [Документация](../index.md)
