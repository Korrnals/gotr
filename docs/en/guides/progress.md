# Progress system in gotr

Language: [Русский](../../ru/guides/progress.md) | English

## Navigation

- [Documentation](../index.md)
  - [Guides](index.md)
    - [Installation](installation.md)
    - [Configuration](configuration.md)
    - [Interactive Mode](interactive-mode.md)
    - [Progress](progress.md)
    - [Commands Index](commands/index.md)
    - [Instructions](instructions/index.md)
  - [Architecture](../architecture/index.md)
  - [Operations](../operations/index.md)
  - [Reports](../reports/index.md)
- [Home](../../../README.md)

## TL;DR

The project uses a **single progress runtime** from `internal/ui`.

- Main API: `ui.RunWithStatus(...)`, `ui.NewOperation(...)`, `TaskHandle`.
- The runtime works with `context.Context` (Ctrl+C / timeout cancel operations cleanly).
- Parallel fetches use the `concurrency.ProgressReporter` / `concurrency.PaginatedProgressReporter` contract.
- Commands must not implement low-level progress bar rendering directly.

The legacy `internal/progress` package and the old mpb-based scenarios are no longer the current runtime contract.

## Architectural role

Data flow:

1. `cmd/*` creates an operation/task via `internal/ui`.
2. The client / concurrency layer receives a reporter (`TaskHandle`) via config.
3. The concurrency layer emits progress events (`OnItemComplete`, `OnBatchReceived`, `OnPageFetched`, `OnError`).
4. The UI runtime renders progress and final statuses.

## Core entities

### 1) `ui.RunWithStatus`

High-level helper for phased/status operations.

```go
_, err := ui.RunWithStatus(ctx, ui.StatusConfig{
    Title: "Loading data",
}, func(ctx context.Context) (struct{}, error) {
    // long-running work
    return struct{}{}, nil
})
```

### 2) `ui.Operation`

Use it when several tasks belong to a single operation.

```go
op := ui.NewOperation(ui.StatusConfig{Title: "Compare cases"})
defer op.Finish()

task := op.AddTask("Project 30", 12)
// task implements ProgressReporter / PaginatedProgressReporter
```

### 3) `TaskHandle` as reporter

`TaskHandle` is passed into `concurrency.ControllerConfig.Reporter` or `FetchParallel*` options.

```go
cfg := &concurrency.ControllerConfig{
    MaxConcurrentSuites: 10,
    MaxConcurrentPages:  6,
    Timeout:             30 * time.Minute,
    Reporter:            task,
}
```

## How this is used in compare

### `compare cases`

- Uses the heavy runtime (`GetCasesParallelCtx`) + `TaskHandle` reporter.
- Supports timeout / cancel / retry / failed pages flow.

### `compare sections` (Stage 11)

- Migrated to the adapter path (`GetSectionsParallelCtx`) with the same runtime config as heavy compare.
- The command layer must no longer implement its own suite/page loop.

## Heavy compare runtime configuration

A single profile for heavy commands (`cases`, `sections`):

- `--parallel-suites`
- `--parallel-pages`
- `--page-retries`
- `--rate-limit`
- `--timeout`

Value sources:

1. Command flags (highest priority)
2. Viper config (`compare.*`, `compare.cases.*`)
3. Defaults

## Rules for new code

1. For long-running operations use only the `internal/ui` runtime.
2. Do not write custom progress renderers in `cmd/*` and do not duplicate concurrency loops.
3. Progress must be built through concurrency reporter contracts.
4. Always pass `ctx` down to the client/concurrency layer.
5. Cancellation errors (`context.Canceled`, `context.DeadlineExceeded`) are a regular flow, not an exception.

## Verification after changes

Minimal checklist:

1. `go test ./cmd/compare/...`
2. `go test ./internal/client/...`
3. `go test ./...`
4. Smoke:
   - `gotr compare cases --pid1 <id> --pid2 <id>`
   - `gotr compare sections --pid1 <id> --pid2 <id>`
   - Ctrl+C during a fetch must not print noise and must terminate the operation cleanly.

## Related documents

- `docs/architecture/overview.md`
- `docs/guides/interactive-mode.md` (Stage 12 roadmap, including Stage 12.3 on test coverage)
- `.github/instructions/STAGE_11.0_DESIGN.md`
- `.github/instructions/PLAN.md`

## Current focus (Stage 12)

1. Stages 12.0–12.2 are done.
2. Active stage: **12.3 — full test audit and coverage closeout**.
3. Next: Stage 12.4 (cleanup wrappers) and Stage 12.5 (docs + release readiness).

---

← [Guides](index.md) · [Documentation](../index.md)
