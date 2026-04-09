# Instructions

Language: [Русский](../../../ru/guides/instructions/index.md) | English

## Navigation

- [Documentation](../../index.md)
  - [Guides](../index.md)
    - [Installation](../installation.md)
    - [Configuration](../configuration.md)
    - [Interactive Mode](../interactive-mode.md)
    - [Progress](../progress.md)
    - [Commands Index](../commands/index.md)
    - [Instructions](index.md)
      - [Full Migration](migration-full.md)
      - [Partial Migration](migration-partial.md)
      - [Shared Steps Migration](migration-shared-steps.md)
      - [Resources Migration](migration-resources.md)
      - [Getting Data](crud-get.md)
      - [Exporting Data](crud-export.md)
      - [Creating Objects](crud-add.md)
      - [Updating Objects](crud-update.md)
      - [Deleting Objects](crud-delete.md)
      - [Comparing Projects](compare.md)
  - [Architecture](../../architecture/index.md)
  - [Operations](../../operations/index.md)
  - [Reports](../../reports/index.md)
- [Home](../../../../README.md)

## Contents

Practical step-by-step instructions for common gotr tasks.
Each instruction is a ready-to-use recipe: prerequisites, commands, result verification.

### Data Migration

Transferring data between TestRail projects via `gotr sync`.

- [Full Migration](migration-full.md) — shared steps + cases in one pass (`sync full`)
- [Partial Migration](migration-partial.md) — cases with mapping from a previous step
- [Shared Steps Migration](migration-shared-steps.md) — transfer only shared test steps
- [Resources Migration](migration-resources.md) — suites, sections between projects

### CRUD Operations

Day-to-day work with TestRail objects.

- [Getting Data](crud-get.md) — `gotr get` for projects, cases, shared steps, etc.
- [Exporting Data](crud-export.md) — `gotr export` to JSON/CSV/HTML with file saving
- [Creating Objects](crud-add.md) — `gotr add` with interactive mode and dry-run
- [Updating Objects](crud-update.md) — `gotr update` entity fields
- [Deleting Objects](crud-delete.md) — `gotr delete` with soft and hard removal

### Comparison

- [Comparing Projects](compare.md) — `gotr compare` for auditing and pre-migration recon

---

← [Guides](../index.md)
