# Snap/Rollback + Interactive UX Audit — 2026-04-18

Language: English | [Русский](../../ru/reports/snap-audit-2026-04-14.md)

## Navigation

- [Documentation](../index.md)
  - [Reports](index.md)

---

## 1. Overall Statistics

| Metric | Value |
| --- | --- |
| Branch | `feat/28-snap-rollback` |
| HEAD | `28d47b0` (32 commits) |
| Files changed | 185 |
| Insertions / Deletions | +17 321 / −714 |
| Source files (snap) | 19 |
| Test files (snap) | 8 |
| Total LOC (snap) | 9 705 |
| Source code LOC (snap) | 4 469 |
| Test code LOC (snap) | 5 236 |
| Test/code ratio (snap) | 1.17:1 |
| Source files (interactive) | 16 |
| Test files (interactive) | 12 |
| Total LOC (interactive) | 3 087 |
| Source code LOC (interactive) | 1 627 |
| Test code LOC (interactive) | 1 460 |

---

## 2. Testing

### 2.1. Test Run Results

| Package | Tests | PASS | FAIL | Time |
| --- | --- | --- | --- | --- |
| `internal/snap` | 119 | 119 | 0 | 0.054s |
| `cmd/snap` | 38 | 38 | 0 | 0.049s |
| `internal/interactive` | 83 | 83 | 0 | 0.015s |
| `cmd/compare` | 300+ | 300+ | 0 | 3.5s |
| `cmd/sync` | 100+ | 100+ | 0 | 0.4s |
| `cmd/` | 400+ | 400+ | 0 | 3.5s |
| **Total** | **1 002** | **1 002** | **0** | **~7.5s** |

### 2.2. Coverage

| Package | Coverage |
| --- | --- |
| `internal/snap` | 74.1% |
| `cmd/snap` | 49.6% |
| `internal/interactive` | 71.4% |
| `cmd/compare` | 88.7% |
| `cmd/sync` | 90.6% |
| `cmd/` | 94.1% |

| File | Coverage | Note |
| --- | --- | --- |
| types.go | 100% | — |
| manifest.go | ~95% | `Latest()` — 44% |
| rollback.go | ~82% | `Rollback()` — 96.2% |
| store.go | ~78% | `atomicWriteJSON` error branches |
| snapshot.go | ~77% | `SnapOrWarn` — 0% (wrapper) |
| attachments.go | ~83% | `decompressToTemp` — 68.8% |
| hook.go | 0% | Thin wrapper, tested via CLI |
| resolve.go | 0% | Reads cobra flags, tested via CLI |
| info.go | 0% | 1 UI wrapper function |

### 2.3. Functions with Zero Coverage (12 total)

- `hook.go`: `NewHook`, `Before`, `FinalizeAdd`, `FinalizeSyncData`, `HookMutation`
- `resolve.go`: `RegisterFlags`, `ResolveDecision`, `ResolveName`, `ReadConfig`
- `snapshot.go`: `SnapOrWarn`
- `store.go`: `NewStore`
- `info.go`: `InfoBanner`

All are thin wrappers over cobra/viper/ui — tested indirectly via CLI smoke tests.

---

## 3. Code Quality

### 3.1. Static Analysis

| Tool | Result |
| --- | --- |
| `go build ./...` | OK |
| `go vet ./...` | 0 findings |
| `go fmt` | Clean |

### 3.2. Project Standards (code-quality.md)

| Rule | Status |
| --- | --- |
| `fmt.Errorf("...: %w", err)` mandatory wrap | PASS — 100% of errors wrapped |
| No `return err` without wrapping | PASS |
| No `os.Exit`/`log.Fatal` in library code | PASS |
| `GetClient()` for HTTP | PASS — via DI `clientFn` |
| `ctx context.Context` on all I/O | PASS |
| Naked returns only in tests | PASS |

### 3.3. Architectural Rules (STANDARDS.md)

| Rule | Status |
| --- | --- |
| `cmd/` → `internal/` (not reverse) | PASS |
| `pkg/` → not `internal/` | PASS |
| Interface Segregation | PASS — `CasesAPI` (5), `AttachmentsAPI` (3), `PromptFunc` (1) |
| Constructor injection | PASS — `Register(rootCmd, clientFn)` |
| `RunE` (not `Run`) | PASS |
| Paths via `paths.*` | PASS |
| Atomic file writes | PASS — `atomicWriteJSON()` |
| Thread-safe manifest | PASS — `sync.Mutex` + defer |

### 3.4. Single Warning

| File | Line | Description |
| --- | --- | --- |
| `cmd/snap/rollback.go` | 61, 63, 102 | Discarded errors from `cmd.Flags().GetBool/GetString` — safe (cobra guarantee), but stylistically imperfect |

---

## 4. Security

**Overall risk level: LOW** — CLI tool with local storage.

| Category | Rating | Justification |
| --- | --- | --- |
| Path Traversal | LOW | ID is generated programmatically, sanitizeName cleans `/\: ` |
| Export path | INFO | CLI tool, analogous to `cp file /path` |
| Temp files | LOW | `sanitizeAttachName` + `os.CreateTemp` in `os.TempDir()` |
| File permissions | LOW | `os.Create()` → umask 0644, `~/.gotr/snaps/` — user-owned |
| Race conditions | LOW | `sync.Mutex` on manifest, single-user |
| JSON deserialization | SAFE | Type-safe `json.Decoder` |
| DoS | SAFE | `AttachMaxFileMB` + `concurrentThreshold` |

---

## 5. Documentation

| Subcommand | RU doc | EN doc | Flags | Examples | Interactive |
| --- | --- | --- | --- | --- | --- |
| list | OK | OK | — | OK | OK |
| info [id] | OK | OK | — | OK | OK |
| rollback [id] | OK | OK | --dry-run, --entity-ids | OK | OK |
| rollback list | OK | OK | — | OK | OK |
| rollback undo [id] | OK | OK | — | OK | OK |
| export [id] [path] | OK | OK | — | OK | OK |
| delete [id] | OK | OK | — | OK | OK |
| gc | OK | OK | — | OK | OK |
| `gotr work` hub | OK | OK | — | OK | OK |
| Cross-navigation | OK | OK | — | OK | OK |

RU/EN synchronization: perfect. Score: 98/100.

---

## 6. Good Practices

- Atomic file writes — `atomicWriteJSON()`
- Thread-safe manifest — `sync.Mutex` + `defer unlock`
- Graceful degradation — `SnapOrWarn()`
- Dry-run support — preview with diff table + confirm
- Resume capability — rollback skips already rolled-back items
- Interface segregation — small interfaces
- Categorized storage — `~/.gotr/snaps/{cases,sections,sync,custom}/`

---

## 7. Summary Matrix

| Category | Score | Verdict |
| --- | --- | --- |
| Tests | 1002 PASS / 0 FAIL, 74–94% coverage | PASS |
| Code standards | 0 critical, 1 warning | PASS |
| Architecture | Full compliance | PASS |
| Security | LOW risk | PASS |
| Documentation | 98/100, RU/EN in sync | PASS |
| Build / Vet | Clean | PASS |

---

## 8. Identified UX Issues (all fixed)

| # | Issue | Severity | Status | Commit |
| --- | --- | --- | --- | --- |
| UX-1 | `snap list` — flat table output | HIGH | ✅ FIXED | `3ec7a7a` |
| UX-2 | `snap info` — raw JSON | HIGH | ✅ FIXED | `3ec7a7a` |
| UX-3 | No grouping by servers | HIGH | ✅ FIXED | `3ec7a7a` |
| UX-4 | Dry-run preview unclear | MEDIUM | ✅ FIXED | `c5951ed` |
| UX-5 | Error 400 without explanation | MEDIUM | ✅ FIXED | `c5951ed` |
| UX-6 | Export without interactive selection | MEDIUM | ✅ FIXED | `3ec7a7a` |
| UX-7 | Picker without server context | HIGH | ✅ FIXED | `595ba00` |

### Additional UX Improvements (commit `28d47b0` — `658a6ad`)

| Improvement | Description |
| --- | --- |
| AlignedLabels | Aligned columns in all interactive pickers |
| Browse | Pagination with ← Back navigation |
| ActionMenu | Unified post-action menus with string keys |
| CrossNavOptions | Cross-navigation Compare/Sync/Snap from any post-action |
| HandleCrossNav | Server guard when switching between commands |
| Server banner | `Server: <url>` in interactive mode |
| `gotr work` | Interactive navigation hub with group menus |
