# Command: get

Language: [Русский](../../../ru/guides/commands/get.md) | English

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
Perform GET requests to the TestRail API.
Subcommands:

## Syntax
```bash
gotr get [command]
```

## Subcommands

| Subcommand | Description |
| --- | --- |
| `case` | Get a single test case by ID |
| `case-fields` | Get list of case fields |
| `case-history` | Get change history of a case by ID |
| `case-types` | Get list of case types |
| `cases` | Get project test cases |
| `project` | Get a single project by ID |
| `projects` | Get all projects |
| `sharedstep` | Get a single shared step by ID |
| `sharedsteps` | Get project shared steps |
| `suite` | Get a single test suite by ID |
| `suites` | Get project test suites |

## Flags

```text
-h, --help   help for get
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
gotr get --help
gotr get case --help
```

## Source of Truth

- Sections above are generated from actual CLI `--help` output from current code.

---

← [Commands](index.md) · [Guides](../index.md) · [Documentation](../../index.md)
