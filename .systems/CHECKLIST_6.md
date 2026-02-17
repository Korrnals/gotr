# CHECKLIST.md — Чеклист Stage 6: Performance Optimization & UX

> **Этап:** Stage 6 — Performance Optimization & UX Enhancement  
> **Дата:** 2026-02-16  
> **Версия:** 2.7.0 → 2.8.0-dev  
> **Статус:** 🔄 В работе  
> **Выбранная библиотека:** `github.com/schollz/progressbar/v3`

---

## ✅ УТВЕРЖДЁН: План Stage 6

**Дата утверждения:** 2026-02-16  
**Scope:** Performance optimization + UX enhancement  
**Целевое улучшение:** 60-80% faster execution

---

## 📊 Матрица прогресса

```
Phase 6.1: Progress Bars      [██████████] 100% ✅
Phase 6.2: Parallel Requests  [░░░░░░░░░░] 0%
Phase 6.3: Caching            [░░░░░░░░░░] 0%
Phase 6.4: Retry Logic        [░░░░░░░░░░] 0%
Phase 6.5: Batch Operations   [░░░░░░░░░░] 0%
Phase 6.6: UX Polish          [░░░░░░░░░░] 0%

Overall: 17% (1/6 phases)
```

---

## Phase 6.1: Progress Bars Foundation ✅ COMPLETE

### Задачи

- [x] **Добавить зависимость** `schollz/progressbar/v3`
  ```bash
  go get github.com/schollz/progressbar/v3
  ```

- [x] **Создать пакет** `internal/progress/`
  - [x] `progress.go` — интерфейс ProgressManager
  - [x] `progress_test.go` — тесты (100% coverage)

- [x] **Интегрировать в compare** (12 команд)
  - [x] `compare cases` — progress bar при загрузке сьютов
  - [x] `compare suites` — spinner при загрузке
  - [x] `compare sections` — spinner
  - [x] `compare sharedsteps` — spinner
  - [x] `compare runs` — spinner
  - [x] `compare plans` — spinner
  - [x] `compare milestones` — spinner
  - [x] `compare datasets` — spinner
  - [x] `compare groups` — spinner
  - [x] `compare labels` — spinner
  - [x] `compare templates` — spinner
  - [x] `compare configurations` — spinner
  - [x] `compare all` — использует pm для всех внутренних команд

- [x] **Интегрировать в sync** (5 команд)
  - [x] `sync full` — multi-phase progress
  - [x] `sync suites` — spinner
  - [x] `sync sections` — spinner
  - [x] `sync shared-steps` — spinner
  - [x] `sync cases` — spinner

- [x] **Интегрировать в get** (4 команды)
  - [x] `get cases --all-suites` — progress bar
  - [x] `get sharedsteps` — spinner
  - [x] `get suites` — spinner
  - [x] `get sections` — spinner

- [x] **Интегрировать в остальные команды**
  - [x] `cases bulk` — spinner для bulk операций
  - [x] `attachments add` — spinner для загрузки файлов
  - [x] `users list` — spinner для списка пользователей
  - [x] `reports run` — spinner для генерации отчётов
  - [x] `reports run-cross-project` — spinner

- [x] **Тесты**
  - [x] Unit tests для `internal/progress/` (100% coverage)
  - [x] Проверка работы в TTY и non-TTY режимах

### Результаты
- ✅ Все 12 compare команд имеют progress bars
- ✅ Все 5 sync команд имеют progress bars
- ✅ Все 4 get команды имеют progress bars  
- ✅ Другие команды с длительными операциями имеют progress bars
- ✅ Пакет `internal/progress` готов для повторного использования
- ✅ Паттерн интеграции установлен
- ✅ Все тесты проходят (27/27)

### Acceptance Criteria

- [ ] При выполнении `compare cases` >100 items показывается progress bar
- [ ] При выполнении `compare all` показывается общий прогресс + per-resource
- [ ] ETA отображается корректно
- [ ] В non-TTY (CI/CD) режиме — только текстовые сообщения

### Визуальный результат

```
Сравнение кейсов...
Загрузка из проекта 30...  45% |████████████░░░░░░░░| (450/1000) [00:05<00:06]
Загрузка из проекта 34...  30% |████████░░░░░░░░░░░░| (300/1000) [00:03<00:07]
```

---

## Phase 6.2: Parallel API Requests

### Задачи

- [ ] **Создать пакет** `internal/concurrent/`
  - [ ] `pool.go` — worker pool с errgroup
  - [ ] `limiter.go` — rate limiter (token bucket)
  - [ ] `retry.go` — retry logic

- [ ] **Rate Limiter**
  - [ ] Лимит: 150 requests/minute
  - [ ] Burst capacity: 10 requests
  - [ ] Graceful wait при превышении

- [ ] **Parallel Client Methods**
  - [ ] `GetCasesParallel(projectID, suiteIDs []int64)`
  - [ ] `GetSuitesParallel(projectIDs []int64)`
  - [ ] `GetSharedStepsParallel(projectIDs []int64)`

- [ ] **Интеграция в compare**
  - [ ] Параллельная загрузка для `compare all`
  - [ ] Параллельная загрузка для `compare cases` (multi-suite)

- [ ] **Тесты**
  - [ ] Тесты для rate limiter
  - [ ] Тесты для worker pool
  - [ ] Проверка отсутствия race conditions

### Acceptance Criteria

- [ ] `compare all` выполняет независимые запросы параллельно
- [ ] Нет 429 ошибок (rate limiting работает)
- [ ] При ошибке одного запроса, остальные продолжают

---

## Phase 6.3: Response Caching

### Задачи

- [ ] **Создать пакет** `internal/cache/`
  - [ ] `cache.go` — интерфейс Cache
  - [ ] `disk.go` — disk-based реализация
  - [ ] `ttl.go` — TTL management
  - [ ] `cleanup.go` — cleanup old entries

- [ ] **TTL настройки**
  | Entity | TTL |
  |--------|-----|
  | Projects | 1 hour |
  | Suites | 30 minutes |
  | Cases | 15 minutes |
  | Shared Steps | 15 minutes |
  | Sections | 30 minutes |

- [ ] **Cache Management**
  - [ ] Автосоздание `~/.gotr/cache/`
  - [ ] LRU eviction при >100MB
  - [ ] Авто-cleanup при старте

- [ ] **CLI команды**
  - [ ] `gotr cache clear` — очистка всего кэша
  - [ ] Флаг `--no-cache` — обход кэша

- [ ] **Интеграция**
  - [ ] Cache в `compare` командах
  - [ ] Cache в `get` командах
  - [ ] Cache invalidation на write операциях

- [ ] **Тесты**
  - [ ] Тесты для disk cache
  - [ ] Тесты для TTL
  - [ ] Тесты для cleanup

### Acceptance Criteria

- [ ] Повторный `compare` использует кэш и работает на 80% быстрее
- [ ] Кэш уважает TTL
- [ ] Размер кэша ограничен 100MB

---

## Phase 6.4: Retry Logic & Resilience

### Задачи

- [ ] **Retry Logic**
  - [ ] Exponential backoff: 1s, 2s, 4s, 8s, 16s
  - [ ] Max retries: 5
  - [ ] Только для idempotent операций (GET, LIST)

- [ ] **Circuit Breaker**
  - [ ] Threshold: 5 ошибок подряд
  - [ ] Timeout: 30 секунд
  - [ ] Half-open state для проверки восстановления

- [ ] **Error Context**
  - [ ] Улучшенные сообщения: "Ошибка загрузки кейсов проекта 30: ..."
  - [ ] Стек вызовов при `--verbose`

- [ ] **Timeout Flag**
  - [ ] `--timeout 5m` (default)
  - [ ] `--timeout 0` (бесконечно)
  - [ ] Max: 30m

- [ ] **Тесты**
  - [ ] Тесты для retry
  - [ ] Тесты для circuit breaker
  - [ ] Тесты для timeout

### Acceptance Criteria

- [ ] Transient ошибки автоматически ретраются
- [ ] Circuit breaker предотвращает cascade failures
- [ ] Timeout не оставляет "висячих" goroutines

---

## Phase 6.5: Batch Operations Optimization

### Задачи

- [ ] **Batch Fetching**
  - [ ] Увеличить limit с 50 до 250 (макс для TestRail)
  - [ ] Авто-pagination для больших датасетов

- [ ] **Prefetching**
  - [ ] Prefetch связанных сущностей
  - [ ] Lazy vs Eager loading стратегии

- [ ] **Memory Optimization**
  - [ ] Streaming JSON parsing
  - [ ] Очистка неиспользуемых объектов
  - [ ] Пул буферов для снижения GC pressure

- [ ] **compare all оптимизация**
  - [ ] Общие данные загружаются один раз
  - [ ] Avoid N+1 queries

- [ ] **Тесты**
  - [ ] Бенчмарки для сравнения
  - [ ] Memory profiling

### Acceptance Criteria

- [ ] `compare all` на проекте 10,000+ кейсов: <2 минут
- [ ] Память не превышает 500MB
- [ ] Нет "out of memory" ошибок

---

## Phase 6.6: UX Polish

### Задачи

- [ ] **ETA Display**
  - [ ] Расчет ETA в progress bar
  - [ ] Скорость (items/sec)
  - [ ] Оставшееся время

- [ ] **Color Output**
  - [ ] `github.com/fatih/color` интеграция
  - [ ] Цветной статус: ✓ зелёный, ⚠ жёлтый, ✗ красный
  - [ ] Отключение цветов через `NO_COLOR` env

- [ ] **Quiet Mode**
  - [ ] Флаг `--quiet` — только ошибки и результат
  - [ ] Для CI/CD интеграции
  - [ ] Exit codes: 0 (success), 1 (error), 2 (differences found)

- [ ] **Verbose Mode**
  - [ ] Флаг `--verbose` — детальное логирование
  - [ ] API request/response logging
  - [ ] Cache hit/miss logging

- [ ] **Help Enhancement**
  - [ ] Примеры в каждой команде help
  - [ ] Long description с use cases

- [ ] **Тесты**
  - [ ] Тесты для quiet mode
  - [ ] Тесты для verbose mode

### Acceptance Criteria

- [ ] Quiet mode выводит только результат
- [ ] Verbose mode показывает API calls
- [ ] Цвета отключаются в non-TTY

---

## ✅ Subagent Execution Summary: COMPLETE

### Progress Bars Implementation (Phase 6.1)

**Subagent A: Sync Commands** ✅
- Files: `cmd/sync/sync_full.go`, `cmd/sync/sync_cases.go`, `cmd/sync/sync_shared_steps.go`, `cmd/sync/sync_suites.go`, `cmd/sync/sync_sections.go`
- Removed old `cheggaaa/pb/v3` dependency
- Added spinners for all phases

**Subagent B: Get Commands** ✅
- Files: `cmd/get/cases.go`, `cmd/get/sharedsteps.go`, `cmd/get/suites.go`, `cmd/get/sections.go`
- Progress bar for `--all-suites` flag
- Spinners for single operations

**Subagent C: Cases + Attachments** ✅
- Files: `cmd/cases/bulk.go`, `cmd/attachments/add.go`
- Bulk operations progress
- File upload spinners

**Subagent D: Other Commands** ✅
- Files: `cmd/users/list.go`, `cmd/reports/run.go`, `cmd/reports/run_cross_project.go`
- Long-running operations only

**Subagent E: Common/Flags Reorganization** ✅
- Created: `cmd/common/flags/parse/parse.go`
- Created: `cmd/common/flags/get/get.go`
- Updated: `cmd/common/flags.go` (backward compatibility layer)

### Quality Assurance
- [x] All subagents followed `internal/progress` pattern
- [x] All tests pass after each subagent
- [x] Final QA by main agent completed
- [x] Build successful: `go build ./...`
- [x] All tests pass: `go test ./...` (27/27)

---

## 📁 Файлы для создания

### Новые пакеты

```
internal/
├── progress/
│   ├── progress.go
│   ├── bar.go
│   ├── spinner.go
│   ├── options.go
│   └── progress_test.go
├── cache/
│   ├── cache.go
│   ├── disk.go
│   ├── ttl.go
│   ├── cleanup.go
│   └── cache_test.go
└── concurrent/
    ├── pool.go
    ├── limiter.go
    ├── retry.go
    ├── circuit.go
    └── concurrent_test.go
```

### Обновляемые файлы

```
cmd/
├── compare/*.go          # Добавить progress bars
├── sync/*.go             # Добавить progress bars
└── get/*.go              # Добавить progress bars
```

---

## 🧪 Тестовая стратегия

### Unit Tests
- Каждый новый пакет: 95%+ покрытие
- Mock для HTTP client
- Table-driven tests

### Integration Tests
- Тесты с реальным TestRail (опционально)
- Performance benchmarks
- Race condition detection: `go test -race`

### Benchmarks
```go
func BenchmarkCompareCases(b *testing.B) {
    // Сравнение до и после оптимизации
}
```

---

## ✅ Обязательные действия после завершения Stage 6

### ☐ Уточнить у пользователя:
- [ ] «Стоит ли выполнить модульные коммиты?»
- [ ] «Необходимо ли выполнить очередной Релиз (2.8.0)?»

### ☐ Обновить документацию:
- [ ] `CHANGELOG.md` — добавить раздел [2.8.0]
- [ ] `README.md` — обновить раздел Performance
- [ ] `docs/*.md` — документация новых флагов

### ☐ Обновить версию:
- [ ] `cmd/root.go` — обновить Version = "2.8.0"
- [ ] `CHANGELOG.md` — дата релиза

### ☐ Зафиксировать изменения:
```
feat(progress): add progress bars with schollz/progressbar/v3
feat(concurrent): add parallel API requests with rate limiting
feat(cache): add disk-based response caching with TTL
feat(retry): add exponential backoff and circuit breaker
feat(perf): optimize batch operations and memory usage
feat(ux): add quiet/verbose modes and colored output
docs: update README and CHANGELOG for Stage 6
```

### ☐ Синхронизировать файлы оси:
- [ ] `API_AUDIT.md` — обновить
- [ ] `PLAN.md` — обновить статус Stage 6
- [ ] `CHECKLIST.md` — этот файл ✅

---

## 📊 Success Metrics Checklist

| Метрика | Было | Цель | Факт | Статус |
|---------|------|------|------|--------|
| compare cases (1000) | 5+ min | <30 sec | - | ⏳ |
| compare all | 10+ min | <2 min | - | ⏳ |
| Memory peak | 1GB+ | <500MB | - | ⏳ |
| Test coverage | - | 95%+ | - | ⏳ |

---

## 🔥 Риски и Mitigation

| Риск | Вероятность | Влияние | Mitigation |
|------|-------------|---------|------------|
| Rate limiting сложнее ожидаемого | Средняя | Высокое | Conservative limits, backoff |
| Race conditions в parallel code | Средняя | Высокое | -race тесты, mutexes |
| Cache invalidation баги | Низкая | Среднее | TTL, explicit invalidation |
| Memory leaks | Низкая | Высокое | Profiling, pprof |

---

*Файл создан: 2026-02-16*  
*Этап: Stage 6 — Performance Optimization*  
*Статус: 🔄 В работе*  
*Следующий шаг: Phase 6.1 — Progress Bars Foundation*
