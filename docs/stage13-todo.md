# Stage 13 - Detailed TODO Plan

Дата: 2026-03-27
Ветка: stage-13.0-final-refactoring
Режим: autonomous
Модель работы: multi-specialist orchestration + поэтапные коммиты после каждого завершенного шага

## Правила выполнения этапа

- Каждый завершенный шаг фиксируется отдельным коммитом.
- После каждого завершенного шага сохраняется checkpoint (/project-checkpoint-now).
- Все изменения проходят build/test/vet; для соответствующих шагов также race/lint.
- Для шагов с изменением поведения обязателен docs shadow-sync.

## Phase 1 - Baseline Capture

- [x] Создать stage-ветку от release-3.0.0.
- [x] Зафиксировать baseline метрики: go build, go test, go vet, инвентарь файлов.
- [x] Создать стартовые артефакты:
- [x] docs/stage13-quality-metrics.md
- [x] docs/stage13-audit-report.md
- [x] Закоммитить шаг Phase 1.

## Phase 2 - Core Audits

- [x] Architecture Conformance (architect + backend-engineer):
- [x] Построить dependency map по пакетам cmd/internal/pkg.
- [x] Зафиксировать layering violations и coupling hotspots.
- [x] Сформировать file risk map (high/medium/low).
- [x] Обновить audit-report (architecture findings + remediation).
- [ ] Commit шага Architecture Conformance.

- [x] CLI Contract Consistency (qa-engineer + backend-engineer):
- [x] Собрать quiet behavior matrix по всем command trees.
- [x] Проверить non-interactive/interactive edge-cases.
- [x] Составить таблицу normalisation по flags (persistent/local/duplicates).
- [x] Обновить audit-report и добавить docs/stage13-cli-contract-matrix.md.
- [ ] Commit шага CLI Contract Audit.

- [ ] API Compliance (architect + qa-engineer + backend-engineer):
- [ ] Сформировать endpoint-by-endpoint compliance matrix.
- [ ] Проверить пагинацию, retries, timeout, error mapping.
- [ ] Зафиксировать Partial/Unsupported с обоснованием.
- [ ] Добавить docs/stage13-api-compliance-matrix.md.
- [ ] Commit шага API Compliance Audit.

- [ ] Reliability/Concurrency (backend-engineer + qa-engineer):
- [ ] Прогнать go test -race ./... и зафиксировать результат.
- [ ] Проверить cancel/timeout propagation и bounded retry behavior.
- [ ] Выделить performance hot paths для compare/sync.
- [ ] Обновить quality-metrics и audit-report.
- [ ] Commit шага Reliability Audit.

- [ ] Security & Supply Chain (devops-engineer + backend-engineer):
- [ ] Провести dependency audit.
- [ ] Проверить утечки чувствительных данных в logs/errors.
- [ ] Проверить input/filepath validation критичных путей.
- [ ] Обновить audit-report (security findings).
- [ ] Commit шага Security Audit.

- [ ] CI/CD Hardening (devops-engineer + release-manager):
- [ ] Формализовать обязательные quality gates.
- [ ] Синхронизировать Makefile verify path и CI path.
- [ ] Проверить reproducibility для release-артефактов.
- [ ] Обновить docs/release-workflow.md при необходимости.
- [ ] Commit шага CI/CD Hardening.

## Добавленные дельты плана (2026-03-27)

- [x] Добавить подпоток "Interactive helper consolidation" как отдельную remediation-зону Stage 13.
- [x] Добавить подпоток "Compare runtime seam hardening" для снижения прямой связки cmd/compare и internal/concurrency.
- [ ] Зафиксировать реализацию этих двух подпотоков в Phase 3 с отдельными коммитами.

## Новые remediation задачи из CLI Contract Audit

- [ ] R1 (HIGH): Убрать локальные quiet-flag декларации из cmd/run/run.go, cmd/test/list.go, cmd/test/get.go, cmd/result/result.go.
- [ ] R2 (MEDIUM): Добавить `interactive.IsNonInteractive(ctx)` helper и мигрировать type assertion pattern.
- [ ] R3 (LOW): Убрать `isQuiet()` wrapper из cmd/sync/sync_helpers.go.
- [ ] R4 (MEDIUM): Аудит прямых fmt.Fprintf/os.Stdout без quiet-guard в 15 command groups.
- [ ] R5 (MEDIUM): Fix ReadJSONResponse body leak — добавить `defer resp.Body.Close()` перед non-200 ветку (internal/client/request.go:54).
- [ ] R6 (LOW): Add GroupsAPI, RolesAPI, DatasetsAPI, VariablesAPI, BDDsAPI, LabelsAPI интерфейсы в interfaces.go.

## Phase 3 - Refactoring Implementation

- [ ] Выполнить remediation по high severity findings (пакетами).
- [ ] Для каждого change-cluster добавить regression tests.
- [ ] Для каждого change-cluster выполнить docs shadow-sync.
- [ ] После каждого change-cluster делать отдельный commit.

## Phase 4 - Validation

- [ ] Полный прогон: go test ./... .
- [ ] Полный прогон: go test -race ./... .
- [ ] Полный прогон: go vet ./... .
- [ ] Линтерный прогон по согласованному gate.
- [ ] Финальное обновление metrics и evidence артефактов.
- [ ] Commit шага Final Validation.

## Phase 5 - Closure

- [ ] Финализировать docs/stage13-audit-report.md.
- [ ] Финализировать docs/stage13-quality-metrics.md.
- [ ] Финализировать docs/stage13-api-compliance-matrix.md.
- [ ] Финализировать docs/stage13-cli-contract-matrix.md.
- [ ] Подготовить release readiness summary.
- [ ] Commit шага Stage 13 Closure.

## Gate Checklist (Blockers)

- [ ] Build gate: go build ./... .
- [ ] Test gate: go test ./... .
- [ ] Race gate: go test -race ./... .
- [ ] Vet gate: go vet ./... .
- [ ] Contract gate: CLI + API matrices completed.
- [ ] Docs gate: все внешние изменения отражены в документации.
