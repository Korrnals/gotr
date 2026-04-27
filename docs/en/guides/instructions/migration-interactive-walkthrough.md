# Instruction: Step-by-Step Interactive Migration Walkthrough

Language: [Русский](../../../ru/guides/instructions/migration-interactive-walkthrough.md) | English

## Navigation

- [Documentation](../../index.md)
  - [Guides](../index.md)
    - [Installation](../installation.md)
    - [Configuration](../configuration.md)
    - [Interactive Mode](../interactive-mode.md)
    - [Progress](../progress.md)
    - [Commands Index](../commands/index.md)
    - [Instructions](index.md)
      - **Interactive Migration Walkthrough**
      - [Full Migration](migration-full.md)
      - [Partial Migration](migration-partial.md)
      - [Shared Steps Migration](migration-shared-steps.md)
      - [Resources Migration](migration-resources.md)
      - [Fetching Data](crud-get.md)
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

This instruction describes **all migration variations** available through the gotr interactive mode.
Each scenario is a step-by-step walkthrough: what the user enters, what the tool displays,
which prompts appear and in what order.

At the end you will find a **real-world task example** — moving shared steps and cases
from project R189 into the "E2E Scenario Testing" project.

> [!TIP]
> All migration commands support a hybrid mode: some parameters can be supplied via
> flags while the rest are chosen interactively. See the
> [Interactive Mode](../interactive-mode.md) guide for details.

## Prerequisites ✅

- [ ] gotr is configured and connected to TestRail (`gotr self-test`)
- [ ] You have read access to the source project and write access to the destination
- [ ] The destination project exists (or will be created via `gotr add project`)

---

## How shared steps migration and case linkage work

When shared steps are moved to another project, TestRail assigns them **new IDs**.
If the cases are simply copied "as is", the references to shared steps inside the
case steps will point to the old IDs (which do not exist in the destination project).

gotr solves this through **automatic ID mapping**.

### Algorithm (two phases)

```text
┌─────────────────────────────────────────────────────────────────┐
│ Phase 1: MigrateSharedSteps                                    │
│                                                                 │
│  1. Load shared steps from SOURCE and DESTINATION               │
│  2. Compare by title (or another field)                         │
│     ├─ Match found → status: "existing"                         │
│     │   mapping: old_id → existing_target_id                    │
│     └─ No match → create in DESTINATION                         │
│         mapping: old_id → new_created_id (status: "created")    │
│  3. Result: mapping table held in memory                        │
│                                                                 │
│  Mapping example:                                               │
│  ┌──────────┬──────────┬──────────┐                             │
│  │ Source ID │ Target ID│ Status   │                             │
│  ├──────────┼──────────┼──────────┤                             │
│  │ 1001     │ 2001     │ created  │                             │
│  │ 1002     │ 2002     │ created  │                             │
│  │ 1003     │ 1850     │ existing │                             │
│  └──────────┴──────────┴──────────┘                             │
├─────────────────────────────────────────────────────────────────┤
│ Phase 2: MigrateCases                                          │
│                                                                 │
│  For every case:                                                │
│    For every step (CustomStepsSeparated):                       │
│      If SharedStepID ≠ 0:                                       │
│        → Look up new_id in the mapping by old SharedStepID      │
│        → Substitute new_id                                      │
│    Create the case in DESTINATION with the updated IDs          │
│                                                                 │
│  Before: Step { SharedStepID: 1001 }  ← old ID                 │
│  After:  Step { SharedStepID: 2001 }  ← new ID in destination  │
└─────────────────────────────────────────────────────────────────┘
```

### Duplicate detection

Before importing, gotr checks whether the destination project already contains a
shared step with the same **title** (the comparison field is configurable). If
one is found, the step is not created again — its ID is added to the mapping
with status `"existing"`. This way cases stay correctly linked even to
shared steps that were already present.

### Filtering by suite

When a source suite is specified, only the shared steps that are **actually used**
by cases of that suite are migrated. Unused ones are skipped. This saves time
and avoids polluting the destination project.

### Persisting the mapping

In a two-step migration (variation 2) the mapping is saved to a file:

```json
{
  "src_project_id": 30,
  "dst_project_id": 34,
  "created_at": "2026-04-18T14:30:45Z",
  "pairs": [
    {"source_id": 1001, "target_id": 2001, "status": "created"},
    {"source_id": 1003, "target_id": 1850, "status": "existing"}
  ]
}
```

It is then passed to `sync cases --mapping-file ...` for ID substitution.

> [!WARNING]
> If the mapping file is not supplied and no mapping was built (for example,
> cases are migrated independently without a preceding `sync shared-steps`),
> the references to shared steps inside the cases will **keep their old IDs**,
> producing dead links.

### With `sync full` everything is automatic

In `sync full` mode both phases run sequentially in a single process. The mapping
is built in memory during phase 1 and used immediately in phase 2 —
**no extra action is required from the user**.

### A frequent question about "free" IDs in the target

- Shared steps are not copied with their original IDs.
- Even if no object with that number exists in the target, that ID is not forced.
- A new ID is always assigned by TestRail when the shared step is created.
- Link stability is ensured not by matching numbers, but by the remap performed via the mapping.

---

## Variation 1: Full migration in a single command (`sync full`)

**When to use:** when both shared steps and test cases need to be migrated in one pass.

### Step 1. Launch

```bash
gotr sync full
```

### Step 2. Interactive prompts

```text
? Select SOURCE project:
  ← Back
  ✕ Exit
  ──────────────────────
  ID: 1  | SAP Hybris
  ID: 2  | SAP CRM
  ...
  ID: 30 | R189
→ Pick: R189
```

```text
? Select SOURCE suite:
  ← Back
  ✕ Exit
  ──────────────────────
  ID: 8411  | R189 IT Suites and Cases
  ID: 9709  | R189 PT Suites and Cases
  ...
  ID: 20069 | Temporary case suite
→ Pick: Temporary case suite
```

```text
? Select DESTINATION project:
  ← Back                          ← returns to source suite selection
  ✕ Exit
  ──────────────────────
  ID: 34 | E2E Scenario Testing
  ...
→ Pick: E2E Scenario Testing
```

```text
? Select DESTINATION suite:
  ← Back
  ✕ Exit
  ──────────────────────
  ID: 19859 | R189 Scenarios (migration)
  ...
→ Pick: R189 Scenarios (migration)
```

### Step 3. Snapshot and confirmation

```text
? 📦 Create snapshot before migration? (recommended) [Y/n] Y
? 🏷  Snapshot label (optional, press Enter to skip): R189 → E2E full migration

  ┌────────────────────────────────────────────┐
  │ Migration Plan                             │
  │ Source:      R189 / Temporary case suite   │
  │ Destination: E2E / R189 Scenarios          │
  │ Shared steps: 12 new, 3 existing           │
  │ Cases: 47 to migrate                       │
  └────────────────────────────────────────────┘

? Continue? [y/N] y
```

### Step 4. Result and post-action menu

```text
✓ Shared steps: 12 created, 3 mapped as existing
✓ Cases: 47 created in destination suite

? What next?
  ✕ Exit
  ↻ Rollback this migration
  📊 Compare: verify current state
  🔄 Sync: migrate data
  📦 Snap: manage snapshots
→ Pick: 📊 Compare — to verify the result
```

### Hybrid variant (some flags supplied)

```bash
# Source set, destination chosen interactively
gotr sync full --src-project 30 --src-suite 20069

# Everything set, only the confirmation is interactive
gotr sync full --src-project 30 --src-suite 20069 --dst-project 34 --dst-suite 19859

# Fully automated (no prompts)
gotr sync full --src-project 30 --src-suite 20069 --dst-project 34 --dst-suite 19859 --approve --save-mapping
```

---

## Variation 2: Two-step migration (shared steps → cases)

**When to use:** when control over each phase is needed. For example, after migrating
shared steps you want to inspect the mapping before moving the cases.

### Step A: Migrate shared steps

```bash
gotr sync shared-steps
```

```text
? Select SOURCE project (copy shared steps from):
→ ID: 30 | R189

? Specify source suite? [y/N] y
  (without a suite — ALL shared steps of the project are migrated)

? Select SOURCE suite:
→ ID: 20069 | Temporary case suite

? Select DESTINATION project (copy shared steps to):
→ ID: 34 | E2E Scenario Testing

  ┌──────────────────────────────────────────┐
  │ Shared Steps Summary                     │
  │ Total in source project: 45              │
  │ Used by cases in suite 20069: 15         │
  │ Already exist in destination: 3          │
  │ To import: 12                            │
  └──────────────────────────────────────────┘

? 📦 Create snapshot before migration? (recommended) [Y/n] Y
? Continue? [y/N] y

✓ Shared steps: 12 created, 3 mapped as existing

? Save mapping? [y/N] y
  → Saved: shared_steps_mapping_2026-04-18_14-30-00.json

? Save filtered shared steps list? [y/N] y
  → Saved: shared_steps_filtered_2026-04-18_14-30-00.json
```

### Step B: Migrate cases with a mapping

```bash
gotr sync cases --mapping-file shared_steps_mapping_2026-04-18_14-30-00.json
```

```text
? Select SOURCE project (copy from):
→ ID: 30 | R189

? Select SOURCE suite:
→ ID: 20069 | Temporary case suite

? Select DESTINATION project (copy to):
→ ID: 34 | E2E Scenario Testing

? Select DESTINATION suite:
→ ID: 19859 | R189 Scenarios (migration)

  ┌────────────────────────────────────────────┐
  │ Cases Migration Plan                       │
  │ Cases to migrate: 47                       │
  │ With shared_step_id replacement: 15        │
  │ Mapping file: shared_steps_mapping_....json│
  └────────────────────────────────────────────┘

? 📦 Create snapshot before migration? (recommended) [Y/n] Y
? Continue? [y/N] y

✓ Cases: 47 created (15 with remapped shared_step_id)
```

---

## Variation 3: Migrating ALL shared steps of a project (no filtering)

**When to use:** when every shared step has to be migrated without limiting to a specific suite.

```bash
gotr sync shared-steps
```

```text
? Select SOURCE project (copy shared steps from):
→ ID: 30 | R189

? Specify source suite? [y/N] n
  (no suite → all shared steps of the project)

? Select DESTINATION project (copy shared steps to):
→ ID: 34 | E2E Scenario Testing

  ┌──────────────────────────────────────┐
  │ Shared Steps Summary                 │
  │ Total in source project: 45          │
  │ Already exist in destination: 8      │
  │ To import: 37                        │
  └──────────────────────────────────────┘

? Continue? [y/N] y

✓ Shared steps: 37 created, 8 mapped as existing
```

---

## Variation 4: Structure migration (suites → sections)

**When to use:** when suites and sections need to be created in the destination project
before the cases are migrated.

### Step A: Migrate suites

```bash
gotr sync suites
```

```text
? Select SOURCE project:
→ ID: 30 | R189

? Select DESTINATION project:
→ ID: 34 | E2E Scenario Testing

  Suites to migrate: 3 (2 new, 1 existing)

? 📦 Create snapshot before migration? (recommended) [Y/n] Y
? Continue? [y/N] y

✓ Suites: 2 created, 1 mapped as existing

? Save mapping? [y/N] y
  → Saved: suites_mapping_2026-04-18_14-35-00.json
```

### Step B: Migrate sections

```bash
gotr sync sections
```

```text
? Select SOURCE project:
→ ID: 30 | R189

? Select SOURCE suite:
→ ID: 20069 | Temporary case suite

? Select DESTINATION project:
→ ID: 34 | E2E Scenario Testing

? Select DESTINATION suite:
→ ID: 19859 | R189 Scenarios (migration)

  Sections to migrate: 8 (6 new, 2 existing)

? 📦 Create snapshot before migration? (recommended) [Y/n] Y
? Continue? [y/N] y

✓ Sections: 6 created (parent-child hierarchy preserved)
```

---

## Variation 5: Full pipeline (structure → shared steps → cases)

**When to use:** full migration with maximum control at every stage.
Order: suites → sections → shared steps → cases.

```bash
# 1. Migrate suites
gotr sync suites --src-project 30 --dst-project 34 --save-mapping --approve

# 2. Migrate sections
gotr sync sections --src-project 30 --src-suite 20069 --dst-project 34 --dst-suite 19859 --save-mapping --approve

# 3. Migrate shared steps (filtered by suite)
gotr sync shared-steps --src-project 30 --src-suite 20069 --dst-project 34 --save-mapping --approve

# 4. Migrate cases (using the mapping from step 3)
gotr sync cases --src-project 30 --src-suite 20069 --dst-project 34 --dst-suite 19859 --mapping-file shared_steps_mapping_*.json

# 5. Verify the result
gotr compare all --pid1 30 --pid2 34 --save
```

---

## Variation 6: Discovery → comparison → migration (cross-navigation)

**When to use:** when the divergences must be assessed first and migrated afterwards.
Cross-navigation lets you jump from compare to sync without leaving the tool.

```bash
gotr compare all
```

```text
? Select first project (pid1):
→ ID: 30 | R189

? Select second project (pid2):
→ ID: 34 | E2E Scenario Testing

  ╔══════════════════════════════════════════╗
  ║ Compare All Results                      ║
  ╠══════════════════════════════════════════╣
  ║ Cases:       147 vs 100 (+47 missing)    ║
  ║ Shared steps: 45 vs 33 (+12 missing)    ║
  ║ Suites:       10 vs 10 (match)          ║
  ║ Sections:     24 vs 18 (+6 missing)     ║
  ╚══════════════════════════════════════════╝

? Compare all complete. What next?
  ✕ Exit
  🔍 Drill-down: view resource details
  💾 Save results to file
  📊 Compare: verify current state
  🔄 Sync: migrate data
  📦 Snap: manage snapshots
→ Pick: 🔄 Sync: migrate data
```

```text
? What do you want to migrate?
  Full migration (cases + shared steps)    ★
  Suites
  Sections
  Shared steps
→ Pick: Full migration

# Automatically pre-filled: --src-project 30 --dst-project 34
# Only the suite still needs to be picked:

? Select SOURCE suite:
→ ID: 20069 | Temporary case suite

? Select DESTINATION suite:
→ ID: 19859 | R189 Scenarios (migration)

? 📦 Create snapshot before migration? (recommended) [Y/n] Y
? Continue? [y/N] y

✓ Migration complete
```

---

## Variation 7: Dry-run before the migration

**When to use:** always recommended before a real migration.

```bash
gotr sync full --dry-run
```

The prompts are the same (project and suite selection), but:

- No data is created
- The full migration plan is shown
- There is no `Continue?` prompt
- There is no post-action menu with rollback

```text
? Select SOURCE project:
→ ID: 30 | R189

? Select SOURCE suite:
→ ID: 20069 | Temporary case suite

? Select DESTINATION project:
→ ID: 34 | E2E Scenario Testing

? Select DESTINATION suite:
→ ID: 19859 | R189 Scenarios (migration)

  [DRY-RUN] Migration Plan:
  Shared steps: 12 would be created, 3 existing
  Cases: 47 would be migrated
  No changes were made.
```

---

## Practical example: R189 → E2E migration task

A real task with five steps. It demonstrates the full workflow through the interactive mode.

### Inputs

| Item | ID | Description |
| --- | --- | --- |
| Source project | 30 | R189 |
| Source suite | S20069 | Temporary case suite with shared steps to migrate |
| Destination project | 34 | E2E Scenario Testing |
| Destination suite | S19859 | R189 Scenarios (migration from the project of the same name) |

### The task

1. Investigate the shared steps of project R189 — export and analyse them
2. Filter the shared steps by their usage in cases of suite S20069
3. Import the filtered shared steps into the destination project
4. Export the cases of suite S20069
5. Import the cases into the destination suite S19859

### Step 1. Discovery — exporting and analysing shared steps

```bash
gotr get sharedsteps
```

```text
? Select project:
→ ID: 30 | R189

[Table with all shared steps of project R189]
```

To save them to a file:

```bash
gotr export sharedsteps -p 30 --save --format json
# → Saved: ~/.gotr/exports/export/sharedsteps_30_2026-04-18_15-00-00.json
```

### Step 2. Export the cases of the suite for analysis

```bash
gotr export cases -p 30 -s 20069 --save --format json
# → Saved: ~/.gotr/exports/export/cases_30_20069_2026-04-18_15-01-00.json
```

Or interactively:

```bash
gotr export
```

```text
? Select export resource:
→ cases

? Select export endpoint:
→ get_cases

? Enter main ID:
→ 30

# Add --suite-id 20069 --save to persist the result
```

### Step 3. Migrate shared steps with filtering

```bash
gotr sync shared-steps
```

```text
? Select SOURCE project (copy shared steps from):
→ ID: 30 | R189

? Specify source suite? [y/N] y

? Select SOURCE suite:
→ ID: 20069 | Temporary case suite

? Select DESTINATION project (copy shared steps to):
→ ID: 34 | E2E Scenario Testing

  Shared Steps Summary:
  Total in R189: 45
  Used by cases in suite 20069: 15
  Already in E2E: 3 (by title match)
  To import: 12

? 📦 Create snapshot before migration? (recommended) [Y/n] Y
? 🏷  Snapshot label: shared-steps R189→E2E

? Continue? [y/N] y

✓ 12 shared steps created, 3 mapped as existing

? Save mapping? [y/N] y
  → shared_steps_mapping_2026-04-18_15-05-00.json

? Save filtered shared steps list? [y/N] y
  → shared_steps_filtered_2026-04-18_15-05-00.json
```

> [!IMPORTANT]
> The mapping file will be needed in the next step to substitute `shared_step_id` inside the cases.

### Step 4. Migrate cases with shared_step_id substitution

```bash
gotr sync cases --mapping-file shared_steps_mapping_2026-04-18_15-05-00.json
```

```text
? Select SOURCE project (copy from):
→ ID: 30 | R189

? Select SOURCE suite:
→ ID: 20069 | Temporary case suite

? Select DESTINATION project (copy to):
→ ID: 34 | E2E Scenario Testing

? Select DESTINATION suite:
→ ID: 19859 | R189 Scenarios (migration)

  Cases Migration Plan:
  Cases to migrate: 47
  With shared_step_id replacement: 15 (using mapping file)
  Without shared steps: 32

? 📦 Create snapshot before migration? (recommended) [Y/n] Y
? 🏷  Snapshot label: cases R189→E2E

? Continue? [y/N] y

✓ 47 cases created (15 with remapped shared_step_id)
```

### Step 5. Verify the result

```bash
gotr compare all --pid1 30 --pid2 34
```

```text
  Compare All Results:
  Cases:        147 vs 147 (match for suite 20069 → 19859)
  Shared steps:  45 vs 45  (match)
```

Or via the post-action menu:

```text
? What next?
→ 📊 Compare: verify current state

# The projects are filled in automatically (30 and 34)
```

### Alternative: everything in a single command

If step-by-step control is not needed, the same task can be solved with a single command:

```bash
gotr sync full
# → Pick R189 → S20069 → E2E → S19859
# → Confirm → Done
```

Or fully non-interactively:

```bash
gotr sync full \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 34 \
  --dst-suite 19859 \
  --approve --save-mapping
```

## Migration rollback: full and partial

If, after a migration, you need to remove the changes applied to the target project,
use rollback through a snapshot.

### Option A: immediately after the migration (from the post-action menu)

After `sync full` / `sync cases` / `sync shared-steps` / `sync sections` / `sync suites`:

```text
? Post-migration:
  ✕ Exit
  ↻ Rollback this migration
  ...
→ Pick: ↻ Rollback this migration
```

### Option B: later, by snapshot ID

```bash
# 1) Find the snapshot
gotr snap list

# 2) Inspect the details
gotr snap info <snapshot_id>

# 3) Dry-run before rolling back
gotr snap rollback <snapshot_id> --dry-run

# 4) Execute the rollback
gotr snap rollback <snapshot_id>
```

### Partial rollback

You can roll back only a subset of the target objects of the current snapshot:

```bash
gotr snap rollback <snapshot_id> --entity-ids 12345,12346
```

### What is rolled back per command

- `sync full`: removes the cases and shared steps that this run created
- `sync cases`: removes the cases that were created
- `sync shared-steps`: removes the created shared steps (`existing` duplicates are not touched)
- `sync sections`: removes the sections that were created
- `sync suites`: removes the suites that were created

### Rollback boundaries

- Rollback does not delete objects that existed in the target before the migration.
- If some objects were already removed manually, rollback marks the corresponding step as processed/partial.
- Rollback can safely be re-run for items that were not finished.

---

## Migration variations map

| Variation | Command(s) | Shared steps | Cases | Sections | Idea |
| --- | --- | --- | --- | --- | --- |
| Full in one pass | `sync full` | ✓ auto | ✓ auto | — | Simplest path |
| Two-step | `sync shared-steps` → `sync cases` | ✓ manual | ✓ with mapping | — | Control over each phase |
| Shared steps only | `sync shared-steps` | ✓ | — | — | Mapping preparation |
| Cases only | `sync cases` | — | ✓ (with/without mapping) | — | Cases without shared steps |
| Full pipeline | `suites` → `sections` → `shared-steps` → `cases` | ✓ | ✓ | ✓ | Maximum control |
| Discovery → migration | `compare all` → cross-navigation → `sync` | depends | depends | depends | Analysis first |

## Tips

- **Always start with `--dry-run`** — you will see the plan without making changes
- **Persist the mapping** (`--save-mapping`) — it is useful for repeated migrations
- **Use snapshots** — they let you roll back the migration through the post-action menu
- **Cross-navigation** — after compare you can jump straight into sync with the projects pre-filled
- **Hybrid mode** — supply the IDs you know via flags, pick the rest interactively
- **`--non-interactive`** — for CI/CD and scripts; every parameter must be provided via flags

---

← [Instructions](index.md) · [Full Migration](migration-full.md) · [Interactive Mode](../interactive-mode.md)