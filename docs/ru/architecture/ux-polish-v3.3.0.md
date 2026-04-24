# Архитектура: UX polish v3.3.0 (#44)

Language: Русский | [English](../../en/architecture/ux-polish-v3.3.0.md)

## Навигация

- [Документация](../index.md)
  - [Гайды](../guides/index.md)
  - [Архитектура](index.md)
    - [Обзор](overview.md)
    - [Concurrency](concurrency.md)
    - [Стандарты](standards.md)
    - [UX polish v3.3.0](ux-polish-v3.3.0.md)
  - [Эксплуатация](../operations/index.md)
  - [Отчёты](../reports/index.md)
- [Главная](../../../README_ru.md)

## Контекст

Релиз v3.3.0 закрывает issue #44 («UX polish after import/export bundles»).
Его цель — превратить «плоские» сторы `~/.gotr/reports/` и
`~/.gotr/exports/` в предсказуемую иерархию, добавить retention/cleanup,
управляемые предупреждения и корпоративный TLS, не ломая существующие
конфиги и данные.

Документ описывает компоненты, контракты и места, где изменилась или была
исправлена логика.

## 1. Категоризованная иерархия отчётов

### 1.1. Классификатор

`internal/report/paths.go`:

```go
type Classification struct {
    Category  string // migrations | coverage | rollbacks | no-snapshot | testrail | _unclassified
    Label     string // "default" либо normalized label / project-id
    YearMonth string // "2026-04" или "" (без month-subdir)
    Project   string // "p<N>" для testrail
}

func ClassifyReport(basename string) Classification
func ClassifyReportWithLabel(basename, label string) Classification
func (c Classification) RelDir() string
```

Правила:

| Паттерн имени файла | Category | Label | YearMonth |
| --- | --- | --- | --- |
| `migration-YYYYMMDD_HHMMSS_<label>.md` | `migrations` | `<label>` | `YYYY-MM` |
| `coverage_p<N>_<label>_<ts>.json` | `coverage` | `<label>` | `YYYY-MM` |
| `rollback_snap_<label>.json` | `rollbacks` | `<label>` | `YYYY-MM` |
| `no-snapshot-report_<ts>.md` | `no-snapshot` | `default` | `YYYY-MM` |
| `testrail_*_<pN>_<YYYYMMDD>T<HHMMSS>Z.json` | `testrail` | — | `YYYY-MM` |
| `testrail_*_<pN>_<YYYYMMDD>.json` (без `T...Z`) | `testrail` | — | `""` |
| всё остальное | `_unclassified` | `default` | `YYYY-MM` (если распознано) |

Relative directory:

```
<category>/<label|"">/<YYYY-MM|"">/<basename>
```

Пустые сегменты выжимаются, так что «testrail без месяца» даёт
`testrail/p234/<file>`.

### 1.2. Writers → категоризованный путь

Writers в `internal/report/*` (миграции, coverage, rollbacks) теперь
получают target через `ResolveReportPath` + `ClassifyReport`. Старый
путь (корень директории) остался только как fallback для уже созданных
артефактов.

### 1.3. Listing

`internal/report.RecursiveListReports(baseDir)` обходит дерево, возвращает
плоский список `[]string` относительных путей. `cmd/report/list` фильтрует
по glob basename ИЛИ по substring в relative path.

### 1.4. INDEX.md

`internal/report.Reindex(baseDir)` регенерирует `INDEX.md`. Вызывается в
трёх точках: после новой генерации отчёта, после import и в конце
`organize`.

### 1.5. Миграция плоского layout

`internal/report.MigrateFlatLayout(baseDir, dryRun) → *OrganizeResult{Plans, Moved, Skipped, DryRun}`:

- Сканирует корень, для каждого файла считает `target := ClassifyReport(name).RelDir()`.
- Никогда не перезаписывает target: коллизия → `Skipped`.
- Пытается `os.Rename`; при `EXDEV` fallback `copy + remove`
  (cross-device FS, контейнерные volume).
- В конце — `Reindex`.

Команда `gotr report organize` — тонкая обёртка.

### 1.6. Обнаружение плоского layout

`internal/report.IsFlatLayout(baseDir)` возвращает true, если в корне
есть хотя бы один файл с расширением отчёта. Используется в
`maybeFlatLayoutHint` (см. §4).

## 2. Категоризованные экспорты

### 2.1. Пути

`internal/paths`:

```go
func ExportsSnapsDirPath() string   // ~/.gotr/exports/snaps
func ExportsReportsDirPath() string // ~/.gotr/exports/reports
func ExportsAPIDirPath() string     // ~/.gotr/exports/api
// + EnsureExportsSnapsDir / EnsureExportsReportsDir / EnsureExportsAPIDir
```

### 2.2. Перенаправление writers

- `internal/snapbundle.DefaultExportPath` → `exports/snaps/`,
- `internal/reportbundle.ExportSingle / ExportAll` → `exports/reports/`,
- `internal/output.GetExportsDir(resource)` → `exports/api/<resource>/`.

### 2.3. Миграция

`internal/exportsorg.MigrateExportsLayout(base, dryRun) → *Result`:

```go
type Category int   // CategorySnap | CategoryReport | CategoryAPI
type Plan struct{ Src, Dst string; Category Category }
type Result struct{ Plans []Plan; Moved, Skipped int; DryRun bool }
```

Классификатор:

- `.tar.gz|.tgz` → snaps,
- `.zip|.pdf|.md|.json` → reports,
- директории-ресурсы (plans/, reports/, runs/…) → api/.

Команда — `gotr export organize [--dry-run]`.

## 3. Bundle + reports

`internal/snapbundle.ExportOptions`:

```go
type ExportOptions struct {
    GotrVersion      string
    Redact           bool
    IncludeReports   bool    // default: true
    ReportsDir       string  // обычно ~/.gotr/reports
}
```

`ExportOne` после сбора `snaps/<id>/` делает `collectReportEntries`:
рекурсивно обходит `ReportsDir`, добавляет в архив те файлы, где
`strings.Contains(filepath.Base(report), filepath.Base(snapID))`. Префикс
в архиве — `reports/<rel>`. Результат — `Result.IncludedReports`.

`Import` видит эти entries в `manifest.Files`, раскладывает их в
`~/.gotr/reports/<original-rel>`, что автоматически укладывается в новую
иерархию (т.к. оригинал уже был категоризован).

CLI: флаг `--with-reports` default ON на `gotr export snap`; `--no-reports`
— opt-out.

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

- При `showAll == true` все варнинги идут, независимо от `suppress`.
- При первом выводе для каждого ключа добавляется tip:
  `(add '<k>' to ui.suppress_warnings to silence this warning)`.
- In-memory map `shownHint` предотвращает повторный вывод одного и того
  же варнинга в рамках процесса.

### 4.2. `internal/state`

```go
type State struct {
    FlatLayoutWarned bool `json:"flat_layout_warned"`
}
func Path() string                          // ~/.gotr/state.json
func Load() (*State, error)
func (*State) Save() error                  // atomic rename, 0o600
```

Используется для one-time подсказок, которые должны пережить рестарты
процесса. v3.3.0 использует только `FlatLayoutWarned`; схема открыта
для будущих флагов.

### 4.3. Точки эмиссии в `cmd/root`

`PersistentPreRunE`:

1. `warnings.Init(viper.GetStringSlice("ui.suppress_warnings"), viper.GetBool("show_warnings"))`.
2. `insecure := viper.GetBool("insecure") || viper.GetBool("tls.insecure")`.
3. Если `insecure` — `warnings.Emitf(os.Stderr, KeyTLSInsecure, "WARNING: TLS verification disabled")`.
4. `caBundle := viper.GetString("tls.ca_bundle")` → `client.WithCABundle(caBundle)` если непустой.

Старая прямая `fmt.Fprintln(os.Stderr, "WARNING: TLS…")` из
`internal/client/client.go` **удалена** — это ключевой архитектурный
сдвиг: все пользовательские предупреждения теперь идут через единый
registry и уважают `suppress_warnings`.

## 5. TLS: ca_bundle

`internal/client/client.go`:

```go
func WithCABundle(path string) ClientOption
func loadCAPool(path string) (*x509.CertPool, error)
```

- Путь пустой → опция no-op.
- Файл читается `os.ReadFile`, парсится `x509.NewCertPool().AppendCertsFromPEM`.
- Ошибка парсинга → возвращается при первом HTTP-вызове (как и было для
  других TLS-опций).
- Применяется через `http.Transport.TLSClientConfig.RootCAs`.

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

Алгоритм для `reports`:

1. Рекурсивно собрать файлы и классифицировать по `Category`.
2. Сгруппировать по `Category`. Whitelist (`KeepCategories`) → skip.
3. Внутри категории — отсортировать по mtime DESC.
4. Пометить к удалению: файлы старше `MaxAgeDays` И/ИЛИ выходящие за
   `MaxCount` самых свежих.
5. Если `DryRun` — только план; иначе `os.Remove`.
6. `Reindex(base)`.

Для `exports` — без категорий, но с whitelist по именам подкаталогов
(`snaps/reports/api/` не удаляются как директории).

### 6.2. CLI

`cmd/cleanup/cleanup.go`: `gotr cleanup {reports,exports,snaps,all} [--dry-run]`.
`snaps` проксирует существующий `gotr snap gc` (backwards-compat), передавая
политику из `retention.snaps`.

## 7. UX: completion + interactive

- **Completion**: `cmd/report/completion.go`, `cmd/bundlecmd/completion.go`
  — `ValidArgsFunction` возвращает реальные имена из
  `RecursiveListReports` / `snap.LoadManifest` / листинга
  `~/.gotr/exports/*`. Учитывает двойное расширение `.tar.gz`.
- **Interactive**: `internal/interactive.Prompter`,
  `TerminalPrompter`, `NonInteractivePrompter`. Команды `show/view/export
  report/snap, import snap/report` принимают `cobra.MaximumNArgs(1)`.
  При отсутствии аргумента на TTY и без `--non-interactive` — survey
  prompt; иначе — явная ошибка «pass as argument or run interactively».
- TTY-guard: `golang.org/x/term.IsTerminal(int(os.Stdin.Fd()))`.

## 8. Изменения логики и баг-фиксы, зафиксированные в v3.3.0

| Файл / функция | Было | Стало | Почему |
| --- | --- | --- | --- |
| `internal/client/client.go` TLS-баннер | `fmt.Fprintln(os.Stderr, "WARNING: TLS...")` | удалён; эмиссия перенесена в `cmd/root` через `warnings.Emitf` | чтобы `ui.suppress_warnings` работал |
| `cmd/report/show.go::openWithOS` | `exec.Cmd.Start()` (fire-and-forget) | `exec.Cmd.Run()` — ждём exit | ненулевой exit viewer'а теперь становится ошибкой CLI |
| `cmd/report/list.go::maybeFlatLayoutHint` | прямой `fmt.Fprintln` | `warnings.Emitf(…, KeyFlatLayout, …)` + persistent gate `state.FlatLayoutWarned` | one-time, подавляемо |
| `cmd/root.go::PersistentPreRunE` (insecure) | только `insecure` top-level | OR-merge `insecure || tls.insecure` | обратная совместимость + новый ключ |
| `internal/report.MigrateFlatLayout` | — (новое) | never-overwrite; `rename → copy+remove` fallback | корректная работа на cross-device FS |
| `internal/snapbundle.ExportOne` | без reports | опция `IncludeReports` (default ON), matching по `filepath.Base(snapID)` | celf-contained bundles |
| `cmd/report/show.go::--print` | — (новое) | cat в stdout независимо от расширения; PDF → явная ошибка | scripting / CI |
| `cmd/report/completion.go`, `cmd/bundlecmd/completion.go` | статический listing | `ValidArgsFunction` с рекурсивным листингом | совпадение с реальной иерархией |
| `cmd/report/list.go::--filter` | только glob basename | glob basename ИЛИ substring в rel path | вложенные категории |

## 9. E2E-покрытие

- `internal/report/e2e_lifecycle_test.go`:
  - `TestE2E_FlatToHierarchy_RoundTrip` — 6 категорий (вкл. testrail без
    месяца), идемпотентный второй `MigrateFlatLayout`, `INDEX.md` после
    `Reindex`.
  - `TestE2E_EmptyReportsDir_IsNotFlat`.
- `internal/snapbundle/e2e_reports_test.go`:
  - `TestE2E_OrganizeThenExportImport_WithReports` — organize →
    `ExportOne{IncludeReports:true}` → `Import` → отчёты появляются в
    категоризованной иерархии.

## 10. Совместимость

- Конфиги v3.2 работают без изменений.
- `insecure: true` / `--insecure` — поддерживаются.
- Старый layout reports/exports — читается и показывается подсказка о
  миграции, **автоматически** не реорганизуется.
- Snap-архивы от v3.2 импортируются без изменений; reports-entries
  просто отсутствуют.

---

← [Архитектура](index.md) · [Документация](../index.md)
