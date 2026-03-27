# Stage 13 - Reliability & Concurrency Audit

Дата: 2026-03-27
Ветка: stage-13.0-final-refactoring
Шаг: Phase 2 → Step 5

## A. go test -race result

**Статус**: EXECUTED (частичный пакетный прогон по критичным зонам).

- Environment: gcc установлен (`/usr/bin/gcc`, GCC 15.2.1), `CGO_ENABLED=1`.
- Проверенные пакеты:
- `./internal/concurrency` — PASS
- `./internal/concurrent` — PASS
- `./internal/client` (targeted concurrent/paginator tests) — PASS
- `./internal/service/...` — PASS
- `./internal/interactive/...` — PASS
- `./cmd/compare/...` — сначала FAIL с `WARNING: DATA RACE`, после фикса — PASS.
- DATA RACE локация: `cmd/compare/fetchers_test.go` (конкурентный append в shared slice `captured` в mock closure).
- Fix: добавить `sync.Mutex` для защиты append (commit `9358ac8`).
- Полный `go test -race ./...` остается обязательным финальным gate (Phase 4) и CI gate (R7).

---

## B. Static Concurrency Analysis (manual)

### B1 — Global Variables Inventory

| Package | Var | Type | Race-safe? | Notes |
| --- | --- | --- | --- | --- |
| internal/ui/display.go | messageQuiet | atomic.Bool | PASS | Store/Load через atomic. |
| internal/log/logger.go | globalLogger | *zap.Logger | PASS | Инициализация через sync.Once. |
| internal/concurrency/types.go | PriorityThresholds | struct (mutable) | WARN | Публичная мутабельная структура — теоретически могут быть concurrent writes если используется из нескольких goroutine — READ-only в practice (только читается в GetPriority). |
| cmd/compare/register.go | getClient, Cmd | func/cobra.Command | INFO | Init-time write в init(), runtime read-only. |

### B2 — Mutex/Atomic Coverage в concurrent пакетах

| File | Protected data | Primitive | Pattern | Status |
| --- | --- | --- | --- | --- |
| internal/concurrency/aggregator.go | cases slice | casesMu (sync.RWMutex) | Lock/Unlock per write, RLock per read | PASS |
| internal/concurrency/aggregator.go | seenIDs map | seenMu (sync.RWMutex) | Lock per write | PASS |
| internal/concurrency/aggregator.go | totalCases/Pages | atomic.Int64 | atomic ops | PASS |
| internal/concurrency/fetch_parallel.go | results map[int64][]T | mu (sync.Mutex) | Lock/Unlock per write | PASS |
| internal/concurrency/fetch_parallel.go | collectedErrors slice | errMu (sync.Mutex) | Lock/Unlock per append | PASS |
| internal/concurrency/fetch_by_suite.go | same pattern | mu, errMu | same | PASS |
| internal/client/concurrent.go | results map | mu (sync.Mutex) | Lock/Unlock per write | PASS |
| internal/service/migration/import.go | shared slices | mu (sync.Mutex) | Lock/Unlock | PASS |
| internal/concurrent/pool.go | shared state | mu (sync.Mutex) | Lock/Unlock | PASS |
| internal/concurrent/retry.go | retry state | mu (sync.Mutex) | Field on struct | PASS |

### B3 — Goroutine Loop Variable Capture

| File | Pattern | Status |
| --- | --- | --- |
| internal/concurrency/fetch_by_suite.go:57 | `sid := sid` перед `g.Go` | PASS |
| internal/concurrency/fetch_parallel.go:107 | `pid := pid` перед `g.Go` | PASS |
| internal/concurrent/pool.go:119,150 | `i, item := i, item` | PASS |
| internal/client/cases.go:418,423 | goroutines с channel communication | PASS |

Note: Go 1.22+ автоматически захватывает loop vars, но явные captures сохраняют обратную совместимость.

### B4 — errgroup Usage

- Используется `golang.org/x/sync/errgroup` во всех параллельных пакетах.
- `g.SetLimit(n)` применяется для управления параллелизмом.
- `ctx` из `errgroup.WithContext` корректно передаётся в worker-функции.
- Context cancellation propagation: проверяется через `ctx.Err() != nil` и `isCancellationError`.

### B5 — Channel Safety

- `internal/client/cases.go`: buffered channel `make(chan result, 2)` для 2 goroutines — корректно.
- Priority queue: `internal/concurrency/priority_queue.go` — heap под `mu sync.RWMutex` — PASS.

---

## C. Findings

| ID | Severity | Location | Description |
| --- | --- | --- | --- |
| F10 | INFO | Reliability gate | Полный `go test -race ./...` по всему репозиторию еще не закрыт в одном прогоне; выполнены targeted race-прогоны по критичным пакетам. |
| F11 | WARN | internal/concurrency/types.go:26 | PriorityThresholds — публичная мутабельная глобальная struct. READ-ONLY в runtime, но ничего технически не запрещает внешний write. Низкий реальный риск. |
| F12 | HIGH (test-only) | cmd/compare/fetchers_test.go:268 | DATA RACE в тесте `TestCompareSectionsInternal_UsesHeavyRuntimeConfig`: конкурентный append в `captured` из 2 goroutines. |

---

## D. Remediation

| ID | Action |
| --- | --- |
| R7 (INFO) | Добавить `go test -race ./...` в Makefile и CI pipeline как обязательный gate. |
| R8 (LOW) | Сделать PriorityThresholds константой или убрать из export (только для internal use). |
| R9 (DONE) | Зафиксирован race-fix в `cmd/compare/fetchers_test.go` (mutex around captured append), commit `9358ac8`. |

---

## E. Status

- Static analysis: PASS — все критические concurrent паттерны корректны.
- Race detector: PARTIAL PASS — critical пакеты проверены, data race в compare-тесте найден и исправлен.
- Mutex/atomic coverage: COMPLETE — все shared mutable state защищены.
- Loop variable capture: PASS — явные captures присутствуют.
