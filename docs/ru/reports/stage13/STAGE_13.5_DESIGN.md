# Stage 13.5 — Full API Coverage & 100% Test Parity

Language: Русский | [English](../../../en/reports/stage13/STAGE_13.5_DESIGN.md)

## Навигация

- [Документация](../../index.md)
  - [Гайды](../../guides/index.md)
  - [Архитектура](../../architecture/index.md)
  - [Эксплуатация](../../operations/index.md)
  - [Отчёты](../index.md)
    - [Stage 13](index.md)
- [Главная](../../../../README_ru.md)

---

## Цель стадии

Довести gotr v3.0 до полного покрытия TestRail API v2 CLI-обёртками и достичь 100% test coverage по всем пакетам. По завершении — обязательный полный повторный аудит по шаблону `.github/prompts/full-project-audit.prompt.md`.

## Предпосылки

По результатам финального аудита Stage 13.0 (2026-04-06):

- **42/42 пакетов PASS** с `-race`, ноль data races
- **Покрытие**: 36 пакетов @ 100%, 6 пакетов @ 96.8-98.7%
- **API клиент**: 147 методов реализовано (суперсет TestRail API v2)
- **CLI exposure**: ~90.5%, 14 клиентских методов без прямой CLI-обёртки
- **api_paths.go**: ~15% эндпоинтов отсутствуют в реестре

---

## Phase 1 — Полнота api_paths.go (документация эндпоинтов)

### 1.1. Аудит и дополнение api_paths.go

Текущее состояние: 128 эндпоинтов задокументированы. Нужно добавить ~15 отсутствующих:

| Группа | Отсутствующие эндпоинты | Тип |
|--------|------------------------|-----|
| Attachments | `get_attachment/{id}`, `get_attachments_for_case/{id}`, `get_attachments_for_plan/{id}`, `get_attachments_for_plan_entry/{id}`, `get_attachments_for_run/{id}`, `get_attachments_for_test/{id}`, `get_attachments_for_project/{id}` | GET |
| Users | `add_user`, `update_user/{id}`, `get_users/{project_id}` | POST/GET |
| Reports | `get_cross_project_reports` | GET |
| Labels | `get_label/{id}`, `get_labels/{project_id}` | GET |

- [ ] Добавить все отсутствующие эндпоинты в `pkg/testrailapi/api_paths.go`
- [ ] Обновить тесты api_paths_test.go (добавить проверку новых путей)
- [ ] Commit: `feat(api): complete api_paths.go endpoint registry`

---

## Phase 2 — CLI-обёртки для недостающих операций

### 2.1. User Management CLI (HIGH)

Текущее состояние: клиентские методы `AddUser()`, `UpdateUser()`, `GetUsersByProject()` реализованы, CLI отсутствует.

- [ ] `gotr users add --name --email --role-id` — обёртка над `AddUser()`
- [ ] `gotr users update <user_id> --name --email --role-id` — обёртка над `UpdateUser()`
- [ ] `gotr users list --project-id` — поддержка `GetUsersByProject()` через existing list
- [ ] Добавить интерактивный режим для `users add/update`
- [ ] Тесты: table-driven для add/update/list-by-project (quiet + JSON + interactive)
- [ ] Commit: `feat(users): add/update CLI commands with interactive mode`

### 2.2. Reference Data CLI (MEDIUM)

Методы `GetPriorities()`, `GetStatuses()`, `GetResultFields()` реализованы в клиенте, но не имеют удобных CLI-обёрток.

- [ ] `gotr list priorities` — эндпоинт через generic list
- [ ] `gotr list statuses` — эндпоинт через generic list
- [ ] `gotr list resultfields` — эндпоинт через generic list
- [ ] Проверить что generic `list` уже маршрутизирует на эти ресурсы; если нет — добавить
- [ ] Тесты: table-driven для каждого нового ресурса
- [ ] Commit: `feat(list): expose priorities/statuses/resultfields via generic list`

### 2.3. Attachments Get/List-by-Context (MEDIUM)

Клиентские методы `GetAttachment()`, `GetAttachmentsFor*()` — 7 методов, ограниченный доступ из CLI.

- [ ] `gotr attachments get <attachment_id>` — download/metadata
- [ ] `gotr attachments list --for-case <id>` / `--for-run <id>` / `--for-plan <id>` / `--for-test <id>` / `--for-project <id>` — контекстные списки
- [ ] Тесты: table-driven для каждого контекста
- [ ] Commit: `feat(attachments): context-aware list and get commands`

### 2.4. Cross-Project Reports (LOW)

Клиентский метод `GetCrossProjectReports()` и `RunCrossProjectReport()` реализованы.

- [ ] Убедиться что `gotr reports list-cross` и `gotr reports run-cross` корректно работают
- [ ] Добавить тесты если отсутствуют
- [ ] Commit: `test(reports): cross-project report coverage`

---

## Phase 3 — Покрытие тестами до 100%

### 3.1. Пакеты с покрытием < 100%

| Пакет | Текущее | Цель | Дельта | Стратегия |
|-------|---------|------|--------|-----------|
| `cmd/sync` | 96.8% | 100% | +3.2% | Покрыть error-ветки, edge cases в sync_full/sync_cases |
| `cmd/get` | 96.9% | 100% | +3.1% | Добавить тесты для непокрытых branches в get/* |
| `cmd/run` | 97.1% | 100% | +2.9% | Error paths в create/update/close |
| `cmd` (root) | 97.3% | 100% | +2.7% | Непокрытые ветки в commands.go/root.go |
| `cmd/result` | 97.6% | 100% | +2.4% | service_wrapper.go error paths |
| `internal/ui` | 98.7% | 100% | +1.3% | Preview edge cases, format edge paths |

- [ ] Для каждого пакета: определить непокрытые строки через `go test -coverprofile`
- [ ] Написать точечные тесты на каждую непокрытую ветку
- [ ] Каждый пакет фиксируется отдельным коммитом:
  - [ ] `test(sync): bring cmd/sync to 100% coverage`
  - [ ] `test(get): bring cmd/get to 100% coverage`
  - [ ] `test(run): bring cmd/run to 100% coverage`
  - [ ] `test(cmd): bring root cmd to 100% coverage`
  - [ ] `test(result): bring cmd/result to 100% coverage`
  - [ ] `test(ui): bring internal/ui to 100% coverage`

---

## Phase 4 — Устранение lint-замечаний (golangci-lint v2)

### Контекст

При миграции CI на golangci-lint v2.11.4 (Go 1.25-совместимый) выявлено ~290 pre-existing замечаний. Линтер v1.64.8 (Go 1.24) никогда не мог запуститься с Go 1.25, поэтому эти проблемы не были видны ранее. Lint step в CI работает с `continue-on-error: true` до завершения этой фазы.

### 4.1. Статистика замечаний (baseline 2026-04-06)

| Линтер | Кол-во | Приоритет |
|--------|--------|-----------|
| gocritic | ~90 | HIGH — стиль и performance hints |
| errcheck | ~52 | HIGH — непроверенные ошибки |
| staticcheck | ~47 | HIGH — потенциальные баги |
| misspell | ~45 | LOW — опечатки в комментариях |
| gocyclo | ~16 | MEDIUM — сложность функций |
| unused | ~15 | MEDIUM — неиспользуемый код |
| nolintlint | ~9 | LOW — невалидные nolint-директивы |
| ineffassign | ~2 | HIGH — присваивания без использования |

### 4.2. План исправления

- [x] **Batch 1 — errcheck + ineffassign** (HIGH): добавить обработку/игнорирование возвращаемых ошибок
- [x] **Batch 2 — staticcheck** (HIGH): исправить потенциальные баги и deprecated вызовы
- [x] **Batch 3 — gocritic** (HIGH): рефакторинг по style/performance рекомендациям
- [x] **Batch 4 — unused** (MEDIUM): удалить неиспользуемый код
- [x] **Batch 5 — gocyclo** (MEDIUM): упростить сложные функции или обосновать `//nolint`
- [x] **Batch 6 — misspell + nolintlint** (LOW): typo-фиксы, очистка директив
- [x] Финальный прогон: `golangci-lint run` EXIT 0
- [x] Убрать `continue-on-error: true` из CI workflow (lint step)
- [x] Commit: `fix(lint): resolve all golangci-lint v2 findings`

---

## Phase 5 — Валидация

- [x] Полный прогон: `go test ./...` — 42/42 PASS
- [x] Полный прогон: `go test -race ./...` — 41/41 PASS, 0 races (excl. concurrency — CI skip)
- [x] Полный прогон: `go test -cover ./...` — 35/42 @ 100%, 7 @ 97.4–99.8%
- [x] `go vet ./...` — PASS
- [x] `go build ./...` — PASS
- [x] `golangci-lint run` — EXIT 0, ноль замечаний
- [x] Coverage: 42/42 PASS (min 97.4% cmd/sync, avg >99.5%)

---

## Phase 6 — Полный повторный аудит

**ОБЯЗАТЕЛЬНЫЙ** финальный аудит по шаблону `.github/prompts/full-project-audit.prompt.md`:

- [x] Phase 0: Scope & skip list
- [x] Phase 1: Architecture & layers — CONDITIONAL PASS (0C/0H/3M/2L)
- [x] Phase 2: TestRail API compliance — PASS (135 endpoints, 98%)
- [x] Phase 3: Code quality & patterns — CONDITIONAL PASS (0C/0H/4M/4L)
- [x] Phase 4: Tests & race detector — PASS (42/42 ≥97.4%, 0 races)
- [x] Phase 5: Documentation — CONDITIONAL PASS (0C/2H/3M/3L)
- [x] Phase 6: CI/Build/Security gates — PASS (6 stdlib vulns, 0 dep)
- [x] Phase 7: Consolidation report — CONDITIONAL PASS

Вердикт аудита: **CONDITIONAL PASS** — 2 HIGH в README требуют фикса перед PR.

---

## Phase 6.5 — Remediation: DRY CRUD + Compare Decouple

**Цель:** Закрыть все finding'ы аудита Phase 6 (B-1, B-5) + i18n-унификация.

### 6.5.1. i18n — Унификация языка в кодовой базе

- [x] Pass 1: Russian→English в тестовых описаниях и комментариях (8 файлов)
- [x] Pass 2: Полный перевод всех 1738 кириллических строк в 170+ Go-файлах
- [x] Верификация: `grep -rn '[а-яА-ЯёЁ]' --include='*.go'` → 0 совпадений
- [x] Build + Tests + Lint — PASS
- [x] Commit: `i18n: translate all Russian text to English in Go source files`

### 6.5.2. B-5 — Compare Resource Registry ✅

**Проблема:** `cmd/compare/all.go` жёстко вызывает 12 compare-функций последовательно. Добавление новой сущности → правка all.go.

**Решение:** Registry pattern — `resourceEntry` struct + `resourceRegistry` массив.

- [x] `resourceEntry` struct (displayName, key, label, accessor, factory)
- [x] `resourceRegistry` — единый массив 12 записей
- [x] `newSimpleResourceEntry()` factory для 9 простых ресурсов
- [x] `init()` автозаполняет `compareAllStages` из registry
- [x] `runCompareAllResources()` строит steps из registry
- [x] Верификация: `go test ./cmd/compare/... -count=1` — PASS (3.5s)
- [x] all.go: 726→~680 LOC, тройное дублирование устранено

### 6.5.3. B-1 — DRY CRUD через Generic Executor ✅

**Проблема:** `cmd/add.go` (1200 LOC), `cmd/update.go` (850 LOC) — 70% boilerplate.

**Решение:** `internal/crud/executor.go` — Go generics `Execute[Req, Resp]` + `DryRun[Req]`.
Для каждой сущности — единственная `buildXxxReq(cmd, validate bool)` функция, shared между execute и dry-run.

| Шаг | Описание | Статус |
| --- | --- | --- |
| 1 | `internal/crud/executor.go` — `Execute[Req, Resp]` + `DryRun[Req]` (79 LOC) | ✅ Done |
| 2 | `internal/crud/executor_test.go` — 7 тестов (162 LOC) | ✅ Done |
| 3 | cmd/add.go: 7 buildReq + 7 slim addXxx + 7 slim dryRunXxx | ✅ Done |
| 4 | cmd/update.go: 6 buildReq + 6 slim updateXxx + 6 slim dryRunXxx | ✅ Done |
| 5 | cmd/delete.go — уже лаконичен (300 LOC), рефакторинг не требуется | ✅ Skip |
| 6 | Верификация: 260 тестов PASS, 0 lint issues, vet OK | ✅ Done |

**Результат:**
- add.go: 1200→1057 LOC (-143, -12%)
- update.go: 850→697 LOC (-153, -18%)
- Net: -217 LOC production code + 241 LOC (executor+tests)

---

## Phase 7 — Closure

- [x] Обновить docs/reports (quality-metrics, audit-report)
- [x] Финализировать CHANGELOG
- [x] Синхронизировать command guides с CLI (attachments, sync, bdds — EN+RU)
- [x] Финальный комплексный аудит (automated gates + static scan + deep audit + security scan)
- [x] Hardening: `formatAPIError()` — `io.LimitReader` на response body
- [x] Зафиксировать результаты Phase 7 в отчётах
- [x] Создать PR: stage-13.5 → release-3.0.0 (PR #17, merged 2026-04-09)
- [x] Создать PR: release-3.0.0 → main
- [ ] Tag v3.0.0 после мерджа в main

---

## Ожидаемые метрики после Stage 13.5

| Метрика | Текущее (13.0) | Целевое (13.5) |
|---------|----------------|----------------|
| Test coverage total | 96.8-100% | **100.0%** |
| Пакеты @ 100% | 36/42 | **42/42** |
| API endpoints в api_paths.go | 128 | **~143** |
| CLI-доступные операции | ~90.5% | **~98%** |
| Data races | 0 | **0** |
| go vet warnings | 0 | **0** |
| golangci-lint findings | ~290 (non-blocking) | **0** |
| Audit verdict | CONDITIONAL PASS | **UNCONDITIONAL PASS** |

---

## Режим работы

- **stepwise**: один шаг → отчёт → подтверждение
- Каждая фаза фиксируется отдельными коммитами
- Docs shadow-sync обязателен для каждого change-cluster
- Checkpoint после каждого завершённого шага

---

← [Stage 13](index.md) · [Отчёты](../index.md) · [Документация](../../index.md)
