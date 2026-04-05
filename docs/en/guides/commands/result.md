# Command: result

Language: [Русский](../../../ru/guides/commands/result.md) | English

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
Commands for adding and retrieving test results in TestRail.
A test result is the outcome of executing an individual test (passed, failed, blocked, etc.)

## Syntax
```bash
gotr result [flags]
gotr result [command]
```

## Subcommands

| Subcommand | Description |
| --- | --- |
| `add` | Add a result for a test |
| `add-bulk` | Bulk add results |
| `add-case` | Add a result for a case in a run |
| `fields` | Get list of result fields |
| `get` | Get results for a test |
| `get-case` | Get results for a case in a run |
| `list` | Get results for a test run |

## Flags

```text
-h, --help   help for result
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
gotr result --help
gotr result add --help
```

## Source of Truth

- Sections above are generated from actual CLI `--help` output from current code.

---

← [Commands](index.md) · [Guides](../index.md) · [Documentation](../../index.md)
