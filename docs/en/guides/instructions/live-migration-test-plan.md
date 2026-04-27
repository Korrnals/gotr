# Live migration test plan on the production server

Language: [Русский](../../../ru/guides/instructions/live-migration-test-plan.md) | English

> ⚠️ **CRITICALLY IMPORTANT — READ BEFORE YOU START**
>
> This document describes the mandatory verification plan for the `gotr` tool against a real TestRail server
> **before any production data migration**.
>
> - **Every step is executed strictly in order**
> - **Nothing is done without explicit confirmation from the responsible person**
> - **Any deviation from the plan — immediate stop**
> - **At the slightest doubt — stop, consult, only then continue**
>
> Failing to follow these rules can cause irreversible data loss in production projects.

---

## Document goal

A standardised step-by-step test plan to verify the full gotr migration cycle:
test data creation → full migration → snapshot → rollback → cleanup.

It is performed **once before the production migration** on isolated test projects.
Once it passes successfully, the real migration may proceed.

A short version for the operator during the live run:
- [Live-run operator card](live-migration-operator-card.md)

---

## Current campaign profile

For the current verification series fixed test projects are used:

- SRC: `Test1` (ID: `48`)
- DST: `Test2` (ID: `49`)

Rules for this profile:

- Projects `48/49` are **not deleted** at the end of the step
- Only test entities inside the projects are cleaned up
- After fixing the bugs that were found, a **repeat full run** is performed:
  - interactive mode
  - non-interactive mode

---

## Self-driven interactive run mode

During interactive steps:

- the `name/title/description/refs` fields and other content fields are filled in automatically by the executor
- the operator does not spend time agreeing on text fields
- the operator only confirms **stages** and strategy changes

Default policy for prompt answers:

- `Create snapshot before migration?` → `Yes`
- `Continue?` → `Yes` (unless a stop is planned)
- `Save mapping?` → `Yes`
- `Save result to file?` → `Yes` (or `No` per scenario, but uniformly within the step)

Goal: test the logic and reliability of chains, not manual text entry.

---

## Preconditions

### Technical

- [ ] A fresh binary is built: `make build` → check version `./gotr version`
- [ ] An API key with **administrator** rights (create/delete projects, suite, section, case, shared steps)
- [ ] Server is reachable: `./gotr self-test` → all checks green
- [ ] Free space in `~/.gotr/` for snapshots (at least 100 MB)

### Organisational

- [ ] The person responsible for the test is identified and present
- [ ] No parallel migrations are running on the server at this time
- [ ] The test start time is recorded

---

## Fully interactive route (no `--non-interactive`)

If the goal of verification is exactly the gotr interactive UX, use this end-to-end route.

Key rule:
- All commands below use only `-i` (interactive mode)
- The `--non-interactive` flag is not used in this scenario

```bash
# Step 1: create 2 test projects
./gotr add project -i
./gotr add project -i

# Step 2: populate SRC with data
./gotr add shared-step <SRC_PROJECT_ID> -i
./gotr add shared-step <SRC_PROJECT_ID> -i
./gotr add shared-step <SRC_PROJECT_ID> -i

./gotr add suite <SRC_PROJECT_ID> -i
./gotr add suite <SRC_PROJECT_ID> -i

./gotr add section <SRC_PROJECT_ID> -i
./gotr add section <SRC_PROJECT_ID> -i
./gotr add section <SRC_PROJECT_ID> -i

./gotr add case <SRC_SECTION_ID_1> -i
./gotr add case <SRC_SECTION_ID_1> -i
./gotr add case <SRC_SECTION_ID_1> -i
./gotr add case <SRC_SECTION_ID_2> -i
./gotr add case <SRC_SECTION_ID_2> -i
./gotr add case <SRC_SECTION_ID_3> -i
./gotr add case <SRC_SECTION_ID_3> -i

# Minimal data in DST
./gotr add suite <DST_PROJECT_ID> -i
./gotr add shared-step <DST_PROJECT_ID> -i

# Step 3: snapshot before migration
./gotr snap create -i

# Step 4: full migration (interactive)
./gotr sync full -i

# Step 5: report
./gotr report list
./gotr report view latest

# Step 6: rollback (interactive)
./gotr snap rollback -i

# Step 7: delete test projects (only after explicit OK)
./gotr delete project <SRC_PROJECT_ID> -i
./gotr delete project <DST_PROJECT_ID> -i
```

What must be recorded during the interactive route:
- IDs of the created projects, suites, sections, cases, shared steps
- ID of the snapshot taken before the migration
- Path to the migration report (`~/.gotr/reports/migration-*.md`)
- All `Yes/No` confirmations on critical steps

---

## Step 0 — Baseline: capture the server state

**Goal:** know the exact state before the test so that everything can be returned afterwards.

```bash
# Get the full list of projects (save the output!)
./gotr get projects --non-interactive --format json > /tmp/gotr-test-baseline-projects.json

# Make sure no test projects exist yet
cat /tmp/gotr-test-baseline-projects.json | grep -i "GOTR-TEST"
# Expected result: empty output
```

**⛔ Stop:** if `GOTR-TEST` is already in the list — find out where it came from before continuing.

---

## Step 1 — Create the test projects

**Two isolated projects are created exclusively for the test.**

All names contain the prefix `[GOTR-TEST]` so that they can be unambiguously identified and removed.

```bash
# Create SRC (migration source)
./gotr add project \
  --name "[GOTR-TEST] SRC Migration Source" \
  --announcement "⚠️ Test SOURCE project for verifying gotr migration. Created automatically. Delete after testing." \
  --non-interactive

# Save the project ID from the output! For example: "id": 99
export GOTR_TEST_SRC_ID=<id from output>

# Create DST (migration target)
./gotr add project \
  --name "[GOTR-TEST] DST Migration Target" \
  --announcement "⚠️ Test TARGET project for verifying gotr migration. Created automatically. Delete after testing." \
  --non-interactive

export GOTR_TEST_DST_ID=<id from output>

echo "SRC=$GOTR_TEST_SRC_ID  DST=$GOTR_TEST_DST_ID"
```

**Verification:**
```bash
./gotr get project $GOTR_TEST_SRC_ID --non-interactive
./gotr get project $GOTR_TEST_DST_ID --non-interactive
```

**⛔ Stop:** both projects must be visible and have the correct names.

---

## Step 2 — Populate SRC with test data

**Population principle:**
- Some entities **share the same name** in SRC and DST — to test the "already exists / skip" logic
- Some entities are **unique to SRC** — to test importing new ones
- All names use the `[GOTR-TEST]` prefix

### 2.1 — Shared steps (first, since they are needed when creating cases)

```bash
# Shared step 1 — will match in DST
./gotr add shared-step $GOTR_TEST_SRC_ID \
  --title "[GOTR-TEST] Open browser and navigate" \
  --non-interactive

# Shared step 2 — unique to SRC
./gotr add shared-step $GOTR_TEST_SRC_ID \
  --title "[GOTR-TEST] Fill login form with credentials" \
  --non-interactive

# Shared step 3 — unique to SRC
./gotr add shared-step $GOTR_TEST_SRC_ID \
  --title "[GOTR-TEST] Verify success notification" \
  --non-interactive
```

Save the IDs of all shared steps from the output.

### 2.2 — Suites in SRC

```bash
# Suite 1 — primary (will match the one we create in DST)
./gotr add suite $GOTR_TEST_SRC_ID \
  --name "[GOTR-TEST] Core Functionality" \
  --description "⚠️ gotr test suite. Delete after the test." \
  --non-interactive

export GOTR_TEST_SRC_SUITE1_ID=<id from output>

# Suite 2 — unique to SRC
./gotr add suite $GOTR_TEST_SRC_ID \
  --name "[GOTR-TEST] Edge Cases" \
  --description "⚠️ gotr test suite. Delete after the test." \
  --non-interactive

export GOTR_TEST_SRC_SUITE2_ID=<id from output>
```

### 2.3 — Sections

```bash
# In Suite 1
./gotr add section $GOTR_TEST_SRC_ID \
  --suite-id $GOTR_TEST_SRC_SUITE1_ID \
  --name "[GOTR-TEST] Authentication" \
  --non-interactive
export GOTR_TEST_SRC_SEC_AUTH=<id>

./gotr add section $GOTR_TEST_SRC_ID \
  --suite-id $GOTR_TEST_SRC_SUITE1_ID \
  --name "[GOTR-TEST] Payments" \
  --non-interactive
export GOTR_TEST_SRC_SEC_PAY=<id>

# In Suite 2
./gotr add section $GOTR_TEST_SRC_ID \
  --suite-id $GOTR_TEST_SRC_SUITE2_ID \
  --name "[GOTR-TEST] Boundary Values" \
  --non-interactive
export GOTR_TEST_SRC_SEC_BV=<id>
```

### 2.4 — Test cases

```bash
# Auth section — 3 cases
./gotr add case $GOTR_TEST_SRC_SEC_AUTH \
  --title "[GOTR-TEST] Login with valid credentials" \
  --non-interactive

./gotr add case $GOTR_TEST_SRC_SEC_AUTH \
  --title "[GOTR-TEST] Login with invalid password" \
  --non-interactive

./gotr add case $GOTR_TEST_SRC_SEC_AUTH \
  --title "[GOTR-TEST] Logout" \
  --non-interactive

# Payments section — 2 cases
./gotr add case $GOTR_TEST_SRC_SEC_PAY \
  --title "[GOTR-TEST] Successful payment" \
  --non-interactive

./gotr add case $GOTR_TEST_SRC_SEC_PAY \
  --title "[GOTR-TEST] Payment with insufficient funds" \
  --non-interactive

# Edge Cases section — 2 cases
./gotr add case $GOTR_TEST_SRC_SEC_BV \
  --title "[GOTR-TEST] Max length input" \
  --non-interactive

./gotr add case $GOTR_TEST_SRC_SEC_BV \
  --title "[GOTR-TEST] Empty input validation" \
  --non-interactive
```

### 2.5 — Minimal data in DST (to test "already exists")

```bash
# Suite with the same name as in SRC — so sync can use it as the DST
./gotr add suite $GOTR_TEST_DST_ID \
  --name "[GOTR-TEST] Core Functionality" \
  --non-interactive
export GOTR_TEST_DST_SUITE_ID=<id>

# One shared step with a matching name
./gotr add shared-step $GOTR_TEST_DST_ID \
  --title "[GOTR-TEST] Open browser and navigate" \
  --non-interactive
```

**Population check:**
```bash
./gotr get suites $GOTR_TEST_SRC_ID --non-interactive
./gotr get cases $GOTR_TEST_SRC_ID $GOTR_TEST_SRC_SUITE1_ID --non-interactive | grep -c '"id"'
./gotr get sharedsteps $GOTR_TEST_SRC_ID --non-interactive | grep -c '"id"'
# Expected: 2 suites, 7 cases, 3 shared steps in SRC
```

**⛔ Stop:** before the next step make sure all data was created correctly.

---

## Step 3 — Take a snapshot of the DST state before the migration

```bash
./gotr snap create \
  --snap-label "before-gotr-test-migration" \
  --non-interactive
```

Save the snapshot ID from the output — it will be needed for rollback.

---

## Step 4A — Full migration SRC → DST (flag-based / semi-automatic mode)

```bash
./gotr sync full \
  --src-project $GOTR_TEST_SRC_ID \
  --src-suite $GOTR_TEST_SRC_SUITE1_ID \
  --dst-project $GOTR_TEST_DST_ID \
  --dst-suite $GOTR_TEST_DST_SUITE_ID \
  --approve \
  --save-mapping \
  --snapshot
```

This variant is convenient for CI and repeatable runs.

---

## Step 4B — Full migration SRC → DST (fully interactive mode)

> Use this when the goal is to test the CLI dialog/prompt behaviour itself.
>
> May be run as the main scenario instead of step 4A.

```bash
./gotr sync full -i
```

### Expected interactive scenario

- [ ] SRC project chosen: `[GOTR-TEST] SRC Migration Source`
- [ ] SRC suite chosen: `[GOTR-TEST] Core Functionality`
- [ ] DST project chosen: `[GOTR-TEST] DST Migration Target`
- [ ] DST suite chosen: `[GOTR-TEST] Core Functionality`
- [ ] Import confirmation answered with `Yes`
- [ ] Save mapping confirmation answered with `Yes`
- [ ] Create snapshot confirmation answered with `Yes`
- [ ] Final migration report produced

### What to record in the logs during the interactive run

- [ ] The `Filter result` block (source/target/matched/new)
- [ ] The `Migration summary` block
- [ ] A line like `Snapshot saved: ...`
- [ ] A line like `Migration report saved: ~/.gotr/reports/migration-*.md`

### Minimal example of prompt answers

Below is a template; the actual prompt wording may differ slightly between CLI versions:

```text
? Select source project: [GOTR-TEST] SRC Migration Source
? Select source suite: [GOTR-TEST] Core Functionality
? Select destination project: [GOTR-TEST] DST Migration Target
? Select destination suite: [GOTR-TEST] Core Functionality
? Continue with migration? Yes
? Save mapping file? Yes
? Save snapshot before/after migration? Yes
```

After step 4B is done, the remaining checks and steps (report, rollback, cleanup) are performed exactly as described below.

**Verify during the run:**
- [ ] The filter result block was shown (source/target/matched/new)
- [ ] The migration summary was shown before import
- [ ] A snapshot was created automatically
- [ ] At the end: `Migration report saved: ~/.gotr/reports/migration-*.md`

**Verification afterwards:**
```bash
./gotr get cases $GOTR_TEST_DST_ID $GOTR_TEST_DST_SUITE_ID --non-interactive | grep -c '"id"'
# Expected: cases from SRC are migrated

./gotr get sharedsteps $GOTR_TEST_DST_ID --non-interactive | grep -c '"id"'
# Shared steps 2 and 3 must appear; shared step 1 was already there
```

**⛔ Stop:** compare the number of cases in DST with what was in SRC.

---

## Step 5 — Review the migration report

```bash
# List all reports
./gotr report list

# View the latest one
./gotr report view latest
```

**What to check in the report:**
- [ ] Migration type is correct (`full`)
- [ ] Source and Destination projects are right
- [ ] Statistics: created / matched / failed
- [ ] A snapshot reference is present
- [ ] Execution time is shown

Copy the report file path and open it in a browser / hand it over for review.

---

## Step 6 — Roll back the migration via the snapshot

```bash
# Show the list of snapshots
./gotr snap list

# Roll back to the "before migration" snapshot
./gotr snap rollback --snap-id <snap-id from step 3>
```

**Verification after rollback:**
```bash
./gotr get cases $GOTR_TEST_DST_ID $GOTR_TEST_DST_SUITE_ID --non-interactive | grep -c '"id"'
# Expected: DST returned to the state BEFORE migration (0 or only the cases that were there)

./gotr get sharedsteps $GOTR_TEST_DST_ID --non-interactive | grep -c '"id"'
# Expected: only 1 shared step (the one that was there before)
```

**⛔ Stop:** make absolutely sure the rollback succeeded before moving on to deletion.

---

## Step 7 — Cleaning up test entities (without deleting the projects)

> ⛔ **PERFORMED ONLY WITH EXPLICIT WRITTEN APPROVAL FROM THE RESPONSIBLE PERSON**
>
> For the current profile (`Test1=48`, `Test2=49`) the projects themselves are not deleted.

Only test entities with the `[GOTR-TEST]` prefix are removed:

```bash
# Roughly in this order:
# 1) runs/plans (if they were created)
# 2) cases
# 3) sections
# 4) suites
# 5) shared steps (if any standalone ones)

# Make sure no test suites/shared steps remain in the projects
./gotr get suites 48 --non-interactive --format json
./gotr get suites 49 --non-interactive --format json
./gotr get sharedsteps 48 --non-interactive --format json
./gotr get sharedsteps 49 --non-interactive --format json
```

**Expected result:** the responses contain no test entities from the current campaign.

---

## Step 8 — Repeat full run after fixing the bugs

After the defects that were found are fixed, run again:

1. The full interactive route (`-i`) with auto-filled fields
2. The full non-interactive route (`--non-interactive`) with explicit flags

Minimal command set for the rerun:

```bash
# Interactive
./gotr sync suites
./gotr sync sections
./gotr sync shared-steps
./gotr sync cases
./gotr sync full

# Non-interactive (example)
./gotr sync suites --src-project 48 --dst-project 49 --approve --save-mapping --snapshot
./gotr sync sections --src-project 48 --src-suite <SRC_SUITE> --dst-project 49 --dst-suite <DST_SUITE> --approve --save-mapping --snapshot
./gotr sync shared-steps --src-project 48 --dst-project 49 --approve --save-mapping --snapshot
./gotr sync cases --src-project 48 --src-suite <SRC_SUITE> --dst-project 49 --dst-suite <DST_SUITE> --snapshot
./gotr sync full --src-project 48 --src-suite <SRC_SUITE> --dst-project 49 --dst-suite <DST_SUITE> --approve --save-mapping --snapshot
```

---

## Checklist: success criteria for the test

| # | Criterion | Result |
|---|---|---|
| 1 | Binary built without errors | ✓/✗ |
| 2 | Self-test green | ✓/✗ |
| 3 | Both test projects created | ✓/✗ |
| 4 | SRC populated: 2 suites, 7 cases, 3 shared steps | ✓/✗ |
| 5 | DST snapshot taken before the migration | ✓/✗ |
| 6 | Migration done (step 4A or 4B), cases migrated | ✓/✗ |
| 7 | Interactive scenario (step 4B) passes without errors | ✓/✗ |
| 8 | Migration report saved and readable | ✓/✗ |
| 9 | Rollback via snapshot completed correctly | ✓/✗ |
| 10 | DST returned to the pre-migration state | ✓/✗ |
| 11 | Test entities cleaned up, projects 48/49 preserved | ✓/✗ |
| 12 | Re-run after fixes: interactive + non-interactive | ✓/✗ |

**Only when all marks are ✓ — the real migration may begin.**

---

## Strict execution rules

> These rules are not recommendations. They are mandatory conditions.

- **No actions without an explicit "OK" from the responsible person** before each destructive step (delete, rollback, overwrite)
- **No parallel execution** — steps are strictly sequential
- **Any API error — immediate stop**, root-cause analysis, only then continue
- **Do not use production projects** as SRC or DST — only specially created `[GOTR-TEST]` projects
- **Snapshot before migration is mandatory** — without it there is no rollback
- **Save the output of each command** — for analysis if anything goes wrong
- **Do not rush** — better to spend an hour on the test than to lose data in production

---

## Recovery on failure

If something goes wrong on any step:

1. **Do not panic** — record the exact error message
2. Check the snapshots: `./gotr snap list`
3. If a snapshot exists — roll back: `./gotr snap rollback --snap-id <id>`
4. If there is no snapshot — contact the TestRail administrator
5. Inspect the DST state manually in the web UI

---

*Document created within the gotr project. The current version is in the repository: `docs/ru/guides/instructions/live-migration-test-plan.md`*
