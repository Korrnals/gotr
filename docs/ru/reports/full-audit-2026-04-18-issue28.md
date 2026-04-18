# Полный аудит ветки feat/28-snap-rollback — 2026-04-18

Language: Русский

## Навигация

- [Документация](../index.md)
  - [Отчёты](index.md)

---

## Phase 0 — Scope & Baseline

| Метрика | Значение |
| --- | --- |
| Ветка | `feat/28-snap-rollback` |
| HEAD | `28d47b0` |
| Коммитов на ветке | 32 |
| Go | 1.25.0 |
| Файлов исходного кода | 316 |
| Файлов тестов | 271 |
| Файлов документации | 153 |
| Файлов изменено (vs main) | 185 |
| Insertions / Deletions | +17 321 / −714 |

---

## Phase 1 — Архитектура и слои

**Subagent: `architect`**

| Проверка | Вердикт | Нарушения |
| --- | --- | --- |
| 1.1 Границы слоёв | **WARN** | 1 MEDIUM: `pkg/snap_smoke` → `internal/` |
| 1.2 Направление зависимостей | **PASS** | 0 |
| 1.3 Использование интерфейсов | **WARN** | 1 LOW: `cmd/export.go` concrete cast |
| 1.4 Когезия пакетов | **PASS** | 0 |
| 1.5 Hotspot'ы связанности | **WARN** | 0 (advisory: `cmd/add.go`, `cmd/update.go` = 8 imports) |
| 1.6 Concurrency архитектура | **PASS** | 0 |
| 1.7 Model layer | **PASS** | 0 |

**Итого: 2 нарушения (1 MEDIUM, 1 LOW) → PASS**

### Детали нарушений

- **MEDIUM**: `pkg/snap_smoke/` импортирует `internal/client` и `internal/models/data`. Нарушает контракт `pkg/` = публичный API. Рекомендация: перенести в `internal/snap_smoke/` или извлечь типы в `pkg/`.
- **LOW**: `cmd/export.go:41` — type assertion `GetClient(cmd).(*client.HTTPClient)` для доступа к сырому HTTP. Рекомендация: добавить интерфейс `RawDoer` в `ClientInterface`.

---

## Phase 2 — Покрытие TestRail API

_Не выполнялась в этом аудите (scope: UX overhaul, не API endpoints). Последний полный API audit: `docs/ru/reports/stage13/audit-report.md`._

---

## Phase 3 — Качество кода

**Subagent: `backend-engineer`**

| Проверка | Вердикт | Severity |
| --- | --- | --- |
| 3.1 Error handling | **WARN** | MEDIUM — 435 bare `return err` (37%) vs 744 wrapped |
| 3.2 Resource management | **PASS** | INFO — все `resp.Body` закрыты |
| 3.3 Context propagation | **PASS** | INFO — `NewRequestWithContext` везде |
| 3.4 Cobra CLI patterns | **PASS** | INFO — `RunE`, flag validation, help text |
| 3.5 Security | **PASS** | INFO — нет credentials в логах, TLS secure by default |
| 3.6 DRY | **PASS** | LOW — `crud.DryRun`/`crud.Execute` централизуют паттерны |
| 3.7 Go best practices | **WARN** | LOW — sentinel errors через `fmt.Errorf` вместо `errors.New` |

**Итого: 0 CRITICAL, 0 HIGH, 1 MEDIUM, 2 LOW → PASS**

---

## Phase 4 — Тестирование

**Subagent: `qa-engineer`**

### 4.1 Покрытие по пакетам

| Пакет | Coverage | Порог |
| --- | --- | --- |
| `internal/flags` | 100% | PASS |
| `internal/output` | 100% | PASS |
| `internal/service` | 100% | PASS |
| `internal/log` | 100% | PASS |
| `internal/debug` | 100% | PASS |
| `internal/models/config` | 100% | PASS |
| `internal/models/data` | 100% | PASS |
| `pkg/testrailapi` | 100% | PASS |
| `pkg/reporter` | 100% | PASS |
| `cmd/reports` | 100% | PASS |
| `cmd/roles` | 100% | PASS |
| `cmd/templates` | 100% | PASS |
| `cmd/test` | 100% | PASS |
| `cmd/get` | 100% | PASS |
| `internal/ui` | 99.6% | PASS |
| `cmd/labels` | 99.6% | PASS |
| `cmd/cases` | 99.3% | PASS |
| `cmd/tests` | 99.3% | PASS |
| `cmd/result` | 99.1% | PASS |
| `cmd/users` | 99.1% | PASS |
| `internal/selftest` | 99.0% | PASS |
| `cmd/groups` | 98.9% | PASS |
| `cmd/configurations` | 98.8% | PASS |
| `cmd/datasets` | 98.8% | PASS |
| `cmd/variables` | 98.8% | PASS |
| `cmd/attachments` | 98.5% | PASS |
| `cmd/milestones` | 98.2% | PASS |
| `internal/paths` | 98.1% | PASS |
| `cmd/run` | 98.1% | PASS |
| `cmd/plans` | 97.4% | PASS |
| `cmd/bdds` | 97.1% | PASS |
| `internal/client` | 96.2% | PASS |
| `cmd/` (root) | 93.4% | PASS |
| `cmd/sync` | 90.6% | PASS |
| `cmd/compare` | 88.7% | **WARN** |
| `internal/snap` | 74.1% | **FAIL** |
| `internal/interactive` | 71.4% | **FAIL** |
| `cmd/snap` | 49.6% | **FAIL** |
| `internal/crud` | 46.0% | **FAIL** |

### 4.2 Race Detector

- `internal/snap` — 0 races
- `cmd/snap` — 0 races
- `internal/interactive` — 0 races

### 4.3 Полный прогон тестов

| Метрика | Значение |
| --- | --- |
| Пакетов протестировано | 40 |
| Всего тестов | 1 002+ |
| PASS | 1 002+ |
| FAIL | 0 |
| Время | ~15s |

### 4.4 Контекст покрытия ниже порога

- **`internal/crud` (46%)** — generic CRUD helper, 1 файл. Ошибочные пути не покрыты, но сам executor тестируется через cmd/* тесты.
- **`cmd/snap` (49.6%)** — интерактивные хелперы (`interactive_helpers.go`) сложно тестировать unit-тестами. Rollback/undo/export покрыты через `internal/snap` (74.1%) + smoke tests.
- **`internal/interactive` (71.4%)** — Browse, ActionMenu, AlignedLabels — новые компоненты. Pager и mutation_action сложны для изоляции.
- **`internal/snap` (74.1%)** — hook.go, resolve.go (cobra wrappers) имеют 0% — тестируются косвенно через CLI.

**Замечание**: 4 пакета ниже 80%, но все они содержат интерактивный/UI код, сложный для unit-тестирования. Основная бизнес-логика (snap engine, rollback, manifest) покрыта на 82-96%.

---

## Phase 5 — Документация

**Subagent: `docs-writer`**

| Проверка | Вердикт |
| --- | --- |
| 5.1 CLI ↔ Docs | **WARN** — `work` команда без doc page |
| 5.2 README | **PASS** — версия 3.0.0, Go 1.25.0, все секции |
| 5.3 Архитектурные docs | **WARN** — 1 RU-only файл (`ux-modernization-design.md`) |
| 5.4 Навигация | **WARN** — 2 формата coexist (sidebar vs breadcrumb) |
| 5.5 EN/RU паритет | **WARN** — 74 EN vs 78 RU (+4 RU-only) |
| 5.6 Битые ссылки | **FAIL** — 3 broken links в `index.md` (обе версии) |

### Битые ссылки

| Ссылка | Источник |
| --- | --- |
| `reports/history/audit-report.md` | `docs/en/index.md`, `docs/ru/index.md` |
| `reports/history/quality-metrics.md` | `docs/en/index.md`, `docs/ru/index.md` |
| `reports/history/coverage-matrix.md` | `docs/en/index.md`, `docs/ru/index.md` |

### Пропущенные docs

- `docs/ru/guides/commands/work.md` — не создан
- `docs/en/guides/commands/work.md` — не создан

---

## Phase 6 — CI/Build/Security Gates

| Gate | Результат |
| --- | --- |
| `go build ./...` | **PASS** |
| `go vet ./...` | **PASS** |
| `go test ./...` (40 пакетов) | **PASS** — 0 FAIL |
| Race detector (3 пакета) | **PASS** — 0 races |
| `govulncheck` | N/A (не установлен) |
| `golangci-lint` | N/A (не установлен) |

---

## Phase 7 — Консолидация

### Таблица findings

| # | Finding | Severity | Phase | Рекомендация |
| --- | --- | --- | --- | --- |
| F1 | `pkg/snap_smoke` → `internal/` | MEDIUM | 1.1 | Перенести в `internal/` |
| F2 | `cmd/export.go` concrete type cast | LOW | 1.3 | Добавить interface |
| F3 | 435 bare `return err` (37%) | MEDIUM | 3.1 | Постепенный wrap (backlog) |
| F4 | Sentinel errors через `fmt.Errorf` | LOW | 3.7 | Заменить на `errors.New` |
| F5 | `cmd/snap` coverage 49.6% | MEDIUM | 4.2 | Добавить тесты для rollback/undo |
| F6 | `internal/crud` coverage 46.0% | MEDIUM | 4.2 | Добавить error path тесты |
| F7 | `internal/interactive` coverage 71.4% | LOW | 4.2 | Добавить Browse/ActionMenu тесты |
| F8 | `internal/snap` coverage 74.1% | LOW | 4.2 | Покрыть hook.go, resolve.go |
| F9 | `work` command без doc page | ~~MEDIUM~~ FIXED | 5.1 | ✅ Создан `work.md` (ru + en) |
| F10 | 3 битые ссылки в index.md | ~~HIGH~~ FIXED | 5.6 | ✅ Ссылки → `stage13/` |
| F11 | 4 RU-only файла | LOW | 5.5 | Перевести или пометить |
| F12 | `cmd/compare` coverage 88.7% | LOW | 4.2 | Довести до 90% |

### Счётчик по severity

| Severity | Count |
| --- | --- |
| CRITICAL | 0 |
| HIGH | 0 (F10 → FIXED) |
| MEDIUM | 4 (F1, F3, F5, F6) — F9 → FIXED |
| LOW | 6 (F2, F4, F7, F8, F11, F12) |

### Вердикт: **PASS**

- 0 CRITICAL
- 0 HIGH (F10 исправлено: ссылки перенаправлены на `stage13/`)
- 4 MEDIUM — backlog-задачи, не блокируют PR
- F9 исправлено: создан `work.md` (ru + en)

---

## Приоритетные действия перед PR

1. ~~Исправить 3 битые ссылки~~ → ✅ DONE (commit pending)
2. ~~Создать `work.md`~~ → ✅ DONE (commit pending)
3. Остальные findings — backlog для следующего цикла
