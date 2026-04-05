# План рекурсивного распараллеливания (Recursive Parallelization)

Language: Русский | [English](../../en/architecture/recursive-parallelization-plan.md)

## Навигация

- [Документация](../index.md)
  - [Гайды](../guides/index.md)
    - [Установка](../guides/installation.md)
    - [Конфигурация](../guides/configuration.md)
    - [Интерактивный режим](../guides/interactive-mode.md)
    - [Прогресс](../guides/progress.md)
    - [Каталог команд](../guides/commands/index.md)
      - [Общие](../guides/commands/index.md#общие)
      - [CRUD операции](../guides/commands/index.md#crud-операции)
      - [Основные ресурсы](../guides/commands/index.md#основные-ресурсы)
      - [Специальные ресурсы](../guides/commands/index.md#специальные-ресурсы)
  - [Архитектура](index.md)
    - [Обзор](overview.md)
    - [Concurrency](concurrency.md)
    - [Стандарты](standards.md)
    - [План распараллеливания](recursive-parallelization-plan.md)
  - [Эксплуатация](../operations/index.md)
  - [Отчёты](../reports/index.md)
- [Главная](../../../README_ru.md)

## Общее описание

Цель: достичь времени выполнения compare cases < 5 минут за счёт максимального использования concurrency на всех уровнях.

## Текущая архитектура (последовательная)

```text
Проект 30 (10 сьютов)
  ├── Suite 1: page 0 → page 250 → page 500 → ... (последовательно)
  ├── Suite 2: page 0 → page 250 → page 500 → ... (последовательно)
  └── ...

Проект 34 (21 сьют)
  ├── Suite 1: page 0 → page 250 → page 500 → ... (последовательно)
  └── ...
```

**Время**: ~12 минут

## Целевая архитектура (рекурсивно-параллельная)

```text
Проект 30 (10 сьютов) ──┐
                         ├── Controller ──┐
Проект 34 (21 сьют) ────┘                │
                                          │
    ┌─────────────────────────────────────┼─────────────────────┐
    │                                     │                     │
Suite 1: pages [0, 250, 500...]    Suite 2: pages [...]    ...
    │                                     │
    ├─ Page 0 (250 cases) ──┐            ├─ Page 0 (250 cases) ──┐
    ├─ Page 250 (250 cases)─┼── Worker    ├─ Page 250 (250 cases)─┼── Worker
    ├─ Page 500 (250 cases)─┘  Pool      ├─ Page 500 (250 cases)─┘
    └─ ...                         (max 20 concurrent pages)
```

**Целевое время**: < 5 минут

## Компоненты системы

### 1. Controller (Оркестратор)

**Ответственность**:

- Управление жизненным циклом всех горутин
- Отслеживание выполнения всех запросов
- Сбор результатов в правильном порядке
- Обработка ошибок и retry

**API**:

```go
type ParallelController struct {
    rateLimiter    *rate.Limiter
    semaphore      chan struct{}     // Ограничение concurrency
    errGroup       *errgroup.Group   // Отслеживание ошибок
    results        sync.Map          // Thread-safe хранилище результатов
    progress       *ProgressTracker  // Общий прогресс
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

### 2. Work Unit (Единица работы)

**Типы work units**:

- `ProjectWork` - загрузка всего проекта
- `SuiteWork` - загрузка одного сьюта
- `PageWork` - загрузка одной страницы (250 кейсов)

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

### 3. Result Aggregator (Сборщик результатов)

**Ответственность**:

- Сбор результатов из разных горутин
- Упорядочивание по suiteID
- Дедупликация кейсов
- Потоковая передача в compare

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

### 4. Adaptive Rate Limiter (Адаптивный rate limiter)

**Функционал**:

- Базовый rate: 180 req/min (3 req/sec)
- Burst: 20 requests
- Адаптация под нагрузку (уменьшение при 429 ошибках)
- Priority queue (сьюты с большим количеством кейсов первыми)

```go
type AdaptiveRateLimiter struct {
    baseRate    rate.Limit
    burst       int
    currentRate atomic.Value // rate.Limit
    priorityQ   PriorityQueue
}
```

## Алгоритм работы

### Фаза 1: Планирование (Planning)

```text
1. Получить список сьютов для обоих проектов
2. Для каждого сьюта:
   a. Сделать HEAD-запрос (или запросить page 0) для определения размера
   b. Рассчитать количество страниц
   c. Создать PageWork units для каждой страницы
3. Отсортировать PageWork по приоритету (большие сьюты первыми)
4. Создать dependency graph
```

### Фаза 2: Выполнение (Execution)

```text
1. Инициализировать Controller с maxWorkers = 20
2. Для каждого проекта:
   a. Запустить горутину ProjectWorker
   b. ProjectWorker запускает SuiteWorkers параллельно
   c. Каждый SuiteWorker запускает PageWorkers через semaphore
3. Controller отслеживает прогресс через ResultAggregator
4. При ошибке: retry с backoff (max 3 attempts)
```

### Фаза 3: Сбор результатов (Aggregation)

```text
1. По мере завершения PageWork:
   a. Добавить кейсы в ResultAggregator
   b. Обновить прогресс
2. Когда все PageWork для сьюта завершены:
   a. Отметить SuiteResult как Completed
   b. Уведомить ProjectWorker
3. Когда все сьюты проекта готовы:
   a. Вернуть результат в compareCasesInternal
```

## Ограничения и гарантии

### Rate Limiting

- **Hard limit**: 180 req/min
- **Max concurrent**: 20 requests
- **Burst**: 20 requests
- **Adaptive**: автоматическое снижение при 429 ошибках

### Надёжность

- **Retry logic**: 3 попытки с exponential backoff
- **Circuit breaker**: отключение при последовательных ошибках
- **Timeout**: 30 сек на запрос
- **Graceful degradation**: при ошибках возвращаем partial results

### Прогресс

- **Global progress**: общий прогресс по всем запросам
- **Per-project progress**: отдельный прогресс для каждого проекта
- **Per-suite progress**: опционально (для больших сьютов)

## Тестирование

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

## Этапы внедрения

### Этап 1: Core Infrastructure ✅

- [x] Создать `ParallelController` — `internal/parallel/controller.go`
- [x] Создать `ResultAggregator` — `internal/parallel/aggregator.go`
- [x] Создать `PriorityQueue` — `internal/parallel/priority_queue.go`
- [x] Создать `SuiteFetcher` interface — `internal/parallel/types.go`
- [x] Unit tests — `controller_test.go`, `aggregator_test.go`, `priority_queue_test.go`

### Этап 2: Integration ✅

- [x] Интегрировать в `GetCasesParallel` — `internal/client/cases.go`
- [x] Интегрировать в `compareCasesInternal` — `cmd/compare/cases.go`
- [x] Pipeline pagination — страницы запрашиваются конвейерно
- [x] Auto-retry failed pages — автоматический повтор упавших страниц
- [x] Config profiles — `fast`, `balanced`, `safe` профили тюнинга
- [x] ANSI live display — динамическая таблица со статистикой в реальном времени

### Этап 3: Testing & Optimization ✅

- [x] Оптимизация rate limiting — configurable, unlocked burst mode
- [x] Обработка edge cases — truncated responses, partial results
- [x] Data completeness verification — верификация полноты данных
- [x] Configurable retries — `--max-retries`, `--retry-delay`

### Этап 4: Unified Output & Documentation ✅

- [x] Централизованный reporter — `internal/ui/reporter/reporter.go`
- [x] Унификация вывода всех 11 compare-команд (reporter вместо progress.Manager)
- [x] `*Ctx` naming convention — функции с `context.Context` имеют суффикс `Ctx`
- [x] Документация — `docs/guides/configuration.md`, `docs/guides/progress.md`

## Фактические результаты

| Метрика | До | После | Улучшение |
|---------|---------|---------|-----------|
| Время (36k cases) | ~12 мин | ~4 мин | **70%** |
| Запросов в секунду | ~3 | ~15-20 | **500-700%** |
| Retry логика | нет | auto-retry failed pages | **∞** |
| Вывод | progress.Manager | centralized reporter | **унифицировано** |

## Архитектура (реализованная, актуализировано в Stage 6.8)

```text
internal/concurrency/          # было: internal/parallel/
├── types.go              # SuiteFetcher, PageRequest, PageResult, PipelineConfig
├── priority_queue.go     # PriorityQueue (heap-based)
├── aggregator.go         # ResultAggregator — сбор результатов из горутин
├── controller.go         # ParallelController — оркестрация pipeline
├── simple.go             # FetchParallel[T], FetchParallelBySuite[T] (Stage 6.8)
├── doc.go                # Документация пакета
└── *_test.go             # Тесты

internal/ui/
└── display.go            # ANSI live display — динамическая таблица

pkg/reporter/                  # было: internal/ui/reporter/ (Stage 6.8)
└── reporter.go           # Builder pattern: Section/Stat/StatIf/StatFmt/Print

internal/ui/
├── runtime.go            # Unified progress runtime (RunWithStatus, Operation, TaskHandle)
└── display.go            # ANSI live display — динамическая таблица
```

## Принятые решения

1. **Max concurrent requests**: 20 (по умолчанию, настраивается через `--workers`)
2. **Priority queue**: да, сортировка по размеру сьюта (большие первыми)
3. **Streaming results**: pipeline — страницы обрабатываются по мере получения
4. **Error handling**: lenient — собираем partial results, retry failed pages

---

**Статус**: ✅ Реализовано
**Ветка**: `stage-6.7-recursive-parallelization` (11 коммитов)  
**Дата завершения**: 2026-03-03  
**См. также**: Stage 6.8 (`STAGE_6.8_DESIGN.md`) — унификация concurrency и перенос `internal/parallel/` → `internal/concurrency/`

---

← [Архитектура](index.md) · [Документация](../index.md)
