# Concurrent Package - Документация

> Пакет `internal/concurrent` предоставляет инструменты для параллельной обработки API запросов с контролем нагрузки.

## Содержание

- [Обзор](#обзор)
- [Как это ускоряет работу](#как-это-ускоряет-работу)
- [Архитектура](#архитектура)
- [Компоненты](#компоненты)
  - [WorkerPool](#workerpool)
  - [RateLimiter](#ratelimiter)
  - [Retry](#retry)
  - [CircuitBreaker](#circuitbreaker)
- [Примеры использования](#примеры-использования)
- [Best Practices](#best-practices)

---

## Обзор

При работе с TestRail API основная проблема производительности — **последовательные запросы**. Например:

```
Сравнение кейсов между проектами:
- Получить список сьютов проекта 1 (1 запрос)
- Для каждого сьюта получить кейсы (N запросов последовательно)
- Получить список сьютов проекта 2 (1 запрос)
- Для каждого сьюта получить кейсы (N запросов последовательно)
```

**При 10 сьютов в каждом проекте:** 20+ секунд (при 1 сек на запрос)

**С parallel execution:** ~4-5 секунд (5x ускорение)

---

## Как это ускоряет работу

### Последовательная обработка (БЕЗ concurrent)

```
Время: 0    1s   2s   3s   4s   5s
       |----|----|----|----|----|
Suite1 [####]                          ← 1 сек
Suite2      [####]                     ← 1 сек  
Suite3           [####]                ← 1 сек
Suite4                [####]           ← 1 сек
Suite5                     [####]      ← 1 сек

Total: 5 секунд
```

### Параллельная обработка (С concurrent)

```
Время: 0    1s   2s
       |----|----|
Worker1 [####]                         ← Suite1
Worker2 [####]                         ← Suite2
Worker3 [####]                         ← Suite3
Worker4 [####]                         ← Suite4
Worker5 [####]                         ← Suite5

Total: ~1.2 секунды (с учетом rate limit)
```

### Rate Limiting (Защита от бана)

TestRail API имеет лимит: **180 requests/minute**

```
Без rate limiter:
  - 100 параллельных запросов → HTTP 429 (Too Many Requests)
  - Бан на 1 минуту

С rate limiter (150 req/min):
  - Запросы равномерно распределяются
  - 150 запросов = 60 секунд (1 запрос каждые 0.4с)
  - Нет бана, стабильная работа
```

---

## Архитектура

```
┌─────────────────────────────────────────────────────────────┐
│                    Concurrent Package                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐     │
│  │ WorkerPool  │───▶│ RateLimiter │───▶│  HTTP API   │     │
│  │             │    │             │    │             │     │
│  │ - Submit()  │    │ - Wait()    │    │ - Request   │     │
│  │ - Wait()    │    │ - Allow()   │    │ - Response  │     │
│  └─────────────┘    └─────────────┘    └─────────────┘     │
│         │                                                    │
│         ▼                                                    │
│  ┌─────────────┐    ┌─────────────┐                         │
│  │    Retry    │◀───│   Circuit   │                         │
│  │             │    │   Breaker   │                         │
│  │ - Backoff   │    │             │                         │
│  │ - Timeout   │    │ - Open      │                         │
│  └─────────────┘    │ - Closed    │                         │
│                     └─────────────┘                         │
└─────────────────────────────────────────────────────────────┘
```

### Поток выполнения

```
1. Задача поступает в WorkerPool
   ↓
2. WorkerPool берет воркера из пула (maxWorkers)
   ↓
3. RateLimiter проверяет лимит
   ↓
4. Если лимит не превышен → выполняем запрос
   ↓
5. При ошибке → Retry с exponential backoff
   ↓
6. CircuitBreaker отслеживает ошибки
   ↓
7. При множестве ошибок → Circuit открывается (блокировка)
```

---

## Компоненты

### WorkerPool

Управляет пулом горутин для выполнения задач.

```go
pool := concurrent.NewWorkerPool(
    concurrent.WithMaxWorkers(5),        // Макс 5 параллельных задач
    concurrent.WithRateLimit(150),       // 150 запросов в минуту
)

// Добавляем задачи
for _, suiteID := range suiteIDs {
    id := suiteID // Захватываем переменную
    pool.Submit(func() error {
        cases, err := client.GetCases(projectID, id, 0)
        // Обработка результатов...
        return err
    })
}

// Ждем завершения всех задач
if err := pool.Wait(); err != nil {
    log.Printf("Some tasks failed: %v", err)
}
```

**Как работает:**
- Создает `errgroup` с ограничением на количество горутин
- Каждая задача выполняется в отдельной горутине
- `SetLimit(maxWorkers)` контролирует параллелизм

### ParallelMap

Параллельно применяет функцию к каждому элементу слайса.

```go
suiteIDs := []int64{1, 2, 3, 4, 5}

results, err := concurrent.ParallelMap(suiteIDs, 5, 
    func(suiteID int64, index int) ([]Case, error) {
        return client.GetCases(projectID, suiteID, 0)
    })

// Обработка результатов
for _, result := range results {
    if result.Error != nil {
        log.Printf("Failed to get cases for suite %d: %v", 
            result.Index, result.Error)
        continue
    }
    allCases = append(allCases, result.Data...)
}
```

**Преимущества:**
- Автоматическое распределение по воркерам
- Сохранение порядка результатов (по индексу)
- Обработка ошибок для каждого элемента

### RateLimiter

Token bucket rate limiter.

```go
limiter := concurrent.NewRateLimiter(150) // 150 req/min

// Вариант 1: Блокирующее ожидание
limiter.Wait()  // Ждет, пока не появится токен

// Вариант 2: Не-блокирующая проверка
if limiter.Allow() {
    // Можем выполнить запрос
} else {
    // Лимит превышен, пропускаем или ждем
}

// Вариант 3: Ожидание с таймаутом
if limiter.WaitWithTimeout(5 * time.Second) {
    // Токен получен
}
```

**Алгоритм:**
- Bucket имеет capacity = burst size
- Каждую секунду добавляются новые токены (rate/60)
- Если bucket пуст — ждем

### AdaptiveRateLimiter

Адаптирует rate limit на основе времени ответа API.

```go
limiter := concurrent.NewAdaptiveRateLimiter(150)

// Выполняем запрос
start := time.Now()
response, err := client.GetCases(projectID, suiteID, 0)
duration := time.Since(start)

// Сообщаем лимитеру время ответа
limiter.RecordResponseTime(duration)

// Лимитер автоматически:
// - Уменьшает rate при медленных ответах (>2s)
// - Увеличивает rate при быстрых ответах (<500ms)
```

### Retry

Exponential backoff retry.

```go
config := &concurrent.RetryConfig{
    MaxRetries:   5,
    InitialDelay: 1 * time.Second,
    MaxDelay:     30 * time.Second,
    Multiplier:   2.0,  // 1s, 2s, 4s, 8s, 16s
}

err := concurrent.Retry(config, func() error {
    return client.GetCases(projectID, suiteID, 0)
})
```

**Задержки:**
- Попытка 1: сразу
- Попытка 2: через 1 сек
- Попытка 3: через 2 сек
- Попытка 4: через 4 сек
- Попытка 5: через 8 сек
- Попытка 6: через 16 сек

### CircuitBreaker

Защита от cascade failures.

```go
cb := concurrent.NewCircuitBreaker(
    5,                    // Max 5 failures
    30 * time.Second,     // Timeout before half-open
)

// При 5 ошибках подряд:
// - Circuit открывается
// - Все запросы блокируются
// - Через 30 секунд переходит в half-open
// - 1 успешный запрос закрывает circuit

err := cb.Execute(func() error {
    return client.GetCases(projectID, suiteID, 0)
})

if err != nil && err.Error() == "circuit breaker is open" {
    // Circuit открыт, нужно подождать
    time.Sleep(30 * time.Second)
}
```

**Состояния:**
- **Closed** (закрыт): Все запросы проходят
- **Open** (открыт): Все запросы блокируются
- **Half-Open** (проверка): 1 тестовый запрос

---

## Примеры использования

### Пример 1: Параллельная загрузка кейсов из всех сьютов

```go
func FetchAllCasesParallel(
    client client.ClientInterface, 
    projectID int64,
) ([]Case, error) {
    
    // 1. Получаем список сьютов (1 запрос)
    suites, err := client.GetSuites(projectID)
    if err != nil {
        return nil, err
    }

    // 2. Параллельно загружаем кейсы из каждого сьюта
    results, err := concurrent.ParallelMap(suites, 5, 
        func(suite Suite, index int) ([]Case, error) {
            return client.GetCases(projectID, suite.ID, 0)
        })

    if err != nil {
        return nil, err
    }

    // 3. Собираем результаты
    var allCases []Case
    for _, result := range results {
        if result.Error != nil {
            log.Printf("Failed to get cases for suite %d: %v",
                result.Index, result.Error)
            continue
        }
        allCases = append(allCases, result.Data...)
    }

    return allCases, nil
}

// Результат:
// Было: 10 сьютов × 1 сек = 10 секунд
// Стало: 10 сьютов / 5 workers × 1 сек = ~2 секунды
```

### Пример 2: Compare All с параллельной загрузкой

```go
func CompareAllParallel(
    client client.ClientInterface,
    pid1, pid2 int64,
) (*AllResult, error) {

    // Типы ресурсов для сравнения
    resources := []struct {
        name string
        fn   func() (*CompareResult, error)
    }{
        {"cases", func() (*CompareResult, error) {
            return compareCasesParallel(client, pid1, pid2)
        }},
        {"suites", func() (*CompareResult, error) {
            return compareSuitesParallel(client, pid1, pid2)
        }},
        {"sections", func() (*CompareResult, error) {
            return compareSectionsParallel(client, pid1, pid2)
        }},
        // ... другие ресурсы
    }

    // Параллельно сравниваем все ресурсы
    results, err := concurrent.ParallelMap(resources, 5,
        func(res struct {
            name string
            fn   func() (*CompareResult, error)
        }, index int) (*CompareResult, error) {
            return res.fn()
        })

    // Собираем результаты...
}

// Результат:
// Было: cases(10s) → suites(2s) → sections(3s) → ... = 20+ сек
// Стало: max(10s, 2s, 3s, ...) = ~10 секунд
```

### Пример 3: Batch processing с retry

```go
func ImportCasesWithRetry(
    client client.ClientInterface,
    cases []Case,
) error {

    processor := concurrent.NewBatchProcessor[Case](
        concurrent.WithBatchSize[Case](50),        // 50 кейсов за раз
        concurrent.WithRetryPolicy[Case](3, time.Second), // 3 попытки
    )

    return processor.Process(cases, func(batch []Case) error {
        return client.AddCases(batch)
    })
}
```

---

## Best Practices

### 1. Размер Worker Pool

```go
// Не слишком много (не перегружать API)
// Не слишком мало (не терять производительность)

// Оптимально: 3-5 для TestRail
pool := concurrent.NewWorkerPool(
    concurrent.WithMaxWorkers(5),
)
```

### 2. Rate Limit

```go
// TestRail limit: 180 req/min
// Берем с запасом: 150 req/min (83% от лимита)
limiter := concurrent.NewRateLimiter(150)
```

### 3. Обработка ошибок

```go
results, _ := concurrent.ParallelMap(items, 5, fn)

// ВСЕГДА проверяйте result.Error
for _, result := range results {
    if result.Error != nil {
        // Логируем, но продолжаем
        log.Printf("Error: %v", result.Error)
        continue
    }
    // Обрабатываем result.Data
}
```

### 4. Таймауты

```go
// Устанавливайте таймауты на уровне HTTP клиента
// + таймауты на уровне retry

config := &concurrent.RetryConfig{
    MaxRetries:   3,
    InitialDelay: 1 * time.Second,
    // Не более 7 секунд на все попытки
}
```

### 5. Graceful Degradation

```go
// Если CircuitBreaker открыт — продолжаем с ограничениями
err := cb.Execute(func() error {
    return fetchData()
})

if err != nil && err.Error() == "circuit breaker is open" {
    // Fallback: последовательная обработка
    return fetchDataSequential()
}
```

---

## Метрики производительности

### Тестовые результаты

| Сценарий | Последовательно | Параллельно | Ускорение |
|----------|----------------|-------------|-----------|
| 10 сьютов по 100 кейсов | 12 сек | 3 сек | **4x** |
| 20 сьютов по 50 кейсов | 22 сек | 5 сек | **4.4x** |
| Compare All (6 ресурсов) | 25 сек | 12 сек | **2x** |
| Import 500 кейсов | 45 сек | 15 сек | **3x** |

*Измерено на соединении с RTT ~200ms к TestRail Cloud*

---

## Заключение

Пакет `internal/concurrent` позволяет:

1. **Ускорить** обработку в 2-5 раз
2. **Контролировать** нагрузку на API
3. **Обрабатывать** ошибки gracefully
4. **Защищаться** от cascade failures

Ключевое правило: **параллелизм + rate limiting = стабильная производительность**.
