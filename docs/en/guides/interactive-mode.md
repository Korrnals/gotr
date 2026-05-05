# Interactive Mode

Language: [Русский](../../ru/guides/interactive-mode.md) | English

## Navigation

- [Documentation](../index.md)
  - [Guides](index.md)
    - [Installation](installation.md)
    - [Configuration](configuration.md)
    - [Interactive Mode](interactive-mode.md)
    - [Progress](progress.md)
    - [Commands Index](commands/index.md)
    - [Instructions](instructions/index.md)
  - [Architecture](../architecture/index.md)
  - [Operations](../operations/index.md)
  - [Reports](../reports/index.md)
- [Home](../../../README.md)

## How it works

If a required parameter is missing, the tool automatically:

1. Fetches the list of available entities from the API
2. Shows an interactive selection menu with navigation
3. Asks the user to pick an item from the list
4. Uses the chosen value and proceeds to the next step

Three operating modes:

- **Auto-interactive** (default) — prompts only appear for unspecified parameters
- **Manual** — every parameter passed via flags, no prompts
- **Non-interactive** (`--non-interactive`) — prompts are forbidden; the command errors out if input is required (CI/CD-friendly)

## Navigation in interactive menus

Every interactive list contains navigation items:

```text
? Select project:
  ← Back                    ← return to the previous step
  ✕ Exit                    ← exit the interactive flow
  ──────────────────────
  ID: 1  | SAP Hybris
  ID: 2  | SAP CRM
  ID: 30 | R189
  ...
  ← Back                    ← duplicated at the bottom of the list (when >5 items)
```

- **← Back** — return to the previous selection step (when available)
- **✕ Exit** — finish without an error
- Items are rendered with column alignment (ID + name)
- If a project has **only one suite**, it is selected automatically (no prompt)

## Commands with interactive mode

### get — read data

```bash
# Fully interactive
gotr get cases
# → Select project: → Select suite: → [JSON with cases]

# Partially interactive (project specified)
gotr get cases 30
# → Select suite: → [JSON with cases]

# Fully manual
gotr get cases 30 --suite-id 20069
```

Interactive mode is available for:

| Command | Prompt 1 | Prompt 2 | Prompt 3 |
| --- | --- | --- | --- |
| `get cases` | Select project | Select suite | — |
| `get case` | Select project | Select suite | Select case |
| `get suites` | Select project | — | — |
| `get suite` | Select project | Select suite | — |
| `get sharedsteps` | Select project | — | — |
| `get sharedstep` | Select project | Select shared step | — |
| `get case-history` | Select project | Select suite → Select case | — |
| `get sharedstep-history` | Select project | Select shared step | — |
| `get project` | Select project | — | — |
| `get sections list` | Select project | — | — |

`get cases` specifics:

- If a project has **only one suite**, it is auto-selected
- The `--all-suites` flag fetches cases from every suite (skipping the picker)

### export — export data

```bash
# Fully interactive
gotr export
# → Select export resource: → Select export endpoint: → [result]

# Partial
gotr export cases get_cases 30 --suite-id 20069 --save --format json
```

Export prompts (each next prompt only when not specified):

1. `Select export resource:` — pick the resource type (cases, suites, sharedsteps...)
2. `Select export endpoint:` — pick the API endpoint
3. `Enter main ID:` — provide the ID (when the endpoint contains `{id}`)

### compare — project comparison

```bash
# Fully interactive
gotr compare cases
# → Select first project (pid1): → Select second project (pid2):
# → [comparison result]
# → Comparison complete. What next?

# Manual
gotr compare all --pid1 30 --pid2 34 --save
```

Compare prompts:

1. `Select first project (pid1):` — pick the first project
2. `Select second project (pid2):` — pick the second one (← Back returns to step 1)
3. `Save compare result to file?` — offer to save (when the comparison was started interactively)

### sync — data migration

All `sync` subcommands support fully interactive mode.

#### sync full

```bash
gotr sync full
# → Select SOURCE project:
# → Select SOURCE suite:
# → Select DESTINATION project:
# → Select DESTINATION suite:
# → 📦 Create snapshot before migration? (recommended) [Y/n]
# → 🏷  Snapshot label (optional, press Enter to skip):
# → [migration summary]
# → Continue? [y/N]
# → [execution: shared steps → cases]
# → [post-action menu]
```

#### sync cases

```bash
gotr sync cases
# → Select SOURCE project (copy from):
# → Select SOURCE suite:
# → Select DESTINATION project (copy to):
# → Select DESTINATION suite:
# → 📦 Create snapshot before migration? (recommended) [Y/n]
# → Continue? [y/N]
# → [case migration]
# → [post-action menu]
```

#### sync shared-steps

```bash
gotr sync shared-steps
# → Select SOURCE project (copy shared steps from):
# → Specify source suite? [y/N]
#   (if yes) → Select SOURCE suite:
# → Select DESTINATION project (copy shared steps to):
# → [filtering and summary]
# → 📦 Create snapshot before migration? (recommended) [Y/n]
# → Continue? [y/N]
# → [shared steps import]
# → Save mapping? [y/N]
# → Save filtered shared steps list? [y/N]
# → [post-action menu]
```

#### sync sections

```bash
gotr sync sections
# → Select SOURCE project:
# → Select SOURCE suite:
# → Select DESTINATION project:
# → Select DESTINATION suite:
# → 📦 Create snapshot before migration? (recommended) [Y/n]
# → Continue? [y/N]
# → [section transfer]
# → Save mapping? [y/N]
# → [post-action menu]
```

#### sync suites

```bash
gotr sync suites
# → Select SOURCE project:
# → Select DESTINATION project:
# → 📦 Create snapshot before migration? (recommended) [Y/n]
# → Continue? [y/N]
# → [suite transfer]
# → Save mapping? [y/N]
# → [post-action menu]
```

### attachments cleanup — bulk attachment cleanup

`gotr attachments cleanup` supports interactive prompts when invoked
with insufficient flags on a TTY. Flags explicitly set on the command
line are never overwritten by the wizard.

```bash
gotr attachments cleanup
# → Scope: All projects / Specific projects
#   (specific) → Project IDs (comma-separated):
# → Parent kinds preset:
#     all (case,run,plan,plan_entry,result,test)  ← default, prints ⚠️ warning
#     case · run · plan,plan_entry · result,test · custom (comma-separated)
# → Older than (e.g. 90d, 3M, 1y):
# → Concurrency:
# → 📦 Create snapshot before deletion? (recommended) [Y/n]
#   (yes) → Snapshot retention (days):
# → Dry-run first? [Y/n]
# → Final confirmation: Y/N
```

Overrides:

- `--non-interactive` — disable prompts; the command errors out if any
  required input is missing (CI-friendly).
- `--force` — skip the final confirmation prompt while still allowing
  the wizard for missing values.
- `--no-snapshot` — opt out of the rollback safety net; rollback will
  not be possible.

See [`gotr attachments cleanup`](commands/attachments.md) for the full
flag reference and rollback recipe.

## Post-action menu and cross-navigation

After sync and compare operations finish, an interactive action menu is shown.

### After sync

```text
? What next?
  ✕ Exit
  ↻ Rollback this migration          ← (only if a snapshot was created)
  📊 Compare: verify current state    ← cross-navigation → gotr compare all
  🔄 Sync: migrate data              ← cross-navigation → gotr sync full
  📦 Snap: manage snapshots          ← cross-navigation → gotr snap list
```

- **Rollback** — undoes the migration through the snapshot created earlier
- **Cross-navigation** — jumps directly to a related command without exiting

### After compare

```text
? Comparison complete. What next?
  ✕ Exit
  📋 View detailed results
  💾 Save results to file
  → Sync: migrate differences        ← (when there are differences)
  📊 Compare: verify current state
  🔄 Sync: migrate data
  📦 Snap: manage snapshots
```

- **Sync: migrate differences** — launches sync, passing project IDs from compare
- **Save results to file** — choice of format (json/yaml/csv/table) and save path

### Parameter inheritance (WorkSession)

The `src-project`, `dst-project`, `src-suite`, `dst-suite` parameters are passed through the session:

```text
compare → sync:  project IDs from compare are auto-fed into sync
sync → compare:  project IDs from sync are auto-fed into compare
```

This means when cross-navigating you **don't have to pick projects again** — they
are inherited from the previous command.

## Snapshot confirmation

Before mutating operations (sync), gotr offers to create a snapshot:

```text
📦 Create snapshot before migration? (recommended) [Y/n]
```

Priority logic:

1. The `--snapshot` flag — explicitly enables/disables
2. The `snap.enabled` config option — when set
3. Interactive prompt — when neither (1) nor (2) is set (default: **yes**)

When a snapshot is created, you are also offered to enter a label:

```text
🏷  Snapshot label (optional, press Enter to skip):
```

## Interactive flow examples

### Example 1: get cases

```text
$ gotr get cases

? Select project:
  ← Back
  ✕ Exit
  ──────────────────────
  ID: 1  | SAP Hybris
  ID: 2  | SAP CRM
  ...
  ID: 30 | R189
  ← Back

→ pick: ID: 30 | R189

? Select suite:
  ← Back
  ✕ Exit
  ──────────────────────
  ID: 8411  | R189 IT Suites and Cases
  ID: 9709  | R189 PT Suites and Cases
  ...
  ID: 20069 | Temporary case suite
  ← Back

→ pick: ID: 20069 | Temporary case suite

[JSON with cases]
```

### Example 2: sync full (full migration)

```text
$ gotr sync full

? Select SOURCE project:
→ ID: 30 | R189

? Select SOURCE suite:
→ ID: 20069 | Temporary case suite

? Select DESTINATION project:
→ ID: 34 | E2E Scenarios Testing

? Select DESTINATION suite:
→ ID: 19859 | R189 Scenarios (transfer)

? 📦 Create snapshot before migration? (recommended) Yes
? 🏷  Snapshot label (optional): R189 → E2E migration

  ┌─────────────────────────────────────┐
  │ Migration summary                   │
  │ Shared steps: 12 new, 3 existing    │
  │ Cases: 47 to migrate                │
  └─────────────────────────────────────┘

? Continue? Yes

✓ Shared steps migrated (12 created, 3 mapped)
✓ Cases migrated (47 created)

? What next?
→ ✕ Exit
```

### Example 3: compare → sync (cross-navigation)

```text
$ gotr compare cases --pid1 30 --pid2 34

  Cases comparison:
  Project 30: 147 cases
  Project 34: 100 cases
  Differences: 47 missing in project 34

? Comparison complete. What next?
→ → Sync: migrate differences

? What do you want to migrate?
→ Full migration (cases + shared steps)

# Projects are inherited from compare → sync full is launched
# with --src-project 30 --dst-project 34
```

## Partial interactive (hybrid mode)

You can specify **part** of the parameters via flags and pick the rest interactively:

```bash
# Only the source is set — destination is picked interactively
gotr sync full --src-project 30 --src-suite 20069

# Only projects are set — suite is picked interactively
gotr sync cases --src-project 30 --dst-project 34

# Mapping file is set, the rest is interactive
gotr sync cases --mapping-file mapping.json
```

## Benefits

1. **No need to memorise IDs** — pick from a list
2. **Visual control** — you see project and suite names
3. **Flexibility** — mix and match: some parameters via flags, some interactively
4. **Automation** — the same commands work in scripts with flags
5. **Cross-navigation** — jump between compare/sync/snap without exiting
6. **Parameter inheritance** — projects propagate between commands through the session
7. **Snapshot safety** — automatic offer to create a rollback point before migration

## Stage 12 roadmap (Interactive System Unification)

Below is the consolidated plan for the interactive mode and the move to a unified
`auto-interactive + --non-interactive` model.

### Stage 12.0 — Foundation (done)

- A single `Prompter` contract was introduced.

- Implementations were added: terminal prompter (survey/v2), non-interactive
prompter, and a mock prompter for tests.

- Context-based prompter injection into the root runtime was added.

- The global `--non-interactive` flag was added.

### Stage 12.1 — Migrate commands to the unified prompter (done)

- Sync (`sync`) was migrated to `SelectProject/SelectSuiteForProject`
and `p.Confirm`.

- `get`, `run`, `result` were migrated to `PrompterFromContext`.

- Tests moved from `os.Stdin` and hand-rolled mock wrappers to `MockPrompter`.

### Stage 12.2 — UX unification: auto-interactive (done)

- For `add`/`update`, the auto-wizard kicks in if the user did not pass
manual input flags.

- The explicit `--interactive` flag is preserved for backward compatibility.

- `--non-interactive` remains the master switch for CI/CD and automation.

### Stage 12.3 — Full test audit and coverage closeout (new stage)

Goal: close test gaps across the codebase (CLI layer first, plus
interactive/safety scenarios) so that Stage 12 meets the quality DoD.

#### 12.3.1 Coverage inventory

- Build a coverage matrix for `cmd/*`, `internal/interactive` and the
critical runtime (`internal/service`, `internal/output`, `internal/flags`).

- Collect a list of files without tests, branches without negative tests,
and commands without `--non-interactive` scenarios.

- Record baseline metrics (`go test -cover ./...` + focused
`-coverprofile`).

#### 12.3.2 Test debt prioritisation

- P0 (mandatory): mutating commands with the dry-run gate and
non-interactive gate, plus interactive selection chains
(`project -> suite -> run`) and selection errors.

- P1: auto-interactive vs manual flag toggling, regressions in error
wrapping and messages.

- P2: edge cases, empty API responses, partial data.

#### 12.3.3 Implement missing tests

- Add table-driven tests for repeating scenarios.

- Use `MockPrompter` as the standard for every interactive branch.

- Add tests for `ErrNonInteractive` at points where input is required,
absence of mutating API calls in dry-run, and a correct fallback to
manual flags without prompts.

#### 12.3.4 Verification and DoD

- Full `go test ./...` is green.

- Focused suites (`cmd/*`, `internal/interactive`) pass without
coverage regressions.

- All identified P0/P1 gaps are closed by tests and reflected in the
stage changelog.

### Stage 12.4 — Cleanup and removal of legacy compatibility wrappers (planned)

- Remove compatibility wrappers in `internal/interactive/*` that are no
longer used.

- Update standards and code samples to the new API
(`PrompterFromContext`, auto-interactive).

### Stage 12.5 — Documentation and release readiness (planned)

- Update docs on interactive mode and non-interactive operation.

- Prepare release notes with a Stage 12 change map.

- Run a final smoke check of CLI scenarios (manual + CI).

### Stage 12.6 — Smoke-check status (done)

Final smoke checks were performed against a freshly built `gotr-test`
binary with a real config and a safe constraint: only read-only
commands or `--dry-run` for mutating operations.

Performed checks:

- `go test ./... -count=1` — **PASS** (full green run).
- `gotr templates list` (interactive project selection) — **PASS**, a
  valid JSON template response was returned.
- `gotr run create --name <name> --dry-run` — interactive project and
  suite selection is reached; the previous block of
  `required flag(s) "suite-id" not set` is gone.
- `gotr result add --status-id 1 --dry-run` — interactive project and
  run selection is reached.

Note about the live runner:

- For multi-step TTY lists, feeding input via `printf | ...` may produce
  `EOF` at later selection steps. This is a non-TTY piping limitation,
  not a Stage 12 functional regression: the key interactive branches
  successfully activate and reach the target selection points.

## Command behaviour matrix

The actual behaviour is reflected in two views:

- Top-level commands and key subcommands (operational view).

- Full map across all `cmd/**` subpackages (architectural view of the
CLI layer).

Legend:

- **Auto**: automatic interactive mode when required input is missing.
- **Manual**: manual mode via explicit flags/arguments without prompts.
- **NI**: `--non-interactive`, which forbids prompts and fails the command if input is required.

| Command/subcommand | Auto | Manual | NI |
| --- | --- | --- | --- |
| `add project` | Yes (wizard) | Yes | Errors when wizard is needed |
| `add suite` | Yes (wizard + auto select project) | Yes | Errors when wizard is needed |
| `add section` | Yes (wizard + auto select project) | Yes | Errors when wizard is needed |
| `add case` | Yes (wizard + auto select section) | Yes | Errors when wizard is needed |
| `add run` | Yes (wizard + auto select project) | Yes | Errors when wizard is needed |
| `add shared-step` | Yes (wizard + auto select project) | Yes | Errors when wizard is needed |
| `add result` | Yes (project/run/test select when `test-id` missing) | Yes | Errors when selection needed |
| `add result-for-case` | Partial (project/run select when `run-id` missing; `case-id` stays manual) | Yes | Errors when selection needed |
| `add attachment` | Partial (container IDs optional in some subcommands, `file_path` manual) | Yes | Errors when selection needed |
| `update project` | Yes (wizard) | Yes | Errors when wizard is needed |
| `update suite` | Yes (wizard) | Yes | Errors when wizard is needed |
| `update section` | Yes (wizard) | Yes | Errors when wizard is needed |
| `update case` | Yes (wizard) | Yes | Errors when wizard is needed |
| `update run` | Yes (wizard) | Yes | Errors when wizard is needed |
| `update shared-step` | Yes (wizard) | Yes | Errors when wizard is needed |
| `update labels` | Partial (`labels update test` with test select; bulk/manual branches partly manual-only) | Yes | Errors when selection needed |
| `get cases` | Yes (project/suite select) | Yes | Errors when selection needed |
| `get suites`, `get sharedsteps` | Yes (project select) | Yes | Errors when selection needed |
| `run list`, `run get`, `run delete`, `run close`, `run update`, `run create` | Yes (project/run select when ID missing; `run create` also picks suite) | Yes | Errors when selection needed |
| `plans list`, `plans get`, `plans add`, `plans update`, `plans delete`, `plans close`, `plans entry add/update/delete` | Yes (project/plan/entry select when ID missing) | Yes | Errors when selection needed |
| `result list`, `result get`, `result get-case`, `result add` | Yes (project/run/test-case select when ID missing) | Yes | Errors when selection needed |
| `users get`, `users update`, `users get-by-email` | Yes (select from users list when ID/email missing) | Yes | Errors when selection needed |
| `roles get` | Yes (select from roles list when ID missing) | Yes | Errors when selection needed |
| `reports list`, `reports run`, `reports run-cross-project`, `templates list`, `bdds add/get` | Yes | Yes | Errors when selection needed |
| `sync *` | Yes (project/suite/select + confirm) | Yes | Errors when selection/confirm is needed |
| `delete` | Yes (endpoint/id select) | Yes | Errors when selection needed |
| `list` | Yes (resource select) | Yes | Errors when selection needed |
| `export` | Yes (resource/endpoint/id input) | Yes | Errors when selection needed |

## Full `cmd/**` map (all subpackages)

The source of truth for root-command registration is `cmd/commands.go`.

Legend:

- **Registered**: the subpackage is wired into `rootCmd` via `*.Register(...)`.

- **Interactive**: there is interactive logic in the package's production code.

- **Coverage**: degree of interactive coverage inside the package.

| Package `cmd/**` | Registered | Interactive | Coverage | Notes |
| --- | --- | --- | --- | --- |
| `attachments` | Yes | Yes | High | Auto in `attachments list case/plan/plan-entry/run/test`, `attachments get`, `attachments delete`, `attachments add case/plan/plan-entry/result/run` |
| `bdds` | Yes | Yes | High | Auto in `add`, `get` via case selection |
| `cases` | Yes | Yes | Partial | Auto in `cases list`, `cases get`, `cases delete`, `cases update`, `cases add`, `cases bulk`; some branches stay manual-only |
| `compare` | Yes | No | None | Manual-only |
| `configurations` | Yes | Yes | High | Auto in `list`, `add-group`, `add-config`, `update-group`, `update-config`, `delete-group`, `delete-config` |
| `datasets` | Yes | Yes | High | Auto in `list`, `add`, `get`, `update`, `delete` via project/dataset selection |
| `get` | Yes | Yes | Partial | Auto in `cases`, `case`, `suites`, `suite`, `sharedsteps`, `sharedstep`, `case-history`, `sharedstep-history`, `project`, `sections list`, `section`; the rest are manual-only |
| `groups` | Yes | Yes | High | Auto in `list`, `get`, `add`, `update`, `delete` via project/group selection |
| `labels` | Yes | Yes | Partial | Auto in `get`, `list`, `update test`, `update-label`; some bulk/manual branches need explicit lists/flags |
| `milestones` | Yes | Yes | High | Auto in `list`, `get`, `add`, `update`, `delete` via project/milestone selection |
| `plans` | Yes | Yes | High | Auto in `list`, `get`, `add`, `update`, `delete`, `close`, `entry add`, `entry update`, `entry delete` via project/plan/entry selection |
| `result` | Yes | Yes | Partial | Auto in `list`, `get`, `get-case`, `add`, `add-case`; `add-bulk`/`fields` stay manual-oriented |
| `run` | Yes | Yes | High | Auto in `list`, `get`, `delete`, `close`, `update`, `create` via project/run/suite selection |
| `sync` | Yes | Yes | High | Interactive selection/confirm chains in the main migration scenarios |
| `test` | Yes | No | None | Manual-only |
| `variables` | Yes | No | None | Manual-only |
| `reports` | Yes | Yes | High | Auto in `list`, `run`, `run-cross-project` |
| `roles` | Yes | Yes | Partial | Auto in `get`; `list` is read-only/manual |
| `templates` | Yes | Yes | Partial | Auto in `list` via project selection |
| `tests` | Yes | Yes | High | Auto in `list`, `get`, `update` via run/test selection |
| `users` | Yes | Yes | Partial | Auto in `get`, `update`, `get-by-email`; `list`/`add` are manual/read-only |
| `list` | No | No | None | Service directory; no standalone Register |
| `internal` | No | No | None | Test helpers, not CLI commands |

### Root commands in `cmd/*.go`

| Command | Interactive | Coverage | Notes |
| --- | --- | --- | --- |
| `add` | Yes | High | Wizard + auto-interactive + `--non-interactive` gate |
| `update` | Yes | High | Wizard + auto-interactive + `--non-interactive` gate |
| `delete` | Yes | High | Auto-select endpoint/id + `--non-interactive` guard |
| `list` | Yes | High | Auto-select resource + `--non-interactive` guard |
| `export` | Yes | Partial | Auto-select inputs; e2e scenarios still need closing |
| `config`, `completion`, `selftest` | No | N/A | Service commands without interactive entity selection |

## Target Auto/Manual/NI matrix across `cmd/**`

The matrix below records both the current state and the target behaviour for the
project-wide unification.

Legend:

- **Current**: current state at this stage.

- **Target**: state after the unification is complete.

- **Priority**: implementation order (`P0` -> `P1` -> `P2`).

| Package | Current (Auto/Manual/NI) | Target (Auto/Manual/NI) | Priority | Scope |
| --- | --- | --- | --- | --- |
| `cmd-root/add` | Yes/Yes/Yes | Yes/Yes/Yes | P0 | Dry-run and NI gate are present; attachment/result paths differ in auto-select depth |
| `cmd-root/update` | Yes/Yes/Yes | Yes/Yes/Yes | P0 | Wizard/NI aligned for root update and package commands |
| `cmd-root/delete` | Yes/Yes/Yes | Yes/Yes/Yes | P0 | Auto-select endpoint/id + NI guard implemented |
| `cmd-root/list` | Yes/Yes/Yes | Yes/Yes/Yes | P1 | Auto-select resource + NI guard implemented |
| `cmd-root/export` | Yes/Yes/Yes | Yes/Yes/Yes | P0 | Auto-select resource/endpoint/id + NI guard implemented |
| `get` | Yes/Yes/Yes | Yes/Yes/Yes | P0 | Main read-only branches are auto-select; the rest are intentionally manual/read-only |
| `run` | Yes/Yes/Yes | Yes/Yes/Yes | P0 | `list/get/delete/close/update/create` closed; NI guard in place in interactive branches |
| `result` | Partial/Yes/Yes | Yes/Yes/Yes | P0 | `list/get/get-case/add/add-case` closed; manual-only is `add-bulk` (required file input) |
| `sync` | Yes/Yes/Yes | Yes/Yes/Yes | P0 | NI closed at select/confirm points (`cases`, `suites`, `sections`, `shared-steps`, `full`) |
| `attachments` | Yes/Yes/Yes | Yes/Yes/Yes | P1 | `list case/plan/plan-entry/run/test`, `get`, `delete`, `add case/plan/plan-entry/result/run` closed |
| `bdds` | Yes/Yes/Yes | Partial/Yes/Yes | P2 | `add/get` support auto-select case_id + NI guard |
| `cases` | Partial/Yes/Yes | Partial/Yes/Yes | P1 | `list/get/delete/update/add/bulk` closed; next: remaining manual-only branches |
| `compare` | No/Yes/N/A | Partial/Yes/Yes | P2 | Interactive presets for source/destination |
| `configurations` | Yes/Yes/Yes | Partial/Yes/Yes | P2 | `list/add-group/add-config/update-group/update-config/delete-group/delete-config` closed |
| `datasets` | Yes/Yes/Yes | Yes/Yes/Yes | P1 | `list/add/get/update/delete` closed with project/dataset select + NI guard |
| `groups` | Yes/Yes/Yes | Yes/Yes/Yes | P1 | `list/get/add/update/delete` closed with project/group select + NI guard |
| `labels` | Partial/Yes/Yes | Partial/Yes/Yes | P2 | Auto in `get/list/update test/update-label`; `update tests` stays manual-only on input |
| `milestones` | Yes/Yes/Yes | Yes/Yes/Yes | P1 | `list/get/add/update/delete` closed with project/milestone select + NI guard |
| `plans` | Yes/Yes/Yes | Yes/Yes/Yes | P1 | `list/get/add/update/delete/close/entry add/update/delete` closed with project/plan/entry select + NI guard |
| `reports` | Yes/Yes/Yes | Partial/Yes/Yes | P2 | `list/run/run-cross-project` support auto-select + NI guard |
| `roles` | Partial/Yes/Yes | No/Yes/N/A | P2 | `get` interactive; `list` read-only/manual |
| `templates` | Partial/Yes/Yes | Partial/Yes/Yes | P2 | `list` with auto-select project + NI guard |
| `test` | No/Yes/N/A | Partial/Yes/Yes | P1 | Select run/test in read branches |
| `tests` | Yes/Yes/Yes | Partial/Yes/Yes | P1 | `list/get/update` with auto-select run/test + NI guard |
| `users` | Partial/Yes/Yes | Partial/Yes/Yes | P2 | `get/update/get-by-email` interactive; `list/add` stay manual/read-only |
| `variables` | Yes/Yes/Yes | Partial/Yes/Yes | P1 | `list/add/update/delete/get` with dataset/variable select + NI guard |

About `roles` separately:

- The package is partially interactive (`roles get`), but it may stay
mostly reference/manual-only without full UX unification of every branch.

- `NI` for `roles get` is already mandatory and implemented (a prompt
point exists).

## Dry-run matrix (mutation-free simulation)

For mutating commands a dry-run is considered correct only when:

- the command runs without a mutating API call;

- the user is shown a simulated operation (method + endpoint + body);

- there is a test confirming there is no real mutation.

| Area | Current | Target | Priority | Note |
| --- | --- | --- | --- | --- |
| `cmd-root/add` | Yes | Yes | P0 | Covered by dry-run routing and tests |
| `cmd-root/update` | Yes | Yes | P0 | Covered by dry-run routing and tests |
| `cmd-root/delete` | Yes | Yes | P0 | Interactive branch added; safety tests added |
| `sync/*` | Yes | Yes | P0 | Dry-run flags and no-mutation tests in place |
| `run/*` mutating | Yes | Yes | P1 | Largely covered; dry-run output formats need a focused review |
| `result/add*` | Yes | Yes | P1 | Dry-run flags and unit tests in place |
| `cases/*` mutating | Yes | Yes | P1 | Dry-run in add/update/delete/bulk |
| `groups/configurations/datasets/milestones/plans/variables` mutating | Yes | Yes | P1 | Dry-run flags and tests in place |
| `users add/update` | Yes | Yes | P0 | Dry-run + no-mutation tests added |
| `labels update-label` | Yes | Yes | P0 | Dry-run + no-mutation test added |
| `attachments delete` | Yes | Yes | P1 | Dry-run + no-mutation test added |
| `reports run*` | Yes | Yes | P2 | Dry-run + no-mutation tests added |

Observation:

- For some read-only commands dry-run already exists but is not required by design.

- Dry-run fix priority is set only for actual mutating branches.

Note:

- The explicit `--interactive` flag is preserved for backward compatibility.
- Recommended usage model: auto-interactive by default and `--non-interactive` for CI/CD.

---

← [Guides](index.md) · [Documentation](../index.md)
