# Instruction: Exporting Data (export)

Language: [Русский](../../../ru/guides/instructions/crud-export.md) | English

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

The `gotr export` command — universal data export from TestRail to file.
Supports **30+ resource types** and 5 output formats.

> [!TIP]
> `export` is a safe read operation. Suitable for backups,
> pre-migration recon, and data preparation for analysis.

## Available Resources

```text
all, cases, casefields, casetypes, configurations, projects, priorities,
runs, tests, suites, sections, statuses, milestones, plans, results,
resultfields, reports, attachments, users, roles, templates, groups,
sharedsteps, variables, labels, datasets, bdds
```

## Examples 🚀

### Exporting Projects and Structure

```bash
# All projects
gotr export projects --save

# All suites of a project
gotr export suites -p 30 --save

# All sections of a suite
gotr export sections -p 30 -s 20069 --save
```

### Exporting Cases

```bash
# All cases of a suite in JSON
gotr export cases -p 30 -s 20069 --save --format json

# Cases of a specific section
gotr export cases -p 30 -s 20069 --section-id 12345 --save

# Cases in CSV for spreadsheet analysis
gotr export cases -p 30 -s 20069 --format csv --save
```

### Exporting Shared Steps

```bash
# All shared steps of a project
gotr export sharedsteps -p 30 --save --format json

# Table view for quick review
gotr export sharedsteps -p 30
```

### Exporting Test Runs and Results

```bash
# All runs of a project
gotr export runs -p 30 --save

# Runs for a specific milestone
gotr export runs -p 30 --milestone-id 5 --save

# All results
gotr export results -p 30 --save
```

### Bulk Export

```bash
# Export ALL supported resources of a project
gotr export all -p 30 --save
```

## Output Formats 🧩

| Format | Flag | Use case |
| --- | --- | --- |
| Table | `--format table` | Terminal viewing (default) |
| JSON | `--format json` | Analysis, scripts, storage |
| CSV | `--format csv` | Import to spreadsheets (Excel, Sheets) |
| Markdown | `--format md` | Documentation |
| HTML | `--format html` | Browser reports |

## Saving

Files are saved to `~/.gotr/exports/export/`:

```bash
# Automatic file naming
gotr export cases -p 30 -s 20069 --save

# Quiet mode (file only, no terminal output)
gotr export cases -p 30 -s 20069 --save --quiet
```

## Syntax 🧩

```bash
gotr export <resource> <endpoint> [id] [flags]
```

## Flags ⚙️

| Flag | Description | Default |
| --- | --- | --- |
| `-p, --project-id` | Project ID | — |
| `-s, --suite-id` | Suite ID (for cases) | — |
| `--section-id` | Section ID (for cases) | — |
| `--milestone-id` | Milestone ID (for runs) | — |
| `--save` | Save to file | `false` |
| `-f, --format` | Format: table, json, csv, md, html | `table` |
| `-q, --quiet` | Suppress service output | `false` |

## Practical Scenario: Pre-migration Preparation 🧩

```bash
# 1. Export shared steps for analysis
gotr export sharedsteps -p 30 --save --format json

# 2. Export cases for review
gotr export cases -p 30 -s 20069 --save --format json

# 3. Analyze data locally
cat ~/.gotr/exports/export/*.json | jq '.[] | .title'

# 4. If everything looks good — proceed to migration
# See Full Migration or Shared Steps Migration
```

## FAQ ❓

- ❓ **Question:** Where are files saved?
  > ↪️ **Answer:** in `~/.gotr/exports/export/`. File name is auto-generated from resource and timestamp.
  >
  > ---

- ❓ **Question:** How to export data from multiple projects?
  > ↪️ **Answer:** run `export` for each project with `-p <id>`. Or use `gotr export all` without `-p` for global resources (users, roles, templates).
  >
  > ---

- ❓ **Question:** Can exported JSON be used for import?
  > ↪️ **Answer:** for import use `gotr sync`, not raw JSON. Export is for analysis and archiving.

---

← [Instructions](index.md)
