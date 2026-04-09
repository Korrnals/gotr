# Stage 13 - Quality Metrics (Baseline)

Language: Русский | [English](../../../en/reports/stage13/quality-metrics.md)

## Навигация

- [Документация](../../index.md)
  - [Гайды](../../guides/index.md)
    - [Установка](../../guides/installation.md)
    - [Конфигурация](../../guides/configuration.md)
    - [Интерактивный режим](../../guides/interactive-mode.md)
    - [Прогресс](../../guides/progress.md)
    - [Каталог команд](../../guides/commands/index.md)
      - [Общие](../../guides/commands/index.md#общие)
      - [CRUD операции](../../guides/commands/index.md#crud-операции)
      - [Основные ресурсы](../../guides/commands/index.md#основные-ресурсы)
      - [Специальные ресурсы](../../guides/commands/index.md#специальные-ресурсы)
  - [Архитектура](../../architecture/index.md)
  - [Эксплуатация](../../operations/index.md)
  - [Отчёты](../index.md)
    - [Stage 13](index.md)
    - [История](../history/index.md)
      - [Final Audit](final-coverage-audit-2026-04-05.md)
      - [Release Summary](release-summary.md)
      - [Audit Report](audit-report.md)
      - [Quality Metrics](quality-metrics.md)
      - [API Compliance](api-compliance-matrix.md)
      - [CLI Contract](cli-contract-matrix.md)
      - [Architecture Conformance](architecture-conformance.md)
      - [Reliability Audit](reliability-audit.md)
      - [Coverage Matrix](test-coverage-matrix.md)
      - [Checklist](coverage-checklist.md)
      - [Layer 2 Wave](layer2-coverage-wave.md)
      - [TODO](todo.md)
- [Главная](../../../../README_ru.md)

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

## Stage 13.3 delta snapshot (2026-04-06)

- Security config hardening: `internal/models/config` сохраняет конфиг с правами `0600`.
- `cmd/config view`: чувствительные ключи (`api_key`, `password`, `token`, `authorization`) редактируются как `"***"`.
- CI hardening:
- `Makefile`: добавлен target `lint`, `verify` обновлен до `test+vet+lint+build+race+vuln`.
- `.github/workflows/ci-stage13.yml`: добавлен шаг `golangci-lint` (`v1.64.8`), `govulncheck` зафиксирован на `v1.1.4`.
- Targeted tests (config/test interactive helpers): PASS.
- request-layer hardening: `internal/client/request.go` (guard for nil response/body + explicit non-OK read-error wrapping).
- Targeted tests (request): PASS (`internal/client/request_test.go`).
- compare seam micro-slice: `cmd/compare/retry_failed_pages.go` использует command-level DTO (`failedPageRecord`) для report JSON и конвертацию в runtime type.
- Targeted tests (compare retry): PASS (`cmd/compare/retry_failed_pages_test.go`).

## Stage 13.3 closure snapshot (2026-04-06)

- R8/R10/R13/R14: закрыты.
- Build gate: PASS (`go build ./...`).
- Test gate: PASS (`go test ./...`).
- Race gate: PASS (`go test -race ./...`).
- Vet gate: PASS (`go vet ./...`).
- Lint gate: PASS (`golangci-lint run --config .golangci.yml --timeout 5m`).
- Vulnerability gate: PASS (`govulncheck ./...`, tool installed: `govulncheck@v1.1.4`).
- Решение по scope closure: coverage 100% target вынесен в post-stage backlog.

## Coverage baseline snapshot (Phase 3.1 COV-1)

- Full coverage command: `go test -vet=off -coverpkg=./... -coverprofile=/tmp/stage13_full.cover ./...`
- Full coverage total: **67.4%** (statements)
- Gap to 100% target: **32.6%**
- Matrix artifact: docs/reports/stage13/test-coverage-matrix.md
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

## Coverage delta snapshot (Phase 3.1 COV-3 partial 16)

- Added tests:
- `internal/client/concurrent_http_test.go` (added HTTP coverage for `GetCasesParallel`, `GetSuitesParallel`, `GetCasesForSuitesParallel`)
- Recomputed full coverage total: **76.8%** (statements)
- Delta vs COV-3 partial 15: **+0.4%**
- Delta vs COV-1 baseline: **+9.4%**
- COV-3 status: **IN PROGRESS**.

## Coverage delta snapshot (Phase 3.1 COV-3 partial 17)

- Added tests:
- `internal/client/client_options_test.go` (covered `WithSkipTlsVerify`, `WithTimeout`, and `NewClient` option/default paths)
- `internal/service/migration/mapping_unit_test.go` (covered `AddPair`, `SortPairs`, `Save`, `LoadSharedStepMapping` flows)
- Recomputed full coverage total: **76.9%** (statements)
- Delta vs COV-3 partial 16: **+0.1%**
- Delta vs COV-1 baseline: **+9.5%**
- COV-3 status: **IN PROGRESS**.

## Coverage delta snapshot (Phase 3.1 COV-3 partial 18)

- Added tests:
- `cmd/config_test.go` (coverage for config init/path/view/edit no-config flows)
- `cmd/completion_test.go`, `cmd/list_test.go` (expanded command branches)
- `main_test.go` with testable `runMain` path in `main.go`
- `internal/selftest/checks_test.go` (helpers + safe checker branches)
- Recomputed full coverage total: **77.4%** (statements)
- Delta vs COV-3 partial 17: **+0.5%**
- Delta vs COV-1 baseline: **+10.0%**
- COV-3 status: **IN PROGRESS**.

## Coverage resync snapshot (2026-03-28 control pass)

- Цель: синхронизировать «источник истины» по total после серии инкрементальных шагов.
- Протокол:
- `go test -vet=off -coverpkg=./... -coverprofile=/tmp/stage13_full.cover ./... >/tmp/stage13_full.log 2>&1`
- `go tool cover -func=/tmp/stage13_full.cover | tail -n 5`
- Актуальный total: **85.1%** (statements).
- Примечание: в истории были нестабильные прогоны `-coverpkg` (долгий teardown в `cmd`), поэтому для фиксации метрики учитываются только успешные полные прогоны по протоколу.

## Coverage delta snapshot (Phase 3.1 COV-3 partial 19)

- Added tests:
- `internal/client/users_test.go` (expanded with non-OK HTTP and decode-error branches for users/priorities/statuses/templates methods)
- Recomputed full coverage total: **86.9%** (statements)
- Delta vs control resync (85.1%): **+1.8%**
- Delta vs COV-1 baseline: **+19.5%**
- COV-3 status: **IN PROGRESS**.

## Coverage delta snapshot (Phase 3.1 COV-3 partial 20)

- Added tests:
- `internal/client/results_test.go` (expanded with non-OK/decode-error branches for add/get result methods)
- Recomputed full coverage total: **86.9%** (statements)
- Delta vs COV-3 partial 19: **+0.0%**
- Delta vs COV-1 baseline: **+19.5%**
- COV-3 status: **IN PROGRESS**.

## Coverage delta snapshot (Phase 3.1 COV-3 partial 21)

- Added tests:
- `internal/client/runs_test.go` (expanded with `GetRuns` success/error and decode-error branches for get/add/update/close)
- Recomputed full coverage total: **87.0%** (statements)
- Delta vs COV-3 partial 20: **+0.1%**
- Delta vs COV-1 baseline: **+19.6%**
- COV-3 status: **IN PROGRESS**.

## Coverage delta snapshot (Phase 3.1 COV-3 partial 22)

- Added tests:
- `internal/client/sharedsteps_test.go` (expanded with request/decode/non-OK branches for shared steps API methods)
- Recomputed full coverage total: **87.0%** (statements)
- Delta vs COV-3 partial 21: **+0.0%**
- Delta vs COV-1 baseline: **+19.6%**
- COV-3 status: **IN PROGRESS**.

## Coverage delta snapshot (Phase 3.1 COV-5 partial 23)

- Added tests:
- `cmd/result/add_test.go` (added unit coverage for `buildAddResultRequest` required/all-fields and extra add-bulk error/dry-run branches)
- Recomputed full coverage total: **87.0%** (statements)
- Delta vs previous snapshot: **+0.0%**
- Delta vs COV-1 baseline: **+19.6%**
- COV-5 status: **IN PROGRESS**.

## Coverage delta snapshot (Phase 3.1 COV-5 partial 24)

- Added tests:
- `cmd/root_test.go` (success/panic coverage for client accessors and `initConfig` invalid-YAML branch)
- Recomputed full coverage total: **87.0%** (statements)
- Delta vs COV-5 partial 23: **+0.0%**
- Delta vs COV-1 baseline: **+19.6%**
- COV-5 status: **IN PROGRESS**.

## Coverage delta snapshot (Phase 3.1 COV-3 partial 25)

- Added tests:
- `internal/client/sections_test.go` (expanded with decode/request/partial-error branches for sections APIs)
- Recomputed full coverage total: **87.1%** (statements)
- Delta vs previous snapshot: **+0.1%**
- Delta vs COV-1 baseline: **+19.7%**
- COV-3 status: **IN PROGRESS**.

## Coverage delta snapshot (Phase 3.1 COV-5 partial 26)

- Added tests/refactor:
- `cmd/resources.go` + `cmd/resources_test.go` (resolved unreachable branches and closed resource utility coverage in package profile)
- Recomputed full coverage total (revalidated twice): **86.3%** (statements)
- Delta vs previous snapshot: **-0.8%** (resync drift)
- Delta vs COV-1 baseline: **+18.9%**
- Notes: dedicated `./cmd` profile confirms `cmd/resources.go` at 100% by functions; full `-coverpkg` baseline fixed at 86.3 for current state.
- COV-5 status: **IN PROGRESS**.

## Coverage audit resync snapshot (2026-03-28, repeated)

- Goal: восстановить точную картину покрытия для матрицы «покрыто / частично / не покрыто».
- Run 1 (global KPI):
- `go test -vet=off -coverpkg=./... -coverprofile=/tmp/stage13_audit.cover.out ./...`
- Result: `total: 86.3%` (statements).
- Run 2 (file-level audit truth):
- `go test -vet=off -coverprofile=/tmp/stage13_pkg.cover.out ./...`
- Result: `total: 85.8%` (statements).
- File buckets from package-local profile:
- fully covered files: **70**
- partially covered files: **157**
- zero-covered files: **2**
- total tracked files: **229**
- Zero-covered files:
- `embedded/jq_embed.go`
- `pkg/testrailapi/api_paths.go`
- Lowest partial hotspots (top-priority backlog):
- `cmd/compare/register.go`
- `cmd/export.go`
- `cmd/root.go`
- `cmd/internal/testhelper/testhelper.go`
- `internal/client/mock.go`
- `internal/output/*`
- `internal/ui/*`
- Methodology note:
- Пер-file bucket-анализ не берется из `-coverpkg` профиля, так как повторная инструментализация в package test binaries искажает агрегированные проценты по файлам. Для audit-классификации используется только package-local профиль.

## Coverage delta snapshot (Phase 3.1 COV-5 partial 27)

- Added tests:
- `embedded/jq_embed_test.go` (success/error branches for embedded jq execution)
- `pkg/testrailapi/api_paths_test.go` (API resource initialization and endpoint aggregation validation)
- Recomputed package-local total: **86.4%** (statements)
- File buckets (package-local): **71 full / 158 partial / 0 zero**
- Zero-covered files: **none** (previously 2)
- Recomputed global KPI (`-coverpkg`): **86.3%** (statements, unchanged)
- Delta vs previous package-local snapshot (85.8%): **+0.6%**
- COV-5 status: **IN PROGRESS**.

## Coverage delta snapshot (Phase 3.1 COV-5 partial 28)

- Added tests:
- `cmd/compare/register_test.go` (coverage for `Register`, persistent flags and subcommands)
- `cmd/export_test.go` (expanded `resolveExportInputs` edge-cases: no-prompter, select/input failures, id prompt)
- `cmd/root_test.go` (expanded `GetClientInterface` fallback/panic and `initConfig` config-not-found path)
- `cmd/internal/testhelper/testhelper_test.go` (full branch coverage for helper functions)
- Recomputed package-local total: **86.8%** (statements)
- File buckets (package-local): **73 full / 156 partial / 0 zero**
- Recomputed global KPI (`-coverpkg`): **86.3%** (statements, unchanged)
- Key per-file deltas:
- `cmd/compare/register.go`: **3.33% -> 100.0%**
- `cmd/export.go`: **41.25% -> 52.50%**
- `cmd/root.go`: **42.42% -> 45.45%**
- `cmd/internal/testhelper/testhelper.go`: **46.67% -> 100.0%**
- Delta vs previous package-local snapshot (86.4%): **+0.4%**
- COV-5 status: **IN PROGRESS**.

## Coverage delta snapshot (Phase 3.1 COV-5 partial 29)

- Added tests:
- `internal/output/dryrun_test.go` (added coverage for `PrintOperation` and `PrintSimple`, including marshal-error body path)
- Recomputed package-local total: **87.0%** (statements)
- File buckets (package-local): **74 full / 155 partial / 0 zero**
- Recomputed global KPI (`-coverpkg`): **86.3%** (latest stable control pass)
- Key per-file delta:
- `internal/output/dryrun.go`: **59.09% -> 100.0%**
- Delta vs previous package-local snapshot (86.8%): **+0.2%**
- COV-5 status: **IN PROGRESS**.

## Coverage delta snapshot (Phase 3.1 COV-5 partial 30)

- Added tests (parallel wave):
- `internal/output/save_test.go` (OutputResult, OutputGetResult branches, OutputResultWithFlags, PrintSuccess, SaveToFileWithPath branches)
- `internal/ui/display_test.go` (refreshLoop, render, Display.Finish, Infof/Successf/Warningf quiet-paths)
- `internal/client/extended_test.go` (non-OK/decode error branches for groups/roles/result-fields/datasets/variables/bdd/labels)
- `cmd/export_test.go` (resolveExportInputs tail branches incl. no-prompter ID and trim normalization)
- `cmd/root_test.go` + `cmd/root.go` (initConfig home-dir error hook + edge coverage)
- Recomputed package-local total: **88.2%** (statements)
- File buckets (package-local): **74 full / 155 partial / 0 zero**
- Recomputed global KPI (`-coverpkg`): **86.3%** (statements)
- Key per-file deltas:
- `internal/output/save.go`: **56.12% -> 83.16%**
- `internal/ui/display.go`: **56.74% -> 98.58%**
- `internal/client/extended.go`: **68.62% -> 75.86%**
- `cmd/export.go`: **52.50% -> 53.75%**
- `cmd/root.go`: **45.45% -> 46.97%**
- Delta vs previous package-local snapshot (87.0%): **+1.2%**
- COV-5 status: **IN PROGRESS**.

## Stage 13.5 Closure Snapshot (2026-04-09)

### Финальные метрики

| Метрика | Stage 13.0 baseline | Stage 13.5 final |
|---------|---------------------|------------------|
| Go files total | 411 | 528 |
| Test files | — | 250 |
| Test packages (ok) | 42 | **43/43** |
| Total tests | — | **3615+** |
| go build | PASS | **PASS** |
| go vet | PASS | **PASS** |
| go test -race | PASS | **PASS (0 data races)** |
| golangci-lint findings | ~290 | **0** |
| TODO/FIXME/HACK markers | — | **0** |
| Audit verdict | CONDITIONAL PASS | **UNCONDITIONAL PASS** |

### Stage 13.5 diff (vs release-3.0.0)

- 28 коммитов
- 336 файлов затронуто
- +6 215 / −4 641 строк

### Выполненные фазы

| Фаза | Статус | Ключевые результаты |
|------|--------|---------------------|
| Phase 1 — api_paths.go | DONE | +14 endpoints, полный реестр |
| Phase 2 — CLI-обёртки | DONE | `attachments list --for-project` |
| Phase 3 — Coverage 100% | DONE | cmd/sync,get,run,result,ui @ 100% |
| Phase 4 — Lint fixes | DONE | 290 → 0 findings (errcheck, staticcheck, gocritic, gocyclo) |
| Phase 5 — Validation | DONE | 42/42 PASS, race: 0, min coverage 97.4% |
| Phase 6 — Audit | DONE | 7 audit rounds (audit-1 → audit-7), все findings закрыты |
| Phase 7 — Closure | DONE | Документация, CHANGELOG, PR |

### Аудит-история (Phase 6)

| Раунд | Коммит | Scope |
|-------|--------|-------|
| audit-1 | `41cf03b` | F-2..F-7: context, bounded parallelism, decouple, errors |
| audit-1b | `891034d` | B-2,B-3,B-4,B-7: ClientInterface, service decouple, unlambda |
| audit-2 | `de6a195` | E-1..E-4 + I-1..I-3: final hardening pass |
| audit-3 | `19cab1c` | json.Marshal errors across codebase (45+ fixes, 17 files) |
| audit-4 | `bab05c3` | safe type assertions + getwd error handling |
| audit-5 | `1832ca5` | io.LimitReader on all io.ReadAll, fd leak, os.Remove |
| audit-6 | `2554803` | panic→os.Exit, save-filtered feature, 3 new tests |
| audit-7 | `c3733ae` | stdin reading for bdds add, 2 new tests |
| phase-7a | `cc1cc3e` | docs sync: attachments/sync/bdds command guides (EN+RU) |
| phase-7b | `9abccc5` | harden formatAPIError io.ReadAll, fix markdown report table |

### Дополнительные работы

- **i18n**: полный перевод Russian→English в Go source (2 прохода, 1738+ строк)
- **refactor**: generic CRUD executor, compare resource registry
- **feat**: `save-filtered` для sync shared-steps, stdin для `bdds add`

---

← [Stage 13](index.md) · [Отчёты](../index.md) · [Документация](../../index.md)
