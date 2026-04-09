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

1. Fetches shared steps from the source project
2. Filters by case association in the source suite ("Used In" field)
3. Deduplicates against the target project by `title`
4. Imports new shared steps and saves a mapping (old ID → new ID)
5. Fetches cases from the source suite
6. Replaces `shared_step_id` in cases using the mapping
7. Imports cases into the target suite

> [!TIP]
> Always start with `--dry-run` to see the migration plan without making changes.

## Prerequisites ✅

- [ ] gotr configured and connected to TestRail (`gotr self-test`)
- [ ] Source project and suite IDs are known
- [ ] Target project and suite IDs are known
- [ ] Target suite already exists in the target project
- [ ] Read access to source project, write access to target project

## Example: Cross-project Migration 🚀

### Input Data

| Parameter | Value | Description |
| --- | --- | --- |
| Source project | `30` | Project R189 |
| Source suite | `20069` | Suite with cases for transfer |
| Target project | `34` | E2E Testing Scenarios |
| Target suite | `19859` | R189 Scenarios (transfer) |

### Step 1. Recon — verify source data

```bash
# Check connection
gotr self-test

# View shared steps in source project
gotr get sharedsteps 30

# View cases in source suite
gotr export cases -p 30 -s 20069 --save --format json
```

### Step 2. Dry-run — preview migration plan

```bash
gotr sync full \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 34 \
  --dst-suite 19859 \
  --dry-run --save-filtered
```

**What to check:**

- Number of shared steps to be transferred
- Number of cases for migration
- Which shared steps are marked as duplicates (already exist in target)

### Step 3. Execute migration

```bash
gotr sync full \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 34 \
  --dst-suite 19859 \
  --save-mapping --approve
```

### Step 4. Verify result

```bash
# Check shared steps in target project
gotr get sharedsteps 34

# Check cases in target suite
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
| `--compare-field` | Field for duplicate detection | `title` |
| `--dry-run` | Show plan without changes | `false` |
| `--save-mapping` | Save mapping to file | `false` |
| `--save-filtered` | Save filtered candidate list | `false` |
| `--approve` | Skip confirmation prompt | `false` |
| `--quiet` | Suppress service output | `false` |

## Expected Result 🧾

### Successful Migration

- Shared steps from source appear in target project
- Test cases created in target suite with correct `shared_step_id`
- Mapping file saved (if `--save-mapping` used)
- Command exits with code `0`

### Artifacts

| File | When created | Contents |
| --- | --- | --- |
| `mapping.json` | with `--save-mapping` | Old shared step IDs → new IDs |
| `filtered.json` | with `--save-filtered` | Candidates after filtering |

## FAQ ❓

- ❓ **Question:** What if shared steps already exist in the target project?
  > ↪️ **Answer:** gotr automatically detects duplicates by `title` (or other field via `--compare-field`). Existing steps are not duplicated — they are added to the mapping as `existing`.
  >
  > ---

- ❓ **Question:** Can I transfer only shared steps without cases?
  > ↪️ **Answer:** yes, use `gotr sync shared-steps` — see [Shared Steps Migration](migration-shared-steps.md).
  >
  > ---

- ❓ **Question:** What if the target suite doesn't exist?
  > ↪️ **Answer:** create it beforehand via `gotr add suite` or use `gotr sync suites` to migrate the entire suite.
  >
  > ---

- ❓ **Question:** How to rollback a migration?
  > ↪️ **Answer:** TestRail API doesn't support bulk rollback. Use `--dry-run` before execution. If needed — delete migrated objects via `gotr delete`.

---

← [Instructions](index.md)
