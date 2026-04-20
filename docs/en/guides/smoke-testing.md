# Smoke Testing (snap_smoke)

Language: [Русский](../../ru/guides/smoke-testing.md) | English

## Navigation

- [Documentation](../index.md)
  - [Guides](index.md)
    - [Installation](installation.md)
    - [Configuration](configuration.md)
    - [Interactive Mode](interactive-mode.md)
    - [Progress](progress.md)
    - [Smoke Testing](smoke-testing.md)
    - [Commands Index](commands/index.md)
    - [Instructions](instructions/index.md)
  - [Architecture](../architecture/index.md)
  - [Operations](../operations/index.md)
  - [Reports](../reports/index.md)
- [Home](../../../README.md)

## Overview

The `pkg/snap_smoke` package provides end-to-end smoke tests for snap/rollback functionality.
Tests verify the full cycle: take snapshot → mutate data → rollback → verify result.

## Architecture

```
pkg/snap_smoke/
├── doc.go           — package documentation
├── config.go        — configuration from env vars, client creation
├── helpers.go       — helpers: testCase (with auto-cleanup) and testSection (reuse)
├── testserver.go    — FakeTestRail: built-in in-memory TestRail API v2 server
└── smoke_test.go    — 6 E2E tests behind //go:build smoke tag
```

### FakeTestRail

Built-in mock server based on `net/http/httptest`. Routing is driven by the
`pkg/testrailapi.APIPath` catalog — URI templates are compiled into regex automatically,
ensuring the mock matches the real API structure. Emulates a subset of TestRail API v2:

| Method | Endpoint | Description |
| ------ | -------- | ----------- |
| GET | `get_case/{case_id}` | Get case |
| POST | `add_case/{section_id}` | Create case |
| POST | `update_case/{case_id}` | Update case (partial update) |
| POST | `delete_case/{case_id}` | Delete case |
| GET | `get_sections/{project_id}` | List sections (paginated wrapper) |
| POST | `add_section/{project_id}` | Create section |

Key features:
- **Routes from testrailapi** — URI templates from `testrailapi.New().Paths()` are compiled into regex (path) + query-template map. If a URI changes in `testrailapi`, the mock updates automatically.
- **Paginated responses** — `get_sections` returns TestRail 6.7+ format: `{"offset":0, "limit":250, "size":N, "_links":{...}, "sections":[...]}`. The client's `fetchAllPages` goes through the full parsing cycle.
- **Storage** — `sync.Mutex` + `map[int64]*Case/Section`. Auto-incrementing IDs (cases: 1000+, sections: 100+).

## Two Modes of Operation

### 1. Built-in Mock (default)

No configuration needed. FakeTestRail starts automatically:

```bash
go test -tags smoke ./pkg/snap_smoke/ -v
```

Benefits:
- No external dependencies
- Fast (~0.03 seconds)
- CI/CD friendly

### 2. Real Server

To verify compatibility with a real TestRail instance, set environment variables:

```bash
export GOTR_SMOKE_URL=http://localhost:8080
export GOTR_SMOKE_USER=admin@example.com
export GOTR_SMOKE_KEY=yourkey
export GOTR_SMOKE_PROJECT=3
export GOTR_SMOKE_SUITE=1
export GOTR_SMOKE_INSECURE=true   # for self-signed certificates

go test -tags smoke ./pkg/snap_smoke/ -v
```

| Variable | Required | Description |
| -------- | -------- | ----------- |
| `GOTR_SMOKE_URL` | Yes | TestRail base URL |
| `GOTR_SMOKE_USER` | Yes | Username / email |
| `GOTR_SMOKE_KEY` | Yes | API key |
| `GOTR_SMOKE_PROJECT` | Yes | Project ID |
| `GOTR_SMOKE_SUITE` | No¹ | Suite ID (for multi-suite projects) |
| `GOTR_SMOKE_INSECURE` | No | Skip TLS verification (`true`/`1`) |

¹ Required for projects in multi-suite mode.

## Test Scenarios

| # | Test | Operation | What it verifies |
| - | ---- | --------- | ---------------- |
| 1 | `TestSmoke_UpdateRollback` | update → rollback | Mutate title+priority → rollback → original restored |
| 2 | `TestSmoke_DeleteRollback` | delete → rollback | Delete → rollback → re-create with new ID (Tier 2) |
| 3 | `TestSmoke_AddRollback` | add → rollback | Create → FinalizeAdd → rollback → case deleted |
| 4 | `TestSmoke_SnapManagementCycle` | list/info/delete | Custom name → list → info → snapshot deletion |
| 5 | `TestSmoke_DoubleRollbackBlocked` | double rollback | Second rollback → error "rolled_back" |
| 6 | `TestSmoke_GCOrphans` | gc | Orphan cleaned, tracked snapshot preserved |

## Extending

### Adding a New Test

1. Add a `TestSmoke_*` function to `smoke_test.go`
2. Use helpers `testCase()` and `testSection()` — they register auto-cleanup
3. Isolate snap storage via `t.Setenv("HOME", t.TempDir())`

### Extending FakeTestRail

To support new endpoints:

1. Add storage (map) to `FakeTestRail`
2. Add a handler method `handleXxx(w, r, params)` — params are extracted from regex groups and query templates
3. Register action → handler in the `handlers` map inside `NewFakeTestRail()`
4. The route is picked up automatically from `testrailapi.APIPath` when the action name matches
5. For list endpoints, use `writePaginatedJSON()` for correct paginated wrapper format

### Build Tag

All files have the `//go:build smoke` tag. Tests are excluded from regular `go test ./...`.

## CLI Smoke Tests (cmd/snap)

Besides the E2E tests in `pkg/snap_smoke`, there are CLI-level smoke tests in `cmd/snap/smoke_test.go`.
They verify the correctness of `gotr snap *` Cobra commands without a real API — using a mock client
and an isolated snap store.

### Running

```bash
go test -tags smoke ./cmd/snap/ -v
```

### Test Scenarios

| # | Test | What it verifies |
| - | ---- | ---------------- |
| 1 | `TestCLI_SnapList_Empty` | Empty list when no snapshots exist |
| 2 | `TestCLI_SnapList_WithEntries` | Table output with data |
| 3 | `TestCLI_SnapInfo` | JSON metadata output |
| 4 | `TestCLI_SnapInfo_NotFound` | Error handling for missing snapshot |
| 5 | `TestCLI_SnapRollback_Update` | Rollback of an update operation |
| 6 | `TestCLI_SnapRollback_Delete` | Rollback of delete with re-create (Tier 2) |
| 7 | `TestCLI_SnapRollback_Add` | Rollback of an add operation |
| 8 | `TestCLI_SnapRollback_NotFound` | Error handling for missing snapshot |
| 9 | `TestCLI_SnapRollback_AlreadyRolledBack` | Protection against double rollback |
| 10 | `TestCLI_SnapDelete` | Snapshot deletion via CLI |
| 11 | `TestCLI_SnapGC_NoOrphans` | GC when no orphans exist |
| 12 | `TestCLI_SnapGC_CleansOrphans` | GC removes orphan directories |
| 13 | `TestCLI_FullCycle_ListRollbackList` | Full cycle list → rollback → list with status check |

### Differences from pkg/snap_smoke

| Aspect | `pkg/snap_smoke` | `cmd/snap/smoke_test.go` |
| ------ | ---------------- | ------------------------ |
| Level | Engine-level (API package) | CLI-level (Cobra commands) |
| Server | FakeTestRail (httptest) | Mock client (no HTTP) |
| Focus | Correctness of snap/rollback logic | Correctness of CLI wrappers |
| Speed | ~0.03 s | ~0.01 s |
| When to use | Verify snap engine | Verify snap commands |

### Full Run of Both Levels

```bash
go test -tags smoke ./pkg/snap_smoke/ ./cmd/snap/ -v
```

---

← [Guides](index.md)
