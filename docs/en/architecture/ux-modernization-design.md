# UX Modernization Design — gotr Interactive Commands

Language: English | [Русский](../../ru/architecture/ux-modernization-design.md)

## Goal

Bring all gotr interactive commands to a unified style, as implemented in `cmd/snap/`:

- Multi-level navigation with Back/Exit
- Aligned columns with headers
- Grouping by categories
- Detail cards + action menus
- JSON priority on save (interactive → table, file save → JSON)

## Reference: cmd/snap/

- 4 levels: Server → Operation → Resource → Snapshot
- `alignedPickerLabels()` — auto-width columns + header
- `errGoBack`/`errExit` sentinels — Back/Exit navigation
- `renderInfoCard()` — detail card
- `postCardAction()` — action menu after card
- `groupByOperation()`, `groupByCategory()` — nested grouping

## Current State by Groups

### Tier 1 — Maximum Impact (large data, frequent use)

| Group | Current Pattern | What's Needed |
|-------|----------------|---------------|
| cmd/cases/ | Project→Suite→Case, flat `[i] ID: X \| Title` | Aligned labels, Back/Exit, info card, grouping by section |
| cmd/get/ | Project→Suite→Item, flat | Aligned labels, Back/Exit, info card |
| cmd/compare/ | Complex flow | Aligned labels, navigation |
| cmd/sync/ | Select+Confirm chains | Back/Exit, aligned labels, preview cards, **snapshot integration** |

### Tier 2 — Medium Impact (moderate lists)

| Group | Current Pattern | What's Needed |
|-------|----------------|---------------|
| cmd/attachments/ | 3-level deep hierarchy | Back/Exit, aligned, cards |
| cmd/test/, cmd/tests/ | Project→Run→Test | Aligned labels, Back/Exit |
| cmd/run/ | Project→Suite→Run | Aligned labels, grouped by status |
| cmd/plans/ | Project→Plan→Entry | Aligned, cards |
| cmd/configurations/ | Group→Config | Aligned, Back/Exit |

### Tier 3 — Minimal (short lists, rarely used)

| Group | Current Pattern | What's Needed |
|-------|----------------|---------------|
| cmd/list/ | Flat resource select | Minimal — already ok |
| cmd/delete/ | Endpoint→Item | Back/Exit, confirm card |
| cmd/add/, cmd/update/ | Input fields | Minimal changes |
| cmd/export/ | Resource→Endpoint→ID | Back/Exit |
| cmd/groups/ | Project→Group | Aligned |
| cmd/templates/ | Project select | Minimal |
| cmd/variables/, cmd/datasets/ | Dataset→Variable | Aligned |
| cmd/labels/, cmd/bdds/ | Flat select | Minimal |
| cmd/users/, cmd/roles/ | Flat select | Minimal |
| cmd/reports/ | Template select | Minimal |
| cmd/milestones/ | Project→Milestone | Aligned |
| cmd/result/ | Run→Test→Case | Back/Exit, aligned |

## Implementation Principles

### 1. Shared Navigation Kit (internal/interactive/)

Extract from `cmd/snap/interactive_helpers.go` into `internal/interactive/`:

- `AlignedLabels(columns []Column, rows []Row) []string` — generic aligned formatter
- `BrowseLoop(cfg BrowseConfig) error` — generic browse with Back/Exit/action
- `GroupBy(entries, keyFn) []Group` — generic grouper
- Navigation sentinels: `ErrGoBack`, `ErrExit`
- `PostAction` enum + `ActionMenu(options)` pattern

### 2. Aligned Label Format

```
[1] ID: 123 │ cases    │ Regression Test Login │ active  │ 2026-04-13
[2] ID: 456 │ cases    │ Payment Flow          │ closed  │ 2026-04-12
```

- Auto-width columns (as in snap)
- Optional header row
- `│` separators

### 3. Info Cards

```
┌─ Case Info ──────────────────┐
│ ID         │ 12345           │
│ Title      │ Login Flow      │
│ Section    │ Auth / Login    │
│ Priority   │ Critical        │
│ Status     │ Active          │
└──────────────────────────────┘
```

- go-pretty table StyleRounded
- Displayed when an item is selected

### 4. Action Menu (after card)

```
? Action:
  ← Back
  ✕ Exit
  ↻ Refresh
  📋 Copy ID
  💾 Save to file (JSON)
```

- Context-specific actions depend on the command
- Rollback/Undo/Delete — only where applicable

### 5. JSON Output Priority

- **Interactive display** → table (human-readable)
- **--format json** → JSON to stdout
- **--save / file save** → JSON by default (not table!)
- **Pipe detection**: if stdout is not a TTY → JSON automatically

### 6. Grouping Strategy

- Projects → grouping by Suite
- Cases → grouping by Section
- Runs → grouping by Status (active/completed)
- Tests → grouping by Run
- Configs → grouping by Config Group
- Attachments → grouping by Entity Type

## Sync — Snapshot Integration (PRIORITY)

### Context

Migration (`gotr sync **`) is the riskiest operation. All safety tools
(snapshot, rollback, undo) must be accessible directly from the migration flow,
rather than forcing the user to switch between commands.

### 1. Mandatory Snapshot Prompt Before Migration

Each `gotr sync *` subcommand asks before execution:

```
? Create snapshot before migration? (recommended) [Y/n]
```

- Default: **true** (Yes) — snapshot is always created unless the user declines
- With `--non-interactive`: snapshot is created automatically (no prompt)
- With `--no-snapshot`: explicitly disabled (flag)
- Snapshot is saved to the manifest, ID is printed after creation:
  `✓ Snapshot created: sync/20260414T120000_sync_cases_0`

### 2. Post-Migration Navigation Menu

After migration completes (success or partial success) — action menu:

```
? Migration complete. What next?
  ← Back to sync menu
  ✕ Exit
  📋 View migration log
  ↻ Rollback this migration
  📦 Browse rollback history
```

- `↻ Rollback this migration` — invokes `snap rollback <snapshot_id>` for the just-created snapshot
- `📦 Browse rollback history` — opens `snap rollback list` (browse rolled-back snapshots)
- `📋 View migration log` — shows diff/summary of executed changes

### 3. Sync Subcommands Scope

All `gotr sync` subcommands receive snapshot integration:

- `gotr sync cases` — sync cases between suites
- `gotr sync suites` — sync suites between projects
- `gotr sync sections` — sync sections between suites
- `gotr sync shared-steps` — sync shared steps
- `gotr sync full` — full project sync (all of the above)

### 4. Flags

- `--snapshot` / `--no-snapshot` — explicit snapshot control (default: true)
- `--rollback-on-error` — automatic rollback on migration error (future)

## Implementation Order

### Phase 0: Sync Snapshot Integration (priority)

1. Add `--snapshot` / `--no-snapshot` flag to all sync subcommands
2. Pre-migration snapshot prompt + auto-create
3. Post-migration action menu (rollback/browse/log)
4. Integration with existing snap engine
5. Tests

### Phase 1: Shared Navigation Kit

6. Extract navigation primitives from cmd/snap/ → internal/interactive/
7. Generic `BrowseLoop`, `AlignedLabels`, `GroupBy`
8. Tests for kit

### Phase 2: Tier 1 Commands (cases, get, compare, sync)

9. cmd/get/ — aligned labels + Back/Exit + info cards
10. cmd/cases/ — grouped by section + aligned + cards
11. cmd/sync/ — navigation + preview cards (on top of Phase 0)
12. cmd/compare/ — aligned output

### Phase 3: Tier 2 Commands

13. cmd/attachments/, cmd/test/, cmd/run/
14. cmd/plans/, cmd/configurations/
15. cmd/result/

### Phase 4: Tier 3 Commands + Polish

16. Remaining commands — minimal aligned labels
17. Pipe detection → auto-JSON
18. Final consistency pass

## Constraints

- Do NOT break the existing non-interactive (`--non-interactive`) contract
- Do NOT change the CLI API (flags, args) — only interactive UX
- JSON schema does not change — only the presentation layer
- Backward-compatible: existing scripts continue to work
