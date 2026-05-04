# Гайды

Language: Русский | [English](../../en/guides/index.md)

## Навигация

- [Документация](../index.md)
  - [Гайды](index.md)
    - [Установка](installation.md)
    - [Конфигурация](configuration.md)
    - [Интерактивный режим](interactive-mode.md)
    - [Прогресс](progress.md)
    - [Smoke-тестирование](smoke-testing.md)
    - [Миграционные архивы](migration-archive.md)
    - [Каталог команд](commands/index.md)
    - [Инструкции](instructions/index.md)
  - [Архитектура](../architecture/index.md)
  - [Эксплуатация](../operations/index.md)
  - [Отчёты](../reports/index.md)
- [Главная](../../../README_ru.md)

## Содержание

### Установка и конфигурация

Начните отсюда — развёртывание gotr и настройка подключения.

- [Установка](installation.md) — скачивание и установка, поддерживаемые платформы
- [Конфигурация](configuration.md) — создание конфига, аутентификация в TestRail

### Команды

Полный справочник 32+ CLI-команд, организованных по 4 группам.

- [Каталог команд](commands/index.md) — полный навигатор со структурой **CRUD операции**, **Основные ресурсы**, **Специальные ресурсы**

### Инструкции

Готовые пошаговые рецепты для типовых задач.

- [Инструкции](instructions/index.md) — миграция данных, CRUD-операции, сравнение проектов
- Новое в v3.3: [жизненный цикл отчётов](instructions/reports-lifecycle.md), [retention/cleanup runbook](instructions/retention-and-cleanup-runbook.md), [миграция TLS на ca_bundle](instructions/tls-ca-bundle-migration.md).
- [Migration guide v3.3](migration-guide-v3.3.md) — как обновиться с v3.2 без потерь.

### Специальные режимы

- [Интерактивный режим](interactive-mode.md) — работа в режиме диалога с автодополнением
- [Прогресс](progress.md) — отслеживание длительных операций с прогресс-барами

### Тестирование

- [Smoke-тестирование](smoke-testing.md) — E2E тесты snap/rollback на встроенном mock-сервере или реальном TestRail

### Перенос состояния между машинами

- [Миграционные архивы](migration-archive.md) — упаковать всё `~/.gotr/` (или произвольный набор снапшотов) в один tar.gz и восстановить на другой машине одной командой

---

← [Документация](../index.md)
