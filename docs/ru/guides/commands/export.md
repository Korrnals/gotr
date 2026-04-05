# Команда: export

Language: Русский | [English](../../../en/guides/commands/export.md)

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
        - [add](add.md)
        - [delete](delete.md)
        - [update](update.md)
        - [list](list.md)
        - [export](export.md)
      - [Основные ресурсы](get.md)
      - [Специальные ресурсы](bdds.md)
  - [Архитектура](../../architecture/index.md)
  - [Эксплуатация](../../operations/index.md)
  - [Отчёты](../../reports/index.md)
- [Главная](../../../../README_ru.md)


## Обзор
Экспортирует данные из TestRail в JSON-файл.
Имя файла для сохранения:

## Синтаксис
```bash
gotr export <resource> <endpoint> [id] [flags]
```

## Флаги

```text
-h, --help                справка для export
--milestone-id string     ID milestone (для get_runs)
-p, --project-id string   ID проекта (для эндпоинтов с {project_id})
--save                    Сохранить ответ в ~/.gotr/exports/export/
--section-id string       ID секции (для get_cases)
-s, --suite-id string     ID тест-сюиты (для get_cases)
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
gotr export --help
```

## Источник

- Данные разделов выше сформированы из фактического вывода `--help` текущего кода CLI.

---

← [Команды](index.md) · [Гайды](../index.md) · [Документация](../../index.md)
