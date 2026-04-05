# Completion

Language: [Русский](../../../ru/guides/commands/completion.md) | English

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

The `gotr completion` command generates shell completion scripts for interactive command input.

## What it does

- Handles API operations for the `completion` command scope.
- Provides deterministic CLI behavior for scripts and CI/CD pipelines.
- Helps reduce manual work by standardizing repetitive workflows.

## When to use

- When you need a predictable CLI flow for automation.
- When you want to minimize manual steps and human error.
- When the operation must run the same way locally and in CI/CD.

## Subcommands

- `gotr completion bash`
- `gotr completion zsh`
- `gotr completion fish`
- `gotr completion powershell`

## Examples

```bash
# Bash
source <(gotr completion bash)

# Zsh
gotr completion zsh > "${fpath[1]}/_gotr"

# Fish
gotr completion fish > ~/.config/fish/completions/gotr.fish
```

## Useful flags

- `--json` for machine-readable output.
- `--output` / `--save` to persist results to files.
- `--verbose` for detailed execution diagnostics.

---

← [Commands](index.md) · [Guides](../index.md) · [Documentation](../../index.md)
