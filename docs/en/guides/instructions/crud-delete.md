# Instruction: Deleting Objects (delete)

Language: [Русский](../../../ru/guides/instructions/crud-delete.md) | English

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

The `gotr delete` command removes objects from TestRail.
Supports soft (`--soft`) and hard deletion.

> [!CAUTION]
> `delete` **irreversibly removes data** from TestRail.
> Always use `--dry-run` before deletion.
> It's recommended to export data first via `gotr export`.

## Examples 🚀

### Deletion with Verification

```bash
# 1. Export for backup first
gotr export cases -p 30 -s 20069 --save --format json

# 2. Dry-run — check what will be deleted
gotr delete case 12345 --dry-run

# 3. Execute deletion
gotr delete case 12345
```

### Deleting Various Resources

```bash
# Delete test case
gotr delete case 12345

# Delete suite
gotr delete suite 20069

# Delete section
gotr delete section 500

# Delete test run
gotr delete run 789

# Delete plan
gotr delete plan 100

# Delete project (soft delete)
gotr delete project 30 --soft

# Delete shared step
gotr delete shared-step 456

# Delete milestone
gotr delete milestone 50
```

### Soft Deletion

```bash
# Soft delete — object is marked as deleted but data remains
gotr delete project 30 --soft

# Hard delete — data is permanently removed
gotr delete project 30
```

## Syntax 🧩

```bash
gotr delete <endpoint> <id> [flags]
```

## Flags ⚙️

| Flag | Description | Default |
| --- | --- | --- |
| `--dry-run` | Show what will be deleted without executing | `false` |
| `--soft` | Soft deletion (where supported) | `false` |

## Safe Deletion Pipeline 🧩

```bash
# 1. Export data for backup
gotr export <resource> -p <project_id> --save --format json

# 2. Check object before deletion
gotr get <resource> <id>

# 3. Dry-run
gotr delete <endpoint> <id> --dry-run

# 4. Delete
gotr delete <endpoint> <id>

# 5. Confirm deletion
gotr get <resource> <id>  # Expect 404 error
```

## FAQ ❓

- ❓ **Question:** Can a deleted object be restored?
  > ↪️ **Answer:** with `--soft` — the object can be restored via TestRail UI. With hard delete — data is lost permanently. Always export before deleting.
  >
  > ---

- ❓ **Question:** What if I delete a suite that contains cases?
  > ↪️ **Answer:** TestRail API will delete the suite along with all cases and sections inside. Make sure this is intentional.
  >
  > ---

- ❓ **Question:** Can I delete multiple objects at once?
  > ↪️ **Answer:** `gotr delete` works with one object per call. For bulk operations use a shell script or `gotr cases bulk`.

---

← [Instructions](index.md)
