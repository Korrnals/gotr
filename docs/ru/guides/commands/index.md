# Команды

Language: Русский | [English](../../../en/guides/commands/index.md)

## Навигация

- [Документация](../../index.md)
  - [Гайды](../index.md)
    - [Установка](../installation.md)
    - [Конфигурация](../configuration.md)
    - [Интерактивный режим](../interactive-mode.md)
    - [Прогресс](../progress.md)
    - [Каталог команд](index.md)
      - [Общие](#общие)
        - [global-flags](global-flags.md)
        - [config](config.md)
        - [completion](completion.md)
        - [self-test](self-test.md)
      - [CRUD операции](#crud-операции)
        - [add](add.md)
        - [delete](delete.md)
        - [update](update.md)
        - [list](list.md)
        - [export](export.md)
      - [Основные ресурсы](#основные-ресурсы)
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
      - [Специальные ресурсы](#специальные-ресурсы)
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

Справочный каталог всех top-level CLI-команд проекта.

## Структура команд

Ниже приведена компактная структура команд: что входит в группу и для каких задач используется.

### Общие

Служебные команды и базовая настройка CLI.

- [global-flags](global-flags.md) — единые флаги подключения, вывода и поведения.
- [config](config.md) — управление локальной конфигурацией клиента.
- [completion](completion.md) — генерация shell completion для bash/zsh/fish/powershell.
- [self-test](self-test.md) — быстрая проверка окружения и доступности API.

### CRUD операции

Универсальные операции создания, изменения, удаления и выгрузки.

- [add](add.md) — создание сущностей через API.
- [delete](delete.md) — удаление ресурсов.
- [update](update.md) — изменение существующих сущностей.
- [list](list.md) — получение списков и базовых выборок.
- [export](export.md) — экспорт данных в файлы и отчётные форматы.

### Основные ресурсы

Основные namespace-команды для ежедневной работы с TestRail.

- [get](get.md) — чтение ресурсов и справочных данных.
- [sync](sync.md) — синхронизация данных между проектами/структурами.
- [compare](compare.md) — сравнение сущностей и отличий.
- [cases](cases.md) — операции с test cases.
- [run](run.md) — работа с test runs.
- [result](result.md) — добавление и просмотр результатов тестов.
- [test](test.md) — операции с отдельными тестами run.
- [tests](tests.md) — массовые операции с набором тестов.
- [attachments](attachments.md) — загрузка и получение вложений.
- [plans](plans.md) — работа с test plans.
- [reports](reports.md) — доступ к отчётам TestRail.

### Специальные ресурсы

Расширенные и специализированные endpoint-группы TestRail.

- [bdds](bdds.md) — BDD-данные кейсов.
- [configurations](configurations.md) — конфигурации и наборы параметров.
- [datasets](datasets.md) — датасеты и связанные структуры.
- [groups](groups.md) — пользовательские группы и доступы.
- [labels](labels.md) — метки и категоризация.
- [milestones](milestones.md) — вехи проекта.
- [roles](roles.md) — роли и разрешения.
- [templates](templates.md) — шаблоны кейсов.
- [users](users.md) — пользователи и атрибуты.
- [variables](variables.md) — переменные и параметры.

---

← [Гайды](../index.md) · [Документация](../../index.md)
