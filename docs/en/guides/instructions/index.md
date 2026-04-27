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
      - [Interactive Migration Walkthrough](migration-interactive-walkthrough.md)
      - [Full Migration](migration-full.md)
      - [Partial Migration](migration-partial.md)
      - [Shared Steps Migration](migration-shared-steps.md)
      - [Resources Migration](migration-resources.md)
      - [⚠️ Live migration test plan (before going live)](live-migration-test-plan.md)
      - [Live-run operator card](live-migration-operator-card.md)
      - [Getting Data](crud-get.md)
      - [Exporting Data](crud-export.md)
      - [Creating Objects](crud-add.md)
      - [Updating Objects](crud-update.md)
      - [Deleting Objects](crud-delete.md)
      - [Comparing Projects](compare.md)
      - [Reports lifecycle](reports-lifecycle.md)
      - [Retention & cleanup runbook](retention-and-cleanup-runbook.md)
      - [TLS: insecure → ca_bundle](tls-ca-bundle-migration.md)
  - [Architecture](../../architecture/index.md)
  - [Operations](../../operations/index.md)
  - [Reports](../../reports/index.md)
- [Home](../../../../README.md)

## Contents

Practical step-by-step instructions for common gotr tasks.
Each instruction is a ready-to-use recipe: prerequisites, commands, result verification.

### Data Migration

Transferring data between TestRail projects via `gotr sync`.

- [Interactive Migration Walkthrough](migration-interactive-walkthrough.md) — every migration variant via interactive mode (walkthrough)
- [Full Migration](migration-full.md) — shared steps + cases in one pass (`sync full`)
- [Partial Migration](migration-partial.md) — cases with mapping from a previous step
- [Shared Steps Migration](migration-shared-steps.md) — transfer only shared test steps
- [Resources Migration](migration-resources.md) — suites, sections between projects
- [⚠️ Live migration test plan](live-migration-test-plan.md) — **mandatory before a real run**: an isolated test on two test projects with rollback and cleanup
- [Live-run operator card](live-migration-operator-card.md) — short checklist and commands for executing the test in a terminal

### CRUD Operations

Day-to-day work with TestRail objects.

- [Getting Data](crud-get.md) — `gotr get` for projects, cases, shared steps, etc.
- [Exporting Data](crud-export.md) — `gotr export` to JSON/CSV/HTML with file saving
- [Creating Objects](crud-add.md) — `gotr add` with interactive mode and dry-run
- [Updating Objects](crud-update.md) — `gotr update` entity fields
- [Deleting Objects](crud-delete.md) — `gotr delete` with soft and hard removal

### Comparison

- [Comparing Projects](compare.md) — `gotr compare` for auditing and pre-migration recon

### v3.3.0 — UX polish

Step-by-step recipes for the new functionality (report hierarchy,
retention/cleanup, corporate TLS). See also
[Migration guide v3.3](../migration-guide-v3.3.md).

- [Reports lifecycle](reports-lifecycle.md) — layout migration → generation → viewing (`--print`, `latest`) → snap export with `--with-reports` → import into a clean environment → cleanup
- [Retention & cleanup runbook](retention-and-cleanup-runbook.md) — base config with the `coverage` whitelist, CI dry-run scenario, compatibility with `gotr snap gc`
- [TLS: insecure → ca_bundle](tls-ca-bundle-migration.md) — corporate CA migration, suppressing the `tls_insecure` banner, troubleshooting

---

← [Guides](../index.md)
