# Migration Guide v3.2 → v3.3

Language: Русский | [English](../../en/guides/migration-guide-v3.3.md)

## Навигация

- [Документация](../index.md)
  - [Гайды](index.md)
  - [Архитектура](../architecture/index.md)
  - [Эксплуатация](../operations/index.md)
  - [Отчёты](../reports/index.md)
- [Главная](../../../README_ru.md)

## TL;DR

```bash
gotr report organize --dry-run   && gotr report organize
gotr export organize --dry-run   && gotr export organize
# (опционально) политики retention/cleanup
gotr cleanup all --dry-run
```

Никакие существующие конфиги не ломаются; данные не удаляются
автоматически.

## Что изменилось

### 1. Иерархия `~/.gotr/reports/`

Было (v3.2):

```
~/.gotr/reports/
├── migration-20260418_...md
├── coverage_p34_...json
├── rollback_snap_rel_9946.json
├── testrail_get_plan_234_20260420T101530Z.json
└── INDEX.md
```

Стало (v3.3):

```
~/.gotr/reports/
├── migrations/rel_9946/2026-04/migration-20260418_...md
├── coverage/rel_9946/2026-04/coverage_p34_...json
├── rollbacks/rel_9946/2026-04/rollback_snap_rel_9946.json
├── testrail/p234/2026-04/testrail_get_plan_234_20260420T101530Z.json
├── _unclassified/2026-04/<misc>
└── INDEX.md
```

- Категоризация делается в `internal/report.ClassifyReport` по имени
  файла.
- Для `testrail_*_<YYYYMMDD>.json` без `T<HHMMSS>Z` — `YearMonth`
  остаётся пустым, файл попадает в `testrail/p<N>/<file>` (без
  month-subdir). Это **нормально**.

Что делать:

```bash
gotr report organize --dry-run
gotr report organize
```

- Команда **не удаляет** ничего.
- Конфликт (файл в target уже существует) → запись пропускается,
  счётчик `Skipped` увеличивается; исходник остаётся в корне.
- Повторный запуск — no-op.

### 2. Иерархия `~/.gotr/exports/`

Было (v3.2):

```
~/.gotr/exports/
├── snap_rel_9946_20260418.tar.gz
├── reports_20260418.zip
├── plans/
│   └── project_30_plans_20260401.json
└── reports/<legacy>
```

Стало (v3.3):

```
~/.gotr/exports/
├── snaps/snap_rel_9946_20260418.tar.gz
├── reports/reports_20260418.zip
└── api/plans/project_30_plans_20260401.json
```

- `snaps/` — tar.gz бандлы snap-ов,
- `reports/` — zip-бандлы и plain-файлы отчётов,
- `api/<resource>/` — legacy «сырые» выгрузки `gotr get ... --save`.

Миграция:

```bash
gotr export organize --dry-run
gotr export organize
```

### 3. TLS — от `insecure` к `ca_bundle`

Было:

```yaml
insecure: true
```

Стало (рекомендуемо):

```yaml
tls:
  insecure: false
  ca_bundle: "/etc/ssl/corp-ca.pem"
```

- Старый top-level `insecure` и флаг `--insecure` продолжают работать
  (OR-merge с `tls.insecure`).
- `tls.ca_bundle` подгружает PEM в `tls.Config.RootCAs` — безопаснее,
  чем `insecure=true`, и проходит корпоративные MITM-прокси.

### 4. Предупреждения — от «всё или ничего» к ключам

Было (не финальный API): `no_warnings: true`.

Стало:

```yaml
ui:
  suppress_warnings:
    - tls_insecure
    - flat_layout
```

- Ключи: `tls_insecure`, `deprecation`, `flat_layout`.
- `--show-warnings` — CLI-флаг для временного показа всех.
- При первой эмиссии любого варнинга в stderr выводится tip:
  «add '<key>' to ui.suppress_warnings to silence this warning».
- Флаг «показано про flat_layout» персистентный —
  `~/.gotr/state.json::flat_layout_warned`. Это означает, что подсказку
  вы увидите **ровно один раз** на инсталляцию.

### 5. Retention и cleanup

Retention по умолчанию **выключен**. Никакие старые артефакты не
удаляются автоматически после апгрейда.

Чтобы включить:

```yaml
retention:
  reports:
    enabled: true
    max_age_days: 90
    max_count: 500
    keep_categories: [coverage]
    dry_run: false
```

И применять вручную:

```bash
gotr cleanup reports --dry-run
gotr cleanup all
```

См. [gotr cleanup](commands/cleanup.md).

### 6. `gotr export snap --with-reports`

Новый флаг, **включён по умолчанию**. При экспорте snap-а в архив
автоматически добавляются связанные отчёты из `~/.gotr/reports/`
(матчинг по basename snap-а как substring в basename отчёта).

Отключить:

```bash
gotr export snap <snap-id> --no-reports
```

Встроенные отчёты видны в `manifest.Files` и распаковываются при
`gotr import snap` в ту же категоризованную иерархию.

## Чек-лист апгрейда

- [ ] `go install github.com/Korrnals/gotr@v3.3.0` (или ребилд из исходников).
- [ ] `gotr report organize --dry-run` → `gotr report organize`.
- [ ] `gotr export organize --dry-run` → `gotr export organize`.
- [ ] Перенёс `insecure: true` → `tls.ca_bundle: /etc/ssl/corp-ca.pem` (если было).
- [ ] (опционально) Добавил `ui.suppress_warnings: [tls_insecure]`,
  если баннер мешает в CI.
- [ ] (опционально) Настроил `retention.*` и прогнал
  `gotr cleanup all --dry-run`.

## Rollback

Если что-то пошло не так:

- Старый layout можно восстановить вручную: переместить файлы из
  подкаталогов обратно в корень `~/.gotr/reports/`. Никакой магии нет
  — `organize` работает на уровне ФС.
- `gotr config view` маскирует ключи, но сам YAML — plain; вернуть
  старый `insecure: true` можно в любой момент.
- Устанавливать предыдущую версию бинарника — безопасно: структура
  `~/.gotr/snaps/` и `~/.gotr/state.json` обратно-совместимы.

---

← [Гайды](index.md) · [Документация](../index.md)
