# Command: snap

Language: [Русский](../../../ru/guides/commands/snap.md) | English

## Navigation

- [Documentation](../../index.md)
  - [Guides](../index.md)
    - [Installation](../installation.md)
    - [Configuration](../configuration.md)
    - [Interactive Mode](../interactive-mode.md)
    - [Progress](../progress.md)
    - [Commands Index](index.md)
      - [General](global-flags.md)
        - [global-flags](global-flags.md)
        - [config](config.md)
        - [completion](completion.md)
        - [self-test](self-test.md)
        - [snap](snap.md)
      - [CRUD Operations](add.md)
      - [Core Resources](get.md)
      - [Special Resources](bdds.md)
    - [Instructions](../instructions/index.md)
  - [Architecture](../../architecture/index.md)
  - [Operations](../../operations/index.md)
  - [Reports](../../reports/index.md)
- [Home](../../../../README.md)


## Overview 🎯

Manage pre-mutation snapshots: list, inspect, rollback mutations, export, and clean up.

Snapshots are automatically created before mutating operations (`update`, `delete`, `add`, etc.)
when `snap.enabled = true` in config or the `--snapshot` flag is set.

> [!TIP]
> Quick start: run `gotr snap list` after any mutating operation
> to see available snapshots.

## Syntax 🧩

```bash
gotr snap <subcommand> [args] [flags]
```

## Subcommands 📋

| Subcommand | Description |
| ---------- | ----------- |
| `list` | Snapshot table grouped by server; interactive two-level picker |
| `info [id]` | Formatted metadata card (JSON via `--format json`) |
| `rollback [id]` | Rollback a mutation using saved data |
| `export [id]` | Export snapshot to a portable JSON file (interactive path prompt) |
| `delete [id]` | Delete snapshot from disk and manifest |
| `gc` | Clean up orphaned snapshots (on disk but not in manifest) |

> All subcommands that accept `[id]` support interactive selection:
> if `id` is omitted, an interactive picker with available snapshots is shown.

## Subcommand Flags ⚙️

### rollback

```text
--dry-run              Preview changes without applying (diff table)
--entity-ids string    Limit rollback to specific entity IDs (comma-separated)
```

### export

Second positional argument is the output file path (default: `snapshot_<id>.json`).
In interactive mode, if path is omitted, prompts for filename and directory.

## Global Flags 🌐

```text
-k, --api-key string    TestRail API key
-c, --config            Create default configuration file
-f, --format string     Output format: table, json, csv, md, html (default "table")
--insecure              Skip TLS certificate verification
--non-interactive       Disable interactive prompts; exit with error if input is required
-q, --quiet             Suppress output (progress, stats, save messages)
--url string            TestRail base URL
-u, --username string   TestRail user email
```

## Rollback Tiers

| Tier | Operation | Behavior |
| ---- | --------- | -------- |
| Tier 1 | `update` | Full rollback — restores original field values |
| Tier 2 | `delete` | Re-creates entity with a new ID (original ID is lost) |
| Tier 2 | `add` | Deletes the created entity |
| Tier 3 | `result`, `labels` | Info-only snapshot (rollback not supported) |

## Examples 🚀

### ▶️ Scenario 1: List snapshots

🎯 **Goal:** view all available snapshots.

```bash
gotr snap list
```

In interactive mode, performs two-level selection: first server, then snapshot.
In non-interactive mode or with `--format`, displays a table with a SERVER column:

```text
 #  ID                              SERVER                    OP      ENTITY  CATEGORY  STATUS     TIMESTAMP
 1  cases/1718000000_update_42      https://my.testrail.io    update  case    cases     available  2025-06-10 12:00:00
 2  custom/my_backup                https://my.testrail.io    delete  suite   custom    available  2025-06-10 12:05:00
```

For JSON output (scripts, pipelines):

```bash
gotr snap list --format json
```

---

### ▶️ Scenario 2: Snapshot details

🎯 **Goal:** inspect full metadata.

```bash
gotr snap info cases/1718000000_update_42
```

Displays a formatted card:

```text
┌──────────── Snapshot Info ────────────┐
│ ID         │ cases/1718000000_update_42 │
│ Server     │ https://my.testrail.io     │
│ Operation  │ update case                │
│ Tier       │ T1 (full rollback)         │
│ Status     │ available                  │
│ Entity IDs │ 42                         │
│ ...        │ ...                        │
└────────────┴────────────────────────────┘
```

For JSON output (scripts):

```bash
gotr snap info cases/1718000000_update_42 --format json
```

---

### ▶️ Scenario 3: Dry-run rollback preview

🎯 **Goal:** see the diff before applying.

```bash
gotr snap rollback cases/1718000000_update_42 --dry-run
```

Example output (with server context):

```text
Server:    https://my.testrail.io
Snapshot:  cases/1718000000_update_42 (update case, T1)

The following changes will be applied:

ENTITY ID  FIELD     CURRENT           SNAPSHOT
42         title     Changed Title     Original Title
42         priority  3                 2
```

> When rolling back a `delete` and the target section is already deleted on the server (404/400),
> the rollback gracefully skips re-creation and continues.

---

### ▶️ Scenario 4: Execute rollback

🎯 **Goal:** rollback a mutation.

```bash
gotr snap rollback cases/1718000000_update_42
```

In interactive mode, shows a diff table first and asks for confirmation.

---

### ▶️ Scenario 5: Partial rollback by entity-ids

🎯 **Goal:** rollback only specific entities from a batch operation.

```bash
gotr snap rollback sync/1718000000_sync_cases --entity-ids 42,43,44
```

---

### ▶️ Scenario 6: Export and cleanup

🎯 **Goal:** save a snapshot as an artifact and clean up garbage.

```bash
# Export
gotr snap export cases/1718000000_update_42 backup.json

# Delete a specific snapshot
gotr snap delete cases/1718000000_update_42

# Clean up orphaned snapshots
gotr snap gc
```

## ⚡ Quick Start (30 seconds)

1. Run a mutating operation with snap enabled:
```bash
gotr update case 42 --json '{"title":"test"}' --snapshot
```
2. List snapshots:
```bash
gotr snap list
```
3. Rollback if needed:
```bash
gotr snap rollback <snapshot_id>
```

## Configuration

Snap can be enabled globally in `~/.gotr.yaml`:

```yaml
snap:
  enabled: true
```

Or via the `--snapshot` flag for individual operations.

Snapshot storage: `~/.gotr/snaps/`.

## 🧪 Pre-run Checklist

- [ ] gotr is configured (`gotr self-test`)
- [ ] Snap is enabled (`snap.enabled: true` or `--snapshot`)
- [ ] For rollback: TestRail API is reachable (update/delete rollbacks make API calls)
- [ ] For dry-run: API access required to fetch current state

## FAQ ❓

- ❓ **Question:** What happens when rolling back a delete?
  > ↪️ **Answer:** The entity is re-created via API with a new ID (Tier 2). The original ID cannot be restored — this is a TestRail API limitation.

- ❓ **Question:** Can I rollback twice?
  > ↪️ **Answer:** No. After rollback, the snapshot status changes to `rolled_back` and further rollbacks are blocked.

- ❓ **Question:** What does gc do?
  > ↪️ **Answer:** Removes snapshot directories on disk that are not tracked in the manifest (orphans). Tracked snapshots are not affected.

- ❓ **Question:** Where is data stored?
  > ↪️ **Answer:** In `~/.gotr/snaps/` — organized by category (`cases/`, `sync/`, `custom/`, etc.). Manifest: `~/.gotr/snaps/manifest.json`.

## 🛠️ Troubleshooting

| Problem | Solution |
| ------- | -------- |
| "snapshot not found" | Check the ID via `gotr snap list` |
| "already rolled_back" | Snapshot already used; create a new one |
| Tier 3 rollback fails | Tier 3 operations (result, labels) do not support rollback |
| Empty list | Ensure `snap.enabled: true` or use `--snapshot` |

---

← [Commands Index](index.md) · [self-test](self-test.md)
