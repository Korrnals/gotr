# Reference: Global Flags

Language: [Русский](../../../ru/guides/commands/global-flags.md) | English

## Navigation

- [Documentation](../../index.md)
  - [Guides](../index.md)
    - [Installation](../installation.md)
    - [Configuration](../configuration.md)
    - [Interactive Mode](../interactive-mode.md)
    - [Progress](../progress.md)
    - [Commands Index](index.md)
      - [General](global-flags.md)
        - [global-flags](global-flags.md)
        - [config](config.md)
        - [completion](completion.md)
        - [self-test](self-test.md)
      - [CRUD Operations](add.md)
      - [Core Resources](get.md)
      - [Special Resources](bdds.md)
  - [Architecture](../../architecture/index.md)
  - [Operations](../../operations/index.md)
  - [Reports](../../reports/index.md)
- [Home](../../../../README.md)


## Overview
gotr — a convenient utility for working with TestRail API v2.
Supports browsing available endpoints, executing requests, and more.

## Syntax
```bash
gotr [command]
```

## Subcommands

| Subcommand | Description |
| --- | --- |
| `add` | Create a new resource (POST request) |
| `attachments` | Manage file attachments |
| `bdds` | Manage BDD scenarios |
| `cases` | Manage test cases |
| `compare` | Compare data between projects |
| `completion` | Generate completion script |
| `config` | Manage gotr configuration |
| `datasets` | Manage datasets (test data) |
| `delete` | Delete a resource (DELETE/POST request) |
| `export` | Export data from TestRail to JSON file |
| `get` | GET requests to TestRail API |
| `groups` | Manage user groups |
| `labels` | Manage test labels |
| `list` | List available TestRail API endpoints by resource |
| `milestones` | Manage project milestones |
| `plans` | Manage test plans |
| `reports` | Manage project reports |
| `result` | Manage test results in TestRail |
| `roles` | Manage user roles |
| `run` | Manage test runs in TestRail |
| `self-test` | Run self-diagnostic tests |
| `sync` | Sync TestRail data between projects |
| `templates` | Manage test case templates |
| `test` | Manage tests in TestRail |
| `tests` | Manage tests |
| `update` | Update an existing resource (POST request) |
| `users` | Manage TestRail users |
| `variables` | Manage test case variables |

## Flags

```text
-k, --api-key string    TestRail API key
-c, --config            Create default configuration file
-f, --format string     Output format: table, json, csv, md, html (default "table")
-h, --help              help for gotr
--insecure              Skip TLS certificate verification
--non-interactive       Disable interactive prompts; exit with error if input is required
-q, --quiet             Suppress output (progress, stats, save messages)
--url string            TestRail base URL
-u, --username string   TestRail user email
-v, --version           version for gotr
```

## Examples

```bash
gotr --help
gotr list --help
```

## Source of Truth

- Sections above are generated from actual CLI `--help` output from current code.

---

← [Commands](index.md) · [Guides](../index.md) · [Documentation](../../index.md)
