# PLAN.md — Актуальный план работ по gotr

> **Обновлено:** 2026-03-12  
> **Версия:** 3.5  
> **Статус:** Active  
> **Источник актуализации:** Stage 8.0 завершён, release-3.0.0 подготовлена

---

## 1) Контекст и правила

- Все системные файлы проекта хранятся в `.protocols/project/`.
- Для контекст-хранилища используются `.protocols/data/db/context.db` и `.protocols/data/contexts/*`.
- При аудите и работах игнорируются директории: `.history/`, `.systems/`, `.testrail/`.
- Текущий план заменяет устаревшие процессные рудименты (включая «алгоритм работы с осью»).

---

## 2) Фактическое состояние проекта (на 2026-02-25)

### 2.1 Кодовая база

- Go-файлов: **383**
- Тестовых файлов (`*_test.go`): **147**
- Командный слой `cmd/`: **272** Go-файла
- Внутренний слой `internal/`: **108** Go-файлов

### 2.2 Архитектурные подсистемы

- `cmd/` — CLI слой на Cobra, широкий набор команд по ресурсам TestRail
- `internal/client/` — HTTP-клиент и API-методы
- `internal/service/` + `internal/service/migration/` — бизнес-логика и миграции
- `internal/parallel/` — новый контур рекурсивной параллелизации (Stage 6.7)
- `internal/progress/`, `internal/concurrent/`, `internal/interactive/` — прогресс, concurrency и UX-утилиты

### 2.3 Текущая стабильность

Результат `go test ./...` (2026-03-04):

- **Все пакеты проходят успешно** — включая `internal/parallel` (ранее были падения, исправлены)
- Сборка `go build ./...` чистая
- Удалён ~907 строк мёртвого кода (коммит `9007d95`)

Вывод: Stage 6.7 стабилизирован, аудит проведён, остался рефакторинг дубликатов и улучшение error wrapping.

---

## 3) Текущая цель проекта

Стабилизировать и довести до production-ready состояние контура параллельной загрузки/сравнения данных (Stage 6.7), не нарушив существующую CLI-функциональность и покрытие API.

---

## 4) Приоритетный roadmap

### P0 — Стабилизация `internal/parallel` ✅ ЗАВЕРШЕНО

> Завершено в Stage 6.7. Все тесты проходят стабильно.

### P1 — Интеграционная валидация compare-потоков ✅ ЗАВЕРШЕНО

> Завершено в Stage 6.7. Реальный прогон `gotr compare cases --pid1 30 --pid2 34` — корректен.

### P2 — Синхронизация документации ✅ ЗАВЕРШЕНО

> Документация актуализирована. Reporter переписан на go-pretty.

### P3 — Подготовка к выпуску ⏳ В ПРОЦЕССЕ

1. Ветка `release-3.0.0` создана для сборки релиза.
2. Stage 6.7 включается через PR в релизную ветку.
3. Stage 6.8 — следующий этап перед финальным релизом.

### P6 — Унификация конкурентности и compare-подкоманд (Stage 6.8) ✅ ЗАВЕРШЕНО

> **Дизайн-документ:** `STAGE_6.8_DESIGN.md`  
> **Ветка:** `stage-6.8-concurrency-unification`  
> **Утверждён:** 2026-03-05

**Критерий завершения P6:** все compare-подкоманды используют пакет `internal/concurrency/`, ~1200 строк копипасты устранено, реальный прогон `gotr compare all` проходит корректно.

---

### P7 — Аудит пагинации и масштабирование конкурентности (Stage 6.9) ✅ ЗАВЕРШЁН

> **Дизайн-документ:** `STAGE_6.9_DESIGN.md`  
> **Аудит:** `AUDIT_2026-03-06_pagination.txt`  
> **Ветка:** `stage-6.9-pagination-audit`  
> **Старт:** 2026-03-05  
> **Завершён:** 2026-03-06  
> **PR:** в `release-3.0.0`

#### Выполнено

1. ✅ Полный аудит всех API-методов на предмет пагинации  
2. ✅ `internal/client/paginator.go` — generic `fetchAllPages[T]` + `decodeListResponse[T]`  
3. ✅ Миграция 9 критичных методов: GetRuns, GetPlans, GetSections, GetSharedSteps, GetMilestones, GetResults, GetResultsForRun, GetTests, GetSuites  
4. ✅ 11 unit-тестов (`paginator_test.go`), все зелёные  
5. ✅ Smoke-тест: compare all P30+P34 — 20 509 + 116 009 кейсов, пагинация подтверждена  
6. ✅ CHANGELOG.md обновлён, документация актуальна  

> **Примечание:** Phase 3 (GetResultsForCase, GetProjects, GetHistoryForCase) — не требует миграции.  
> Эти методы возвращают < 250 записей по природе запроса (один кейс/ран/проект).

**Критерий завершения P7:** ✅ выполнен — все критичные методы используют пагинацию, compare-подкоманды корректно работают с большими данными.

---

### P8 — Проброс context.Context (Stage 7.0) ✅ ЗАВЕРШЁН

> **Дизайн-документ:** `STAGE_7.0_DESIGN.md`
> **Ветка:** `stage-7.0-context-propagation`
> **Дизайн утверждён:** 2026-03-06
> **Завершён:** 2026-03-07
> **PR:** в `release-3.0.0`

#### Проблема

Все ~100 методов `*HTTPClient` и весь `cmd/`-слой (~270 файлов) не передают
`context.Context` в HTTP-запросы. Следствие:
- Ctrl+C не прерывает in-flight запросы (goroutine leak)
- Нет timeout-propagation из Cobra-команд
- Stage P4 (Rate Limiting с Retry-After) невозможен без ctx

#### Решение (AD принятые в дизайн-документе)

1. **Без переименования** — изменить сигнатуры, не создавать `*Ctx` варианты
2. **Bottom-up**: `client.go` → `paginator.go` → `internal/client/*.go` →
   `internal/service/` → `internal/concurrent/` → `cmd/`
3. **Cobra даёт ctx бесплатно**: `ctx := cmd.Context()` — автоотмена по Ctrl+C
4. **`http.NewRequestWithContext`** — в `Get`/`Post` в `client.go`

#### Оценка
~1-2 дня (большая часть — механическая замена в cmd/ через `go build` навигацию).

**Критерий завершения P8:** ✅ выполнен — build чист, тесты зелёные, Ctrl+C работает (signal.NotifyContext + ExecuteContext), нет одиночных Get без ctx.

#### Выполнено

1. ✅ `http.NewRequestWithContext` в transport (`client.go:DoRequest`)
2. ✅ `fetchAllPages[T](ctx, ...)` в `paginator.go`
3. ✅ Все ~100 методов `*HTTPClient` принимают ctx
4. ✅ `interfaces.go` + `mock.go` обновлены
5. ✅ `internal/service/`, `internal/concurrent/` обновлены
6. ✅ Все ~270 файлов `cmd/` используют `ctx := cmd.Context()`
7. ✅ `main.go`: `signal.NotifyContext` + `cmd.Execute(ctx)`
8. ✅ `cmd/root.go`: `Execute(ctx context.Context)` + `rootCmd.ExecuteContext(ctx)`
9. ✅ `cancel_test.go`: `TestGetRuns_Cancellation` + `TestHTTPClient_CancelledContext`

---

### P9 — UI/Output Refactoring (Stage 8.0) ✅ ЗАВЕРШЁН

> **Дизайн-документ:** `STAGE_8.0_DESIGN.md`
> **Ветка:** `stage-8.0-output-refactoring` (от `stage-7.0-context-propagation`)
> **Старт:** 2026-03-07
> **Завершён:** 2026-03-12
> **Влит в:** `release-3.0.0`

#### Выполнено

1. ✅ [8.1] `tabwriter` → `ui.Table` во всех list/get командах
2. ✅ [8.2] `json.MarshalIndent` → `ui.JSON`
3. ✅ [8.3] `--format` PersistentFlag на root
4. ✅ [8.4] `*Var` → `GetFlag` миграция
5. ✅ [8.5] `internal/flags` пропагация + перевод ошибок на English
6. ✅ [8.6] sync/interactive — закрыто (консолидация не нужна)
7. ✅ [8.7] `os.Exit` → `panic` в `GetClient*`
8. ✅ [8.8] `fmt.Print*` → `ui.*` (49 файлов, +272/−185)

**Коммит:** `0e49b5f` (код) + `432dfac` (документация)

**114 `fmt.Print*` оставлены осознанно** — interactive prompts, data output, debug.

---

### P10 — Release 3.0.0 ⏳ ПОДГОТОВЛЕН

> **Ветка:** `release-3.0.0`
> **Branch protection:** main защищён, только через PR
> **Протокол:** `.github/copilot/instructions/workflow/RELEASE.md`

#### Состояние

- ✅ Stages 6.7–8.0 влиты в `release-3.0.0`
- ✅ CHANGELOG.md обновлён
- ✅ `go test ./...` — 32 ok, 0 fail
- ✅ `go build` — чисто
- ✅ `go vet` — чисто
- ⬜ Smoke tests с реальным TestRail
- ⬜ PR: `release-3.0.0` → `main`
- ⬜ Тег `v3.0.0`

---

### Следующие этапы (после release 3.0.0)

| Stage | Описание | Дизайн-документ | Ветка |
|-------|----------|-----------------|-------|
| 9.0 | Стандарты проекта | `STAGE_9.0_DESIGN.md` | `stage-9.0-standards` |
| 10.0 | Полный аудит и рефакторинг | `STAGE_10.0_DESIGN.md` | `stage-10.0-audit` |
| 11.0 | Единая интерактивная система | `STAGE_11.0_DESIGN.md` | `stage-11.0-interactive` |
| 12.0 | Финальный рефакторинг | `STAGE_12.0_DESIGN.md` | `stage-12.0-final` |

---

### P4 — Rate Limiting и устойчивость к API-ограничениям (отложено)

> **Источник:** Официальная документация TestRail API  
> **Ссылка:** https://support.testrail.com/hc/en-us/articles/7077083596436  
> **Добавлено:** 2026-02-26

#### Контекст
По API-документации TestRail Cloud имеет rate limits:
- **180 req/min** — Professional
- **300 req/min** — Enterprise  
- **Без ограничений** — TestRail Server (self-hosted)

При параллельной загрузке кейсов из ~30 сьютов мы легко можем превысить лимит
Cloud (десятки запросов в секунду).

#### Что нужно реализовать

1. **Обработка HTTP 429 (Too Many Requests):**
   - Парсить заголовок `Retry-After` из ответа
   - Автоматический retry с backoff после паузы
   - Не считать 429 за фатальную ошибку

2. **Опциональный Rate Limiter:**
   - Конфигурируемый лимит запросов в минуту (флаг `--rate-limit`)
   - Дефолт: `0` (без ограничений для Server)
   - Presets: `--rate-limit cloud-pro` (180), `--rate-limit cloud-ent` (300)
   - Реализация через `golang.org/x/time/rate` или аналог

3. **Адаптивное регулирование:**
   - При получении 429 автоматически снижать concurrency
   - Логирование warning при приближении к лимиту

**Критерий завершения P4:** gotr корректно работает с TestRail Cloud без 429-ошибок
при дефолтных настройках.

---

## 5) Текущее состояние `.protocols/data`

### 5.1 SQLite (`.protocols/data/db/context.db`)

- Таблицы: `sessions`, `contexts`, `checkpoints`
- Активная сессия: `sess_20260304_audit_refactor`
- Последний checkpoint: `checkpoint-dead-code-removal-2026-03-04`

### 5.2 Context-файлы (`.protocols/data/contexts/`)

- `session-latest.json` — инициализирована активная сессия
- `checkpoint-init.md` — чекпоинт инициализации
- `checkpoint-audit-2026-02-25.md` — расширенный аудит

---

### P5 — Кэширование и инкрементальная загрузка (планируется)

> **Добавлено:** 2026-02-26  
> **Источник:** обсуждение оптимизации после отказа от pre-count

#### Проблема
TestRail API не предоставляет `total count` за один вызов. Полная загрузка всех кейсов при каждом запуске — основной bottleneck. Для 36k+ кейсов это сотни API-запросов.

#### Решение: двухфазный кэш

1. **Первый запуск** — полная индексация всех кейсов, сохранение в локальный кэш (SQLite или JSON):
   - `case_id`, `title`, `suite_id`, `section_id`, `updated_on`, `created_on`
   - Метка времени последней синхронизации

2. **Последующие запуски** — инкрементальная загрузка:
   - Фильтр `&updated_after=<last_sync_timestamp>` для получения только изменённых
   - Фильтр `&created_after=<last_sync_timestamp>` для получения новых
   - Merge с кэшем

3. **Валидация** — периодическая полная сверка (раз в N запусков или по флагу `--force-full`)

#### API поддержка
TestRail API поддерживает фильтры:
- `created_after` / `created_before` (UNIX timestamp)
- `updated_after` / `updated_before` (UNIX timestamp)

#### Предварительная оценка
- Хранилище: `~/.gotr/cache/` или `$XDG_CACHE_HOME/gotr/`
- Формат: SQLite (один файл per TestRail instance)
- Флаги CLI: `--no-cache`, `--force-full`, `--cache-ttl`
- Приоритет: после стабилизации P0-P1

---

## 6) Опорные артефакты проекта

- Архитектура проекта: `.protocols/project/ARCHITECTURE.md`
- Stage 6.7 дизайн: `.protocols/project/STAGE_6.7_DESIGN.md`
- Stage 6.7 статус: `.protocols/project/STAGE_6.7_IMPLEMENTATION.md`
- Stage 6.9 дизайн: `.protocols/project/STAGE_6.9_DESIGN.md`
- Stage 7.0 дизайн: `.protocols/project/STAGE_7.0_DESIGN.md`
- Аудит пагинации (актуальный): `.protocols/project/AUDIT_2026-03-06_pagination.txt`
- TestRail API pagination ref: `.protocols/project/TESTRAIL_API_PAGINATION_REF.md`
- Пользовательская архитектура: `docs/architecture.md`

---

## 7) Рабочий протокол обновления этого плана

План обновляется после каждого значимого изменения одного из пунктов:

1. стабильность тестов (`internal/parallel`),
2. архитектурная схема/границы слоёв,
3. статус `.protocols/data` (session/checkpoint),
4. готовность к релизному этапу.

---

**Версия плана:** 3.1  
**Владелец:** проектный контур `.protocols/project/`
