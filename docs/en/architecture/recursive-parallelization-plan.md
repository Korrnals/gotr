# Recursive parallelization plan

Language: [Русский](../../ru/architecture/recursive-parallelization-plan.md) | English

## Navigation

- [Documentation](../index.md)
  - [Guides](../guides/index.md)
    - [Installation](../guides/installation.md)
    - [Configuration](../guides/configuration.md)
    - [Interactive Mode](../guides/interactive-mode.md)
    - [Progress](../guides/progress.md)
    - [Commands Index](../guides/commands/index.md)
      - [General](../guides/commands/index.md#general)
      - [CRUD Operations](../guides/commands/index.md#crud-operations)
      - [Core Resources](../guides/commands/index.md#core-resources)
      - [Special Resources](../guides/commands/index.md#special-resources)
    - [Instructions](../guides/instructions/index.md)
  - [Architecture](index.md)
    - [Overview](overview.md)
    - [Concurrency](concurrency.md)
    - [Standards](standards.md)
    - [Parallelization Plan](recursive-parallelization-plan.md)
  - [Operations](../operations/index.md)
  - [Reports](../reports/index.md)
- [Home](../../../README.md)

## Summary

Goal: bring `compare cases` runtime under 5 minutes by maximising the use of concurrency at every level.

## Current architecture (sequential)

```text
Project 30 (10 suites)
  ├── Suite 1: page 0 → page 250 → page 500 → ... (sequential)
  ├── Suite 2: page 0 → page 250 → page 500 → ... (sequential)
  └── ...

Project 34 (21 suites)
  ├── Suite 1: page 0 → page 250 → page 500 → ... (sequential)
  └── ...
```

**Time**: ~12 minutes

## Target architecture (recursively parallel)

```text
Project 30 (10 suites) ──┐
                          ├── Controller ──┐
Project 34 (21 suites) ──┘                │
                                           │
    ┌──────────────────────────────────────┼─────────────────────┐
    │                                      │                     │
Suite 1: pages [0, 250, 500...]    Suite 2: pages [...]    ...
    │                                      │
    ├─ Page 0 (250 cases) ──┐             ├─ Page 0 (250 cases) ──┐
    ├─ Page 250 (250 cases)─┼── Worker     ├─ Page 250 (250 cases)─┼── Worker
    ├─ Page 500 (250 cases)─┘  Pool       ├─ Page 500 (250 cases)─┘
    └─ ...                         (max 20 concurrent pages)
```

**Target time**: < 5 minutes

## System components

### 1. Controller (orchestrator)

**Responsibilities**:

- Manage the lifecycle of all goroutines
- Track the execution of every request
- Collect results in the right order
- Handle errors and retries

**API**:

```go
type ParallelController struct {
    rateLimiter    *rate.Limiter
    semaphore      chan struct{}     // Concurrency limit
    errGroup       *errgroup.Group   // Error tracking
    results        sync.Map          // Thread-safe result store
    progress       *ProgressTracker  // Shared progress
}

func (c *ParallelController) FetchProjectCases(
    projectID int64,
    suites []Suite,
) (`map[int64][]Case`, error)

func (c *ParallelController) FetchSuiteCases(
    projectID int64,
    suiteID int64,
) ([]Case, error)

func (c *ParallelController) FetchPageCtx(
    projectID int64,
    suiteID int64,
    offset int64,
) ([]Case, error)
```

### 2. Work Unit (unit of work)

**Work unit types**:

- `ProjectWork` — load an entire project
- `SuiteWork` — load a single suite
- `PageWork` — load a single page (250 cases)

```go
type WorkUnit interface {
    Execute(ctx context.Context) error
    Priority() int
    Dependencies() []WorkUnit
}

type PageWork struct {
    ProjectID int64
    SuiteID   int64
    Offset    int64
    Limit     int64
    Result    []Case
}
```

### 3. Result Aggregator (result collector)

**Responsibilities**:

- Collect results from different goroutines
- Order them by suiteID
- Deduplicate cases
- Stream into compare

```go
type ResultAggregator struct {
    mu       sync.RWMutex
    results  map[int64]SuiteResult
    complete chan struct{}
}

type SuiteResult struct {
    SuiteID   int64
    Cases     []Case
    Completed bool
    Error     error
}
```

### 4. Adaptive Rate Limiter

**Functionality**:

- Base rate: 180 req/min (3 req/sec)
- Burst: 20 requests
- Adapts to load (decreases on 429 errors)
- Priority queue (suites with more cases first)

```go
type AdaptiveRateLimiter struct {
    baseRate    rate.Limit
    burst       int
    currentRate atomic.Value // rate.Limit
    priorityQ   PriorityQueue
}
```

## Algorithm

### Phase 1: Planning

```text
1. Get the suite list for both projects
2. For each suite:
   a. Issue a HEAD request (or fetch page 0) to determine the size
   b. Compute the number of pages
   c. Create PageWork units for every page
3. Sort PageWork by priority (large suites first)
4. Build the dependency graph
```

### Phase 2: Execution

```text
1. Initialise the Controller with maxWorkers = 20
2. For each project:
   a. Spawn a ProjectWorker goroutine
   b. ProjectWorker spawns SuiteWorkers in parallel
   c. Each SuiteWorker spawns PageWorkers gated by a semaphore
3. The Controller tracks progress via the ResultAggregator
4. On error: retry with backoff (max 3 attempts)
```

### Phase 3: Aggregation

```text
1. As each PageWork finishes:
   a. Add its cases to the ResultAggregator
   b. Update progress
2. When every PageWork for a suite is complete:
   a. Mark SuiteResult as Completed
   b. Notify ProjectWorker
3. When all suites of a project are ready:
   a. Return the result to compareCasesInternal
```

## Limits and guarantees

### Rate Limiting

- **Hard limit**: 180 req/min
- **Max concurrent**: 20 requests
- **Burst**: 20 requests
- **Adaptive**: automatic back-off on 429 errors

### Reliability

- **Retry logic**: 3 attempts with exponential backoff
- **Circuit breaker**: trip on consecutive errors
- **Timeout**: 30 seconds per request
- **Graceful degradation**: return partial results on errors

### Progress

- **Global progress**: aggregated progress across all requests
- **Per-project progress**: separate progress per project
- **Per-suite progress**: optional (for large suites)

## Testing

### Unit Tests

```go
func TestParallelController_FetchProjectCases(t *testing.T)
func TestResultAggregator_Ordering(t *testing.T)
func TestAdaptiveRateLimiter_Backoff(t *testing.T)
func TestWorkUnit_Priority(t *testing.T)
```

### Integration Tests

```go
func TestParallelCompareCases_SmallProjects(t *testing.T)     // < 1000 cases
func TestParallelCompareCases_MediumProjects(t *testing.T)    // ~ 10k cases
func TestParallelCompareCases_LargeProjects(t *testing.T)     // ~ 50k cases
func TestParallelCompareCases_WithErrors(t *testing.T)        // retry logic
func TestParallelCompareCases_RateLimiting(t *testing.T)      // 429 handling
```

### Performance Benchmarks

```go
func BenchmarkSequentialCompare(b *testing.B)
func BenchmarkParallelCompare(b *testing.B)
```

## Implementation stages

### Stage 1: Core Infrastructure ✅

- [x] Build `ParallelController` — `internal/parallel/controller.go`
- [x] Build `ResultAggregator` — `internal/parallel/aggregator.go`
- [x] Build `PriorityQueue` — `internal/parallel/priority_queue.go`
- [x] Build `SuiteFetcher` interface — `internal/parallel/types.go`
- [x] Unit tests — `controller_test.go`, `aggregator_test.go`, `priority_queue_test.go`

### Stage 2: Integration ✅

- [x] Wire it into `GetCasesParallel` — `internal/client/cases.go`
- [x] Wire it into `compareCasesInternal` — `cmd/compare/cases.go`
- [x] Pipeline pagination — pages are requested in a pipeline
- [x] Auto-retry failed pages — failed pages are retried automatically
- [x] Config profiles — `fast`, `balanced`, `safe` tuning profiles
- [x] ANSI live display — dynamic table with real-time statistics

### Stage 3: Testing & Optimization ✅

- [x] Rate-limiting optimisation — configurable, unlocked burst mode
- [x] Edge-case handling — truncated responses, partial results
- [x] Data completeness verification
- [x] Configurable retries — `--max-retries`, `--retry-delay`

### Stage 4: Unified Output & Documentation ✅

- [x] Centralised reporter — `internal/ui/reporter/reporter.go`
- [x] Unified output across all 11 compare commands (reporter instead of progress.Manager)
- [x] `*Ctx` naming convention — functions taking `context.Context` are suffixed with `Ctx`
- [x] Documentation — `docs/guides/configuration.md`, `docs/guides/progress.md`

## Actual results

| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| Time (36k cases) | ~12 min | ~4 min | **70%** |
| Requests per second | ~3 | ~15–20 | **500–700%** |
| Retry logic | none | auto-retry failed pages | **∞** |
| Output | progress.Manager | centralised reporter | **unified** |

## Architecture (as implemented, refreshed in Stage 6.8)

```text
internal/concurrency/          # was: internal/parallel/
├── types.go              # SuiteFetcher, PageRequest, PageResult, PipelineConfig
├── priority_queue.go     # PriorityQueue (heap-based)
├── aggregator.go         # ResultAggregator — collects results from goroutines
├── controller.go         # ParallelController — pipeline orchestration
├── simple.go             # FetchParallel[T], FetchParallelBySuite[T] (Stage 6.8)
├── doc.go                # Package documentation
└── *_test.go             # Tests

internal/ui/
└── display.go            # ANSI live display — dynamic table

pkg/reporter/                  # was: internal/ui/reporter/ (Stage 6.8)
└── reporter.go           # Builder pattern: Section/Stat/StatIf/StatFmt/Print

internal/ui/
├── runtime.go            # Unified progress runtime (RunWithStatus, Operation, TaskHandle)
└── display.go            # ANSI live display — dynamic table
```

## Decisions made

1. **Max concurrent requests**: 20 (default, configurable via `--workers`)
2. **Priority queue**: yes, sorted by suite size (large suites first)
3. **Streaming results**: pipeline — pages are processed as they arrive
4. **Error handling**: lenient — keep partial results, retry failed pages

---

**Status**: ✅ Implemented
**Branch**: `stage-6.7-recursive-parallelization` (11 commits)
**Completion date**: 2026-03-03
**See also**: Stage 6.8 (`STAGE_6.8_DESIGN.md`) — concurrency unification and the move of `internal/parallel/` → `internal/concurrency/`

---

← [Architecture](index.md) · [Documentation](../index.md)
