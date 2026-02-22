# План рекурсивного распараллеливания (Recursive Parallelization)

## Общее описание

Цель: достичь времени выполнения compare cases < 5 минут за счёт максимального использования concurrency на всех уровнях.

## Текущая архитектура (последовательная)

```
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

```
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
) (map[int64][]Case, error)

func (c *ParallelController) FetchSuiteCases(
    projectID int64,
    suiteID int64,
) ([]Case, error)

func (c *ParallelController) FetchPage(
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

```
1. Получить список сьютов для обоих проектов
2. Для каждого сьюта:
   a. Сделать HEAD-запрос (или запросить page 0) для определения размера
   b. Рассчитать количество страниц
   c. Создать PageWork units для каждой страницы
3. Отсортировать PageWork по приоритету (большие сьюты первыми)
4. Создать dependency graph
```

### Фаза 2: Выполнение (Execution)

```
1. Инициализировать Controller с maxWorkers = 20
2. Для каждого проекта:
   a. Запустить горутину ProjectWorker
   b. ProjectWorker запускает SuiteWorkers параллельно
   c. Каждый SuiteWorker запускает PageWorkers через semaphore
3. Controller отслеживает прогресс через ResultAggregator
4. При ошибке: retry с backoff (max 3 attempts)
```

### Фаза 3: Сбор результатов (Aggregation)

```
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

### Этап 1: Core Infrastructure (2-3 дня)
- [ ] Создать `ParallelController`
- [ ] Создать `ResultAggregator`
- [ ] Создать `AdaptiveRateLimiter`
- [ ] Создать `WorkUnit` interfaces
- [ ] Unit tests

### Этап 2: Integration (2-3 дня)
- [ ] Интегрировать в `GetCasesParallel`
- [ ] Интегрировать в `compareCasesInternal`
- [ ] Обновить progress bars для множественных операций
- [ ] Integration tests

### Этап 3: Testing & Optimization (2-3 дня)
- [ ] Performance benchmarks
- [ ] Load testing с большими проектами
- [ ] Оптимизация rate limiting
- [ ] Обработка edge cases

### Этап 4: Documentation (1 день)
- [ ] Обновить docs/progress.md
- [ ] Добавить архитектурную диаграмму
- [ ] Описать алгоритм в комментариях

**Общая оценка**: 7-10 дней работы

## Ожидаемые результаты

| Метрика | Текущее | Целевое | Улучшение |
|---------|---------|---------|-----------|
| Время (36k cases) | ~12 мин | < 5 мин | 60% |
| Запросов в секунду | ~3 | ~10-15 | 300-500% |
| CPU usage | низкий | средний | приемлемо |
| Memory | низкий | средний | приемлемо |

## Риски

| Риск | Вероятность | Влияние | Митигация |
|------|-------------|---------|-----------|
| Rate limiting от API | Высокое | Высокое | Adaptive rate limiter |
| Race conditions | Среднее | Высокое | Тщательное тестирование |
| Memory leak | Низкое | Среднее | Профилирование |
| Сложность отладки | Среднее | Среднее | Подробное логирование |

## Решения для принятия

1. **Max concurrent requests**: 20 (текущий) или больше?
2. **Priority queue**: сортировать по размеру сьюта?
3. **Streaming results**: передавать результаты чанками?
4. **Error handling**: strict (fail fast) или lenient (partial results)?

---

**Статус**: На рассмотрении
**Приоритет**: High
**Зависит от**: Stage 6 completion (mpb migration)
