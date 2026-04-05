# Команды test

Language: Русский | [English](../../../en/guides/commands/test.md)

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

Команда `gotr test` работает с отдельными тестами внутри run.

## Что делает

- Оперирует отдельным тестом внутри run.
- Позволяет точечно читать/обновлять статус конкретного test-экземпляра.
- Удобен для ручной диагностики проблемных тестов.

## Когда использовать

- Когда нужно получить предсказуемый CLI-поток для автоматизации.
- Когда важно сократить ручные действия и риск ошибок.
- Когда операция должна одинаково выполняться локально и в CI/CD.

## Примеры

```bash
# Справка по команде
gotr test --help

# Справка по подкоманде
gotr test get --help

# Базовый вызов
gotr test --json
```

## Полезные флаги

- `--json` для вывода в машинно-обрабатываемом формате.
- `--output` / `--save` для сохранения результата в файл.
- `--verbose` для детальной диагностики выполнения.

---

← [Команды](index.md) · [Гайды](../index.md) · [Документация](../../index.md)
