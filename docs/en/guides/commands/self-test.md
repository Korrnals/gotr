# Self Test

Language: [Русский](../../../ru/guides/commands/self-test.md) | English

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

The `gotr self-test` command performs local environment and API availability checks.

## What it does

- Handles API operations for the `self-test` command scope.
- Provides deterministic CLI behavior for scripts and CI/CD pipelines.
- Helps reduce manual work by standardizing repetitive workflows.

## When to use

- When you need a predictable CLI flow for automation.
- When you want to minimize manual steps and human error.
- When the operation must run the same way locally and in CI/CD.

## Examples

```bash
# Basic check
gotr self-test

# Verbose output
gotr self-test --verbose

# Check with alternative config
gotr self-test --config ~/.gotr/config/default.yaml
```

## Useful flags

- `--json` for machine-readable output.
- `--output` / `--save` to persist results to files.
- `--verbose` for detailed execution diagnostics.

---

← [Команды](index.md) · [Гайды](../index.md) · [Документация](../../index.md)
