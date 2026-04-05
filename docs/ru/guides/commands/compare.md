# Команды COMPARE (анализ и сравнение)

Language: Русский | [English](../../../en/guides/commands/compare.md)

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

## Обзор

Команда `gotr compare` выполняет **глубокий анализ различий** между проектами, сьютами, кейсами и другими структурами TestRail. Используется для валидации миграций, обнаружения несоответствий, аудита данных и подготовки отчётов перед синхронизацией.

## Что делает

- **Сравнение структур**: Сьюты, секции, иерархия
- **Сравнение контента**: Кейсы, шаги, ожидаемые результаты
- **Анализ дубликатов**: Поиск кейсов с идентичным содержанием
- **Статистика**: Подсчёт new, modified, deleted, duplicates
- **Экспорт отчётов**: JSON, CSV для дальнейшей обработки
- **Интерактивный режим**: Просмотр в диалоге
- **Фильтрация**: По полям, диапазонам ID, статусам

## Когда использовать

- До миграции (sync): проверить наличие дубликатов в dst
- После миграции: валидировать что всё скопировалось корректно
- Для аудита: регулярная проверка целостности данных
- Для отчётов: анализ расхождений между версиями

## Подкоманды

### `gotr compare projects` — сравнение проектов

Обзор: статистика по кейсам, сьютам, шаблонам, конфигурациям.

```bash
gotr compare projects --src-project 30 --dst-project 31
```

### `gotr compare suites` — сравнение сьютов

Структурные различия (секции, порядок, параметры).

```bash
gotr compare suites --src-project 30 --src-suite 20069 \
                    --dst-project 31 --dst-suite 19859 \
                    --report summary
```

### `gotr compare cases` — сравнение кейсов

Детальный анализ кейсов: what's new, modified, deleted, duplicates.

```bash
gotr compare cases --src-project 30 --src-suite 20069 \
                   --dst-project 31 --dst-suite 19859 \
                   --report detailed \
                   --output diff.csv
```

### `gotr compare sections` — сравнение иерархии секций

Структура папок и организация.

```bash
gotr compare sections --src-project 30 --src-suite 20069 \
                      --dst-project 31 --dst-suite 19859 \
                      --output sections_diff.json
```

## Примеры реальных сценариев

### Пример 1: Аудит после миграции (проверить полноту)

```bash
# Быстрая статистика
gotr compare suites --src-project 30 --src-suite 20069 \
                    --dst-project 31 --dst-suite 19859 \
                    --report summary

# Если есть расхождения → детальный анализ
gotr compare cases --src-project 30 --src-suite 20069 \
                   --dst-project 31 --dst-suite 19859 \
                   --report detailed \
                   --output details.csv
```

### Пример 2: Поиск дубликатов в целевом проекте

```bash
# Перед миграцией — есть ли уже похожие кейсы?
gotr compare cases --src-project 30 --src-suite 20069 \
                   --dst-project 31 \
                   --find-duplicates \
                   --compare-field title
```

### Пример 3: Экспорт различий для анализа

```bash
gotr compare cases --src-project 30 --src-suite 20069 \
                   --dst-project 31 --dst-suite 19859 \
                   --output report_for_qa.csv

# Откройте в Excel для ручного анализа
```

## Флаги и параметры

### Обязательные

| Флаг | Описание | Пример |
| --- | --- | --- |
| `--src-project` | ID исходного проекта | `--src-project 30` |
| `--dst-project` | ID целевого проекта | `--dst-project 31` |

### Контекст (опциональные)

| Флаг | Описание | По умолчанию |
| --- | --- | --- |
| `--src-suite` | Исходный сьют | Все сьюты src-project |
| `--dst-suite` | Целевой сьют | Все сьюты dst-project |

### Стратегия сравнения

| Флаг | Описание | Когда |
| --- | --- | --- |
| `--compare-field` | Поле для сравнения (title/name/id) | Когда `title` недостаточно |
| `--ignore-draft` | Пропустить черновики | Когда draft не важны |
| `--ignore-case` | Case-insensitive matching | Для нечувствительного матча |

### Вывод и отчёты

| Флаг | Описание | Опции |
| --- | --- | --- |
| `--report` | Формат отчёта | `summary` / `detailed` |
| `--output` | Сохранить результаты | `report.csv`, `report.json` |
| `--quiet` | Минимум вывода | Для CI |
| `--verbose` | Максимум деталей | Для отладки |

### Анализ

| Флаг | Описание | Пример |
| --- | --- | --- |
| `--find-duplicates` | Поиск дубликатов | `--find-duplicates` |
| `--group-by` | Группировка результатов | `--group-by section` |

## Алгоритм сравнения

```txt
compare cases выполняет:

1. Load данные
   ├─ src_data = fetch src project/suite/cases
   └─ dst_data = fetch dst project/suite/cases

2. Индексирование по compare-field
   ├─ src_index[compare_field] = case_id
   └─ dst_index[compare_field] = case_id

3. Проход по каждому src case
   ├─ lookup: есть ли в dst по compare-field?
   ├─ if found → структурное сравнение
   │   ├─ Identical → status = PRESENT
   │   └─ Different → status = MODIFIED
   └─ if not found → поиск похожего
       ├─ if similar found → status = POTENTIAL_DUPLICATE
       └─ if not found → status = MISSING

4. Классификация
   ├─ PRESENT: совпадают (успешная синхронизация)
   ├─ MODIFIED: есть но отличаются (нужно обновить)
   ├─ MISSING: не найдены в dst (не синхронизированы)
   └─ POTENTIAL_DUPLICATE: похожи по содержанию

5. Экспорт результатов
   ├─ summary: только статистика
   ├─ detailed: полный список с diff
   └─ --output: в файл JSON/CSV
```

## Обработка ошибок

| Ошибка | Причина | Решение |
| --- | --- | --- |
| `Project not found` | Проект не существует | `gotr get projects` |
| `Suite not found` | Сьют не найден | Проверить `--src-suite` / `--dst-suite` |
| `Compare field invalid` | Неправильное поле | Используйте: title, name, id |
| `Output file permission denied` | Нет доступа к файлу | Проверить права на директорию |
| `API timeout` | Сравнение заняло > timeout | ↑ `--timeout` или используйте `--src-suite` |
| `Memory exceeded` | Слишком большой dataset | Используйте `--src-suite` для сужения |

## Типичный workflow

```bash
# 1. Подготовка к миграции
gotr compare cases --src-project 30 --src-suite 20069 \
                   --dst-project 31 \
                   --find-duplicates \
                   --report summary
# Если дубликатов мало → можем мигрировать

# 2. После миграции (sync full)
gotr compare cases --src-project 30 --src-suite 20069 \
                   --dst-project 31 --dst-suite 19859 \
                   --report detailed \
                   --output post_sync_diff.csv

# 3. Проанализировать различия
cat post_sync_diff.csv | grep "MODIFIED"
```

## FAQ

**Q: Чем compare отличается от sync --dry-run?**  
A: `sync --dry-run` показывает что будет **создано**. `compare` анализирует **уже существующие** различия.

**Q: Как найти повреждённые данные при миграции?**  
A: `gotr compare cases --report detailed` → ищите статус `MODIFIED`.

**Q: Можно ли сравнивать между разными TestRail инстансами?**  
A: Нет. Для этого используйте `export` + `import`.

**Q: На 100k+ кейсов что будет?**  
A: Используйте `--src-suite` для сужения scope, иначе может занять часы.

**Q: Что означает "POTENTIAL_DUPLICATE"?**  
A: Кейс существует в dst но под другим именем (похож по содержанию). Проверьте вручную.

---

← [Команды](index.md) · [Гайды](../index.md) · [Документация](../../index.md)
