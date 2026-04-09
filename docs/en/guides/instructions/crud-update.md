# Instruction: Updating Objects (update)

Language: [Русский](../../../ru/guides/instructions/crud-update.md) | English

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

The `gotr update` command modifies existing objects in TestRail.
Supports updating by ID, interactive mode, and updating from JSON file.

> [!WARNING]
> `update` **modifies data** in TestRail. Use `--dry-run` to verify.

## Examples 🚀

### Updating a Project

```bash
# Change name
gotr update project 30 --name "R189 (updated)"

# Change description
gotr update project 30 --description "New project description"

# Dry-run
gotr update project 30 --name "R189 (updated)" --dry-run
```

### Updating a Test Case

```bash
# Change title
gotr update case 12345 --title "Updated title"

# Change priority and type
gotr update case 12345 --priority-id 4 --type-id 2

# From JSON file
gotr update case 12345 --json-file case-update.json

# Interactive mode
gotr update case 12345 -i
```

### Updating a Suite

```bash
gotr update suite 20069 --name "Updated suite" --description "New description"
```

### Updating a Shared Step

```bash
# Change title
gotr update shared-step 456 --title "Updated step"

# From JSON with full data
gotr update shared-step 456 --json-file step-update.json
```

### Updating a Test Run

```bash
# Change name and description
gotr update run 789 --name "Regression v2" --description "Updated run"

# Reassign
gotr update run 789 --assignedto-id 10
```

### Updating a Milestone

```bash
gotr update milestone 50 --name "Release 3.1" --description "New milestone"
```

## Update Modes 🧩

### Flags (inline)

```bash
gotr update <endpoint> <id> --title "New value"
```

### JSON file

```bash
gotr update <endpoint> <id> --json-file data.json
```

### Interactive wizard

```bash
gotr update <endpoint> <id> -i
```

### Dry-run (preview)

```bash
gotr update <endpoint> <id> --title "New value" --dry-run
```

## Main Flags ⚙️

| Flag | Description |
| --- | --- |
| `--dry-run` | Show what will be changed without sending |
| `-i, --interactive` | Interactive wizard |
| `--json-file` | Path to JSON file with data |
| `--title` | New title |
| `-n, --name` | New name |
| `--description` | New description |
| `--priority-id` | New priority |
| `--type-id` | New type |
| `--labels` | New labels |

## Result Verification

```bash
# After updating — verify via get
gotr get case 12345
gotr get project 30
gotr get suite 20069
```

## Typical Pipeline: get → verify → update → get

```bash
# 1. Get current state
gotr get case 12345 --format json

# 2. Check what will be changed
gotr update case 12345 --title "New title" --dry-run

# 3. Execute update
gotr update case 12345 --title "New title"

# 4. Verify result
gotr get case 12345
```

## FAQ ❓

- ❓ **Question:** Which fields can be updated?
  > ↪️ **Answer:** depends on the endpoint. Use `gotr update <endpoint> --help` for the full list of flags for a specific resource.
  >
  > ---

- ❓ **Question:** Can I update multiple objects at once?
  > ↪️ **Answer:** for bulk case updates use `gotr cases bulk`. For others — sequential calls.
  >
  > ---

- ❓ **Question:** What if I specify a non-existent ID?
  > ↪️ **Answer:** TestRail API will return a 400/404 error, gotr will display the error message.

---

← [Instructions](index.md)
