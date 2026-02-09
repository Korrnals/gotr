# Команды GET

Команды для получения данных из TestRail.

## Получение проектов

```bash
# Все проекты
gotr get projects

# Конкретный проект
gotr get project 30
```

## Получение сьютов

```bash
# Сьютs проекта (интерактивный выбор проекта, если не указан)
gotr get suites
gotr get suites 30
gotr get suites --project-id 30

# Конкретный сьют
gotr get suite 20069
```

## Получение кейсов

### Интерактивный режим

```bash
# Полностью интерактивно
gotr get cases

# С указанием проекта
gotr get cases 30
```

При запуске без `--suite-id`:

1. Если в проекте **один сьют** — используется автоматически
2. Если **несколько сьютов** — предлагается выбор из списка

### Явное указание параметров

```bash
# С конкретным сьютом
gotr get cases 30 --suite-id 20069

# Все кейсы из всех сьютов проекта
gotr get cases 30 --all-suites

# С фильтрацией по секции
gotr get cases 30 --suite-id 20069 --section-id 100
```

### Получение одного кейса

```bash
gotr get case 12345
```

## Получение shared steps

```bash
# Все shared steps проекта (интерактивно)
gotr get sharedsteps
gotr get sharedsteps 30

# Конкретный shared step
gotr get sharedstep 45678
```

## Дополнительные команды

```bash
# Типы кейсов
gotr get case-types

# Поля кейсов
gotr get case-fields

# История изменений кейса
gotr get case-history 12345

# История shared step
gotr get sharedstep-history 45678
```

## Общие флаги

| Флаг | Описание |
|------|----------|
| `-t, --type` | Формат вывода: `json`, `json-full`, `table` |
| `-o, --output` | Сохранить в файл |
| `-q, --quiet` | Тихий режим (без вывода) |
| `-j, --jq` | Включить jq-форматирование |
| `--jq-filter` | jq фильтр |
| `-b, --body-only` | Сохранить только тело ответа |

## Примеры с jq

```bash
# Фильтрация вывода
gotr get projects --jq --jq-filter '.[] | {id: .id, name: .name}'

# Красивый вывод
gotr get case 12345 --jq
```

---

## Примечание: Полное покрытие API

Все 106 endpoint'ов TestRail API v2 реализованы в Client Layer (`internal/client/`).
CLI команды для новых API (milestones, plans, tests, configurations, users, reports, etc.)
будут добавлены в **Stage 5**.
