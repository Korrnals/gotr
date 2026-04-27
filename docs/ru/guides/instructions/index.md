# Инструкции

Language: Русский | [English](../../../en/guides/instructions/index.md)

## Навигация

- [Документация](../../index.md)
  - [Гайды](../index.md)
    - [Установка](../installation.md)
    - [Конфигурация](../configuration.md)
    - [Интерактивный режим](../interactive-mode.md)
    - [Прогресс](../progress.md)
    - [Каталог команд](../commands/index.md)
    - [Инструкции](index.md)
      - [Пошаговая интерактивная миграция](migration-interactive-walkthrough.md)
      - [Полная миграция](migration-full.md)
      - [Частичная миграция](migration-partial.md)
      - [Миграция shared steps](migration-shared-steps.md)
      - [Миграция ресурсов](migration-resources.md)
      - [⚠️ Живой тест-план миграции (перед боевым запуском)](live-migration-test-plan.md)
      - [Операторская карточка живого прогона](live-migration-operator-card.md)
      - [Получение данных](crud-get.md)
      - [Экспорт данных](crud-export.md)
      - [Создание объектов](crud-add.md)
      - [Обновление объектов](crud-update.md)
      - [Удаление объектов](crud-delete.md)
      - [Сравнение проектов](compare.md)
      - [Жизненный цикл отчётов](reports-lifecycle.md)
      - [Retention и cleanup runbook](retention-and-cleanup-runbook.md)
      - [TLS: insecure → ca_bundle](tls-ca-bundle-migration.md)
  - [Архитектура](../../architecture/index.md)
  - [Эксплуатация](../../operations/index.md)
  - [Отчёты](../../reports/index.md)
- [Главная](../../../../README_ru.md)

## Содержание

Практические пошаговые инструкции для типовых задач с gotr.
Каждая инструкция — готовый рецепт: предусловия, команды, проверка результата.

### Миграция данных

Перенос данных между проектами TestRail через `gotr sync`.

- [Пошаговая интерактивная миграция](migration-interactive-walkthrough.md) — все вариации миграций через интерактивный режим (walkthrough)
- [Полная миграция](migration-full.md) — shared steps + cases за один проход (`sync full`)
- [Частичная миграция](migration-partial.md) — cases с подстановкой mapping из предыдущего шага
- [Миграция shared steps](migration-shared-steps.md) — перенос только общих тестовых шагов
- [Миграция ресурсов](migration-resources.md) — suites, sections между проектами
- [⚠️ Живой тест-план миграции](live-migration-test-plan.md) — **обязательно перед боевым запуском**: изолированный тест на двух тестовых проектах с откатом и очисткой
- [Операторская карточка живого прогона](live-migration-operator-card.md) — краткий чеклист и команды для выполнения теста в терминале

### CRUD-операции

Повседневная работа с объектами TestRail.

- [Получение данных](crud-get.md) — `gotr get` для проектов, кейсов, shared steps и др.
- [Экспорт данных](crud-export.md) — `gotr export` в JSON/CSV/HTML с сохранением
- [Создание объектов](crud-add.md) — `gotr add` с интерактивным режимом и dry-run
- [Обновление объектов](crud-update.md) — `gotr update` полей сущностей
- [Удаление объектов](crud-delete.md) — `gotr delete` с мягким и жёстким удалением

### Сравнение

- [Сравнение проектов](compare.md) — `gotr compare` для аудита и предмиграционной разведки

### v3.3.0 — UX polish

Пошаговые кейсы под новый функционал (иерархия отчётов, retention/cleanup,
корпоративный TLS). См. также [Migration guide v3.3](../migration-guide-v3.3.md).

- [Жизненный цикл отчётов](reports-lifecycle.md) — миграция layout → генерация → просмотр (`--print`, `latest`) → экспорт snap с `--with-reports` → импорт в чистое окружение → очистка
- [Retention и cleanup runbook](retention-and-cleanup-runbook.md) — базовый конфиг с whitelist `coverage`, CI-сценарий dry-run, совместимость с `gotr snap gc`
- [TLS: insecure → ca_bundle](tls-ca-bundle-migration.md) — перенос корпоративного CA, подавление баннера `tls_insecure`, troubleshooting

---

← [Гайды](../index.md)
