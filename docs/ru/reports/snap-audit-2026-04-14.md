# Аудит snap/rollback — 2026-04-14

Language: Русский

## Навигация

- [Документация](../index.md)
  - [Отчёты](index.md)

---

## 1. Общая статистика

| Метрика | Значение |
| --- | --- |
| Ветка | `feat/28-snap-rollback` |
| HEAD | `20a53e1` (11 коммитов) |
| Файлов изменено | 106 |
| Insertions / Deletions | +8 527 / −4 |
| Файлов исходного кода (snap) | 21 |
| Файлов тестов (snap) | 8 |
| Всего LOC (snap) | 6 892 |
| Исходный код LOC | 3 534 |
| Тестовый код LOC | 3 358 |
| Соотношение тест/код | 0.95:1 |

---

## 2. Тестирование

### 2.1. Результаты прогона

| Пакет | Тестов | PASS | FAIL | Время |
| --- | --- | --- | --- | --- |
| `internal/snap` | 79 | 79 | 0 | 0.049s |
| `cmd/snap` | 19 | 19 | 0 | 0.026s |
| **Итого** | **98** | **98** | **0** | **0.075s** |

### 2.2. Покрытие `internal/snap` — 76.0%

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
| export [id] [path] | OK | OK | — | OK | OK |
| delete [id] | OK | OK | — | OK | OK |
| gc | OK | OK | — | OK | OK |

Синхронизация RU/EN: идеальная. Оценка: 95/100.

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
| Тесты | 98/98 PASS, 76% coverage | PASS |
| Стандарты кода | 0 критических, 1 warning | PASS |
| Архитектура | Полное соответствие | PASS |
| Безопасность | LOW risk | PASS |
| Документация | 95/100, RU/EN синхрон | PASS |
| Build / Vet | Чисто | PASS |

---

## 8. Выявленные UX-проблемы (из ручного тестирования)

| # | Проблема | Серьёзность |
| --- | --- | --- |
| UX-1 | `snap list` — плоский табличный вывод, при 1000+ записей — дезориентация | HIGH |
| UX-2 | `snap info` — сырой JSON, непригоден для интерактива | HIGH |
| UX-3 | Нет группировки снапшотов по серверам | HIGH |
| UX-4 | `snap rollback` — Dry-run preview непонятен пользователю | MEDIUM |
| UX-5 | `snap rollback` — ошибка 400 без понятного объяснения и рекомендации | MEDIUM |
| UX-6 | `snap export` — нет интерактивного выбора имени файла и директории | MEDIUM |
| UX-7 | Интерактивный пикер не показывает контекст сервера | HIGH |

Эти проблемы вынесены в отдельный план исправлений.
