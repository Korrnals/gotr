# Instruction: Full Migration

Language: [Русский](../../../ru/guides/instructions/migration-full.md) | English

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

Full migration transfers **shared steps + test cases** from one project/suite to another in a single pass.
The `gotr sync full` command automatically:

1. Loads shared steps from the source project
2. Applies shared steps filtering with the source suite in mind
3. Deduplicates against the target project by `title`
4. Imports new shared steps and saves a mapping (old ID → new ID)
5. Loads cases from the source suite
6. Replaces `shared_step_id` in cases according to the mapping
7. Imports cases into the target suite

> [!TIP]
> Always start with `--dry-run` to see the migration plan without making changes.

## Important: how shared step IDs are assigned

- The shared step ID from the source project is not carried over "as is" to the target.
- When a shared step is created the `add_shared_step/<project_id>` API is called, and the new ID is assigned by TestRail.
- Even if the same numeric ID is "free" in the target, the utility cannot force-claim it.
- The link between cases and shared steps is preserved via the `source_shared_step_id -> target_shared_step_id` mapping.
- In `sync full` the mapping is built during the shared steps transfer step and is applied automatically when cases are migrated.
- If a shared step already exists in the target (a duplicate by the comparison field, usually `title`), it is recorded in the mapping with the `existing` status and cases are pointed at the already existing target ID.

## Prerequisites ✅

- [ ] gotr is configured and connected to TestRail (`gotr self-test`)
- [ ] Source project and suite IDs are known
- [ ] Target project and suite IDs are known
- [ ] The target suite already exists in the target project
- [ ] Read access to the source project and write access to the target project

## Example: Cross-project Migration 🚀

### Input Data

| Parameter | Value | Description |
| --- | --- | --- |
| Source project | `30` | Project R189 |
| Source suite | `20069` | Suite with cases to transfer |
| Target project | `34` | E2E Scenario Testing |
| Target suite | `19859` | R189 Scenarios (transfer) |

### Step 1. Recon — verify the source data

```bash
# Check the connection
gotr self-test

# View shared steps in the source project
gotr get sharedsteps 30

# View cases in the source suite
gotr export cases -p 30 -s 20069 --save --format json
```

### Step 2. Dry-run — preview the migration plan

```bash
gotr sync full \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 34 \
  --dst-suite 19859 \
  --dry-run --save-filtered
```

**What to check:**

- Number of shared steps that will be transferred
- Number of cases for migration
- Which shared steps are marked as duplicates (already present in the target project)

### Step 3. Execute the migration

```bash
gotr sync full \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 34 \
  --dst-suite 19859 \
  --save-mapping --approve
```

### Step 4. Verify the result

```bash
# Check shared steps in the target project
gotr get sharedsteps 34

# Check cases in the target suite
gotr export cases -p 34 -s 19859 --save --format json

# Compare projects for verification
gotr compare all --pid1 30 --pid2 34 --save
```

## Syntax 🧩

```bash
gotr sync full \
  --src-project <ID> \
  --src-suite <ID> \
  --dst-project <ID> \
  --dst-suite <ID> \
  [--compare-field <field>] \
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
| `--src-suite` | Source suite ID | required |
| `--dst-project` | Target project ID | required |
| `--dst-suite` | Target suite ID | required |
| `--compare-field` | Field used for duplicate detection | `title` |
| `--dry-run` | Show plan without changes | `false` |
| `--save-mapping` | Save mapping to file | `false` |
| `--save-filtered` | Save filtered candidate list | `false` |
| `--approve` | Skip confirmation prompt | `false` |
| `--quiet` | Suppress service output | `false` |

## Expected Result 🧾

### Successful Migration

- Shared steps from the source project appear in the target project
- Test cases are created in the target suite with correct `shared_step_id` values
- Mapping file is saved (if `--save-mapping` was used)
- Command exits with code `0`

### Artifacts

| File | When created | Contents |
| --- | --- | --- |
| `mapping.json` | with `--save-mapping` | Mapping from old shared step IDs to new IDs |
| `filtered.json` | with `--save-filtered` | Candidate list after filtering |

## Migration Rollback

`sync full` supports rollback via snapshot.

### Quick rollback right after migration

In the post-action menu choose:

- `↻ Rollback this migration`

The utility will delete the entities created in the target in a safe dependency order:

1. cases
2. shared steps

### Rollback later by snapshot ID

```bash
# Find the snapshot
gotr snap list

# View details
gotr snap info <snapshot_id>

# Rollback preview without changes
gotr snap rollback <snapshot_id> --dry-run

# Execute the rollback
gotr snap rollback <snapshot_id>
```

### Partial rollback

You can roll back only a subset of the created target objects:

```bash
gotr snap rollback <snapshot_id> --entity-ids 12345,12346
```

### Important rollback boundaries

- Rollback removes only objects created during this migration.
- Pre-existing objects in the target (`existing` duplicates) are not deleted.
- If some entities have already been removed manually, rollback continues and marks the partial result as resumable.
- Re-running the same rollback only reprocesses the unsuccessful or unprocessed items.

## FAQ ❓

- ❓ **Question:** What if shared steps already exist in the target project?
  > ↪️ **Answer:** gotr automatically detects duplicates by the `title` field (or another field via `--compare-field`). Existing steps are not duplicated and are added to the mapping as `existing`.
  >
  > ---

- ❓ **Question:** Can I transfer only shared steps without cases?
  > ↪️ **Answer:** yes, use `gotr sync shared-steps` — see [Shared Steps Migration](migration-shared-steps.md).
  >
  > ---

- ❓ **Question:** What if the target suite does not exist?
  > ↪️ **Answer:** create it beforehand via `gotr add suite` or use `gotr sync suites` to migrate the entire suite.
  >
  > ---

- ❓ **Question:** How do I roll back a migration?
  > ↪️ **Answer:** via snapshots: `gotr snap rollback <snapshot_id>` or the `↻ Rollback this migration` menu item right after `sync full`.

---

← [Instructions](index.md)
