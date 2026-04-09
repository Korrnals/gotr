# Instruction: Comparing Projects (compare)

Language: [Русский](../../../ru/guides/instructions/compare.md) | English

## Navigation

- [Documentation](../../index.md)
  - [Guides](../index.md)
    - [Installation](../installation.md)
    - [Configuration](../configuration.md)
    - [Interactive Mode](../interactive-mode.md)
    - [Progress](../progress.md)
    - [Commands Index](../commands/index.md)
    - [Instructions](index.md)
      - [Full Migration](migration-full.md)
      - [Partial Migration](migration-partial.md)
      - [Shared Steps Migration](migration-shared-steps.md)
      - [Resources Migration](migration-resources.md)
      - [Getting Data](crud-get.md)
      - [Exporting Data](crud-export.md)
      - [Creating Objects](crud-add.md)
      - [Updating Objects](crud-update.md)
      - [Deleting Objects](crud-delete.md)
      - [Comparing Projects](compare.md)
  - [Architecture](../../architecture/index.md)
  - [Operations](../../operations/index.md)
  - [Reports](../../reports/index.md)
- [Home](../../../../README.md)

## Overview 🎯

The `gotr compare` command — compares resources between two TestRail projects.
Used for **pre-migration auditing**, **post-migration verification**, and **drift monitoring**.

> [!TIP]
> `compare` is a safe read operation. Supports parallel data loading
> and saving results in JSON/CSV/HTML.

## Available Resources

| Subcommand | Description |
| --- | --- |
| `all` | Compare ALL supported resources |
| `cases` | Test cases |
| `suites` | Test suites |
| `sections` | Sections |
| `sharedsteps` | Shared steps |
| `runs` | Test runs |
| `plans` | Test plans |
| `milestones` | Milestones |
| `configurations` | Configurations |
| `datasets` | Datasets |
| `groups` | Groups |
| `labels` | Labels |
| `templates` | Templates |

## Examples 🚀

### Full Comparison of Two Projects

```bash
# Compare all resources
gotr compare all --pid1 30 --pid2 34 --save

# Save to a specific file
gotr compare all --pid1 30 --pid2 34 --save-to comparison-report.json
```

### Comparing Specific Resources

```bash
# Shared steps
gotr compare sharedsteps --pid1 30 --pid2 34

# Cases
gotr compare cases --pid1 30 --pid2 34

# Suites
gotr compare suites --pid1 30 --pid2 34

# Sections
gotr compare sections --pid1 30 --pid2 34
```

### Saving in Different Formats

```bash
# JSON (for scripts and analysis)
gotr compare all --pid1 30 --pid2 34 --format json --save

# HTML (for reports)
gotr compare all --pid1 30 --pid2 34 --format html --save

# CSV (for spreadsheets)
gotr compare cases --pid1 30 --pid2 34 --format csv --save
```

### Performance Tuning

```bash
# Increase parallelism for large projects
gotr compare all --pid1 30 --pid2 34 \
  --parallel-suites 15 \
  --parallel-pages 10 \
  --timeout 60m
```

## Syntax 🧩

```bash
gotr compare <resource> --pid1 <ID> --pid2 <ID> [flags]
```

## Flags ⚙️

| Flag | Description | Default |
| --- | --- | --- |
| `-1, --pid1` | First project ID | required |
| `-2, --pid2` | Second project ID | required |
| `--save` | Save to `~/.gotr/exports/` | `false` |
| `--save-to` | Save to specific file | — |
| `-f, --format` | Format: json, csv, md, html | `table` |
| `--parallel-suites` | Suite parallelism | `10` |
| `--parallel-pages` | Page parallelism | `6` |
| `--rate-limit` | API rate limit (-1=auto, 0=unlimited) | `-1` |
| `--timeout` | Operation timeout | `30m` |
| `-q, --quiet` | Suppress service output | `false` |

## Scenario: Pre-migration Audit 🧩

```bash
# 1. Compare all resources for full picture
gotr compare all --pid1 30 --pid2 34 --save-to pre-migration-audit.json

# 2. Check shared steps separately
gotr compare sharedsteps --pid1 30 --pid2 34

# 3. Check cases
gotr compare cases --pid1 30 --pid2 34

# 4. Based on results — decide which migration type is needed
# → Full migration: migration-full.md
# → Shared steps only: migration-shared-steps.md
```

## Scenario: Post-migration Verification 🧩

```bash
# 1. Compare shared steps — check all transferred
gotr compare sharedsteps --pid1 30 --pid2 34

# 2. Compare cases — check all cases in place
gotr compare cases --pid1 30 --pid2 34

# 3. Full report for documentation
gotr compare all --pid1 30 --pid2 34 \
  --format html --save-to post-migration-report.html
```

## FAQ ❓

- ❓ **Question:** What does the comparison show?
  > ↪️ **Answer:** object counts in each project, matches (by title/name), elements unique to each project, shared and differing items.
  >
  > ---

- ❓ **Question:** How to speed up comparison for large projects?
  > ↪️ **Answer:** increase `--parallel-suites` and `--parallel-pages`. For projects with 10000+ cases, `--timeout 60m` is recommended.
  >
  > ---

- ❓ **Question:** Can I compare a single specific resource?
  > ↪️ **Answer:** yes, use a specific subcommand: `gotr compare cases`, `gotr compare sharedsteps`, etc.
  >
  > ---

- ❓ **Question:** What about `retry-failed-pages`?
  > ↪️ **Answer:** if some pages failed to load during comparison (timeout, rate limit), use `gotr compare retry-failed-pages` for retry.

---

← [Instructions](index.md)
