# Команды completion

Команда `gotr completion` — автодополнение для shell

## What it does

- Основная операция для completion
- Интеграция с другими командами  
- Поддержка интерактивного режима

## When to use

- Для операций completion в тестировании
- Когда нужна автоматизация процесса
- В CI/CD конвейерах

## Examples

```bash
gotr completion --help      # справка
gotr completion --project 30 # базовый запуск
```

## Main flags

| Флаг | Описание |
| --- | --- |
| `--help` | Справка по команде |
| `--verbose` | Детальный вывод |
| `--dry-run` | Предпросмотр |

## FAQ

**Q: Как получить справку?**  
A: `gotr completion --help`.

---

← [Commands](index.md) · [Guides](../index.md) · [Documentation](../../index.md)
