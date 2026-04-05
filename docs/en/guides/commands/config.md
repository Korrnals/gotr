# Command: config

Language: [Русский](../../../ru/guides/commands/config.md) | English

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
Commands for managing the gotr configuration file.

## Syntax
```bash
gotr config [command]
```

## Subcommands

| Subcommand | Description |
| --- | --- |
| `edit` | Open config file in the default editor |
| `init` | Create default configuration file |
| `path` | Show path to current config file |
| `view` | Show contents of current config file |

## Flags

```text
-h, --help   help for config
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
gotr config --help
gotr config edit --help
```

## Source of Truth

- Sections above are generated from actual CLI `--help` output from current code.

---

← [Commands](index.md) · [Guides](../index.md) · [Documentation](../../index.md)
