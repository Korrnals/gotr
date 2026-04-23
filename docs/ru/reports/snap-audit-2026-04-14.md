# Аудит snap/rollback + interactive UX — 2026-04-18

Language: Русский

## Навигация

- [Документация](../index.md)
  - [Отчёты](index.md)

---

## 1. Общая статистика

| Метрика | Значение |
| --- | --- |
| Ветка | `feat/28-snap-rollback` |
| HEAD | `28d47b0` (32 коммита) |
| Файлов изменено | 185 |
| Insertions / Deletions | +17 321 / −714 |
| Файлов исходного кода (snap) | 19 |
| Файлов тестов (snap) | 8 |
| Всего LOC (snap) | 9 705 |
| Исходный код LOC (snap) | 4 469 |
| Тестовый код LOC (snap) | 5 236 |
| Соотношение тест/код (snap) | 1.17:1 |
| Файлов исходного кода (interactive) | 16 |
| Файлов тестов (interactive) | 12 |
| Всего LOC (interactive) | 3 087 |
| Исходный код LOC (interactive) | 1 627 |
| Тестовый код LOC (interactive) | 1 460 |

---

## 2. Тестирование

### 2.1. Результаты прогона

| Пакет | Тестов | PASS | FAIL | Время |
| --- | --- | --- | --- | --- |
| `internal/snap` | 119 | 119 | 0 | 0.054s |
| `cmd/snap` | 38 | 38 | 0 | 0.049s |
| `internal/interactive` | 83 | 83 | 0 | 0.015s |
| `cmd/compare` | 300+ | 300+ | 0 | 3.5s |
| `cmd/sync` | 100+ | 100+ | 0 | 0.4s |
| `cmd/` | 400+ | 400+ | 0 | 3.5s |
| **Итого** | **1 002** | **1 002** | **0** | **~7.5s** |

### 2.2. Покрытие

| Пакет | Покрытие |
| --- | --- |
| `internal/snap` | 74.1% |
| `cmd/snap` | 49.6% |
| `internal/interactive` | 71.4% |
| `cmd/compare` | 88.7% |
| `cmd/sync` | 90.6% |
| `cmd/` | 94.1% |

| Файл | Покрытие | Примечание |
| --- | --- | --- |
| types.go | 100% | — |
| manifest.go | ~95% | `Latest()` — 44% |
| rollback.go | ~82% | `Rollback()` — 96.2% |
| store.go | ~78% | `atomicWriteJSON` error branches |
| snapshot.go | ~77% | `SnapOrWarn` — 0% (обёртка) |
| attachments.go | ~83% | `decompressToTemp` — 68.8% |
| hook.go | 0% | Тонкая обёртка, тестируется через CLI |
| resolve.go | 0% | Чтение флагов cobra, через CLI |
| info.go | 0% | 1 функция UI wrapper |

### 2.3. Функции с нулевым покрытием (12 шт.)

- `hook.go`: `NewHook`, `Before`, `FinalizeAdd`, `FinalizeSyncData`, `HookMutation`
- `resolve.go`: `RegisterFlags`, `ResolveDecision`, `ResolveName`, `ReadConfig`
- `snapshot.go`: `SnapOrWarn`
- `store.go`: `NewStore`
- `info.go`: `InfoBanner`

Все являются тонкими обёртками над cobra/viper/ui — тестируются косвенно через CLI smoke tests.

---

## 3. Качество кода

### 3.1. Статический анализ

| Инструмент | Результат |
| --- | --- |
| `go build ./...` | OK |
| `go vet ./...` | 0 findings |
| `go fmt` | Чисто |

### 3.2. Стандарты проекта (code-quality.md)

| Правило | Статус |
| --- | --- |
| `fmt.Errorf("...: %w", err)` обязательный wrap | PASS — 100% ошибок обёрнуты |
| Нет `return err` без обёртки | PASS |
| Нет `os.Exit`/`log.Fatal` в библиотечном коде | PASS |
| `GetClient()` для HTTP | PASS — через DI `clientFn` |
| `ctx context.Context` на всех I/O | PASS |
| Naked returns только в тестах | PASS |

### 3.3. Архитектурные правила (STANDARDS.md)

| Правило | Статус |
| --- | --- |
| `cmd/` → `internal/` (не наоборот) | PASS |
| `pkg/` → не `internal/` | PASS |
| Interface Segregation | PASS — `CasesAPI` (5), `AttachmentsAPI` (3), `PromptFunc` (1) |
| Constructor injection | PASS — `Register(rootCmd, clientFn)` |
| `RunE` (не `Run`) | PASS |
| Paths через `paths.*` | PASS |
| Атомарная запись файлов | PASS — `atomicWriteJSON()` |
| Thread-safe manifest | PASS — `sync.Mutex` + defer |

### 3.4. Единственный warning

| Файл | Строка | Описание |
| --- | --- | --- |
| `cmd/snap/rollback.go` | 61, 63, 102 | Дискарт ошибок от `cmd.Flags().GetBool/GetString` — безопасно (cobra гарантия), но стилистически неидеально |

---

## 4. Безопасность

**Общий уровень риска: LOW** — CLI-утилита с локальным хранилищем.

| Категория | Оценка | Обоснование |
| --- | --- | --- |
| Path Traversal | LOW | ID генерируется программно, sanitizeName чистит `/\: ` |
| Export path | INFO | CLI-утилита, аналог `cp file /path` |
| Temp files | LOW | `sanitizeAttachName` + `os.CreateTemp` в `os.TempDir()` |
| File permissions | LOW | `os.Create()` → umask 0644, `~/.gotr/snaps/` — пользовательская |
| Race conditions | LOW | `sync.Mutex` на manifest, однопользовательский |
| JSON deserialization | SAFE | Типобезопасный `json.Decoder` |
| DoS | SAFE | `AttachMaxFileMB` + `concurrentThreshold` |

---

## 5. Документация

| Подкоманда | RU doc | EN doc | Флаги | Примеры | Интерактив |
| --- | --- | --- | --- | --- | --- |
| list | OK | OK | — | OK | OK |
| info [id] | OK | OK | — | OK | OK |
| rollback [id] | OK | OK | --dry-run, --entity-ids | OK | OK |
| rollback list | OK | OK | — | OK | OK |
| rollback undo [id] | OK | OK | — | OK | OK |
| export [id] [path] | OK | OK | — | OK | OK |
| delete [id] | OK | OK | — | OK | OK |
| gc | OK | OK | — | OK | OK |
| `gotr work` хаб | OK | OK | — | OK | OK |
| Cross-navigation | OK | OK | — | OK | OK |

Синхронизация RU/EN: идеальная. Оценка: 98/100.

---

## 6. Хорошие практики

- Atomic file writes — `atomicWriteJSON()`
- Thread-safe manifest — `sync.Mutex` + `defer unlock`
- Graceful degradation — `SnapOrWarn()`
- Dry-run support — preview с diff-таблицей + confirm
- Resume capability — rollback пропускает уже откаченные
- Interface segregation — малые интерфейсы
- Categorized storage — `~/.gotr/snaps/{cases,sections,sync,custom}/`

---

## 7. Итоговая матрица

| Категория | Оценка | Вердикт |
| --- | --- | --- |
| Тесты | 1002 PASS / 0 FAIL, 74–94% coverage | PASS |
| Стандарты кода | 0 критических, 1 warning | PASS |
| Архитектура | Полное соответствие | PASS |
| Безопасность | LOW risk | PASS |
| Документация | 98/100, RU/EN синхрон | PASS |
| Build / Vet | Чисто | PASS |

---

## 8. Выявленные UX-проблемы (все исправлены)

| # | Проблема | Серьёзность | Статус | Коммит |
| --- | --- | --- | --- | --- |
| UX-1 | `snap list` — плоский табличный вывод | HIGH | ✅ FIXED | `3ec7a7a` |
| UX-2 | `snap info` — сырой JSON | HIGH | ✅ FIXED | `3ec7a7a` |
| UX-3 | Нет группировки по серверам | HIGH | ✅ FIXED | `3ec7a7a` |
| UX-4 | Dry-run preview непонятен | MEDIUM | ✅ FIXED | `c5951ed` |
| UX-5 | Ошибка 400 без объяснения | MEDIUM | ✅ FIXED | `c5951ed` |
| UX-6 | Export без интерактивного выбора | MEDIUM | ✅ FIXED | `3ec7a7a` |
| UX-7 | Пикер без контекста сервера | HIGH | ✅ FIXED | `595ba00` |

### Дополнительные UX-улучшения (commit `28d47b0` — `658a6ad`)

| Улучшение | Описание |
| --- | --- |
| AlignedLabels | Выровненные колонки во всех интерактивных пикерах |
| Browse | Пагинация с ← Back навигацией |
| ActionMenu | Унифицированные post-action меню с строковыми ключами |
| CrossNavOptions | Кросс-навигация Compare/Sync/Snap из любого post-action |
| HandleCrossNav | Серверный guard при переходе между командами |
| Server banner | `Server: <url>` в интерактивном режиме |
| `gotr work` | Интерактивный навигационный хаб с групповыми меню |
