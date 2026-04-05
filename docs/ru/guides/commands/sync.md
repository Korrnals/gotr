# Команды SYNC (миграция)

Language: Русский | [English](../../../en/guides/commands/sync.md)

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

Команда `gotr sync` синхронизирует сущности между проектами и структурами.

## Что делает

- Синхронизирует данные между источником и целевой структурой.
- Автоматизирует перенос кейсов/шагов/связей между проектами.
- Снижает ручные ошибки при миграциях и регулярной репликации данных.

## Когда использовать

- Когда нужно получить предсказуемый CLI-поток для автоматизации.
- Когда важно сократить ручные действия и риск ошибок.
- Когда операция должна одинаково выполняться локально и в CI/CD.

## Подкоманды

- `gotr sync full`
- `gotr sync cases`
- `gotr sync shared-steps`

## Примеры

```bash
# Полная синхронизация
gotr sync full --src-project 30 --dst-project 31 --approve

# Только кейсы
gotr sync cases --src-project 30 --dst-project 31 --approve

# Только shared steps
gotr sync shared-steps --src-project 30 --dst-project 31 --approve
```

## Полезные флаги

- `--json` для вывода в машинно-обрабатываемом формате.
- `--output` / `--save` для сохранения результата в файл.
- `--verbose` для детальной диагностики выполнения.

---

← [Команды](index.md) · [Гайды](../index.md) · [Документация](../../index.md)
