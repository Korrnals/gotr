# Инструкция: retention и cleanup runbook

Language: Русский | [English](../../../en/guides/instructions/retention-and-cleanup-runbook.md)

## Навигация

- [Документация](../../index.md)
  - [Гайды](../index.md)
    - [Инструкции](index.md)
      - [Retention и cleanup runbook](retention-and-cleanup-runbook.md)
  - [Архитектура](../../architecture/index.md)
  - [Эксплуатация](../../operations/index.md)
  - [Отчёты](../../reports/index.md)
- [Главная](../../../../README_ru.md)

Пошаговые сценарии настройки и применения retention-политик в v3.3.0.
По умолчанию retention выключен — ничего не удаляется автоматически.

## Предусловия

- v3.3.0+, иерархия reports/exports уже приведена к новой схеме (см.
  [Жизненный цикл отчётов](reports-lifecycle.md), кейс A).

## Кейс A: Базовый конфиг с защитой coverage

**Сценарий:** держать последние 500 отчётов, не старше 90 дней;
coverage-артефакты никогда не трогать.

```yaml
# ~/.gotr/config/default.yaml
retention:
  reports:
    enabled: true
    max_age_days: 90
    max_count: 500
    keep_categories: [coverage]
    dry_run: false
  exports:
    enabled: true
    max_age_days: 30
  snaps:
    enabled: true
    max_age_days: 180
    max_count: 100
```

```bash
# Всегда — dry-run первым
gotr cleanup all --dry-run

# Боевой запуск
gotr cleanup all
```

## Кейс B: Только reports, без трогания snap/exports

```bash
gotr cleanup reports --dry-run
gotr cleanup reports
```

`keep_categories: [coverage, migrations]` — защитит и migration-отчёты.

## Кейс C: Аварийно временно отключить retention

```yaml
retention:
  reports:
    enabled: false
```

Команда `gotr cleanup reports` явно завершится сообщением
«retention for reports disabled», код выхода 0.

## Кейс D: CI: дросселируем локальный сторадж после миграции

```bash
# В CI-скрипте после гарантированно успешной миграции
gotr cleanup reports --dry-run > cleanup-plan.txt
# Ручной ревью артефакта; если ок — следующий запуск без --dry-run
```

Флаг `--dry-run` CLI имеет приоритет над `retention.*.dry_run` в YAML.

## Кейс E: «Только snap gc» через единый CLI

Исторически был `gotr snap gc`. В v3.3 `gotr cleanup snaps` —
совместимая обёртка, читающая `retention.snaps`. Старый путь продолжает
работать:

```bash
gotr snap gc                   # как раньше
gotr cleanup snaps             # через retention.snaps
```

## Критерии успеха

- `INDEX.md` обновлён после `cleanup reports`.
- Файлы в `keep_categories` не уменьшились.
- В `~/.gotr/exports/snaps/` остались только файлы свежее
  `max_age_days` или в пределах `max_count`.
- `gotr cleanup all --dry-run` возвращает exit 0 и план в stdout.

## Важные предупреждения

> **Dry-run всегда.** Первый запуск новой политики — **только** с
> `--dry-run`. Retention смотрит на mtime, а не на логический возраст;
> свежеимпортированный старый архив может неожиданно попасть в план
> удаления при `max_age_days`.

> **`keep_categories` применяется только к `reports`.** Для `snaps` и
> `exports` используйте `max_count` / `max_age_days` или держите их
> отключёнными.

## См. также

- [gotr cleanup](../commands/cleanup.md).
- [Конфигурация → Retention](../configuration.md).

---

← [Инструкции](index.md) · [Документация](../../index.md)
