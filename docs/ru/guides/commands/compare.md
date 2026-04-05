# Команда: compare

Language: Русский | [English](../../../en/guides/commands/compare.md)

## Навигация

- [Документация](../../index.md)
  - [Гайды](../index.md)
    - [Установка](../installation.md)
    - [Конфигурация](../configuration.md)
    - [Интерактивный режим](../interactive-mode.md)
    - [Прогресс](../progress.md)
    - [Каталог команд](index.md)
      - [Общие](global-flags.md)
      - [CRUD операции](add.md)
      - [Основные ресурсы](get.md)
        - [get](get.md)
        - [sync](sync.md)
        - [compare](compare.md)
        - [cases](cases.md)
        - [run](run.md)
        - [result](result.md)
        - [test](test.md)
        - [tests](tests.md)
        - [attachments](attachments.md)
        - [plans](plans.md)
        - [reports](reports.md)
      - [Специальные ресурсы](bdds.md)
  - [Архитектура](../../architecture/index.md)
  - [Эксплуатация](../../operations/index.md)
  - [Отчёты](../../reports/index.md)
- [Главная](../../../../README_ru.md)


## Обзор
Выполнение сравнения ресурсов между двумя проектами.
Поддерживаемые ресурсы:

## Синтаксис
```bash
gotr compare [command]
```

## Подкоманды

| Подкоманда | Описание |
| --- | --- |
| `all` | Compare all resources between two projects |
| `cases` | Compare test cases between projects |
| `configurations` | Сравнить конфигурации между проектами |
| `datasets` | Сравнить datasets между проектами |
| `groups` | Сравнить группы между проектами |
| `labels` | Сравнить метки между проектами |
| `milestones` | Сравнить milestones между проектами |
| `plans` | Сравнить test plans между проектами |
| `runs` | Сравнить test runs между проектами |
| `sections` | Compare sections between projects |
| `sharedsteps` | Сравнить shared steps между проектами |
| `suites` | Compare test suites between projects |
| `templates` | Сравнить шаблоны между проектами |

## Флаги

```text
-h, --help               справка для compare
--page-retries int       Количество retry для каждой страницы в основном этапе загрузки (default 5)
--parallel-pages int     Максимальное количество параллельных страниц внутри сьюта (default 6)
--parallel-suites int    Максимальное количество параллельных сьютов (default 10)
-1, --pid1 string        ID первого проекта (обязательно)
-2, --pid2 string        ID второго проекта (обязательно)
--rate-limit int         Лимит API-запросов в минуту. -1 = авто по profile/deployment, 0 = без лимита, >0 = фиксированное значение. (default -1)
--retry-attempts int     Количество попыток при точечном авто-ретрае failed pages (default 5)
--retry-delay duration   Пауза между попытками одной страницы при авто-ретрае (default 200ms)
--retry-workers int      Количество параллельных воркеров при авто-ретрае failed pages (default 12)
--save                   Сохранить результат в файл (по умолчанию в ~/.gotr/exports/)
--save-to string         Сохранить результат в указанный файл
--timeout duration       Таймаут для операции сравнения (default 30m0s)
```

## Глобальные флаги

```text
-k, --api-key string    API ключ TestRail
-c, --config            Создать дефолтный файл конфигурации
-f, --format string     Формат вывода: table, json, csv, md, html (default "table")
--insecure              Пропустить проверку TLS сертификата
--non-interactive       Отключить интерактивные подсказки; завершить с ошибкой если требуется ввод
-q, --quiet             Подавить служебный вывод (прогресс, статистику, сообщения о сохранении)
--url string            Базовый URL TestRail
-u, --username string   Email пользователя TestRail
```

## Примеры

```bash
gotr compare --help
gotr compare all --help
```

## Источник

- Данные разделов выше сформированы из фактического вывода `--help` текущего кода CLI.

---

← [Команды](index.md) · [Гайды](../index.md) · [Документация](../../index.md)
