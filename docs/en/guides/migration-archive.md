# Migration archives — full-state cross-machine transfer

`gotr export migration-archive` and `gotr import migration-archive`
provide a single-file mechanism to move a snapshot, a hand-picked group
of snapshots, or **the entire `~/.gotr/` state** between machines. Use
it when you need to:

- hand off a migration result to another operator;
- mirror a workstation's `gotr` state to a CI runner or to a colleague;
- archive everything before reinstalling the system.

---

## TL;DR

```bash
# On source machine: pack the WHOLE local gotr state into one archive
gotr export migration-archive
# → ~/.gotr/exports/all/migration_bundle_all_<N>snaps_<UTC-ts>.tar.gz

# Copy the file to the target machine, then:
gotr import migration-archive ~/Downloads/migration_bundle_all_*.tar.gz
# Snaps, reports and logs are restored under ~/.gotr/, and the imported
# snaps are auto-registered in ~/.gotr/snaps/manifest.json — gotr snap list
# shows them immediately.
```

---

## Export modes

| Invocation | Pack scope | Archive kind | Default location |
|---|---|---|---|
| `gotr export migration-archive` | All snapshots + entire `reports/` + entire `logs/` | `migration_bundle` | `~/.gotr/exports/all/migration_bundle_all_<N>snaps_<ts>.tar.gz` |
| `gotr export migration-archive --all` | Same as above (explicit) | `migration_bundle` | `~/.gotr/exports/all/` |
| `gotr export migration-archive --label <substr>` | Snaps whose `label` contains substring + their union of reports/logs | `migration_bundle` | `~/.gotr/exports/all/migration_bundle_label-<slug>_<N>snaps_<ts>.tar.gz` |
| `gotr export migration-archive <id1> <id2> ...` | Listed snaps + their union of reports/logs | `migration_bundle` | `~/.gotr/exports/all/migration_bundle_multi_<N>snaps_<ts>.tar.gz` |
| `gotr export migration-archive <single-id>` | One snap + its reports + its ±1h log window | `snap` (single) | `~/.gotr/exports/snaps/migration_<id>_<ts>.tar.gz` |
| `gotr export migration-archive --out PATH` | Same as the chosen mode | — | Custom destination, regardless of mode |

`<substr>` matching for `--label` is **case-insensitive**.

### Archive layout

A `migration_bundle` archive contains, at the archive root:

```text
manifest.json     — kind, schema version, snap_ids, file checksums
SHA256SUMS        — for tooling-friendly verification
README.txt        — human-readable summary
snaps/<id>/...    — one directory per snapshot (full meta+data tree)
reports/...       — embedded reports (full ~/.gotr/reports/ tree in --all mode)
logs/...          — embedded logs (full ~/.gotr/logs/ tree in --all mode)
```

A single-snap archive (`Kind=snap`) is laid out the same way but with
exactly one `snaps/<id>/` directory and only the per-snap windowed
reports/logs.

---

## Import

### Path argument is mandatory

```bash
gotr import migration-archive <path-to-file>
```

The argument is the **source archive file**, not a destination — the
target is **always** the system directory `~/.gotr/` (or
`$GOTR_HOME` when set). There is no `--out`-style flag for import:
restoring elsewhere would defeat the purpose of cross-machine state
transfer.

In interactive mode (TTY, no `--non-interactive`) and with no
argument, the picker scans **both** `~/.gotr/exports/snaps/` and
`~/.gotr/exports/all/` and offers a unified list.

### Auto-detection

Import inspects `manifest.json` inside the archive and dispatches to the
correct path automatically:

- `Kind=snap` → restores one snapshot. `--rename-id <new-id>` is
  honoured for collision resolution.
- `Kind=migration_bundle` → restores every snapshot in `snap_ids`,
  plus embedded reports and logs. `--rename-id` is **not** supported
  in multi-snap mode (a 1→1 mapping is impossible).

### Manifest auto-registration (since v3.4.0-dev)

After files are relocated on disk, import automatically updates
`~/.gotr/snaps/manifest.json` so each imported snapshot becomes
visible to `gotr snap list` without requiring a follow-up
`gotr snap manifest repair`.

- Without `--overwrite`: new entries are appended.
- With `--overwrite`: any pre-existing entry for an imported ID is
  pruned first, then re-added — no duplicates.
- If a snap's `meta.json` is unreadable for any reason, the snap
  remains on disk but is not indexed. A subsequent
  `gotr snap manifest repair` will pick it up.

### Other flags

```text
--dry-run         Report what would be imported, do not modify disk
--overwrite       Replace existing snapshots (a backup is moved to <store>/.trash/)
--rename-id ID    Single-snap only: relocate the snapshot under a new ID
--skip-reports    Do not restore embedded reports/
--skip-logs       Do not restore embedded logs/
```

Existing files in `~/.gotr/reports/` and `~/.gotr/logs/` are **never
overwritten**; import logs them as `skipped` and continues.

---

## Where archives live

```text
~/.gotr/exports/
├── snaps/    — single-snap archives (snap_<id>_<ts>.tar.gz, migration_<id>_<ts>.tar.gz)
├── reports/  — exported reports (.zip / .pdf / .md / .json)
├── api/      — raw API dumps from `gotr export <resource> --save`
└── all/      — full or multi-snap migration_bundle_*.tar.gz (since v3.4.0-dev)
```

The `all/` directory is created on first use of the multi-snap export
mode. Old multi-snap archives that pre-date this change keep working;
import accepts any explicit path.

---

## Verifying a transfer

After import:

```bash
gotr snap list                 # Imported snaps appear immediately
gotr snap manifest repair --dry-run   # Should report 0 changes
ls ~/.gotr/reports/            # Restored reports tree
ls ~/.gotr/logs/               # Restored logs tree
```

For a stricter byte-level check, compare `SHA256SUMS` from the archive
against the on-disk extracted files.

---

## Common pitfalls

- **"snap already exists" on import.** Pass `--overwrite` (the
  existing snap is moved to `<store>/.trash/`) or `--rename-id` for
  single-snap archives.
- **Empty `gotr snap list` after import on older versions.** Auto-
  registration was added in v3.4.0-dev; on older builds run
  `gotr snap manifest repair` once.
- **Reports/logs not restored.** Check whether you exported with
  `--skip-reports` / `--skip-logs`, or the source machine had nothing
  matching the snap window. Use `--all` for unconditional inclusion.
- **Archive ended up in `exports/snaps/` even though it has many
  snaps.** Pre-v3.4.0-dev behaviour. The file is still importable from
  any path; new exports go to `exports/all/`.
