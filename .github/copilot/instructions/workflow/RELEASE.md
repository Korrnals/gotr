# Workflow: Release 3.0.0

> **Created:** 2026-03-05  
> **Updated:** 2026-03-11  
> **Status:** IN PROGRESS  
> **Release branch:** `release-3.0.0`  
> **Base protocol:** `.protocols/core/04-release-flow.md`  

---

## Release Scope

### Stages included in release-3.0.0:

| Stage | Branch | Status | Description |
|-------|--------|--------|-------------|
| 6.7 | `stage-6.7-recursive-parallelization` | ✅ Merged | Recursive parallelization, reporter rewrite |
| 6.8 | `stage-6.8-concurrency-unification` | ✅ Merged | Concurrency unification, generic compare |
| 6.9 | `stage-6.9-pagination-audit` | ✅ Merged | Generic paginator, 9 API methods migrated |
| 7.0 | `stage-7.0-context-propagation` | ✅ Merged | context.Context in all layers, Ctrl+C support |
| 8.0 | `stage-8.0-output-refactoring` | ✅ Merged | UI/output refactoring |

> **Обоснование major version 3.0.0:** Breaking changes в API-сигнатурах (добавлен `context.Context` во все ~100 методов клиента), глобальная унификация concurrency, пагинация.

---

## Workflow Steps

### Step 1: Create release branch ✅

```bash
git checkout -b release-3.0.0 origin/main
git push -u origin release-3.0.0
```

### Step 2: Merge Stages 6.7–7.0 → release-3.0.0 ✅

All four stages merged via PRs.

### Step 3: Complete Stage 8.0 ✅

1. Branch `stage-8.0-output-refactoring` from `stage-7.0-context-propagation`
2. Implement Stage 8.0 (see `STAGE_8.0_DESIGN.md`)
3. Merged into `release-3.0.0`

### Step 4: Finalize release-3.0.0 ⏳

1. Verify on release branch:
   - `go test ./...` — all tests pass ✅
   - `go build ./...` — build succeeds ✅
   - `go vet ./...` — no warnings ✅
   - `gotr compare cases` — smoke test with real TestRail
   - `gotr compare all` — smoke test orchestrator
2. Update CHANGELOG.md ✅
3. PR: `release-3.0.0` → `main` (ONLY via Pull Request!)
4. Review & Merge PR
5. Tag: `v3.0.0`

---

## Branch Protection Rules

**`main` branch is protected** (configured 2026-03-12):

- **Direct push:** FORBIDDEN
- **Direct merge:** FORBIDDEN
- **Only via Pull Request:** YES
- **Enforce for admins:** YES
- **Force push:** FORBIDDEN
- **Branch deletion:** FORBIDDEN

> **НИКОГДА** не мержить напрямую в `main`! Только через PR.
> Это правило действует для всех, включая admin.

---

## Smoke Tests for release-3.0.0

### Minimum set:

```bash
# Build
go build -o gotr ./...

# Unit tests
go test ./...

# Functional tests (require TestRail credentials)
./gotr compare cases -p1 <PROJECT1> -p2 <PROJECT2>
./gotr compare all -p1 <PROJECT1> -p2 <PROJECT2>
./gotr compare suites -p1 <PROJECT1> -p2 <PROJECT2>
```

### Verification checklist:

- [ ] Reporter displays stats correctly (no emoji, ANSI colors, proper alignment)
- [ ] Parallel case loading works
- [ ] `compare all` orchestrator works with reporter
- [ ] Errors handled gracefully (retry, warning instead of fatal)

---

## Versioning

- **3.0.0** — major release (breaking API changes):
  - Stage 6.7: recursive parallelization + reporter rewrite
  - Stage 6.8: concurrency unification + generic compare subcommands
  - Stage 6.9: generic paginator for all list API methods
  - Stage 7.0: context.Context propagation + Ctrl+C support
  - Stage 8.0: UI/output refactoring, unified rendering
- **go.mod:** `github.com/Korrnals/gotr` — module path unchanged
- **Binary:** `gotr` — binary name unchanged

---

## Related Documents

- [STAGE_6.7_DESIGN.md](../STAGE_6.7_DESIGN.md)
- [STAGE_6.7_IMPLEMENTATION.md](../STAGE_6.7_IMPLEMENTATION.md)
- [STAGE_6.8_DESIGN.md](../STAGE_6.8_DESIGN.md)
- [PLAN.md](../PLAN.md)
- [AUDIT_2026-03-03.md](../AUDIT_2026-03-03.md)

---

**END OF RELEASE WORKFLOW**
