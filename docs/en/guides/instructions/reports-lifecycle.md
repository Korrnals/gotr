# Instruction: reports lifecycle in v3.3

Language: [Русский](../../../ru/guides/instructions/reports-lifecycle.md) | English

## Navigation

- [Documentation](../../index.md)
  - [Guides](../index.md)
    - [Instructions](index.md)
      - [Reports lifecycle](reports-lifecycle.md)
  - [Architecture](../../architecture/index.md)
  - [Operations](../../operations/index.md)
  - [Reports](../../reports/index.md)
- [Home](../../../../README.md)

This recipe covers every stage of working with local reports in v3.3.0:
migration → generation → listing → viewing → export with embedded
reports → import into a clean environment → policy-driven cleanup.

## Prerequisites

- `gotr` v3.3.0+ (`gotr --version`).
- A config at `~/.gotr/config/default.yaml` with valid `base_url`, `username`, `api_key`.
- At least one finished snap/migration run (so there is something to organise).

## Case A: Layout migration from v3.2 to v3.3 (one-time)

**Scenario:** upgraded from v3.2; `~/.gotr/reports/` contains a flat list of files.

```bash
# 1. View the migration plan without changes
gotr report organize --dry-run
# → prints a list of src → dst pairs

# 2. Apply
gotr report organize
# → Moved: N, Skipped: 0
# INDEX.md is regenerated automatically

# 3. Check the result
tree -L 3 ~/.gotr/reports | head -30
gotr report list
```

**Rollback.** If the new hierarchy is undesirable, move the files back to
the root manually
(`find ~/.gotr/reports -type f -exec mv -t ~/.gotr/reports {} +`) and
delete the empty subdirectories. Nothing is encrypted.

## Case B: First session on v3.3 (flat layout + hint)

The first `gotr report list` prints a **one-time** hint about the flat
layout. The flag is persisted in `~/.gotr/state.json`. Ways to remove it:

1. Run the migration (case A) — the hint stops appearing because there
   is no flat layout left.
2. Suppress it in the config without migrating:
   ```yaml
   ui:
     suppress_warnings: [flat_layout]
   ```

## Case C: View and print a report

```bash
# Interactive selection from the recursive hierarchy (on a TTY)
gotr report show

# By latest — the freshest file by mtime
gotr report show latest

# By basename — the matcher finds the file in any category
gotr report show migration-20260418_123456_rel_9946.md

# By relative path — for an exact pick
gotr report show migrations/rel_9946/2026-04/migration-20260418_123456_rel_9946.md

# Print contents to stdout (for CI/pipeline)
gotr report show latest --print > /tmp/report.md
gotr report show coverage_p34_... --print | jq '.summary'

# A PDF opens via the OS viewer; the exit code is propagated
gotr report show full-audit-2026-04-18.pdf
echo "viewer exit: $?"
```

**What is forbidden:** `gotr report show <file.pdf> --print` → an
explicit error about binary content.

## Case D: Export a snap together with its reports

**Scenario:** hand a snap to a colleague with every related artefact
(migration report, coverage, rollback-blueprint).

```bash
# By default --with-reports=true
gotr export snap rel_9946

# The archive lands in exports/snaps/
ls -lh ~/.gotr/exports/snaps/

# Inspect the contents
tar -tzf ~/.gotr/exports/snaps/snap_rel_9946_*.tar.gz | head -20
# → manifest.json, SHA256SUMS, snaps/rel_9946/*, reports/migrations/..., reports/coverage/...

# Disable embedding reports
gotr export snap rel_9946 --no-reports
```

Matching criterion: `filepath.Base(report)` contains
`filepath.Base(snapID)`. Example: snapID `rel_9946` pulls in
`migration-*_rel_9946.md`, `coverage_p34_rel_9946.json`, but not
`rollback_snap_rel_other.json`.

## Case E: Round-trip via import into another environment

```bash
# On the dst machine with a clean ~/.gotr/
gotr import snap /path/to/snap_rel_9946_20260420.tar.gz

# The snap lands in ~/.gotr/snaps/rel_9946/
# Reports are automatically laid out in the categorised hierarchy:
gotr report list | grep rel_9946
# → migrations/rel_9946/2026-04/migration-*.md
# → coverage/rel_9946/2026-04/coverage_*.json
```

The archive layout (`reports/<rel>`) matches the target hierarchy, so no
extra `organize` is required after `import`.

## Case F: Periodic cleanup

```yaml
# ~/.gotr/config/default.yaml
retention:
  reports:
    enabled: true
    max_age_days: 90
    max_count: 500
    keep_categories: [coverage]
```

```bash
# Always start with --dry-run
gotr cleanup reports --dry-run

# Apply
gotr cleanup reports

# Or, in one command — reports + exports + snaps
gotr cleanup all
```

The `coverage` category is whitelisted: retention never deletes it.
`INDEX.md` is regenerated after deletion.

## Success criteria

- `gotr report list` shows hierarchical paths, not flat names.
- `~/.gotr/exports/` contains `snaps/`, `reports/`, `api/` subdirectories.
- `tar -tzf <snap.tar.gz>` contains both `snaps/<id>/` and `reports/<rel>`.
- `~/.gotr/state.json::flat_layout_warned=true` (if you saw the hint).
- After `gotr cleanup reports` `INDEX.md` is updated and the deleted-count is > 0.

## See also

- [gotr report](../commands/report.md) — full command reference.
- [gotr cleanup](../commands/cleanup.md) — retention executor.
- [Migration guide v3.3](../migration-guide-v3.3.md).
- [Architecture: UX polish v3.3.0](../../architecture/ux-polish-v3.3.0.md).

---

← [Instructions](index.md) · [Documentation](../../index.md)
