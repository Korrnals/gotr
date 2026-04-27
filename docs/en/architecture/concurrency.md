# Concurrent Package — Documentation

Language: [Русский](../../ru/architecture/concurrency.md) | English

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

## Contents

- [Overview](#overview)
- [How it speeds things up](#how-it-speeds-things-up)
- [Architecture](#architecture)
- [Components](#components)
  - [WorkerPool](#workerpool)
  - [RateLimiter](#ratelimiter)
  - [Retry](#retry)
  - [CircuitBreaker](#circuitbreaker)
- [Usage examples](#usage-examples)
- [Best Practices](#best-practices)

---

## Overview

When working with the TestRail API the main performance issue is **sequential requests**. For example:

```text
Comparing cases between projects:
- Get the suite list of project 1 (1 request)
- For each suite, fetch its cases (N requests sequentially)
- Get the suite list of project 2 (1 request)
- For each suite, fetch its cases (N requests sequentially)
```

**With 10 suites in each project:** 20+ seconds (at 1 second per request)

**With parallel execution:** ~4–5 seconds (5x speed-up)

---

## How it speeds things up

### Sequential processing (WITHOUT concurrent)

```text
Time:  0    1s   2s   3s   4s   5s
       |----|----|----|----|----|
Suite1 [####]                          ← 1 sec
Suite2      [####]                     ← 1 sec
Suite3           [####]                ← 1 sec
Suite4                [####]           ← 1 sec
Suite5                     [####]      ← 1 sec

Total: 5 seconds
```

### Parallel processing (WITH concurrent)

```text
Time:  0    1s   2s
       |----|----|
Worker1 [####]                         ← Suite1
Worker2 [####]                         ← Suite2
Worker3 [####]                         ← Suite3
Worker4 [####]                         ← Suite4
Worker5 [####]                         ← Suite5

Total: ~1.2 seconds (accounting for rate limit)
```

### Rate Limiting (protection against bans)

The TestRail API has a limit: **180 requests/minute**

```text
Without a rate limiter:
  - 100 parallel requests → HTTP 429 (Too Many Requests)
  - 1-minute ban

With a rate limiter (150 req/min):
  - Requests are spread out evenly
  - 150 requests = 60 seconds (one request every 0.4s)
  - No ban, stable operation
```

---

## Architecture

```text
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

### Execution flow

```text
1. A task arrives in the WorkerPool
   ↓
2. WorkerPool picks a worker from the pool (maxWorkers)
   ↓
3. RateLimiter checks the limit
   ↓
4. If the limit is not exceeded → execute the request
   ↓
5. On error → Retry with exponential backoff
   ↓
6. CircuitBreaker tracks errors
   ↓
7. On many errors → the circuit opens (blocking)
```

---

## Components

### WorkerPool

Manages a pool of goroutines to execute tasks.

```go
pool := concurrent.NewWorkerPool(
    concurrent.WithMaxWorkers(5),        // Max 5 parallel tasks
    concurrent.WithRateLimit(150),       // 150 requests per minute
)

// Submit tasks
for _, suiteID := range suiteIDs {
    id := suiteID // Capture the variable
    pool.Submit(func() error {
        cases, err := client.GetCases(projectID, id, 0)
        // Process the result...
        return err
    })
}

// Wait for all tasks to finish
if err := pool.Wait(); err != nil {
    log.Printf("Some tasks failed: %v", err)
}
```

### How it works

- Creates an `errgroup` with a goroutine count limit
- Each task runs in its own goroutine
- `SetLimit(maxWorkers)` controls the parallelism

### ParallelMap

Applies a function to every element of a slice in parallel.

```go
suiteIDs := []int64{1, 2, 3, 4, 5}

results, err := concurrent.ParallelMap(suiteIDs, 5,
    func(suiteID int64, index int) ([]Case, error) {
        return client.GetCases(projectID, suiteID, 0)
    })

// Process the results
for _, result := range results {
    if result.Error != nil {
        log.Printf("Failed to get cases for suite %d: %v",
            result.Index, result.Error)
        continue
    }
    allCases = append(allCases, result.Data...)
}
```

### Benefits

- Automatic distribution across workers
- Result order is preserved (by index)
- Per-element error handling

### RateLimiter

Token bucket rate limiter.

```go
limiter := concurrent.NewRateLimiter(150) // 150 req/min

// Option 1: blocking wait
limiter.Wait()  // Waits until a token is available

// Option 2: non-blocking check
if limiter.Allow() {
    // We can run the request
} else {
    // Limit exceeded, skip or wait
}

// Option 3: wait with timeout
if limiter.WaitWithTimeout(5 * time.Second) {
    // Token acquired
}
```

### Algorithm

- The bucket has capacity = burst size
- Every second new tokens are added (rate/60)
- If the bucket is empty — we wait

### AdaptiveRateLimiter

Adapts the rate limit based on API response time.

```go
limiter := concurrent.NewAdaptiveRateLimiter(150)

// Run a request
start := time.Now()
response, err := client.GetCases(projectID, suiteID, 0)
duration := time.Since(start)

// Tell the limiter the response time
limiter.RecordResponseTime(duration)

// The limiter automatically:
// - Decreases the rate on slow responses (>2s)
// - Increases the rate on fast responses (<500ms)
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

### Delays

- Attempt 1: immediately
- Attempt 2: after 1 sec
- Attempt 3: after 2 sec
- Attempt 4: after 4 sec
- Attempt 5: after 8 sec
- Attempt 6: after 16 sec

### CircuitBreaker

Protection against cascade failures.

```go
cb := concurrent.NewCircuitBreaker(
    5,                    // Max 5 failures
    30 * time.Second,     // Timeout before half-open
)

// After 5 failures in a row:
// - The circuit opens
// - All requests are blocked
// - After 30 seconds it switches to half-open
// - 1 successful request closes the circuit

err := cb.Execute(func() error {
    return client.GetCases(projectID, suiteID, 0)
})

if err != nil && err.Error() == "circuit breaker is open" {
    // Circuit is open, need to wait
    time.Sleep(30 * time.Second)
}
```

### States

- **Closed**: all requests pass through
- **Open**: all requests are blocked
- **Half-Open**: one probe request

---

## Usage examples

### Example 1: parallel fetch of cases from all suites

```go
func FetchAllCasesParallel(
    client client.ClientInterface,
    projectID int64,
) ([]Case, error) {

    // 1. Get the list of suites (1 request)
    suites, err := client.GetSuites(projectID)
    if err != nil {
        return nil, err
    }

    // 2. Fetch cases from each suite in parallel
    results, err := concurrent.ParallelMap(suites, 5,
        func(suite Suite, index int) ([]Case, error) {
            return client.GetCases(projectID, suite.ID, 0)
        })

    if err != nil {
        return nil, err
    }

    // 3. Collect the results
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

// Result:
// Before: 10 suites × 1 sec = 10 seconds
// After: 10 suites / 5 workers × 1 sec = ~2 seconds
```

### Example 2: Compare All with parallel fetch

```go
func CompareAllParallel(
    client client.ClientInterface,
    pid1, pid2 int64,
) (*AllResult, error) {

    // Resource types to compare
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
        // ... other resources
    }

    // Compare all resources in parallel
    results, err := concurrent.ParallelMap(resources, 5,
        func(res struct {
            name string
            fn   func() (*CompareResult, error)
        }, index int) (*CompareResult, error) {
            return res.fn()
        })

    // Collect the results...
}

// Result:
// Before: cases(10s) → suites(2s) → sections(3s) → ... = 20+ sec
// After: max(10s, 2s, 3s, ...) = ~10 seconds
```

### Example 3: batch processing with retry

```go
func ImportCasesWithRetry(
    client client.ClientInterface,
    cases []Case,
) error {

    processor := concurrent.NewBatchProcessor[Case](
        concurrent.WithBatchSize[Case](50),        // 50 cases per batch
        concurrent.WithRetryPolicy[Case](3, time.Second), // 3 attempts
    )

    return processor.Process(cases, func(batch []Case) error {
        return client.AddCases(batch)
    })
}
```

---

## Best Practices

### 1. Worker pool size

```go
// Not too many (don't overload the API)
// Not too few (don't lose performance)

// Sweet spot: 3–5 for TestRail
pool := concurrent.NewWorkerPool(
    concurrent.WithMaxWorkers(5),
)
```

### 2. Rate limit

```go
// TestRail limit: 180 req/min
// Take with a margin: 150 req/min (83% of the limit)
limiter := concurrent.NewRateLimiter(150)
```

### 3. Error handling

```go
results, _ := concurrent.ParallelMap(items, 5, fn)

// ALWAYS check result.Error
for _, result := range results {
    if result.Error != nil {
        // Log it, but keep going
        log.Printf("Error: %v", result.Error)
        continue
    }
    // Process result.Data
}
```

### 4. Timeouts

```go
// Set timeouts at the HTTP client level
// + timeouts at the retry level

config := &concurrent.RetryConfig{
    MaxRetries:   3,
    InitialDelay: 1 * time.Second,
    // No more than 7 seconds across all attempts
}
```

### 5. Graceful Degradation

```go
// If the CircuitBreaker is open — keep going with restrictions
err := cb.Execute(func() error {
    return fetchData()
})

if err != nil && err.Error() == "circuit breaker is open" {
    // Fallback: sequential processing
    return fetchDataSequential()
}
```

---

## Performance metrics

### Test results

| Scenario | Sequential | Parallel | Speed-up |
|----------|------------|----------|----------|
| 10 suites × 100 cases | 12 sec | 3 sec | **4x** |
| 20 suites × 50 cases | 22 sec | 5 sec | **4.4x** |
| Compare All (6 resources) | 25 sec | 12 sec | **2x** |
| Import 500 cases | 45 sec | 15 sec | **3x** |

Measured on a connection with ~200ms RTT to TestRail Cloud.

---

## Conclusion

The `internal/concurrent` package allows you to:

1. **Speed up** processing by 2–5×
2. **Control** the load on the API
3. **Handle** errors gracefully
4. **Protect** against cascade failures

The key rule: **parallelism + rate limiting = stable performance**.

---

← [Architecture](index.md) · [Documentation](../index.md)
