# Smoke-тестирование (snap_smoke)

Language: Русский | [English](../../en/guides/smoke-testing.md)

## Навигация

- [Документация](../index.md)
  - [Гайды](index.md)
    - [Установка](installation.md)
    - [Конфигурация](configuration.md)
    - [Интерактивный режим](interactive-mode.md)
    - [Прогресс](progress.md)
    - [Smoke-тестирование](smoke-testing.md)
    - [Каталог команд](commands/index.md)
    - [Инструкции](instructions/index.md)
  - [Архитектура](../architecture/index.md)
  - [Эксплуатация](../operations/index.md)
  - [Отчёты](../reports/index.md)
- [Главная](../../../README_ru.md)

## Обзор

Пакет `pkg/snap_smoke` предоставляет end-to-end smoke-тесты для функциональности snap/rollback.
Тесты проверяют полный цикл: создание снапшота → мутация данных → откат → верификация результата.

## Архитектура

```
pkg/snap_smoke/
├── doc.go           — документация пакета
├── config.go        — загрузка конфигурации из env-переменных, создание клиента
├── helpers.go       — хелперы: testCase (с автоочисткой) и testSection (переиспользование)
├── testserver.go    — FakeTestRail: встроенный in-memory TestRail API v2 сервер
└── smoke_test.go    — 6 E2E тестов за build-тегом //go:build smoke
```

### FakeTestRail

Встроенный mock-сервер на базе `net/http/httptest`. Маршрутизация построена на каталоге
`pkg/testrailapi.APIPath` — URI-шаблоны компилируются в regex автоматически, что гарантирует
соответствие мока реальной структуре API. Эмулирует подмножество TestRail API v2:

| Метод | Эндпоинт | Описание |
| ----- | -------- | -------- |
| GET | `get_case/{case_id}` | Получить кейс |
| POST | `add_case/{section_id}` | Создать кейс |
| POST | `update_case/{case_id}` | Обновить кейс (partial update) |
| POST | `delete_case/{case_id}` | Удалить кейс |
| GET | `get_sections/{project_id}` | Список секций (paginated wrapper) |
| POST | `add_section/{project_id}` | Создать секцию |

Особенности:
- **Роуты из testrailapi** — URI-шаблоны из `testrailapi.New().Paths()` компилируются в regex (path) + query-template map. Если в `testrailapi` меняется URI — мок обновляется автоматически.
- **Paginated responses** — `get_sections` возвращает ответ в формате TestRail 6.7+: `{"offset":0, "limit":250, "size":N, "_links":{...}, "sections":[...]}`. Клиентский `fetchAllPages` проходит полный цикл парсинга.
- **Хранение** — `sync.Mutex` + `map[int64]*Case/Section`. ID автоинкрементные (cases: 1000+, sections: 100+).

## Два режима работы

### 1. Встроенный mock (по умолчанию)

Не требует настройки. FakeTestRail поднимается автоматически:

```bash
go test -tags smoke ./pkg/snap_smoke/ -v
```

Преимущества:
- Без внешних зависимостей
- Быстро (~0.03 секунды)
- Подходит для CI/CD

### 2. Реальный сервер

Для проверки совместимости с настоящим TestRail задайте переменные окружения:

```bash
export GOTR_SMOKE_URL=http://localhost:8080
export GOTR_SMOKE_USER=admin@example.com
export GOTR_SMOKE_KEY=yourkey
export GOTR_SMOKE_PROJECT=3
export GOTR_SMOKE_SUITE=1
export GOTR_SMOKE_INSECURE=true   # для самоподписанных сертификатов

go test -tags smoke ./pkg/snap_smoke/ -v
```

| Переменная | Обязательна | Описание |
| ---------- | ----------- | -------- |
| `GOTR_SMOKE_URL` | Да | Базовый URL TestRail |
| `GOTR_SMOKE_USER` | Да | Логин / email |
| `GOTR_SMOKE_KEY` | Да | API-ключ |
| `GOTR_SMOKE_PROJECT` | Да | ID проекта |
| `GOTR_SMOKE_SUITE` | Нет¹ | ID сьюта (для multi-suite проектов) |
| `GOTR_SMOKE_INSECURE` | Нет | Пропустить проверку TLS (`true`/`1`) |

¹ Обязательна для проектов в режиме multi-suite.

## Тестовые сценарии

| # | Тест | Операция | Что проверяет |
| - | ---- | -------- | ------------- |
| 1 | `TestSmoke_UpdateRollback` | update → rollback | Мутация title+priority → откат → оригинал восстановлен |
| 2 | `TestSmoke_DeleteRollback` | delete → rollback | Удаление → откат → re-create с новым ID (Tier 2) |
| 3 | `TestSmoke_AddRollback` | add → rollback | Создание → FinalizeAdd → откат → кейс удалён |
| 4 | `TestSmoke_SnapManagementCycle` | list/info/delete | Custom name → список → инфо → удаление снапшота |
| 5 | `TestSmoke_DoubleRollbackBlocked` | double rollback | Повторный rollback → ошибка "rolled_back" |
| 6 | `TestSmoke_GCOrphans` | gc | Orphan удалён, tracked снапшот сохранён |

## Расширение

### Добавление нового теста

1. Добавьте функцию `TestSmoke_*` в `smoke_test.go`
2. Используйте хелперы `testCase()` и `testSection()` — они регистрируют автоочистку
3. Изолируйте snap-хранилище через `t.Setenv("HOME", t.TempDir())`

### Расширение FakeTestRail

Для поддержки новых эндпоинтов:

1. Добавьте хранилище (map) в `FakeTestRail`
2. Добавьте handler-метод `handleXxx(w, r, params)` — параметры извлекаются из regex-групп и query templates
3. Зарегистрируйте action → handler в `handlers` map внутри `NewFakeTestRail()`
4. Маршрут подхватится автоматически из `testrailapi.APIPath` при совпадении action name
5. Для list-эндпоинтов используйте `writePaginatedJSON()` для корректного paginated wrapper

### Build-тег

Все файлы имеют тег `//go:build smoke`. Тесты не попадают в обычный `go test ./...`.

## CLI Smoke-тесты (cmd/snap)

Помимо E2E-тестов в `pkg/snap_smoke`, существуют CLI-level smoke-тесты в `cmd/snap/smoke_test.go`.
Они проверяют корректность Cobra-команд `gotr snap *` без реального API — используя mock-клиент
и изолированное snap-хранилище.

### Запуск

```bash
go test -tags smoke ./cmd/snap/ -v
```

### Тестовые сценарии

| # | Тест | Что проверяет |
| - | ---- | ------------- |
| 1 | `TestCLI_SnapList_Empty` | Пустой список при отсутствии снапшотов |
| 2 | `TestCLI_SnapList_WithEntries` | Табличный вывод с данными |
| 3 | `TestCLI_SnapInfo` | JSON-вывод метаданных |
| 4 | `TestCLI_SnapInfo_NotFound` | Обработка ошибки "не найден" |
| 5 | `TestCLI_SnapRollback_Update` | Откат update-операции |
| 6 | `TestCLI_SnapRollback_Delete` | Откат delete с re-create (Tier 2) |
| 7 | `TestCLI_SnapRollback_Add` | Откат add-операции |
| 8 | `TestCLI_SnapRollback_NotFound` | Обработка ошибки при отсутствии снапшота |
| 9 | `TestCLI_SnapRollback_AlreadyRolledBack` | Защита от двойного rollback |
| 10 | `TestCLI_SnapDelete` | Удаление снапшота через CLI |
| 11 | `TestCLI_SnapGC_NoOrphans` | GC при отсутствии orphans |
| 12 | `TestCLI_SnapGC_CleansOrphans` | GC удаляет orphan-директории |
| 13 | `TestCLI_FullCycle_ListRollbackList` | Полный цикл list → rollback → list с проверкой статуса |

### Отличия от pkg/snap_smoke

| Аспект | `pkg/snap_smoke` | `cmd/snap/smoke_test.go` |
| ------ | ---------------- | ------------------------ |
| Уровень | Engine-level (API пакет) | CLI-level (Cobra команды) |
| Сервер | FakeTestRail (httptest) | Mock-клиент (без HTTP) |
| Фокус | Корректность snap/rollback логики | Корректность CLI-обёрток |
| Скорость | ~0.03 с | ~0.01 с |
| Когда использовать | Проверка snap-движка | Проверка snap-команд |

### Полный прогон обоих уровней

```bash
go test -tags smoke ./pkg/snap_smoke/ ./cmd/snap/ -v
```

---

← [Гайды](index.md)
