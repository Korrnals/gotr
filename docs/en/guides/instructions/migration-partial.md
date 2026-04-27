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
      - [Interactive Migration Walkthrough](migration-interactive-walkthrough.md)
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
Used when shared steps have **already been transferred** separately and you have a mapping file with the correspondence between old and new IDs.

The `gotr sync cases` command automatically:

1. Loads cases from the source suite
2. Replaces `shared_step_id` in each case using the mapping file
3. Imports cases into the target suite

> [!TIP]
> This scenario is the second step after `gotr sync shared-steps --save-mapping`.
> To transfer everything in one go use [Full Migration](migration-full.md).

## Important: how shared step IDs are handled

- Shared steps in the target receive IDs assigned by TestRail at creation time.
- The source ID is not "preserved" even if such a number is technically free in the target.
- A correct link between cases and shared steps is guaranteed only by remapping with the mapping file.
- For cases with `shared_step_id` it is therefore strongly recommended to always use `--mapping-file`.

## Prerequisites ✅

- [ ] gotr is configured and connected to TestRail (`gotr self-test`)
- [ ] Shared steps have already been transferred (or are not used by the cases)
- [ ] A mapping file from the previous `sync shared-steps` step is available
- [ ] The target suite already exists in the target project

## Example: Transfer Cases After Shared Steps Migration 🚀

### Input Data

| Parameter | Value | Description |
| --- | --- | --- |
| Source project | `30` | Project R189 |
| Source suite | `20069` | Suite with cases |
| Target project | `34` | E2E Scenario Testing |
| Target suite | `19859` | R189 Scenarios (transfer) |
| Mapping file | `mapping.json` | Result of the previous `sync shared-steps` |

### Step 1. Make sure the mapping file is in place

```bash
# Check the contents of the mapping file
cat mapping.json
```

The mapping file contains `old_id → new_id` pairs for shared steps.

### Step 2. Dry-run — verify the plan

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

### Step 3. Execute the migration

```bash
gotr sync cases \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 34 \
  --dst-suite 19859 \
  --mapping-file mapping.json
```

### Step 4. Verify the result

```bash
# Check cases in the target suite
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
| `--mapping-file` | Path to the shared steps mapping file | — |
| `--compare-field` | Field used for duplicate detection | `title` |
| `--output` | Path for the JSON results file | — |
| `--dry-run` | Show plan without changes | `false` |
| `--quiet` | Suppress service output | `false` |

## Step-by-step scenario: two steps instead of sync full 🧩

If for some reason `sync full` is not suitable, run the two steps separately:

```bash
# Step A: transfer shared steps and save the mapping
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

## Rollback for partial migration

Snapshot rollback also works for `sync cases`.

### Right after the command

- In the post-action menu choose `↻ Rollback this migration`

### Later, by snapshot ID

```bash
# View snapshots
gotr snap list

# Check the diff before rollback
gotr snap rollback <snapshot_id> --dry-run

# Execute the rollback
gotr snap rollback <snapshot_id>
```

### What exactly is rolled back

- For `sync cases`, the cases created in the target during a specific run are deleted.
- Shared steps that were transferred separately are not deleted by the `sync cases` rollback.
- To roll back a two-step migration completely you typically run rollback for both snapshots:
  - first for `sync cases`
  - then for `sync shared-steps`

## FAQ ❓

- ❓ **Question:** What if no mapping file is provided but cases reference shared steps?
  > ↪️ **Answer:** cases will be transferred with the original `shared_step_id` values. If those IDs do not exist in the target project, the references will be broken.
  >
  > ---

- ❓ **Question:** If the target project has a free ID that matches the source ID, will it be used?
  > ↪️ **Answer:** no. The shared step ID is always assigned by the TestRail server during `add_shared_step`; the utility does not set it manually.
  >
  > ---

- ❓ **Question:** Can I transfer cases without shared steps at all?
  > ↪️ **Answer:** yes, if the cases do not use shared steps — simply omit `--mapping-file`.
  >
  > ---

- ❓ **Question:** What if some shared steps were already present in the target project?
  > ↪️ **Answer:** the mapping file from `sync shared-steps` contains `existing` entries for duplicates — substitution will happen correctly.

---

← [Instructions](index.md)
