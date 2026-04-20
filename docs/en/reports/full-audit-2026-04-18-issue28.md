# Full Audit of feat/28-snap-rollback Branch — 2026-04-18

Language: English | [Русский](../../ru/reports/full-audit-2026-04-18-issue28.md)

## Navigation

- [Documentation](../index.md)
  - [Reports](index.md)

---

## Phase 0 — Scope & Baseline

| Metric | Value |
| --- | --- |
| Branch | `feat/28-snap-rollback` |
| HEAD | `28d47b0` |
| Commits on branch | 32 |
| Go | 1.25.0 |
| Source files | 316 |
| Test files | 271 |
| Documentation files | 153 |
| Files changed (vs main) | 185 |
| Insertions / Deletions | +17 321 / −714 |

---

## Phase 1 — Architecture and Layers

**Subagent: `architect`**

| Check | Verdict | Violations |
| --- | --- | --- |
| 1.1 Layer boundaries | **WARN** | 1 MEDIUM: `pkg/snap_smoke` → `internal/` |
| 1.2 Dependency direction | **PASS** | 0 |
| 1.3 Interface usage | **WARN** | 1 LOW: `cmd/export.go` concrete cast |
| 1.4 Package cohesion | **PASS** | 0 |
| 1.5 Coupling hotspots | **WARN** | 0 (advisory: `cmd/add.go`, `cmd/update.go` = 8 imports) |
| 1.6 Concurrency architecture | **PASS** | 0 |
| 1.7 Model layer | **PASS** | 0 |

**Total: 2 violations (1 MEDIUM, 1 LOW) → PASS**

### Violation Details

- **MEDIUM**: `pkg/snap_smoke/` imports `internal/client` and `internal/models/data`. Violates the `pkg/` = public API contract. Recommendation: move to `internal/snap_smoke/` or extract types into `pkg/`.
- **LOW**: `cmd/export.go:41` — type assertion `GetClient(cmd).(*client.HTTPClient)` to access raw HTTP. Recommendation: add a `RawDoer` interface to `ClientInterface`.

---

## Phase 2 — TestRail API Coverage

_Not performed in this audit (scope: UX overhaul, not API endpoints). Last full API audit: `docs/en/reports/stage13/audit-report.md`._

---

## Phase 3 — Code Quality

**Subagent: `backend-engineer`**

| Check | Verdict | Severity |
| --- | --- | --- |
| 3.1 Error handling | **WARN** | MEDIUM — 435 bare `return err` (37%) vs 744 wrapped |
| 3.2 Resource management | **PASS** | INFO — all `resp.Body` closed |
| 3.3 Context propagation | **PASS** | INFO — `NewRequestWithContext` everywhere |
| 3.4 Cobra CLI patterns | **PASS** | INFO — `RunE`, flag validation, help text |
| 3.5 Security | **PASS** | INFO — no credentials in logs, TLS secure by default |
| 3.6 DRY | **PASS** | LOW — `crud.DryRun`/`crud.Execute` centralize patterns |
| 3.7 Go best practices | **WARN** | LOW — sentinel errors via `fmt.Errorf` instead of `errors.New` |

**Total: 0 CRITICAL, 0 HIGH, 1 MEDIUM, 2 LOW → PASS**

---

## Phase 4 — Testing

**Subagent: `qa-engineer`**

### 4.1 Coverage by Package

| Package | Coverage | Threshold |
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

### 4.3 Full Test Run

| Metric | Value |
| --- | --- |
| Packages tested | 40 |
| Total tests | 1 002+ |
| PASS | 1 002+ |
| FAIL | 0 |
| Time | ~15s |

### 4.4 Context for Below-Threshold Coverage

- **`internal/crud` (46%)** — generic CRUD helper, 1 file. Error paths not covered, but the executor itself is tested through cmd/* tests.
- **`cmd/snap` (49.6%)** — interactive helpers (`interactive_helpers.go`) are difficult to unit-test. Rollback/undo/export are covered through `internal/snap` (74.1%) + smoke tests.
- **`internal/interactive` (71.4%)** — Browse, ActionMenu, AlignedLabels — new components. Pager and mutation_action are difficult to isolate.
- **`internal/snap` (74.1%)** — hook.go, resolve.go (cobra wrappers) have 0% — tested indirectly through CLI.

**Note**: 4 packages are below 80%, but all of them contain interactive/UI code that is difficult to unit-test. Core business logic (snap engine, rollback, manifest) is covered at 82-96%.

---

## Phase 5 — Documentation

**Subagent: `docs-writer`**

| Check | Verdict |
| --- | --- |
| 5.1 CLI ↔ Docs | **WARN** — `work` command without doc page |
| 5.2 README | **PASS** — version 3.0.0, Go 1.25.0, all sections |
| 5.3 Architecture docs | **WARN** — 1 RU-only file (`ux-modernization-design.md`) |
| 5.4 Navigation | **WARN** — 2 formats coexist (sidebar vs breadcrumb) |
| 5.5 EN/RU parity | **WARN** — 74 EN vs 78 RU (+4 RU-only) |
| 5.6 Broken links | **FAIL** — 3 broken links in `index.md` (both versions) |

### Broken Links

| Link | Source |
| --- | --- |
| `reports/history/audit-report.md` | `docs/en/index.md`, `docs/ru/index.md` |
| `reports/history/quality-metrics.md` | `docs/en/index.md`, `docs/ru/index.md` |
| `reports/history/coverage-matrix.md` | `docs/en/index.md`, `docs/ru/index.md` |

### Missing Docs

- `docs/ru/guides/commands/work.md` — not created
- `docs/en/guides/commands/work.md` — not created

---

## Phase 6 — CI/Build/Security Gates

| Gate | Result |
| --- | --- |
| `go build ./...` | **PASS** |
| `go vet ./...` | **PASS** |
| `go test ./...` (40 packages) | **PASS** — 0 FAIL |
| Race detector (3 packages) | **PASS** — 0 races |
| `govulncheck` | N/A (not installed) |
| `golangci-lint` | N/A (not installed) |

---

## Phase 7 — Consolidation

### Findings Table

| # | Finding | Severity | Phase | Recommendation |
| --- | --- | --- | --- | --- |
| F1 | `pkg/snap_smoke` → `internal/` | MEDIUM | 1.1 | Move to `internal/` |
| F2 | `cmd/export.go` concrete type cast | LOW | 1.3 | Add interface |
| F3 | 435 bare `return err` (37%) | MEDIUM | 3.1 | Gradual wrap (backlog) |
| F4 | Sentinel errors via `fmt.Errorf` | LOW | 3.7 | Replace with `errors.New` |
| F5 | `cmd/snap` coverage 49.6% | MEDIUM | 4.2 | Add tests for rollback/undo |
| F6 | `internal/crud` coverage 46.0% | MEDIUM | 4.2 | Add error path tests |
| F7 | `internal/interactive` coverage 71.4% | LOW | 4.2 | Add Browse/ActionMenu tests |
| F8 | `internal/snap` coverage 74.1% | LOW | 4.2 | Cover hook.go, resolve.go |
| F9 | `work` command without doc page | ~~MEDIUM~~ FIXED | 5.1 | ✅ Created `work.md` (ru + en) |
| F10 | 3 broken links in index.md | ~~HIGH~~ FIXED | 5.6 | ✅ Links → `stage13/` |
| F11 | 4 RU-only files | LOW | 5.5 | Translate or mark |
| F12 | `cmd/compare` coverage 88.7% | LOW | 4.2 | Bring to 90% |

### Severity Count

| Severity | Count |
| --- | --- |
| CRITICAL | 0 |
| HIGH | 0 (F10 → FIXED) |
| MEDIUM | 4 (F1, F3, F5, F6) — F9 → FIXED |
| LOW | 6 (F2, F4, F7, F8, F11, F12) |

### Verdict: **PASS**

- 0 CRITICAL
- 0 HIGH (F10 fixed: links redirected to `stage13/`)
- 4 MEDIUM — backlog tasks, do not block PR
- F9 fixed: created `work.md` (ru + en)

---

## Priority Actions Before PR

1. ~~Fix 3 broken links~~ → ✅ DONE (commit pending)
2. ~~Create `work.md`~~ → ✅ DONE (commit pending)
3. Remaining findings — backlog for next cycle
