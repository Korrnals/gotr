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

---

← [Гайды](index.md)
