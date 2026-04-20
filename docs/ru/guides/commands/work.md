# Команда: work

Language: Русский | [English](../../../en/guides/commands/work.md)

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
        - [work](work.md)
      - [CRUD операции](add.md)
      - [Основные ресурсы](get.md)
      - [Специальные ресурсы](bdds.md)
    - [Инструкции](../instructions/index.md)
  - [Архитектура](../../architecture/index.md)
  - [Эксплуатация](../../operations/index.md)
  - [Отчёты](../../reports/index.md)
- [Главная](../../../../README_ru.md)


## Обзор 🎯

Интерактивный навигационный хаб для работы с TestRail. Объединяет все ключевые операции
(compare, sync, snap, CRUD) в единое меню с кросс-навигацией между командами.

> [!TIP]
> `gotr work` — рекомендуемая точка входа для интерактивной работы.
> Все подкоманды доступны через иерархическое меню.

## Синтаксис 🧩

```bash
gotr work
```

## Как это работает

### Главное меню

При запуске `gotr work` отображается выбор сервера, затем — главное меню:

```text
? What to do?
  ← Exit
  ─────────────────
  📊 Compare — compare projects
  🔄 Sync — migrate data
  📦 Snap — snapshots and rollback
  📋 Get — read resources
  ➕ Add — create resources
  ✏️  Update — modify resources
  🗑  Delete — remove resources
  📤 Export — export resources
```

### Групповые хабы

Выбор группы открывает подменю с конкретными подкомандами:

**Compare:**
```text
? Compare — what to compare?
  ← Back
  ─────────────────
  All resources (full scan)
  Cases
  ...
```

**Snap:**
```text
? Snap — what to do?
  ← Back
  ─────────────────
  📋 List snapshots
  🔍 View snapshot details
  ↻ Rollback a snapshot
  📤 Export snapshot
  🗑  Delete snapshot
```

### Кросс-навигация

После выполнения операции post-action меню предлагает переход к связанным командам:

- **📊 Compare** — перейти к сравнению проектов
- **🔄 Sync** — перейти к миграции
- **📦 Snap** — перейти к управлению снапшотами

Это позволяет работать в едином потоке: compare → sync → snap → compare, не выходя из интерактивного режима.

### Сессия

При подключении к серверу отображается баннер:

```text
Server: https://your.testrail.io
```

Контекст сессии (проект, suite) наследуется между командами внутри хаба.

## Примеры 🚀

### ▶️ Запуск навигационного хаба

```bash
gotr work
```

### ▶️ Типичный рабочий поток

1. `gotr work` → выбрать сервер
2. **Compare** → сравнить проекты
3. Из post-action → **Sync** → мигрировать данные
4. Из post-action → **Snap** → проверить снапшоты
5. ← Back → вернуться в главное меню

## 🧪 Чек-лист перед использованием

- [ ] Конфигурация gotr настроена (`gotr self-test`)
- [ ] Интерактивный режим включён (без `--non-interactive`)

## FAQ ❓

- ❓ **Вопрос:** Чем `gotr work` отличается от прямых команд?
  > ↪️ **Ответ:** `gotr work` — это навигационная обёртка. Все те же команды (`gotr compare all`, `gotr snap list`) работают напрямую. Хаб добавляет меню, кросс-навигацию и сессионный контекст.

- ❓ **Вопрос:** Можно ли использовать `gotr work` в CI/скриптах?
  > ↪️ **Ответ:** Нет. `gotr work` — строго интерактивная команда. Для CI используйте прямые команды с флагами.

---

← [Каталог команд](index.md) · [snap](snap.md)
