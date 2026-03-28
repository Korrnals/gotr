# Stage 13 - Quality Metrics (Baseline)

Дата baseline: 2026-03-27
Ветка: stage-13.0-final-refactoring
Коммит baseline: 537ddad

## Инвентарь кода

- Go files total: 411
- cmd files: 289
- internal files: 119
- pkg files: 3
- docs files: 18
- embedded files: 3

## Baseline gates (Phase 1)

- go build ./...: PASS
- go test ./...: PASS (262 passed, 0 failed)
- go vet ./...: PASS
- go test -race ./...: NOT RUN (планируется в отдельном шаге reliability-аудита)
- golangci-lint: NOT RUN (инструмент и gate будут верифицированы в workstream CI/CD + code quality)

## Baseline notes

- Стартовые quality gates на момент входа в Stage 13 зеленые для build/test/vet.
- Полный Stage 13 quality-gate пакет будет считаться только после race/lint/contract matrices.
- Метрики этого файла обновляются после каждого закрытого workstream шага Stage 13.

## Step 2 snapshot - Architecture Conformance

- `internal/pkg -> cmd` violations: 0
- `pkg -> internal` violations: 0
- `cmd/compare -> internal/concurrency` direct coupling files: 4
- `cmd/*/interactive_helpers.go` duplication points: 17

Interpretation:

- Базовые архитектурные границы соблюдены.
- Основные риски шага 2 лежат в зоне поддерживаемости и связности, а не в запретных импортах.

## Step 3 snapshot - CLI Contract Audit

- command groups analyzed: 22
- local quiet-flag overrides (non-test): 7 declarations / 4 locations (HIGH)
- groups with quiet=n/a (no quiet check, rely on global seam): 15 (MEDIUM — need R4 audit)
- non-interactive type assertion occurrences: ~15 files (fragile pattern, MEDIUM)
- isQuiet() wrapper duplicates: 1 function in cmd/sync/sync_helpers.go (LOW)
- CLI contract findings: F1(HIGH), F2(MEDIUM), F3(LOW), F4(MEDIUM)
- Remediation added to Phase 3: R1(HIGH), R2(MEDIUM), R3(LOW), R4(MEDIUM)

## Step 4 snapshot - API Compliance Audit

- Transport checks: 7/7 PASS
- URL construction checks: 5/5 PASS
- Response handling: 5/6 PASS, 1 FAIL (body leak F5)
- Interface coverage: 4/4 PASS + 1 NOTE (F6)
- Pagination: 3/3 PASS
- Security: 3/3 PASS
- MEDIUM findings: 1 (F5 body leak)
- LOW findings: 1 (F6 interface gaps)
- Remediation added: R5(MEDIUM), R6(LOW)

## Step 5 snapshot - Reliability & Concurrency Audit

- go test -race: PARTIAL PASS (targeted packages with CGO_ENABLED=1)
- Race-tested packages: internal/concurrency PASS, internal/concurrent PASS, internal/client targeted PASS, internal/service PASS, internal/interactive PASS, cmd/compare PASS after fix
- DATA RACE detected and fixed: cmd/compare/fetchers_test.go (captured append protected by mutex), commit 9358ac8
- Static mutex analysis: 10/10 concurrent data structures protected
- Loop variable capture: PASS (all goroutine launches)
- errgroup usage: PASS (g.SetLimit + ctx propagation)
- Buffered channels: PASS
- Global mutable state findings: 1 WARN (F11 PriorityThresholds)
- Remediation added: R7(INFO), R8(LOW), R9(DONE)

## Step 6 snapshot - Security & Supply Chain Audit

- go mod verify: PASS (`all modules verified`)
- dependency update candidates: present (57 module lines in `go list -m -u all` output)
- govulncheck: NOT AVAILABLE in environment (tool not installed)
- secret-pattern scan: no exposed credentials detected
- risky exec usage: limited to editor/selftest contexts
- Security findings: F13(LOW), F14(LOW)
- Remediation added: R10(LOW), R11(MEDIUM)

## Step 7 snapshot - CI/CD Hardening Audit

- `.github/workflows`: absent (no CI workflow files)
- Makefile verify target: absent (no unified quality gate command)
- build/release separation: partial (build currently triggers `sync-tag` side-effect)
- reproducibility checks: absent (no checksum verification flow)
- CI/CD findings: F15(HIGH), F16(MEDIUM), F17(MEDIUM)
- Remediation added: R12(HIGH), R13(MEDIUM), R14(MEDIUM)

## Coverage baseline snapshot (Phase 3.1 COV-1)

- Full coverage command: `go test -vet=off -coverpkg=./... -coverprofile=/tmp/stage13_full.cover ./...`
- Full coverage total: **67.4%** (statements)
- Gap to 100% target: **32.6%**
- Matrix artifact: docs/stage13-test-coverage-matrix.md
- Dedicated workstream added: COV-1..COV-6

## Coverage delta snapshot (Phase 3.1 COV-2 complete)

- Added tests:
- `internal/paths/paths_test.go`
- `internal/models/config/config_test.go`
- `internal/selftest/types_test.go`
- `internal/log/logger_test.go`
- Recomputed full coverage total: **68.7%** (statements)
- Delta vs COV-1 baseline: **+1.3%**
- COV-2 status: **DONE**.

## Coverage delta snapshot (Phase 3.1 COV-3 partial)

- Added tests:
- `internal/client/projects_test.go`
- `internal/client/accessor_test.go`
- `internal/service/test_test.go`
- Recomputed full coverage total: **69.2%** (statements)
- Delta vs COV-2 complete: **+0.5%**
- Delta vs COV-1 baseline: **+1.8%**
- COV-3 status: **IN PROGRESS**.

## Coverage delta snapshot (Phase 3.1 COV-3 partial 2)

- Added tests:
- `internal/service/migration/export_loader_log_test.go`
- Recomputed full coverage total: **69.9%** (statements)
- Delta vs COV-3 partial: **+0.7%**
- Delta vs COV-1 baseline: **+2.5%**
- COV-3 status: **IN PROGRESS**.

## Coverage delta snapshot (Phase 3.1 COV-3 partial 3)

- Added tests:
- `internal/client/request_test.go` (expanded with ReadResponse/PrintResponseFromData/SaveResponseToFile coverage)
- Recomputed full coverage total: **70.2%** (statements)
- Delta vs COV-3 partial 2: **+0.3%**
- Delta vs COV-1 baseline: **+2.8%**
- COV-3 status: **IN PROGRESS**.

## Coverage delta snapshot (Phase 3.1 COV-3 partial 4)

- Added tests:
- `internal/service/result_test.go` (expanded with constructors/get/add/parse method coverage)
- `internal/client/reports_test.go` (expanded with HTTP tests for run/cross reports)
- Recomputed full coverage total: **70.4%** (statements)
- Delta vs COV-3 partial 3: **+0.2%**
- Delta vs COV-1 baseline: **+3.0%**
- COV-3 status: **IN PROGRESS**.

## Coverage delta snapshot (Phase 3.1 COV-3 partial 5)

- Added tests:
- `internal/client/plans_test.go` (expanded with HTTP tests for update/close/entry operations)
- Recomputed full coverage total: **70.8%** (statements)
- Delta vs COV-3 partial 4: **+0.4%**
- Delta vs COV-1 baseline: **+3.4%**
- COV-3 status: **IN PROGRESS**.

## Coverage delta snapshot (Phase 3.1 COV-3 partial 6)

- Added tests:
- `internal/client/configs_test.go` (expanded with HTTP tests for add/update/delete config group and config)
- Recomputed full coverage total: **71.2%** (statements)
- Delta vs COV-3 partial 5: **+0.4%**
- Delta vs COV-1 baseline: **+3.8%**
- COV-3 status: **IN PROGRESS**.

## Coverage delta snapshot (Phase 3.1 COV-3 partial 7)

- Added tests:
- `internal/client/extended_test.go` (expanded with HTTP tests across extended API methods)
- Recomputed full coverage total: **72.7%** (statements)
- Delta vs COV-3 partial 6: **+1.5%**
- Delta vs COV-1 baseline: **+5.3%**
- COV-3 status: **IN PROGRESS**.

## Coverage delta snapshot (Phase 3.1 COV-3 partial 8)

- Added tests:
- `internal/client/attachments_test.go` (expanded with HTTP tests for read/delete endpoints and upload wrappers)
- Recomputed full coverage total: **73.4%** (statements)
- Delta vs COV-3 partial 7: **+0.7%**
- Delta vs COV-1 baseline: **+6.0%**
- COV-3 status: **IN PROGRESS**.

## Coverage delta snapshot (Phase 3.1 COV-3 partial 9)

- Added tests:
- `internal/ui/table_test.go`
- `internal/ui/runtime_display_test.go`
- `internal/debug/print_test.go`
- Recomputed full coverage total: **73.7%** (statements)
- Delta vs COV-3 partial 8: **+0.3%**
- Delta vs COV-1 baseline: **+6.3%**
- COV-3 status: **IN PROGRESS**.

## Coverage delta snapshot (Phase 3.1 COV-3 partial 10)

- Added tests:
- `internal/client/cases_test.go` (expanded with decode/get/page/history/bulk/meta coverage and casesEqualByField assertions)
- Recomputed full coverage total: **74.7%** (statements)
- Delta vs COV-3 partial 9: **+1.0%**
- Delta vs COV-1 baseline: **+7.3%**
- COV-3 status: **IN PROGRESS**.

## Coverage delta snapshot (Phase 3.1 COV-3 partial 11)

- Added tests:
- `internal/client/cases_test.go` (expanded with coverage for `DiffCasesData`, `GetCasesParallelCtx`, `casesFetcher.FetchPageCtx`)
- Recomputed full coverage total: **75.2%** (statements)
- Delta vs COV-3 partial 10: **+0.5%**
- Delta vs COV-1 baseline: **+7.8%**
- COV-3 status: **IN PROGRESS**.

## Coverage delta snapshot (Phase 3.1 COV-3 partial 12)

- Added tests:
- `internal/client/suites_test.go` (added GET coverage for `GetSuites` and `GetSuite`)
- `internal/client/sharedsteps_test.go` (added GET coverage for `GetSharedSteps`, `GetSharedStep`, `GetSharedStepHistory`)
- Recomputed full coverage total: **75.5%** (statements)
- Delta vs COV-3 partial 11: **+0.3%**
- Delta vs COV-1 baseline: **+8.1%**
- COV-3 status: **IN PROGRESS**.

## Coverage delta snapshot (Phase 3.1 COV-3 partial 13)

- Added tests:
- `internal/client/users_test.go` (added HTTP coverage for `GetUsersByProject`, `GetUser`, `GetUserByEmail`, `AddUser`, `UpdateUser`, `GetStatuses`, `GetTemplates`)
- Recomputed full coverage total: **76.0%** (statements)
- Delta vs COV-3 partial 12: **+0.5%**
- Delta vs COV-1 baseline: **+8.6%**
- COV-3 status: **IN PROGRESS**.

## Coverage delta snapshot (Phase 3.1 COV-3 partial 14)

- Added tests:
- `internal/client/results_test.go` (added coverage for `GetResults`, `GetResultsForRun`, `GetResultsForCase`)
- `internal/client/runs_test.go` (added coverage for `GetRun`)
- `internal/client/sections_test.go` (added coverage for `GetSection`)
- `internal/client/tests_test.go` (added coverage for `GetTestsByStatus`, `GetTestsAssignedTo`)
- Recomputed full coverage total: **76.3%** (statements)
- Delta vs COV-3 partial 13: **+0.3%**
- Delta vs COV-1 baseline: **+8.9%**
- COV-3 status: **IN PROGRESS**.

## Coverage delta snapshot (Phase 3.1 COV-3 partial 15)

- Added tests:
- `internal/service/run_test.go` (expanded coverage for `Get`, `GetByProject`, `Create`, `Update`, `Close`, `Delete` branches)
- `internal/service/result_test.go` (expanded validation/error branches for `GetForRun`, `GetRunsForProject`, `AddForCase`, `AddResults`, `AddResultsForCases`)
- Recomputed full coverage total: **76.4%** (statements)
- Delta vs COV-3 partial 14: **+0.1%**
- Delta vs COV-1 baseline: **+9.0%**
- COV-3 status: **IN PROGRESS**.
