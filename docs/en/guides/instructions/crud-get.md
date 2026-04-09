# Instruction: Getting Data (get)

Language: [Русский](../../../ru/guides/instructions/crud-get.md) | English

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

The `gotr get` command is the primary tool for **reading data** from TestRail.
Supports fetching lists and individual objects for all resource types.

> [!TIP]
> All `get` commands are safe — they only read data and don't change anything in TestRail.

## Available Resources

| Resource | Syntax | Description |
| --- | --- | --- |
| `projects` | `gotr get projects` | All projects |
| `project` | `gotr get project <id>` | Single project |
| `suites` | `gotr get suites <project_id>` | Project suites |
| `suite` | `gotr get suite <id>` | Single suite |
| `cases` | `gotr get cases <project_id>` | Project cases |
| `case` | `gotr get case <id>` | Single case |
| `case-fields` | `gotr get case-fields` | Case fields |
| `case-types` | `gotr get case-types` | Case types |
| `case-history` | `gotr get case-history <id>` | Case history |
| `sharedsteps` | `gotr get sharedsteps <project_id>` | Shared steps |
| `sharedstep` | `gotr get sharedstep <id>` | Single shared step |
| `sharedstep-history` | `gotr get sharedstep-history <id>` | Shared step history |
| `runs` | `gotr get runs <project_id>` | Project test runs |
| `run` | `gotr get run <id>` | Single run |
| `tests` | `gotr get tests <run_id>` | Tests in a run |
| `test` | `gotr get test <id>` | Single test |
| `results` | `gotr get results <test_id>` | Test results |
| `plans` | `gotr get plans <project_id>` | Test plans |
| `plan` | `gotr get plan <id>` | Single plan |
| `milestones` | `gotr get milestones <project_id>` | Project milestones |
| `users` | `gotr get users` | All users |
| `user` | `gotr get user <id>` | Single user |

## Examples 🚀

### Projects and Structure

```bash
# All projects
gotr get projects

# Specific project
gotr get project 30

# Suites of a project
gotr get suites 30

# Specific suite
gotr get suite 20069
```

### Cases

```bash
# All cases in a project (requires --suite-id for multi-suite projects)
gotr get cases 30 --suite-id 20069

# Specific case
gotr get case 12345

# Case change history
gotr get case-history 12345
```

### Shared Steps

```bash
# All shared steps of a project
gotr get sharedsteps 30

# Specific shared step
gotr get sharedstep 456

# Shared step history
gotr get sharedstep-history 456
```

### Test Runs and Results

```bash
# All runs of a project
gotr get runs 30

# Tests in a specific run
gotr get tests 789

# Results of a specific test
gotr get results 101
```

## Output Format 🧩

```bash
# Table (default)
gotr get projects

# JSON
gotr get projects --format json

# CSV
gotr get projects --format csv

# Markdown
gotr get projects --format md

# HTML
gotr get projects --format html
```

## Interactive Mode 🧩

Running `get` without an ID activates interactive selection:

```bash
# Interactive project selection, then suites
gotr get suites

# Interactive case selection
gotr get case
```

## Saving Results

```bash
# Save to ~/.gotr/exports/
gotr get sharedsteps 30 --save

# JSON with saving
gotr get cases 30 --suite-id 20069 --format json --save
```

## FAQ ❓

- ❓ **Question:** How does `get` differ from `export`?
  > ↪️ **Answer:** `get` fetches data through specific GET endpoints with typed output. `export` is more universal, works with any resource, and is oriented towards file saving. For quick viewing use `get`, for saving archives — `export`.
  >
  > ---

- ❓ **Question:** How to get cases from a specific section?
  > ↪️ **Answer:** use `gotr export cases -p <project_id> -s <suite_id> --section-id <id>`.

---

← [Instructions](index.md)
