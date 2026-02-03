# Архитектура gotr

> Общее описание архитектуры CLI-утилиты gotr для пользователей  
> **Важно:** Этот файл актуализируется при добавлении новых команд или изменении структуры проекта. Последнее обновление: 2026-02-04.

## Что такое gotr

`gotr` — это CLI-клиент для TestRail API v2, построенный по многослойной архитектуре. Это означает, что код организован в отдельные слои, каждый из которых отвечает за свою задачу.

## Общая схема

```
┌─────────────────────────────────────┐
│  Пользователь → CLI команды         │
│  (cmd/*)                            │
│  Парсинг аргументов, флагов         │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│  Service Layer                      │
│  (internal/service/*)               │
│  Бизнес-логика, валидация           │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│  Client Layer                       │
│  (internal/client/*)                │
│  HTTP запросы к TestRail API        │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│  TestRail API v2                    │
└─────────────────────────────────────┘
```

## Слои подробно

### 1. CLI Layer (`cmd/`)

**Что делает:** Принимает команды от пользователя и показывает результаты

**Пример:**
```bash
gotr run get 12345
```

Здесь `run` — группа команд, `get` — команда, `12345` — аргумент

**Где находится:**
- `cmd/run/` — команды для работы с test runs
- `cmd/result/` — команды для работы с результатами
- `cmd/sync/` — команды для миграции данных
- `cmd/get/` — GET-запросы к API

### 2. Service Layer (`internal/service/`)

**Что делает:** Проверяет корректность данных и управляет бизнес-логикой

**Пример:** Перед созданием run проверяет, что:
- ID проекта > 0
- Название не пустое
- Suite ID указан корректно

**Сервисы:**
- `RunService` — работа с test runs
- `ResultService` — работа с результатами тестов
- `migration` — миграция данных между проектами

### 3. Client Layer (`internal/client/`)

**Что делает:** Отправляет HTTP запросы к TestRail API

**API методы:**
- `GetRun`, `AddRun`, `UpdateRun`, `CloseRun`, `DeleteRun`
- `AddResult`, `GetResults`
- `GetCases`, `AddCase`
- `GetSuites`, `AddSuite`

### 4. Models (`internal/models/data/`)

**Что делает:** Описывает структуры данных (DTO)

**Примеры структур:**
- `Run` — test run
- `Result` — результат теста
- `Case` — тест-кейс
- `Suite` — тест-сюита

## Поток данных

### Пример: Создание test run

```
Пользователь
    ↓
gotr run create 30 --suite-id 100 --name "Smoke"
    ↓
CLI Layer (cmd/run/create.go)
    ↓
RunService.Create(projectID=30, req={suite_id:100, name:"Smoke"})
    ↓
Валидация: projectID>0? suite_id>0? name не пустое?
    ↓
HTTPClient.AddRun(30, req)
    ↓
POST /index.php?/api/v2/add_run/30
    ↓
TestRail API
```

### Пример: Добавление результата

```
Пользователь
    ↓
gotr result add 12345 --status-id 1 --comment "Passed"
    ↓
CLI Layer (cmd/result/add.go)
    ↓
ResultService.AddForTest(testID=12345, req={status_id:1, ...})
    ↓
Валидация: testID>0? status_id>0?
    ↓
HTTPClient.AddResult(12345, req)
    ↓
POST /index.php?/api/v2/add_result/12345
    ↓
TestRail API
```

## Директории проекта

```
gotr/
├── cmd/                    # CLI команды
│   ├── run/               #   Управление runs
│   ├── result/            #   Управление results
│   ├── sync/              #   Миграция данных
│   └── get/               #   GET запросы
├── internal/              # Внутренний код
│   ├── service/           #   Бизнес-логика
│   │   ├── run.go
│   │   ├── result.go
│   │   └── migration/
│   ├── client/            #   HTTP клиент
│   │   ├── runs.go
│   │   ├── results.go
│   │   └── ...
│   ├── models/            #   Модели данных
│   │   └── data/
│   │       ├── runs.go
│   │       ├── results.go
│   │       └── ...
│   └── utils/             #   Утилиты
├── pkg/                   # Публичные пакеты
│   └── testrailapi/       #   Описания API endpoint'ов
├── docs/                  # Документация
└── .systems/              # Системные документы разработки
```

## Доступные команды

### Управление test runs (`gotr run`)
- `gotr run get <id>` — получить информацию о run
- `gotr run list <project-id>` — список runs проекта
- `gotr run create <project-id>` — создать run
- `gotr run update <id>` — обновить run
- `gotr run close <id>` — закрыть run
- `gotr run delete <id>` — удалить run

### Управление результатами (`gotr result`)
- `gotr result get <test-id>` — получить результаты test
- `gotr result get-case <run-id> <case-id>` — получить результаты case
- `gotr result add <test-id>` — добавить результат
- `gotr result add-case <run-id>` — добавить результат для case
- `gotr result add-bulk <run-id>` — массовое добавление из файла

### Миграция данных (`gotr sync`)
- `gotr sync full` — полная миграция (suites → sections → shared-steps → cases)
- `gotr sync cases` — миграция кейсов
- `gotr sync shared-steps` — миграция shared steps
- `gotr sync suites` — миграция suites
- `gotr sync sections` — миграция sections

### Получение данных (`gotr get`)
- `gotr get case <id>` — получить кейс
- `gotr get cases <project-id>` — получить кейсы проекта
- `gotr get suite <id>` — получить сьюту
- `gotr get suites <project-id>` — получить сьюты
- `gotr get project <id>` — получить проект
- `gotr get projects` — получить все проекты
- `gotr get sharedstep <id>` — получить shared step
- `gotr get sharedsteps <project-id>` — получить shared steps

### Прочие команды
- `gotr add <endpoint>` — POST запросы
- `gotr update <endpoint>` — UPDATE запросы
- `gotr delete <endpoint>` — DELETE запросы
- `gotr list <resource>` — список API endpoints
- `gotr export <resource>` — экспорт данных
- `gotr import <resource>` — импорт данных
- `gotr compare` — сравнение проектов
- `gotr config` — управление конфигурацией

## Почему такая архитектура

### Преимущества

1. **Чёткое разделение** — каждый слой знает только про свой уровень
2. **Легко тестировать** — можно тестировать сервисы без реальных HTTP запросов
3. **Легко расширять** — добавление новой команды не требует изменения client
4. **Переиспользование** — один сервис используется в разных командах

### Что если нужно добавить retry?

Если TestRail начнёт возвращать ошибки "rate limit", retry логику добавляется только в Service Layer, не затрагивая CLI команды:

```go
// Service Layer
func (s *RunService) Get(id int64) (*data.Run, error) {
    // Добавляем retry здесь
    var run *data.Run
    err := retry.Do(3, func() error {
        var err error
        run, err = s.client.GetRun(id)
        return err
    })
    return run, err
}
```

CLI команды даже не заметят изменений!

## Для разработчиков

Если вы хотите внести изменения:

- **Новая команда** → создавайте в `cmd/`
- **Новая валидация** → добавляйте в `internal/service/`
- **Новый API метод** → добавляйте в `internal/client/`
- **Новая структура данных** → добавляйте в `internal/models/data/`

Подробная техническая документация находится в `.systems/ARCHITECTURE.md`
