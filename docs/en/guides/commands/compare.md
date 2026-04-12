# Command: compare

Language: [Русский](../../../ru/guides/commands/compare.md) | English

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
    - [Instructions](../instructions/index.md)
  - [Architecture](../../architecture/index.md)
  - [Operations](../../operations/index.md)
  - [Reports](../../reports/index.md)
- [Home](../../../../README.md)


## Overview 🎯
Compare resources between two TestRail projects.
Supported resources:

> [!TIP]
> For a quick `compare` workflow: inspect `--help`, then run a
> safe/validation command before operational execution.

## Syntax 🧩
```bash
gotr compare [command]
```

## Subcommands

| Subcommand | Description |
| --- | --- |
| `all` | Compare all resources between two projects |
| `cases` | Compare test cases between projects |
| `configurations` | Compare configurations between projects |
| `datasets` | Compare datasets between projects |
| `groups` | Compare groups between projects |
| `labels` | Compare labels between projects |
| `milestones` | Compare milestones between projects |
| `plans` | Compare test plans between projects |
| `retry-failed-pages` | Retry only failed case pages from a report |
| `runs` | Compare test runs between projects |
| `sections` | Compare sections between projects |
| `sharedsteps` | Compare shared steps between projects |
| `suites` | Compare test suites between projects |
| `templates` | Compare templates between projects |

## Flags ⚙️

```text
-h, --help               help for compare
--page-retries int       Number of retries per page in the main loading phase (default 5)
--parallel-pages int     Maximum number of parallel pages within a suite (default 6)
--parallel-suites int    Maximum number of parallel suites (default 10)
-1, --pid1 string        First project ID (required)
-2, --pid2 string        Second project ID (required)
--rate-limit int         API request limit per minute. -1 = auto by profile/deployment, 0 = no limit, >0 = fixed value. (default -1)
--retry-attempts int     Number of attempts for auto-retry of failed pages (default 5)
--retry-delay duration   Pause between retries for a single page during auto-retry (default 200ms)
--retry-workers int      Number of parallel workers during auto-retry of failed pages (default 12)
--save                   Save result to file (default: ~/.gotr/exports/)
--save-to string         Save result to specified file
--timeout duration       Timeout for compare operation (default 30m0s)
```

### Retry & Rate Limiting

| Flag | Description | Default |
| --- | --- | --- |
| `--rate-limit` | API request limit per minute (-1 = auto, 0 = no limit, >0 = fixed) | `-1` |
| `--page-retries` | Number of retries per page in the main loading phase | `5` |
| `--retry-attempts` | Number of attempts for auto-retry of failed pages | `5` |
| `--retry-workers` | Number of parallel workers during auto-retry | `12` |
| `--retry-delay` | Delay between retries for a single page | `200ms` |

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
gotr compare --help
```

✅ **Why this matters:** keeps execution aligned with the current CLI version and avoids stale command assumptions.

---

### ▶️ Scenario 2: Targeted action validation
🎯 **Goal:** validate the exact subcommand/shape for the operation you need.

```bash
gotr compare all -1 <project_id_1> -2 <project_id_2> --save
```

✅ **Why this matters:** prevents wrong endpoint selection and reduces trial-and-error in production pipelines.

---

### ▶️ Scenario 3: Safe or machine-readable run
🎯 **Goal:** get deterministic output for analysis and automation.

```bash
gotr compare cases -1 <project_id_1> -2 <project_id_2> --timeout 30m
```

✅ **Why this matters:** enables safer checks and structured post-processing in CI/CD.

---

### ▶️ Scenario 4: Mini operational pipeline
🎯 **Goal:** demonstrate a practical flow: validate -> run -> persist artifact.

```bash
gotr compare all -1 <project_id_1> -2 <project_id_2> --save-to ./compare-report.json
```

✅ **Why this matters:** provides a reusable template for runbooks and scripted operations.

---

## ⚡ Quick Start (30 seconds)

1. Validate syntax and available flags quickly:
```bash
gotr compare --help
```
2. Choose the operational execution path:
```bash
gotr compare all -1 <project_id_1> -2 <project_id_2>
```
3. Execute safe/operational run:
```bash
gotr compare all -1 <project_id_1> -2 <project_id_2> --save
```

---

## 🧪 Pre-run Checklist

- [ ] URL, credentials, and TestRail access are validated.
- [ ] Project/suite/case identifiers are confirmed.
- [ ] A safe/diagnostic run was executed (`--help`, `--dry-run`, `--json`, or `--save`).
- [ ] Output format and artifact storage location are defined.

---

## 🎯 When To Use

- Use `compare` when the task belongs to this command domain and you need predictable repeatable behavior.
- Use it when you want a clear flow from syntax validation to operational execution.

---

## 🚫 When Not To Use

- Do not run directly if target IDs/endpoints are uncertain: validate with `--help` and a safe check first.
- Do not force this command for bulk operations outside its domain: pick a more specialized command/subcommand.

---

## FAQ ❓

- ❓ **Question:** When should I use `compare`?
  > ↪️ **Answer:** use it when your task belongs to this command domain and you want predictable resource-focused behavior.
  >
  > ---

- ❓ **Question:** Where should I start if parameters are unclear?
  > ↪️ **Answer:** always start with `gotr compare --help`, then inspect the target subcommand help before execution.
  >
  > ---

- ❓ **Question:** Which subcommands should be validated first?
  > ↪️ **Answer:** recommended starting set: all, cases, configurations, datasets, groups, labels, milestones, plans. Begin with the highest-frequency operation in your release workflow.
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

### Run artifacts

- With `--save`, output is persisted to file for audit and diff workflows.


---

## 🔎 Result Verification via Neighbor Commands

- Store baseline output and diff against rerun using the same `pid1/pid2` pair.

```bash
# baseline
gotr compare all -1 <project_id_1> -2 <project_id_2> --save-to ./baseline.json

# rerun after changes
gotr compare all -1 <project_id_1> -2 <project_id_2> --save-to ./after.json
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
