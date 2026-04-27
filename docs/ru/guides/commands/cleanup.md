# Команда: cleanup

Language: Русский | [English](../../../en/guides/commands/cleanup.md)

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
      - [cleanup](cleanup.md)
    - [Инструкции](../instructions/index.md)
  - [Архитектура](../../architecture/index.md)
  - [Эксплуатация](../../operations/index.md)
  - [Отчёты](../../reports/index.md)
- [Главная](../../../../README_ru.md)

## Обзор 🎯

`gotr cleanup` — ручной executor retention-политик над локальным стором
`~/.gotr/`. По умолчанию ничего не удаляется автоматически: политика
описывается в конфиге (секция `retention.*`) и применяется только по
команде пользователя.

> [!IMPORTANT]
> Перед первым боевым запуском всегда используйте `--dry-run` и проверяйте
> план.

## Синтаксис 🧩

```bash
gotr cleanup {reports | exports | snaps | all} [--dry-run]
```

## Подкоманды

| Подкоманда | Что чистит | Источник политики |
| --- | --- | --- |
| `reports` | `~/.gotr/reports/**` (иерархия по категориям) | `retention.reports` |
| `exports` | `~/.gotr/exports/{snaps,reports,api}/**` | `retention.exports` |
| `snaps` | `~/.gotr/snaps/**` (делегирует `gotr snap gc`) | `retention.snaps` |
| `all` | последовательно `reports` → `exports` → `snaps` | все три секции |

## Флаги ⚙️

```text
--dry-run    только показать план, ничего не удалять (override над retention.*.dry_run)
-h, --help   справка
```

## Конфигурация политики

```yaml
retention:
  reports:
    enabled: true
    max_age_days: 90
    max_count: 500
    keep_categories: [coverage]   # whitelist категорий
    dry_run: false
  exports:
    enabled: true
    max_age_days: 30
  snaps:
    enabled: true
    max_age_days: 180
    max_count: 100
```

- `enabled: false` → команда явно выходит с сообщением «retention disabled».
- `keep_categories` учитывается только для `reports` (coverage-артефакты
  обычно защищают whitelist-ом).
- Внутри `reports` удаление идёт по категориям: сначала фильтр по возрасту,
  затем trim до `max_count` самых свежих. Перед выходом регенерируется
  `INDEX.md`.

## Примеры 🚀

### Сценарий 1: Посмотреть план очистки отчётов

```bash
gotr cleanup reports --dry-run
```

Покажет список файлов-кандидатов по категориям, без удаления.

### Сценарий 2: Полная очистка

```bash
gotr cleanup all --dry-run
gotr cleanup all
```

### Сценарий 3: Только экспорты

```bash
gotr cleanup exports
```

## Взаимодействие с `gotr snap gc`

`gotr cleanup snaps` — тонкая обёртка над существующим `gotr snap gc`.
Старый workflow продолжает работать без изменений; новая команда нужна
для единого UX `cleanup all` и для чтения политики из `retention.snaps`.

## См. также

- [Конфигурация → Retention](../configuration.md)
- [Архитектура: UX polish v3.3.0](../../architecture/ux-polish-v3.3.0.md)

---

← [Каталог команд](index.md) · [Документация](../../index.md)
