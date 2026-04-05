# Справка: Глобальные флаги

Language: Русский | [English](../../../en/guides/commands/global-flags.md)

## Навигация

- [Документация](../../index.md)
  - [Гайды](../index.md)
    - [Установка](../installation.md)
    - [Конфигурация](../configuration.md)
    - [Интерактивный режим](../interactive-mode.md)
    - [Прогресс](../progress.md)
    - [Каталог команд](index.md)
      - [Общие](global-flags.md)
        - [global-flags](global-flags.md)
        - [config](config.md)
        - [completion](completion.md)
        - [self-test](self-test.md)
      - [CRUD операции](add.md)
      - [Основные ресурсы](get.md)
      - [Специальные ресурсы](bdds.md)
  - [Архитектура](../../architecture/index.md)
  - [Эксплуатация](../../operations/index.md)
  - [Отчёты](../../reports/index.md)
- [Главная](../../../../README_ru.md)


## Обзор
gotr — удобная утилита для работы с TestRail API v2.
Поддерживает просмотр доступных эндпоинтов, выполнение запросов и многое другое.

## Синтаксис
```bash
gotr [command]
```

## Подкоманды

| Подкоманда | Описание |
| --- | --- |
| `add` | Создать новый ресурс (POST-запрос) |
| `attachments` | Управление файловыми вложениями |
| `bdds` | Управление BDD сценариями |
| `cases` | Управление тест-кейсами |
| `compare` | Comparison данных между проектами |
| `completion` | Generate completion script |
| `config` | Управление конфигурацией gotr |
| `datasets` | Управление датасетами (тестовыми данными) |
| `delete` | Удалить ресурс (DELETE/POST-запрос) |
| `export` | Экспорт данных из TestRail в JSON-файл |
| `get` | GET-запросы к TestRail API |
| `groups` | Управление группами пользователей |
| `labels` | Управление метками тестов |
| `list` | Вывод списка доступных эндпоинтов TestRail API по ресурсу |
| `milestones` | Управление майлстонами (этапами) проекта |
| `plans` | Управление тест-планами |
| `reports` | Управление отчётами проекта |
| `result` | Управление результатами тестов в TestRail |
| `roles` | Управление ролями пользователей |
| `run` | Управление test runs в TestRail |
| `self-test` | Run self-diagnostic tests |
| `sync` | Синхронизация данных TestRail между проектами |
| `templates` | Управление шаблонами тест-кейсов |
| `test` | Управление тестами в TestRail |
| `tests` | Управление тестами |
| `update` | Обновить существующий ресурс (POST-запрос) |
| `users` | Управление пользователями TestRail |
| `variables` | Управление переменными тест-кейсов |

## Флаги

```text
-k, --api-key string    API ключ TestRail
-c, --config            Создать дефолтный файл конфигурации
-f, --format string     Формат вывода: table, json, csv, md, html (default "table")
-h, --help              справка для gotr
--insecure              Пропустить проверку TLS сертификата
--non-interactive       Отключить интерактивные подсказки; завершить с ошибкой если требуется ввод
-q, --quiet             Подавить служебный вывод (прогресс, статистику, сообщения о сохранении)
--url string            Базовый URL TestRail
-u, --username string   Email пользователя TestRail
-v, --version           version for gotr
```

## Примеры

```bash
gotr --help
gotr list --help
```

## Источник

- Данные разделов выше сформированы из фактического вывода `--help` текущего кода CLI.

---

← [Команды](index.md) · [Гайды](../index.md) · [Документация](../../index.md)
