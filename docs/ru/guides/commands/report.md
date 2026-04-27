# Команда: report

Language: Русский | [English](../../../en/guides/commands/report.md)

## Навигация

- [Документация](../../index.md)
  - [Гайды](../index.md)
    - [Установка](../installation.md)
    - [Конфигурация](../configuration.md)
    - [Интерактивный режим](../interactive-mode.md)
    - [Прогресс](../progress.md)
    - [Каталог команд](index.md)
      - [Общие](global-flags.md)
      - [CRUD операции](add.md)
      - [Основные ресурсы](get.md)
      - [Специальные ресурсы](bdds.md)
      - [report](report.md)
      - [cleanup](cleanup.md)
    - [Инструкции](../instructions/index.md)
  - [Архитектура](../../architecture/index.md)
  - [Эксплуатация](../../operations/index.md)
  - [Отчёты](../../reports/index.md)
- [Главная](../../../../README_ru.md)

## Обзор 🎯

`gotr report` — управление локальными отчётами `gotr` в
`~/.gotr/reports/`. Начиная с v3.3.0 отчёты хранятся в категоризованной
иерархии:

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

Классификация — в `internal/report.ClassifyReport` по префиксу имени
файла. Для `testrail_*_<YYYYMMDD>T<HHMMSS>Z.json` извлекается и
`p<project>`, и месяц; для формата без `T<HHMMSS>Z` (только дата) файл
кладётся в `testrail/p<N>/<basename>` без month-subdir.

## Подкоманды

| Подкоманда | Описание |
| --- | --- |
| `list` | Рекурсивный листинг отчётов, с фильтром по glob basename или подстроке относительного пути |
| `show` | Показать содержимое отчёта: PDF через системный viewer, md/json/txt — cat в stdout или через `--print` |
| `view` | Печать содержимого отчёта в stdout с заголовком `# <path>`; работает для любых текстовых форматов, не вызывает OS-viewer |
| `organize` | Миграция «плоского» layout в иерархию (v3.2 → v3.3); идемпотентна |

## `gotr report list`

```bash
gotr report list [--filter <glob-or-substring>] [--limit <N>]
```

- Рекурсивный обход `~/.gotr/reports/`.
- `--filter` применяется по **basename** (glob) или как **substring**
  по относительному пути (`migrations/rel_9946/…`).
- `--limit <N>` ограничивает количество выводимых отчётов (по умолчанию: `20`).
- При обнаружении файлов в корне директории (плоский layout v3.2) один
  раз показывается подсказка `flat_layout` → рекомендация запустить
  `gotr report organize`. Подсказку можно подавить:
  `ui.suppress_warnings: [flat_layout]`.

## `gotr report show`

```bash
gotr report show <filename|latest> [--print]
```

- `<filename>` — basename, relative path или абсолютный путь.
  `latest` — самый свежий файл по mtime.
- PDF открывается через OS-launcher (`xdg-open`/`open`/`rundll32`);
  ненулевой exit лаунчера теперь пропагируется как ошибка CLI
  (exec.Cmd.Run, не Start — fix v3.3.0).
- `md`/`json`/`txt` по умолчанию cat-ятся в stdout.
- `--print` — принудительно печатать содержимое в stdout вне зависимости
  от расширения. Для PDF явная ошибка «cannot --print a binary PDF».
- Shell completion: tab-autocomplete по фактическому содержимому
  reports-иерархии.

## `gotr report view`

```bash
gotr report view <filename|latest>
```

- Печатает содержимое любого текстового отчёта (md/json/txt) в stdout,
  с заголовком `# <абсолютный-путь>`.
- В отличие от `show`, никогда не вызывает OS-viewer; удобно в скриптах
  и pipelines, когда нужен плоский вывод для любого формата.
- Без аргумента и на TTY открывает тот же survey-prompt, что
  `report show`.

## `gotr report organize`

```bash
gotr report organize [--dry-run]
```

- Один раз после апдейта на v3.3.0: переносит `*.md|*.json|*.pdf` из
  корня `~/.gotr/reports/` в подкаталоги по результатам
  `ClassifyReport`.
- `--dry-run` печатает план (`Plans`) без изменений.
- Никогда не перезаписывает существующие файлы: при коллизии путь
  добавляется в `Skipped`.
- Работает через rename; при `EXDEV` fallback-ит на copy+remove
  (cross-device FS).
- В конце вызывает `Reindex` — регенерирует `INDEX.md`.

### Пример

```bash
gotr report organize --dry-run
# → 48 moved, 0 skipped (dry-run)

gotr report organize
# → реально перемещает, затем пишет INDEX.md
```

## Интерактивный режим

Без аргумента и на TTY `report show` открывает survey-prompt со списком
кандидатов. В non-interactive окружении (pipe, `--non-interactive`)
требует явный argument и завершается с понятной ошибкой.

## См. также

- [Миграция v3.3.0](../migration-guide-v3.3.md)
- [Архитектура: UX polish v3.3.0](../../architecture/ux-polish-v3.3.0.md)
- [export](export.md) · [cleanup](cleanup.md)

---

← [Каталог команд](index.md) · [Документация](../../index.md)
