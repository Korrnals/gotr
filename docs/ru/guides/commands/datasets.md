# Команды datasets

Language: Русский | [English](../../../en/guides/commands/datasets.md)

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
        - [other](other.md)
  - [Архитектура](../../architecture/index.md)
  - [Эксплуатация](../../operations/index.md)
  - [Отчёты](../../reports/index.md)
- [Главная](../../../../README_ru.md)

Команда `gotr datasets` работает с dataset-ресурсами.

## Что делает

- Работает с dataset-сущностями.
- Позволяет поддерживать тестовые наборы данных для автосценариев.
- Используется для подготовки репрезентативных входных данных.

## Когда использовать

- Когда нужно получить предсказуемый CLI-поток для автоматизации.
- Когда важно сократить ручные действия и риск ошибок.
- Когда операция должна одинаково выполняться локально и в CI/CD.

## Примеры

```bash
# Справка по команде
gotr datasets --help

# Справка по подкоманде
gotr datasets get --help

# Базовый вызов
gotr datasets --json
```

## Полезные флаги

- `--json` для вывода в машинно-обрабатываемом формате.
- `--output` / `--save` для сохранения результата в файл.
- `--verbose` для детальной диагностики выполнения.

---

← [Команды](index.md) · [Гайды](../index.md) · [Документация](../../index.md)
