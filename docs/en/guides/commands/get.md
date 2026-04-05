# Команды GET

Команда `gotr get` — получение данных из TestRail

## What it does

- Основная операция для get
- Интеграция с другими командами  
- Поддержка интерактивного режима

## When to use

- Для операций get в тестировании
- Когда нужна автоматизация процесса
- В CI/CD конвейерах

## Examples

```bash
gotr get --help      # справка
gotr get --project 30 # базовый запуск
```

## Main flags

| Флаг | Описание |
| --- | --- |
| `--help` | Справка по команде |
| `--verbose` | Детальный вывод |
| `--dry-run` | Предпросмотр |

## FAQ

**Q: Как получить справку?**  
A: `gotr get --help`.

---

← [Commands](index.md) · [Guides](../index.md) · [Documentation](../../index.md)
