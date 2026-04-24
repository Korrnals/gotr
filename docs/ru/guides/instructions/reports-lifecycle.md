# Инструкция: жизненный цикл отчётов в v3.3

Language: Русский | [English](../../../en/guides/instructions/reports-lifecycle.md)

## Навигация

- [Документация](../../index.md)
  - [Гайды](../index.md)
    - [Инструкции](index.md)
      - [Жизненный цикл отчётов](reports-lifecycle.md)
  - [Архитектура](../../architecture/index.md)
  - [Эксплуатация](../../operations/index.md)
  - [Отчёты](../../reports/index.md)
- [Главная](../../../../README_ru.md)

Рецепт покрывает все стадии работы с локальными отчётами в v3.3.0:
миграция → генерация → листинг → просмотр → экспорт со встроенными
отчётами → импорт в чистое окружение → очистка по политике.

## Предусловия

- `gotr` v3.3.0+ (`gotr --version`).
- Конфиг в `~/.gotr/config/default.yaml` с корректным `base_url`, `username`, `api_key`.
- Хотя бы один завершённый snap/migration прогон (чтобы было что организовать).

## Кейс A: Миграция layout с v3.2 на v3.3 (однократно)

**Сценарий:** обновились с v3.2, в `~/.gotr/reports/` — плоский список файлов.

```bash
# 1. Посмотреть план миграции без изменений
gotr report organize --dry-run
# → напечатает список пар src → dst

# 2. Применить
gotr report organize
# → Moved: N, Skipped: 0
# INDEX.md регенерируется автоматически

# 3. Проверить результат
tree -L 3 ~/.gotr/reports | head -30
gotr report list
```

**Откат.** Если новая иерархия не нужна — верните файлы в корень вручную
(`find ~/.gotr/reports -type f -exec mv -t ~/.gotr/reports {} +`) и
удалите пустые подкаталоги. Ничего не шифруется.

## Кейс B: Первая сессия на v3.3 (плоский layout + подсказка)

При первом `gotr report list` появится **однократная** подсказка про
flat layout. Флаг сохраняется в `~/.gotr/state.json`. Способы убрать:

1. Выполнить миграцию (кейс A) — подсказка перестанет появляться, т.к.
   плоского layout больше нет.
2. Подавить в конфиге, не мигрируя:
   ```yaml
   ui:
     suppress_warnings: [flat_layout]
   ```

## Кейс C: Просмотр и печать отчёта

```bash
# Интерактивный выбор из рекурсивной иерархии (на TTY)
gotr report show

# По latest — самый свежий файл по mtime
gotr report show latest

# По basename — matcher найдёт файл в любой категории
gotr report show migration-20260418_123456_rel_9946.md

# По relative path — для точного выбора
gotr report show migrations/rel_9946/2026-04/migration-20260418_123456_rel_9946.md

# Печать содержимого в stdout (для CI/pipeline)
gotr report show latest --print > /tmp/report.md
gotr report show coverage_p34_... --print | jq '.summary'

# PDF открывается через OS-viewer; exit-code пропагируется
gotr report show full-audit-2026-04-18.pdf
echo "viewer exit: $?"
```

**Что нельзя:** `gotr report show <file.pdf> --print` → явная ошибка про
binary.

## Кейс D: Экспорт snap вместе с отчётами

**Сценарий:** передать snap коллеге со всеми связанными артефактами
(migration report, coverage, rollback-blueprint).

```bash
# По умолчанию --with-reports=true
gotr export snap rel_9946

# Архив попадает в exports/snaps/
ls -lh ~/.gotr/exports/snaps/

# Посмотреть, что внутри
tar -tzf ~/.gotr/exports/snaps/snap_rel_9946_*.tar.gz | head -20
# → manifest.json, SHA256SUMS, snaps/rel_9946/*, reports/migrations/..., reports/coverage/...

# Выключить встраивание отчётов
gotr export snap rel_9946 --no-reports
```

Критерий matching'а: `filepath.Base(report)` содержит
`filepath.Base(snapID)`. Пример: snapID `rel_9946` → подтянет
`migration-*_rel_9946.md`, `coverage_p34_rel_9946.json`, но не
`rollback_snap_rel_other.json`.

## Кейс E: Round-trip через import в другое окружение

```bash
# На dst-машине с чистым ~/.gotr/
gotr import snap /path/to/snap_rel_9946_20260420.tar.gz

# Snap ляжет в ~/.gotr/snaps/rel_9946/
# Отчёты автоматически раскладываются в категоризованную иерархию:
gotr report list | grep rel_9946
# → migrations/rel_9946/2026-04/migration-*.md
# → coverage/rel_9946/2026-04/coverage_*.json
```

Схема архива (`reports/<rel>`) совпадает с целевой иерархией, поэтому
дополнительного `organize` после `import` не требуется.

## Кейс F: Периодическая очистка

```yaml
# ~/.gotr/config/default.yaml
retention:
  reports:
    enabled: true
    max_age_days: 90
    max_count: 500
    keep_categories: [coverage]
```

```bash
# Первый раз — всегда dry-run
gotr cleanup reports --dry-run

# Применить
gotr cleanup reports

# Или одной командой — reports + exports + snaps
gotr cleanup all
```

Категория `coverage` в whitelist: никогда не удаляется retention-ом.
Файл `INDEX.md` регенерируется после удаления.

## Критерии успеха

- `gotr report list` показывает иерархические пути, не плоские имена.
- В `~/.gotr/exports/` есть подкаталоги `snaps/`, `reports/`, `api/`.
- `tar -tzf <snap.tar.gz>` содержит и `snaps/<id>/`, и `reports/<rel>`.
- `~/.gotr/state.json::flat_layout_warned=true` (если видели подсказку).
- После `gotr cleanup reports` `INDEX.md` обновлён, счётчик удалений > 0.

## См. также

- [gotr report](../commands/report.md) — полная справка по команде.
- [gotr cleanup](../commands/cleanup.md) — retention executor.
- [Migration guide v3.3](../migration-guide-v3.3.md).
- [Архитектура: UX polish v3.3.0](../../architecture/ux-polish-v3.3.0.md).

---

← [Инструкции](index.md) · [Документация](../../index.md)
