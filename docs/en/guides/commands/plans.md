# Command: plans

Language: [Русский](../../../ru/guides/commands/plans.md) | English

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


## Overview 🎯
Manage test plans: create, update, close, delete, and manage entries.
A test plan is a collection of test runs (entries) united by a common goal.

> [!TIP]
> For a quick `plans` workflow: inspect `--help`, then run a
> safe/validation command before operational execution.

## Syntax 🧩
```bash
gotr plans [command]
```

## Subcommands

| Subcommand | Description |
| --- | --- |
| `add` | Create a new test plan |
| `close` | Close a test plan |
| `delete` | Delete a test plan |
| `entry` | Manage plan entries |
| `get` | Get a test plan by ID |
| `list` | List test plans |
| `update` | Update a test plan |

## Flags ⚙️

```text
-h, --help   help for plans
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
gotr plans --help
```

✅ **Why this matters:** keeps execution aligned with the current CLI version and avoids stale command assumptions.

---

### ▶️ Scenario 2: Targeted action validation
🎯 **Goal:** validate the exact subcommand/shape for the operation you need.

```bash
gotr plans add --help
```

✅ **Why this matters:** prevents wrong endpoint selection and reduces trial-and-error in production pipelines.

---

### ▶️ Scenario 3: Safe or machine-readable run
🎯 **Goal:** get deterministic output for analysis and automation.

```bash
gotr plans
```

✅ **Why this matters:** enables safer checks and structured post-processing in CI/CD.

---

### ▶️ Scenario 4: Mini operational pipeline
🎯 **Goal:** demonstrate a practical flow: validate -> run -> persist artifact.

```bash
gotr plans add --help && gotr plans add
```

✅ **Why this matters:** provides a reusable template for runbooks and scripted operations.

---

## ⚡ Quick Start (30 seconds)

1. Validate syntax and available flags quickly:
```bash
gotr plans --help
```
2. Choose the operational execution path:
```bash
gotr plans add --help
```
3. Execute safe/operational run:
```bash
gotr plans add --help
```

---

## 🧪 Pre-run Checklist

- [ ] URL, credentials, and TestRail access are validated.
- [ ] Project/suite/case identifiers are confirmed.
- [ ] A safe/diagnostic run was executed (`--help`, `--dry-run`, `--json`, or `--save`).
- [ ] Output format and artifact storage location are defined.

---

## 🎯 When To Use

- Use `plans` when the task belongs to this command domain and you need predictable repeatable behavior.
- Use it when you want a clear flow from syntax validation to operational execution.

---

## 🚫 When Not To Use

- Do not run directly if target IDs/endpoints are uncertain: validate with `--help` and a safe check first.
- Do not force this command for bulk operations outside its domain: pick a more specialized command/subcommand.

---

## FAQ ❓

- ❓ **Question:** When should I use `plans`?
  > ↪️ **Answer:** use it when your task belongs to this command domain and you want predictable resource-focused behavior.
  >
  > ---

- ❓ **Question:** Where should I start if parameters are unclear?
  > ↪️ **Answer:** always start with `gotr plans --help`, then inspect the target subcommand help before execution.
  >
  > ---

- ❓ **Question:** Which subcommands should be validated first?
  > ↪️ **Answer:** recommended starting set: add, close, delete, entry, get, list, update. Begin with the highest-frequency operation in your release workflow.
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

- Command exits with code `0` and confirms operation application on target resource.
- Resource state in TestRail matches provided input after execution.
- Follow-up `get/list` on target ID reflects the expected change.

### Practical result check

- Execute `get/list` for the same resource after mutation and verify changed fields explicitly.


---

## 🔎 Result Verification via Neighbor Commands

- Run a neighboring verification step via `get/list` for the same resource.

```bash
# primary operation
gotr <command> ...

# verification
gotr get <resource> <id>  # or gotr list <resource>
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
