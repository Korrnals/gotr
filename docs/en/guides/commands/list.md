# Command: list

Language: [Русский](../../../ru/guides/commands/list.md) | English

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
    - [Instructions](../instructions/index.md)
  - [Architecture](../../architecture/index.md)
  - [Operations](../../operations/index.md)
  - [Reports](../../reports/index.md)
- [Home](../../../../README.md)


## Overview 🎯
List available TestRail API v2 endpoints for the specified resource.
Examples:

> [!TIP]
> For a quick `list` workflow: inspect `--help`, then run a
> safe/validation command before operational execution.

## Syntax 🧩
```bash
gotr list <resource> [flags]
```

## Flags ⚙️

```text
-h, --help   help for list
--json       Output as JSON
--short      Short output (URI only)
```

## Global Flags 🌐

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

## Examples 🚀

### ▶️ Scenario 1: Capability discovery
🎯 **Goal:** inspect valid syntax and available flags before running operational actions.

```bash
gotr list --help
```

✅ **Why this matters:** keeps execution aligned with the current CLI version and avoids stale command assumptions.

---

### ▶️ Scenario 2: Targeted action validation
🎯 **Goal:** validate the exact subcommand/shape for the operation you need.

```bash
gotr list <resource>
```

✅ **Why this matters:** prevents wrong endpoint selection and reduces trial-and-error in production pipelines.

---

### ▶️ Scenario 3: Safe or machine-readable run
🎯 **Goal:** get deterministic output for analysis and automation.

```bash
gotr list --json
```

✅ **Why this matters:** enables safer checks and structured post-processing in CI/CD.

---

### ▶️ Scenario 4: Mini operational pipeline
🎯 **Goal:** demonstrate a practical flow: validate -> run -> persist artifact.

```bash
gotr list <resource>
```

✅ **Why this matters:** provides a reusable template for runbooks and scripted operations.

---

## ⚡ Quick Start (30 seconds)

1. Validate syntax and available flags quickly:
```bash
gotr list --help
```
2. Choose the operational execution path:
```bash
gotr list <resource>
```
3. Execute safe/operational run:
```bash
gotr list --json
```

---

## 🧪 Pre-run Checklist

- [ ] URL, credentials, and TestRail access are validated.
- [ ] Project/suite/case identifiers are confirmed.
- [ ] A safe/diagnostic run was executed (`--help`, `--dry-run`, `--json`, or `--save`).
- [ ] Output format and artifact storage location are defined.

---

## 🎯 When To Use

- Use `list` when the task belongs to this command domain and you need predictable repeatable behavior.
- Use it when you want a clear flow from syntax validation to operational execution.

---

## 🚫 When Not To Use

- Do not run directly if target IDs/endpoints are uncertain: validate with `--help` and a safe check first.
- Do not force this command for bulk operations outside its domain: pick a more specialized command/subcommand.

---

## FAQ ❓

- ❓ **Question:** When should I use `list`?
  > ↪️ **Answer:** use it when your task belongs to this command domain and you want predictable resource-focused behavior.
  >
  > ---

- ❓ **Question:** Where should I start if parameters are unclear?
  > ↪️ **Answer:** always start with `gotr list --help`, then inspect the target subcommand help before execution.
  >
  > ---

- ❓ **Question:** Which subcommands should be validated first?
  > ↪️ **Answer:** recommended starting set: no subcommands. Begin with the highest-frequency operation in your release workflow.
  >
  > ---

- ❓ **Question:** How do I run safely in production-like environments?
  > ↪️ **Answer:** follow a staged approach: syntax validation, constrained trial run, then final execution with saved artifacts.
  >
  > ---

- ❓ **Question:** How do I integrate this command into CI/CD?
  > ↪️ **Answer:** use stable parameter sets, machine-readable output where available, and explicit exit-code checks.

---

## 🧾 Expected Execution Result

### Success criteria

- Command exits with code `0` and no terminal diagnostics errors.
- Requested resource/compare data is returned in the selected format.
- Dataset scope matches provided filters and identifiers.

### Output shape example (`--json`)

```json
{
  "status": "ok",
  "data": [
    { "...": "..." }
  ]
}
```

Note: this is a structural example, not a literal API snapshot.


---

## 🔎 Result Verification via Neighbor Commands

- After mutations, verify aggregated list includes/excludes target entity as expected.

```bash
gotr list <resource>
```


---

## Best Practices 🧭

- ✅ **Practice: Keep reusable command templates**
  > Store proven command variants for project/suite/case identifiers in your internal runbook to reduce manual mistakes.
  >
  > ---

- ✅ **Practice: Log execution context**
  > Capture key parameters (IDs, URL, selected flags, timestamp) before execution to simplify incident analysis.
  >
  > ---

- ✅ **Practice: Separate diagnostic and operational runs**
  > Use help/safe checks first, then run production actions. This significantly lowers risk of unintended TestRail changes.

---

## Common Pitfalls and Diagnostics 🛠️

- ⚠️ **Pitfall: Command succeeds but output is not what you expected**
  > Validate target IDs and subcommand selection; mismatched endpoint/arguments are the most frequent cause.
  >
  > ---

- ⚠️ **Pitfall: Automation fails intermittently**
  > Ensure required parameters are always provided and interactive input expectations are disabled in CI contexts.
  >
  > ---

- ⚠️ **Pitfall: Hard to compare outcomes between runs**
  > Persist artifacts to files and keep output format consistent for repeatable diff/analysis.

## Source of Truth

- Sections above are generated from actual CLI `--help` output from current code.

---

← [Commands](index.md) · [Guides](../index.md) · [Documentation](../../index.md)
