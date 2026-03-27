# Stage 13 - Матрица тестового покрытия (к цели 100%)

Дата: 2026-03-27
Ветка: stage-13.0-final-refactoring
Источник данных:

- `go test -vet=off -coverpkg=./... -coverprofile=/tmp/stage13_full.cover ./...`
- `go tool cover -func=/tmp/stage13_full.cover`

## 1) Базовый срез

- Total statements coverage: **67.4%**
- Текущий разрыв до цели 100%: **32.6%**

## 2) Пакеты с наибольшим дефицитом покрытия (по package-report)

| Пакет | Покрытие |
| --- | --- |
| github.com/Korrnals/gotr | 0.0% |
| github.com/Korrnals/gotr/internal/models/config | 0.0% |
| github.com/Korrnals/gotr/internal/selftest | 0.0% |
| github.com/Korrnals/gotr/internal/log | 0.0% |
| github.com/Korrnals/gotr/internal/paths | 0.0% |
| github.com/Korrnals/gotr/internal/ui | 0.0% |
| github.com/Korrnals/gotr/internal/client | 5.1% |
| github.com/Korrnals/gotr/internal/concurrency | 4.0% |
| github.com/Korrnals/gotr/internal/concurrent | 1.7% |
| github.com/Korrnals/gotr/cmd/* family | 1.1% - 12.0% |

## 3) Файловая матрица (приоритет по риску)

| Приоритет | Зона | Текущее avg file coverage | Zero-func count | Комментарий |
| --- | --- | --- | --- | --- |
| P0 | internal/client (attachments/projects/concurrent/extended/request) | 0% - 17% | высокий | Прямой API-контракт, ошибки влияют на runtime данные.
| P0 | internal/service (run/result/migration loader/log) | 0% - 20% | высокий | Бизнес-логика миграций и результатов.
| P1 | internal/concurrency + internal/concurrent | 0% - 75% | средний | Критично для parallel execution и reliability.
| P1 | cmd/compare + cmd/sync + cmd/resources | 0% - 75% | средний | CLI orchestration и пользовательские сценарии.
| P2 | internal/log, internal/paths, internal/models/config | 0% - 20% | средний | Инфраструктурные пакеты, легко покрываются unit-тестами.
| P3 | thin cmd wrappers (cases/configurations/datasets/...) | 0% - 10% | высокий (по count), низкий (по риску) | Большое количество оберток, можно закрывать table-driven smoke tests.

## 4) Матрица стратегий покрытия

| Track | Цель | Что тестируем | Формат тестов | Критерий готовности |
| --- | --- | --- | --- | --- |
| T1 | Закрыть 0%-файлы инфраструктуры | `internal/paths`, `internal/models/config`, `internal/log`, `internal/selftest/types` | Unit + table-driven | Все функции в этих файлах >= 95%, целевой 100% |
| T2 | API/service ядро | `internal/client/*`, `internal/service/*` | Mocked integration-style unit tests | Нулевых функций нет, package coverage >= 90% |
| T3 | Concurrency core | `internal/concurrency/*`, `internal/concurrent/*` | Deterministic concurrency tests + race-safe assertions | package coverage >= 90%, no uncovered critical branches |
| T4 | CLI command orchestration | `cmd/*` high-traffic команды | Cobra command tests + output/assertion tests | Все ветки non-interactive/interactive/save/quiet покрыты |
| T5 | Финальный добор до 100 | Оставшиеся edge branches | targeted micro-tests per function | total statements = 100.0% |

## 5) Отдельный шаг текущей стадии (Coverage 100%)

Шаг добавлен как отдельный workstream Stage 13:

- COV-1: baseline + матрица (этот документ) — DONE
- COV-2: закрыть все 0%-файлы в internal/*
- COV-3: довести internal/client + internal/service до 95%+
- COV-4: довести internal/concurrency + internal/concurrent до 95%+
- COV-5: закрыть cmd/* thin wrappers массовыми table-driven тестами
- COV-6: финальный проход до `total: 100.0%`

## 6) Gate для шага покрытия

- `go test ./...` == PASS
- `CGO_ENABLED=1 go test -race ./...` == PASS
- `go test -vet=off -coverpkg=./... -coverprofile=/tmp/stage13_full.cover ./...` == PASS
- `go tool cover -func=/tmp/stage13_full.cover` == `total: 100.0%`

## 7) Риски и реалистичность

- Текущая цель 100% достижима только при массовом расширении unit-тестов почти для всех thin wrappers в `cmd/*`.
- Для контроля времени внедряем wave-подход (T1→T5), но финальный критерий остается жестким: 100.0%.
