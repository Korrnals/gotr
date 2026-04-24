# Changelog

All notable changes to the `gotr` project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

> Russian translation: see [CHANGELOG.ru.md](CHANGELOG.ru.md).

---

## [Unreleased]

## [3.3.1] - 2026-04-25

### Fixed

- **Case ordering during import** (`internal/service/migration/import.go`).
  `ImportCases` and `ImportCasesReport` create cases in parallel
  (`maxImportConcurrency = 10`); TestRail records cases in the order
  `add_case` calls arrive, and because of non-deterministic goroutine
  scheduling cases ended up shuffled relative to the source order.
  Now, after the parallel-create phase the import calls
  `move_cases_to_section` per dst section with `case_ids` sorted by
  the source index in `filtered`. Sections with `≤1` case are
  skipped. Reorder errors are logged as `WARN` but do not abort the
  migration (data is already imported, ordering is a UX concern).
  Regression covered by 4 unit tests in
  `internal/service/migration/import_order_test.go`.

## [3.3.0] - 2026-04-24

UX polish release (issue #44): categorized hierarchy for reports and
exports, shell completion, interactive mode, retention/cleanup,
warnings management and corporate TLS (`ca_bundle`).

### Added — `~/.gotr/reports/` and `~/.gotr/exports/` hierarchy

- **Report categorization** into the tree
  `~/.gotr/reports/<category>/<label|default>/<YYYY-MM>/<file>`.
  New categories: `migrations`, `coverage`, `rollbacks`, `no-snapshot`,
  `testrail/p<N>`, `_unclassified`. Classification is performed by
  `internal/report.ClassifyReport` based on filename patterns.
- **`gotr report organize [--dry-run]`** — migrates the legacy "flat"
  layout into the new hierarchy. Idempotent; conflicts (target file
  already exists) are skipped. `--dry-run` prints the plan without
  touching disk. After a successful move it calls `Reindex`.
- **Exports hierarchy** `~/.gotr/exports/{snaps,reports,api}/`.
  `snaps/` — tar.gz bundles, `reports/` — zip bundles and plain
  copies, `api/<resource>/` — legacy dumps from `gotr get
  plans/reports/...`. Migrator: `gotr export organize [--dry-run]`.
- **`gotr export snap --with-reports`** — ON by default. Recursively
  scans the reports directory and embeds files whose basename
  contains `filepath.Base(snapID)` into the archive. `--no-reports`
  opts out. Archive prefix: `reports/<rel>`. The result is reflected
  in `manifest.Files`.
- **`gotr cleanup {reports,snaps,exports,all} [--dry-run]`** — manual
  executor for retention policies. Configuration in
  `retention.{reports,snaps,exports}` (see below). `snaps` delegates
  to the existing `gotr snap gc`.
- **`gotr report show --print`** — prints report contents to stdout
  (cat-like) regardless of extension for md/json/txt. Binary PDFs
  are explicitly rejected with a clear error.
- **INDEX.md**: automatically regenerated after generate/import/organize,
  contains links to all reports in the hierarchy.

### Added — shell completion

- Dynamic `ValidArgsFunction` for `report show/view`,
  `export report/snap`, `import snap/report`. Recursive listing via
  `intreport.RecursiveListReports`; for snap commands — by
  `snap.LoadManifest`; for import — files with extensions
  `.zip/.pdf/.md/.json` or `.tar.gz/.tgz` in the matching
  directories. Handles two-dot `.tar.gz`.

### Added — interactive mode (TTY-guard)

- `report show/view`, `export report/snap`, `import snap/report` now
  accept `cobra.MaximumNArgs(1)`. If an argument is missing, stdin
  is a TTY and `--non-interactive` is not set, the user is shown a
  survey-prompt with candidate listing. In non-interactive mode
  without TTY — explicit error with hint "pass as argument or run
  interactively".

### Added — warnings suppression + TLS

- **`ui.suppress_warnings: []`** (list of keys) — silences individual
  non-critical warnings. Keys:
  - `tls_insecure` — banner shown when `insecure=true` or
    `tls.insecure=true`,
  - `deprecation` — reserved,
  - `flat_layout` — hint shown when a flat layout is detected.
- **`--show-warnings`** CLI flag (`show_warnings` viper key) —
  temporary override that displays all warnings regardless of config.
- **`tls.insecure`** — new config key. The legacy top-level
  `insecure` and `--insecure` flag are kept for backward
  compatibility; whichever source enables it wins.
- **`tls.ca_bundle: "/path/to/ca.pem"`** — corporate CA. The path is
  read, parsed through `x509.NewCertPool` + `AppendCertsFromPEM`,
  and injected into `tls.Config.RootCAs`. Preferred alternative to
  `insecure=true`. `client.WithCABundle(path)` is the public option.
- A one-time tip is appended on the first display of each warning:
  "add '<key>' to ui.suppress_warnings to silence this warning".
- The "shown the flat layout hint" flag is now persistent —
  `~/.gotr/state.json::flat_layout_warned`. Shown once per
  installation.

### Added — retention/cleanup configuration

```yaml
retention:
  reports:
    enabled: false          # OFF by default (safe)
    max_age_days: 90
    max_count: 500
    keep_categories: [coverage]
    dry_run: true
  snaps:
    enabled: false
    max_age_days: 180
    max_count: 100
    dry_run: false
  exports:
    enabled: false
    max_age_days: 30
```

- A missing `retention` block is not an error: defaults are filled
  in automatically.
- `keep_categories` — whitelist: such categories are never removed
  by the retention policy (important for coverage artifacts).

### Changed

- **`internal/paths`**: added `ExportsSnapsDirPath`,
  `ExportsReportsDirPath`, `ExportsAPIDirPath` plus Ensure\*
  variants. Writers redirected: `internal/snapbundle.DefaultExportPath`,
  `internal/reportbundle.ExportSingle`/`ExportAll`,
  `internal/output.GetExportsDir(resource)` now write into the
  `exports/` subcategories.
- **`cmd/report/list`**: `--filter` is applied by basename (glob)
  OR by substring within the relative path; listing is recursive.
- **`cmd/report/show`**: `openWithOS` uses `exec.Cmd.Run()` instead
  of `Start()`, so a non-zero exit from the OS launcher
  (`xdg-open`/`open`/`rundll32`) is propagated as a CLI error.
- **`cmd/root.PersistentPreRunE`**: warnings registry is initialised
  before any other output; the `tls_insecure` banner now flows
  through `warnings.Emitf` and respects `suppress_warnings`. Direct
  `fmt.Fprintln(os.Stderr, "WARNING: TLS...")` from
  `internal/client` removed.

### Fixed

- Removed false-positive on the first display of the flat-layout
  hint when the user has already migrated the hierarchy and removed
  `state.json` manually: `warnings.Emitf` additionally blocks
  re-display within the same process via an in-memory `shownHint`
  map.

### Migration notes (v3.2 → v3.3)

1. **Reports layout.** On the first invocation of any `report
   list/show` command a one-time warning is shown about the "flat"
   layout if files are detected at the root of `~/.gotr/reports/`.
   Recommended:
   ```bash
   gotr report organize --dry-run   # preview the plan
   gotr report organize             # perform the migration
   ```
   The command does not delete anything: on collision the original
   stays at the root and the `skipped` counter is incremented. After
   success `INDEX.md` is regenerated.
2. **Exports layout.** Likewise:
   ```bash
   gotr export organize --dry-run
   gotr export organize
   ```
   The classifier moves `*.tar.gz|*.tgz` → `exports/snaps/`,
   `*.zip|*.pdf|*.md|*.json` → `exports/reports/`, resource
   directories (plans/, reports/, runs/…) — into `exports/api/`.
3. **Insecure TLS.** Legacy `insecure: true` and `--insecure` keep
   working unchanged. Recommended migration:
   ```yaml
   tls:
     insecure: false
     ca_bundle: "/etc/ssl/corp-ca.pem"
   ```
4. **Warnings suppression.** Was: global `no_warnings: true` (in
   plan). Final design — per-key list:
   ```yaml
   ui:
     suppress_warnings: [tls_insecure, flat_layout]
   ```
   The `--show-warnings` flag shows all warnings regardless of the
   list.
5. **Retention.** OFF by default; no legacy artifacts are removed
   automatically. To migrate to the new policy explicitly set
   `retention.*.enabled: true` and run `gotr cleanup all --dry-run`
   before the production run.

### Internal

- New packages: `internal/warnings` (registry), `internal/state`
  (persistent JSON store), `internal/exportsorg` (migrator),
  `internal/retention` (policy + executor).
- Lifecycle E2E tests: `internal/report/e2e_lifecycle_test.go`
  (flat→hierarchy, 6 categories, idempotency),
  `internal/snapbundle/e2e_reports_test.go` (organize → export
  --with-reports → import round-trip).

---

### Added — export/import bundles (#42, #43)

- **`gotr export snap <id>` → portable tar.gz** with `manifest.json`
  (`schema_version=1`, gotr version, created_at, per-file SHA-256 and
  `~/.gotr/...` relative paths), `SHA256SUMS`, `README.txt` and the full
  `~/.gotr/snaps/<id>/` subtree. Default destination
  `~/.gotr/exports/snap_<id>_<label>_<YYYYMMDD>.tar.gz`. Flags:
  `--out`, `--redact` (strips `assignee`, `assignee_email`, `assignee_name`,
  `email` from JSON payloads and audits stripped paths via
  `manifest.redacted_fields`).
- **`gotr import snap <path.tar.gz>`** verifies SHA256SUMS and
  `schema_version` before touching the store, then extracts into
  `~/.gotr/snaps/<id>/`. Flags: `--overwrite` (backs up existing snap to
  `~/.gotr/snaps/.trash/<id>_<ts>/`), `--rename-id <new>` (imports under a
  new id and rewrites `meta.json.id` accordingly), `--dry-run`.
- **`gotr export report <filename|all>` + `gotr import report <path>`.**
  Single file → plain copy into `~/.gotr/exports/<basename>`; `all` →
  zip bundle at `~/.gotr/exports/reports_<YYYYMMDD>.zip` with the same
  manifest/SHA256SUMS/README layout. Import handles both plain files and
  zip bundles. Flags: `--out`, `--filter <glob>`, `--overwrite`, `--dry-run`.
- **`gotr report show <filename|latest>`** opens PDFs via the OS default
  viewer (`xdg-open`/`open`/`rundll32`) and cats MD/JSON/TXT to stdout.
- **`gotr report list --filter <glob>`** applies a glob to report
  filenames and now also accepts `.pdf`/`.json` reports (previously only
  `migration-*.md` was visible).
- New internal packages: `internal/bundle` (tar.gz + zip + SHA-256 + manifest
  primitives with zip-slip protection and deterministic archive output),
  `internal/snapbundle`, `internal/reportbundle`.

### Fixed — migration engine (3.2.0 follow-up)

- **Coverage gate no longer false-negatives on name-resolved sections.**
  `internal/service/migration/import.go`:
  `resolveSectionMapByName` now registers every `src→dst` pair it resolves by
  section name into `m.mapping` (`AddPair(src, dst, "existing")`). Previously
  the map lived only as a local variable consumed by `ImportCasesReport`,
  while `VerifyCasesCoverage → resolveDstSectionIDForFilter` consulted
  `m.mapping` exclusively and therefore reported all cases as missing after a
  successful migration. Regression test:
  `TestResolveSectionMapByName_PopulatesGlobalMapping`.
- **Sync snapshots now expose real `data_size_bytes`.**
  `internal/snap/hook.go` `Hook.FinalizeSyncData` rewrites `meta.json` with
  the actual payload size and calls the new
  `Manifest.UpdateDataSize` so list/retention/UI no longer see
  `data_size_bytes=0` for sync snaps. Rollback was already reading
  `data.json` directly, so this is purely observational for existing flows
  but unblocks retention policies tied to size metadata and fixes inspection
  output. Regression: extended `TestHook_FinalizeSyncData_Happy` asserts both
  meta and manifest entry are updated.
- **Flag parity: `--verify-coverage` exposed on `sync full` and `sync cases`.**
  `cmd/sync/sync.go` registers the flag explicitly on both commands
  (previously only `sync shared-steps` / `suites` / `sections` saw it via
  `addSyncFlags`).

---

## [3.2.0] - 2026-04-23

Migration bugfix release: eliminates a hidden case-loss
discrepancy caused by erroneous "silent" filtering and section
re-parenting behaviour in the migration engine.

### Fixed — migration engine

- **Multiset matching by `(dst_section_id, compare_field)`.** Before
  3.2.0 the filter treated equal `title` values within a single
  `section_id` as "already in target" and skipped source cases with
  the same title even when the target did not actually contain such
  a case (or had it in a different scope). The new implementation
  (`internal/service/migration/match.go`, `filter.go`) uses a
  multiset with FIFO consumption: each source case consumes exactly
  one target case by matching key.
- **Strict dst-scope resolution.** `resolveDstSectionIDForFilter` no
  longer "collapses" unmapped sections to target root — for an
  unresolvable `section_id` it returns the negative sentinel
  `-srcSectionID`, guaranteeing that such source cases are never
  mismatched with cases of foreign sections.
- **No more silent root fallback during section import.**
  `ImportSections`: a section with an unmapped `parent_id` is now
  rejected with an error and counted in `failedImports`, instead of
  being silently re-parented under root (this was the root cause of
  silent case loss when the source had sections with unmapped parents).
- **`FailedCount()` reflects real import failures.** `ImportCases`,
  `ImportCasesReport`, `ImportSections`, `ImportSuites`,
  `ImportSharedSteps` now increment `failedImports` on every API
  error or import refusal. The `max(len(errs), FailedCount())`
  convention in sync command reports correctly turns silently-missing
  cases from invisible "skipped" into explicit `errors`/`failed`.

### Added — compare pipeline

- **`--suite1` / `--suite2` (persistent).** On `gotr compare *` a
  separate suite can be pinned for each project (`0 = all suites`,
  same as before). On multi-suite targets the comparison no longer
  silently traverses different scopes.
- **`--match-field` (persistent, shared).** New unified compare
  field flag — takes precedence over `--field`, applied uniformly
  across all `compare` subcommands.
- **`filterSuitesByID`**: returns an empty scope for an unknown
  suite-id (instead of falling back to all project suites) — no
  silent fallback at the comparison level.

### Added — sync safety gate

- **`--verify-coverage` (opt-in, default `false`)** on `sync cases`
  and `sync full`. After import re-fetches the target and asserts
  that every source case has a multiset-key match; on a gap exits
  with `coverage gap: ...` (non-zero exit code) and log lines
  `  - [id] "title" (src_section=X, dst_section=Y)` (up to 50). This
  gate is a direct guardian against silent case-loss regressions
  recurring in the future.
- **`Migration.VerifyCasesCoverage`**
  (`internal/service/migration/coverage.go`): a standalone API
  method consumed inside `runCoverageGate` and callable from user
  code.

### Added — interactive UX

- **`SelectMatchField`** (`internal/interactive/match_field.go`):
  new TTY-guarded prompt to pick the compare field for `compare`/
  `sync` with normalisation (`MatchFieldCases`, `MatchFieldSections`,
  `MatchFieldSuites`, `MatchFieldSharedSteps`).

### Changed

- Refactor `cmd/compare/cases.go`: extracted `resolveCompareField`
  and `applySuiteScope` helpers — reduces cyclomatic complexity of
  `newCasesCmd` and `compareCasesInternal`.
- `internal/service/migration/filter.go`: tag-less `switch` in the
  `dstParentID` resolver replaced with `if/else` (staticcheck
  QF1002).

### Tests

- New regression tests
  (`internal/service/migration/import_test.go`,
  `cmd/sync/sync_helpers_coverage_test.go`,
  `cmd/sync/sync_flags_test.go`) lock the contract:
  - source case with unresolvable `SectionID` → `FailedCount++`,
    `AddCase` not called.
  - source section with unmapped `ParentID` → rejected, failure
    counted.
  - `AddCase` error appears both in `[]errs` and `FailedCount`.
  - `--verify-coverage` defaults to `false`, gate is no-op without
    the flag.
  - coverage breach returns an error containing the substring
    `"coverage gap"` (stable contract for CI/grep).

### Docs

- Updated `README.md` / `README_ru.md`: new flags and the coverage
  gate.
- Updated `docs/{en,ru}/guides/commands/{compare,sync}.md`.

---

## [3.1.0] - 2026-04-19

### Added

#### #28: Snap backup and rollback engine

- **`gotr snap backup`** — create named snapshots of TestRail project state before migration.
- **`gotr snap restore`** — roll back project to a named snapshot.
- **`gotr snap list`** — list available snapshots with metadata.
- **`gotr snap gc`** — garbage-collect stale snapshots respecting retention policy.
- **`gotr snap pin` / `gotr snap unpin`** — pin snapshots to protect from GC.
- **`internal/snap/`** — full snapshot engine: store, hooks (67), rollback engine, conflict resolution.
- **`pkg/snap_smoke/`** — smoke test package for integration-level validation.
- Sync commands now auto-create a snapshot before destructive operations.
- Server naming context propagated through all snap operations.

#### #29: Interactive navigation kit and Compare UX modernization

- **`internal/interactive`** — shared navigation kit: `aligned.go`, `browse.go`, `navigation.go`, `action_menu.go`, `group.go`, `pager.go` (Phase 5).
- **`cmd/compare`** — post-action menu, pager, drill-down UX modernization (Phase 6).
- Adds `golang.org/x/term` for terminal detection.

#### #30: Navigation hub, AlignedLabels UX, cross-command dispatcher

- **`gotr work`** — new unified work hub command as the interactive entry point.
- **`internal/interactive/session.go`** — server session context with URL display.
- **`internal/interactive/nav_dispatch.go`**, **`server.go`** — extracted `SelectServer` and `NavigateMenu` to eliminate duplication.
- **`internal/interactive/mutation_action.go`** — standardized post-mutation action menu across all mutation commands.
- **`internal/interactive/cross_nav.go`** — cross-command navigation dispatcher (E1-E4, B4 navigation patterns).
- **AlignedLabels Tier 1-3** — all interactive pickers now use aligned column layout for consistent UX.

#### #31: Audit hardening, migration docs, sync rollback reliability

- **F1-F12 audit** — full backlog execution across 142 files (3 879 additions); covers nil-guards, error wrapping, context propagation, test coverage.
- **Bare error wrapping** — 16 `return err` paths wrapped with `fmt.Errorf` context.
- **`cmd/snap` coverage** — boosted to 81%.
- **`internal/report/`** — new package: `service.go`, `types.go`, `doc.go` + tests.
- **Migration walkthrough docs** — RU + EN step-by-step guides added to `docs/`.
- **Shared-step ID remapping** — semantics documented in both languages.
- **`sync` rollback fix** — snapshot rollback is now reliable for all migration flows.

#### #32: Snap retention policy, GC, pin/unpin, migration reports

- **`internal/snap/labeling.go`** — label generation (`batch_`, `interactive_`, `auto_` prefixes + timestamp), validation, `IsProtectedLabel`.
- **`internal/snap/interactive_prompt.go`** — interactive label prompt (use default / custom / pin / skip).
- **Retention policy config** — `gotr.yaml` block with keep-count, max-age, protected-prefixes + validation.
- **`cmd/snap/gc.go`** — GC respects retention policy; dry-run mode available.
- **`cmd/snap/pin.go`** — pin/unpin commands.
- **`cmd/report/`** — `gotr report list`, `gotr report view` — migration report commands.
- **`cmd/sync/sync_report_helpers.go`** — report generation helpers integrated into sync flows.
- **Docs** — live migration test plan + operator card added to `docs/ru/guides/instructions/`.

---

## [3.0.1] - 2026-04-12

### Fixed

- Removed dead `--soft` flag from `gotr delete` (was declared but never used).
- Removed misleading milestone/plan/entry references from `gotr add` and `gotr update` help text.
- Fixed `--save-filtered` flag: wired into `sync full` command to actually save filtered shared steps list after migration.
- Fixed self-test table alignment for dynamic content widths.
- Fixed spinner first-frame delay in `ui.RunWithStatus` (immediate render before ticker).

### Added

- Progress spinners for all API-calling commands (~48 commands across cases, attachments, configurations, milestones, plans, groups, labels, run, result, tests, reports, export).
- Spinner wrapper for `gotr self-test` execution.

### Changed

- Documentation: added missing compare flags (`--include-refs`, `--include-custom-statuses`, `--include-custom-steps`, `--include-updated-by`, `--include-details`) to EN/RU guides.
- Documentation: corrected CRUD instruction redirects for milestones/plans.
- Documentation: fixed artifact filenames in migration-shared-steps guide.

---

## [3.0.0] - 2026-04-09

### Added

#### Stage 13.5: Quality Hardening & Full Audit

- **`api_paths.go`** — +14 endpoints added to the endpoint registry, complete coverage of TestRail API v2.
- **`attachments list --for-project`** — new subcommand wrapping `GetAttachmentsForProject()`.
- **`bdds add`** — stdin reading support: `cat scenario.feature | gotr bdds add 12345`.
- **`sync shared-steps --save-filtered`** — automatic/interactive saving of filtered shared steps list via `ExportSharedSteps()`.
- **Generic CRUD executor** (`internal/crud`) — eliminates boilerplate for simple CRUD commands.
- **Compare resource registry** (`cmd/compare`) — declarative resource registration replacing manual wiring.

### Changed

- `compare all`: stage-by-stage progress tracker in terminal (`done/active/pending`) for all resources.
- `compare all`: shared suites prefetch for `cases/suites/sections` to avoid repeated `get_suites` calls.
- `compare all`: resource failures are now marked as `PARTIAL` (instead of misleading `INTERRUPTED`).
- `compare all`: unsupported TestRail endpoints (`404 Unknown method`) are shown as `UNSUPPORTED` with a dedicated `Unsupported endpoints` summary block.
- `compare all` JSON/YAML meta now distinguishes real errors from unsupported endpoints:
  - `error_summary_count` / `error_resources` for real failures
  - `unsupported_summary_count` / `unsupported_resources` for server-unsupported methods
- Legacy `internal/progress` package removed; progress/status flow is unified via `internal/ui` runtime.
- All Russian text in Go source files translated to English (i18n pass: 1738+ lines across 2 passes).
- `panic(err)` in `main.go` and `cmd/commands.go` replaced with `fmt.Fprintf(os.Stderr)` + `os.Exit(1)`.
- `ClientInterface` unified across all service packages (B-2..B-4 audit remediation).

### Fixed

- `internal/client` paginator: fixed potential infinite loop for flat-array API responses with page size at or above 250.
- `compare sections`: stabilized loading path via client pagination behavior and added regression coverage in paginator tests.
- All `io.ReadAll(resp.Body)` calls wrapped with `io.LimitReader` (10 MiB cap) to prevent unbounded memory allocation.
- File descriptor leak in `migration/types.go` — `logFile` now properly closed in `Migration.Close()`.
- `os.Remove` error paths in `embedded/jq_embed.go` now checked and logged.
- All `json.Marshal` errors across the codebase handled (45+ fixes in 17 files).
- Safe type assertions with comma-ok pattern throughout; `os.Getwd` errors properly handled.
- Context propagation ensured across all API calls (F-2..F-7 audit findings).

### Security

- Bounded parallelism enforced in all concurrent operations.
- All HTTP response body reads are size-limited.
- `ReadResponse` documentation clarifies `resp.Body` ownership contract.

### Quality

- **golangci-lint**: 290 findings → **0** (errcheck, staticcheck, gocritic, gocyclo, misspell, unused, nolintlint, ineffassign).
- **Test suite**: 43/43 packages pass with race detector, 0 data races.
- **0 TODO/FIXME/HACK** markers in production code.
- **Audit verdict**: UNCONDITIONAL PASS (7 audit rounds completed).

## [3.0.0] - 2026-03-12

### Added

#### Stage 6.8: Concurrency Unification & Compare Subcommands

- **`internal/concurrency/`** — new unified concurrency package (renamed from `internal/parallel/`)
  - Three strategy tiers:
    - `FetchParallel[T]` — light: one API call per project, parallel loading of P1+P2
    - `FetchParallelBySuite[T]` — medium: per-suite requests (for `sections`)
    - `FetchParallelPaginated` — heavy: `ParallelController` with pagination (for `cases`)
  - Generic API via Go generics `[T any]`

- **`pkg/reporter/`** — universal reporter extracted into a public package (from `internal/ui/reporter/`)
  - Builder pattern: `Section` / `Stat` / `StatIf` / `StatFmt` / `Print`
  - go-pretty/v6 for aligned, boxed output

- **Generic `newSimpleCompareCmd`** — single generic factory replacing 9 identical files (`cmd/compare/simple.go`)
  - Eliminated ~1200 lines of copy-paste
  - All simple subcommands use `FetchParallel[T]` for parallel project loading

- **`compare sections`** — parallel section loading per suite via `FetchParallelBySuite[T]`

- **`compare all`** — uniform output via `pkg/reporter`, partial results when an API is unavailable

### Changed

- `internal/parallel/` → `internal/concurrency/` (package rename and all imports)
- `internal/ui/reporter/` → `pkg/reporter/` (extracted into a public package)
- All 13 compare subcommands use `pkg/reporter` for stat output
- `OnSuiteComplete` → `OnItemComplete` in the `ProgressReporter` interface
- Default values: `parallel-suites=10`, `parallel-pages=6` (stable for TestRail Server)

### Fixed

- `compare all` no longer uses `fmt.Println` with emoji and box-drawing characters
- Fixed misaligned statistics in terminals without emoji support

### Performance

- Simple compare subcommands (runs, plans, milestones, etc.): P1 and P2 loaded **in parallel**
- `compare sections`: per-suite parallel loading instead of sequential

---

#### Stage 6.9: Generic Paginator & Pagination Audit

### Added

- **`internal/client/paginator.go`** — universal generic paginator `fetchAllPages[T]`
  - Handles both TestRail API formats without branches in business logic:
    - **Paginated wrapper** (TestRail 6.7+): `{"offset":0,"limit":250,"size":N,"<key>":[...]}`
    - **Flat array** (legacy TestRail Server): `[item1, item2, ...]`
  - Auto-detects format by the first response byte
  - Standard page size: 250 items (TestRail default)
  - Exit condition: `len(page) < limit` (last page)

- **Migration of 9 critical list methods** to `fetchAllPages[T]`:
  - `GetRuns(projectID)` — runs no longer truncated at 250
  - `GetPlans(projectID)` — plans no longer truncated at 250
  - `GetSections(projectID, suiteID)` — sections (critical for `compare sections`)
  - `GetSharedSteps(projectID)` — shared steps
  - `GetMilestones(projectID)` — milestones
  - `GetResults(runID)` — run results
  - `GetResultsForRun(runID)` — extended variant
  - `GetTests(runID)` — run tests
  - `GetSuites(projectID)` — project suites

### Changed

- All 9 migrated methods: method body simplified from ~30 lines of manual loop to a single `fetchAllPages` call
- Removed ~145 lines of duplicated pagination boilerplate from `internal/client/`

### Tests

- `internal/client/paginator_test.go` — 11 new unit tests:
  - Both response formats (paginated wrapper and flat array)
  - Multi-page accumulation
  - Edge cases: empty response, last partial page
  - Server error test (HTTP 500)
  - Table-driven tests for `decodeListResponse`

### Verified

- `compare all --pid1 30 --pid2 34`: 20 509 cases (87 pages) + 116 009 cases (475 pages) — pagination confirmed against real data
- `compare runs`, `compare plans`, `compare milestones`, `compare sections`, `compare sharedsteps`: all working correctly
- `go test ./...` — all tests green

---

#### Stage 7.0: Context Propagation

### Added

- **`context.Context`** propagated to all ~100 methods of `ClientInterface`
  - `signal.NotifyContext` → graceful shutdown on Ctrl+C
  - Context flows CLI → Service → Client → HTTP

### Changed

- All API methods accept `ctx context.Context` as the first argument
- `cmd.ExecuteContext()` instead of `cmd.Execute()`
- `MockClient` updated to match the new signatures

---

#### Stage 8.0: UI/Output Refactoring

### Added

- **`internal/ui/`** — universal helpers:
  - `ui.Table(headers, rows)` — go-pretty wrapper replacing tabwriter
  - `ui.JSON(v)` — formatted JSON output
  - `ui.Success()`, `ui.Warn()`, `ui.Error()`, `ui.Info()` — coloured messages
  - `ui.Print()`, `ui.Printf()`, `ui.Println()` — stdout wrappers
- **`--format` PersistentFlag** — global output format flag at the root level
- Mass migration: `tabwriter` → `ui.Table`, `json.MarshalIndent` → `ui.JSON`, `fmt.Print*` → `ui.*` (49 files)

### Changed

- `internal/flags/`: `*Var` → `GetFlag`, `ValidateRequiredID`
- `os.Exit` → `panic` in `GetClient*` (testability)
- All error messages translated to English

---

## [2.7.0] - 2026-02-20

### Added

#### Stage 6: Performance Optimization & UX Enhancement (In Progress)

- **Universal Progress Monitoring**: Channel-based progress system
  - `internal/progress.Monitor` — decoupled from UI
  - Real-time updates via buffered channels
  - Thread-safe, non-blocking implementation
  - Works with any long-running operation
  - See [docs/guides/progress.md](docs/guides/progress.md) for details

- **Multi-Progress-Bars (mpb)**: Visual feedback for long-running operations
  - `github.com/vbauerster/mpb/v8` integration (multi-progress-bar library)
  - Multiple simultaneous progress bars on separate lines
  - Real-time updates for parallel project loading
  - ETA, speed, and percentage decorators

- **Parallel API Requests**: 60-80% performance improvement
  - Worker pool pattern for concurrent requests
  - Rate limiting (180 req/min — TestRail maximum)
  - Parallel fetching for cases, suites, shared steps
  - Integrated with progress monitoring
  - **Page-level progress**: GetCasesWithProgress updates after each 250-case page

- **Compare Cases Command**: Full comparison with parallel loading
  - Two-phase progress: spinner → progress bars
  - Parallel loading of both projects simultaneously
  - Project-level statistics (suites count, cases count, duration)
  - Analysis phase with timing
  - Debug mode support via `--debug` flag

- **Response Caching**: Disk-based cache with TTL
  - Cache location: `~/.gotr/cache/`
  - TTL: Projects 1h, Cases 15min, Suites 30min
  - `--no-cache` flag to bypass

- **Retry Logic**: Exponential backoff for resilience
  - Automatic retry on transient failures
  - Circuit breaker pattern
  - `--timeout` flag (default: 5min)

- **Batch Operations**: Optimized for large projects
  - Batch fetching (250 items per request)
  - Streaming output for large datasets
  - Memory optimization (<500MB peak)

### Changed

- **Progress Bar Library**: Migrated from `progressbar/v3` to `mpb/v8`
  - Better support for multiple simultaneous progress bars
  - Improved UX with parallel operations
  - New API: methods called directly on bar objects (`bar.Add()`, `bar.Finish()`)

#### Stage 5: CLI Test Coverage (986 tests, 10 packages at 100%) - COMPLETE

- **Comprehensive test suite** for all CLI commands:
  - `cmd/run` — 38 tests (95.2% coverage)
  - `cmd/result` — 46+ tests (90.6% coverage)
  - `cmd/get` — 71 tests (89.7% coverage)
  - `cmd/attachments` — 18 tests (100% coverage)
  - `cmd/labels` — 60+ tests (100% coverage)
  - `cmd/groups` — 26 tests (100% coverage)
  - `cmd/cases` — 87 tests (99.2% coverage)
  - `cmd/test` + `cmd/tests` — 35 tests (90.5%+ coverage)
  - `cmd/templates`, `cmd/reports`, `cmd/sync`, `cmd/milestones`, `cmd/plans`
  - `cmd/users`, `cmd/variables`, `cmd/configurations`, `cmd/datasets`
  - `cmd/bdds`, `cmd/roles` — all at 87-100% coverage
- **Test infrastructure:**
  - Shared `testhelper` package for mock client injection
  - `serviceWrapper` pattern for interface-based testing
  - Constructor pattern for all commands (`newXxxCmd`)
- **Total:** 110 test files, 986 test functions, 18 packages ≥ 90% coverage

#### Stage 5.2: Project Comparison Command + Unified --save Flag

- **New `cmd/compare/` package** with subcommand structure:
  - `gotr compare cases` - compare test cases with field-based diff
  - `gotr compare suites` - compare test suites
  - `gotr compare sections` - compare sections
  - `gotr compare sharedsteps` - compare shared steps
  - `gotr compare runs` - compare test runs
  - `gotr compare plans` - compare test plans
  - `gotr compare milestones` - compare milestones
  - `gotr compare datasets` - compare datasets
  - `gotr compare groups` - compare groups
  - `gotr compare labels` - compare labels
  - `gotr compare templates` - compare templates
  - `gotr compare configurations` - compare configurations
  - `gotr compare all` - compare all resources at once with formatted table output
- **Enhanced `--save` and new `--save-to` flags:**
  - `--save` - saves table output as text file to `~/.gotr/exports/{resource}/`
  - `--save-to <path>` - saves to specified path with format from `--format` or auto-detected from extension
  - Auto-detection: `.json` → JSON, `.yaml`/`.yml` → YAML, `.csv` → CSV, `.txt` → table
  - Supports JSON, YAML, CSV, and table (text) formats
  - Affects all `compare` subcommands
- **BREAKING CHANGE: `--save` flag replaces `--output` across ALL commands:**
  - `--save` is now a boolean flag (no value required)
  - Saves to `~/.gotr/exports/{resource}/{resource}_YYYY-MM-DD_HH-MM-SS.{format}`
  - Supports JSON, YAML, and CSV formats via `--format` flag (where applicable)
  - Affected commands: `get`, `export`, `users list`, `labels list`, `reports list-cross-project`,
    `test get/list`, `tests list`, `groups add/update`, and all `compare` subcommands
- **Field-based comparison** for cases: `--field title`, `--field priority_id`, etc.
- **Formatted table output** for `compare all`:
  - Unicode box-drawing characters for clean presentation
  - Status indicators: ✓ (perfect match), ⚠ (has differences), ✗ (error loading)
  - Compact summary showing counts per resource type
  - Error section for failed resource comparisons
- **Package structure:**
  - `types.go` - shared types (CompareResult, ItemInfo, CommonItemInfo)
  - `register.go` - command registration with root command
  - Individual files per resource (cases.go, suites.go, etc.)

#### Save Package (cmd/common/flags/save)

- **New package** for standardized output saving across all commands:
  - `SaveWithOptions()` - unified save function supporting JSON, YAML, CSV formats
  - `GenerateFilename()` - generates timestamped filenames: `{resource}_YYYY-MM-DD_HH-MM-SS.{ext}`
  - `GetExportsDir()` - returns `~/.gotr/exports/` directory path
  - Automatic directory creation with 0755 permissions
  - CSV export with dynamic header detection from struct tags
  - Over 40 comprehensive tests (100% coverage)

#### Build System Improvements

- **Automatic version sync in Makefile:**
  - `make build` now extracts the version from `cmd/root.go` (single source of truth)
  - For release versions (no `-dev` suffix) a git tag is auto-created/verified
  - Version priority: 1) `make build VERSION=x`, 2) version from code, 3) git tag
  - Tag normalisation: supports both `VERSION=v2.7.0` and `VERSION=2.7.0`

#### Stage 4: Complete API Coverage (106/106 endpoints)

- **Attachments API** — 5 endpoints:
  - `AddAttachmentToCase`, `AddAttachmentToPlan`, `AddAttachmentToPlanEntry`
  - `AddAttachmentToResult`, `AddAttachmentToRun`
  - multipart/form-data support for file uploads
- **Configurations API** — 7 endpoints:
  - `GetConfigs`, `AddConfigGroup`, `AddConfig`
  - `UpdateConfigGroup`, `UpdateConfig`, `DeleteConfigGroup`, `DeleteConfig`
- **Users API** — 4 endpoints:
  - `GetUsers`, `GetUser`, `GetUserByEmail`
- **Reference Data APIs** — 3 endpoints:
  - `GetPriorities`, `GetStatuses`, `GetTemplates`
- **Reports API** — 3 endpoints:
  - `GetReports`, `RunReport`, `RunCrossProjectReport`
- **Extended APIs** — 21 endpoints:
  - Groups: `GetGroups`, `GetGroup`, `AddGroup`, `UpdateGroup`, `DeleteGroup`
  - Roles: `GetRoles`, `GetRole`
  - ResultFields: `GetResultFields`
  - Datasets: `GetDatasets`, `GetDataset`, `AddDataset`, `UpdateDataset`, `DeleteDataset`
  - Variables: `GetVariables`, `AddVariable`, `UpdateVariable`, `DeleteVariable`
  - BDDs: `GetBDD`, `AddBDD`
  - Labels: `UpdateTestLabels`, `UpdateTestsLabels`

**Total implemented:** 44 new endpoints
**Overall coverage:** 106/106 TestRail API endpoints (100%)

### Added

#### Dry-run mode

- **Flag** `--dry-run` — single flag for all state-mutating commands:
  - `add` — project, suite, section, case, run, result, shared-step
  - `update` — project, suite, section, case, run, shared-step
  - `delete` — project, suite, section, case, run, shared-step
  - `run create/update/close/delete`
  - `result add/add-case/add-bulk`
- **Package** `cmd/common/dryrun/` — centralised formatting for dry-run output

#### Interactive wizard mode

- **Flag** `--interactive/-i` — interactive mode for commands:
  - `add` — project, suite, case, run
  - `update` — project, suite, case
- **Package** `cmd/common/wizard/` — library of interactive prompts on survey/v2
- Pattern: input → preview → confirm/cancel

### Changed

- **Flag** `-i` is now used for `--interactive` (instead of `--insecure`)
- **Flag** `--insecure` — long form only (no shorthand)

---

## [2.5.0] - 2026-02-05

### Added

#### Interactive mode

- **Command** `gotr run list` — interactive project picker when no arguments are given
- **Command** `gotr result list` — interactive picker: project → test run
- **Package** `internal/interactive/` — unified interactive selection mechanism

#### Client Interface + Mock (architectural improvement)

- **Package** `internal/client/interfaces.go` — full composite interface:
  - `ProjectsAPI` — 5 methods
  - `CasesAPI` — 14 methods
  - `SuitesAPI` — 5 methods
  - `SectionsAPI` — 5 methods
  - `SharedStepsAPI` — 6 methods
  - `RunsAPI` — 6 methods
  - `ResultsAPI` — 7 methods
- **Package** `internal/client/mock.go` — full `MockClient` (43 methods)
- Compile-time check: `var _ ClientInterface = (*HTTPClient)(nil)`

#### Shared utilities (refactor)

- **Package** `cmd/common/client.go` — `ClientAccessor` for unified access to the HTTP client
- **Package** `cmd/common/flags.go` — shared flag-parsing helpers
- Refactor of `cmd/result/`, `cmd/run/`, `cmd/sync/` — uses `common.ClientAccessor`
- Removed `getClientSafe` duplication across 3 packages

### Fixed

#### Migration interface unification

- **Removed duplicate package** `internal/migration` (kept `internal/service/migration`)
- **Unified interface** — `internal/service/migration` now uses `client.ClientInterface`
- **Updated `MockClient`** — default return values prevent nil pointer dereference
- **Refactored sync tests** — all 10 tests rewritten using `client.MockClient`
- Removed test skips (`t.Skip`) — every test now passes

### Changed

- **README.md** — restructured description, acknowledgements moved to the end
- Version bumped to `2.5.0`

---

## [2.4.0] - 2026-02-04

### Added

#### Results API (full implementation)

- **New client** `internal/client/results.go` with methods:
  - `AddResult` — add a result for a test
  - `AddResultForCase` — add a result for a case in a run
  - `AddResults` — bulk add results
  - `AddResultsForCases` — bulk add for cases
  - `GetResults` — fetch results for a test
  - `GetResultsForRun` — fetch all results for a run
  - `GetResultsForCase` — fetch results for a case in a run

#### Runs API (full implementation)

- **New client** `internal/client/runs.go` with methods:
  - `GetRun` — fetch run info
  - `GetRuns` — list project runs
  - `AddRun` — create a new run
  - `UpdateRun` — update an existing run
  - `CloseRun` — close a run
  - `DeleteRun` — delete a run

#### CLI commands for Results

- **New package** `cmd/result/` with commands:
  - `gotr result get <test-id>` — get results
  - `gotr result get-case <run-id> <case-id>` — get results for a case
  - `gotr result add <test-id>` — add a result
  - `gotr result add-case <run-id>` — add a result for a case
  - `gotr result add-bulk <run-id>` — bulk add from a JSON file

#### CLI commands for Runs

- **New package** `cmd/run/` with commands:
  - `gotr run get <run-id>` — fetch run info
  - `gotr run list <project-id>` — list project runs
  - `gotr run create <project-id>` — create a run
  - `gotr run update <run-id>` — update a run
  - `gotr run close <run-id>` — close a run
  - `gotr run delete <run-id>` — delete a run

#### Service Layer (architectural improvement)

- **New package** `internal/service/`:
  - `RunService` — business logic for runs with validation
  - `ResultService` — business logic for results with validation
  - `internal/service/migration/` — migrated from `internal/migration/`
- **Validation** in services:
  - ID > 0 checks
  - Required fields (name, suite_id, status_id)
  - Bulk request validation (non-empty arrays)
- **Utilities** in `internal/utils/helpers.go`:
  - `ParseID` — parse an ID
  - `OutputResult` — output result (JSON + save to file)
  - `PrintSuccess` — output messages
  - `SaveToFile` — save data to a JSON file

#### Architecture documentation

- **System docs** in `.github/copilot/instructions/`:
  - Full description of the 4 architectural layers
  - Responsibility split tables
  - Component inventory (22 commands, 3 services, 40+ API methods)
  - Refactoring examples
- **User docs** in `docs/architecture/overview.md` (243 lines):
  - Simplified architecture description
  - Data-flow examples
  - Full command list

#### Tests

- **Service Layer tests**:
  - `internal/service/run_test.go` — 6 tests for RunService validation
  - `internal/service/result_test.go` — 9 tests for ResultService validation

---

## [2.3.0] - 2026-02-03

### Added

#### Models for Results and Runs API

- **New data models** in `internal/models/data/`:
  - `results.go` — `Result`, `AddResultRequest`, `AddResultsRequest`, `AddResultsForCasesRequest`
  - `runs.go` — `Run`, `AddRunRequest`, `UpdateRunRequest`, `CloseRunRequest`
  - `tests.go` — `Test`, `UpdateTestRequest`
  - `statuses.go` — `Status` plus status constants
- Preparation for the Results and Runs API implementation

#### Audit-driven fixes

- **Updated request structures** in `cases.go`:
  - `AddCaseRequest` — added `custom_steps` and `custom_expected` (text format)
  - `UpdateCaseRequest` — added `type_id`, `suite_id`, `section_id`, `template_id` for case relocation
- **Fixed `Section` model** — `omitempty` added to optional fields
- **Removed duplicate** `AddCaseRequest` method from `internal/client/cases.go`
- **Fixed `GetSections`** — `suite_id` is now passed as a query parameter

#### System changes

- Development system files moved to internal instructions in `.github/copilot/instructions/`
- Adopted explicit Semantic Versioning

---

## [2.2.3] - 2026-02-03

### Added

#### Interactive mode

- **Interactive selection** for all `get` and `sync` commands:
  - `gotr get cases` — interactive project + suite picker
  - `gotr get suites` — interactive project picker
  - `gotr get sharedsteps` — interactive project picker
  - `gotr sync cases` — interactive source/destination project + suite pickers
  - `gotr sync shared-steps` — interactive project pickers
  - `gotr sync sections` — interactive project + suite pickers
  - `gotr sync full` — interactive selection for full migration
- **Auto-selection**: if a project has only one suite — it is used automatically
- **`--all-suites` flag** for `gotr get cases` — fetch cases from all project suites

#### Code restructure

- New `cmd/` package layout:
  - `cmd/get/` — dedicated package for GET commands
  - `cmd/sync/` — dedicated package for SYNC commands
  - `cmd/commands.go` — centralised command registration
  - `cmd/interactive.go` — shared interactive selection helpers
- Dependency injection to avoid circular dependencies between packages

#### Documentation

- Created the `docs/` directory with detailed documentation:
  - `guides/installation.md` — installation
  - `guides/configuration.md` — configuration
  - `guides/commands/get.md` — data-retrieval commands
  - `guides/commands/sync.md` — synchronisation commands
  - `guides/interactive-mode.md` — interactive mode
  - `guides/commands/other.md` — other commands

### Changed

- Improved `suite-id` handling in `gotr get cases`:
  - Removed the hard requirement of `--suite-id`
  - Interactive picker when the flag is absent
  - Clear API error message when suite_id is missing on multi-suite projects
- Updated `Long` descriptions of all sync commands with interactive-mode notes
- Flag registration moved out of `init()` functions into `cmd/sync/sync.go`

### Fixed

- Removed flag duplication in `sync` commands
- Removed unused variables in tests

---

## [2.1.0] - 2026-01-24

### Added

- `gotr sync suites` — new suites synchronisation command: Fetch → Filter → Import.
- `gotr sync sections` — new sections synchronisation command.
- Shared helper `addSyncFlags()` to unify flags across `sync/*` commands.
- Unit tests for `sync suites` and `sync sections`.

### Changed

- `sync/*` commands moved onto a unified migration flow (`internal/migration`) and now share the centralised Fetch → Filter → Import logic.
- Improved `Long` command descriptions and added Russian "Steps" comments inside command code for the convenience of Russian-speaking users.

### Testing

- Tests use a separate logs folder: `.testrail/logs/test_runs`.
- Introduced a test seam `sync_helpers.go` (variable `newMigration`) to inject mock migrations in tests.

---

## [2.0.0] - 2026-01-15

### Breaking Changes

- Full rework of the `get` command: switched to subcommands instead of a universal approach.
  - Now `gotr get <resource>` with subcommands: `cases`, `case`, `projects`, `project`, `sharedsteps`, `sharedstep`, `sharedstep-history`, `suites`, `suite`.
  - Removed the legacy universal calls (e.g. `gotr get get_cases 30`).
- All IDs are now strictly typed as `int64` in client methods and structs (some places previously used `string`).
- `get_cases` now requires `suite_id` (mandatory for projects in multiple-suites mode).
- Response shape changed for some endpoints (e.g. `GetProjectsResponse`, `GetSharedStepsResponse` are now slices instead of objects with a wrapping field).

### Added

- New subcommands in the `get` group:
  - `gotr get case <case-id>` — fetch a single case by ID.
  - `gotr get case-history <case-id>` — fetch case change history.
  - `gotr get sharedstep <step-id>` — fetch a single shared step by ID.
  - `gotr get sharedstep-history <step-id>` — fetch shared-step change history.
  - `gotr get suites` — fetch project test-suite list.
  - `gotr get suite <suite-id>` — fetch a single test suite by ID.
- Support for **positional arguments** for project ID in `cases`, `sharedsteps`, `suites`.
- Explicit and informative `Short`/`Long` hints for all subcommands.
- Required-parameter checks in `RunE` with clear error messages.
- Suites client methods: `GetSuites`, `GetSuite`, `AddSuite`, `UpdateSuite`, `DeleteSuite`.

### Changed

- Improved error handling in the client: StatusCode checked before decoding, informative messages.
- All list responses (projects, cases, shared steps, suites) return a slice directly (an array) instead of an object with a wrapping field.
- Removed redundant wrappers in response structs (GetProjectResponse → Project, GetCaseResponse → Case, etc.).
- `help` hints made as clear as possible: indicate the required ID and where to obtain it.

### Fixed

- Fixed array decoding from the API (projects, shared steps, cases).
- Fixed `MarkFlagRequired` issue — positional arguments now work without conflicting with required flags.
- Fixed `is_deleted` field on Case (now int because API returns 0/1).

---

## [2.0.0] - 2025-12-21

### Breaking Changes

- Changed environment variable prefix from `GOTR_` to `TESTRAIL_` for better compatibility with the TestRail ecosystem (e.g. `TESTRAIL_BASE_URL`, `TESTRAIL_USERNAME`, `TESTRAIL_API_KEY`).
- Removed legacy keys in config and Viper (`testrail_base_url`, `testrail_username`, etc.) — now only `base_url`, `username`, `api_key` are used.

### Added

- Configuration file `~/.gotr/config/default.yaml` with automatic reading (Viper).
- New subcommands in the `config` group:
  - `gotr config init` — create a default config with comments.
  - `gotr config path` — print the config path.
  - `gotr config view` — print config contents.
  - `gotr config edit` — open the config in the default editor (`$EDITOR`).
- Bash completion (via `gotr completion bash`).
- Disabled mandatory checks for utility commands (`config`, `completion`).
- Conditional message output (no "Using config file" line for clean stdout).
- `insecure` support in the config (skip TLS verification).

### Changed

- Unified Viper keys: `base_url`, `username`, `api_key` (no `testrail_` prefix).
- Improved env-variable handling with the `TESTRAIL_` prefix.

### Fixed

- Removed duplicated "Using config file" messages.
- Fixed completion (no garbage from files or output).

### Removed

- Legacy env variables with the `GOTR_` prefix.

---

## [1.0.0] - 2025-12-19 (previous release)

- Base version with `list`, `get`, `add`, etc. commands.
- TestRail API v2 support via an HTTP client.
- Global flags `--url`, `--username`, `--api-key`.
