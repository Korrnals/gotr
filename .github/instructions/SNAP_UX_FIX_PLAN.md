# План исправлений snap/rollback UX — 2026-04-14

Language: Русский

---

## Контекст

По результатам аудита и ручного тестирования выявлено 7 UX-проблем.
Решения согласованы с пользователем (см. аудит: `docs/ru/reports/snap-audit-2026-04-14.md`).

---

## Фаза 1: Данные — привязка снапшотов к серверу

### 1.1. Добавить `ServerURL` в Meta и ManifestEntry

**Проблема**: Снапшоты не содержат информации о сервере, с которым была совершена операция.
При наличии нескольких серверов — невозможно понять, к какому серверу относится снапшот.

**Решение**:

- `internal/snap/types.go`: Добавить поле `ServerURL string` в `Meta`
- `internal/snap/manifest.go`: Добавить поле `ServerURL string` в `ManifestEntry`
- `internal/snap/snapshot.go` → `BuildMeta()`: Принимать `serverURL` как параметр
- `internal/snap/hook.go` → `HookMutation()`: Читать `viper.GetString("base_url")` и передавать в `BuildMeta`
- Обратная совместимость: старые снапшоты без `server_url` — показывать как `(unknown)`

**Файлы**: `types.go`, `manifest.go`, `snapshot.go`, `hook.go`

---

## Фаза 2: snap list — интерактивный select со скроллом

### 2.1. Двухуровневый пикер: сервер → снапшоты

**Проблема**: Плоский список 1000+ записей дезориентирует.

**Решение**:

Интерактивный режим (по умолчанию):
1. Сгруппировать снапшоты по `ServerURL`
2. Шаг 1: `p.Select("Select server:", serverOptions)` — если серверов >1
3. Шаг 2: `p.Select("Select snapshot:", filteredSnapshots)` — показать снапшоты выбранного сервера
4. После выбора — вывести команду `gotr snap info <id>` как подсказку

Non-interactive / `--format` режим:
- Оставить табличный вывод как есть (для CI/CD)
- Добавить колонку `SERVER` в таблицу

**Файлы**: `cmd/snap/list.go`, `cmd/snap/interactive_helpers.go`

### 2.2. Формат элементов в select-пикере

Каждый элемент в select:
```
[N] op entity_type | status | tier | project_id | 2026-04-14 07:38
```

Примеры:
```
[1] add case        | available     | T1 | P:3  | 2026-04-14 07:38
[2] update section  | rolled_back   | T2 | P:3  | 2026-04-14 07:35
[3] sync_cases      | available     | T2 | P:10 | 2026-04-14 07:30
```

---

## Фаза 3: snap info — табличная карточка

### 3.1. Заменить JSON на структурированный табличный вывод

**Проблема**: Сырой JSON на 1000 строк — непригоден для интерактивной работы.

**Решение**:

Интерактивный режим (по умолчанию):
```
╔══ Snapshot Info ═══════════════════════════════════════╗

  ID:         cases/20260414T073804_add_bulk_0
  Server:     https://testrail.example.com
  Operation:  add case
  Category:   cases
  Tier:       T1 (full rollback)
  Status:     available
  Project:    3
  CLI:        gotr cases add --suite-id 1 ...
  Created:    2026-04-14 07:38:04
  Data size:  1.2 KB

╠══ Entities (5) ═══════════════════════════════════════╣

  #   TYPE    ID      
  1   case    555     
  2   case    556     
  3   case    557     
  ...

╠══ Rollback Log (0) ═══════════════════════════════════╣

  (no rollback attempts)

╚═══════════════════════════════════════════════════════╝
```

JSON-вывод — только по `--format json` или `--json` (для скриптов):
- Старое поведение сохраняется как fallback
- `--format json` выдаёт полный meta + data dump

**Файлы**: `cmd/snap/info.go`, возможно новый `cmd/snap/info_render.go`

---

## Фаза 4: snap rollback — UX ошибок и dry-run

### 4.1. Понятный dry-run preview

**Проблема**: Пользователь не понимает что такое "Dry-run preview" и что происходит.

**Решение**:

Заменить текст:
- `"Dry-run preview: ..."` → `"Предпросмотр изменений:"`  (ну, мы пишем на EN, но более понятно)
- Перед таблицей diff — пояснение: `"The following changes will be applied to <server>:"`
- После таблицы — `"Apply this rollback? (Y/n)"` оставить как есть

### 4.2. Мягкий пропуск при 400/404 ошибках

**Проблема**: При rollback add-операции — undo = delete case. Если case уже удалён вручную, API возвращает 400. Пользователь видит сырую ошибку и не знает что делать.

**Решение**:

В `internal/snap/rollback.go`:
- При `rollbackCaseAdd()` → `DeleteCase()` → ошибка 400/404:
  - Не фейлить весь rollback
  - Пометить entity как `skipped (already deleted)`
  - Добавить в rollback log с пометкой
  - Показать warning: `Entity 555 already deleted, skipping`
  - Продолжить с остальными entities
- То же для `rollbackCaseDelete()` → `AddCase()` → если section не существует

**Файлы**: `internal/snap/rollback.go`

### 4.3. Контекст сервера в rollback

Перед интерактивным подтверждением показать:
```
Server: https://testrail.example.com
Snapshot: cases/20260414T073804_add_bulk_0 (add case, T1)
```

**Файлы**: `cmd/snap/rollback.go`

---

## Фаза 5: snap export — интерактивный выбор файла

### 5.1. Промпт для имени файла и директории

**Проблема**: export молча сохраняет в текущую директорию без запроса.

**Решение**:

В интерактивном режиме, если `output_path` не указан:
1. `p.Input("Save directory:", "./")` — с дефолтом на текущую
2. `p.Input("Filename:", "snapshot_<sanitized_id>.json")` — с дефолтом
3. Объединить и сохранить

Non-interactive — текущее поведение (дефолтное имя в cwd).

**Файлы**: `cmd/snap/export.go`

---

## Фаза 6: Общий интерактивный пикер — сервер header

### 6.1. Обновить selectSnapshot для всех команд

Текущий `selectSnapshot()` в `interactive_helpers.go` — обновить:
- Если у снапшотов >1 уникального `ServerURL` — сначала пикер сервера
- Показать выбранный сервер как header перед пикером снапшота
- Формат элементов — как в 2.2

**Файлы**: `cmd/snap/interactive_helpers.go`

---

## Фаза 7: Тесты

### 7.1. Обновить существующие тесты

- Все тесты с `ManifestEntry` и `Meta` — добавить `ServerURL`
- Тесты интерактивного пикера — обновить формат строк
- Тест группировки по серверам

### 7.2. Новые тесты

- `TestCLI_SnapList_Interactive` — двухуровневый пикер
- `TestCLI_SnapInfo_TableOutput` — табличный вывод vs JSON
- `TestRollback_GracefulSkip_404` — мягкий пропуск при 400/404
- `TestExport_InteractivePrompt` — промпт имя + диектория
- `TestSelectSnapshot_MultiServer` — группировка по серверам

**Файлы**: `cmd/snap/smoke_test.go`, `internal/snap/*_test.go`

---

## Фаза 8: Документация

- Обновить `docs/ru/guides/commands/snap.md`
- Обновить `docs/en/guides/commands/snap.md`
- Синхронизировать оба языка
- Добавить примеры нового табличного вывода
- Описать группировку по серверам

---

## Порядок выполнения

| # | Фаза | Зависимости | Оценка сложности |
| --- | --- | --- | --- |
| 1 | Данные: ServerURL в Meta | — | Низкая |
| 2 | snap list: интерактивный select | Фаза 1 | Средняя |
| 3 | snap info: табличная карточка | Фаза 1 | Средняя |
| 4 | snap rollback: UX ошибок | — | Средняя |
| 5 | snap export: промпт файла | — | Низкая |
| 6 | Общий пикер: сервер header | Фаза 1 | Средняя |
| 7 | Тесты | Фазы 1-6 | Средняя |
| 8 | Документация | Фазы 1-7 | Низкая |

---

## Режим работы

Предлагаемый режим: **stepwise** — каждая фаза → отчёт → подтверждение.
