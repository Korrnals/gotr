# Command: bdds

Language: [Русский](../../../ru/guides/commands/bdds.md) | English

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
        - [bdds](bdds.md)
        - [configurations](configurations.md)
        - [datasets](datasets.md)
        - [groups](groups.md)
        - [labels](labels.md)
        - [milestones](milestones.md)
        - [roles](roles.md)
        - [templates](templates.md)
        - [users](users.md)
        - [variables](variables.md)
  - [Architecture](../../architecture/index.md)
  - [Operations](../../operations/index.md)
  - [Reports](../../reports/index.md)
- [Home](../../../../README.md)


## Overview
Manage BDD (Behavior Driven Development) scenarios.
BDD scenarios describe system behavior in the Given-When-Then format

## Syntax
```bash
gotr bdds [command]
```

## Subcommands

| Subcommand | Description |
| --- | --- |
| `add` | Add a BDD scenario to a test case |
| `get` | Get the BDD scenario for a test case |

## Flags

```text
-h, --help   help for bdds
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
gotr bdds --help
gotr bdds add --help
```

## Source of Truth

- Sections above are generated from actual CLI `--help` output from current code.

---

← [Commands](index.md) · [Guides](../index.md) · [Documentation](../../index.md)
