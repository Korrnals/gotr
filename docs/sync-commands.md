# Команды SYNC (миграция)

Команды для синхронизации данных между проектами.

## Общие принципы

- **Интерактивный режим**: если параметры не указаны, будет предложен выбор
- **Dry-run**: флаг `--dry-run` для проверки без изменений
- **Mapping**: автоматическая генерация mapping для связи старых и новых ID
- **Подтверждение**: интерактивное подтверждение перед импортом (или `--approve`)

## Полная миграция (full)

Выполняет миграцию shared steps и cases за один проход:

```bash
# Полностью интерактивно
gotr sync full

# Через флаги
gotr sync full \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 31 \
  --dst-suite 19859 \
  --approve \
  --save-mapping
```

## Миграция shared steps

```bash
# Интерактивно
gotr sync shared-steps

# С указанием параметров
gotr sync shared-steps \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 31 \
  --approve \
  --save-mapping \
  --output mapping.json
```

**Особенности:**

- Генерирует mapping (shared_step_id_old → shared_step_id_new)
- Можно сохранить mapping в файл для последующей миграции кейсов

## Миграция кейсов

```bash
# Интерактивно
gotr sync cases

# С mapping-файлом
gotr sync cases \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 31 \
  --dst-suite 19859 \
  --mapping-file mapping.json \
  --approve

# Dry-run (проверка без импорта)
gotr sync cases \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 31 \
  --dst-suite 19859 \
  --dry-run
```

## Миграция сьютов

```bash
gotr sync suites \
  --src-project 30 \
  --dst-project 31 \
  --approve
```

## Миграция секций

```bash
# Интерактивно
gotr sync sections

# Через флаги
gotr sync sections \
  --src-project 30 \
  --src-suite 20069 \
  --dst-project 31 \
  --dst-suite 19859 \
  --approve
```

## Общие флаги

| Флаг | Описание |
|------|----------|
| `--src-project` | ID source проекта |
| `--src-suite` | ID source сьюта |
| `--dst-project` | ID destination проекта |
| `--dst-suite` | ID destination сьюта |
| `--compare-field` | Поле для сравнения (default: `title`) |
| `--approve` | Автоматическое подтверждение |
| `--dry-run` | Проверка без импорта |
| `--save-mapping` | Сохранить mapping автоматически |
| `--mapping-file` | Файл mapping для загрузки |
| `--output` | Файл для сохранения результатов |

## Рабочий процесс миграции

### Сценарий 1: Полная миграция

```bash
# Шаг 1: Полная миграция shared steps + cases
gotr sync full --src-project 30 --src-suite 20069 --dst-project 31 --dst-suite 19859 --approve --save-mapping
```

### Сценарий 2: Пошаговая миграция

```bash
# Шаг 1: Миграция shared steps с сохранением mapping
gotr sync shared-steps --src-project 30 --src-suite 20069 --dst-project 31 --approve --save-mapping

# Шаг 2: Миграция кейсов с использованием mapping
gotr sync cases --src-project 30 --src-suite 20069 --dst-project 31 --dst-suite 19859 --mapping-file mapping.json --approve
```

### Сценарий 3: Только новые кейсы

```bash
# Dry-run для проверки
gotr sync cases --src-project 30 --src-suite 20069 --dst-project 31 --dst-suite 19859 --dry-run

# Реальный импорт
gotr sync cases --src-project 30 --src-suite 20069 --dst-project 31 --dst-suite 19859 --approve
```
