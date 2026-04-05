# Get

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

The `gotr get` command reads resources and reference data from TestRail.

## What it does

- Handles API operations for the `get` command scope.
- Provides deterministic CLI behavior for scripts and CI/CD pipelines.
- Helps reduce manual work by standardizing repetitive workflows.

## When to use

- When you need a predictable CLI flow for automation.
- When you want to minimize manual steps and human error.
- When the operation must run the same way locally and in CI/CD.

## Examples

```bash
# All projects
gotr get projects

# List cases by project/suite
gotr get cases 30 --suite-id 20069

# Get run
gotr get run 12345
```

## Useful flags

- `--json` for machine-readable output.
- `--output` / `--save` to persist results to files.
- `--verbose` for detailed execution diagnostics.

---

← [Команды](index.md) · [Гайды](../index.md) · [Документация](../../index.md)
