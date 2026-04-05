# Command: compare

Language: [Русский](../../../ru/guides/commands/compare.md) | English

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
        - [get](get.md)
        - [sync](sync.md)
        - [compare](compare.md)
        - [cases](cases.md)
        - [run](run.md)
        - [result](result.md)
        - [test](test.md)
        - [tests](tests.md)
        - [attachments](attachments.md)
        - [plans](plans.md)
        - [reports](reports.md)
      - [Special Resources](bdds.md)
  - [Architecture](../../architecture/index.md)
  - [Operations](../../operations/index.md)
  - [Reports](../../reports/index.md)
- [Home](../../../../README.md)


## Overview
Compare resources between two TestRail projects.
Supported resources:

## Syntax
```bash
gotr compare [command]
```

## Subcommands

| Subcommand | Description |
| --- | --- |
| `all` | Compare all resources between two projects |
| `cases` | Compare test cases between projects |
| `configurations` | Compare configurations between projects |
| `datasets` | Compare datasets between projects |
| `groups` | Compare groups between projects |
| `labels` | Compare labels between projects |
| `milestones` | Compare milestones between projects |
| `plans` | Compare test plans between projects |
| `runs` | Compare test runs between projects |
| `sections` | Compare sections between projects |
| `sharedsteps` | Compare shared steps between projects |
| `suites` | Compare test suites between projects |
| `templates` | Compare templates between projects |

## Flags

```text
-h, --help               help for compare
--page-retries int       Number of retries per page in the main loading phase (default 5)
--parallel-pages int     Maximum number of parallel pages within a suite (default 6)
--parallel-suites int    Maximum number of parallel suites (default 10)
-1, --pid1 string        First project ID (required)
-2, --pid2 string        Second project ID (required)
--rate-limit int         API request limit per minute. -1 = auto by profile/deployment, 0 = no limit, >0 = fixed value. (default -1)
--retry-attempts int     Number of attempts for auto-retry of failed pages (default 5)
--retry-delay duration   Pause between retries for a single page during auto-retry (default 200ms)
--retry-workers int      Number of parallel workers during auto-retry of failed pages (default 12)
--save                   Save result to file (default: ~/.gotr/exports/)
--save-to string         Save result to specified file
--timeout duration       Timeout for compare operation (default 30m0s)
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
gotr compare --help
gotr compare all --help
```

## Source of Truth

- Sections above are generated from actual CLI `--help` output from current code.

---

← [Commands](index.md) · [Guides](../index.md) · [Documentation](../../index.md)
