# Changelog

Все заметные изменения в проекте `gotr` будут документироваться в этом файле.

Формат основан на [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
и проект использует [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

## [3.6.0] - 2026-05-07

### Added — `attachments cleanup`: resume-aware backup and ghost tolerance

- **Resume-aware backup.** When a cleanup is restarted via
  `--resume <RUN_ID>`, the backup phase now reads the existing
  `mapping.json`/`attachments.json` of the in-progress snapshot and
  skips attachments that were already downloaded and hashed in the
  previous attempt. Only the remaining IDs are fetched, hashed and
  appended; partial snapshots are no longer re-downloaded from scratch.
- **Ghost-attachment tolerance.** TestRail occasionally reports an
  attachment ID in listings that immediately returns
  `400 "attachment does not exist"` (or `404`) when fetched — a race
  between listing and a concurrent deletion. Such IDs are now classified
  as **ghosts** via `client.IsAttachmentNotFound`, skipped with a
  `WARN: ghost attachment <id> skipped` line, and recorded in the new
  `BackupResult.GhostIDs` slice so the rest of the run is not aborted.
- **`Skipped` / `Ghost` counters in summaries.** The Markdown and PDF
  cleanup reports gained a `Backed up (skipped, ghosts)` row in the
  Summary block, and `ExecuteResult` exposes `BackupSkipped` and
  `GhostIDs` for downstream tooling.
- **Resume hint on interrupted Execute.** When `attachments cleanup`
  fails after a `RUN_ID` has been issued, the CLI now prints
  `⚠️ Cleanup interrupted. To resume from where it left off, run: gotr attachments cleanup --resume <run-id>`
  before exiting non-zero.

### Added — `attachments cleanup`: snapshot completeness (v2 layout)

- **SHA-256 integrity per attachment.** Every downloaded binary is
  hashed inline during streaming; the digest is persisted in
  `mapping.json` and verified at restore.
- **`mapping.json` (schema_version=2).** Versioned, incremental ledger
  written atomically after each successful download. Records original
  ID, sha256, size, parent entity, file path, compressed flag, and the
  `restorable` boolean (test-bound attachments are flagged
  non-restorable with reason).
- **`references.json` sidecar.** During backup, markdown fields of the
  referencing entities (case `custom_steps*`, `comment`, `description`;
  result `comment`; run/plan/milestone `description`/`comment`) are
  scanned for `index.php?/attachments/get/<id|md5>` links. Matches are
  persisted with their dotted field path so restore can rewrite them.
- **`integrity.json`.** Top-level Merkle index over `data.json`,
  `mapping.json`, `references.json` and every binary under `files/`.
  Restore recomputes and verifies the root before any re-upload.
- **Reference rewrite on rollback.** After re-uploading attachments,
  the rollback fetches each affected case/result/run/plan/milestone,
  substitutes `attachments/get/<old_id>` → `<new_id>` in the recorded
  fields, and pushes the update. Numeric IDs are rewritten; md5-only
  refs are reported `not_rewritten`.
- **`--backup-concurrency N`** flag (default `min(8, --concurrency)`):
  parallelises download+hash+write while a single committer goroutine
  drains journal entries and persists `mapping.json` atomically.
- **`--skip-references`** flag on both backup and `gotr snap rollback`:
  short-circuits both the reference scan and the restore-time rewrite.
- **`--verify-integrity`** flag on rollback: recomputes the Merkle root
  before re-uploading; aborts on mismatch unless `--force`.
- **Backward compat.** Old snapshots (only `data.json`, no
  `mapping.json`/`references.json`/`integrity.json`) restore via the
  legacy v1 path; an `INFO` line announces the downgrade. New
  snapshots set `meta.json.schema_version = 2`.
- Closes [#74].

### Added — `attachments cleanup`: progress UI and per-entity-type summary

- **Multiline live scan UI.** The single-line spinner is replaced by a
  5-line progress block on stderr (when stderr is a TTY and `--quiet`
  is not set) covering project N/M, the current phase
  (`project → suites → cases → runs → plans → tests`) with a 10-char
  bar and a done/total counter, running totals (`found`, `eligible`,
  size), and elapsed time + ETA. Progress events are throttled to
  ~50 ms to avoid stalling the scanner during high-concurrency phases.
  Closes [#72].
- **`INFO`-line fallback** on non-TTY writers / `--quiet`: per-project
  `INFO: project X/Y done: …` and per-chunk
  `INFO: chunk N/M done — running totals: …` lines keep CI logs
  grep-friendly while preserving the same data.
- **Per-project × per-entity-type summary table** is printed after the
  executor finishes (`Breakdown by project × entity type:`) with
  `case run plan plan_entry result test` columns, per-row Total/Size
  and a footer aggregating across the whole run.
- **Final summary block** lists the absolute paths of every audit
  report file, the snapshot id with its absolute path under
  `~/.gotr/snaps/cleanup-attachments/<id>/`, and a `Next steps:`
  section with copy-pasteable rollback (`gotr snap rollback <id>`)
  and resume (`gotr attachments cleanup --resume <run-id>`) hints.
- New `internal/cleanup.ScanProgress` interface +
  `ScanProgressReceiver` capability lets both `ProjectScanner` and
  `EntityScanner` emit per-phase events without leaking UI concerns
  into the scan layer; `BuildPlanWithScannerProgress` and
  `ChunkConfig.Progress` are the new public entry points.
- New `internal/ui.MultilineStatus` component with TTY-aware rendering
  (`golang.org/x/term`), ANSI cursor-up redraw, deterministic
  `RenderForTest` helper for golden tests, and `HumanBytes` /
  `fmtDurationShort` formatters.

[#72]: https://github.com/Korrnals/gotr/issues/72

### Added — `attachments cleanup`: chunked execution, checkpoints, resume

- **Chunked plan-build with crash-safe checkpoints.** Long entity-walk
  scans across many projects are now broken into chunks of
  `--chunk-size` (default `10`). After every chunk the run state is
  flushed to `~/.gotr/cache/cleanup-attachments/<RUN_ID>/checkpoint.json`
  plus `partial-plan.json` via `tmp + fsync + rename`, so a Ctrl-C,
  network failure, or OS crash leaves a recoverable artifact instead
  of a half-built plan. Closes [#73].
- **`--resume <RUN_ID>`** picks up an interrupted run, restoring the
  original scope, filters, scan strategy, limit, and chunk size from
  the checkpoint and only retrying projects in `pending`/`retry_pending`.
  Mutually exclusive with scope/filter flags — those are restored from
  the checkpoint, not re-supplied. Filter mismatches abort with a
  `checkpoint mismatch` error rather than silently mixing two filters.
- **`--scan-timeout-per-project <dur>`** (default `10m`) caps each
  per-project scan; on timeout the project is recorded as `timeout`
  in the checkpoint and the run continues. `--resume` re-tries it.
  Pass `0` to disable the deadline.
- **`--list-checkpoints`** prints a table of all known checkpoints
  (RUN_ID, started, updated, totals, done, pending, failed, timeout)
  and the `--resume` hint. Mutually exclusive with every other flag.
- **Auto-cleanup on success.** A run that completes with every project
  in `done` automatically removes its checkpoint directory; otherwise
  the artifacts are preserved and a `WARN:` summary is printed.
- **Per-chunk progress.** `INFO: chunk N/M done — running totals: X
  projects with hits, Y attachments, Z bytes` is logged to stderr
  after every flushed chunk.

### Documentation

- New "Recovery & resume" section in `docs/en/guides/commands/attachments.md`
  and `docs/ru/guides/commands/attachments.md` covering run-id format,
  checkpoint location, atomic write semantics, and the resume contract.

[#73]: https://github.com/Korrnals/gotr/issues/73

---

## [3.5.2] - 2026-05-05

### Fixed — entity scanner: complete coverage of run/result attachments

- **Run-bound attachments are no longer dropped.** TestRail Server
  builds (notably self-hosted < 7.5) return
  `get_attachments_for_run/{id}` payloads with `run_id`/`case_id`/
  `result_id` set to `null`. The previous scanner relied on those
  fields, so `InferredEntityType()` returned an empty string and
  `--entity-type run` filtered every run-bound attachment out. The
  scanner now stamps the parent ID after each per-entity fetch.
- **Result-/test-bound attachments are now scanned.** The entity
  scanner only walked `cases → runs → plans`; result-bound files were
  reachable solely via `get_attachments_for_test/{test_id}` and were
  silently skipped. A new tests walk enumerates every run
  (project-level + plan entries) → tests → attachments, and
  `--entity-type result|test` now route to it.
- **Mapping change:** `EntityScannerOptionsFromTypes` no longer
  conflates `result`/`test` with the case walk. The new
  `EntityScannerOptions.WalkTests` flag drives the dedicated walk.
  Default behavior (no `--entity-type` filter) walks **all four**
  kinds (cases, runs, plans, tests). v3.5.1 is impacted; deletion
  runs that omitted `--entity-type` would have missed run/result
  attachments — re-run cleanup on v3.5.2 if you relied on that path.
- **Regression tests:** `TestEntityScanner_StampsRunIDOnRunBoundAttachments`,
  `TestEntityScanner_WalksTestsForResultBoundAttachments`,
  `TestEntityScanner_CollectRunIDsFromPlanEntries`,
  `TestEntityScanner_WalksPlanEntriesForEntryBoundAttachments` lock
  every fix.
- **Plan-entry attachments are now scanned.** The plan walk previously
  only called `get_attachments_for_plan/{id}`, so attachments uploaded
  via `add_attachment_to_plan_entry/{plan_id}/{entry_id}` were
  unreachable for `--entity-type plan_entry`. The walk now expands
  every plan via `GetPlan` and fetches
  `get_attachments_for_plan_entry/{plan_id}/{entry_id}` per entry,
  stamping `plan_id` + `entry_id` on each result.

### Changed — `attachments cleanup --entity-type` default is now ALL kinds

- **Default scope is the full set** `case,run,plan,plan_entry,result,test`
  (was `result` only). This matches the entity-scanner fix above so
  invocations without `--entity-type` no longer silently skip run/case/
  plan-bound attachments.
- **Mandatory scope notice** is printed before scanning. When the
  resolved scope covers ALL kinds the notice is rendered as a ⚠️
  warning with a hint on how to narrow with `--entity-type`. Narrower
  scopes get a one-line ℹ️ confirmation.
- **Interactive prompt** now warns before asking and offers the
  presets `all` / `case` / `run` / `plan,plan_entry` / `result,test` /
  `custom` so users can correct the scope in one keystroke.
- The pre-flight summary table and the final `Proceed with deletion?`
  confirmation already exist and continue to gate destructive runs.

### Removed

- **v3.5.1 release** is withdrawn from GitHub Releases after v3.5.2
  binaries are published — the v3.5.1 entity scanner skipped
  run/result attachments. Use v3.5.2.

---

## [3.5.1] - 2026-05-05

### Fixed — emit cleanup deletion report

- **`gotr attachments cleanup` now writes an audit report.** After every
  invocation (including `--dry-run`) the command emits the deletion
  audit in four formats — Markdown, JSON, CSV and PDF — under
  `~/.gotr/reports/cleanup-attachments/<label>/<YYYY-MM>/`. The report
  contains the run header (snapshot ID, label, CLI args, gotr version),
  the applied filters, a summary, a per-project breakdown, the full
  list of deleted attachments (id, name, size, parent, created date)
  and any per-attachment failures. The Markdown rendition includes the
  rollback command, and `INDEX.md` is refreshed automatically.
- **`--no-report`** opts out of writing the report files (useful for
  CI pipelines that capture stdout only).
- **Dry-run reports** are flagged with a `DRY-RUN` marker in the
  Markdown header and `dry_run=true,deleted=false` in CSV rows so they
  cannot be confused with the real artefact.
- **Backward compatible.** Default behaviour change is the addition of
  four files per cleanup run; no existing flags or output paths
  changed.

### Internal

- New package `internal/report/cleanup` (types + Markdown / JSON / CSV
  renderers + writer) and a `pdf.NewCleanupGenerator()` PDF generator
  reusing the embedded NotoSans fonts.
- New report category `cleanup-attachments` added to
  `internal/report/paths.go` and `internal/retention/exports.go` (with
  `.csv` extension support).

---

## [3.5.0] - 2026-05-04

### Added — bulk attachments cleanup with snapshot + rollback

- **New command `gotr attachments cleanup`** — removes attachments older
  than a configurable cutoff across one project, several projects, or
  every visible project. Selection can be filtered by parent kind
  (`case`, `run`, `plan`, `plan_entry`, `result`, `test`) and capped
  via `--limit`.

- **Default snapshot + rollback safety net.** Before any deletion the
  command stores a snapshot of every selected attachment (binary +
  metadata) under the new snap category `cleanup-attachments`. The
  standard snap workflow restores them:
  ```bash
  gotr snap list --category cleanup-attachments
  gotr snap rollback <snap-id>
  ```
  Rollback re-uploads each binary to its original parent and records the
  old → new attachment ID mapping. Test-bound attachments (no
  `add_attachment_to_test` endpoint exists in TestRail) are reported as
  **Skipped** with the original IDs preserved in the failure list.

- **Seven-day default retention for cleanup snapshots.** `snap gc` now
  honors a per-category TTL map (`snap.retention.category_ttl_days`)
  with a built-in default of 7 days for `cleanup-attachments`. The
  global `--ttl-days` override and existing `protected_prefixes` /
  `frozen_snapshots` rules continue to apply unchanged.

- **Compatibility with TestRail Server &lt; 7.5.** The cleanup walker now
  ships **two scan strategies** behind a new `--scan-strategy` flag:
  `project` (single bulk `get_attachments_for_project` call, TestRail
  7.5+ / Cloud) and `entities` (walk
  `get_suites → get_cases → get_attachments_for_case`, plus optional
  `get_runs` / `get_plans` based on `--entity-type`, deduplicated by
  attachment ID, TestRail 5.7+). The default is `auto`: the command
  probes the project endpoint once on the first project and falls back
  to the entity walk **only** when the canonical
  `404 Unknown method 'get_attachments_for_project'` is returned. Any
  other probe error aborts the run — no silent fallback. The fallback
  is announced with an `INFO:` line on stderr; `--scan-strategy=project`
  or `--scan-strategy=entities` pins the strategy and skips the probe.

- **Interactive survey** mirrors every CLI flag (project scope, parent
  kinds, `--older-than`, concurrency, snapshot toggle, retention,
  dry-run) and is gated by a TTY check — non-interactive contexts skip
  the survey entirely. The final pre-flight confirmation is suppressed
  by `--force` or `--dry-run`.

### Added — Attachment model & client

- `data.Attachment` gained `EntryID`, `TestID`, `EntityType`, and
  `EntityID` fields plus an `InferredEntityType()` helper that derives
  the parent kind from whichever ID is populated.
- All `client.HTTPClient.GetAttachmentsFor*` methods now use the
  generic paginator and stream every page of results.

### Fixed — multipart attachment upload

- `client.HTTPClient.DoRequest` now extracts the `Content-Type`
  override from `queryParams` **before** building the URL query string.
  Previously the multipart `Content-Type` header (with its boundary
  parameter) leaked into the GET parameters, and TestRail rejected the
  upload with `Invalid characters in GET: [Content-Type]
  [multipart/form-data; boundary=...]`. As a result, `gotr attachments
  add` and the cleanup snapshot rollback path (which re-uploads
  binaries) now work against TestRail Server again.

---

## [3.4.0] - 2026-05-04

<details>
<summary>Details</summary>

### Added — multi-snap migration bundles & full-state cross-machine transfer

- **`gotr export migration-archive` with no arguments** now defaults to
  packing the **full state** of `~/.gotr/`: all snapshots from the local
  store + the entire `~/.gotr/reports/` tree + the entire `~/.gotr/logs/`
  tree. The archive has type `migration_bundle` and is saved into the new
  directory **`~/.gotr/exports/all/`** under the name
  `migration_bundle_all_<N>snaps_<UTC-timestamp>.tar.gz`. This is the
  recommended artifact for transferring `gotr` state between machines
  with a single command.

- **Selectors for multi-snap mode**:
  - `gotr export migration-archive --label <substr>` — packs all
    snapshots whose `label` contains the substring (case-insensitive).
  - `gotr export migration-archive <id1> <id2> ...` — packs the
    listed snapshots + the union of their reports and logs.
  - `gotr export migration-archive --all` — explicit form of the default.
  For a single `<id>` the behavior is unchanged: the archive stays in
  `~/.gotr/exports/snaps/`.

- **`gotr import migration-archive`** automatically detects the archive
  type (single-snap vs `migration_bundle`) from its manifest and picks
  the correct restore path — no flags required. Import still requires
  an explicit file-path argument; extraction always goes into the
  system directory `~/.gotr/` (or `$GOTR_HOME`). In interactive mode the
  picker shows archives from both directories: `exports/snaps/` and
  `exports/all/`.

### Fixed — manifest auto-registration after import

- **`gotr import` now automatically registers imported snapshots in
  `~/.gotr/snaps/manifest.json`** (store index), so `gotr snap list`
  immediately shows them without needing to run
  `gotr snap manifest repair`. Works for both single-snap imports
  (`Kind=snap`) and multi-snap bundles (`Kind=migration_bundle`).
  With `--overwrite` the stale manifest entry is replaced by the fresh
  one (no duplicates). If an individual snap's `meta.json` is unreadable,
  indexing is skipped without error — data remains on disk and will be
  picked up by a subsequent `manifest repair`.

- **`gotr snap manifest repair`** now performs a single atomic
  `manifest.save()` per run instead of O(N) writes for each
  orphan/missing entry. On a store with thousands of snapshots this
  removes O(N²) disk load — observed speed-up from tens of seconds
  to tenths of a second (`internal/snap.AddMany` / `RemoveMany`).

### Changed — `~/.gotr/exports/` layout

- Added constant `paths.ExportsAllSubdir = "all"` and helpers
  `paths.ExportsAllDirPath()` / `paths.EnsureExportsAllDirPath()`.
  The `~/.gotr/exports/` tree is now:

  ```text
  ~/.gotr/exports/
  ├── snaps/    — single snap_<id>_<ts>.tar.gz / migration_<id>_<ts>.tar.gz
  ├── reports/  — exported reports (.zip / .pdf / .md / .json)
  ├── api/      — raw API dumps (`gotr export <resource> --save`)
  └── all/      — full / multi-snap migration_bundle_*.tar.gz (new)
  ```

  Old migration bundles remain functional — import accepts an explicit
  file path regardless of location.

</details>

## [3.3.2] - 2026-04-28

<details>
<summary>Details</summary>

Patch release combining two hotfixes.

### Fixed

- **Case import order** (`internal/service/migration/import.go`).
  `ImportCases` and `ImportCasesReport` create cases in parallel
  (`maxImportConcurrency = 10`); TestRail records insertion order from
  `add_case` calls, so non-deterministic goroutine scheduling scrambled
  the original order. After the parallel creation phase, import now calls
  `move_cases_to_section` for each dst-section with case_ids sorted by
  original index in `filtered`. Sections with `≤1` case are skipped.
  Reorder errors are logged as `WARN` but do not abort the migration
  (data is already imported; order is a UX concern). Regression covered
  by 4 unit tests in `internal/service/migration/import_order_test.go`.

- **Snapshot manifest drift recovery**.
  Added `gotr snap manifest repair [--dry-run]` to reconcile
  `~/.gotr/snaps/manifest.json` with on-disk snapshot directories.
  The command re-indexes valid snapshots missing from the manifest,
  removes orphan manifest entries whose directories no longer exist,
  and reports unreadable `meta.json` entries without destructive changes.

- **Atomic manifest write race** (`internal/snap/store.go`).
  Replaced fixed `manifest.json.tmp` temp-file name with a unique
  temp file via `os.CreateTemp`, eliminating rename failures under
  concurrent writers or overlapping processes.

- **Test isolation for local snapshot store**.
  Snapshot-related tests no longer contaminate the operator's real
  `~/.gotr` data. Test-mode path sandboxing redirects home paths to
  per-process temporary directories; explicit per-test overrides still
  work as expected.

</details>

## [3.3.0] - 2026-04-24

<details>
<summary>Details</summary>

UX polish релиз (issue #44): категоризованная иерархия отчётов и
экспортов, shell completion, интерактивный режим, retention/cleanup,
управление предупреждениями и корпоративный TLS (`ca_bundle`).

### Added — иерархия `~/.gotr/reports/` и `~/.gotr/exports/`

- **Категоризация отчётов** в дерево
  `~/.gotr/reports/<category>/<label|default>/<YYYY-MM>/<file>`.
  Новые категории: `migrations`, `coverage`, `rollbacks`, `no-snapshot`,
  `testrail/p<N>`, `_unclassified`. Классификация выполняется в
  `internal/report.ClassifyReport` по шаблонам имён файлов.
- **`gotr report organize [--dry-run]`** — миграция старого «плоского»
  layout в новую иерархию. Идемпотентна; конфликты (уже существующий
  файл в целевом пути) скипаются. `--dry-run` печатает план без
  изменений на диске. После успешного переноса вызывает `Reindex`.
- **Иерархия экспортов** `~/.gotr/exports/{snaps,reports,api}/`.
  `snaps/` — tar.gz-бандлы, `reports/` — zip-бандлы и plain-копии,
  `api/<resource>/` — legacy-выгрузки `gotr get plans/reports/...`.
  Миграция — `gotr export organize [--dry-run]`.
- **`gotr export snap --with-reports`** — по умолчанию ON. Сканирует
  reports-директорию рекурсивно, встраивает в архив файлы, чья
  basename содержит `filepath.Base(snapID)`. `--no-reports` —
  opt-out. Архивный префикс: `reports/<rel>`. Результат отражается
  в `manifest.Files`.
- **`gotr cleanup {reports,snaps,exports,all} [--dry-run]`** —
  ручной executor для retention-политик. Конфигурация в
  `retention.{reports,snaps,exports}` (см. ниже). `snaps` делегирует
  существующему `gotr snap gc`.
- **`gotr report show --print`** — вывод содержимого отчёта в stdout
  (cat-like) независимо от расширения для md/json/txt. Бинарные
  PDF явно отклоняются с понятной ошибкой.
- **INDEX.md**: автоматически регенерируется после generate/import/organize,
  содержит ссылки на все отчёты в иерархии.

### Added — shell completion

- Динамический `ValidArgsFunction` для `report show/view`,
  `export report/snap`, `import snap/report`. Рекурсивный листинг
  через `intreport.RecursiveListReports`; для snap-команд — по
  `snap.LoadManifest`; для import — файлы с расширениями
  `.zip/.pdf/.md/.json` или `.tar.gz/.tgz` в соответствующих
  директориях. Handles two-dot `.tar.gz`.

### Added — interactive mode (TTY-guard)

- `report show/view`, `export report/snap`, `import snap/report`
  теперь принимают `cobra.MaximumNArgs(1)`. Если argument отсутствует,
  stdin — TTY и `--non-interactive` не установлен, пользователю
  показывается survey-prompt со списком кандидатов. В
  non-interactive режиме и без TTY — явная ошибка с подсказкой
  «pass as argument or run interactively».

### Added — warnings suppression + TLS

- **`ui.suppress_warnings: []`** (list of keys) — подавление отдельных
  не-критичных предупреждений. Ключи:
  - `tls_insecure` — баннер при `insecure=true` или `tls.insecure=true`,
  - `deprecation` — зарезервирован,
  - `flat_layout` — подсказка при обнаружении плоского layout.
- **`--show-warnings`** CLI-флаг (`show_warnings` viper key) —
  временный override, показывает все варнинги независимо от конфига.
- **`tls.insecure`** — новый config-ключ. Старый top-level `insecure`
  и флаг `--insecure` сохранены для обратной совместимости;
  включающий источник побеждает.
- **`tls.ca_bundle: "/path/to/ca.pem"`** — корпоративный CA. Путь
  читается, парсится `x509.NewCertPool` + `AppendCertsFromPEM`,
  подставляется в `tls.Config.RootCAs`. Предпочтительная альтернатива
  `insecure=true`. `client.WithCABundle(path)` — публичная опция.
- При первом показе каждого варнинга добавляется one-time tip:
  «add '<key>' to ui.suppress_warnings to silence this warning».
- Флаг «показывали про flat layout» теперь персистентный —
  `~/.gotr/state.json::flat_layout_warned`. Показывается один раз за
  инсталляцию.

### Added — retention/cleanup конфиг

```yaml
retention:
  reports:
    enabled: false          # по умолчанию ВЫКЛЮЧЕНО (безопасно)
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

- Отсутствие секции `retention` не является ошибкой: дефолты
  подставляются автоматически.
- `keep_categories` — whitelist: такие категории никогда не
  удаляются retention-политикой (важно для coverage-артефактов).

### Changed

- **`internal/paths`**: введены `ExportsSnapsDirPath`,
  `ExportsReportsDirPath`, `ExportsAPIDirPath` + Ensure\*-варианты.
  Writers перенаправлены: `internal/snapbundle.DefaultExportPath`,
  `internal/reportbundle.ExportSingle`/`ExportAll`,
  `internal/output.GetExportsDir(resource)` теперь пишут в
  подкатегории `exports/`.
- **`cmd/report/list`**: `--filter` применяется по basename (glob)
  ИЛИ по substring в relative path; листинг — рекурсивный.
- **`cmd/report/show`**: `openWithOS` использует `exec.Cmd.Run()`
  вместо `Start()`, так что нулевой exit OS-лаунчера
  (`xdg-open`/`open`/`rundll32`) пропагируется как ошибка CLI.
- **`cmd/root.PersistentPreRunE`**: warnings registry инициализируется
  до любого другого вывода; баннер `tls_insecure` теперь идёт через
  `warnings.Emitf` и, соответственно, уважает `suppress_warnings`.
  Прямая `fmt.Fprintln(os.Stderr, "WARNING: TLS...")` из
  `internal/client` удалена.

### Fixed

- Снятие false-positive при первом показе flat-layout подсказки, если
  пользователь уже мигрировал иерархию и удалил `state.json` вручную:
  `warnings.Emitf` дополнительно блокирует повторный вывод в рамках
  одного процесса через in-memory `shownHint` map.

### Migration notes (v3.2 → v3.3)

1. **Reports layout.** При первом запуске любой команды `report list/show`
   будет показано однократное предупреждение о «плоском» layout, если
   в `~/.gotr/reports/` обнаружены файлы в корне. Рекомендуется:
   ```bash
   gotr report organize --dry-run   # посмотреть план
   gotr report organize             # выполнить миграцию
   ```
   Команда не удаляет ничего: при коллизии оригинал остаётся в корне,
   счётчик `skipped` увеличивается. После успеха `INDEX.md` регенерируется.
2. **Exports layout.** Аналогично:
   ```bash
   gotr export organize --dry-run
   gotr export organize
   ```
   Классификатор перемещает `*.tar.gz|*.tgz` → `exports/snaps/`,
   `*.zip|*.pdf|*.md|*.json` → `exports/reports/`, директории ресурсов
   (plans/, reports/, runs/…) — в `exports/api/`.
3. **Insecure TLS.** Старый `insecure: true` и `--insecure` продолжают
   работать без изменений. Рекомендуется миграция:
   ```yaml
   tls:
     insecure: false
     ca_bundle: "/etc/ssl/corp-ca.pem"
   ```
4. **Подавление предупреждений.** Было: глобальное `no_warnings: true`
   (в плане). Финальное решение — per-key list:
   ```yaml
   ui:
     suppress_warnings: [tls_insecure, flat_layout]
   ```
   Флаг `--show-warnings` показывает все варнинги независимо от списка.
5. **Retention.** По умолчанию выключен, никакие старые артефакты не
   удаляются автоматически. Для миграции на новую политику — явно
   задать `retention.*.enabled: true` и прогнать `gotr cleanup all
   --dry-run` перед боевым запуском.

### Internal

- Новые пакеты: `internal/warnings` (registry), `internal/state`
  (persistent JSON store), `internal/exportsorg` (migrator), 
  `internal/retention` (policy + executor).
- E2E-тесты жизненного цикла: `internal/report/e2e_lifecycle_test.go`
  (flat→hierarchy, 6 категорий, идемпотентность),
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

- **Coverage-gate no longer false-negatives on name-resolved sections.**
  `internal/service/migration/import.go`:
  `resolveSectionMapByName` now registers every `src→dst` pair it resolves by
  section name into `m.mapping` (`AddPair(src, dst, "existing")`). Previously
  the map lived only as a local variable consumed by `ImportCasesReport`,
  while `VerifyCasesCoverage → resolveDstSectionIDForFilter` consulted
  `m.mapping` exclusively and therefore reported all cases as `missing`
  after a successful migration. Regression test:
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

</details>

## [3.2.0] - 2026-04-23

<details>
<summary>Details</summary>

Полный багфикс миграции TestRail: устраняет скрытые расхождения счётчиков
импорта, вызванные ошибочным «молчаливым» поведением фильтрации
и парентинга секций в движке миграции.

### Fixed — migration engine

- **Multiset-matching по `(dst_section_id, compare_field)`.**
  До 3.2.0 фильтр воспринимал одинаковое значение `title` внутри одной
  `section_id` как «уже есть в target» и пропускал source-кейсы с тем же
  заголовком, даже если в target этого кейса не было (или был в другом
  scope). Новая реализация (`internal/service/migration/match.go`,
  `filter.go`) использует мультимножество с FIFO-поглощением: каждый
  source-кейс потребляет ровно один target-кейс по совпадающему ключу.
- **Резолвинг dst-scope стал строгим.** `resolveDstSectionIDForFilter`
  больше не «схлопывает» несмапленные секции в target root — для неразрешимого
  `section_id` выдаётся отрицательный sentinel `-srcSectionID`,
  гарантирующий, что такие source-кейсы не будут ошибочно сматчены
  с кейсами чужих секций.
- **Отказ от silent-root-fallback при импорте секций.**
  `ImportSections`: секция с несмапленным `parent_id` теперь отклоняется
  с ошибкой и учитывается в `failedImports`, вместо того чтобы молча
  перепарентиться в root.
- **`FailedCount()` отражает реальные отказы импорта.**
  `ImportCases`, `ImportCasesReport`, `ImportSections`, `ImportSuites`,
  `ImportSharedSteps` теперь инкрементируют `failedImports` при каждой
  ошибке API или отказе от импорта. `max(len(errs), FailedCount())`
  в отчётах sync-команд теперь корректно превращает пропущенные
  импорты из невидимых «skipped» в явные `errors`/`failed`.

### Added — compare pipeline

- **`--suite1` / `--suite2` (persistent).** На `gotr compare *`
  можно зафиксировать отдельный suite для каждого проекта
  (`0 = all suites`, как раньше). На мульти-suite целях сравнение
  больше не будет молча идти через разные scope.
- **`--match-field` (persistent, shared).** Новый унифицированный
  флаг поля сравнения — имеет приоритет над `--field`, применяется
  единообразно во всех `compare` подкомандах.
- **`filterSuitesByID`**: для неизвестного suite-id возвращается пустой
  scope (а не fallback на все suites проекта) — отказ от silent-fallback
  на уровне сравнения.

### Added — sync safety gate

- **`--verify-coverage` (opt-in, default `false`)** на `sync cases` и
  `sync full`. После импорта повторно фетчит target и проверяет, что
  каждый source-кейс имеет совпадение по multiset-ключу; при пробеле
  выходит с ошибкой `coverage gap: ...` (non-zero exit code) и лог-строками
  `  - [id] "title" (src_section=X, dst_section=Y)` (до 50 штук).
  Гейт защищает от повторения этого класса ошибок в будущем.
- **`Migration.VerifyCasesCoverage`** (`internal/service/migration/coverage.go`):
  автономный API-метод, который используется внутри `runCoverageGate`
  и может быть вызван из пользовательского кода.

### Added — interactive UX

- **`SelectMatchField`** (`internal/interactive/match_field.go`):
  новый TTY-guarded промпт выбора поля сравнения для `compare`/`sync`
  с нормализацией (`MatchFieldCases`, `MatchFieldSections`,
  `MatchFieldSuites`, `MatchFieldSharedSteps`).

### Changed

- Рефакторинг `cmd/compare/cases.go`: вынесены `resolveCompareField` и
  `applySuiteScope` хелперы — снижение цикломатической сложности
  `newCasesCmd` и `compareCasesInternal`.
- `internal/service/migration/filter.go`: tag-less `switch` в резолвере
  `dstParentID` заменён на `if/else` (staticcheck QF1002).

### Tests

- Новые регрессионные тесты (`internal/service/migration/import_test.go`,
  `cmd/sync/sync_helpers_coverage_test.go`, `cmd/sync/sync_flags_test.go`)
  фиксируют контракт:
  - source-кейс с неразрешимым `SectionID` → `FailedCount++`, `AddCase`
    не вызывается.
  - source-секция с несмапленным `ParentID` → отклонение, учёт отказа.
  - ошибка `AddCase` одновременно в `[]errs` и `FailedCount`.
  - `--verify-coverage` по умолчанию `false`, гейт-no-op без флага.
  - пробой покрытия возвращает ошибку с подстрокой `"coverage gap"`
    (стабильный контракт для CI/grep).

### Docs

- Обновлён `README.md` / `README_ru.md`: новые флаги и coverage-gate.
- Обновлены `docs/{en,ru}/guides/commands/{compare,sync}.md`.

---

</details>

## [3.1.0] - 2026-04-19

<details>
<summary>Details</summary>

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

</details>

## [3.0.1] - 2026-04-12

<details>
<summary>Details</summary>

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

</details>

## [3.0.0] - 2026-04-09

<details>
<summary>Details</summary>

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

</details>

## [3.0.0] - 2026-03-12

<details>
<summary>Details</summary>

### Added

#### Stage 6.8: Concurrency Unification & Compare Subcommands

- **`internal/concurrency/`** — новый unified concurrency-пакет (переименован из `internal/parallel/`)
  - Три уровня стратегий:
    - `FetchParallel[T]` — лёгкая: один API-вызов на проект, параллельная загрузка P1+P2
    - `FetchParallelBySuite[T]` — средняя: per-suite запросы (для `sections`)
    - `FetchParallelPaginated` — тяжёлая: `ParallelController` с пагинацией (для `cases`)
  - Generic API через Go generics `[T any]`

- **`pkg/reporter/`** — универсальный reporter вынесен в публичный пакет (из `internal/ui/reporter/`)
  - Builder pattern: `Section` / `Stat` / `StatIf` / `StatFmt` / `Print`
  - go-pretty/v6 для выровненного boxed-вывода

- **Generic `newSimpleCompareCmd`** — одна generic-фабрика вместо 9 идентичных файлов (`cmd/compare/simple.go`)
  - Устранено ~1200 строк копипасты
  - Все простые подкоманды используют `FetchParallel[T]` для параллельной загрузки проектов

- **`compare sections`** — параллельная загрузка секций по сьютам через `FetchParallelBySuite[T]`

- **`compare all`** — единообразный вывод через `pkg/reporter`, partial results при недоступных API

### Changed

- `internal/parallel/` → `internal/concurrency/` (переименование пакета и всех импортов)
- `internal/ui/reporter/` → `pkg/reporter/` (вынесен в публичный пакет)
- Все 13 compare-подкоманд используют `pkg/reporter` для вывода статистики
- `OnSuiteComplete` → `OnItemComplete` в интерфейсе `ProgressReporter`
- Дефолтные значения: `parallel-suites=10`, `parallel-pages=6` (стабильные для TestRail Server)

### Fixed

- `compare all` более не использует `fmt.Println` с emoji и box-drawing символами
- Устранено некорректное выравнивание статистики в терминалах без поддержки emoji

### Performance

- Простые compare-подкоманды (runs, plans, milestones и др.): загрузка P1 и P2 **параллельно**
- `compare sections`: параллельная загрузка по сьютам вместо последовательной

---

#### Stage 6.9: Generic Paginator & Pagination Audit

### Added

- **`internal/client/paginator.go`** — универсальный generic-пагинатор `fetchAllPages[T]`
  - Обрабатывает оба формата TestRail API без ветвлений в бизнес-логике:
    - **Paginated wrapper** (TestRail 6.7+): `{"offset":0,"limit":250,"size":N,"<key>":[...]}`
    - **Flat array** (старые TestRail Server): `[item1, item2, ...]`
  - Автоматическое определение формата по первому байту ответа
  - Стандартный размер страницы: 250 элементов (TestRail default)
  - Выход по условию: `len(page) < limit` (последняя страница)

- **Миграция 9 критичных list-методов** на `fetchAllPages[T]`:
  - `GetRuns(projectID)` — runs теперь не обрезаются на 250
  - `GetPlans(projectID)` — планы теперь не обрезаются на 250
  - `GetSections(projectID, suiteID)` — секции (критично для `compare sections`)
  - `GetSharedSteps(projectID)` — shared steps
  - `GetMilestones(projectID)` — milestones
  - `GetResults(runID)` — результаты рана
  - `GetResultsForRun(runID)` — расширенный вариант
  - `GetTests(runID)` — тесты рана
  - `GetSuites(projectID)` — сьюты проекта

### Changed

- Все 9 мигрированных методов: тело метода упрощено с ~30 строк ручного цикла до 1 вызова `fetchAllPages`
- Удалено ~145 строк дублированного pagination boilerplate из `internal/client/`

### Tests

- `internal/client/paginator_test.go` — 11 новых unit-тестов:
  - Оба формата ответа (paginated wrapper и flat array)
  - Многостраничная загрузка (multi-page accumulation)
  - Граничные случаи: пустой ответ, последняя неполная страница
  - Тест на ошибку сервера (HTTP 500)
  - Table-driven tests для `decodeListResponse`

### Verified

- `compare all --pid1 30 --pid2 34`: 20 509 кейсов (87 стр.) + 116 009 кейсов (475 стр.) — пагинация подтверждена на реальных данных
- `compare runs`, `compare plans`, `compare milestones`, `compare sections`, `compare sharedsteps`: все работают корректно
- `go test ./...` — все тесты зелёные

---

#### Stage 7.0: Context Propagation

### Added

- **`context.Context`** во все ~100 методов `ClientInterface`
  - `signal.NotifyContext` → корректное завершение по Ctrl+C
  - Контекст пробрасывается CLI → Service → Client → HTTP

### Changed

- Все API-методы принимают `ctx context.Context` первым аргументом
- `cmd.ExecuteContext()` вместо `cmd.Execute()`
- `MockClient` обновлён под новые сигнатуры

---

#### Stage 8.0: UI/Output Refactoring

### Added

- **`internal/ui/`** — универсальные хелперы:
  - `ui.Table(headers, rows)` — обёртка над go-pretty вместо tabwriter
  - `ui.JSON(v)` — форматированный JSON-вывод
  - `ui.Success()`, `ui.Warn()`, `ui.Error()`, `ui.Info()` — цветные сообщения
  - `ui.Print()`, `ui.Printf()`, `ui.Println()` — обёртки стандартного вывода
- **`--format` PersistentFlag** — глобальный флаг формата вывода на root-уровне
- Массовая миграция: `tabwriter` → `ui.Table`, `json.MarshalIndent` → `ui.JSON`, `fmt.Print*` → `ui.*` (49 файлов)

### Changed

- `internal/flags/`: `*Var` → `GetFlag`, `ValidateRequiredID`
- `os.Exit` → `panic` в `GetClient*` (тестируемость)
- Все error messages переведены на английский

---

</details>

## [2.7.0] - 2026-02-20

<details>
<summary>Details</summary>

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
  - **Page-level progress**: GetCasesWithProgress updates after each 250 cases page
  
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

- **Автоматическая синхронизация версии в Makefile:**
  - Команда `make build` теперь извлекает версию из `cmd/root.go` (единый источник правды)
  - Для релизных версий (без `-dev`) автоматически создаётся/проверяется git tag
  - Приоритет версии: 1) `make build VERSION=x`, 2) версия из кода, 3) git tag
  - Нормализация тега: поддержка `VERSION=v2.7.0` и `VERSION=2.7.0`

#### Stage 4: Complete API Coverage (106/106 endpoints)

- **Attachments API** — 5 endpoints:
  - `AddAttachmentToCase`, `AddAttachmentToPlan`, `AddAttachmentToPlanEntry`
  - `AddAttachmentToResult`, `AddAttachmentToRun`
  - Поддержка multipart/form-data для загрузки файлов
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

**Всего реализовано:** 44 новых endpoint'а  
**Общее покрытие:** 106/106 endpoint'ов TestRail API (100%)

### Added

#### Dry-run режим

- **Флаг** `--dry-run` — единый флаг для всех команд, изменяющих состояние:
  - `add` — project, suite, section, case, run, result, shared-step
  - `update` — project, suite, section, case, run, shared-step
  - `delete` — project, suite, section, case, run, shared-step
  - `run create/update/close/delete`
  - `result add/add-case/add-bulk`
- **Пакет** `cmd/common/dryrun/` — централизованное форматирование вывода dry-run

#### Интерактивный wizard mode

- **Флаг** `--interactive/-i` — интерактивный режим для команд:
  - `add` — project, suite, case, run
  - `update` — project, suite, case
- **Пакет** `cmd/common/wizard/` — библиотека интерактивных prompt'ов на survey/v2
- Паттерн: ввод → предпросмотр → подтверждение/отмена

### Changed

- **Флаг** `-i` теперь используется для `--interactive` (вместо `--insecure`)
- **Флаг** `--insecure` — только длинная форма (без shorthand)

---

</details>

## [2.5.0] - 2026-02-05

<details>
<summary>Details</summary>

### Added

#### Интерактивный режим

- **Команда** `gotr run list` — интерактивный выбор проекта при отсутствии аргументов
- **Команда** `gotr result list` — интерактивный выбор проекта → test run
- **Пакет** `internal/interactive/` — единый механизм интерактивного выбора

#### Client Interface + Mock (Архитектурное улучшение)

- **Пакет** `internal/client/interfaces.go` — полный композитный интерфейс:
  - `ProjectsAPI` — 5 методов
  - `CasesAPI` — 14 методов
  - `SuitesAPI` — 5 методов
  - `SectionsAPI` — 5 методов
  - `SharedStepsAPI` — 6 методов
  - `RunsAPI` — 6 методов
  - `ResultsAPI` — 7 методов
- **Пакет** `internal/client/mock.go` — полный `MockClient` (43 метода)
- Проверка компиляции: `var _ ClientInterface = (*HTTPClient)(nil)`

#### Общие утилиты (Рефакторинг)

- **Пакет** `cmd/common/client.go` — `ClientAccessor` для единого доступа к HTTP клиенту
- **Пакет** `cmd/common/flags.go` — общие функции парсинга флагов
- Рефакторинг `cmd/result/`, `cmd/run/`, `cmd/sync/` — использование `common.ClientAccessor`
- Удалено дублирование `getClientSafe` из 3 пакетов

### Fixed

#### Унификация интерфейсов миграции

- **Удалён дублирующий пакет** `internal/migration` (оставлен `internal/service/migration`)
- **Унифицирован интерфейс** — `internal/service/migration` теперь использует `client.ClientInterface`
- **Обновлён `MockClient`** — дефолтные возвращаемые значения предотвращают nil pointer dereference
- **Рефакторинг sync тестов** — все 10 тестов переписаны с использованием `client.MockClient`
- Убраны пропуски тестов (`t.Skip`) — все тесты проходят

### Changed

- **README.md** — реструктурировано описание, acknowledgements перенесены в конец
- Версия обновлена до `2.5.0`

---

</details>

## [2.4.0] - 2026-02-04

<details>
<summary>Details</summary>

### Added

#### Results API (Полная реализация)

- **Новый client** `internal/client/results.go` с методами:
  - `AddResult` — добавление результата для теста
  - `AddResultForCase` — добавление результата для кейса в run
  - `AddResults` — массовое добавление результатов (bulk)
  - `AddResultsForCases` — массовое добавление для кейсов (bulk)
  - `GetResults` — получение результатов для теста
  - `GetResultsForRun` — получение всех результатов run
  - `GetResultsForCase` — получение результатов для кейса в run

#### Runs API (Полная реализация)

- **Новый client** `internal/client/runs.go` с методами:
  - `GetRun` — получение информации о run
  - `GetRuns` — список runs проекта
  - `AddRun` — создание нового run
  - `UpdateRun` — обновление существующего run
  - `CloseRun` — закрытие run
  - `DeleteRun` — удаление run

#### CLI команды для Results

- **Новый пакет** `cmd/result/` с командами:
  - `gotr result get <test-id>` — получить результаты
  - `gotr result get-case <run-id> <case-id>` — получить результаты для кейса
  - `gotr result add <test-id>` — добавить результат
  - `gotr result add-case <run-id>` — добавить результат для кейса
  - `gotr result add-bulk <run-id>` — массовое добавление из JSON-файла

#### CLI команды для Runs

- **Новый пакет** `cmd/run/` с командами:
  - `gotr run get <run-id>` — получить информацию о run
  - `gotr run list <project-id>` — список runs проекта
  - `gotr run create <project-id>` — создать run
  - `gotr run update <run-id>` — обновить run
  - `gotr run close <run-id>` — закрыть run
  - `gotr run delete <run-id>` — удалить run

#### Service Layer (Архитектурное улучшение)

- **Новый пакет** `internal/service/`:
  - `RunService` — бизнес-логика для runs с валидацией
  - `ResultService` — бизнес-логика для results с валидацией
  - `internal/service/migration/` — перенесён из `internal/migration/`
- **Валидация** в сервисах:
  - Проверка ID > 0
  - Проверка обязательных полей (name, suite_id, status_id)
  - Валидация bulk-запросов (непустые массивы)
- **Утилиты** в `internal/utils/helpers.go`:
  - `ParseID` — парсинг ID
  - `OutputResult` — вывод результата (JSON + сохранение в файл)
  - `PrintSuccess` — вывод сообщений
  - `SaveToFile` — сохранение данных в JSON-файл

#### Архитектурная документация

- **Системная документация** `.github/copilot/instructions/`:
  - Полное описание 4 слоёв архитектуры
  - Таблицы разделения ответственности
  - Полный перечень компонентов (22 команды, 3 сервиса, 40+ API методов)
  - Примеры рефакторинга
- **Пользовательская документация** `docs/architecture/overview.md` (243 строки):
  - Упрощённое описание архитектуры
  - Примеры потоков данных
  - Полный список команд

#### Тесты

- **Тесты для Service Layer**:
  - `internal/service/run_test.go` — 6 тестов для валидации RunService
  - `internal/service/result_test.go` — 9 тестов для валидации ResultService

---

</details>

## [2.3.0] - 2026-02-03

<details>
<summary>Details</summary>

### Added

#### Модели для Results и Runs API

- **Новые модели данных** в `internal/models/data/`:
  - `results.go` — модели `Result`, `AddResultRequest`, `AddResultsRequest`, `AddResultsForCasesRequest`
  - `runs.go` — модели `Run`, `AddRunRequest`, `UpdateRunRequest`, `CloseRunRequest`
  - `tests.go` — модели `Test`, `UpdateTestRequest`
  - `statuses.go` — модель `Status` с константами статусов
- Подготовка к реализации Results и Runs API

#### Исправления по результатам аудита

- **Обновлены request-структуры** в `cases.go`:
  - `AddCaseRequest` — добавлены поля `custom_steps` и `custom_expected` (текстовый формат)
  - `UpdateCaseRequest` — добавлены поля `type_id`, `suite_id`, `section_id`, `template_id` для перемещения кейсов
- **Исправлена модель `Section`** — добавлены `omitempty` к необязательным полям
- **Удалён дубликат метода** `AddCaseRequest` из `internal/client/cases.go`
- **Исправлен метод `GetSections`** — `suite_id` теперь передаётся как query-параметр

#### Системные изменения

- Системные файлы разработки вынесены в служебные инструкции `.github/copilot/instructions/`
- Внедрено осознанное версионирование (Semantic Versioning)

---

</details>

## [2.2.3] - 2026-02-03

<details>
<summary>Details</summary>

### Added

#### Интерактивный режим

- **Интерактивный выбор** для всех команд `get` и `sync`:
  - `gotr get cases` — интерактивный выбор проекта и сьюта
  - `gotr get suites` — интерактивный выбор проекта
  - `gotr get sharedsteps` — интерактивный выбор проекта
  - `gotr sync cases` — интерактивный выбор source/destination проектов и сьютов
  - `gotr sync shared-steps` — интерактивный выбор проектов
  - `gotr sync sections` — интерактивный выбор проектов и сьютов
  - `gotr sync full` — интерактивный выбор для полной миграции
- **Автоматический выбор**: если в проекте один сьют — используется автоматически
- **Флаг `--all-suites`** для `gotr get cases` — получение кейсов из всех сьютов проекта

#### Реструктуризация кода

- Новая структура пакета `cmd/`:
  - `cmd/get/` — отдельный пакет для GET-команд
  - `cmd/sync/` — отдельный пакет для SYNC-команд
  - `cmd/commands.go` — централизованная регистрация всех команд
  - `cmd/interactive.go` — общие функции интерактивного выбора
- Dependency injection для избежания циклических зависимостей между пакетами

#### Документация

- Создана директория `docs/` с подробной документацией:
  - `guides/installation.md` — установка
  - `guides/configuration.md` — конфигурация
  - `guides/commands/get.md` — команды получения данных
  - `guides/commands/sync.md` — команды синхронизации
  - `guides/interactive-mode.md` — интерактивный режим
  - `guides/commands/other.md` — другие команды

### Changed

- Улучшена работа с `suite-id` в `gotr get cases`:
  - Убрано жёсткое требование флага `--suite-id`
  - Интерактивный выбор при отсутствии флага
  - Понятное сообщение об ошибке от API при отсутствии suite_id для multiple suites
- Обновлены `Long` описания всех sync-команд с описанием интерактивного режима
- Регистрация флагов перенесена из `init()` функций в `cmd/sync/sync.go`

### Fixed

- Исправлено дублирование флагов в `sync` командах
- Убраны неиспользуемые переменные в тестах

---

</details>

## [2.1.0] - 2026-01-24

<details>
<summary>Details</summary>

### Added

- `gotr sync suites` — новая команда синхронизации suites: Fetch → Filter → Import.
- `gotr sync sections` — новая команда синхронизации sections.
- Общий хелпер `addSyncFlags()` для унификации флагов команд `sync/*`.
- Unit-тесты для `sync suites` и `sync sections`.

### Changed

- Команды `sync/*` переведены на единый поток миграции (internal/migration) и теперь используют централизованную логику Fetch → Filter → Import.
- Улучшены `Long` описания команд и добавлены русские комментарии-«Шаги» в коде команд для удобства русскоязычных пользователей.

### Testing

- В тестах используется отдельная папка логов: `.testrail/logs/test_runs`.
- Введён тестовый seam `sync_helpers.go` (переменная `newMigration`) для инъекции мок-миграций в тестах.

---

</details>

## [2.0.0] - 2026-01-15

<details>
<summary>Details</summary>

### Breaking Changes

- Полная переработка команды `get`: переход на подкоманды вместо универсального подхода.
  - Теперь `gotr get <resource>` с подкомандами: `cases`, `case`, `projects`, `project`, `sharedsteps`, `sharedstep`, `sharedstep-history`, `suites`, `suite`.
  - Убраны старые универсальные вызовы (например, `gotr get get_cases 30`).
- Все ID теперь строго типизированы как `int64` в методах клиента и структурах (было string в некоторых местах).
- `get_cases` теперь требует `suite_id` (обязательно для проектов в режиме multiple suites).
- Изменена структура ответов для некоторых эндпоинтов (например, `GetProjectsResponse`, `GetSharedStepsResponse` стали срезами вместо объектов с полем).

### Added

- Новые подкоманды в группе `get`:
  - `gotr get case <case-id>` — получить один кейс по ID кейса.
  - `gotr get case-history <case-id>` — получить историю изменений кейса.
  - `gotr get sharedstep <step-id>` — получить один shared step по ID шага.
  - `gotr get sharedstep-history <step-id>` — получить историю изменений shared step.
  - `gotr get suites` — получить список тест-сюит проекта.
  - `gotr get suite <suite-id>` — получить одну тест-сюиту по ID.
- Поддержка **позиционных аргументов** для ID проекта в `cases`, `sharedsteps`, `suites`.
- Явные и информативные подсказки в `Short` и `Long` для всех подкоманд.
- Проверка обязательных параметров в `RunE` с понятными сообщениями об ошибках.
- Методы клиента для suites: `GetSuites`, `GetSuite`, `AddSuite`, `UpdateSuite`, `DeleteSuite`.

### Changed

- Улучшена обработка ошибок в клиенте: проверка StatusCode перед декодированием, информативные сообщения.
- Все ответы на список (projects, cases, shared steps, suites) возвращают срез напрямую (массив), а не объект с полем.
- Убраны лишние обёртки в структурах ответов (GetProjectResponse → Project, GetCaseResponse → Case и т.д.).
- Подсказки в `help` теперь максимально понятные: указывают, какой ID нужен и где его взять.

### Fixed

- Исправлено декодирование массивов из API (projects, shared steps, cases).
- Исправлена проблема с `MarkFlagRequired` — теперь позиционные аргументы работают без конфликта с обязательными флагами.
- Исправлено поле `is_deleted` в Case (теперь int, так как API возвращает 0/1).

---

</details>

## [2.0.0] - 2025-12-21

<details>
<summary>Details</summary>

### Breaking Changes

- Изменён префикс переменных окружения с `GOTR_` на `TESTRAIL_` для лучшей совместимости с экосистемой TestRail (например, `TESTRAIL_BASE_URL`, `TESTRAIL_USERNAME`, `TESTRAIL_API_KEY`).
- Убраны старые ключи в конфиге и Viper (`testrail_base_url`, `testrail_username` и т.д.) — теперь используются `base_url`, `username`, `api_key`.

### Added

- Поддержка конфигурационного файла `~/.gotr/config/default.yaml` с автоматическим чтением (Viper).
- Новые подкоманды в группе `config`:
  - `gotr config init` — создание дефолтного конфига с комментариями.
  - `gotr config path` — показ пути к конфигу.
  - `gotr config view` — вывод содержимого конфига.
  - `gotr config edit` — открытие конфига в редакторе по умолчанию (`$EDITOR`).
- Автодополнение для bash (через `gotr completion bash`).
- Отключение обязательных проверок для служебных команд (`config`, `completion`).
- Условный вывод сообщений (без "Using config file" для чистоты stdout).
- Поддержка `insecure` в конфиге (для пропуска TLS-проверки).

### Changed

- Унифицированы ключи Viper: `base_url`, `username`, `api_key` (без `testrail_`).
- Улучшена обработка env-переменных с префиксом `TESTRAIL_`.

### Fixed

- Убрано дублирование сообщений "Using config file".
- Исправлено автодополнение (без мусора из файлов и вывода).

### Removed

- Старые env-переменные с префиксом `GOTR_`.

---

</details>

## [1.0.0] - 2025-12-19 (предыдущий релиз)

<details>
<summary>Details</summary>

- Базовая версия с командами `list`, `get`, `add` и т.д.
- Поддержка TestRail API v2 через HTTP-клиент.
- Глобальные флаги `--url`, `--username`, `--api-key`.

</details>
