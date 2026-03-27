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
