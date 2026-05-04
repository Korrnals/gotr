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

## Случай F: Массовая очистка вложений TestRail

**Сценарий:** освободить место на стороне TestRail, удалив вложения
старше N дней, со страховочным снапшотом для отката. Полный список
флагов и сценарий отката описан в
[`gotr attachments cleanup`](../commands/attachments.md).

Раннер из 5 шагов:

1. **Dry-run preview.** Всегда начинайте с `--dry-run`, чтобы увидеть
   pre-flight summary (область проектов, типы сущностей, общее
   количество вложений, оценка размера).

   ```bash
   gotr attachments cleanup --all-projects \
     --entity-type result --older-than 90d --dry-run
   ```

2. **Просмотр pre-flight summary.** Проверьте ID проектов, типы
   сущностей и прогноз количества/размера. Если область слишком
   широкая — повторите dry-run с более узкими `--project` /
   `--entity-type` / `--limit`.

3. **Подтверждение и запуск.** Уберите `--dry-run`. Команда создаст
   снапшот в категории `cleanup-attachments` перед удалением.

   ```bash
   gotr attachments cleanup --all-projects \
     --entity-type result --older-than 90d --concurrency 4
   ```

   Используйте `--force`, чтобы пропустить финальное подтверждение в
   скриптах.

4. **Местоположение снапшота.** Снапшоты лежат в
   `~/.gotr/snaps/cleanup-attachments/<id>/` (`data.json` + дерево
   `files/`). Default TTL для этой категории — **7 дней**; см.
   [snap → Per-category retention TTL](../commands/snap.md#per-category-retention-ttl),
   чтобы продлить окно через `snap.retention.category_ttl_days`, если
   требуется более длинный rollback-window.

   ```bash
   gotr snap list --category cleanup-attachments
   ```

5. **Откат при необходимости.** Перезаливает снапшотные блобы обратно в
   TestRail. Перезалитые вложения получают **новые** TestRail-ID
   (исходные ID восстановить нельзя). Тип сущности `test` при откате
   корректно пропускается, поскольку TestRail не имеет endpoint
   `add_attachment_to_test`.

   ```bash
   gotr snap rollback <snapshot_id>
   ```

> **Подсказка.** `--no-snapshot` стоит использовать только если вы
> сознательно отказываетесь от возможности отката (например, разовая
> очистка без требований к восстановлению). Режим со снапшотом — по
> умолчанию и рекомендуется.

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
