# Команда: get

Language: Русский | [English](../../../en/guides/commands/get.md)

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
Выполняет GET-запросы к TestRail API.
Подкоманды:

## Синтаксис
```bash
gotr get [command]
```

## Подкоманды

| Подкоманда | Описание |
| --- | --- |
| `case` | Получить один кейс по ID кейса |
| `case-fields` | Получить список полей кейсов |
| `case-history` | Получить историю изменений кейса по ID кейса |
| `case-types` | Получить список типов кейсов |
| `cases` | Получить кейсы проекта |
| `project` | Получить один проект по ID проекта |
| `projects` | Получить все projects |
| `sharedstep` | Получить один shared step по ID шага |
| `sharedsteps` | Получить shared steps проекта |
| `suite` | Получить одну тест-сюиту по ID сюиты |
| `suites` | Получить тест-сюиты проекта |

## Флаги

```text
-h, --help   справка для get
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
gotr get --help
gotr get case --help
```

## Источник

- Данные разделов выше сформированы из фактического вывода `--help` текущего кода CLI.

---

← [Команды](index.md) · [Гайды](../index.md) · [Документация](../../index.md)
