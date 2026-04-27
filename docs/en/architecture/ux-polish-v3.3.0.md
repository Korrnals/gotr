# Architecture: UX polish v3.3.0 (#44)

Language: [Русский](../../ru/architecture/ux-polish-v3.3.0.md) | English

## Navigation

- [Documentation](../index.md)
  - [Guides](../guides/index.md)
  - [Architecture](index.md)
    - [Overview](overview.md)
    - [Concurrency](concurrency.md)
    - [Standards](standards.md)
    - [UX polish v3.3.0](ux-polish-v3.3.0.md)
  - [Operations](../operations/index.md)
  - [Reports](../reports/index.md)
- [Home](../../../README.md)

## Context

Release v3.3.0 closes issue #44 ("UX polish after import/export bundles").
Its goal is to turn the "flat" `~/.gotr/reports/` and `~/.gotr/exports/`
stores into a predictable hierarchy, to add retention/cleanup,
manageable warnings, and corporate TLS, without breaking existing
configs and data.

This document describes the components, contracts, and the places where
logic was changed or fixed.

## 1. Categorised report hierarchy

### 1.1. Classifier

`internal/report/paths.go`:

```go
type Classification struct {
    Category  string // migrations | coverage | rollbacks | no-snapshot | testrail | _unclassified
    Label     string // "default" or normalized label / project-id
    YearMonth string // "2026-04" or "" (no month-subdir)
    Project   string // "p<N>" for testrail
}

func ClassifyReport(basename string) Classification
func ClassifyReportWithLabel(basename, label string) Classification
func (c Classification) RelDir() string
```

Rules:

| File-name pattern | Category | Label | YearMonth |
| --- | --- | --- | --- |
| `migration-YYYYMMDD_HHMMSS_<label>.md` | `migrations` | `<label>` | `YYYY-MM` |
| `coverage_p<N>_<label>_<ts>.json` | `coverage` | `<label>` | `YYYY-MM` |
| `rollback_snap_<label>.json` | `rollbacks` | `<label>` | `YYYY-MM` |
| `no-snapshot-report_<ts>.md` | `no-snapshot` | `default` | `YYYY-MM` |
| `testrail_*_<pN>_<YYYYMMDD>T<HHMMSS>Z.json` | `testrail` | — | `YYYY-MM` |
| `testrail_*_<pN>_<YYYYMMDD>.json` (no `T...Z`) | `testrail` | — | `""` |
| everything else | `_unclassified` | `default` | `YYYY-MM` (if recognised) |

Relative directory:

```
<category>/<label|"">/<YYYY-MM|"">/<basename>
```

Empty segments are squeezed out, so "testrail without month" yields
`testrail/p234/<file>`.

### 1.2. Writers → categorised path

Writers in `internal/report/*` (migrations, coverage, rollbacks) now
obtain the target via `ResolveReportPath` + `ClassifyReport`. The old
path (the directory root) remains only as a fallback for already
existing artefacts.

### 1.3. Listing

`internal/report.RecursiveListReports(baseDir)` walks the tree and
returns a flat `[]string` of relative paths. `cmd/report/list` filters
by glob basename OR by substring in the relative path.

### 1.4. INDEX.md

`internal/report.Reindex(baseDir)` regenerates `INDEX.md`. It is called
at three points: after a new report is generated, after import, and at
the end of `organize`.

### 1.5. Migrating the flat layout

`internal/report.MigrateFlatLayout(baseDir, dryRun) → *OrganizeResult{Plans, Moved, Skipped, DryRun}`:

- Scans the root, computes `target := ClassifyReport(name).RelDir()` for each file.
- Never overwrites the target: a collision → `Skipped`.
- Tries `os.Rename`; on `EXDEV` falls back to `copy + remove`
  (cross-device FS, container volumes).
- Finishes with `Reindex`.

The `gotr report organize` command is a thin wrapper.

### 1.6. Detecting the flat layout

`internal/report.IsFlatLayout(baseDir)` returns true if at least one
file with a report extension exists in the root. Used by
`maybeFlatLayoutHint` (see §4).

## 2. Categorised exports

### 2.1. Paths

`internal/paths`:

```go
func ExportsSnapsDirPath() string   // ~/.gotr/exports/snaps
func ExportsReportsDirPath() string // ~/.gotr/exports/reports
func ExportsAPIDirPath() string     // ~/.gotr/exports/api
// + EnsureExportsSnapsDir / EnsureExportsReportsDir / EnsureExportsAPIDir
```

### 2.2. Redirecting writers

- `internal/snapbundle.DefaultExportPath` → `exports/snaps/`,
- `internal/reportbundle.ExportSingle / ExportAll` → `exports/reports/`,
- `internal/output.GetExportsDir(resource)` → `exports/api/<resource>/`.

### 2.3. Migration

`internal/exportsorg.MigrateExportsLayout(base, dryRun) → *Result`:

```go
type Category int   // CategorySnap | CategoryReport | CategoryAPI
type Plan struct{ Src, Dst string; Category Category }
type Result struct{ Plans []Plan; Moved, Skipped int; DryRun bool }
```

Classifier:

- `.tar.gz|.tgz` → snaps,
- `.zip|.pdf|.md|.json` → reports,
- resource directories (plans/, reports/, runs/…) → api/.

The command is `gotr export organize [--dry-run]`.

## 3. Bundle + reports

`internal/snapbundle.ExportOptions`:

```go
type ExportOptions struct {
    GotrVersion      string
    Redact           bool
    IncludeReports   bool    // default: true
    ReportsDir       string  // typically ~/.gotr/reports
}
```

After collecting `snaps/<id>/`, `ExportOne` runs `collectReportEntries`:
it walks `ReportsDir` recursively and adds those files into the archive
where `strings.Contains(filepath.Base(report), filepath.Base(snapID))`.
The archive prefix is `reports/<rel>`. The output is `Result.IncludedReports`.

`Import` sees these entries in `manifest.Files` and lays them out in
`~/.gotr/reports/<original-rel>`, which automatically falls into the
new hierarchy (since the original was already categorised).

CLI: the `--with-reports` flag defaults to ON for `gotr export snap`;
`--no-reports` is the opt-out.

## 4. Warnings registry + persistent state

### 4.1. `internal/warnings`

```go
type Key string
const (
    KeyTLSInsecure = "tls_insecure"
    KeyDeprecation = "deprecation"
    KeyFlatLayout  = "flat_layout"
)

func Init(suppress []string, showAll bool)
func Suppressed(k Key) bool
func Emit(w io.Writer, k Key, msg string)
func Emitf(w io.Writer, k Key, format string, args ...any)
```

- When `showAll == true`, all warnings are emitted regardless of `suppress`.
- The first time a key is emitted, a tip is appended:
  `(add '<k>' to ui.suppress_warnings to silence this warning)`.
- An in-memory `shownHint` map prevents the same warning from being
  printed twice within a single process.

### 4.2. `internal/state`

```go
type State struct {
    FlatLayoutWarned bool `json:"flat_layout_warned"`
}
func Path() string                          // ~/.gotr/state.json
func Load() (*State, error)
func (*State) Save() error                  // atomic rename, 0o600
```

Used for one-time hints that must survive process restarts. v3.3.0
uses only `FlatLayoutWarned`; the schema is open to future flags.

### 4.3. Emission points in `cmd/root`

`PersistentPreRunE`:

1. `warnings.Init(viper.GetStringSlice("ui.suppress_warnings"), viper.GetBool("show_warnings"))`.
2. `insecure := viper.GetBool("insecure") || viper.GetBool("tls.insecure")`.
3. If `insecure` — `warnings.Emitf(os.Stderr, KeyTLSInsecure, "WARNING: TLS verification disabled")`.
4. `caBundle := viper.GetString("tls.ca_bundle")` → `client.WithCABundle(caBundle)` if non-empty.

The old direct `fmt.Fprintln(os.Stderr, "WARNING: TLS…")` from
`internal/client/client.go` has been **removed** — this is the key
architectural shift: every user-facing warning now flows through a
single registry and respects `suppress_warnings`.

## 5. TLS: ca_bundle

`internal/client/client.go`:

```go
func WithCABundle(path string) ClientOption
func loadCAPool(path string) (*x509.CertPool, error)
```

- An empty path → the option is a no-op.
- The file is read with `os.ReadFile`, parsed with `x509.NewCertPool().AppendCertsFromPEM`.
- A parse error → returned on the first HTTP call (same behaviour as
  other TLS options).
- Applied via `http.Transport.TLSClientConfig.RootCAs`.

## 6. Retention / cleanup

### 6.1. Policy

`internal/retention`:

```go
type Policy struct {
    Enabled        bool
    MaxAgeDays     int
    MaxCount       int
    KeepCategories []string
    DryRun         bool
}

func CleanupReports(base string, p Policy) (*Result, error)
func CleanupExports(base string, p Policy) (*Result, error)
```

Algorithm for `reports`:

1. Recursively gather files and classify them by `Category`.
2. Group by `Category`. Whitelist (`KeepCategories`) → skip.
3. Within a category, sort by mtime DESC.
4. Mark for removal: files older than `MaxAgeDays` AND/OR those past
   the `MaxCount` most recent.
5. If `DryRun` — only the plan; otherwise `os.Remove`.
6. `Reindex(base)`.

For `exports` — no categories, but with a whitelist over subdirectory
names (`snaps/reports/api/` are not removed as directories).

### 6.2. CLI

`cmd/cleanup/cleanup.go`: `gotr cleanup {reports,exports,snaps,all} [--dry-run]`.
`snaps` proxies the existing `gotr snap gc` (backwards-compat), passing
the policy from `retention.snaps`.

## 7. UX: completion + interactive

- **Completion**: `cmd/report/completion.go`, `cmd/bundlecmd/completion.go`
  — `ValidArgsFunction` returns real names from
  `RecursiveListReports` / `snap.LoadManifest` / listings of
  `~/.gotr/exports/*`. The double extension `.tar.gz` is honoured.
- **Interactive**: `internal/interactive.Prompter`,
  `TerminalPrompter`, `NonInteractivePrompter`. The `show/view/export
  report/snap, import snap/report` commands accept `cobra.MaximumNArgs(1)`.
  When the argument is missing on a TTY and `--non-interactive` is not
  set — a survey prompt is shown; otherwise an explicit error
  "pass as argument or run interactively".
- TTY guard: `golang.org/x/term.IsTerminal(int(os.Stdin.Fd()))`.

## 8. Logic changes and bug fixes recorded in v3.3.0

| File / function | Was | Became | Why |
| --- | --- | --- | --- |
| `internal/client/client.go` TLS banner | `fmt.Fprintln(os.Stderr, "WARNING: TLS...")` | removed; emission moved to `cmd/root` via `warnings.Emitf` | so that `ui.suppress_warnings` works |
| `cmd/report/show.go::openWithOS` | `exec.Cmd.Start()` (fire-and-forget) | `exec.Cmd.Run()` — wait for exit | a non-zero viewer exit now becomes a CLI error |
| `cmd/report/list.go::maybeFlatLayoutHint` | direct `fmt.Fprintln` | `warnings.Emitf(…, KeyFlatLayout, …)` + persistent gate `state.FlatLayoutWarned` | one-time, suppressible |
| `cmd/root.go::PersistentPreRunE` (insecure) | only top-level `insecure` | OR-merge `insecure || tls.insecure` | backwards compatibility + new key |
| `internal/report.MigrateFlatLayout` | — (new) | never-overwrite; `rename → copy+remove` fallback | correct behaviour on cross-device FS |
| `internal/snapbundle.ExportOne` | no reports | `IncludeReports` option (default ON), matching by `filepath.Base(snapID)` | self-contained bundles |
| `cmd/report/show.go::--print` | — (new) | cat to stdout regardless of extension; PDF → explicit error | scripting / CI |
| `cmd/report/completion.go`, `cmd/bundlecmd/completion.go` | static listing | `ValidArgsFunction` with recursive listing | matches the actual hierarchy |
| `cmd/report/list.go::--filter` | only glob basename | glob basename OR substring in rel path | nested categories |

## 9. E2E coverage

- `internal/report/e2e_lifecycle_test.go`:
  - `TestE2E_FlatToHierarchy_RoundTrip` — 6 categories (including
    testrail without month), idempotent second `MigrateFlatLayout`,
    `INDEX.md` after `Reindex`.
  - `TestE2E_EmptyReportsDir_IsNotFlat`.
- `internal/snapbundle/e2e_reports_test.go`:
  - `TestE2E_OrganizeThenExportImport_WithReports` — organize →
    `ExportOne{IncludeReports:true}` → `Import` → reports show up in
    the categorised hierarchy.

## 10. Compatibility

- v3.2 configs work unchanged.
- `insecure: true` / `--insecure` are supported.
- The old reports/exports layout is still read and a migration hint is
  shown; **no automatic** reorganisation happens.
- v3.2 snap archives are imported without changes; reports entries
  are simply absent.

---

← [Architecture](index.md) · [Documentation](../index.md)
