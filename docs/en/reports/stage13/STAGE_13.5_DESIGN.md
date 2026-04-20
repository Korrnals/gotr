# Stage 13.5 — Full API Coverage & 100% Test Parity

Language: English | [Русский](../../../ru/reports/stage13/STAGE_13.5_DESIGN.md)

## Navigation

- [Documentation](../../index.md)
  - [Guides](../../guides/index.md)
    - [Instructions](../../guides/instructions/index.md)
  - [Architecture](../../architecture/index.md)
  - [Operations](../../operations/index.md)
  - [Reports](../index.md)
    - [Stage 13](index.md)
- [Home](../../../../README.md)

---

## Stage Goal

Bring gotr v3.0 to full TestRail API v2 CLI wrapper coverage and achieve 100% test coverage across all packages. Upon completion — mandatory full re-audit per `.github/prompts/full-project-audit.prompt.md` template.

## Background

Based on the Stage 13.0 final audit results (2026-04-06):

- **42/42 packages PASS** with `-race`, zero data races
- **Coverage**: 36 packages @ 100%, 6 packages @ 96.8-98.7%
- **API client**: 147 methods implemented (superset of TestRail API v2)
- **CLI exposure**: ~90.5%, 14 client methods without a direct CLI wrapper
- **api_paths.go**: ~15% of endpoints missing from the registry

---

## Phase 1 — api_paths.go Completeness (endpoint documentation)

### 1.1. Audit and Supplement api_paths.go

Current state: 128 endpoints documented. Need to add ~15 missing:

| Group | Missing Endpoints | Type |
|-------|------------------|------|
| Attachments | `get_attachment/{id}`, `get_attachments_for_case/{id}`, `get_attachments_for_plan/{id}`, `get_attachments_for_plan_entry/{id}`, `get_attachments_for_run/{id}`, `get_attachments_for_test/{id}`, `get_attachments_for_project/{id}` | GET |
| Users | `add_user`, `update_user/{id}`, `get_users/{project_id}` | POST/GET |
| Reports | `get_cross_project_reports` | GET |
| Labels | `get_label/{id}`, `get_labels/{project_id}` | GET |

- [ ] Add all missing endpoints to `pkg/testrailapi/api_paths.go`
- [ ] Update api_paths_test.go tests (add checks for new paths)
- [ ] Commit: `feat(api): complete api_paths.go endpoint registry`

---

## Phase 2 — CLI Wrappers for Missing Operations

### 2.1. User Management CLI (HIGH)

Current state: client methods `AddUser()`, `UpdateUser()`, `GetUsersByProject()` are implemented, CLI is missing.

- [ ] `gotr users add --name --email --role-id` — wrapper over `AddUser()`
- [ ] `gotr users update <user_id> --name --email --role-id` — wrapper over `UpdateUser()`
- [ ] `gotr users list --project-id` — support `GetUsersByProject()` via existing list
- [ ] Add interactive mode for `users add/update`
- [ ] Tests: table-driven for add/update/list-by-project (quiet + JSON + interactive)
- [ ] Commit: `feat(users): add/update CLI commands with interactive mode`

### 2.2. Reference Data CLI (MEDIUM)

Methods `GetPriorities()`, `GetStatuses()`, `GetResultFields()` are implemented in the client but lack convenient CLI wrappers.

- [ ] `gotr list priorities` — endpoint via generic list
- [ ] `gotr list statuses` — endpoint via generic list
- [ ] `gotr list resultfields` — endpoint via generic list
- [ ] Verify that generic `list` already routes to these resources; if not — add routing
- [ ] Tests: table-driven for each new resource
- [ ] Commit: `feat(list): expose priorities/statuses/resultfields via generic list`

### 2.3. Attachments Get/List-by-Context (MEDIUM)

Client methods `GetAttachment()`, `GetAttachmentsFor*()` — 7 methods, limited CLI access.

- [ ] `gotr attachments get <attachment_id>` — download/metadata
- [ ] `gotr attachments list --for-case <id>` / `--for-run <id>` / `--for-plan <id>` / `--for-test <id>` / `--for-project <id>` — context-aware lists
- [ ] Tests: table-driven for each context
- [ ] Commit: `feat(attachments): context-aware list and get commands`

### 2.4. Cross-Project Reports (LOW)

Client methods `GetCrossProjectReports()` and `RunCrossProjectReport()` are implemented.

- [ ] Verify that `gotr reports list-cross` and `gotr reports run-cross` work correctly
- [ ] Add tests if missing
- [ ] Commit: `test(reports): cross-project report coverage`

---

## Phase 3 — Test Coverage to 100%

### 3.1. Packages with Coverage < 100%

| Package | Current | Target | Delta | Strategy |
|---------|---------|--------|-------|----------|
| `cmd/sync` | 96.8% | 100% | +3.2% | Cover error branches, edge cases in sync_full/sync_cases |
| `cmd/get` | 96.9% | 100% | +3.1% | Add tests for uncovered branches in get/* |
| `cmd/run` | 97.1% | 100% | +2.9% | Error paths in create/update/close |
| `cmd` (root) | 97.3% | 100% | +2.7% | Uncovered branches in commands.go/root.go |
| `cmd/result` | 97.6% | 100% | +2.4% | service_wrapper.go error paths |
| `internal/ui` | 98.7% | 100% | +1.3% | Preview edge cases, format edge paths |

- [ ] For each package: identify uncovered lines via `go test -coverprofile`
- [ ] Write targeted tests for each uncovered branch
- [ ] Each package is committed separately:
  - [ ] `test(sync): bring cmd/sync to 100% coverage`
  - [ ] `test(get): bring cmd/get to 100% coverage`
  - [ ] `test(run): bring cmd/run to 100% coverage`
  - [ ] `test(cmd): bring root cmd to 100% coverage`
  - [ ] `test(result): bring cmd/result to 100% coverage`
  - [ ] `test(ui): bring internal/ui to 100% coverage`

---

## Phase 4 — Lint Finding Resolution (golangci-lint v2)

### Context

During CI migration to golangci-lint v2.11.4 (Go 1.25-compatible), ~290 pre-existing findings were discovered. Linter v1.64.8 (Go 1.24) could never run with Go 1.25, so these issues were not visible before. The lint step in CI runs with `continue-on-error: true` until this phase is complete.

### 4.1. Finding Statistics (baseline 2026-04-06)

| Linter | Count | Priority |
|--------|-------|----------|
| gocritic | ~90 | HIGH — style and performance hints |
| errcheck | ~52 | HIGH — unchecked errors |
| staticcheck | ~47 | HIGH — potential bugs |
| misspell | ~45 | LOW — typos in comments |
| gocyclo | ~16 | MEDIUM — function complexity |
| unused | ~15 | MEDIUM — unused code |
| nolintlint | ~9 | LOW — invalid nolint directives |
| ineffassign | ~2 | HIGH — assignments without use |

### 4.2. Remediation Plan

- [x] **Batch 1 — errcheck + ineffassign** (HIGH): add handling/ignoring of returned errors
- [x] **Batch 2 — staticcheck** (HIGH): fix potential bugs and deprecated calls
- [x] **Batch 3 — gocritic** (HIGH): refactor per style/performance recommendations
- [x] **Batch 4 — unused** (MEDIUM): remove unused code
- [x] **Batch 5 — gocyclo** (MEDIUM): simplify complex functions or justify `//nolint`
- [x] **Batch 6 — misspell + nolintlint** (LOW): typo fixes, directive cleanup
- [x] Final run: `golangci-lint run` EXIT 0
- [x] Remove `continue-on-error: true` from CI workflow (lint step)
- [x] Commit: `fix(lint): resolve all golangci-lint v2 findings`

---

## Phase 5 — Validation

- [x] Full run: `go test ./...` — 42/42 PASS
- [x] Full run: `go test -race ./...` — 41/41 PASS, 0 races (excl. concurrency — CI skip)
- [x] Full run: `go test -cover ./...` — 35/42 @ 100%, 7 @ 97.4–99.8%
- [x] `go vet ./...` — PASS
- [x] `go build ./...` — PASS
- [x] `golangci-lint run` — EXIT 0, zero findings
- [x] Coverage: 42/42 PASS (min 97.4% cmd/sync, avg >99.5%)

---

## Phase 6 — Full Re-Audit

**MANDATORY** final audit per `.github/prompts/full-project-audit.prompt.md` template:

- [x] Phase 0: Scope & skip list
- [x] Phase 1: Architecture & layers — CONDITIONAL PASS (0C/0H/3M/2L)
- [x] Phase 2: TestRail API compliance — PASS (135 endpoints, 98%)
- [x] Phase 3: Code quality & patterns — CONDITIONAL PASS (0C/0H/4M/4L)
- [x] Phase 4: Tests & race detector — PASS (42/42 ≥97.4%, 0 races)
- [x] Phase 5: Documentation — CONDITIONAL PASS (0C/2H/3M/3L)
- [x] Phase 6: CI/Build/Security gates — PASS (6 stdlib vulns, 0 dep)
- [x] Phase 7: Consolidation report — CONDITIONAL PASS

Audit verdict: **CONDITIONAL PASS** — 2 HIGH in README require fix before PR.

---

## Phase 6.5 — Remediation: DRY CRUD + Compare Decouple

**Goal:** Close all Phase 6 audit findings (B-1, B-5) + i18n unification.

### 6.5.1. i18n — Codebase Language Unification

- [x] Pass 1: Russian→English in test descriptions and comments (8 files)
- [x] Pass 2: Full translation of all 1738 Cyrillic strings across 170+ Go files
- [x] Verification: `grep -rn '[а-яА-ЯёЁ]' --include='*.go'` → 0 matches
- [x] Build + Tests + Lint — PASS
- [x] Commit: `i18n: translate all Russian text to English in Go source files`

### 6.5.2. B-5 — Compare Resource Registry ✅

**Problem:** `cmd/compare/all.go` hard-codes calls to 12 compare functions sequentially. Adding a new entity → editing all.go.

**Solution:** Registry pattern — `resourceEntry` struct + `resourceRegistry` array.

- [x] `resourceEntry` struct (displayName, key, label, accessor, factory)
- [x] `resourceRegistry` — single array of 12 entries
- [x] `newSimpleResourceEntry()` factory for 9 simple resources
- [x] `init()` auto-populates `compareAllStages` from registry
- [x] `runCompareAllResources()` builds steps from registry
- [x] Verification: `go test ./cmd/compare/... -count=1` — PASS (3.5s)
- [x] all.go: 726→~680 LOC, triple duplication eliminated

### 6.5.3. B-1 — DRY CRUD via Generic Executor ✅

**Problem:** `cmd/add.go` (1200 LOC), `cmd/update.go` (850 LOC) — 70% boilerplate.

**Solution:** `internal/crud/executor.go` — Go generics `Execute[Req, Resp]` + `DryRun[Req]`.
For each entity — a single `buildXxxReq(cmd, validate bool)` function, shared between execute and dry-run.

| Step | Description | Status |
| --- | --- | --- |
| 1 | `internal/crud/executor.go` — `Execute[Req, Resp]` + `DryRun[Req]` (79 LOC) | ✅ Done |
| 2 | `internal/crud/executor_test.go` — 7 tests (162 LOC) | ✅ Done |
| 3 | cmd/add.go: 7 buildReq + 7 slim addXxx + 7 slim dryRunXxx | ✅ Done |
| 4 | cmd/update.go: 6 buildReq + 6 slim updateXxx + 6 slim dryRunXxx | ✅ Done |
| 5 | cmd/delete.go — already concise (300 LOC), no refactoring needed | ✅ Skip |
| 6 | Verification: 260 tests PASS, 0 lint issues, vet OK | ✅ Done |

**Result:**
- add.go: 1200→1057 LOC (-143, -12%)
- update.go: 850→697 LOC (-153, -18%)
- Net: -217 LOC production code + 241 LOC (executor+tests)

---

## Phase 7 — Closure

- [x] Update docs/reports (quality-metrics, audit-report)
- [x] Finalize CHANGELOG
- [x] Sync command guides with CLI (attachments, sync, bdds — EN+RU)
- [x] Final comprehensive audit (automated gates + static scan + deep audit + security scan)
- [x] Hardening: `formatAPIError()` — `io.LimitReader` on response body
- [x] Record Phase 7 results in reports
- [x] Create PR: stage-13.5 → release-3.0.0 (PR #17, merged 2026-04-09)
- [x] Create PR: release-3.0.0 → main
- [ ] Tag v3.0.0 after merge to main

---

## Expected Metrics After Stage 13.5

| Metric | Current (13.0) | Target (13.5) |
|--------|----------------|---------------|
| Test coverage total | 96.8-100% | **100.0%** |
| Packages @ 100% | 36/42 | **42/42** |
| API endpoints in api_paths.go | 128 | **~143** |
| CLI-accessible operations | ~90.5% | **~98%** |
| Data races | 0 | **0** |
| go vet warnings | 0 | **0** |
| golangci-lint findings | ~290 (non-blocking) | **0** |
| Audit verdict | CONDITIONAL PASS | **UNCONDITIONAL PASS** |

---

## Working Mode

- **stepwise**: one step → report → confirmation
- Each phase is committed separately
- Docs shadow-sync is mandatory for each change cluster
- Checkpoint after each completed step

---

← [Stage 13](index.md) · [Reports](../index.md) · [Documentation](../../index.md)
