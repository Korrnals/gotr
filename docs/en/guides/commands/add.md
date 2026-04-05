# Command: add

Language: [Русский](../../../ru/guides/commands/add.md) | English

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
        - [add](add.md)
        - [delete](delete.md)
        - [update](update.md)
        - [list](list.md)
        - [export](export.md)
      - [Core Resources](get.md)
      - [Special Resources](bdds.md)
  - [Architecture](../../architecture/index.md)
  - [Operations](../../operations/index.md)
  - [Reports](../../reports/index.md)
- [Home](../../../../README.md)


## Overview
Creates a new object in TestRail via POST API.
Supported endpoints:

## Syntax
```bash
gotr add <endpoint> [id] [flags]
```

## Flags

```text
--announcement string   Announcement (for project)
--assignedto-id int     ID of the assigned user
--case-ids string       Comma-separated case IDs (for run)
--comment string        Comment (for result)
--defects string        Defects (for result)
--description string    Description/announcement
--dry-run               Show what would be executed without making changes
--elapsed string        Elapsed time (for result)
-h, --help              help for add
--include-all           Include all cases (for run) (default true)
-i, --interactive       Interactive mode (wizard)
--json-file string      Path to JSON file with data
--milestone-id int      Milestone ID
-n, --name string       Resource name
--priority-id int       Priority ID (for case)
--refs string           References
--save                  Save output to file in ~/.gotr/exports/
--section-id int        Section ID
--show-announcement     Show announcement
--status-id int         Status ID (for result)
--suite-id int          Suite ID
--template-id int       Template ID (for case)
--title string          Title (for case)
--type-id int           Type ID (for case)
```

## Global Flags

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

## Examples

```bash
gotr add --help
```

## Source of Truth

- Sections above are generated from actual CLI `--help` output from current code.

---

← [Commands](index.md) · [Guides](../index.md) · [Documentation](../../index.md)
