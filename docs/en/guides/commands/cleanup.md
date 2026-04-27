# Command: cleanup

Language: [Русский](../../../ru/guides/commands/cleanup.md) | English

## Navigation

- [Documentation](../../index.md)
  - [Guides](../index.md)
    - [Installation](../installation.md)
    - [Configuration](../configuration.md)
    - [Interactive Mode](../interactive-mode.md)
    - [Progress](../progress.md)
    - [Commands Index](index.md)
      - [General](global-flags.md)
      - [CRUD Operations](add.md)
      - [Core Resources](get.md)
      - [Special Resources](bdds.md)
      - [cleanup](cleanup.md)
    - [Instructions](../instructions/index.md)
  - [Architecture](../../architecture/index.md)
  - [Operations](../../operations/index.md)
  - [Reports](../../reports/index.md)
- [Home](../../../../README.md)

## Overview 🎯

`gotr cleanup` is a manual executor of retention policies over the local
store at `~/.gotr/`. By default nothing is deleted automatically: the
policy is described in the config (the `retention.*` section) and only
applied on explicit user command.

> [!IMPORTANT]
> Always use `--dry-run` before the first real run and review the plan.

## Syntax 🧩

```bash
gotr cleanup {reports | exports | snaps | all} [--dry-run]
```

## Subcommands

| Subcommand | What it cleans | Policy source |
| --- | --- | --- |
| `reports` | `~/.gotr/reports/**` (category hierarchy) | `retention.reports` |
| `exports` | `~/.gotr/exports/{snaps,reports,api}/**` | `retention.exports` |
| `snaps` | `~/.gotr/snaps/**` (delegates to `gotr snap gc`) | `retention.snaps` |
| `all` | sequentially `reports` → `exports` → `snaps` | all three sections |

## Flags ⚙️

```text
--dry-run    only show the plan, do not delete anything (overrides retention.*.dry_run)
-h, --help   help
```

## Policy configuration

```yaml
retention:
  reports:
    enabled: true
    max_age_days: 90
    max_count: 500
    keep_categories: [coverage]   # category whitelist
    dry_run: false
  exports:
    enabled: true
    max_age_days: 30
  snaps:
    enabled: true
    max_age_days: 180
    max_count: 100
```

- `enabled: false` → the command exits explicitly with a "retention disabled"
  message.
- `keep_categories` is honoured only for `reports` (coverage artifacts are
  typically protected via the whitelist).
- Inside `reports`, deletion proceeds per category: first an age filter,
  then a trim down to `max_count` newest entries. Before exit `INDEX.md`
  is regenerated.

## Examples 🚀

### Scenario 1: Preview the report cleanup plan

```bash
gotr cleanup reports --dry-run
```

Shows the list of candidate files per category, without deletion.

### Scenario 2: Full cleanup

```bash
gotr cleanup all --dry-run
gotr cleanup all
```

### Scenario 3: Exports only

```bash
gotr cleanup exports
```

## Interaction with `gotr snap gc`

`gotr cleanup snaps` is a thin wrapper around the existing `gotr snap gc`.
The old workflow continues to work unchanged; the new command exists for
the unified `cleanup all` UX and to read policy from `retention.snaps`.

## See also

- [Configuration → Retention](../configuration.md)
- [Architecture: UX polish v3.3.0](../../architecture/ux-polish-v3.3.0.md)

---

← [Commands Index](index.md) · [Documentation](../../index.md)
