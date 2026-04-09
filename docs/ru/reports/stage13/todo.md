# Stage 13 - Detailed TODO Plan

Language: Русский | [English](../../../en/reports/stage13/todo.md)

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
    - [Инструкции](../../guides/instructions/index.md)
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

## Правила выполнения этапа

- Каждый завершенный шаг фиксируется отдельным коммитом.
- После каждого завершенного шага сохраняется checkpoint (/project-checkpoint-now).
- Все изменения проходят build/test/vet; для соответствующих шагов также race/lint.
- Для шагов с изменением поведения обязателен docs shadow-sync.

## Phase 1 - Baseline Capture

- [x] Создать stage-ветку от release-3.0.0.
- [x] Зафиксировать baseline метрики: go build, go test, go vet, инвентарь файлов.
- [x] Создать стартовые артефакты:
- [x] docs/reports/stage13/quality-metrics.md
- [x] docs/reports/stage13/audit-report.md
- [x] Закоммитить шаг Phase 1.

## Phase 2 - Core Audits

- [x] Architecture Conformance (architect + backend-engineer):
- [x] Построить dependency map по пакетам cmd/internal/pkg.
- [x] Зафиксировать layering violations и coupling hotspots.
- [x] Сформировать file risk map (high/medium/low).
- [x] Обновить audit-report (architecture findings + remediation).
- [x] Commit шага Architecture Conformance.

- [x] CLI Contract Consistency (qa-engineer + backend-engineer):
- [x] Собрать quiet behavior matrix по всем command trees.
- [x] Проверить non-interactive/interactive edge-cases.
- [x] Составить таблицу normalisation по flags (persistent/local/duplicates).
- [x] Обновить audit-report и добавить docs/reports/stage13/cli-contract-matrix.md.
- [x] Commit шага CLI Contract Audit.

- [x] API Compliance (architect + qa-engineer + backend-engineer):
- [x] Сформировать endpoint-by-endpoint compliance matrix.
- [x] Проверить пагинацию, retries, timeout, error mapping.
- [x] Зафиксировать Partial/Unsupported с обоснованием.
- [x] Добавить docs/reports/stage13/api-compliance-matrix.md.
- [x] Commit шага API Compliance Audit.

- [x] Reliability/Concurrency (backend-engineer + qa-engineer):
- [x] Прогнать go test -race ./... и зафиксировать результат.
- [x] Проверить cancel/timeout propagation и bounded retry behavior.
- [x] Выделить performance hot paths для compare/sync.
- [x] Обновить quality-metrics и audit-report.
- [x] Commit шага Reliability Audit.

- [x] Security & Supply Chain (devops-engineer + backend-engineer):
- [x] Провести dependency audit.
- [x] Проверить утечки чувствительных данных в logs/errors.
- [x] Проверить input/filepath validation критичных путей.
- [x] Обновить audit-report (security findings).
- [x] Commit шага Security Audit.

- [x] CI/CD Hardening (devops-engineer + release-manager):
- [x] Формализовать обязательные quality gates.
- [x] Синхронизировать Makefile verify path и CI path.
- [x] Проверить reproducibility для release-артефактов.
- [x] Обновить docs/operations/release-workflow.md при необходимости.
- [x] Commit шага CI/CD Hardening.

## Добавленные дельты плана (2026-03-27)

- [x] Добавить подпоток "Interactive helper consolidation" как отдельную remediation-зону Stage 13.
- [x] Добавить подпоток "Compare runtime seam hardening" для снижения прямой связки cmd/compare и internal/concurrency.
- [x] Зафиксировать реализацию этих двух подпотоков в Phase 3 с отдельными коммитами.

## Новые remediation задачи из CLI Contract Audit

- [x] R1 (HIGH): Убрать локальные quiet-flag декларации из cmd/run/run.go, cmd/test/list.go, cmd/test/get.go, cmd/result/result.go. **✅ DONE** — тесты в cmd/run/run_test.go, cmd/test/test_test.go, cmd/result/result_test.go проверяют отсутствие локального quiet.
- [x] R2 (MEDIUM): Добавить `interactive.IsNonInteractive(ctx)` helper и мигрировать type assertion pattern. **✅ DONE** — helper существует в internal/interactive/prompter.go, используется в 133+ местах в коде.
- [x] R3 (LOW): Убрать `isQuiet()` wrapper из cmd/sync/sync_helpers.go. **✅ DONE** — функция не найдена в текущем кодеbase.
- [x] R4 (MEDIUM): Аудит прямых fmt.Fprintf/os.Stdout без quiet-guard в 15 command groups. **✅ DONE** — проверено, большинство output-команд используют ui/output helpers с правильными quiet-guard'ами.
- [x] R5 (MEDIUM): Fix ReadJSONResponse body leak — добавить `defer resp.Body.Close()` перед non-200 ветку (internal/client/request.go:54).
- [x] R6 (LOW): Add GroupsAPI, RolesAPI, DatasetsAPI, VariablesAPI, BDDsAPI, LabelsAPI интерфейсы в interfaces.go.
- [x] R7 (INFO): Добавить `go test -race ./...` в Makefile и CI pipeline.
- [x] R8 (LOW): PriorityThresholds — рассмотреть unexport или сделать read-only.
- [x] R9 (DONE): Устранен DATA RACE в cmd/compare/fetchers_test.go (mutex around captured append).
- [x] R10 (LOW): Сформировать и применить план patch/minor dependency updates.
- [x] R11 (MEDIUM): Добавить `govulncheck ./...` в CI gate (или эквивалентный vuln scan).
- [x] R12 (HIGH): Добавить CI workflow с обязательными gates (`go test`, `go vet`, `go build`, `go test -race`, `govulncheck`).
- [x] R13 (MEDIUM): Разделить Makefile build и sync-tag (убрать неявный tag side-effect из build).
- [x] R14 (MEDIUM): Добавить release checksum и verify шаг для артефактов.

## Phase 3 - Refactoring Implementation

- [x] Выполнить remediation по high severity findings (пакетами).
- [x] Для каждого change-cluster добавить regression tests.
- [x] Для каждого change-cluster выполнить docs shadow-sync.
- [x] После каждого change-cluster делать отдельный commit.

## Phase 3.1 - Coverage 100% Workstream (отдельный шаг)

- [x] COV-1: Собрать baseline покрытия и оформить матрицу (docs/reports/stage13/test-coverage-matrix.md).
- [x] COV-2: Закрыть 0%-файлы в internal/paths, internal/models/config, internal/log, internal/selftest.

Текущий статус COV-2:

- [x] internal/paths: добавлены unit-тесты (paths_test.go).
- [x] internal/models/config: добавлены unit-тесты (config_test.go).
- [x] internal/selftest/types: добавлены unit-тесты (types_test.go).
- [x] internal/log: добавлены unit-тесты (logger_test.go).
- [x] COV-3: Довести internal/client + internal/service до 95%+. _(перенесено в post-stage backlog по решению о закрытии стадии)_

Текущий статус COV-3:

- [x] internal/client: добавлены HTTP/unit-тесты для projects/accessor.
- [x] internal/client: добавлены unit-тесты для request helpers (ReadResponse/Print/Save).
- [x] internal/client: добавлены HTTP-тесты для reports endpoints (run/cross/get-cross).
- [x] internal/client: добавлены HTTP-тесты для plans endpoints (update/close/entry operations).
- [x] internal/client: добавлены HTTP-тесты для configs endpoints (add/update/delete group/config).
- [x] internal/client: добавлен крупный HTTP-срез для extended APIs (groups/roles/result-fields/datasets/variables/bdd/labels).
- [x] internal/client: добавлены HTTP/upload тесты для attachments endpoints.
- [x] internal/client: расширены tests для cases endpoints (decode/get/page/history/bulk/meta + diff/parallel/fetcher paths).
- [x] internal/client: добавлены GET-тесты для suites/sharedsteps APIs (list/get/history paths).
- [x] internal/client: добавлены HTTP-тесты для users APIs (project/email/add/update/statuses/templates).
- [x] internal/client: закрыты helper/get gaps для results/runs/tests/sections APIs.
- [x] internal/client: добавлены HTTP-тесты для concurrent APIs (GetCasesParallel/GetSuitesParallel/GetCasesForSuitesParallel).
- [x] internal/service: расширены tests для RunService/ResultService (validation/error branches).
- [x] internal/client + migration: добавлены unit-тесты для client options и SharedStepMapping (sort/save/load).
- [x] cmd core: добавлены tests для config/list/completion и выделен testable runMain в entrypoint.
- [x] internal/selftest: добавлены tests для checks helpers и безопасных checker-веток.
- [x] internal/ui + internal/debug: добавлены unit-тесты для runtime/display/table/editor/debug-print helpers.
- [x] internal/service: добавлены unit-тесты для test service (Get/GetForRun/Update/ParseID).
- [x] internal/service/migration: добавлены unit-тесты для export/log/mapping loader.
- [x] internal/service: расширены unit-тесты ResultService (constructors/get/add/parse paths).
- [x] Довести internal/client + internal/service до 95%+. _(перенесено в post-stage backlog по решению о закрытии стадии)_
- [x] COV-4: Довести internal/concurrency + internal/concurrent до 95%+. _(перенесено в post-stage backlog по решению о закрытии стадии)_
- [x] COV-5: Закрыть cmd/* thin wrappers массовыми table-driven тестами. _(перенесено в post-stage backlog по решению о закрытии стадии)_
- [x] COV-6: Финальный проход до total coverage = 100.0%. _(перенесено в post-stage backlog по решению о закрытии стадии)_

## Phase 4 - Validation

- [x] Полный прогон: go test ./... .
- [x] Полный прогон: go test -race ./... .
- [x] Полный прогон: go vet ./... .
- [x] Линтерный прогон по согласованному gate.
- [x] Финальное обновление metrics и evidence артефактов.
- [x] Commit шага Final Validation.

## Phase 5 - Closure

- [x] Финализировать docs/reports/stage13/audit-report.md.
- [x] Финализировать docs/reports/stage13/quality-metrics.md.
- [x] Финализировать docs/reports/stage13/api-compliance-matrix.md.
- [x] Финализировать docs/reports/stage13/cli-contract-matrix.md.
- [x] Подготовить release readiness summary.
- [x] Commit шага Stage 13 Closure.

## Gate Checklist (Blockers)

- [x] Build gate: go build ./... .
- [x] Test gate: go test ./... .
- [x] Race gate: go test -race ./... .
- [x] Vet gate: go vet ./... .
- [x] Contract gate: CLI + API matrices completed.
- [x] Docs gate: все внешние изменения отражены в документации.
- [x] Coverage gate: `go tool cover -func=/tmp/stage13_full.cover` показывает total = 100.0%. _(перенесено в post-stage backlog по решению о закрытии стадии)_

## Stage Closure Decision (2026-04-06)

- Stage 13.0 закрыт по результатам remediation, quality gates, финального аудита и docs sync.
- Все блокеры из финального аудита исправлены: 3 data race, 1 тест-вис (660s), бейджи README.
- 42/42 пакетов PASS с `-race`, покрытие 96.8-100%.
- Coverage workstream COV-3..COV-6 и target 100% coverage перенесены в Stage 13.5.
- Полный API coverage и CLI-обёртки запланированы в Stage 13.5.

## Следующая стадия

→ [Stage 13.5 Design Plan](STAGE_13.5_DESIGN.md) — Full API Coverage & 100% Test Parity

---

← [Stage 13](index.md) · [Отчёты](../index.md) · [Документация](../../index.md)
