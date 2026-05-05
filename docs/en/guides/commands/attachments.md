# Command: attachments

Language: [Русский](../../../ru/guides/commands/attachments.md) | English

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
Manage file attachments for test cases, plans, results, and runs.
Supported resource types for file attachments:

> [!TIP]
> For a quick `attachments` workflow: inspect `--help`, then run a
> safe/validation command before operational execution.

## Syntax 🧩
```bash
gotr attachments [command]
```

## Subcommands

| Subcommand | Description |
| --- | --- |
| `add` | Add an attachment to a resource |
| `list` | List attachments for case, plan, plan-entry, project, run, and test |
| `cleanup` | Bulk-delete attachments older than a cutoff with snapshot+rollback |

## Flags ⚙️

```text
-h, --help   help for attachments
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
gotr attachments --help
```

✅ **Why this matters:** keeps execution aligned with the current CLI version and avoids stale command assumptions.

---

### ▶️ Scenario 2: Targeted action validation
🎯 **Goal:** validate the exact subcommand/shape for the operation you need.

```bash
gotr attachments add --help
```

✅ **Why this matters:** prevents wrong endpoint selection and reduces trial-and-error in production pipelines.

---

### ▶️ Scenario 3: Safe or machine-readable run
🎯 **Goal:** get deterministic output for analysis and automation.

```bash
gotr attachments
```

✅ **Why this matters:** enables safer checks and structured post-processing in CI/CD.

---

### ▶️ Scenario 4: Mini operational pipeline
🎯 **Goal:** demonstrate a practical flow: validate -> run -> persist artifact.

```bash
gotr attachments add --help && gotr attachments add
```

✅ **Why this matters:** provides a reusable template for runbooks and scripted operations.

---

### ▶️ Scenario 5: List project attachments
🎯 **Goal:** inspect already uploaded files at project scope.

```bash
gotr attachments list project 1
```

✅ **Why this matters:** helps validate migration/upload outcomes without opening UI manually.

---

### ▶️ Scenario 6: Bulk cleanup of old attachments (with snapshot + rollback)

🎯 **Goal:** reclaim storage by removing attachments older than a cutoff, while keeping a recovery snapshot.

```bash
# Preview only — lists candidates per project, no snapshot, no deletes.
gotr attachments cleanup --all-projects --older-than 90d --dry-run

# Real cleanup of result-bound attachments older than 3 months across
# every visible project. A snapshot is taken automatically and kept
# for 7 days (default category TTL for `cleanup-attachments`).
gotr attachments cleanup --all-projects --older-than 3M --entity-type result

# Restrict to specific projects, raise concurrency, and disable the
# snapshot (NOT recommended outside of throwaway environments).
gotr attachments cleanup --project 7,9 --older-than 30d --concurrency 8 --no-snapshot --force
```

**Flags worth knowing:**

| Flag | Purpose |
| --- | --- |
| `--project <ids>` / `--all-projects` | Required scope selector (mutually exclusive). |
| `--older-than <dur>` | Cutoff (e.g. `7d`, `3M`, `1y`, `720h`). Required unless `--dry-run`. |
| `--entity-type` | Filter by parent kind: `case`, `run`, `plan`, `plan_entry`, `result`, `test`. **Default since v3.5.2: all six kinds.** When the resolved scope covers ALL kinds, the command prints a ⚠️ pre-scan warning to stderr; narrow scopes get a one-line `ℹ️` confirmation. Pass an explicit list (e.g. `--entity-type result`) to limit the walk. |
| `--dry-run` | Preview only; no snapshot, no deletes. |
| `--no-snapshot` | Skip the safety-net snapshot. |
| `--snapshot-retention <dur>` | Override snapshot TTL (default `7d`; honored by `gotr snap gc`). |
| `--max-size-gb <n>` | Abort if estimated snapshot exceeds this size unless `--force`. |
| `--concurrency <n>` | Parallel delete workers (default `4`). |
| `--backup-concurrency <n>` | Worker count for snapshot downloads (default `0` = auto: `min(8, --concurrency)`). Since v3.6.0. |
| `--skip-references` | Disable the markdown reference scan during backup. References that point at the deleted attachments will not be rewritten on restore. Use only when the entity API is unhealthy. Since v3.6.0. |
| `--limit <n>` | Cap total attachments processed across all projects. |
| `--scan-strategy` | How to enumerate attachments: `auto` (default), `project`, `entities`. See **Compatibility** below. |
| `--chunk-size <n>` | Number of projects scanned per chunk; the checkpoint file is flushed atomically after every chunk. Default `10`. See **Recovery & resume** below. |
| `--scan-timeout-per-project <dur>` | Per-project scan deadline (default `10m`). On timeout, that project is marked `timeout` in the checkpoint and the run continues; resume re-tries it. Pass `0` to disable. |
| `--resume <RUN_ID>` | Resume an interrupted run from its checkpoint. Cannot be combined with scope/filter flags — they are restored from the checkpoint. |
| `--list-checkpoints` | Print all known checkpoints and exit. Mutually exclusive with every other flag. |
| `--no-report` | Skip emitting the deletion-audit report (Markdown + JSON + CSV + PDF). See **Deletion-audit report** below. |
| `--force` | Skip the final confirmation prompt. |

**🧭 Compatibility (TestRail Server vs Cloud):**

The bulk endpoint `get_attachments_for_project` was added in **TestRail 7.5**. Older self-hosted servers (TestRail Server &lt; 7.5) respond with `404 Unknown method`. To stay compatible with both, `gotr` ships **two scan strategies**:

| Strategy | How attachments are listed | Works on |
| --- | --- | --- |
| `project` | Single bulk call to `get_attachments_for_project`. | TestRail 7.5+ / Cloud |
| `entities` | Walk `get_suites → get_cases → get_attachments_for_case` plus (when enabled by `--entity-type`) `get_runs → get_attachments_for_run`, `get_plans → get_attachments_for_plan` + per-entry `get_attachments_for_plan_entry`, and `get_runs → get_tests → get_attachments_for_test` (the only path to result-bound attachments on Server &lt; 7.5). Results are deduplicated by attachment ID. | TestRail 5.7+ |
| `auto` (default) | Probe the project endpoint once on the first project; fall back to `entities` only when the canonical `404 Unknown method` is returned. Any other error aborts the run — no silent fallback. | both |

When the auto-probe falls back, an `INFO:` line is printed to stderr explaining the switch. In CI/non-interactive flows you can pin the strategy explicitly with `--scan-strategy=entities` or `--scan-strategy=project` to skip the probe call entirely.

**Rollback:** the snapshot is recorded under category `cleanup-attachments` and can be restored with the standard snap workflow:

```bash
gotr snap list --category cleanup-attachments
gotr snap rollback <snap-id>

# Verify SHA-256 of every snapshot file before restore (since v3.6.0).
gotr snap rollback <snap-id> --verify-integrity

# Skip the markdown reference rewrite step (since v3.6.0).
gotr snap rollback <snap-id> --skip-references
```

⚠️ TestRail's API has no `add_attachment_to_test` endpoint, so attachments whose only parent is a `test` are reported as **Skipped** during rollback — clean them up only when re-upload of test-bound binaries is acceptable.

**🔐 Snapshot completeness (since v3.6.0):**

Every cleanup snapshot now ships with the artifacts required to fully reverse the deletion, including an audit index of markdown bodies that referenced the deleted binaries:

```
<snap_id>/
├── meta.json            # snapshot metadata
├── attachments.json     # v2 ledger: per-attachment SHA-256 + restorability flag + new_id slot
├── references.json      # audit index: markdown URLs in other entities that pointed at the deleted attachments
├── integrity.json       # per-file SHA-256 + merkle-style root over the whole snapshot
└── files/<id>(.gz)      # binary payloads, optionally gzip-compressed
```

- **`attachments.json`** carries one entry per attachment: original ID, on-disk SHA-256, parent entity, a `restorable` flag (`false` for test-bound attachments, since TestRail has no `add_attachment_to_test` endpoint), and a `new_id` slot populated on rollback with the freshly issued TestRail ID.
- **`references.json`** is the audit index of every markdown URL pointing at one of the deleted attachments, scanned across `case`, `run`, `plan`, and `milestone` description fields plus `case.custom_steps_separated`. **In v3.6.0 the index is recorded but external entity bodies are NOT modified** — neither on cleanup nor on rollback. After a cleanup, links of the form `[label](.../attachments/get/<old_id>)` inside other entities will return 404; after a rollback the attachments come back under **new** TestRail IDs, so the old links remain broken. Manual rewrite is supported via the saved index. Automatic rewrite is on the roadmap (deferred from v3.6.0 due to read-modify-write race risk on shared TestRail entities).
- **`integrity.json`** records SHA-256 for every file inside the snapshot directory and a merkle-style root computed over the sorted `<path>|<sha256>` lines. Pass `--verify-integrity` to `gotr snap rollback` to re-verify before any API call.
- Backup downloads are parallelized via `--backup-concurrency`; the SHA-256 is computed inline so no second pass is needed.

**⚠️ Known limitations of cleanup + rollback (v3.6.0):**

Even after a successful `--verify-integrity` rollback, the restored state is **not byte-identical** to the pre-cleanup state. TestRail's API forces the following deltas:

| Field | Before cleanup | After rollback | Why |
| --- | --- | --- | --- |
| `attachment_id` | original | **new** (mapped in `attachments.json` `new_id`) | `add_attachment_to_*` always issues a fresh ID |
| `created_on` | original timestamp | **rollback timestamp** | API does not accept a back-dated value |
| `created_by` | original uploader | **API user running rollback** | API does not accept a back-dated user |
| URL | `.../attachments/get/<old_id>` | `.../attachments/get/<new_id>` | follows from the new ID |
| Markdown refs in OTHER entity bodies | pointed at `<old_id>` | **still point at `<old_id>` → 404** | gotr does not modify external entities; see `references.json` |
| Test-bound attachments | present | **skipped** | no `add_attachment_to_test` endpoint |

Practical consequences:

- For audit/compliance flows that rely on original `created_on`/`created_by`, treat rollback as a **content-recovery** tool, not a time-machine.
- For markdown integrity inside case/run/plan/milestone bodies, use `references.json` from the snapshot to drive a manual or scripted rewrite. The deletion-audit report includes a `Known limitations: markdown references` section enumerating the per-entity-type counts.
- If broken external links are unacceptable, prefer `--dry-run` + manual archival over `cleanup` for those projects until automatic rewrite ships.

**📑 Deletion-audit report (since v3.5.1):**

Every `attachments cleanup` run — including `--dry-run` — automatically writes a deletion-audit report under:

```
~/.gotr/reports/cleanup-attachments/<label>/<YYYY-MM>/cleanup-attachments-<UTC-timestamp>-<snapshot-id>.<ext>
```

Four formats are produced in lockstep:

| Format | Use case |
| --- | --- |
| `.md` | Human review. Contains the run header, applied filters, summary, per-project breakdown, the full list of deleted attachments and the rollback command. |
| `.json` | Machine consumption / audit pipelines. Stable schema (`internal/report/cleanup.Report`). |
| `.csv` | Spreadsheet review of every removed attachment (`project_id, attachment_id, name, size_bytes, parent_kind, parent_id, created_unix, deleted, dry_run, snapshot_id, …`). |
| `.pdf` | Self-contained handover artifact (NotoSans embedded). |

`INDEX.md` in the reports root is regenerated automatically after each run.

Dry-run reports carry a `DRY-RUN` marker in the Markdown title and `dry_run=true,deleted=false` in CSV rows so they cannot be confused with the real artifact.

Pass `--no-report` to suppress the four files entirely (useful for CI flows that already capture stdout).

✅ **Why this matters:** bulk cleanup without a snapshot is irreversible. The default flow keeps a one-week recovery window so you can roll back any over-eager deletion before the snapshot is GC'd.

**🔁 Recovery & resume (since v3.6.0):**

Long entity-walk scans across many projects can fail mid-flight (network blip, Ctrl-C, OS reboot). To make those runs recoverable, `cleanup` chunks the project set and persists progress between chunks:

- **Run id:** every run gets a stable id (e.g. `20260324T204109-3f9a12`) printed once at start: `INFO: cleanup run-id=<RUN_ID> (resume with --resume <RUN_ID>)`.
- **Checkpoint location:** `~/.gotr/cache/cleanup-attachments/<RUN_ID>/checkpoint.json` plus `partial-plan.json`.
- **Chunking:** projects are processed in groups of `--chunk-size` (default `10`). After each chunk finishes, both files are written via `tmp + fsync + rename` (atomic, crash-safe).
- **Per-project timeout:** `--scan-timeout-per-project` (default `10m`) caps each project's scan. Timed-out and failed projects are recorded in the checkpoint and skipped from the produced plan, but the run continues.
- **Clean completion:** when every project reaches `done`, the checkpoint directory is removed automatically. Otherwise it is preserved with a `WARN:` summary listing the affected project ids.
- **Listing checkpoints:** `gotr attachments cleanup --list-checkpoints` prints a table (RUN_ID, started, updated, totals, done, pending, failed, timeout) and the resume hint.
- **Resuming:** `gotr attachments cleanup --resume <RUN_ID>` restores the original scope, filters, scan strategy, limit, and chunk size from the checkpoint and only retries projects in `pending`/`retry_pending` (failures and timeouts). The resume command rejects mutually-exclusive scope/filter flags — change the cutoff or scope by starting a fresh run instead.

> The `--resume` invocation must match the original filter set. If the cached `FilterSnapshot` no longer matches what the CLI would build, the run aborts with `checkpoint mismatch` rather than silently mixing two filters.

**📊 Progress reporting (since v3.6.1):**

The scan phase emits a 5-line live block on stderr when stderr is a TTY (and `--quiet` is not set):

```
Scanning attachments — entity
   project 4/12  →  Demo project
   phase: cases      ████░░░░░░ 28 / 70
   found: 137  eligible: 89  size: 24.50 MiB
   ⏱ 1m12s  ETA ~3m05s
```

- **Phase progression:** `project → suites → cases → runs → plans → tests`. Each phase shows its own done/total counter and a 10-character bar; counts come from the API responses (number of suites, cases per suite, runs, plans, tests per run).
- **Throttling:** progress events are throttled to ~50 ms to avoid stalling the scanner during high-concurrency phases.
- **Non-TTY / `--quiet`:** the multiline UI is suppressed and the command falls back to `INFO: project X/Y done: …` and `INFO: chunk N/M done — running totals: …` lines so logs stay grep-friendly in CI.
- **Per-project × per-entity-type table:** after the executor finishes, a breakdown table is printed under `Breakdown by project × entity type:` with one row per project and per-entity-type columns (`case run plan plan_entry result test`) plus a `Total` / `Size` per row and a `Total` footer.
- **Final summary block:** a `Final summary:` section follows with the absolute path of every audit report file, the snapshot id and absolute path under `~/.gotr/snaps/cleanup-attachments/<id>/`, and a `Next steps:` block with copy-pasteable rollback (`gotr snap rollback <id>`) and resume (`gotr attachments cleanup --resume <run-id>`) commands.

---

## ⚡ Quick Start (30 seconds)

1. Validate syntax and available flags quickly:
```bash
gotr attachments --help
```
2. Choose the operational execution path:
```bash
gotr attachments add --help
```
3. Execute safe/operational run:
```bash
gotr attachments add --help
```

---

## 🧪 Pre-run Checklist

- [ ] URL, credentials, and TestRail access are validated.
- [ ] Project/suite/case identifiers are confirmed.
- [ ] A safe/diagnostic run was executed (`--help`, `--dry-run`, `--json`, or `--save`).
- [ ] Output format and artifact storage location are defined.

---

## 🎯 When To Use

- Use `attachments` when the task belongs to this command domain and you need predictable repeatable behavior.
- Use it when you want a clear flow from syntax validation to operational execution.

---

## 🚫 When Not To Use

- Do not run directly if target IDs/endpoints are uncertain: validate with `--help` and a safe check first.
- Do not force this command for bulk operations outside its domain: pick a more specialized command/subcommand.

---

## FAQ ❓

- ❓ **Question:** When should I use `attachments`?
  > ↪️ **Answer:** use it when your task belongs to this command domain and you want predictable resource-focused behavior.
  >
  > ---

- ❓ **Question:** Where should I start if parameters are unclear?
  > ↪️ **Answer:** always start with `gotr attachments --help`, then inspect the target subcommand help before execution.
  >
  > ---

- ❓ **Question:** Which subcommands should be validated first?
  > ↪️ **Answer:** recommended starting set: add, list. Begin with the highest-frequency operation in your release workflow.
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
