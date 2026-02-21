package client

import (
	"fmt"
	"sync"

	"github.com/Korrnals/gotr/internal/concurrent"
	"github.com/Korrnals/gotr/internal/models/data"
)

// defaultWorkers — количество параллельных воркеров по умолчанию
const defaultWorkers = 5

// GetCasesParallel получает кейсы из нескольких сьютов параллельно.
// Использует WorkerPool для ограничения параллелизма и RateLimiter
// для соблюдения лимита запросов к API (150 req/min).
//
// Параметры:
//   - projectID: ID проекта
//   - suiteIDs: список ID сьютов для загрузки
//   - workers: количество параллельных воркеров (0 = defaultWorkers)
//   - monitor: опциональный монитор прогресса (может быть nil)
//
// Возвращает:
//   - map[suiteID] => список кейсов
//   - ошибку, если хотя бы один запрос не удался
//
// Пример без прогресса:
//
//	cases, err := client.GetCasesParallel(30, []int64{1, 2, 3}, 5, nil)
//	if err != nil {
//	    log.Printf("Some suites failed: %v", err)
//	}
//
// Пример с прогресс-баром:
//
//	progressChan := make(chan int, 100)
//	monitor := progress.NewMonitor(progressChan, len(suiteIDs))
//	go func() {
//	    for range progressChan {
//	        bar.Add(1)
//	    }
//	}()
//	cases, err := client.GetCasesParallel(30, suiteIDs, 5, monitor)
func (c *HTTPClient) GetCasesParallel(projectID int64, suiteIDs []int64, workers int, monitor ProgressMonitor) (map[int64]data.GetCasesResponse, error) {
	if len(suiteIDs) == 0 {
		return make(map[int64]data.GetCasesResponse), nil
	}

	if workers <= 0 {
		workers = defaultWorkers
	}

	// Результаты
	results := make(map[int64]data.GetCasesResponse, len(suiteIDs))
	var mu sync.Mutex

	// Ошибки
	var errs []error
	var errMu sync.Mutex

	// Worker pool с ограничением, rate limiter и монитором прогресса
	opts := []concurrent.PoolOption{
		concurrent.WithMaxWorkers(workers),
		concurrent.WithRateLimit(180),
	}
	if monitor != nil {
		opts = append(opts, concurrent.WithProgressMonitor(monitor))
	}
	pool := concurrent.NewWorkerPool(opts...)

	// Запускаем задачи
	for _, suiteID := range suiteIDs {
		sid := suiteID // Захватываем переменную
		pool.Submit(func() error {
			// Выполняем запрос (без внутреннего прогресса, только через pool)
			cases, err := c.GetCases(projectID, sid, 0)
			if err != nil {
				errMu.Lock()
				errs = append(errs, fmt.Errorf("suite %d: %w", sid, err))
				errMu.Unlock()
				return err
			}

			// Сохраняем результат
			mu.Lock()
			results[sid] = cases
			mu.Unlock()

			return nil
		})
	}

	// Ждем завершения всех задач
	if err := pool.Wait(); err != nil {
		return results, fmt.Errorf("parallel execution failed: %w", err)
	}

	// Если были ошибки, возвращаем их
	if len(errs) > 0 {
		return results, fmt.Errorf("partial failure: %d/%d suites failed", len(errs), len(suiteIDs))
	}

	return results, nil
}

// GetSuitesParallel получает сьюты из нескольких проектов параллельно.
// Полезно для команд compare all, когда нужно получить сьюты из двух проектов.
//
// Параметры:
//   - projectIDs: список ID проектов
//   - workers: количество параллельных воркеров (0 = defaultWorkers)
//   - monitor: опциональный монитор прогресса (может быть nil)
//
// Возвращает:
//   - map[projectID] => список сьютов
//   - ошибку, если хотя бы один запрос не удался
func (c *HTTPClient) GetSuitesParallel(projectIDs []int64, workers int, monitor ProgressMonitor) (map[int64]data.GetSuitesResponse, error) {
	if len(projectIDs) == 0 {
		return make(map[int64]data.GetSuitesResponse), nil
	}

	if workers <= 0 {
		workers = defaultWorkers
	}

	// Результаты
	results := make(map[int64]data.GetSuitesResponse, len(projectIDs))
	var mu sync.Mutex

	// Ошибки
	var errs []error
	var errMu sync.Mutex

	// Worker pool с опциональным монитором
	opts := []concurrent.PoolOption{
		concurrent.WithMaxWorkers(workers),
		concurrent.WithRateLimit(180),
	}
	if monitor != nil {
		opts = append(opts, concurrent.WithProgressMonitor(monitor))
	}
	pool := concurrent.NewWorkerPool(opts...)

	// Запускаем задачи
	for _, projectID := range projectIDs {
		pid := projectID // Захватываем переменную
		pool.Submit(func() error {
			// Выполняем запрос
			suites, err := c.GetSuites(pid)
			if err != nil {
				errMu.Lock()
				errs = append(errs, fmt.Errorf("project %d: %w", pid, err))
				errMu.Unlock()
				return err
			}

			// Сохраняем результат
			mu.Lock()
			results[pid] = suites
			mu.Unlock()

			return nil
		})
	}

	// Ждем завершения
	if err := pool.Wait(); err != nil {
		return results, fmt.Errorf("parallel execution failed: %w", err)
	}

	// Если были ошибки
	if len(errs) > 0 {
		return results, fmt.Errorf("partial failure: %d/%d projects failed", len(errs), len(projectIDs))
	}

	return results, nil
}

// GetCasesForSuitesParallel получает все кейсы для списка сьютов одного проекта.
// Объединяет результаты в плоский список кейсов.
//
// Параметры:
//   - projectID: ID проекта
//   - suiteIDs: список ID сьютов
//   - workers: количество параллельных воркеров
//   - monitor: опциональный монитор прогресса (может быть nil)
//
// Возвращает:
//   - плоский список всех кейсов из всех сьютов
//   - ошибку, если хотя бы один запрос не удался
func (c *HTTPClient) GetCasesForSuitesParallel(projectID int64, suiteIDs []int64, workers int, monitor ProgressMonitor) (data.GetCasesResponse, error) {
	if len(suiteIDs) == 0 {
		return data.GetCasesResponse{}, nil
	}

	// Получаем кейсы параллельно
	results, err := c.GetCasesParallel(projectID, suiteIDs, workers, monitor)
	if err != nil && len(results) == 0 {
		return nil, err
	}

	// Объединяем результаты в плоский список
	var allCases data.GetCasesResponse
	for _, suiteCases := range results {
		allCases = append(allCases, suiteCases...)
	}

	return allCases, err
}
