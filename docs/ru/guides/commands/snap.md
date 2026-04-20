# Команда: snap

Language: Русский | [English](../../../en/guides/commands/snap.md)

## Навигация

- [Документация](../../index.md)
  - [Гайды](../index.md)
    - [Установка](../installation.md)
    - [Конфигурация](../configuration.md)
    - [Интерактивный режим](../interactive-mode.md)
    - [Прогресс](../progress.md)
    - [Каталог команд](index.md)
      - [Общие](global-flags.md)
        - [global-flags](global-flags.md)
        - [config](config.md)
        - [completion](completion.md)
        - [self-test](self-test.md)
        - [snap](snap.md)
      - [CRUD операции](add.md)
      - [Основные ресурсы](get.md)
      - [Специальные ресурсы](bdds.md)
    - [Инструкции](../instructions/index.md)
  - [Архитектура](../../architecture/index.md)
  - [Эксплуатация](../../operations/index.md)
  - [Отчёты](../../reports/index.md)
- [Главная](../../../../README_ru.md)


## Обзор 🎯

Управление pre-mutation снапшотами: просмотр, инспекция, откат мутаций, экспорт и очистка.

Снапшоты автоматически создаются перед мутирующими операциями (`update`, `delete`, `add` и др.),
когда `snap.enabled = true` в конфигурации или передан флаг `--snapshot`.

> [!TIP]
> Для быстрого старта: выполните `gotr snap list` после любой мутирующей операции,
> чтобы увидеть доступные снапшоты.

## Синтаксис 🧩

```bash
gotr snap <subcommand> [args] [flags]
```

## Подкоманды 📋

| Подкоманда | Описание |
| ---------- | -------- |
| `list` | Таблица снапшотов с группировкой по серверам; интерактивный двухуровневый выбор |
| `info [id]` | Табличная карточка метаданных снапшота (JSON через `--format json`) |
| `rollback [id]` | Откат мутации по сохранённым данным |
| `export [id]` | Экспорт снапшота в портативный JSON-файл (с интерактивным выбором пути) |
| `delete [id]` | Удаление снапшота с диска и из манифеста |
| `gc` | Очистка orphan-снапшотов (есть на диске, нет в манифесте) |

> Все подкоманды, принимающие `[id]`, поддерживают интерактивный выбор:
> если `id` не указан, показывается интерактивный picker со списком снапшотов.

## Флаги подкоманд ⚙️

### rollback

```text
--dry-run              Предпросмотр изменений без применения (таблица diff)
--entity-ids string    Ограничить откат конкретными ID сущностей (через запятую)
```

### export

Второй позиционный аргумент — путь к файлу вывода (по умолчанию: `snapshot_<id>.json`).
В интерактивном режиме, если путь не указан, запрашиваются имя файла и директория.

## Глобальные флаги 🌐

```text
-k, --api-key string    API ключ TestRail
-c, --config            Создать дефолтный файл конфигурации
-f, --format string     Формат вывода: table, json, csv, md, html (default "table")
--insecure              Пропустить проверку TLS сертификата
--non-interactive       Отключить интерактивные подсказки; завершить с ошибкой если требуется ввод
-q, --quiet             Подавить служебный вывод (прогресс, статистику, сообщения о сохранении)
--url string            Базовый URL TestRail
-u, --username string   Email пользователя TestRail
```

## Уровни отката (Tiers)

| Tier | Операция | Поведение |
| ---- | -------- | --------- |
| Tier 1 | `update` | Полный откат — восстановление оригинальных значений полей |
| Tier 2 | `delete` | Re-create сущности с новым ID (оригинальный ID утерян) |
| Tier 2 | `add` | Удаление созданной сущности |
| Tier 3 | `result`, `labels` | Только информационный снапшот (откат невозможен) |

## Примеры 🚀

### ▶️ Сценарий 1: Просмотр снапшотов

🎯 **Цель:** увидеть все доступные снапшоты.

```bash
gotr snap list
```

В интерактивном режиме выполняется двухуровневый выбор: сначала сервер, затем снапшот.
В non-interactive режиме или при `--format` выводится таблица с колонкой SERVER:

```text
 #  ID                              SERVER                    OP      ENTITY  CATEGORY  STATUS     TIMESTAMP
 1  cases/1718000000_update_42      https://my.testrail.io    update  case    cases     available  2025-06-10 12:00:00
 2  custom/my_backup                https://my.testrail.io    delete  suite   custom    available  2025-06-10 12:05:00
```

Для JSON-вывода (скрипты, пайплайны):

```bash
gotr snap list --format json
```

---

### ▶️ Сценарий 2: Детали снапшота

🎯 **Цель:** просмотреть полные метаданные.

```bash
gotr snap info cases/1718000000_update_42
```

Выводит форматированную карточку:

```text
┌──────────── Snapshot Info ────────────┐
│ ID         │ cases/1718000000_update_42 │
│ Server     │ https://my.testrail.io     │
│ Operation  │ update case                │
│ Tier       │ T1 (full rollback)         │
│ Status     │ available                  │
│ Entity IDs │ 42                         │
│ ...        │ ...                        │
└────────────┴────────────────────────────┘
```

Для JSON-вывода (скрипты):

```bash
gotr snap info cases/1718000000_update_42 --format json
```

---

### ▶️ Сценарий 3: Предпросмотр отката (dry-run)

🎯 **Цель:** увидеть diff перед применением.

```bash
gotr snap rollback cases/1718000000_update_42 --dry-run
```

Пример вывода (с контекстом сервера):

```text
Server:    https://my.testrail.io
Snapshot:  cases/1718000000_update_42 (update case, T1)

The following changes will be applied:

ENTITY ID  FIELD     CURRENT           SNAPSHOT
42         title     Changed Title     Original Title
42         priority  3                 2
```

> При откате `delete`, если секция уже удалена на сервере (404/400),
> откат мягко пропускает re-create и продолжает работу.

---

### ▶️ Сценарий 4: Выполнение отката

🎯 **Цель:** откатить мутацию.

```bash
gotr snap rollback cases/1718000000_update_42
```

В интерактивном режиме сначала покажет diff-таблицу и запросит подтверждение.

---

### ▶️ Сценарий 5: Частичный откат по entity-ids

🎯 **Цель:** откатить только конкретные сущности из batch-операции.

```bash
gotr snap rollback sync/1718000000_sync_cases --entity-ids 42,43,44
```

---

### ▶️ Сценарий 6: Экспорт и очистка

🎯 **Цель:** сохранить снапшот как артефакт и очистить мусор.

```bash
# Экспорт
gotr snap export cases/1718000000_update_42 backup.json

# Удаление конкретного снапшота
gotr snap delete cases/1718000000_update_42

# Очистка orphan-снапшотов
gotr snap gc
```

## ⚡ Быстрый старт (30 секунд)

1. Выполните мутирующую операцию с включённым snap:
```bash
gotr update case 42 --json '{"title":"test"}' --snapshot
```
2. Просмотрите список снапшотов:
```bash
gotr snap list
```
3. Откатите при необходимости:
```bash
gotr snap rollback <snapshot_id>
```

## Конфигурация

Snap может быть включён глобально в `~/.gotr.yaml`:

```yaml
snap:
  enabled: true
```

Или через флаг `--snapshot` для отдельных операций.

Хранилище снапшотов: `~/.gotr/snaps/`.

## 🧪 Чек-лист перед использованием

- [ ] Конфигурация gotr настроена (`gotr self-test`)
- [ ] Snap включён (`snap.enabled: true` или `--snapshot`)
- [ ] Для rollback: TestRail API доступен (update/delete откаты выполняют API-вызовы)
- [ ] Для dry-run: API-доступ необходим для получения текущего состояния

## FAQ ❓

- ❓ **Вопрос:** Что происходит при откате delete?
  > ↪️ **Ответ:** Сущность пересоздаётся через API с новым ID (Tier 2). Оригинальный ID не восстанавливается — это ограничение TestRail API.

- ❓ **Вопрос:** Можно ли откатить дважды?
  > ↪️ **Ответ:** Нет. После отката статус снапшота меняется на `rolled_back`, повторный откат блокируется.

- ❓ **Вопрос:** Что делает gc?
  > ↪️ **Ответ:** Удаляет директории снапшотов на диске, которые отсутствуют в манифесте (orphans). Tracked снапшоты не затрагиваются.

- ❓ **Вопрос:** Где хранятся данные?
  > ↪️ **Ответ:** В `~/.gotr/snaps/` — по категориям (`cases/`, `sync/`, `custom/` и т.д.). Манифест: `~/.gotr/snaps/manifest.json`.

## 🛠️ Диагностика

| Проблема | Решение |
| -------- | ------- |
| "snapshot not found" | Проверьте ID через `gotr snap list` |
| "already rolled_back" | Снапшот уже использован; создайте новый |
| Tier 3 rollback fails | Tier 3 операции (result, labels) не поддерживают откат |
| Пустой список | Убедитесь что `snap.enabled: true` или используйте `--snapshot` |

---

← [Каталог команд](index.md) · [self-test](self-test.md)
