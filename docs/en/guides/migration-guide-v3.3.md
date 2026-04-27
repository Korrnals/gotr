# Migration Guide v3.2 → v3.3

Language: [Русский](../../ru/guides/migration-guide-v3.3.md) | English

## Navigation

- [Documentation](../index.md)
  - [Guides](index.md)
  - [Architecture](../architecture/index.md)
  - [Operations](../operations/index.md)
  - [Reports](../reports/index.md)
- [Home](../../../README.md)

## TL;DR

```bash
gotr report organize --dry-run   && gotr report organize
gotr export organize --dry-run   && gotr export organize
# (optional) retention/cleanup policies
gotr cleanup all --dry-run
```

No existing config is broken; nothing is deleted automatically.

## What changed

### 1. `~/.gotr/reports/` hierarchy

Before (v3.2):

```
~/.gotr/reports/
├── migration-20260418_...md
├── coverage_p34_...json
├── rollback_snap_rel_9946.json
├── testrail_get_plan_234_20260420T101530Z.json
└── INDEX.md
```

After (v3.3):

```
~/.gotr/reports/
├── migrations/rel_9946/2026-04/migration-20260418_...md
├── coverage/rel_9946/2026-04/coverage_p34_...json
├── rollbacks/rel_9946/2026-04/rollback_snap_rel_9946.json
├── testrail/p234/2026-04/testrail_get_plan_234_20260420T101530Z.json
├── _unclassified/2026-04/<misc>
└── INDEX.md
```

- Categorisation is performed in `internal/report.ClassifyReport`
  based on the file name.
- For `testrail_*_<YYYYMMDD>.json` (date only, no `T<HHMMSS>Z`)
  `YearMonth` stays empty and the file is placed under
  `testrail/p<N>/<file>` (no month subdir). This is **expected**.

What to do:

```bash
gotr report organize --dry-run
gotr report organize
```

- The command **does not delete** anything.
- On conflict (the target file already exists) the entry is skipped,
  the `Skipped` counter is incremented; the source remains at the root.
- Re-running is a no-op.

### 2. `~/.gotr/exports/` hierarchy

Before (v3.2):

```
~/.gotr/exports/
├── snap_rel_9946_20260418.tar.gz
├── reports_20260418.zip
├── plans/
│   └── project_30_plans_20260401.json
└── reports/<legacy>
```

After (v3.3):

```
~/.gotr/exports/
├── snaps/snap_rel_9946_20260418.tar.gz
├── reports/reports_20260418.zip
└── api/plans/project_30_plans_20260401.json
```

- `snaps/` — tar.gz snap bundles,
- `reports/` — zip bundles and plain report files,
- `api/<resource>/` — legacy "raw" dumps from `gotr get ... --save`.

Migration:

```bash
gotr export organize --dry-run
gotr export organize
```

### 3. TLS — from `insecure` to `ca_bundle`

Before:

```yaml
insecure: true
```

After (recommended):

```yaml
tls:
  insecure: false
  ca_bundle: "/etc/ssl/corp-ca.pem"
```

- The legacy top-level `insecure` and the `--insecure` flag still work
  (OR-merge with `tls.insecure`).
- `tls.ca_bundle` loads the PEM into `tls.Config.RootCAs` — safer than
  `insecure=true`, and works behind corporate MITM proxies.

### 4. Warnings — from "all or nothing" to per-key

Before (not the final API): `no_warnings: true`.

After:

```yaml
ui:
  suppress_warnings:
    - tls_insecure
    - flat_layout
```

- Keys: `tls_insecure`, `deprecation`, `flat_layout`.
- `--show-warnings` is a CLI flag to temporarily show all warnings.
- On the first emission of any warning a tip is printed to stderr:
  "add '<key>' to ui.suppress_warnings to silence this warning".
- The "shown about flat_layout" flag is persistent —
  `~/.gotr/state.json::flat_layout_warned`. This means you will see
  the hint **exactly once** per installation.

### 5. Retention and cleanup

Retention is **disabled** by default. No old artifacts are deleted
automatically after the upgrade.

To enable:

```yaml
retention:
  reports:
    enabled: true
    max_age_days: 90
    max_count: 500
    keep_categories: [coverage]
    dry_run: false
```

And apply manually:

```bash
gotr cleanup reports --dry-run
gotr cleanup all
```

See [gotr cleanup](commands/cleanup.md).

### 6. `gotr export snap --with-reports`

A new flag, **enabled by default**. When a snap is exported into an
archive, related reports from `~/.gotr/reports/` are added
automatically (matching by snap basename as a substring of the report
basename).

To disable:

```bash
gotr export snap <snap-id> --no-reports
```

Embedded reports are visible in `manifest.Files` and are unpacked by
`gotr import snap` into the same categorised hierarchy.

## Upgrade checklist

- [ ] `go install github.com/Korrnals/gotr@v3.3.0` (or rebuild from source).
- [ ] `gotr report organize --dry-run` → `gotr report organize`.
- [ ] `gotr export organize --dry-run` → `gotr export organize`.
- [ ] Migrated `insecure: true` → `tls.ca_bundle: /etc/ssl/corp-ca.pem` (if applicable).
- [ ] (optional) Added `ui.suppress_warnings: [tls_insecure]` if the
  banner is noisy in CI.
- [ ] (optional) Configured `retention.*` and ran
  `gotr cleanup all --dry-run`.

## Rollback

If something goes wrong:

- The old layout can be restored manually: move files from the
  subdirectories back into the root of `~/.gotr/reports/`. There is no
  magic — `organize` works at the filesystem level.
- `gotr config view` masks keys, but the YAML itself is plain; you can
  put `insecure: true` back at any time.
- Installing the previous binary version is safe: the
  `~/.gotr/snaps/` and `~/.gotr/state.json` structures are
  backward-compatible.

---

← [Guides](index.md) · [Documentation](../index.md)
