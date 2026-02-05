# Архитектура gotr

> Общее описание архитектуры CLI-утилиты gotr для пользователей  
> **Важно:** Этот файл актуализируется при добавлении новых команд или изменении структуры проекта. Последнее обновление: 2026-02-05 (v2.5.0).

## Что такое gotr

`gotr` — это CLI-клиент для TestRail API v2, построенный по многослойной архитектуре с чётким разделением ответственности между слоями.

## Общая схема

```
┌─────────────────────────────────────────────────────────────┐
│  CLI Layer (cmd/*)                                          │
│  • Парсинг аргументов и флагов                              │
│  • Интерактивный выбор (internal/interactive)               │
│  • Вызов сервисов                                           │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│  Service Layer (internal/service/*)                         │
│  • Бизнес-логика                                            │
│  • Валидация данных                                         │
│  • Миграция данных (migration)                              │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│  Client Layer (internal/client/*)                           │
│  • HTTPClient — реальный клиент                             │
│  • ClientInterface — абстракция для тестов                  │
│  • MockClient — реализация для тестирования                 │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│  TestRail API v2                                            │
└─────────────────────────────────────────────────────────────┘
```

## Слои подробно

### 1. CLI Layer (`cmd/`)

**Ответственность:** Принимает команды от пользователя, парсит аргументы, вызывает сервисы.

**Структура:**
```
cmd/
├── common/              # Общие компоненты
│   ├── client.go       #   ClientAccessor — единый доступ к HTTP клиенту
│   └── flags.go        #   Парсинг общих флагов
├── get/                # GET команды
├── result/             # Команды для работы с результатами
├── run/                # Команды для работы с test runs
├── sync/               # Команды миграции данных
├── root.go             # Корневая команда
└── commands.go         # Регистрация всех команд
```

**Пример:**
```bash
gotr run get 12345 --jq
# cmd/run/get.go → RunService.Get(12345) → вывод с jq-фильтром
```

### 2. Service Layer (`internal/service/`)

**Ответственность:** Бизнес-логика, валидация, оркестрация операций.

**Компоненты:**
- `RunService` — работа с test runs
- `ResultService` — работа с результатами тестов
- `migration/` — миграция данных между проектами
  - `types.go` — контекст миграции
  - `fetch.go` — получение данных
  - `filter.go` — фильтрация дубликатов
  - `import.go` — импорт сущностей
  - `export.go` — экспорт данных
  - `mapping.go` — управление mapping ID

**Валидация:**
```go
// Проверки перед созданием run:
// - projectID > 0
// - name не пустое
// - suite_id > 0 (если указан)
```

### 3. Client Layer (`internal/client/`)

**Ответственность:** HTTP-запросы к TestRail API.

**Структура:**
```
internal/client/
├── client.go           # HTTPClient — основной HTTP клиент
├── interfaces.go       # ClientInterface + 7 API групп
├── mock.go             # MockClient для тестирования
├── projects.go         # ProjectsAPI (5 методов)
├── cases.go            # CasesAPI (14 методов)
├── suites.go           # SuitesAPI (5 методов)
├── sections.go         # SectionsAPI (5 методов)
├── sharedsteps.go      # SharedStepsAPI (6 методов)
├── runs.go             # RunsAPI (6 методов)
└── results.go          # ResultsAPI (7 методов)
```

**ClientInterface:**
- 43 метода общего назначения
- Композиция из 7 интерфейсов по доменам
- Поддержка mock-реализации для тестов

### 4. Interactive Layer (`internal/interactive/`)

**Ответственность:** Интерактивный выбор проектов, сьютов, ранов.

**Использование:**
- `gotr run list` — выбор проекта → список ранов
- `gotr result list` — выбор проекта → выбор рана → результаты
- `gotr get cases` — выбор проекта → выбор сьюта

### 5. Models (`internal/models/data/`)

**Ответственность:** DTO (Data Transfer Objects) для API.

**Основные структуры:**
- `Project`, `Suite`, `Section`, `Case`
- `Run`, `Test`, `Result`
- `SharedStep`
- `Status` — константы статусов

### 6. Utilities (`internal/utils/`)

**Ответственность:** Вспомогательные функции.

**Компоненты:**
- `helpers.go` — парсинг ID, вывод результатов, сохранение в файл
- `log.go` — директории для логов

## Поток данных

### Пример 1: Создание test run

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

### Пример 2: Миграция данных (sync full)

```
Пользователь
    ↓
gotr sync full --src-project 30 --dst-project 31
    ↓
CLI Layer (cmd/sync/sync_full.go)
    ↓
migration.NewMigration(client, 30, 0, 31, 0, "title", logDir)
    ↓
Migration.FetchSharedStepsData()  // Получение данных
    ↓
Migration.FilterSharedSteps()     // Фильтрация дубликатов
    ↓
Migration.ImportSharedSteps()     // Импорт
    ↓
Аналогично для cases
    ↓
TestRail API (src) → Migration → TestRail API (dst)
```

## Полная структура проекта

```
gotr/
├── cmd/                          # CLI команды
│   ├── common/                   #   Общие компоненты
│   │   ├── client.go            #     ClientAccessor
│   │   └── flags.go             #     Парсинг флагов
│   ├── get/                     #   GET команды
│   ├── result/                  #   Команды results
│   ├── run/                     #   Команды runs
│   ├── sync/                    #   Команды миграции
│   ├── root.go                  #   Корневая команда
│   └── commands.go              #   Регистрация команд
├── docs/                         # Документация
│   ├── architecture.md          #   Этот файл
│   ├── configuration.md         #   Настройка
│   ├── get-commands.md          #   GET команды
│   ├── sync-commands.md         #   SYNC команды
│   ├── installation.md          #   Установка
│   ├── interactive-mode.md      #   Интерактивный режим
│   └── other-commands.md        #   Прочие команды
├── embedded/                     # Встроенные утилиты
│   └── jq.go                    #   Встроенный jq
├── internal/                     # Внутренний код
│   ├── client/                  #   HTTP клиент
│   │   ├── client.go           #     HTTPClient
│   │   ├── interfaces.go       #     ClientInterface (43 метода)
│   │   ├── mock.go             #     MockClient
│   │   ├── projects.go         #     ProjectsAPI
│   │   ├── cases.go            #     CasesAPI
│   │   ├── suites.go           #     SuitesAPI
│   │   ├── sections.go         #     SectionsAPI
│   │   ├── sharedsteps.go      #     SharedStepsAPI
│   │   ├── runs.go             #     RunsAPI
│   │   └── results.go          #     ResultsAPI
│   ├── interactive/            #   Интерактивный выбор
│   │   └── selector.go         #     Селекторы проектов/сьютов
│   ├── service/                #   Бизнес-логика
│   │   ├── run.go              #     RunService
│   │   ├── result.go           #     ResultService
│   │   └── migration/          #     Миграция данных
│   │       ├── types.go       #       Migration struct
│   │       ├── fetch.go       #       Получение данных
│   │       ├── filter.go      #       Фильтрация
│   │       ├── import.go      #       Импорт
│   │       ├── export.go      #       Экспорт
│   │       ├── mapping.go     #       Mapping ID
│   │       └── log.go         #       Логирование
│   ├── models/                 #   Модели данных
│   │   └── data/              #     DTO для API
│   │       ├── projects.go    #       Project
│   │       ├── cases.go       #       Case
│   │       ├── suites.go      #       Suite
│   │       ├── sections.go    #       Section
│   │       ├── sharedsteps.go #       SharedStep
│   │       ├── runs.go        #       Run, Test
│   │       ├── results.go     #       Result
│   │       └── statuses.go    #       Status
│   └── utils/                  #   Утилиты
│       ├── helpers.go         #     Вспомогательные функции
│       └── log.go             #     Работа с логами
├── pkg/                          # Публичные пакеты
│   └── testrailapi/            #   Описания API endpoints
│       └── api_paths.go
├── .systems/                     # Системная документация
│   └── ARCHITECTURE.md         #   Детальная архитектура для разработчиков
├── dist/                         # Артефакты сборки (в .gitignore)
├── main.go                       # Точка входа
├── go.mod                        # Go модули
├── Makefile                     # Сборка
└── README.md                    # Основная документация
```

## Доступные команды

### Управление test runs (`gotr run`)
| Команда | Описание |
|---------|----------|
| `gotr run get <id>` | Получить информацию о run |
| `gotr run list [project-id]` | Список runs (интерактивный выбор) |
| `gotr run create <project-id>` | Создать run |
| `gotr run update <id>` | Обновить run |
| `gotr run close <id>` | Закрыть run |
| `gotr run delete <id>` | Удалить run |

### Управление результатами (`gotr result`)
| Команда | Описание |
|---------|----------|
| `gotr result get <test-id>` | Получить результаты test |
| `gotr result get-case <run-id> <case-id>` | Получить результаты case |
| `gotr result list [--run-id <id>]` | Список результатов (интерактивно) |
| `gotr result add <test-id>` | Добавить результат |
| `gotr result add-case <run-id>` | Добавить результат для case |
| `gotr result add-bulk <run-id>` | Массовое добавление из файла |

### Миграция данных (`gotr sync`)
| Команда | Описание |
|---------|----------|
| `gotr sync full` | Полная миграция (shared-steps + cases) |
| `gotr sync cases` | Миграция кейсов |
| `gotr sync shared-steps` | Миграция shared steps |
| `gotr sync suites` | Миграция suites |
| `gotr sync sections` | Миграция sections |

### Получение данных (`gotr get`)
| Команда | Описание |
|---------|----------|
| `gotr get projects` | Все проекты |
| `gotr get project <id>` | Конкретный проект |
| `gotr get suites [project-id]` | Сьютs проекта |
| `gotr get suite <id>` | Конкретный сьют |
| `gotr get cases [project-id]` | Кейсы (интерактивный выбор сьюта) |
| `gotr get case <id>` | Конкретный кейс |
| `gotr get sharedsteps <project-id>` | Shared steps |
| `gotr get sections <project-id>` | Секции |

### Прочие команды
| Команда | Описание |
|---------|----------|
| `gotr add <endpoint>` | POST запросы |
| `gotr update <endpoint>` | UPDATE запросы |
| `gotr delete <endpoint>` | DELETE запросы |
| `gotr list <resource>` | Список API endpoints |
| `gotr export <resource>` | Экспорт данных |
| `gotr import <resource>` | Импорт данных |
| `gotr compare` | Сравнение проектов |
| `gotr config` | Управление конфигурацией |
| `gotr self-test` | Самодиагностика |

## Почему такая архитектура

### Преимущества

1. **Чёткое разделение** — каждый слой знает только про свой уровень
2. **Тестируемость** — можно тестировать сервисы с MockClient без реальных HTTP запросов
3. **Расширяемость** — добавление новой команды не требует изменения client
4. **Переиспользование** — один сервис используется в разных командах
5. **Интерактивность** — единый механизм выбора в `internal/interactive/`

### Добавление retry

Если TestRail возвращает "rate limit", retry добавляется только в Service Layer:

```go
func (s *RunService) Get(id int64) (*data.Run, error) {
    var run *data.Run
    err := retry.Do(3, func() error {
        var err error
        run, err = s.client.GetRun(id)
        return err
    })
    return run, err
}
```

CLI команды не требуют изменений!

## Для разработчиков

Где вносить изменения:

| Задача | Локация |
|--------|---------|
| Новая команда | `cmd/<group>/*.go` |
| Новая валидация | `internal/service/*.go` |
| Новый API метод | `internal/client/*.go` + `interfaces.go` |
| Новая структура данных | `internal/models/data/*.go` |
| Интерактивный выбор | `internal/interactive/selector.go` |

Подробная техническая документация: `.systems/ARCHITECTURE.md`
