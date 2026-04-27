# Command: report

Language: [Русский](../../../ru/guides/commands/report.md) | English

## Navigation

- [Documentation](../../index.md)
  - [Guides](../index.md)
    - [Installation](../installation.md)
    - [Configuration](../configuration.md)
    - [Interactive Mode](../interactive-mode.md)
    - [Progress](../progress.md)
    - [Commands Index](index.md)
      - [General](global-flags.md)
      - [CRUD Operations](add.md)
      - [Core Resources](get.md)
      - [Special Resources](bdds.md)
      - [report](report.md)
      - [cleanup](cleanup.md)
    - [Instructions](../instructions/index.md)
  - [Architecture](../../architecture/index.md)
  - [Operations](../../operations/index.md)
  - [Reports](../../reports/index.md)
- [Home](../../../../README.md)

## Overview 🎯

`gotr report` manages local `gotr` reports stored in
`~/.gotr/reports/`. Starting with v3.3.0 reports are stored in a
categorized hierarchy:

```
~/.gotr/reports/
├── migrations/<label>/<YYYY-MM>/…
├── coverage/<label>/<YYYY-MM>/…
├── rollbacks/<label>/<YYYY-MM>/…
├── no-snapshot/<label>/<YYYY-MM>/…
├── testrail/p<N>/<YYYY-MM>/…
├── _unclassified/<YYYY-MM>/…
└── INDEX.md
```

Classification is performed in `internal/report.ClassifyReport` based on
the file name prefix. For `testrail_*_<YYYYMMDD>T<HHMMSS>Z.json` both
`p<project>` and the month are extracted; for the date-only format
(without `T<HHMMSS>Z`) the file is placed under
`testrail/p<N>/<basename>` with no month subdir.

## Subcommands

| Subcommand | Description |
| --- | --- |
| `list` | Recursive listing of reports, with a filter by basename glob or relative-path substring |
| `show` | Show report contents: PDF via the system viewer, md/json/txt — cat to stdout or via `--print` |
| `view` | Print report contents to stdout with a `# <path>` header; works for any text format, does not invoke the OS viewer |
| `organize` | Migrate the "flat" layout into the hierarchy (v3.2 → v3.3); idempotent |

## `gotr report list`

```bash
gotr report list [--filter <glob-or-substring>] [--limit <N>]
```

- Recursive walk of `~/.gotr/reports/`.
- `--filter` is applied either as a **basename** glob or as a **substring**
  match against the relative path (`migrations/rel_9946/…`).
- `--limit <N>` caps the number of reports shown (default: `20`).
- When files are detected at the directory root (the v3.2 flat layout), a
  `flat_layout` hint is shown once recommending `gotr report organize`.
  The hint can be suppressed via
  `ui.suppress_warnings: [flat_layout]`.

## `gotr report show`

```bash
gotr report show <filename|latest> [--print]
```

- `<filename>` accepts a basename, relative path, or absolute path.
  `latest` is the newest file by mtime.
- PDF is opened through the OS launcher (`xdg-open`/`open`/`rundll32`);
  a non-zero exit from the launcher is now propagated as a CLI error
  (exec.Cmd.Run, not Start — fix in v3.3.0).
- `md`/`json`/`txt` are cat'd to stdout by default.
- `--print` forces printing the contents to stdout regardless of the
  extension. For PDF an explicit error is returned: "cannot --print a
  binary PDF".
- Shell completion: tab-autocomplete works against the actual contents
  of the reports hierarchy.

## `gotr report view`

```bash
gotr report view <filename|latest>
```

- Prints the contents of any text report (md/json/txt) to stdout,
  prefixed with a `# <absolute-path>` header.
- Unlike `show`, never invokes the OS viewer; useful in scripts and
  pipelines where you want plain output for any file type.
- Without an argument and on a TTY, opens the same survey prompt as
  `report show`.

## `gotr report organize`

```bash
gotr report organize [--dry-run]
```

- Run once after upgrading to v3.3.0: moves `*.md|*.json|*.pdf` from the
  root of `~/.gotr/reports/` into subdirectories according to
  `ClassifyReport`.
- `--dry-run` prints the plan (`Plans`) without changes.
- Never overwrites existing files: on collision the path is added to
  `Skipped`.
- Operates via rename; on `EXDEV` falls back to copy+remove (cross-device
  FS).
- At the end calls `Reindex` — regenerates `INDEX.md`.

### Example

```bash
gotr report organize --dry-run
# → 48 moved, 0 skipped (dry-run)

gotr report organize
# → actually moves files, then writes INDEX.md
```

## Interactive mode

Without an argument and on a TTY, `report show` opens a survey prompt
listing the candidates. In a non-interactive environment (pipe,
`--non-interactive`) it requires an explicit argument and exits with a
clear error.

## See also

- [Migration v3.3.0](../migration-guide-v3.3.md)
- [Architecture: UX polish v3.3.0](../../architecture/ux-polish-v3.3.0.md)
- [export](export.md) · [cleanup](cleanup.md)

---

← [Commands Index](index.md) · [Documentation](../../index.md)
