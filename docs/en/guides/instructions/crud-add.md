# Instruction: Creating Objects (add)

Language: [Русский](../../../ru/guides/instructions/crud-add.md) | English

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

The `gotr add` command creates new objects in TestRail via POST API.
Supports interactive mode (wizard), dry-run, and creation from JSON file.

> [!WARNING]
> `add` **modifies data** in TestRail. Always use `--dry-run` to verify before creating.

> [!NOTE]
> The `gotr add` dispatcher does not support `milestone`, `plan`, or `entry` endpoints directly.
> Use dedicated subcommands instead:
> - `gotr milestones add <project_id> --name "..."` — create a milestone
> - `gotr plans add <project_id> --name "..."` — create a plan
> - `gotr plans entry add <plan_id>` — add a plan entry

## Examples 🚀

### Creating a Project

```bash
# Via flags
gotr add project --name "New project" --description "Description"

# Dry-run — check what will be created
gotr add project --name "New project" --dry-run

# Interactive wizard
gotr add project -i
```

### Creating a Suite

```bash
# Create suite in project
gotr add suite 30 --name "Regression" --description "Regression cases"

# Dry-run
gotr add suite 30 --name "Regression" --dry-run
```

### Creating a Test Case

```bash
# With minimal fields
gotr add case --title "Auth verification" \
  --suite-id 20069 --section-id 500

# With additional fields
gotr add case --title "Auth verification" \
  --suite-id 20069 --section-id 500 \
  --type-id 1 --priority-id 3 --template-id 1

# From JSON file
gotr add case --json-file case-data.json

# Interactive wizard
gotr add case -i
```

### Creating a Test Run

```bash
# Create run with all cases
gotr add run 30 --name "Smoke test" --suite-id 20069

# Create run with specific cases
gotr add run 30 --name "Smoke test" \
  --suite-id 20069 \
  --include-all=false \
  --case-ids "101,102,103"

# Assign to user
gotr add run 30 --name "Smoke test" \
  --suite-id 20069 \
  --assignedto-id 5
```

### Creating a Shared Step

```bash
# Create shared step
gotr add shared-step 30 --title "Log into system"

# From JSON file with steps
gotr add shared-step 30 --json-file shared-step-data.json
```

### Adding a Test Result

```bash
# Add result
gotr add result --test-id 12345 \
  --status-id 1 --comment "Passed" --elapsed "30s"
```

## Creation Modes 🧩

### Flags (inline)

```bash
gotr add <endpoint> [id] --name "Name" --description "Description"
```

### JSON file

```bash
gotr add <endpoint> [id] --json-file data.json
```

### Interactive wizard

```bash
gotr add <endpoint> -i
```

### Dry-run (preview)

```bash
gotr add <endpoint> [id] --name "Name" --dry-run
```

## Main Flags ⚙️

| Flag | Description |
| --- | --- |
| `--dry-run` | Show what will be created without sending |
| `-i, --interactive` | Interactive wizard |
| `--json-file` | Path to JSON file with data |
| `--save` | Save result to file |
| `-n, --name` | Resource name |
| `--title` | Title (for case) |
| `--description` | Description |
| `--suite-id` | Suite ID |
| `--section-id` | Section ID |
| `--type-id` | Type ID (for case) |
| `--priority-id` | Priority ID (for case) |
| `--template-id` | Template ID (for case) |
| `--milestone-id` | Milestone ID |
| `--assignedto-id` | User ID |
| `--case-ids` | Comma-separated case IDs (for run) |
| `--include-all` | Include all cases (for run) |

## Result Verification

```bash
# After creation — verify via get
gotr get project <new_id>
gotr get case <new_id>
gotr get suite <new_id>
```

## FAQ ❓

- ❓ **Question:** How to create an object from a prepared JSON?
  > ↪️ **Answer:** `gotr add <endpoint> --json-file data.json`. JSON format matches TestRail API.
  >
  > ---

- ❓ **Question:** What if dry-run is OK but creation fails?
  > ↪️ **Answer:** dry-run validates data format locally. API errors (duplicates, insufficient permissions, missing dependencies) only occur during the actual request.
  >
  > ---

- ❓ **Question:** Can I create multiple objects at once?
  > ↪️ **Answer:** for bulk case creation use `gotr cases bulk`. For other resources — sequential `add` calls.

---

← [Instructions](index.md)
