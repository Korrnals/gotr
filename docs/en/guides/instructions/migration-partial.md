# Instruction: Partial Migration (cases with mapping)

Language: [Русский](../../../ru/guides/instructions/migration-partial.md) | English

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

Partial migration — transferring **only test cases** between suites of two projects.
Used when shared steps have **already been transferred** separately and you have a mapping file with old-to-new ID correspondence.

The `gotr sync cases` command automatically:

1. Fetches cases from the source suite
2. Replaces `shared_step_id` in each case using the mapping file
3. Imports cases into the target suite

> [!TIP]
> This scenario is the second step after `gotr sync shared-steps --save-mapping`.
> For transferring everything at once, use [Full Migration](migration-full.md).

## Prerequisites ✅

- [ ] gotr configured and connected to TestRail (`gotr self-test`)
- [ ] Shared steps already transferred (or not used in cases)
- [ ] Mapping file from previous `sync shared-steps` step
- [ ] Target suite already exists in target project

## Example: Transfer Cases After Shared Steps Migration 🚀

### Input Data

| Parameter | Value | Description |
| --- | --- | --- |
| Source project | `30` | Project R189 |
| Source suite | `20069` | Suite with cases |
| Target project | `34` | E2E Testing Scenarios |
| Target suite | `19859` | R189 Scenarios (transfer) |
| Mapping file | `mapping.json` | Result from previous `sync shared-steps` |

### Step 1. Verify mapping file is present

```bash
# Check mapping file contents
cat mapping.json
```

The mapping file contains pairs `old_id → new_id` for shared steps.

### Step 2. Dry-run — verify plan

```bash
gotr sync cases \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 34 \
  --dst-suite 19859 \
  --mapping-file mapping.json \
  --dry-run
```

**What to check:**

- Number of cases for transfer
- Correctness of `shared_step_id` replacement

### Step 3. Execute migration

```bash
gotr sync cases \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 34 \
  --dst-suite 19859 \
  --mapping-file mapping.json
```

### Step 4. Verify result

```bash
# Check cases in target suite
gotr export cases -p 34 -s 19859 --save --format json

# Compare suites
gotr compare cases --pid1 30 --pid2 34 --save
```

## Syntax 🧩

```bash
gotr sync cases \
  --src-project <ID> \
  --src-suite <ID> \
  --dst-project <ID> \
  --dst-suite <ID> \
  [--mapping-file <path>] \
  [--compare-field <field>] \
  [--output <path>] \
  [--dry-run] \
  [--quiet]
```

## Flags ⚙️

| Flag | Description | Default |
| --- | --- | --- |
| `--src-project` | Source project ID | required |
| `--src-suite` | Source suite ID | required |
| `--dst-project` | Target project ID | required |
| `--dst-suite` | Target suite ID | required |
| `--mapping-file` | Path to shared steps mapping file | — |
| `--compare-field` | Field for duplicate detection | `title` |
| `--output` | Path for JSON results file | — |
| `--dry-run` | Show plan without changes | `false` |
| `--quiet` | Suppress service output | `false` |

## Step-by-step: Two Steps Instead of sync full 🧩

If `sync full` doesn't suit your needs, execute two steps separately:

```bash
# Step A: transfer shared steps and save mapping
gotr sync shared-steps \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 34 \
  --save-mapping --approve

# Step B: transfer cases with ID substitution
gotr sync cases \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 34 \
  --dst-suite 19859 \
  --mapping-file mapping.json
```

## FAQ ❓

- ❓ **Question:** What if no mapping file is provided but cases reference shared steps?
  > ↪️ **Answer:** cases will be transferred with original `shared_step_id`. If those IDs don't exist in the target project, the references will be broken.
  >
  > ---

- ❓ **Question:** Can I transfer cases without shared steps at all?
  > ↪️ **Answer:** yes, if cases don't use shared steps — simply omit `--mapping-file`.
  >
  > ---

- ❓ **Question:** What if some shared steps already existed in the target project?
  > ↪️ **Answer:** the mapping file from `sync shared-steps` contains `existing` entries for duplicates — replacement will work correctly.

---

← [Instructions](index.md)
