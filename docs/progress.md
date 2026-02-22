# Progress Monitoring System

## Overview

The `progress` package provides a universal, decoupled progress monitoring system for long-running operations in `gotr`. It uses channel-based communication to separate business logic from UI updates, allowing any method to report progress without knowing about progress bars or other UI components.

The package now uses **`github.com/vbauerster/mpb/v8`** (multi-progress-bar) which supports rendering multiple progress bars on separate lines simultaneously.

## Key Features

- **Universal**: Works with any method that supports progress reporting
- **Non-blocking**: Channel-based updates don't block execution
- **Thread-safe**: Safe to use from multiple goroutines
- **Decoupled**: Business logic knows nothing about UI
- **Flexible**: Can be used with progress bars, logs, or any other UI
- **Multi-bar Support**: Multiple progress bars render simultaneously on separate lines

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Business Logic                           │
│  ┌─────────────────────┐    ┌─────────────────────┐             │
│  │   GetCasesParallel  │    │  GetSuitesParallel  │             │
│  │  (accepts Monitor)  │    │  (accepts Monitor)  │             │
│  └──────────┬──────────┘    └──────────┬──────────┘             │
│             │                          │                        │
│             ▼                          ▼                        │
│  ┌──────────────────────────────────────────┐                   │
│  │         WorkerPool with Monitor           │                   │
│  │  (calls monitor.Increment() after task)   │                   │
│  └──────────┬───────────────────────────────┘                   │
└─────────────┼──────────────────────────────────────────────────┘
              │ sends to channel
              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     UI / Progress Layer                          │
│  ┌─────────────────────┐    ┌─────────────────────┐             │
│  │    progress.Bar     │    │      Logger         │             │
│  │  (receives updates) │    │  (receives updates) │             │
│  └─────────────────────┘    └─────────────────────┘             │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              mpb.Progress Container                      │   │
│  │         (manages multiple bars rendering)               │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## Components

### 1. Manager

The `Manager` creates and manages an `mpb.Progress` container which handles all progress bars:

```go
type Manager struct {
    progress *mpb.Progress
    output   io.Writer
    quiet    bool
    enabled  bool
    wg       sync.WaitGroup
}
```

**Key Methods:**
- `NewManager(opts ...Option)` - Creates a new manager with mpb container
- `NewBar(total int64, description string) *Bar` - Creates a new progress bar
- `NewSpinner(description string) *Bar` - Creates a spinner for indeterminate operations
- `Wait()` - Waits for all bars to complete

### 2. Bar

The `Bar` type represents a single progress bar with methods called directly on the object:

```go
type Bar struct {
    bar    *mpb.Bar
    total  int64
    mgr    *Manager
}
```

**Key Methods (called on bar object):**
- `bar.Add(n)` - Increment by n
- `bar.Increment()` - Increment by 1
- `bar.Finish()` - Complete the bar
- `bar.Describe(description)` - Update description

### 3. Monitor

The `Monitor` type tracks progress and sends updates through a channel:

```go
type Monitor struct {
    ProgressChan chan<- int  // Channel for updates
    Total        int         // Expected total
    completed    int64       // Atomic counter
}
```

**Key Methods:**
- `Increment()` - Increment by 1 (thread-safe, non-blocking)
- `IncrementBy(n)` - Increment by n
- `Completed()` - Get completed count
- `Percentage()` - Get completion percentage

### 4. WorkerPool Integration

The `concurrent.WorkerPool` automatically calls `monitor.Increment()` after each task completes:

```go
pool := concurrent.NewWorkerPool(
    concurrent.WithMaxWorkers(5),
    concurrent.WithRateLimit(150),
    concurrent.WithProgressMonitor(monitor),  // Optional
)
```

### 5. Client Methods

All parallel client methods accept an optional `ProgressMonitor`:

```go
func (c *HTTPClient) GetCasesParallel(
    projectID int64, 
    suiteIDs []int64, 
    workers int, 
    monitor ProgressMonitor,  // Can be nil
) (map[int64]data.GetCasesResponse, error)
```

## Usage Examples

### Basic Example with Progress Bar

```go
package main

import (
    "github.com/Korrnals/gotr/internal/client"
    "github.com/Korrnals/gotr/internal/progress"
)

func main() {
    cli := client.NewHTTPClient(...)
    
    // Create manager with mpb container
    pm := progress.NewManager()
    
    // Create progress bar (method on manager)
    bar := pm.NewBar(int64(len(suiteIDs)), "Loading cases...")
    
    // Create monitor with channel
    progressChan := make(chan int, 100)
    monitor := progress.NewMonitor(progressChan, len(suiteIDs))
    
    // Goroutine to update progress bar (method on bar object)
    go func() {
        for range progressChan {
            bar.Increment()  // Called on bar, not package
        }
    }()
    
    // Execute with progress tracking
    cases, err := cli.GetCasesParallel(30, suiteIDs, 5, monitor)
    if err != nil {
        log.Fatal(err)
    }
    
    // Finish the bar (method on bar object)
    bar.Finish()
    
    // Wait for all bars to complete
    pm.Wait()
    
    // Use results...
    fmt.Printf("Loaded %d cases\n", len(cases))
}
```

### Two-Phase Progress with Multiple Bars

With mpb, multiple bars render simultaneously on separate lines:

```go
func compareCasesWithProgress(cli client.ClientInterface, pid1, pid2 int64) {
    pm := progress.NewManager()
    
    // Phase 1: Spinner for getting suites
    spinner := pm.NewSpinner("Getting suites list...")
    suitesMap, err := cli.GetSuitesParallel([]int64{pid1, pid2}, 2, nil)
    spinner.Finish()  // Method on bar object
    
    // Phase 2: Two progress bars render simultaneously (mpb feature)
    totalSuites := len(suitesMap[pid1]) + len(suitesMap[pid2])
    bar := pm.NewBar(int64(totalSuites), "Loading cases...")
    
    progressChan := make(chan int, totalSuites)
    monitor := progress.NewMonitor(progressChan, totalSuites)
    
    go func() {
        for range progressChan {
            bar.Increment()  // Method on bar object
        }
    }()
    
    // Both projects load in parallel with shared progress
    var wg sync.WaitGroup
    wg.Add(2)
    
    go func() {
        defer wg.Done()
        cli.GetCasesParallel(pid1, suiteIDs1, 5, monitor)
    }()
    
    go func() {
        defer wg.Done()
        cli.GetCasesParallel(pid2, suiteIDs2, 5, monitor)
    }()
    
    wg.Wait()
    bar.Finish()  // Method on bar object
    
    // Wait for all bars to render
    pm.Wait()
}
```

### Multiple Simultaneous Progress Bars

mpb supports multiple bars rendering at the same time:

```go
func loadMultipleProjects(cli client.ClientInterface, projectIDs []int64) {
    pm := progress.NewManager()
    
    // Each bar renders on its own line
    bars := make(map[int64]*progress.Bar)
    for _, pid := range projectIDs {
        bars[pid] = pm.NewBar(100, fmt.Sprintf("Project %d", pid))
    }
    
    var wg sync.WaitGroup
    for _, pid := range projectIDs {
        wg.Add(1)
        go func(id int64) {
            defer wg.Done()
            // Work...
            for i := 0; i < 100; i++ {
                bars[id].Increment()  // Each bar updated independently
                time.Sleep(10 * time.Millisecond)
            }
            bars[id].Finish()
        }(pid)
    }
    
    wg.Wait()
    pm.Wait()
}
```

### Without Progress (Backwards Compatible)

```go
// Simply pass nil as monitor
cases, err := cli.GetCasesParallel(30, suiteIDs, 5, nil)
```

## Interface Design

### Why Interface?

The `ProgressMonitor` interface allows:
1. **Testing** - Easy to mock
2. **Flexibility** - Different implementations (silent, logging, etc.)
3. **Decoupling** - Client doesn't depend on progress package

```go
// In client/interfaces.go
type ProgressMonitor interface {
    Increment()
}

// In progress/monitor.go - full implementation
type Monitor struct {
    ProgressChan chan<- int
    Total        int
    completed    int64
}

func (m *Monitor) Increment() { ... }
func (m *Monitor) IncrementBy(n int) { ... }
```

## Best Practices

1. **Always use buffered channels** to avoid blocking:
   ```go
   progressChan := make(chan int, 100)  // Buffer = number of expected updates
   ```

2. **Close the channel** when done:
   ```go
   defer close(progressChan)
   ```

3. **Handle nil gracefully** - methods should work with `monitor = nil`:
   ```go
   if monitor != nil {
       monitor.Increment()
   }
   ```

4. **Update UI in separate goroutine** to avoid blocking workers:
   ```go
   go func() {
       for range progressChan {
           bar.Increment()  // Method on bar object
       }
   }()
   ```

5. **Use appropriate buffer size** based on expected number of tasks

6. **Call pm.Wait()** at the end to ensure all bars finish rendering:
   ```go
   pm := progress.NewManager()
   // ... create and use bars ...
   pm.Wait()  // Wait for mpb to finish rendering
   ```

## Migration from progressbar/v3 to mpb

### Package Import Change

**Old:**
```go
import "github.com/schollz/progressbar/v3"
```

**New:**
```go
import "github.com/vbauerster/mpb/v8"
```

### API Changes

**Old (progressbar/v3):**
```go
bar := progressbar.NewOptions64(total, ...)
progress.Add(bar, 1)        // Package function
progress.Finish(bar)        // Package function
```

**New (mpb):**
```go
pm := progress.NewManager()
bar := pm.NewBar(total, "Description")  // Method on manager
bar.Add(1)                              // Method on bar
bar.Increment()                         // Method on bar
bar.Finish()                            // Method on bar
pm.Wait()                               // Wait for container
```

### Key Differences

| Feature | progressbar/v3 | mpb |
|---------|---------------|-----|
| Library | `github.com/schollz/progressbar/v3` | `github.com/vbauerster/mpb/v8` |
| Multiple bars | Overwrite each other | Render on separate lines |
| Update API | Package functions (`progress.Add()`) | Methods on bar (`bar.Add()`) |
| Container | None | `mpb.Progress` managed by Manager |
| Wait | Not needed | `pm.Wait()` required |

## Comparison with Alternatives

| Approach | Pros | Cons |
|----------|------|------|
| **Callback** (old) | Simple | Blocks worker, hard to manage |
| **Channel + Monitor** (current) | Non-blocking, flexible, clean | Slightly more code |
| **Context with values** | Standard pattern | Verbose, harder to type-safe |
| **Global state** | Easy | Not thread-safe, hard to test |

## Migration Guide

### From Callback to Monitor

**Old:**
```go
onComplete := func(suiteID int64) {
    bar.Add(1)
}
cases, err := cli.GetCasesParallelWithCallback(pid, suites, 5, onComplete)
```

**New:**
```go
progressChan := make(chan int, len(suites))
monitor := progress.NewMonitor(progressChan, len(suites))

go func() {
    for range progressChan {
        bar.Increment()  // Method on bar object
    }
}()

cases, err := cli.GetCasesParallel(pid, suites, 5, monitor)
```

## User Experience

Система прогресса использует эмодзи и понятные сообщения для лучшего UX:

```
🔍 Получение структуры проектов 30 и 34...

📥 Параллельная загрузка данных:
   Проект 30: 15 сьютов | Проект 34: 16 сьютов

⏳ Проект 30 (15 сьютов)...  [███████████████░░░░░]  75%  (12/15, 45 items/min) [2m15s:30s]
⏳ Проект 34 (16 сьютов)...  [████████████████░░░░]  80%  (13/16, 38 items/min) [5m12s:15s]

📊 Результаты загрузки:
  ✅ Проект 30: 15 сьютов → 1854 кейсов (3m15s)
  ✅ Проект 34: 16 сьютов → 24073 кейсов (8m42s)

🔍 Выполняется анализ и сверка данных...
  ✅ Анализ завершён (125ms)

Результат сохранён: /home/user/.gotr/exports/compare/compare_2026-02-21_01-00-48.json

┌──────────────────────────────────────────────────────────────┐
│          📊 СТАТИСТИКА: cases                                │
├──────────────────────────────────────────────────────────────┤
│  ⏱️  Время выполнения: 11m48s                                │
│  📦 Всего обработано: 36425                                  │
├──────────────────────────────────────────────────────────────┤
│  ✅ Только в проекте 30: 1854                                │
│  ✅ Только в проекте 34: 24073                               │
│  🔗 Общих: 10498                                             │
└──────────────────────────────────────────────────────────────┘
```

### Отладочный режим

При использовании флага `--debug` выводится подробная отладочная информация:

```bash
gotr compare cases --pid1 30 --pid2 34 --debug
```

Пример вывода с `--debug`:
```
2026/02/20 23:38:52 [DEBUG] [Compare] Fetching suites for projects 30 and 34
2026/02/20 23:38:52 [DEBUG] [Compare] Found suites: P30=15, P34=16, total=31
...
2026/02/20 23:38:55 [DEBUG] [Project 30] Processing suite 12345: 150 cases
2026/02/20 23:38:56 [DEBUG] [Project 30] Returning 1854 unique cases
```

### Значения эмодзи

| Эмодзи | Значение |
|--------|----------|
| 🔍 | Поиск / Получение структуры |
| 📥 | Загрузка данных |
| ⏳ | Ожидание / Выполнение |
| 📊 | Статистика / Результат |
| ✅ | Успешно / Только в проекте |
| 🔗 | Общие / Связанные |
| ⏱️ | Время выполнения |
| 📦 | Всего обработано |

## Performance Optimizations

### Parallel Pagination

Для ускорения загрузки больших сьютов (>250 кейсов) используется параллельная пагинация:

```
Без оптимизации:  page1 → wait → page2 → wait → page3 ... (последовательно)
С оптимизацией:   page1 → [page2, page3, page4, page5] ... (параллельно, max 5 concurrent)
```

Настройки в `internal/client/cases.go`:
- `maxConcurrency = 5` - максимум параллельных запросов страниц
- `limit = 250` - размер страницы (максимум TestRail)

### Parallel Project Loading

Оба проекта загружаются параллельно с отдельными прогресс-барами (mpb renders them on separate lines):

```go
// Project 1
go func() {
    cases1, err1 = fetchCasesForProjectWithStats(...)
}()

// Project 2  
go func() {
    cases2, err2 = fetchCasesForProjectWithStats(...)
}()
```

### Rate Limiting

TestRail API ограничен 180 req/min. Все запросы проходят через rate limiter:
- Burst: 20 запросов
- Rate: 3 запроса/сек

## Future Enhancements

- [x] Support for multiple simultaneous progress bars (mpb)
- [ ] Support for cancellable contexts
- [ ] Progress reporting with item counts (not just increments)
- [ ] Support for nested progress (parent/child bars)
- [ ] Integration with OpenTelemetry for distributed tracing

## Related Packages

- `internal/concurrent` - Worker pool with progress support
- `internal/progress` - Progress bar manager with mpb
- `internal/client` - HTTP client with progress-aware methods
