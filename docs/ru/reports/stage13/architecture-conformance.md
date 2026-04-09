# Stage 13 - Architecture Conformance (Step 2)

Language: Русский | [English](../../../en/reports/stage13/architecture-conformance.md)

## Навигация

- [Документация](../../index.md)
  - [Гайды](../../guides/index.md)
    - [Установка](../../guides/installation.md)
    - [Конфигурация](../../guides/configuration.md)
    - [Интерактивный режим](../../guides/interactive-mode.md)
    - [Прогресс](../../guides/progress.md)
    - [Каталог команд](../../guides/commands/index.md)
      - [Общие](../../guides/commands/index.md#общие)
      - [CRUD операции](../../guides/commands/index.md#crud-операции)
      - [Основные ресурсы](../../guides/commands/index.md#основные-ресурсы)
      - [Специальные ресурсы](../../guides/commands/index.md#специальные-ресурсы)
  - [Архитектура](../../architecture/index.md)
  - [Эксплуатация](../../operations/index.md)
  - [Отчёты](../index.md)
    - [Stage 13](index.md)
    - [История](../history/index.md)
      - [Final Audit](final-coverage-audit-2026-04-05.md)
      - [Release Summary](release-summary.md)
      - [Audit Report](audit-report.md)
      - [Quality Metrics](quality-metrics.md)
      - [API Compliance](api-compliance-matrix.md)
      - [CLI Contract](cli-contract-matrix.md)
      - [Architecture Conformance](architecture-conformance.md)
      - [Reliability Audit](reliability-audit.md)
      - [Coverage Matrix](test-coverage-matrix.md)
      - [Checklist](coverage-checklist.md)
      - [Layer 2 Wave](layer2-coverage-wave.md)
      - [TODO](todo.md)
- [Главная](../../../../README_ru.md)

## Проверки зависимостей

- Forbidden dependency check (`internal/pkg -> cmd`): PASS
- Public boundary check (`pkg -> internal`): PASS
- Command-layer runtime coupling (`cmd/compare -> internal/concurrency`): FOUND
- Interactive helper duplication in command tree: FOUND

## Сводка наблюдений

### A1. Запретные зависимости не обнаружены (Low, accepted)

- `internal/**` не импортирует `cmd/**`.
- `pkg/**` не импортирует `internal/**`.

Вывод:

- Базовые архитектурные границы соблюдены.

### A2. Прямая связка compare-команд с concurrency runtime (Medium, needs follow-up)

Найдено прямое использование `internal/concurrency` в compare-командах:

- cmd/compare/cases.go
- cmd/compare/sections.go
- cmd/compare/simple.go
- cmd/compare/retry_failed_pages.go

Вывод:

- Для Stage 13 это допустимо как текущий runtime seam, но повышает риск регрессий при рефакторинге конкурентного ядра.

Remediation direction:

- Выделить стабильный adapter/facade слой для `cmd/compare` с минимальной surface area.

### A3. Дублирование interactive helper паттерна в command tree (Medium, must-fix in Stage 13)

Обнаружено 17 файлов вида `cmd/*/interactive_helpers.go`:

- cmd/attachments/interactive_helpers.go
- cmd/bdds/interactive_helpers.go
- cmd/cases/interactive_helpers.go
- cmd/configurations/interactive_helpers.go
- cmd/datasets/interactive_helpers.go
- cmd/groups/interactive_helpers.go
- cmd/labels/interactive_helpers.go
- cmd/milestones/interactive_helpers.go
- cmd/plans/interactive_helpers.go
- cmd/reports/interactive_helpers.go
- cmd/roles/interactive_helpers.go
- cmd/run/interactive_helpers.go
- cmd/templates/interactive_helpers.go
- cmd/test/interactive_helpers.go
- cmd/tests/interactive_helpers.go
- cmd/users/interactive_helpers.go
- cmd/variables/interactive_helpers.go

Вывод:

- Это hotspot связности и единая зона для стандартизации interactive contract (quiet/non-interactive behavior).

Remediation direction:

- Провести консолидацию повторяющихся helper-паттернов через shared layer с сохранением UX-контракта.

## File Risk Map (initial)

- High:
- internal/client/client.go (transport/retry/timeout criticality)
- cmd/root.go (global flags/context wiring)
- internal/ui/display.go (quiet/output contract)

- Medium:
- cmd/compare/cases.go
- cmd/compare/sections.go
- cmd/compare/simple.go
- cmd/compare/retry_failed_pages.go
- cmd/*/interactive_helpers.go (family)

- Low:
- Тестовые вспомогательные файлы без runtime side effects

## Плановые дельты, добавленные в Stage 13

1. Добавлен отдельный подпоток "Interactive helper consolidation" в Core Audits/Refactoring.
2. Добавлен подпоток "Compare runtime seam hardening" для снижения прямой связки с `internal/concurrency`.
3. Обновлены TODO и системный PLAN/Stage design под текущий статус Step 2.

---

← [Stage 13](index.md) · [Отчёты](../index.md) · [Документация](../../index.md)
