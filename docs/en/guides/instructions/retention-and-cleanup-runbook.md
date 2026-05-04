# Instruction: retention & cleanup runbook

Language: [Русский](../../../ru/guides/instructions/retention-and-cleanup-runbook.md) | English

## Navigation

- [Documentation](../../index.md)
  - [Guides](../index.md)
    - [Instructions](index.md)
      - [Retention & cleanup runbook](retention-and-cleanup-runbook.md)
  - [Architecture](../../architecture/index.md)
  - [Operations](../../operations/index.md)
  - [Reports](../../reports/index.md)
- [Home](../../../../README.md)

Step-by-step scenarios for configuring and applying retention policies
in v3.3.0. By default retention is disabled — nothing is deleted
automatically.

## Prerequisites

- v3.3.0+, the reports/exports hierarchy is already migrated to the new
  layout (see [Reports lifecycle](reports-lifecycle.md), case A).

## Case A: Base config with coverage protection

**Scenario:** keep the last 500 reports, no older than 90 days; never
touch coverage artefacts.

```yaml
# ~/.gotr/config/default.yaml
retention:
  reports:
    enabled: true
    max_age_days: 90
    max_count: 500
    keep_categories: [coverage]
    dry_run: false
  exports:
    enabled: true
    max_age_days: 30
  snaps:
    enabled: true
    max_age_days: 180
    max_count: 100
```

```bash
# Always — dry-run first
gotr cleanup all --dry-run

# Real run
gotr cleanup all
```

## Case B: Reports only, leaving snap/exports alone

```bash
gotr cleanup reports --dry-run
gotr cleanup reports
```

`keep_categories: [coverage, migrations]` will also protect migration
reports.

## Case C: Temporarily disable retention in an emergency

```yaml
retention:
  reports:
    enabled: false
```

`gotr cleanup reports` will explicitly exit with a message
"retention for reports disabled" and exit code 0.

## Case D: CI — throttle local storage after migration

```bash
# In a CI script after a guaranteed successful migration
gotr cleanup reports --dry-run > cleanup-plan.txt
# Manual review of the artefact; if it looks good, the next run drops --dry-run
```

The CLI `--dry-run` flag takes precedence over `retention.*.dry_run` in
the YAML.

## Case E: "snap gc only" via the unified CLI

Historically there was `gotr snap gc`. In v3.3 `gotr cleanup snaps` is a
compatible wrapper that reads `retention.snaps`. The legacy path still
works:

```bash
gotr snap gc                   # as before
gotr cleanup snaps             # via retention.snaps
```

## Case F: Bulk cleanup of TestRail attachments

**Scenario:** reclaim space on the TestRail side by deleting attachments
older than N days, with a snapshot safety net for rollback. See
[`gotr attachments cleanup`](../commands/attachments.md) for the full
flag reference.

Five-step runbook:

1. **Dry-run preview.** Always start with `--dry-run` to inspect the
   pre-flight summary (project scope, entity types, total attachments,
   estimated size).

   ```bash
   gotr attachments cleanup --all-projects \
     --entity-type result --older-than 90d --dry-run
   ```

2. **Review the pre-flight summary.** Verify project IDs, entity types,
   and the projected count/size. Re-run the dry-run with narrower
   `--project` / `--entity-type` / `--limit` if the scope is too broad.

3. **Confirm and execute.** Drop `--dry-run`. The command will create a
   snapshot under category `cleanup-attachments` before deleting.

   ```bash
   gotr attachments cleanup --all-projects \
     --entity-type result --older-than 90d --concurrency 4
   ```

   Use `--force` to skip the final confirmation prompt in scripts.

4. **Locate the snapshot.** Snapshots live under
   `~/.gotr/snaps/cleanup-attachments/<id>/` (`data.json` + `files/`
   tree). Default TTL for this category is **7 days** — see
   [snap → Per-category retention TTL](../commands/snap.md#per-category-retention-ttl)
   to extend it via `snap.retention.category_ttl_days` if you need a
   longer rollback window.

   ```bash
   gotr snap list --category cleanup-attachments
   ```

5. **Rollback if needed.** Re-uploads the snapshotted blobs to TestRail.
   Re-uploaded attachments receive **new** TestRail IDs (original IDs
   cannot be restored). The `test` entity type is gracefully skipped on
   rollback because TestRail has no `add_attachment_to_test` endpoint.

   ```bash
   gotr snap rollback <snapshot_id>
   ```

> **Tip.** Use `--no-snapshot` only when you intentionally accept that
> rollback will be impossible (e.g. one-off cleanup with no recovery
> requirement). The default snapshot path is the recommended mode.

## Success criteria

- `INDEX.md` is updated after `cleanup reports`.
- Files in `keep_categories` did not shrink.
- `~/.gotr/exports/snaps/` contains only files newer than
  `max_age_days` or within `max_count`.
- `gotr cleanup all --dry-run` returns exit 0 and a plan on stdout.

## Important warnings

> **Always dry-run.** The very first run of a new policy must be
> **only** with `--dry-run`. Retention looks at mtime, not logical age;
> a freshly imported old archive can unexpectedly fall into the
> deletion plan under `max_age_days`.

> **`keep_categories` applies only to `reports`.** For `snaps` and
> `exports` use `max_count` / `max_age_days` or keep them disabled.

## See also

- [gotr cleanup](../commands/cleanup.md).
- [Configuration → Retention](../configuration.md).

---

← [Instructions](index.md) · [Documentation](../../index.md)
