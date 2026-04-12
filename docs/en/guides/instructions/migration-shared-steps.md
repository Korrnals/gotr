# Instruction: Shared Steps Migration

Language: [Русский](../../../ru/guides/instructions/migration-shared-steps.md) | English

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

## Overview 🎯

Migration of only **shared test steps** between projects.
The `gotr sync shared-steps` command:

1. Fetches all shared steps from the source project
2. Fetches cases from the specified suite (if `--src-suite` is set)
3. Filters: keeps only shared steps linked to suite cases ("Used In / case_ids" field)
4. Deduplicates: excludes steps already present in the target project (by `title`)
5. Imports new shared steps into the target project
6. Saves mapping of old IDs → new IDs

> [!TIP]
> Use `--save-mapping` to save the correspondence file for subsequent
> `gotr sync cases --mapping-file` — this ensures correct references in cases.

## Prerequisites ✅

- [ ] gotr configured and connected to TestRail (`gotr self-test`)
- [ ] Source project ID is known
- [ ] Target project ID is known
- [ ] (Optional) Suite ID for shared steps filtering

## Scenario 1: Transfer Shared Steps Filtered by Suite 🚀

### Input Data

| Parameter | Value | Description |
| --- | --- | --- |
| Source project | `30` | Project R189 |
| Source suite | `20069` | Suite for filtering |
| Target project | `34` | E2E Testing Scenarios |

### Step 1. Recon

```bash
# View all shared steps in source project
gotr get sharedsteps 30

# Export for detailed analysis
gotr export sharedsteps -p 30 --save --format json
```

### Step 2. Dry-run

```bash
gotr sync shared-steps \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 34 \
  --dry-run --save-filtered
```

**What it shows:**

- How many shared steps are linked to cases in suite 20069
- How many already exist in the target project (duplicates)
- How many new steps will be created

### Step 3. Execute migration

```bash
gotr sync shared-steps \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 34 \
  --save-mapping --approve
```

### Step 4. Verify result

```bash
# Check shared steps in target project
gotr get sharedsteps 34

# Compare shared steps between projects
gotr compare sharedsteps --pid1 30 --pid2 34
```

## Scenario 2: Transfer ALL Shared Steps 🚀

Without `--src-suite`, all project shared steps are transferred:

```bash
# Dry-run — preview plan
gotr sync shared-steps \
  --src-project 30 \
  --dst-project 34 \
  --dry-run

# Execute
gotr sync shared-steps \
  --src-project 30 \
  --dst-project 34 \
  --save-mapping --approve
```

## Syntax 🧩

```bash
gotr sync shared-steps \
  --src-project <ID> \
  [--src-suite <ID>] \
  --dst-project <ID> \
  [--compare-field <field>] \
  [--output <path>] \
  [--dry-run] \
  [--save-mapping] \
  [--save-filtered] \
  [--approve] \
  [--quiet]
```

## Flags ⚙️

| Flag | Description | Default |
| --- | --- | --- |
| `--src-project` | Source project ID | required |
| `--src-suite` | Suite ID for "Used In" filtering | — (all steps) |
| `--dst-project` | Target project ID | required |
| `--compare-field` | Field for duplicate detection | `title` |
| `--output` | Path to save mapping | — |
| `--dry-run` | Show plan without changes | `false` |
| `--save-mapping` | Save mapping (old ID → new) | `false` |
| `--save-filtered` | Save filtered list | `false` |
| `--approve` | Skip confirmation prompt | `false` |
| `--quiet` | Suppress service output | `false` |

## How Filtering Works 📐

The shared steps filtering algorithm by suite:

```text
For each shared step in the source project:
  1. Check the case_ids field ("Used In")
  2. If at least one case_id belongs to cases from --src-suite → step is a CANDIDATE
  3. Among candidates: compare title with target project steps
     - Match → add to mapping as "existing" (don't import)
     - No match → IMPORT
```

## Expected Result 🧾

### Artifacts

| File | Contents |
| --- | --- |
| `shared_steps_mapping_YYYY-MM-DD_HH-MM-SS.json` | `{ "source_id": 123, "target_id": 456, "status": "created" }` — for new ones |
| `shared_steps_mapping_YYYY-MM-DD_HH-MM-SS.json` | `{ "source_id": 789, "target_id": 101, "status": "existing" }` — for duplicates |
| `shared_steps_filtered_YYYY-MM-DD_HH-MM-SS.json` | List of shared steps that passed filtering |

## FAQ ❓

- ❓ **Question:** What if a shared step is used in multiple suites?
  > ↪️ **Answer:** filtering checks the intersection of `case_ids` with cases of the specified suite. If at least one case from the suite uses the step — it qualifies as a candidate.
  >
  > ---

- ❓ **Question:** What happens on repeated execution?
  > ↪️ **Answer:** steps already present in the target project (by title) will be marked as `existing` and won't be duplicated.
  >
  > ---

- ❓ **Question:** How to use the mapping file afterwards?
  > ↪️ **Answer:** pass it to `gotr sync cases --mapping-file mapping.json` — see [Partial Migration](migration-partial.md).

---

← [Instructions](index.md)
