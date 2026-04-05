# Команды delete

Команда `gotr delete` — удаление элементов (БЕЗОПАСНО!)

## What it does

- Основная операция для delete
- Интеграция с другими командами  
- Поддержка интерактивного режима

## When to use

- Для операций delete в тестировании
- Когда нужна автоматизация процесса
- В CI/CD конвейерах

## Examples

```bash
gotr delete --help      # справка
gotr delete --project 30 # базовый запуск
```

## Main flags

| Флаг | Описание |
| --- | --- |
| `--help` | Справка по команде |
| `--verbose` | Детальный вывод |
| `--dry-run` | Предпросмотр |

## FAQ

**Q: Как получить справку?**  
A: `gotr delete --help`.

---

← [Commands](index.md) · [Guides](../index.md) · [Documentation](../../index.md)
