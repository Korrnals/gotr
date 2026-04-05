# Команды export

Команда `gotr export` — экспорт данных в различные форматы

## What it does

- Основная операция для export
- Интеграция с другими командами  
- Поддержка интерактивного режима

## When to use

- Для операций export в тестировании
- Когда нужна автоматизация процесса
- В CI/CD конвейерах

## Examples

```bash
gotr export --help      # справка
gotr export --project 30 # базовый запуск
```

## Main flags

| Флаг | Описание |
| --- | --- |
| `--help` | Справка по команде |
| `--verbose` | Детальный вывод |
| `--dry-run` | Предпросмотр |

## FAQ

**Q: Как получить справку?**  
A: `gotr export --help`.

---

← [Commands](index.md) · [Guides](../index.md) · [Documentation](../../index.md)
