# UX Modernization Design — gotr Interactive Commands

## Цель

Привести весь интерактив gotr к единой стилистике, реализованной в `cmd/snap/`:

- Многоуровневая навигация с Back/Exit
- Выровненные колонки с заголовками
- Группировка по категориям
- Карточки деталей + action-меню
- JSON-приоритет при сохранении (interactive → table, file save → JSON)

## Эталон: cmd/snap/

- 4 уровня: Server → Operation → Resource → Snapshot
- `alignedPickerLabels()` — авто-ширина колонок + header
- `errGoBack`/`errExit` sentinels — навигация Back/Exit
- `renderInfoCard()` — детальная карточка
- `postCardAction()` — action-меню после карточки
- `groupByOperation()`, `groupByCategory()` — вложенная группировка

## Текущее состояние по группам

### Tier 1 — Максимальный импакт (много данных, частое использование)

| Группа | Текущий паттерн | Что нужно |
|--------|----------------|-----------|
| cmd/cases/ | Project→Suite→Case, flat `[i] ID: X \| Title` | Aligned labels, Back/Exit, info card, grouping by section |
| cmd/get/ | Project→Suite→Item, flat | Aligned labels, Back/Exit, info card |
| cmd/compare/ | Complex flow | Aligned labels, navigation |
| cmd/sync/ | Select+Confirm chains | Back/Exit, aligned labels, preview cards, **snapshot integration** |

### Tier 2 — Средний импакт (умеренные списки)

| Группа | Текущий паттерн | Что нужно |
|--------|----------------|-----------|
| cmd/attachments/ | 3-level deep hierarchy | Back/Exit, aligned, cards |
| cmd/test/, cmd/tests/ | Project→Run→Test | Aligned labels, Back/Exit |
| cmd/run/ | Project→Suite→Run | Aligned labels, grouped by status |
| cmd/plans/ | Project→Plan→Entry | Aligned, cards |
| cmd/configurations/ | Group→Config | Aligned, Back/Exit |

### Tier 3 — Минимальный (короткие списки, редко)

| Группа | Текущий паттерн | Что нужно |
|--------|----------------|-----------|
| cmd/list/ | Flat resource select | Minimal — уже ок |
| cmd/delete/ | Endpoint→Item | Back/Exit, confirm card |
| cmd/add/, cmd/update/ | Input fields | Minimal changes |
| cmd/export/ | Resource→Endpoint→ID | Back/Exit |
| cmd/groups/ | Project→Group | Aligned |
| cmd/templates/ | Project select | Minimal |
| cmd/variables/, cmd/datasets/ | Dataset→Variable | Aligned |
| cmd/labels/, cmd/bdds/ | Flat select | Minimal |
| cmd/users/, cmd/roles/ | Flat select | Minimal |
| cmd/reports/ | Template select | Minimal |
| cmd/milestones/ | Project→Milestone | Aligned |
| cmd/result/ | Run→Test→Case | Back/Exit, aligned |

## Принципы реализации

### 1. Shared Navigation Kit (internal/interactive/)

Вынести из `cmd/snap/interactive_helpers.go` в `internal/interactive/`:

- `AlignedLabels(columns []Column, rows []Row) []string` — generic aligned formatter
- `BrowseLoop(cfg BrowseConfig) error` — generic browse with Back/Exit/action
- `GroupBy(entries, keyFn) []Group` — generic grouper
- Navigation sentinels: `ErrGoBack`, `ErrExit`
- `PostAction` enum + `ActionMenu(options)` pattern

### 2. Aligned Label Format

```
[1] ID: 123 │ cases    │ Regression Test Login │ active  │ 2026-04-13
[2] ID: 456 │ cases    │ Payment Flow          │ closed  │ 2026-04-12
```

- Авто-ширина колонок (как в snap)
- Header row опциональный
- `│` разделители

### 3. Info Cards

```
┌─ Case Info ──────────────────┐
│ ID         │ 12345           │
│ Title      │ Login Flow      │
│ Section    │ Auth / Login    │
│ Priority   │ Critical        │
│ Status     │ Active          │
└──────────────────────────────┘
```

- go-pretty таблица StyleRounded
- Отображается при выборе элемента

### 4. Action Menu (после карточки)

```
? Action:
  ← Back
  ✕ Exit
  ↻ Refresh
  📋 Copy ID
  💾 Save to file (JSON)
```

- Контекстные действия зависят от команды
- Rollback/Undo/Delete — только где применимо

### 5. JSON Output Priority

- **Interactive display** → table (human-readable)
- **--format json** → JSON to stdout
- **--save / file save** → JSON по умолчанию (не table!)
- **Pipe detection**: если stdout не TTY → JSON автоматически

### 6. Grouping Strategy

- Projects → группировка по Suite
- Cases → группировка по Section
- Runs → группировка по Status (active/completed)
- Tests → группировка по Run
- Configs → группировка по Config Group
- Attachments → группировка по Entity Type

## Sync — Snapshot Integration (ПРИОРИТЕТ)

### Контекст

Миграция (`gotr sync **`) — самая рискованная операция. Все инструменты безопасности
(snapshot, rollback, undo) должны быть доступны прямо из потока миграции, а не
заставлять пользователя переключаться между командами.

### 1. Обязательный Snapshot Prompt перед миграцией

Каждая подкоманда `gotr sync *` перед выполнением спрашивает:

```
? Create snapshot before migration? (recommended) [Y/n]
```

- Default: **true** (Yes) — снапшот создаётся всегда, если пользователь не откажется
- При `--non-interactive`: снапшот создаётся автоматически (без вопроса)
- При `--no-snapshot`: явно отключается (флаг)
- Снапшот сохраняется в manifest, ID выводится после создания:
  `✓ Snapshot created: sync/20260414T120000_sync_cases_0`

### 2. Post-Migration Navigation Menu

После завершения миграции (успех или частичный успех) — action-меню:

```
? Migration complete. What next?
  ← Back to sync menu
  ✕ Exit
  📋 View migration log
  ↻ Rollback this migration
  📦 Browse rollback history
```

- `↻ Rollback this migration` — вызывает `snap rollback <snapshot_id>` для только что созданного снапшота
- `📦 Browse rollback history` — открывает `snap rollback list` (browse rolled-back snapshots)
- `📋 View migration log` — показывает diff/summary выполненных изменений

### 3. Sync Subcommands Scope

Все подкоманды `gotr sync` получают snapshot integration:

- `gotr sync cases` — sync cases between suites
- `gotr sync suites` — sync suites between projects
- `gotr sync sections` — sync sections between suites
- `gotr sync shared-steps` — sync shared steps
- `gotr sync full` — full project sync (все вышеперечисленное)

### 4. Флаги

- `--snapshot` / `--no-snapshot` — явное управление снапшотом (default: true)
- `--rollback-on-error` — автоматический rollback при ошибке миграции (future)

## Порядок реализации

### Phase 0: Sync Snapshot Integration (приоритет)

1. Добавить `--snapshot` / `--no-snapshot` флаг ко всем sync подкомандам
2. Pre-migration snapshot prompt + auto-create
3. Post-migration action menu (rollback/browse/log)
4. Интеграция с существующим snap engine
5. Tests

### Phase 1: Shared Navigation Kit

6. Extract navigation primitives из cmd/snap/ → internal/interactive/
7. Generic `BrowseLoop`, `AlignedLabels`, `GroupBy`
8. Tests для kit

### Phase 2: Tier 1 Commands (cases, get, compare, sync)

9. cmd/get/ — aligned labels + Back/Exit + info cards
10. cmd/cases/ — grouped by section + aligned + cards
11. cmd/sync/ — navigation + preview cards (поверх Phase 0)
12. cmd/compare/ — aligned output

### Phase 3: Tier 2 Commands

13. cmd/attachments/, cmd/test/, cmd/run/
14. cmd/plans/, cmd/configurations/
15. cmd/result/

### Phase 4: Tier 3 Commands + Polish

16. Remaining commands — minimal aligned labels
17. Pipe detection → auto-JSON
18. Final consistency pass

## Ограничения

- НЕ ломать существующий non-interactive (`--non-interactive`) контракт
- НЕ менять CLI API (flags, args) — только interactive UX
- JSON schema не меняется — только presentation layer
- Backward-compatible: старые скрипты продолжают работать
