# Команды update

Language: Русский | [English](../../../en/guides/commands/update.md)

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

Команда `gotr update` обновляет существующие ресурсы.

## Что делает

- Обновляет поля существующих сущностей без пересоздания.
- Позволяет синхронизировать метаданные между инструментами.
- Используется в миграционных и регулярных maintenance-скриптах.

## Когда использовать

- Когда нужно получить предсказуемый CLI-поток для автоматизации.
- Когда важно сократить ручные действия и риск ошибок.
- Когда операция должна одинаково выполняться локально и в CI/CD.

## Примеры

```bash
# Обновить кейс
gotr update case 12345 --title "Updated title"

# Обновить run
gotr update run 9988 --name "Nightly"

# Обновить milestone
gotr update milestone 77 --name "Release 2.1"
```

## Полезные флаги

- `--json` для вывода в машинно-обрабатываемом формате.
- `--output` / `--save` для сохранения результата в файл.
- `--verbose` для детальной диагностики выполнения.

---

← [Команды](index.md) · [Гайды](../index.md) · [Документация](../../index.md)
