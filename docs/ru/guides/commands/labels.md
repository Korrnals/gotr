# Команда: labels

Language: Русский | [English](../../../en/guides/commands/labels.md)

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
      - [Специальные ресурсы](bdds.md)
        - [bdds](bdds.md)
        - [configurations](configurations.md)
        - [datasets](datasets.md)
        - [groups](groups.md)
        - [labels](labels.md)
        - [milestones](milestones.md)
        - [roles](roles.md)
        - [templates](templates.md)
        - [users](users.md)
        - [variables](variables.md)
  - [Архитектура](../../architecture/index.md)
  - [Эксплуатация](../../operations/index.md)
  - [Отчёты](../../reports/index.md)
- [Главная](../../../../README_ru.md)


## Обзор
Обновление меток (labels) для тестов и тестовых прогонов.
Метки позволяют категоризировать и группировать тесты для удобного анализа.

## Синтаксис
```bash
gotr labels [command]
```

## Подкоманды

| Подкоманда | Описание |
| --- | --- |
| `get` | Получить информацию о метке |
| `list` | Получить список меток проекта |
| `update` | Обновить метки для тестов |

## Флаги

```text
-h, --help   справка для labels
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
gotr labels --help
gotr labels get --help
```

## Источник

- Данные разделов выше сформированы из фактического вывода `--help` текущего кода CLI.

---

← [Команды](index.md) · [Гайды](../index.md) · [Документация](../../index.md)
