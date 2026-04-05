# Команды config

Language: Русский | [English](../../../en/guides/commands/config.md)

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

Команда `gotr config` управляет локальной конфигурацией клиента TestRail.

## Что делает

- Создаёт и обновляет локальный конфиг CLI.
- Показывает текущие параметры подключения и активный профиль.
- Позволяет безопасно переключаться между окружениями и аккаунтами.

## Когда использовать

- Когда нужно получить предсказуемый CLI-поток для автоматизации.
- Когда важно сократить ручные действия и риск ошибок.
- Когда операция должна одинаково выполняться локально и в CI/CD.

## Подкоманды

- `gotr config init`
- `gotr config path`
- `gotr config view`
- `gotr config edit`

## Примеры

```bash
# Инициализация конфигурации
gotr config init

# Путь к активному конфигу
gotr config path

# Проверка текущих параметров
gotr config view
```

## Полезные флаги

- `--json` для вывода в машинно-обрабатываемом формате.
- `--output` / `--save` для сохранения результата в файл.
- `--verbose` для детальной диагностики выполнения.

---

← [Команды](index.md) · [Гайды](../index.md) · [Документация](../../index.md)
