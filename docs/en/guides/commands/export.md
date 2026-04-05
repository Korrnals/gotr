# Export

Language: [Русский](../../../ru/guides/commands/export.md) | English

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

The `gotr export` command exports data to JSON and file artifacts.

## What it does

- Handles API operations for the `export` command scope.
- Provides deterministic CLI behavior for scripts and CI/CD pipelines.
- Helps reduce manual work by standardizing repetitive workflows.

## When to use

- When you need a predictable CLI flow for automation.
- When you want to minimize manual steps and human error.
- When the operation must run the same way locally and in CI/CD.

## Examples

```bash
# Export cases
gotr export cases 30 --suite-id 20069 --output cases.json

# Export to directory
gotr export run 12345 --save

# Filtered export
gotr export cases 30 --section-id 900 --output section_cases.json
```

## Useful flags

- `--json` for machine-readable output.
- `--output` / `--save` to persist results to files.
- `--verbose` for detailed execution diagnostics.

---

← [Команды](index.md) · [Гайды](../index.md) · [Документация](../../index.md)
