# Архитектура

Language: Русский | [English](../../en/architecture/index.md)

## Навигация

- [Документация](../index.md)
  - [Гайды](../guides/index.md)
    - [Установка](../guides/installation.md)
    - [Конфигурация](../guides/configuration.md)
    - [Интерактивный режим](../guides/interactive-mode.md)
    - [Прогресс](../guides/progress.md)
    - [Каталог команд](../guides/commands/index.md)
      - [Общие](../guides/commands/index.md#общие)
      - [CRUD операции](../guides/commands/index.md#crud-операции)
      - [Основные ресурсы](../guides/commands/index.md#основные-ресурсы)
      - [Специальные ресурсы](../guides/commands/index.md#специальные-ресурсы)
    - [Инструкции](../guides/instructions/index.md)
  - [Архитектура](index.md)
    - [Обзор](overview.md)
    - [Concurrency](concurrency.md)
    - [Стандарты](standards.md)
    - [План распараллеливания](recursive-parallelization-plan.md)
  - [Эксплуатация](../operations/index.md)
  - [Отчёты](../reports/index.md)
- [Главная](../../../README_ru.md)

## Содержание

### Принципы проектирования

Общие подходы, слои системы и стратегии расширения.

- [Обзор](overview.md) — архитектурные решения, слои и их ответственность
- [Стандарты](standards.md) — соглашения кодирования, именования, структуры пакетов

### Параллелизация

Модели concurrent-обработки и масштабирования.

- [Concurrency](concurrency.md) — пакет concurrent, worker pool, каналы
- [План распараллеливания](recursive-parallelization-plan.md) — схема рекурсивного разбиения работы

---

← [Документация](../index.md)
