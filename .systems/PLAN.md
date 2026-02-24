# PLAN.md — Общий план разработки gotr

> **Дата создания:** 2026-02-03  
> **Последнее обновление:** 2026-02-10  
> **Версия плана:** 2.0  
> **Статус:** Активный  
> **ВАЖНО:** Этот файл обновляется только после аудита и синхронизации с API_ANALYSIS.md  
> **НЕЛЬЗЯ создавать отдельные файлы планов для этапов — только этот файл!**

---

## ⚠️ ВАЖНО: Алгоритм работы с осью (ОБЯЗАТЕЛЕН)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         АЛГОРИТМ РАБОТЫ С ОСЬЮ                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  1. ОБЩИЙ АУДИТ (API_AUDIT.md)                                              │
│     ↓ Проверка всего проекта, выявление проблем                             │
│                                                                             │
│  2. СВЕРКА С ДОКУМЕНТАЦИЕЙ (API_ANALYSIS.md)                                │
│     ↓ Актуализация согласно текущему состоянию                              │
│                                                                             │
│  3. КОРРЕКТИРОВКА ОБЩЕГО ПЛАНА (PLAN.md) ← ТЫ ЗДЕСЬ                        │
│     ↓ Синхронизация плана с реальным состоянием                             │
│                                                                             │
│  4. ФОРМИРОВАНИЕ ЧЕКЛИСТА (CHECKLIST.md)                                    │
│     ↓ Подробное задание для текущего этапа                                  │
│                                                                             │
│  5. ПЛАН ДЕЙСТВИЙ ЭТАПА → УТВЕРЖДЕНИЕ                                       │
│     ↓ Получение подтверждения перед началом работы                          │
│                                                                             │
│  6. ВЫПОЛНЕНИЕ ЭТАПА → РАБОТА С КОДОМ                                       │
│     ↓ Следование лучшим практикам (DRY, KISS, SOLID)                        │
│                                                                             │
│  7. ЗАВЕРШЕНИЕ ЭТАПА → ОБЯЗАТЕЛЬНЫЕ ДЕЙСТВИЯ:                              │
│     • Уточнить: «Стоит ли выполнить модульные коммиты?»                     │
│     • Уточнить: «Необходимо ли выполнить очередной Релиз?»                  │
│     • Обновить CHANGELOG.md                                                 │
│     • Обновить версию проекта (semver)                                      │
│     • Обновить документацию проекта                                         │
│     • Зафиксировать изменения в комментариях коммитов                       │
│     • Синхронизировать все файлы оси                                        │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Правила для AI Assistant:

0. **ЛЮБАЯ ФИКСАЦИЯ ИЗМЕНЕНИЙ** должна быть подтверждена пользователем ЯВНО.
1. **ЗАПРЕЩЕНО УДАЛЕНИЕ** файлов - без ЯВНОГО подтверждения! Необходимо ВЫДЕЛЯТЬ такие изменения ОСОБО.
2. **КАЖДЫЙ этап** начинается с ОБЩЕГО АУДИТА (API_AUDIT.md)
3. **НЕЛЬЗЯ** создавать отдельные файлы планов для этапов — только PLAN.md
4. **ВСЕ файлы оси** должны быть максимально синхронизированы
5. **Перед началом** этапа — получить утверждение плана
6. **После завершения** — выполнить все обязательные действия
7. **При разработке** — следовать DRY, KISS, SOLID, проводить рефакторинг

---

## 📌 Версионирование проекта

Проект использует [Semantic Versioning 2.0.0](https://semver.org/lang/ru/):

```
MAJOR.MINOR.PATCH[-prerelease]

MAJOR — ломающие изменения (несовместимые API изменения)
MINOR — новые фичи (обратно совместимые)
PATCH — багфиксы (обратно совместимые)
-prerelease — предрелизные теги (-dev, -alpha, -beta, -rc)
```

### Текущая версия
- **Разработка:** `2.7.0-dev` (Stage 5 в процессе)
- **Последний релиз:** `2.2.3`

### Правила обновления версии

#### При разработке (feature-ветки):
- Начинаем с `-dev` (например, `2.4.0-dev`)
- По мере стабилизации: `-alpha` → `-beta` → `-rc`
- Релиз: убираем суффикс

#### При релизе:
- Новые фичи: `MINOR++` (2.3.0 → 2.4.0)
- Багфиксы: `PATCH++` (2.3.0 → 2.3.1)
- Ломающие изменения: `MAJOR++` (2.3.0 → 3.0.0)

---

## 🎯 Общее видение проекта

### Цель
Создать полнофункциональный CLI-клиент для TestRail API v2 с:
- Полным покрытием API (106 endpoint'ов)
- Удобным интерактивным режимом
- Возможностью миграции/синхронизации данных
- Высоким качеством кода (тесты, документация)

### Текущее состояние (актуально на 2026-02-10)

```
┌────────────────────────────────────────────────────────────┐
│                    СТАТИСТИКА ПРОЕКТА                       │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  📊 Покрытие API:                                          │
│  • Реализовано: 106/106 endpoint'ов (100%)                 │
│  • CLI команды: покрытие высокое, осталось частичное       │
│  • Тесты: 70 тестовых файлов (клиент/сервис/CLI)            │
│                                                            │
│  ✅ Полностью реализовано (Client API):                    │
│  • Cases API (10/10)                                       │
│  • Runs API (6/6)                                          │
│  • Results API (7/7)                                       │
│  • Projects API (5/5)                                      │
│  • Tests API (3/3)                                         │
│  • Milestones API (5/5)                                    │
│  • Plans API (9/9)                                         │
│  • Attachments API (5/5)                                   │
│  • Configurations API (7/7)                                │
│  • Users API (4/4)                                         │
│  • Priorities (1/1)                                        │
│  • Statuses (1/1)                                          │
│  • Templates (1/1)                                         │
│  • Reports API (3/3)                                       │
│  • Groups API (5/5)                                        │
│  • Roles API (2/2)                                         │
│  • ResultFields (1/1)                                      │
│  • Datasets API (5/5)                                      │
│  • Variables API (4/4)                                     │
│  • BDDs API (2/2)                                          │
│  • Labels API (2/2)                                        │
│                                                            │
│  🟡 Частично реализовано (Client ✅, CLI ⏳):               │
│  • Sections API (GET) — нет CLI                            │
│  • Users/Roles (GET) — нет CLI                             │
│                                                            │
│  ⚪ Не реализовано:                                        │
│  • Нет — все 106 endpoint'ов реализованы!                  │
│                                                            │
│  ⚠️ Критические проблемы:                                  │
│  • Интерактивный режим — непоследовательность              │
│  • Sync тесты — отключены (нужен рефакторинг)              │
│  • Нет валидации дефолтных конфиг-значений                 │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

---

## 📋 ЭТАПЫ РАЗРАБОТКИ

### ЭТАП 0: Анализ и аудит ✅ (Завершён)

**Период:** 2026-02-03  
**Статус:** ✅ Завершён  
**Версия:** —

**Задачи:**
- [x] Составить таблицу всех 91 endpoint'а TestRail API
- [x] Отметить реализованные endpoint'ы
- [x] Выявить критические пробелы
- [x] Создать систему файлов оси

**Артефакты:**
- `.systems/API_ANALYSIS.md` — полный анализ API
- `.systems/API_AUDIT.md` — аудит качества
- `.systems/ARCHITECTURE.md` — архитектурная документация

---

### ЭТАП 1: Реализация Results + Runs API ✅ (Завершён)

**Период:** 2026-02-03 — 2026-02-04  
**Статус:** ✅ Завершён  
**Версия:** 2.3.0 → 2.4.0-dev

**Задачи:**
- [x] Реализовать Results API (7 endpoints)
- [x] Реализовать Runs API (6 endpoints)
- [x] Создать Service Layer
- [x] Добавить CLI команды
- [x] Написать тесты

**Артефакты:**
- `internal/client/results.go` — 7 методов
- `internal/client/runs.go` — 6 методов
- `cmd/result/*.go` — CLI команды
- `cmd/run/*.go` — CLI команды
- `internal/service/*.go` — Service Layer

---

### ЭТАП 2.5: Тестирование и логирование ✅ (Завершён)

**Период:** 2026-02-04 — 2026-02-05  
**Статус:** ✅ Завершён  
**Версия:** 2.4.0-dev

**Задачи:**
- [x] Добавить POST методы для Cases/Suites/sections/shared_steps
- [x] Интегрировать zap-логгер
- [x] Реализовать self-test команду
- [x] Исправить путь конфигурации
- [x] Добавить проверку дефолтных конфиг-значений

**Критические исправления:**
- [x] Путь конфигурации: `~/.gotr/config/default.yaml`
- [x] Проверка placeholder'ов в конфиге
- [x] Исправлены дублирующиеся объявления

**Артефакты:**
- 22 новых POST теста
- `cmd/selftest.go`
- `internal/paths/`
- `internal/log/`

---

### ЭТАП 3: Исправление базовых компонентов ✅ (Завершён)

**Период:** 2026-02-05 — 2026-02-07  
**Статус:** ✅ Завершён  
**Версия:** 2.5.0 (релиз)

#### ✅ Аудит и подготовка (выполнено):
- [x] Провести полный аудит проекта
- [x] Обновить алгоритм работы с осью
- [x] Синхронизировать все файлы оси
- [x] Удалить лишние stage-файлы

#### 🔄 Задачи в работе:

**Часть A: Интерактивный режим (Приоритет 1)** ✅ ЗАВЕРШЕНА
- [x] Создать единый пакет `internal/interactive`
- [x] Рефакторинг `cmd/get/` — использование общего пакета
- [x] Рефакторинг `cmd/run/list.go` — добавлен интерактивный выбор
- [x] Рефакторинг `cmd/result/` — добавлена команда `list`

**Часть B: Полный ClientInterface + MockClient (Приоритет 2)** ✅ ЗАВЕРШЕНА
- [x] Создать `internal/client/interfaces.go` — полный `ClientInterface` (43 метода)
- [x] Создать `internal/client/mock.go` — `MockClient` со всеми методами
- [x] Проверка компиляции и тестов

**Часть C: Рефакторинг дублирования (Приоритет 3)** ✅ ЗАВЕРШЕНА
- [x] Создать `cmd/common` пакет с общими утилитами
- [x] Рефакторинг `cmd/result`, `cmd/run`, `cmd/sync`
- [x] Удалить дублирование кода доступа к клиенту
- [x] Обновить тесты с migrationMock

**Критические проблемы для решения:**
- ✅ Непоследовательность интерактивного режима — ИСПРАВЛЕНО
- 🔄 Отключенные sync-тесты — В РАБОТЕ
- ⏳ Дублирование кода в CLI командах — ЗАПЛАНИРОВАНО
- ⏳ Отсутствие единого стиля обработки ошибок — ЗАПЛАНИРОВАНО

#### 📋 Дополнительно выполнено (вне плана):
- [x] Единый `--dry-run` флаг для всех state-changing команд
- [x] Централизованный пакет `cmd/common/dryrun`
- [x] Базовый wizard пакет `cmd/common/wizard` на survey/v2
- [x] `--interactive/-i` флаг для `add` и `update` (основные endpoints)

#### ⏳ Расширение интерактивности (техдолг):
- [ ] Довести `--interactive` до всех команд (delete, run, result, sync)
- [ ] Единый UX pattern: ввод → предпросмотр → подтверждение/отмена

---

### ЭТАП 4: Полное покрытие API ✅ (Завершён)

**Период:** 2026-02-07 — 2026-02-08  
**Статус:** ✅ Завершён  
**Версия:** 2.5.0 → 2.7.0-dev

**Цель:** Реализовать ВСЕ оставшиеся endpoint'ы TestRail API (106/106)

#### Приоритет 1 (Критично для CI/CD):

**Tests API** — 3 endpoint'а ✅ (100%):
- [x] `GetTest(testID int64)` — получение информации о тесте
- [x] `GetTests(runID int64, filters ...)` — список тестов в ране
- [x] `UpdateTest(testID int64, req *UpdateTestRequest)` — обновление теста
- [x] CLI: `gotr test get`, `gotr test list`

**Plans API** — 9 endpoint'ов ✅ (100%):
- [x] `GetPlan(planID int64)` — получение плана
- [x] `GetPlans(projectID int64)` — список планов проекта
- [x] `AddPlan(projectID int64, req *AddPlanRequest)` — создание плана
- [x] `UpdatePlan(planID int64, req *UpdatePlanRequest)` — обновление плана
- [x] `ClosePlan(planID int64)` — закрытие плана
- [x] `DeletePlan(planID int64)` — удаление плана
- [x] `AddPlanEntry(planID int64, req *AddPlanEntryRequest)` — добавление entry
- [x] `UpdatePlanEntry(planID, entryID string, req *UpdatePlanEntryRequest)` — обновление entry
- [x] `DeletePlanEntry(planID, entryID string)` — удаление entry
- [x] CLI: `gotr plan get`, `gotr plan list`, `gotr plan add`, `gotr plan update`, `gotr plan close`, `gotr plan delete`, `gotr plan add-entry`, `gotr plan update-entry`, `gotr plan delete-entry`

#### Приоритет 2 (Важно):

**Milestones API** — 5 endpoint'ов ✅ (100%):
- [x] `GetMilestone(milestoneID int64)` — получение milestone
- [x] `GetMilestones(projectID int64)` — список milestone проекта
- [x] `AddMilestone(projectID int64, req *AddMilestoneRequest)` — создание
- [x] `UpdateMilestone(milestoneID int64, req *UpdateMilestoneRequest)` — обновление
- [x] `DeleteMilestone(milestoneID int64)` — удаление
- [x] CLI: `gotr milestone get`, `gotr milestone list`, `gotr milestone add`, `gotr milestone update`, `gotr milestone delete`

#### Артефакты (будут созданы):
- 📁 `internal/models/data/tests.go` — уже существует, дополнить при необходимости
- 📁 `internal/models/data/plans.go` — новый файл
- 📁 `internal/models/data/milestones.go` — новый файл
- 📁 `internal/client/tests.go` — новый файл
- 📁 `internal/client/plans.go` — новый файл
- 📁 `internal/client/milestones.go` — новый файл
- 📁 `cmd/test/*.go` — CLI команды для Tests
- 📁 `cmd/plan/*.go` — CLI команды для Plans
- 📁 `cmd/milestone/*.go` — CLI команды для Milestones

#### Реализовано в Stage 4 (44 endpoint'а):

**Приоритет 1 — Основной функционал:**
- ✅ Tests API — 3 endpoint'а
- ✅ Plans API — 9 endpoint'ов
- ✅ Milestones API — 5 endpoint'ов
- ✅ Attachments API — 5 endpoint'ов
- ✅ Configurations API — 7 endpoint'ов
- ✅ Users/Priorities/Statuses/Templates — 7 endpoint'ов

**Приоритет 2 — Расширенный функционал:**
- ✅ Reports API — 3 endpoint'а
- ✅ Groups API — 5 endpoint'ов
- ✅ Roles API — 2 endpoint'а
- ✅ ResultFields API — 1 endpoint
- ✅ Datasets API — 5 endpoint'ов
- ✅ Variables API — 4 endpoint'а
- ✅ BDDs API — 2 endpoint'а
- ✅ Labels API — 2 endpoint'а

#### Критерии завершения:
- [x] Все 106 endpoint'ов реализованы (106/106 ✅)
- [x] Все модели данных созданы
- [x] Все client методы покрыты тестами
- [x] MockClient обновлён для всех API
- [x] Все тесты проходят (118+ тестов)
- [x] Обновлена документация (API_ANALYSIS.md, PLAN.md)
- [ ] CLI команды — запланировано на Stage 5

---

### ЭТАП 5: Полный CLI coverage ✅ (ЗАВЕРШЁН)

**Период:** 2026-02-11 — 2026-02-11  
**Статус:** ✅ Завершён (Вариант C — Полный)  
**Версия:** 2.7.0-dev → 2.7.0
**Scope:** 28 endpoints → CLI ~94% ✅

**Результаты:**
- CLI покрытие: ~74% → ~94% (+20%)
- Добавлено команд: 20+
- Создано тестов: 50+
- Создано пакетов: 7 (users, templates, roles, reports, + дополнения groups/tests/get)

**Цель:** Довести CLI покрытие с ~74% до ~90%+

#### 📊 Текущее состояние (по результатам аудита 2026-02-11)

```
Client API:    106/106 (100%) ✅
CLI покрытие:  ~78/106 (~74%)
Полностью:     14 ресурсов
Частично:      4 ресурса (Sections, Groups, Tests, CaseFields)
Не реализовано: 6 ресурсов (Users, Priorities, Statuses, Templates, Roles, ResultFields, Reports)
```

#### 🎯 Варианты Scope (на утверждение)

**Вариант A: Минимальный (P0) — 8 endpoints**
| Приоритет | Ресурс | Endpoints | CLI команды |
|-----------|--------|-----------|-------------|
| 🔴 P0 | Users | 3 | `users list`, `users get`, `users get-by-email` |
| 🔴 P0 | Priorities | 1 | `get priorities` |
| 🔴 P0 | Statuses | 1 | `get statuses` |
| 🔴 P0 | Templates | 1 | `templates list` |
| 🔴 P0 | Sections | 2 | `get section`, `get sections` |
| **Итого** | | **8** | |

**Результат:** CLI ~81% (+7%)

---

**Вариант B: Расширенный (P0+P1) — 15 endpoints**
Дополнительно к варианту A:
| Приоритет | Ресурс | Endpoints | CLI команды |
|-----------|--------|-----------|-------------|
| 🟡 P1 | Tests | 2 | `tests get`, `tests list` |
| 🟡 P1 | Groups | 3 | `groups add`, `groups update`, `groups delete` |
| 🟡 P1 | Roles | 2 | `roles list`, `roles get` |
| **Итого с A** | | **15** | |

**Результат:** CLI ~88% (+14%)

---

**Вариант C: Полный — 28 endpoints**
Дополнительно к варианту B:
| Приоритет | Ресурс | Endpoints | CLI команды |
|-----------|--------|-----------|-------------|
| 🟢 P2 | ResultFields | 1 | `get result-fields` |
| 🟢 P2 | Reports | 3 | `reports list`, `reports run`, `reports run-cross-project` |
| 🟢 P2 | CaseFields | 1 | `add case-field` |
| **Итого** | | **28** | |

**Результат:** CLI ~94% (+20%)

#### 📋 Детальный план (рекомендуется Вариант B)

**Часть 1: Справочники (Reference Data)**
- [ ] `cmd/users/` — Users CLI (list, get, get-by-email)
- [ ] `cmd/roles/` — Roles CLI (list, get)
- [ ] Добавить в `cmd/get/types_fields.go`:
  - [ ] `get priorities`
  - [ ] `get statuses`
  - [ ] `get result-fields`
- [ ] `cmd/templates/` — Templates CLI (list)

**Часть 2: Незавершенные ресурсы**
- [ ] `cmd/get/sections.go` — Sections GET (get, get sections)
- [ ] `cmd/tests/get.go`, `cmd/tests/list.go` — Tests GET
- [ ] `cmd/groups/add.go`, `cmd/groups/update.go`, `cmd/groups/delete.go` — Groups CRUD

**Часть 3: Тесты (следуя примеру существующих)**
- [ ] Тесты для Users (как `cmd/milestones/*_test.go`)
- [ ] Тесты для Sections GET
- [ ] Тесты для Tests GET
- [ ] Тесты для Groups CRUD

**Часть 4: Регистрация команд**
- [ ] Обновить `cmd/commands.go`
- [ ] Обновить `cmd/resources.go` (если нужно)

#### 🧪 Тестовое покрытие (требования)

Каждая новая команда должна иметь тесты, аналогичные существующим:

```go
// Пример теста (как в cmd/milestones/add_test.go)
func TestAddCommand(t *testing.T) {
    tests := []struct {
        name    string
        args    []string
        wantErr bool
    }{
        {
            name:    "valid add",
            args:    []string{"add", "--project-id", "1", "--name", "Test"},
            wantErr: false,
        },
        // ...
    }
}
```

#### 📝 Шаблон новой команды

```go
// cmd/[resource]/[action].go
package [resource]

import (
    "github.com/Korrnals/gotr/internal/client"
    "github.com/spf13/cobra"
)

func New[Action]Cmd(getClient func() client.ClientInterface) *cobra.Command {
    return &cobra.Command{
        Use:   "[action] [args]",
        Short: "Description",
        RunE: func(cmd *cobra.Command, args []string) error {
            // implementation
        },
    }
}
```

#### ⚠️ Важные замечания

1. **API Client уже реализован** — все методы есть в `internal/client/`
2. **Паттерн тестов известен** — использовать `cmd/milestones/` как референс
3. **Регистрация в commands.go** — не забыть добавить в `init()`
4. **Dry-run флаг** — добавить `--dry-run` для state-changing команд

#### 📁 Файлы для создания (пример для Варианта B)

```
cmd/
├── users/
│   ├── users.go          # Родительская команда
│   ├── list.go           # GET get_users
│   ├── get.go            # GET get_user/{id}
│   ├── list_test.go
│   ├── get_test.go
│   └── test_helper.go
├── roles/
│   ├── roles.go
│   ├── list.go           # GET get_roles
│   ├── get.go            # GET get_role/{id}
│   ├── list_test.go
│   └── get_test.go
├── templates/
│   ├── templates.go
│   ├── list.go           # GET get_templates
│   └── list_test.go
├── get/
│   └── sections.go       # GET get_section, get_sections
├── tests/
│   ├── get.go            # GET get_test
│   ├── list.go           # GET get_tests
│   ├── get_test.go
│   └── list_test.go
└── groups/
    ├── add.go            # POST add_group
    ├── update.go         # POST update_group
    ├── delete.go         # POST delete_group
    └── *_test.go
```

#### 🚦 Критерии завершения Stage 5

- [ ] Все команды из выбранного scope реализованы
- [ ] Тесты для всех новых команд (покрытие как в milestones)
- [ ] Команды зарегистрированы в `commands.go`
- [ ] Обновлена документация (README.md, docs/)
- [ ] CHANGELOG.md обновлён
- [ ] Все тесты проходят (`go test ./...`)

---

### ЭТАП 6: Performance Optimization & UX Enhancement 🔄 (В работе)

**Период:** 2026-02-16 — TBD  
**Статус:** 🔄 В работе (Phases 6.1-6.2, 6.4, 6.6 ✅)  
**Версия:** 2.7.0 → 2.8.0-dev

**Цель:** Устранение "зависаний" и улучшение производительности на 60-80%

#### Задачи

**Phase 6.1: Progress Bars Foundation** ✅ (Завершено)
- [x] Интеграция `github.com/vbauerster/mpb/v8` (вместо progressbar/v3)
- [x] Создать `internal/progress` пакет
  - `Manager` — управление прогресс-барами
  - `Bar` — индивидуальные прогресс-бары с ETA
  - `Monitor` — channel-based мониторинг
  - `AsyncProgress` — асинхронные обновления
  - `ProgressTracker` — multi-phase прогресс
- [x] Добавить progress bar в `compare` команды
- [x] Добавить progress bar в `get` команды с большими датасетами

**Phase 6.2: Parallel API Requests** ✅ (Завершено)
- [x] Создать `internal/concurrent` пакет
- [x] Worker pool pattern (`WorkerPool` с `errgroup`)
- [x] Rate limiter (150 req/min для TestRail, token bucket)
- [x] Adaptive Rate Limiter (подстройка по response time)
- [x] Parallel fetching (`ParallelMap`, `ParallelForEach`)
- [x] Batch Processor (пакетная обработка с retry)
- [x] Graceful error handling при параллельных запросах

**Phase 6.3: Response Caching** ⏳ (Запланировано)
- [ ] Создать `internal/cache` пакет
- [ ] Disk-based cache с TTL
- [ ] Cache invalidation при write операциях
- [ ] Флаг `--no-cache` для обхода кэша
- [ ] Команда `gotr cache clear`

**Phase 6.4: Retry Logic & Resilience** ✅ (Завершено)
- [x] Exponential backoff retry (1s, 2s, 4s, 8s, 16s, max 30s)
- [x] Circuit breaker pattern (Closed/Open/Half-Open)
- [x] Retry with context cancellation support
- [x] Configurable retry policies
- [ ] Флаг `--timeout` (default: 5min) — перенесено в 6.7

**Phase 6.5: Batch Operations Optimization** 🔄 (Частично)
- [x] Batch fetching (250 items per request) — `BatchProcessor`
- [ ] Prefetching связанных сущностей
- [ ] Оптимизация памяти для больших сравнений
- [ ] Streaming output для больших датасетов

**Phase 6.6: UX Polish** ✅ (Завершено)
- [x] ETA расчет в progress bars (EWMA-based)
- [x] Цветной вывод (эмодзи)
- [x] Флаг `--quiet` для CI/CD
- [x] Флаг `--debug` для детального логирования

**Phase 6.7: Recursive Parallelization** 🔄 (В работе — приоритет)
- [ ] Создать `ParallelController` для оркестрации запросов
- [ ] Создать `ResultAggregator` для сбора результатов
- [ ] Priority queue (большие сьюты первыми)
- [ ] Параллельная пагинация внутри каждого сьюта
- [ ] Контроллер для отслеживания всех горутин
- [ ] Гарантированная целостность данных
- [ ] Unit + Integration тесты
- [ ] Performance benchmarks
- [ ] Флаг `--timeout` (default: 5min)

**Детали реализации:**
- Максимум 20 concurrent requests
- Adaptive rate limiting (снижение при 429)
- Priority queue (большие сьюты первыми)
- Graceful error handling с retry
- Потоковая передача результатов

**Целевое время**: < 5 минут для 36k+ cases (сейчас ~12 мин)

**Документация**: [docs/recursive-parallelization-plan.md](../docs/recursive-parallelization-plan.md)

#### Целевые метрики

| Метрика | Текущее | Цель |
|---------|---------|------|
| `compare cases` (1000 items) | 5+ min | **<30 sec** (-90%) |
| `compare all` | 10+ min | **<2 min** (-80%) |
| Пиковая память | 1GB+ | **<500MB** |
| "Зависания" | Минуты | **Нет** |

#### Новые зависимости

```go
github.com/schollz/progressbar/v3 v3.14.1  // Progress bars
golang.org/x/sync v0.6.0                    // errgroup для параллельности
```

#### Новые пакеты

```
internal/
├── progress/           # Progress bar interface
│   ├── progress.go
│   ├── spinner.go
│   └── bar.go
├── cache/              # Caching layer
│   ├── cache.go
│   ├── disk.go
│   └── ttl.go
└── concurrent/         # Concurrency utilities
    ├── pool.go
    ├── limiter.go
    └── retry.go
```

#### Критерии завершения

- [ ] Все 6 фаз завершены
- [ ] Производительность улучшена на 60-80%
- [ ] 95%+ покрытие тестами новых пакетов
- [ ] Документация обновлена
- [ ] Бенчмарки добавлены в CI
- [ ] Нет регрессий в существующей функциональности

---

### ЭТАП 7: Общий рефакторинг и стандартизация ⏳ (Запланирован)

**Период:** TBD (после Stage 6)  
**Статус:** ⏳ Запланирован  
**Версия:** 2.8.0 → 3.0.0

**Цель:** Приведение всей кодовой базы к единым стандартам и архитектурным паттернам

#### Задачи

**Code Organization:**
- [ ] Унификация структуры всех команд (одинаковые паттерны для всех cmd/)
- [ ] Выделение общих компонентов в internal/
- [ ] Удаление дублирующегося кода
- [ ] Стандартизация именования файлов и пакетов

**API Consistency:**
- [ ] Единый подход к обработке ошибок во всех командах
- [ ] Стандартизация флагов (имена, дефолты, описания)
- [ ] Унификация вывода (форматы, цвета, структура)
- [ ] Единый паттерн для интерактивного режима

**Testing:**
- [ ] Доведение покрытия тестами до 90%+ во всех пакетах
- [ ] Унификация тестовых паттернов
- [ ] Создание общих тестовых хелперов
- [ ] Интеграционные тесты для критических путей

**Documentation:**
- [ ] Обновление всех README и документации
- [ ] Примеры использования для всех команд
- [ ] Документация архитектурных решений

#### Критерии завершения
- [ ] Все команды следуют единым паттернам
- [ ] Нет дублирующегося кода
- [ ] 90%+ покрытие тестами
- [ ] Документация актуальна
- [ ] Готовность к релизу 3.0.0

---

## 📊 Матрица покрытия API (актуально на 2026-02-11)

### Client API (полностью реализовано)

| Категория | Всего | Client | Процент | Статус |
|-----------|-------|--------|---------|--------|
| Cases | 10 | 10 | 100% | 🟢 |
| Suites | 5 | 5 | 100% | 🟢 |
| Sections | 5 | 5 | 100% | 🟢 |
| Shared Steps | 6 | 6 | 100% | 🟢 |
| Runs | 6 | 6 | 100% | 🟢 |
| Results | 7 | 7 | 100% | 🟢 |
| Tests | 3 | 3 | 100% | 🟢 |
| Milestones | 5 | 5 | 100% | 🟢 |
| Plans | 9 | 9 | 100% | 🟢 |
| Attachments | 5 | 5 | 100% | 🟢 |
| Configurations | 7 | 7 | 100% | 🟢 |
| Users | 3 | 3 | 100% | 🟢 |
| Projects | 5 | 5 | 100% | 🟢 |
| Priorities | 1 | 1 | 100% | 🟢 |
| Statuses | 1 | 1 | 100% | 🟢 |
| Templates | 1 | 1 | 100% | 🟢 |
| CaseFields | 2 | 2 | 100% | 🟢 |
| CaseTypes | 1 | 1 | 100% | 🟢 |
| ResultFields | 1 | 1 | 100% | 🟢 |
| Reports | 3 | 3 | 100% | 🟢 |
| Roles | 2 | 2 | 100% | 🟢 |
| Groups | 5 | 5 | 100% | 🟢 |
| Datasets | 5 | 5 | 100% | 🟢 |
| Variables | 4 | 4 | 100% | 🟢 |
| Labels | 2 | 2 | 100% | 🟢 |
| BDDs | 2 | 2 | 100% | 🟢 |
| **ИТОГО** | **106** | **106** | **100%** | 🟢 |

### CLI Coverage (Stage 5 - в процессе)

| Категория | Всего | CLI | Процент | Статус | Stage 5 Приоритет |
|-----------|-------|-----|---------|--------|-------------------|
| Cases | 10 | 10 | 100% | 🟢 | — |
| Suites | 5 | 5 | 100% | 🟢 | — |
| Sections | 5 | 3 | 60% | 🟡 | 🔴 P0 |
| Shared Steps | 6 | 6 | 100% | 🟢 | — |
| Runs | 6 | 6 | 100% | 🟢 | — |
| Results | 7 | 7 | 100% | 🟢 | — |
| Tests | 3 | 1 | 33% | 🟡 | 🟡 P1 |
| Milestones | 5 | 5 | 100% | 🟢 | — |
| Plans | 9 | 9 | 100% | 🟢 | — |
| Attachments | 5 | 5 | 100% | 🟢 | — |
| Configurations | 7 | 7 | 100% | 🟢 | — |
| Users | 3 | 0 | 0% | ❌ | 🔴 P0 |
| Projects | 5 | 2 | 40% | 🟡 | 🟢 P2 (admin) |
| Priorities | 1 | 0 | 0% | ❌ | 🔴 P0 |
| Statuses | 1 | 0 | 0% | ❌ | 🔴 P0 |
| Templates | 1 | 0 | 0% | ❌ | 🔴 P0 |
| CaseFields | 2 | 1 | 50% | 🟡 | 🟢 P2 |
| CaseTypes | 1 | 1 | 100% | 🟢 | — |
| ResultFields | 1 | 0 | 0% | ❌ | 🟡 P1 |
| Reports | 3 | 0 | 0% | ❌ | 🟢 P2 |
| Roles | 2 | 0 | 0% | ❌ | 🟡 P1 |
| Groups | 5 | 2 | 40% | 🟡 | 🟡 P1 |
| Datasets | 5 | 5 | 100% | 🟢 | — |
| Variables | 4 | 4 | 100% | 🟢 | — |
| Labels | 2 | 2 | 100% | 🟢 | — |
| BDDs | 2 | 2 | 100% | 🟢 | — |
| **ИТОГО** | **106** | **~78** | **~74%** | 🟡 | Stage 5 |

---

## 🎯 Приоритеты разработки Stage 5

### 🔴 P0 — Высокий (основные справочники)
| Ресурс | Endpoints | CLI команды | Почему важно |
|--------|-----------|-------------|--------------|
| Users | 3 | `users list/get/get-by-email` | Назначение тестов |
| Priorities | 1 | `get priorities` | Создание кейсов |
| Statuses | 1 | `get statuses` | Результаты тестов |
| Templates | 1 | `templates list` | Создание кейсов |
| Sections | 2 | `get section/sections` | Структура проекта |

### 🟡 P1 — Средний (функциональность)
| Ресурс | Endpoints | CLI команды | Почему важно |
|--------|-----------|-------------|--------------|
| Tests | 2 | `tests get/list` | Инфо о тестах |
| Groups | 3 | `groups add/update/delete` | Управление группами |
| Roles | 2 | `roles list/get` | Администрирование |
| ResultFields | 1 | `get result-fields` | Кастомизация |

### 🟢 P2 — Низкий (специфичное)
| Ресурс | Endpoints | CLI команды | Почему важно |
|--------|-----------|-------------|--------------|
| Reports | 3 | `reports list/run/*` | Отчёты |
| CaseFields POST | 1 | `add case-field` | Админская функция |
| Projects POST | 3 | `add/update/delete project` | Админская функция |

---

## 📝 Правила работы с этим файлом

### Обновление плана:

1. **ТОЛЬКО после аудита** — сначала API_AUDIT.md
2. **Синхронизация** — все изменения согласованы с API_ANALYSIS.md
3. **Версионирование** — при значительных изменениях обновлять версию плана
4. **Комментарии** — пояснять причины изменений

### Добавление этапов:

1. НЕ создавать отдельные файлы (STAGE_X.md)
2. Добавлять секцию в этот файл
3. Следовать шаблону этапа
4. Получать утверждение перед началом

---

## 🔗 Связанные документы

- `API_AUDIT.md` — Общий аудит проекта
- `API_ANALYSIS.md` — Сверка с официальной документацией
- `CHECKLIST.md` — Чеклисты текущего этапа
- `AGENTS.md` — Руководство для AI
- `ARCHITECTURE.md` — Архитектурная документация
- `../CHANGELOG.md` — История изменений

---

*Файл обновлён: 2026-02-11*  
*Версия плана: 2.1*  
*Алгоритм: API_AUDIT → API_ANALYSIS → PLAN → CHECKLIST → Code*  
*Статус: Stage 5 — ожидает утверждения плана*  
*Следующий шаг: Получить утверждение на Этап 3*
