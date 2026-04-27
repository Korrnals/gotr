# Operator card: live migration run

Language: [Русский](../../../ru/guides/instructions/live-migration-operator-card.md) | English

> ⚠️ Critical: execute the steps strictly in order.
>
> ⚠️ Do nothing without explicit confirmation from the responsible person.
>
> ⚠️ Only `[GOTR-TEST]` projects. No production projects.

## Execution mode (current)

- The operator only confirms steps and strategy changes.
- All interactive form fields (`name/title/description/refs` etc.) are filled in automatically by the executor.
- For the `Create snapshot`, `Continue`, `Save mapping` questions — answer `Yes` by default unless the step is marked as an exception.

Current fixed projects:

- `SRC_PROJECT_ID=48` (`Test1`)
- `DST_PROJECT_ID=49` (`Test2`)

## Variables (fill in once before starting)

Copy the block below, plug in the values and run it in the terminal.

```bash
# Project IDs
export SRC_PROJECT_ID=""
export DST_PROJECT_ID=""

# Suite/section IDs in SRC
export SRC_SUITE_CORE_ID=""
export SRC_SUITE_EDGE_ID=""
export SRC_SECTION_1_ID=""
export SRC_SECTION_2_ID=""
export SRC_SECTION_3_ID=""

# Suite ID in DST (for the migration)
export DST_SUITE_CORE_ID=""

# Snapshot before migration
export SNAP_ID_BEFORE=""

# Test entity prefix
export TEST_PREFIX="[GOTR-TEST]"
```

Check that the variables are set:

```bash
echo "SRC_PROJECT_ID=$SRC_PROJECT_ID"
echo "DST_PROJECT_ID=$DST_PROJECT_ID"
echo "SRC_SUITE_CORE_ID=$SRC_SUITE_CORE_ID"
echo "SRC_SECTION_1_ID=$SRC_SECTION_1_ID"
echo "DST_SUITE_CORE_ID=$DST_SUITE_CORE_ID"
echo "SNAP_ID_BEFORE=$SNAP_ID_BEFORE"
```

## 0) Preparation

- [ ] Fresh binary built
- [ ] Access verified: `./gotr self-test`
- [ ] Write rights confirmed (create/update/delete)
- [ ] Start time recorded

```bash
make build
./gotr version
./gotr self-test
```

## 1) Create 2 test projects (interactive)

- [ ] SRC project `[GOTR-TEST] SRC Migration Source` created
- [ ] DST project `[GOTR-TEST] DST Migration Target` created
- [ ] `SRC_PROJECT_ID` and `DST_PROJECT_ID` saved

```bash
./gotr add project -i
./gotr add project -i
```

## 2) Populate SRC with data (interactive)

- [ ] 3 shared steps
- [ ] 2 suites
- [ ] 3 sections
- [ ] 7 cases
- [ ] All entities with the `[GOTR-TEST]` prefix
- [ ] Descriptions contain the note:
- [ ] `⚠️ Created automatically by the gotr tool for migration testing. Safe to delete.`

```bash
./gotr add shared-step $SRC_PROJECT_ID -i
./gotr add shared-step $SRC_PROJECT_ID -i
./gotr add shared-step $SRC_PROJECT_ID -i

./gotr add suite $SRC_PROJECT_ID -i
./gotr add suite $SRC_PROJECT_ID -i

./gotr add section $SRC_PROJECT_ID -i
./gotr add section $SRC_PROJECT_ID -i
./gotr add section $SRC_PROJECT_ID -i

./gotr add case $SRC_SECTION_1_ID -i
./gotr add case $SRC_SECTION_1_ID -i
./gotr add case $SRC_SECTION_1_ID -i
./gotr add case $SRC_SECTION_2_ID -i
./gotr add case $SRC_SECTION_2_ID -i
./gotr add case $SRC_SECTION_3_ID -i
./gotr add case $SRC_SECTION_3_ID -i
```

## 3) Minimal data in DST (interactive)

- [ ] A suite with the same name as in SRC was created
- [ ] One shared step with the same name as in SRC was created

```bash
./gotr add suite $DST_PROJECT_ID -i
./gotr add shared-step $DST_PROJECT_ID -i
```

## 4) Snapshot before the migration

- [ ] Snapshot created
- [ ] `SNAP_ID_BEFORE` saved

```bash
./gotr snap create -i
```

## 5) Full migration (interactive)

- [ ] `sync full -i` executed
- [ ] All critical-action prompts confirmed
- [ ] The log contains `Filter result`
- [ ] The log contains `Migration summary`
- [ ] The log contains `Migration report saved: ...`

```bash
./gotr sync full -i
```

## 6) Report check

- [ ] Report list obtained
- [ ] `latest` opened
- [ ] source/destination, stats, snapshot reference checked
- [ ] Path to the report saved

```bash
./gotr report list
./gotr report view latest
```

## 7) Rollback (interactive)

- [ ] Rollback launched
- [ ] Rollback to `SNAP_ID_BEFORE` confirmed
- [ ] Verified that DST is back in its initial state

```bash
./gotr snap rollback -i
```

## 8) Cleanup (only after explicit OK)

- [ ] Written approval for deletion received
- [ ] Test entities `[GOTR-TEST]` inside `48/49` removed
- [ ] Projects `48/49` themselves were not deleted
- [ ] Verified that no test suites/shared steps remain in `48/49`

```bash
# Example control check after cleaning up the entities
./gotr get suites 48 --non-interactive --format json
./gotr get suites 49 --non-interactive --format json
./gotr get sharedsteps 48 --non-interactive --format json
./gotr get sharedsteps 49 --non-interactive --format json
```

## 9) Re-run after fixing bugs

- [ ] Full interactive route (`-i`) repeated with auto-filled fields
- [ ] Full non-interactive route (`--non-interactive` + flags) repeated
- [ ] Results of the two runs compared (filter summary, migration summary, snapshot, report)

## Stop rules

- [ ] Any 4xx/5xx API error -> immediate stop
- [ ] No snapshot before migration -> migration forbidden
- [ ] No confirmation from the responsible person -> destructive steps forbidden
- [ ] Any doubt about the project being correct -> stop and validate manually in the UI
