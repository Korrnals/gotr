# Команды labels

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
        - [other](other.md)
  - [Архитектура](../../architecture/index.md)
  - [Эксплуатация](../../operations/index.md)
  - [Отчёты](../../reports/index.md)
- [Главная](../../../../README_ru.md)

Команда `gotr labels` управляет labels и категоризацией.

## Что делает

- Управляет метками и категоризацией.
- Помогает организовать фильтрацию и навигацию по кейсам.
- Используется для отчётности по срезам и тегам.

## Когда использовать

- Когда нужно получить предсказуемый CLI-поток для автоматизации.
- Когда важно сократить ручные действия и риск ошибок.
- Когда операция должна одинаково выполняться локально и в CI/CD.

## Примеры

```bash
# Справка по команде
gotr labels --help

# Справка по подкоманде
gotr labels get --help

# Базовый вызов
gotr labels --json
```

## Полезные флаги

- `--json` для вывода в машинно-обрабатываемом формате.
- `--output` / `--save` для сохранения результата в файл.
- `--verbose` для детальной диагностики выполнения.

---

← [Команды](index.md) · [Гайды](../index.md) · [Документация](../../index.md)
