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
