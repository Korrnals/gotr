# Команда: add

Language: Русский | [English](../../../en/guides/commands/add.md)

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
Создаёт новый объект в TestRail через POST API.
Поддерживаемые эндпоинты:

## Синтаксис
```bash
gotr add <endpoint> [id] [flags]
```

## Флаги

```text
--announcement string   Announcement (для проекта)
--assignedto-id int     ID назначенного пользователя
--case-ids string       ID кейсов через запятую (для run)
--comment string        Комментарий (для result)
--defects string        Дефекты (для result)
--description string    Описание/announcement
--dry-run               Показать что будет выполнено без реальных изменений
--elapsed string        Время выполнения (для result)
-h, --help              справка для add
--include-all           Включить все кейсы (для run) (default true)
-i, --interactive       Интерактивный режим (wizard)
--json-file string      Путь к JSON-файлу с данными
--milestone-id int      ID milestone
-n, --name string       Название ресурса
--priority-id int       ID приоритета (для case)
--refs string           Ссылки (references)
--save                  Save output to file in ~/.gotr/exports/
--section-id int        ID секции
--show-announcement     Показывать announcement
--status-id int         ID статуса (для result)
--suite-id int          ID сьюта
--template-id int       ID шаблона (для case)
--title string          Заголовок (для case)
--type-id int           ID типа (для case)
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
gotr add --help
```

## Источник

- Данные разделов выше сформированы из фактического вывода `--help` текущего кода CLI.

---

← [Команды](index.md) · [Гайды](../index.md) · [Документация](../../index.md)
