# Stage 13 - Audit Report (Work in Progress)

Дата старта: 2026-03-27
Статус: In Progress
Ветка: stage-13.0-final-refactoring

## Scope

- Полный аудит репозитория в рамках Stage 13.0 (architecture, quality, API contract, CLI contract, reliability, CI/CD, release readiness).

## Baseline findings (initial)

### 1) CI/CD и release gates (High)

- В репозитории отсутствует автоматический CI workflow с обязательными quality gates.
- Риски: незафиксированные регрессии до merge/release, нестабильность quality bar.
- Источник диагностики: devops-анализ Stage 13.

План remediation:

- Добавить формализованные gates для test/race/vet/lint/build.
- Синхронизировать локальный Makefile verify-path и CI-path.

### 2) Сборка и воспроизводимость (High)

- Build/release путь зависит от параметров окружения и несет риск недетерминированных артефактов.
- Риски: сложность проверки релизных бинарников и воспроизводимости.

План remediation:

- Разделить build и tagging flow.
- Добавить reproducibility checks и checksum-проверки в release flow.

### 3) Архитектурные и контрактные hotspot-зоны (Medium)

- Предварительно выявлены зоны повышенной связности в интерактивных helper-паттернах команд и в контрактных точках cmd/internal.*.
- Риски: регрессии при cross-cutting рефакторинге quiet/flags/non-interactive.

План remediation:

- Выполнить file-by-file архитектурный аудит по пакетам cmd/internal/pkg.
- Зафиксировать dependency map и нарушители границ слоев.

### 4) Тестовые и контрактные риски (Medium)

- На baseline все тесты проходят, но матрицы strict quiet/flags/non-interactive и API-compliance еще не собраны.
- Риски: скрытые edge-cases вне текущего набора тестов.

План remediation:

- Сформировать CLI contract matrix и API compliance matrix.
- Расширить regression suite для критичных command trees.

## Tooling constraints observed

- Для части multi-agent диагностики получен 401 token issue в subagent-вызовах.
- Использован fallback read-only analysis path (search diagnostics + локальные baseline проверки).
- На выполнение Stage 13 это не блокер: реализация и проверки продолжаются штатно.

## Step 2 Results - Architecture Conformance

Артефакт шага:

- docs/stage13-architecture-conformance.md

Ключевые результаты:

- Forbidden dependency `internal/pkg -> cmd`: не обнаружено.
- Public boundary `pkg -> internal`: не обнаружено.
- Выявлен coupling hotspot: прямая связка `cmd/compare` с `internal/concurrency` (4 runtime файла + тестовые точки).
- Выявлен duplication hotspot: 17 файлов `cmd/*/interactive_helpers.go`.

Влияние на план Stage 13:

- Добавлен remediation-подпоток "Compare runtime seam hardening".
- Добавлен remediation-подпоток "Interactive helper consolidation".

Статус шага:

- Architecture audit completed.
- Далее: CLI contract matrix (quiet/flags/non-interactive).

## Next workstream

1. API compliance audit: endpoint-by-endpoint matrix and deviations.
2. Reliability pass: race checks и concurrency risk verification.
3. Security & Supply Chain audit.

## Step 3 Results - CLI Contract Audit

Артефакт шага:

- docs/stage13-cli-contract-matrix.md

Ключевые находки:

- HIGH: local `--quiet, -q` override в cmd/run/run.go, cmd/test/list.go, cmd/test/get.go, cmd/result/result.go — shadowing global PersistentFlag.
- MEDIUM: fragile type assertion pattern для non-interactive check (~15 файлов) — нужен `interactive.IsNonInteractive(ctx)` helper.
- LOW: дублирующая `isQuiet()` wrapper функция в cmd/sync/sync_helpers.go.
- MEDIUM: прямые fmt.Fprintf/os.Stdout без quiet-guard в 15 command groups требуют точечного аудита.

Влияние на план Stage 13:

- Добавлены remediation задачи R1-R4 в статус must-fix для Phase 3.

Статус шага:

- CLI Contract Audit completed.
- Далее: API Compliance Matrix.

## Step 4 Results - API Compliance Audit

Артефакт шага:

- docs/stage13-api-compliance-matrix.md

Ключевые находки:

- PASS: Transport/Auth — Basic Auth, Content-Type, User-Agent, ctx propagation, TLS настройка.
- PASS: URL construction — единственная const apiPrefix, нет прямых строк вне client, корректный TestRail & encoding.
- PASS: Pagination — dual-mode decodeListResponse, paginator body.Close в каждой итерации.
- PASS: Interface coverage — compile-time check присутствует, MockClient реализован.
- MEDIUM BUG (F5): internal/client/request.go:54 — ReadJSONResponse non-200 ветка: body leak (нет defer resp.Body.Close()).
- LOW (F6): Group/Role/Dataset/../Label операции без отдельных API-интерфейсных типов.

Влияние на план Stage 13:

- R5 (MEDIUM): Fix ReadJSONResponse body leak — добавить в Phase 3.
- R6 (LOW): Add grouped interfaces — Phase 3, low priority.

## Step 5 Results - Reliability & Concurrency Audit

Артефакт шага:

- docs/stage13-reliability-audit.md

Ключевые находки:

- `go test -race` теперь исполняется (gcc установлен, `CGO_ENABLED=1`).
- Critical package race-run: `internal/concurrency`, `internal/concurrent`, `internal/client` (targeted), `internal/service`, `internal/interactive` — PASS.
- `cmd/compare` race-run выявил `WARNING: DATA RACE` в тесте `TestCompareSectionsInternal_UsesHeavyRuntimeConfig`.
- Причина race: конкурентный append в shared slice `captured` внутри mock closure (`cmd/compare/fetchers_test.go:268`).
- Статус: race исправлен (mutex around append), commit `9358ac8`; повторный `go test -race ./cmd/compare/...` — PASS.
- Оставшийся reliability risk: мутабельная global struct `PriorityThresholds` (LOW/WARN).

Влияние на план Stage 13:

- R7 (INFO): формализовать полный `go test -race ./...` как CI gate.
- R8 (LOW): `PriorityThresholds` сделать read-only/unexport.
- R9 (DONE): test-race fix для compare уже внедрен.

## Step 6 Results - Security & Supply Chain Audit

Ключевые результаты:

- `go mod verify` -> PASS (`all modules verified`).
- `go list -m -u all` выполнен: обнаружены доступные minor/patch updates у части зависимостей (например `github.com/creack/pty`, `github.com/fatih/color`, `github.com/go-viper/mapstructure/v2`, `github.com/google/go-cmp`).
- `govulncheck` отсутствует в environment (инструмент не установлен), vulnerability scan не выполнен автоматически.
- Базовый secret-pattern scan по репозиторию не выявил признаков реальных секретов; найденные совпадения относятся к флагам/полям (`api-key`, `password`) и placeholder-конфигу (`your_api_key_here`).
- Exec usage ограничен и контролируем: `internal/ui/editor.go` (launch editor через `exec.Command`) и `internal/selftest/checks.go` (локальные selftest команды).

Найденные риски:

- LOW (F13): dependency freshness debt — есть устаревшие модули; требует планового dependency refresh.
- LOW (F14): `govulncheck` отсутствует локально — нет автоматического vuln gate в текущем окружении.

Влияние на план Stage 13:

- R10 (LOW): подготовить dependency bump план для безопасных patch/minor обновлений.
- R11 (MEDIUM): добавить `govulncheck ./...` в CI gate (или эквивалентный vulnerability scan шаг).

## Step 7 Results - CI/CD Hardening Audit

Ключевые результаты:

- В `.github/workflows` отсутствуют workflow-файлы (CI pipeline не формализован).
- В `Makefile` отсутствует единая verify-цель с обязательными gates (`build`, `test`, `vet`, `race`, `lint`).
- `build` зависит от `sync-tag`, который может мутировать git-состояние (создание локального tag) во время локальной сборки.
- Release path (`release`, `release-compressed`) не публикует checksums и не фиксирует воспроизводимые build inputs.
- `release-workflow.md` описывает branch flow, но не содержит machine-checkable quality gate contract.

Найденные риски:

- HIGH (F15): отсутствует CI workflow с обязательными quality gates.
- MEDIUM (F16): `build -> sync-tag` смешивает compile и release/tagging responsibilities.
- MEDIUM (F17): отсутствуют checksum/verify шаги для release артефактов.

Влияние на план Stage 13:

- R12 (HIGH): добавить CI workflow с gates: `go test`, `go vet`, `go build`, `go test -race`, `govulncheck`.
- R13 (MEDIUM): разделить `build` и `sync-tag` в Makefile (tagging только явной release целью).
- R14 (MEDIUM): добавить release checksum + verification шаг (например `sha256sum` для каждого артефакта).
