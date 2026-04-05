# Команды SYNC (миграция)

Language: Русский | [English](../../../en/guides/commands/sync.md)

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

## Overview

Команда `gotr sync` — комплексный инструмент для **миграции и синхронизации данных** между проектами и структурами TestRail. Поддерживает интеллектуальное сопоставление, обнаружение дубликатов, генерацию маппингов и параллельную обработку больших объёмов данных в CI/CD конвейерах.

## Что делает

- **Миграция сущностей**: Перемещает кейсы, shared steps, секции, сьюты и другие структуры между проектами
- **Умное сопоставление**: Обнаруживает дубликаты по настраиваемому полю (`title`, `name`, `id`)
- **Маппинг для связей**: Генерирует таблицу соответствия ID между исходным и целевым проектом
- **Параллельная загрузка**: Использует горутины для ускорения больших миграций
- **Интерактивный режим**: При отсутствии флагов предлагает выбор в диалоге
- **Сухой запуск**: `--dry-run` показывает план без реального изменения данных
- **Автоматизация**: Идеально для CI/CD, scheduled синхронизаций и репликаций

## Когда использовать

- Когда нужно получить предсказуемый CLI-поток для автоматизации.
- Когда важно сократить ручные действия и риск ошибок.
- Когда операция должна одинаково выполняться локально и в CI/CD.

## Подкоманды

### `gotr sync full` — полная миграция

Автоматическая миграция со всеми зависимостями: shared-steps → cases.

```bash
# Интерактивный режим
gotr sync full

# С явными параметрами
gotr sync full --src-project 30 --src-suite 20069 \
               --dst-project 31 --dst-suite 19859 \
               --approve
```

**Уровни сложности:**

1. **Базовый**: `gotr sync full` — интерактивный выбор
2. **Средний**: С параметрами + `--dry-run` для проверки
3. **Продвинутый**: Full + маппинги + параллелизм + сохранение результатов

### `gotr sync shared-steps` — миграция общих шагов

Миграция shared steps отдельно (подготовка к миграции cases).

```bash
gotr sync shared-steps --src-project 30 --src-suite 20069 \
                       --dst-project 31 \
                       --approve \
                       --output mapping.json
```

### `gotr sync cases` — миграция кейсов

Использует маппинг для замены ID shared-steps.

```bash
gotr sync cases --src-project 30 --src-suite 20069 \
                --dst-project 31 --dst-suite 19859 \
                --mapping mapping.json \
                --approve
```

### `gotr sync sections` — миграция структуры секций

Копирует иерархию секций без кейсов.

```bash
gotr sync sections --src-project 30 --src-suite 20069 \
                   --dst-project 31 --dst-suite 19859 \
                   --approve
```

### `gotr sync suites` — миграция целого сьюта

Полное копирование сьюта со всей структурой.

```bash
gotr sync suites --src-project 30 --src-suite 20069 \
                 --dst-project 31 \
                 --approve
```

## Примеры реальных сценариев

### Пример 1: Пошаговая миграция перед production

```bash
# Шаг 1: сухой запуск для анализа
gotr sync full --src-project 30 --src-suite 20069 \
               --dst-project 31 --dst-suite 19859 \
               --dry-run --verbose

# Шаг 2: если резулітать ОК → реальная миграция
gotr sync full --src-project 30 --src-suite 20069 \
               --dst-project 31 --dst-suite 19859 \
               --approve --save-mapping
```

### Пример 2: CI/CD интеграция с параллелизмом

```bash
gotr sync full --src-project 30 --src-suite 20069 \
               --dst-project 31 --dst-suite 19859 \
               --approve \
               --parallel-suites 8 \
               --parallel-pages 12 \
               --rate-limit 50
```

### Пример 3: Поиск дубликатов перед миграцией

```bash
# Использовать custom field для поиска дубликатов
gotr sync cases --src-project 30 --src-suite 20069 \
                --dst-project 31 --dst-suite 19859 \
                --compare-field custom_id \
                --mapping mapping.json \
                --approve
```

## Флаги и параметры

### Обязательные

| Флаг | Описание | Пример |
| --- | --- | --- |
| `--src-project` | ID исходного проекта | `--src-project 30` |
| `--dst-project` | ID целевого проекта | `--dst-project 31` |
| `--src-suite` | ID исходного сьюта | `--src-suite 20069` |

### Опциональные: идентификация

| Флаг | Описание | По умолчанию | Когда |
| --- | --- | --- | --- |
| `--dst-suite` | ID целевого сьюта | Выбор интерактивно | Когда хотите явно |
| `--compare-field` | Поле для поиска дубликатов (title/name/id) | `title` | Для custom логики |

### Опциональные: контроль выполнения

| Флаг | Описание | Когда |
| --- | --- | --- |
| `--approve` | Авто-подтверждение без диалога | В CI/CD |
| `--dry-run` | Просмотр плана без реальных изменений | **ВСЕГДА перед production** |
| `--quiet` | Минимум вывода | CI логи |
| `--verbose` | Максимум деталей | Отладка |

### Опциональные: результаты и маппинги

| Флаг | Описание | Пример |
| --- | --- | --- |
| `--mapping-file` | Файл маппинга для загрузки | `--mapping-file map.json` |
| `--output` | Файл для сохранения маппинга | `--output results.json` |
| `--save-mapping` | Автосохранение в `.testrail/mappings/` | `--save-mapping` |

## Алгоритм работы

```txt
sync full выполняет:

1. Валидация (проверка параметров и доступа)
   ├─ Load configurations
   ├─ Validate projects & suites
   └─ Setup logging

2. Фаза 1: Миграция shared steps
   ├─ Fetch: получить все shared steps из src
   ├─ Compare: сравнить с dst (по compare-field)
   ├─ Classify: new → создать, duplicates → пропустить
   ├─ Create: bulk insert в dst
   └─ Map: ID src → ID dst (сохранить JSON)

3. Фаза 2: Подготовка cases
   ├─ Fetch: получить все cases из src
   ├─ Mutate: заменить shared_step_id по маппингу
   └─ Prepare: batch statements

4. Фаза 3: Миграция cases
   ├─ Fetch (dst): получить существующие cases в dst
   ├─ Compare: сравнить с подготовленными
   ├─ Classify: как в фазе 1
   └─ Create/Update: bulk операции

5. Финализация
   ├─ Save: маппинг + результаты (если --save-mapping)
   ├─ Logs: запись в .testrail/logs/
   ├─ Summary: статистика
   └─ Exit
```

## Обработка ошибок

| Ошибка | Причина | Решение |
| --- | --- | --- |
| `Project not found` | Проект не существует | `gotr get projects` для списка |
| `Suite not found` | Сьют не найден | Проверить `--src-suite` / `--dst-suite` |
| `Mapping file not found` | Маппинг не существует (для cases) | Запустить `sync shared-steps` сначала |
| `API rate limit exceeded` | Слишком много параллельных запросов | ↓ `--parallel-suites`, ↓ `--parallel-pages` |
| `Duplicate mapping keys` | Конфликт в маппинге | Пересоздать маппинг или отредактировать |
| `Context deadline exceeded` | Операция заняла > --timeout | ↑ `--timeout` значение |

## Оптимизация для больших объёмов

```bash
# Маленький проект (<500 кейсов)
gotr sync full ... --parallel-suites 2 --parallel-pages 4

# Средний проект (500-5000 кейсов)  
gotr sync full ... --parallel-suites 4 --parallel-pages 8

# Крупный проект (5000-50000 кейсов)
gotr sync full ...--parallel-suites 8 --parallel-pages 12

# Очень крупный проект (>50000 кейсов)
gotr sync full ... --parallel-suites 16 --parallel-pages 20 --rate-limit 50
```

## Логирование и восстановление

```bash
# Просмотреть логи последней операции
tail -f .testrail/logs/sync_full_*.log

# Просмотреть сохранённые маппинги
ls -la .testrail/mappings/

# Повторный запуск с тем же маппингом
gotr sync cases ... --mapping .testrail/mappings/shared_steps_latest.json
# Операция будет idempotent (пропустит уже созданные)
```

## Типичный workflow

```bash
# 1. Анализ
gotr sync full --src-project 30 --src-suite 20069 \
               --dst-project 31 --dst-suite 19859 \
               --dry-run --verbose

# 2. Если ОК → миграция
gotr sync full --src-project 30 --src-suite 20069 \
               --dst-project 31 --dst-suite 19859 \
               --approve --save-mapping

# 3. Валидация результатов
gotr compare cases --src-project 30 --src-suite 20069 \
                   --dst-project 31 --dst-suite 19859

# 4. Проверить маппинг сохранился
cat .testrail/mappings/shared_steps_latest.json | jq .
```

## FAQ

**Q: Что произойдёт с дубликатами?**  
A: По умолчанию пропускаются (обнаруживаются по `--compare-field`). Используйте `gotr compare` перед миграцией.

**Q: Можно ли обновить существующие кейсы вместо пропуска?**  
A: Текущая версия пропускает дубликаты. Для обновления используйте `gotr update`.

**Q: Как восстановиться от ошибки посередине?**  
A: Маппинги сохраняются автоматически. Повторный запуск будет идемпотентным.

**Q: Сколько времени займёт миграция 10000 кейсов?**  
A: 5-15 минут при оптимальных настройках в зависимости от сетевой задержки.

---

← [Команды](index.md) · [Гайды](../index.md) · [Документация](../../index.md)
