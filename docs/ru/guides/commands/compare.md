# Команды compare

Language: Русский | [English](../../../en/guides/commands/compare.md)

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

Команда `gotr compare` сравнивает структуры и данные между источником и целевой средой.

## Что делает

- Сравнивает два набора данных и показывает отличия.
- Помогает обнаружить пропуски, дубликаты и рассинхрон.
- Используется перед синхронизацией как этап контроля качества.

## Когда использовать

- Когда нужно получить предсказуемый CLI-поток для автоматизации.
- Когда важно сократить ручные действия и риск ошибок.
- Когда операция должна одинаково выполняться локально и в CI/CD.

## Подкоманды

- `gotr compare all`
- `gotr compare cases`
- `gotr compare sections`

## Примеры

```bash
# Сравнить кейсы
gotr compare cases --src-project 30 --dst-project 31

# Сравнить секции
gotr compare sections --src-project 30 --dst-project 31

# Полное сравнение
gotr compare all --src-project 30 --dst-project 31
```

## Полезные флаги

- `--json` для вывода в машинно-обрабатываемом формате.
- `--output` / `--save` для сохранения результата в файл.
- `--verbose` для детальной диагностики выполнения.

---

← [Команды](index.md) · [Гайды](../index.md) · [Документация](../../index.md)
